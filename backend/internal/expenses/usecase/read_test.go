package usecase_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func TestGet_ReturnsALiveExpenseWithoutWarnings(t *testing.T) {
	h := seeded(t, func(in *usecase.CreateInput) { in.Draft.HasCFDI = false })
	h.taxpayers.byID[taxpayerA].archived = true // reads stay available

	got, err := h.uc.Get(t.Context(), userA, taxpayerA, expense1)

	require.NoError(t, err)
	assert.Equal(t, expense1, got.ID)
	assert.Empty(t, got.Warnings)
	assert.Equal(t, domain.Legend, got.Legend)
	assert.Contains(t, h.tx.modes, "ro")
}

func TestGet_NotFound(t *testing.T) {
	h := seeded(t)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))

	for name, in := range map[string][3]string{
		"deleted":          {userA, taxpayerA, expense1},
		"another user":     {userB, taxpayerB, expense1},
		"another taxpayer": {userA, taxpayerM, expense1},
		"malformed":        {userA, taxpayerA, "x"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := h.uc.Get(t.Context(), in[0], in[1], in[2])

			requireAppErr(t, err, apperr.CodeNotFound)
		})
	}
}

// seedMonth creates n September expenses of 100.00 + 16.00 on days 1..n,
// plus one in October and one deleted in September.
func seedMonth(t *testing.T, h *harness, n int) {
	t.Helper()
	for i := 1; i <= n+2; i++ {
		in := createInput(t, fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", i))
		in.Draft.Amount, in.Draft.IVA = 100_00, 16_00
		in.Draft.PaidOn = date(t, fmt.Sprintf("2026-09-%02d", min(i, 28)))
		if i == n+1 {
			in.Draft.PaidOn = date(t, "2026-10-01")
		}
		_, err := h.uc.Create(t.Context(), in)
		require.NoError(t, err)
		if i == n+2 {
			require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, in.ID))
		}
	}
}

func listInput(cursor string, limit int) usecase.ListInput {
	return usecase.ListInput{UserID: userA, TaxpayerID: taxpayerA, Period: "2026-09", Cursor: cursor, Limit: limit}
}

func TestList_PagesByKeysetWithPeriodTotals(t *testing.T) {
	h := newHarness(t)
	seedMonth(t, h, 5)

	first, err := h.uc.List(t.Context(), listInput("", 2))
	require.NoError(t, err)
	require.Len(t, first.Items, 2)
	assert.Equal(t, "2026-09-05", first.Items[0].PaidOn.String(), "newest first")
	assert.Equal(t, "2026-09-04", first.Items[1].PaidOn.String())
	assert.NotEmpty(t, first.NextCursor)
	assert.Equal(t, usecase.Totals{Count: 5, Amount: 500_00, IVA: 80_00}, first.Totals, "the whole period, not the page")
	assert.Equal(t, domain.Legend, first.Legend)

	var seen []string
	for page, cursor := first, first.NextCursor; ; {
		for _, e := range page.Items {
			seen = append(seen, e.PaidOn.String())
			assert.Empty(t, e.Warnings)
		}
		if cursor == "" {
			break
		}
		page, err = h.uc.List(t.Context(), listInput(cursor, 2))
		require.NoError(t, err)
		cursor = page.NextCursor
	}
	assert.Equal(t, []string{"2026-09-05", "2026-09-04", "2026-09-03", "2026-09-02", "2026-09-01"}, seen)
}

func TestList_LastFullPageHasNoCursorWhenNothingFollows(t *testing.T) {
	h := newHarness(t)
	seedMonth(t, h, 2)

	page, err := h.uc.List(t.Context(), listInput("", 2))

	require.NoError(t, err)
	assert.Len(t, page.Items, 2)
	assert.Empty(t, page.NextCursor)
}

func TestList_DefaultLimitAndRegimeFilter(t *testing.T) {
	h := newHarness(t)
	seedMonth(t, h, 3)
	in := createInput(t, expense2)
	in.Regime, in.Draft.Category = fiscal.RegimeRent, "predial"
	_, err := h.uc.Create(t.Context(), in)
	require.NoError(t, err)

	all, err := h.uc.List(t.Context(), listInput("", 0))
	require.NoError(t, err)
	assert.Len(t, all.Items, 4)

	filtered := listInput("", 0)
	filtered.Regime = fiscal.RegimeRent
	only, err := h.uc.List(t.Context(), filtered)
	require.NoError(t, err)
	require.Len(t, only.Items, 1)
	assert.Equal(t, usecase.Totals{Count: 1, Amount: 500_00, IVA: 80_00}, only.Totals)
}

func TestList_ArchivedTaxpayerIsReadable(t *testing.T) {
	h := newHarness(t)
	seedMonth(t, h, 1)
	h.taxpayers.byID[taxpayerA].archived = true

	page, err := h.uc.List(t.Context(), listInput("", 10))

	require.NoError(t, err)
	assert.Len(t, page.Items, 1)
}

func TestList_Refusals(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*usecase.ListInput)
		code   apperr.Code
		field  []string
	}{
		{"foreign taxpayer", func(in *usecase.ListInput) { in.TaxpayerID = taxpayerB }, apperr.CodeNotFound, nil},
		{"malformed taxpayer wins over a bad period", func(in *usecase.ListInput) { in.TaxpayerID, in.Period = "x", "bad" },
			apperr.CodeNotFound, nil},
		{"period missing", func(in *usecase.ListInput) { in.Period = "" }, apperr.CodeValidationFailed, []string{"period", "required"}},
		{"period malformed", func(in *usecase.ListInput) { in.Period = "2026-9" }, apperr.CodeValidationFailed, []string{"period", "invalid"}},
		{"period with a day", func(in *usecase.ListInput) { in.Period = "2026-09-01" }, apperr.CodeValidationFailed, []string{"period", "invalid"}},
		{"cursor garbage", func(in *usecase.ListInput) { in.Cursor = "%%%" }, apperr.CodeValidationFailed, []string{"cursor", "invalid"}},
		{"limit too big", func(in *usecase.ListInput) { in.Limit = 101 }, apperr.CodeValidationFailed, []string{"limit", "out_of_range"}},
		{"limit negative", func(in *usecase.ListInput) { in.Limit = -1 }, apperr.CodeValidationFailed, []string{"limit", "out_of_range"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			in := listInput("", 10)
			c.mutate(&in)

			_, err := h.uc.List(t.Context(), in)

			requireAppErr(t, err, c.code, c.field...)
		})
	}
}

func categoriesInput(regime fiscal.RegimeCode, on string) usecase.CategoriesInput {
	return usecase.CategoriesInput{UserID: userA, TaxpayerID: taxpayerA, Regime: regime, On: on}
}

func hintOf(t *testing.T, list usecase.CategoryList, code domain.CategoryCode) usecase.Category {
	t.Helper()
	for _, c := range list.Items {
		if c.Code == code {
			return c
		}
	}
	t.Fatalf("category %q missing", code)
	return usecase.Category{}
}

func TestCategories_ResolvedForTheRegimeAndItsOptions(t *testing.T) {
	h := newHarness(t)
	h.access.readOnly = true // reads stay available

	business, err := h.uc.Categories(t.Context(), categoriesInput(fiscal.RegimeBusiness, ""))
	require.NoError(t, err)
	assert.Len(t, business.Items, len(businessRules))
	software := hintOf(t, business, "software")
	assert.Equal(t, usecase.Category{
		Code: "software", Name: "Software y suscripciones", Deductibility: domain.Deductible,
		DefaultDeductible: true, LegalBasis: "LISR 103",
	}, software)
	assert.Equal(t, "Sólo deducible si lo pagas con tarjeta", hintOf(t, business, "combustible").Note)
	assert.Equal(t, money.Cents(2_000_00), business.CashPaymentLimit)
	assert.Equal(t, []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}, business.IVARatesAllowed)
	assert.Equal(t, domain.Legend, business.Legend)

	blind, err := h.uc.Categories(t.Context(), categoriesInput(fiscal.RegimeRent, "2026-09-12"))
	require.NoError(t, err)
	assert.Equal(t, domain.ISRNotApplicable, hintOf(t, blind, "mantenimiento").Deductibility)
	assert.False(t, hintOf(t, blind, "mantenimiento").DefaultDeductible)
	assert.Equal(t, domain.Deductible, hintOf(t, blind, "predial").Deductibility)

	resico, err := h.uc.Categories(t.Context(), categoriesInput(fiscal.RegimeRESICO, "2025-05-01"))
	require.NoError(t, err)
	assert.Equal(t, domain.ISRNotApplicable, hintOf(t, resico, "software").Deductibility)
	assert.Empty(t, h.access.lastCalls)
}

func TestCategories_ArchivedTaxpayerIsReadable(t *testing.T) {
	h := newHarness(t)
	h.taxpayers.byID[taxpayerA].archived = true

	_, err := h.uc.Categories(t.Context(), categoriesInput(fiscal.RegimeBusiness, ""))

	require.NoError(t, err)
}

func TestCategories_Refusals(t *testing.T) {
	cases := []struct {
		name  string
		in    usecase.CategoriesInput
		code  apperr.Code
		field []string
	}{
		{"foreign taxpayer", usecase.CategoriesInput{UserID: userA, TaxpayerID: taxpayerB, Regime: "612"}, apperr.CodeNotFound, nil},
		{"regime not held on the date", categoriesInput(fiscal.RegimeRESICO, ""), apperr.CodeBusinessRuleViolation, []string{"regime", "regime_not_active"}},
		{"regime never held", categoriesInput("605", ""), apperr.CodeBusinessRuleViolation, []string{"regime", "regime_not_active"}},
		{"regime missing", categoriesInput("", ""), apperr.CodeValidationFailed, []string{"regime", "required"}},
		{"on malformed", categoriesInput("612", "2026/09/01"), apperr.CodeValidationFailed, []string{"on", "invalid"}},
		{"on in the future", categoriesInput("612", "2026-10-14"), apperr.CodeValidationFailed, []string{"on", "out_of_range"}},
		{"on before the catalog", categoriesInput("612", "2021-12-31"), apperr.CodeValidationFailed, []string{"on", "out_of_range"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)

			_, err := h.uc.Categories(t.Context(), c.in)

			requireAppErr(t, err, c.code, c.field...)
		})
	}
}

func TestCategories_DateWithoutParametersIsOutOfRange(t *testing.T) {
	h := newHarness(t)
	h.parameters.uncovered = true
	h.wire()

	_, err := h.uc.Categories(t.Context(), categoriesInput(fiscal.RegimeBusiness, ""))

	requireAppErr(t, err, apperr.CodeValidationFailed, "on", "out_of_range")
}
