package usecase_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func TestGet(t *testing.T) {
	h := seeded(t)
	got, err := h.uc.Get(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err)
	assert.Equal(t, incomeID, got.ID)
	assert.Equal(t, []string{"rw", "ro"}, h.tx.modes, "reads run read-only")

	for name, ids := range map[string][2]string{
		"another taxpayer's path": {taxpayerA2, incomeID},
		"unknown income":          {taxpayerA, generated},
		"malformed income id":     {taxpayerA, "x"},
		"malformed taxpayer id":   {"x", incomeID},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := h.uc.Get(t.Context(), userA, ids[0], ids[1])
			requireApp(t, err, apperr.CodeNotFound)
		})
	}
	t.Run("soft-deleted", func(t *testing.T) {
		require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
		_, err := h.uc.Get(t.Context(), userA, taxpayerA, incomeID)
		requireApp(t, err, apperr.CodeNotFound)
	})
}

// seedMonth creates n incomes of taxpayerA in September 2026 (612, cash),
// one per day, with ids ordered like the days, plus one in August and one
// 606 in September.
func seedMonth(t *testing.T, h *harness, n int) {
	t.Helper()
	add := func(id, regime, collected string, amount money.Cents) {
		in := usecase.IncomeInput{
			Regime: fiscal.RegimeCode(regime), Counterparty: fiscal.ReceptorPFNoActivityOrPublic,
			Amount: amount, IVARate: fiscal.IVAExempt, Dates: domain.Dates{CollectedOn: date(t, collected)},
		}
		_, err := h.uc.Create(t.Context(), userA, create(id, taxpayerA, in))
		require.NoError(t, err)
	}
	for i := range n {
		add(fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", i+1), "612", fmt.Sprintf("2026-09-%02d", i%28+1), 100_00)
	}
	add("0199d4a3-0000-7000-8000-0000000000aa", "612", "2026-08-31", 1_00)
	add("0199d4a3-0000-7000-8000-0000000000bb", "606", "2026-09-15", 7_00)
}

func TestList_KeysetPagesCoverTheMonth(t *testing.T) {
	h := newHarness(t)
	seedMonth(t, h, 5)

	var seen []string
	query := usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Regime: "612", Limit: 2}
	for page := 0; ; page++ {
		require.Less(t, page, 5, "pagination must end")
		got, err := h.uc.List(t.Context(), userA, query)
		require.NoError(t, err)
		assert.Equal(t, usecase.PeriodTotals{Count: 5, Amount: 500_00}, got.Totals, "totals cover the period, not the page")
		for _, item := range got.Items {
			seen = append(seen, item.ID)
		}
		if got.NextCursor == "" {
			break
		}
		query.Cursor = got.NextCursor
	}
	assert.Equal(t, []string{
		"0199d4a3-0000-7000-8000-000000000005", "0199d4a3-0000-7000-8000-000000000004",
		"0199d4a3-0000-7000-8000-000000000003", "0199d4a3-0000-7000-8000-000000000002",
		"0199d4a3-0000-7000-8000-000000000001",
	}, seen, "newest first, no repeats")
	assert.Equal(t, "2026-09-01", h.store.lastList.filter.From.String())
	assert.Equal(t, "2026-10-01", h.store.lastList.filter.Until.String())
	assert.Equal(t, 3, h.store.lastList.limit, "one extra row tells whether a next page exists")
}

func TestList_FiltersAndDefaults(t *testing.T) {
	h := newHarness(t)
	seedMonth(t, h, 2)

	all, err := h.uc.List(t.Context(), userA, usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09"})
	require.NoError(t, err)
	assert.Len(t, all.Items, 3, "both regimes")
	assert.Equal(t, int64(3), all.Totals.Count)
	assert.Empty(t, all.NextCursor)
	assert.Equal(t, usecase.DefaultLimit+1, h.store.lastList.limit)

	only606, err := h.uc.List(t.Context(), userA, usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Regime: "606"})
	require.NoError(t, err)
	require.Len(t, only606.Items, 1)
	assert.Equal(t, money.Cents(7_00), only606.Totals.Amount)

	december, err := h.uc.List(t.Context(), userA, usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-12"})
	require.NoError(t, err)
	assert.Empty(t, december.Items)
	assert.Equal(t, "2027-01-01", h.store.lastList.filter.Until.String(), "December rolls over the year")
}

func TestList_ArchivedTaxpayerIsListed(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.List(t.Context(), userA, usecase.ListQuery{TaxpayerID: archivedA, Period: "2026-09"})
	require.NoError(t, err)
}

func TestList_Refusals(t *testing.T) {
	tests := []struct {
		name  string
		query usecase.ListQuery
		code  apperr.Code
		field []string
	}{
		{"malformed taxpayer id", usecase.ListQuery{TaxpayerID: "x", Period: "bad"}, apperr.CodeNotFound, nil},
		{"unknown taxpayer", usecase.ListQuery{TaxpayerID: "0199d4a2-1111-7000-8000-0000000000ff", Period: "2026-09"}, apperr.CodeNotFound, nil},
		{"malformed period", usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-13"}, apperr.CodeValidationFailed, []string{"period", "invalid"}},
		{"missing period", usecase.ListQuery{TaxpayerID: taxpayerA}, apperr.CodeValidationFailed, []string{"period", "required"}},
		{"limit above 100", usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Limit: 101}, apperr.CodeValidationFailed, []string{"limit", "out_of_range"}},
		{"negative limit", usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Limit: -1}, apperr.CodeValidationFailed, []string{"limit", "out_of_range"}},
		{"forged cursor", usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Cursor: "not-a-cursor"}, apperr.CodeValidationFailed, []string{"cursor", "invalid"}},
		{"cursor with a bad id", usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Cursor: "MjAyNi0wOS0wMXx4"}, apperr.CodeValidationFailed, []string{"cursor", "invalid"}},
		{"cursor with a bad date", usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09", Cursor: "eHx5"}, apperr.CodeValidationFailed, []string{"cursor", "invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newHarness(t).uc.List(t.Context(), userA, tt.query)
			requireApp(t, err, tt.code, tt.field...)
		})
	}
}

func TestDelete(t *testing.T) {
	h := seeded(t)
	h.access.readOnly = true // reads and deletes stay for a read-only user

	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID), "idempotent")

	assert.Equal(t, 1, h.store.deletes, "the second delete changes nothing")
	assert.Zero(t, h.access.ensures+h.access.reserves, "delete needs no access check")
	assert.Contains(t, h.logs.String(), `"msg":"income.deleted"`)
	page, err := h.uc.List(t.Context(), userA, usecase.ListQuery{TaxpayerID: taxpayerA, Period: "2026-09"})
	require.NoError(t, err)
	assert.Empty(t, page.Items)
	assert.Zero(t, page.Totals.Count)
}

func TestDelete_Refusals(t *testing.T) {
	tests := []struct {
		name, taxpayer, id string
		code               apperr.Code
		field              []string
	}{
		{"another taxpayer's path", taxpayerA2, incomeID, apperr.CodeNotFound, nil},
		{"unknown income", taxpayerA, generated, apperr.CodeNotFound, nil},
		{"malformed id", taxpayerA, "x", apperr.CodeNotFound, nil},
		{"archived taxpayer", archivedA, incomeID, apperr.CodeBusinessRuleViolation, []string{"taxpayerId", "taxpayer_archived"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := seeded(t)
			if tt.taxpayer == archivedA {
				_, err := h.uc.Create(t.Context(), userA, create(generated, archivedA, us1(t)))
				require.Error(t, err, "archived: no create either")
				h.store.rows[incomeID] = withTaxpayer(h.store.rows[incomeID], archivedA)
			}
			err := h.uc.Delete(t.Context(), userA, tt.taxpayer, tt.id)
			requireApp(t, err, tt.code, tt.field...)
			assert.Zero(t, h.store.deletes)
		})
	}
}

func withTaxpayer(row usecase.StoredIncome, taxpayerID string) usecase.StoredIncome {
	row.TaxpayerID = taxpayerID
	return row
}

func TestRestore(t *testing.T) {
	h := seeded(t)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
	h.access.readOnly = true // coordinator decision: a read-only user may undo a delete
	h.access.limit, h.access.used = 5, 5

	got, err := h.uc.Restore(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err)
	assert.Equal(t, incomeID, got.Income.ID)
	assert.Zero(t, h.access.ensures+h.access.reserves, "restore never consumes a slot nor checks access")
	assert.Equal(t, 5, h.access.used, "incomes_created stays 5")
	assert.Equal(t, []regimeCall{{taxpayerA, "612", "2026-09"}}, h.taxpayers.calls, "the regime window is re-checked")
	assert.Contains(t, h.logs.String(), `"msg":"income.restored"`)

	_, err = h.uc.Get(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err, "back in place")

	again, err := h.uc.Restore(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err, "restoring a live income is a no-op")
	assert.Equal(t, got.Income, again.Income)
	assert.Equal(t, 2, h.store.deletes, "one delete, one restore")
}

func TestRestore_WarnsOfADuplicateFolio(t *testing.T) {
	h := seeded(t)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
	_, err := h.uc.Create(t.Context(), userA, create(generated, taxpayerA, us1(t)))
	require.NoError(t, err)

	got, err := h.uc.Restore(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err)
	assert.Equal(t, []domain.Warning{{Code: domain.WarningDuplicateCFDIUUID, Field: "cfdi_uuid"}}, got.Warnings)
}

func TestRestore_Refusals(t *testing.T) {
	tests := []struct {
		name     string
		taxpayer string
		id       string
		prepare  func(*harness)
		code     apperr.Code
		field    []string
	}{
		{name: "regime ended before the income's date", taxpayer: taxpayerA, id: incomeID,
			prepare: func(h *harness) { h.taxpayers.windows[taxpayerA+"|612"] = [2]string{"2022-01", "2026-08"} },
			code:    apperr.CodeBusinessRuleViolation, field: []string{"regime_code", "regime_not_active"}},
		{name: "archived taxpayer", taxpayer: taxpayerA, id: incomeID,
			prepare: func(h *harness) {
				p := h.taxpayers.profiles[taxpayerA]
				p.Archived = true
				h.taxpayers.profiles[taxpayerA] = p
			}, code: apperr.CodeBusinessRuleViolation, field: []string{"taxpayerId", "taxpayer_archived"}},
		{name: "purged or unknown", taxpayer: taxpayerA, id: generated, code: apperr.CodeNotFound},
		{name: "another taxpayer's path", taxpayer: taxpayerA2, id: incomeID, code: apperr.CodeNotFound},
		{name: "malformed taxpayer id", taxpayer: "x", id: incomeID, code: apperr.CodeNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := seeded(t)
			require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
			if tt.prepare != nil {
				tt.prepare(h)
			}
			_, err := h.uc.Restore(t.Context(), userA, tt.taxpayer, tt.id)
			requireApp(t, err, tt.code, tt.field...)
			assert.True(t, h.store.rows[incomeID].Deleted, "still deleted")
		})
	}
}

// Restore derives again (ADR-0008): with the taxpayer's current 625 options,
// and it warns like any write.
func TestRestore_ReDerivesTheWithholding(t *testing.T) {
	h := newHarness(t)
	in := usecase.IncomeInput{
		Regime: "625", Counterparty: fiscal.ReceptorPFNoActivityOrPublic, Amount: 15_000_00,
		IVARate: fiscal.IVARate16, IVA: 2_400_00, Dates: domain.Dates{CollectedOn: date(t, "2026-09-20")},
	}
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, platformA, in))
	require.NoError(t, err)
	require.NoError(t, h.uc.Delete(t.Context(), userA, platformA, incomeID))
	p := h.taxpayers.profiles[platformA]
	p.Regimes[0].RFCGivenToPlatform = false
	h.taxpayers.profiles[platformA] = p

	got, err := h.uc.Restore(t.Context(), userA, platformA, incomeID)

	require.NoError(t, err)
	assert.Equal(t, fiscal.Withholdings{ISR: 3_000_00, IVA: 2_400_00}, got.Income.Withholding.Withholdings, "20 % and all the IVA without RFC")
	assert.Equal(t, got.Income.Withholding, h.store.rows[incomeID].Withholding)
}

func TestRestore_WarnsOfAnUnverifiedRule(t *testing.T) {
	h := newHarness(t)
	in := us1(t)
	in.Counterparty = fiscal.ReceptorGovFederal
	_, err := h.uc.Create(t.Context(), userA, create(incomeID, taxpayerA, in))
	require.NoError(t, err)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, incomeID))
	got, err := h.uc.Restore(t.Context(), userA, taxpayerA, incomeID)
	require.NoError(t, err)
	assert.Equal(t, []domain.Warning{{Code: domain.WarningWithholdingCPAUnverified, Field: domain.FieldWithholding}}, got.Warnings)
}
