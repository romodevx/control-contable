package usecase_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func TestCreate_RecordsADeductibleExpenseAndReservesOneSlot(t *testing.T) {
	h := newHarness(t)

	got, err := h.uc.Create(t.Context(), createInput(t, expense1))

	require.NoError(t, err)
	assert.True(t, got.Created)
	e := got.Expense
	assert.Equal(t, expense1, e.ID)
	assert.Equal(t, taxpayerA, e.TaxpayerID)
	assert.Equal(t, fiscal.RegimeBusiness, e.Regime)
	assert.True(t, e.Deductible)
	assert.Equal(t, domain.SourceCategoryDefault, e.Source)
	assert.Empty(t, e.Warnings)
	assert.Equal(t, domain.Legend, e.Legend)
	assert.Equal(t, "Suscripción anual de diseño", e.Description, "trimmed")
	assert.Equal(t, []access.RecordKind{access.RecordExpense}, h.access.reserved)
	assert.Equal(t, []string{"612@2026-09"}, h.taxpayers.checked)
	assert.Equal(t, []string{expense1}, h.expenses.locked, "a client id is serialized before the lookup")
	assert.True(t, logsMention(h, `"msg":"expense.created"`))
	assert.True(t, logsMention(h, expense1))
	assert.False(t, logsMention(h, "Suscripción"), "the description is never logged")
}

// MODULE-incomes §2.10: idempotency lookup → read-only check → CheckRegime →
// rules → ReserveRecord → insert.
func TestCreate_ChecksAccessBeforeTheRegimeAndReservesLast(t *testing.T) {
	h := newHarness(t)

	_, err := h.uc.Create(t.Context(), createInput(t, expense1))

	require.NoError(t, err)
	assert.Equal(t, []string{"writable", "reserve"}, h.access.lastCalls)
}

func TestCreate_WithoutAClientIDUsesAServerID(t *testing.T) {
	h := newHarness(t)

	got, err := h.uc.Create(t.Context(), createInput(t, ""))

	require.NoError(t, err)
	assert.Equal(t, serverID, got.Expense.ID)
	assert.Empty(t, h.expenses.locked)
}

func TestCreate_CanonicalizesTheClientID(t *testing.T) {
	h := newHarness(t)

	got, err := h.uc.Create(t.Context(), createInput(t, strings.ToUpper(expense1)))

	require.NoError(t, err)
	assert.Equal(t, expense1, got.Expense.ID)
}

func TestCreate_RefusesAMalformedClientID(t *testing.T) {
	h := newHarness(t)

	_, err := h.uc.Create(t.Context(), createInput(t, "not-a-uuid"))

	requireAppErr(t, err, apperr.CodeValidationFailed, "id", "invalid")
}

func TestCreate_IdempotentRetryReturnsTheOriginalWithoutASlot(t *testing.T) {
	h := newHarness(t)
	first, err := h.uc.Create(t.Context(), createInput(t, expense1))
	require.NoError(t, err)
	h.access.capSpent = true // a retry at 5/5 still gets the original

	again, err := h.uc.Create(t.Context(), createInput(t, expense1))

	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Equal(t, first.Expense.ID, again.Expense.ID)
	assert.Equal(t, first.Expense.CreatedAt, again.Expense.CreatedAt)
	assert.Len(t, h.access.reserved, 1)
	assert.Equal(t, domain.Legend, again.Expense.Legend)
}

func TestCreate_IdempotentRetryWithAnExplicitDefaultOverrideMatches(t *testing.T) {
	h := newHarness(t)
	in := createInput(t, expense1)
	in.Draft.Deductible = domain.ClaimDeductible
	_, err := h.uc.Create(t.Context(), in)
	require.NoError(t, err)

	again, err := h.uc.Create(t.Context(), in)

	require.NoError(t, err)
	assert.False(t, again.Created)
	assert.Equal(t, domain.SourceUser, again.Expense.Source)
}

func TestCreate_ReusedIDConflicts(t *testing.T) {
	cases := map[string]func(*usecase.CreateInput){
		"other amount":      func(in *usecase.CreateInput) { in.Draft.Amount, in.Draft.IVA = 600_00, 96_00 },
		"other description": func(in *usecase.CreateInput) { in.Description = "otra" },
		"other override":    func(in *usecase.CreateInput) { in.Draft.Deductible = domain.ClaimNotDeductible },
		"other taxpayer":    func(in *usecase.CreateInput) { in.TaxpayerID = taxpayerM; in.Regime = "601" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			_, err := h.uc.Create(t.Context(), createInput(t, expense1))
			require.NoError(t, err)
			in := createInput(t, expense1)
			mutate(&in)

			_, err = h.uc.Create(t.Context(), in)

			requireAppErr(t, err, apperr.CodeConflict, "id", "id_reused")
			assert.Len(t, h.access.reserved, 1)
		})
	}
}

func TestCreate_ReusedIDOfADeletedExpenseConflicts(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), createInput(t, expense1))
	require.NoError(t, err)
	require.NoError(t, h.uc.Delete(t.Context(), userA, taxpayerA, expense1))

	_, err = h.uc.Create(t.Context(), createInput(t, expense1))

	requireAppErr(t, err, apperr.CodeConflict, "id", "id_reused")
}

func TestCreate_ForeignIDConflictsWithoutDetail(t *testing.T) {
	h := newHarness(t)
	in := createInput(t, expense1)
	in.UserID, in.TaxpayerID = userB, taxpayerB
	_, err := h.uc.Create(t.Context(), in)
	require.NoError(t, err)

	_, err = h.uc.Create(t.Context(), createInput(t, expense1))

	appErr := appErrOf(t, err, apperr.CodeConflict)
	assert.Empty(t, appErr.Fields, "a foreign id reveals no ownership")
	require.ErrorIs(t, err, usecase.ErrIDConflict)
}

func TestCreate_AccessRefusals(t *testing.T) {
	t.Run("read-only", func(t *testing.T) {
		h := newHarness(t)
		h.access.readOnly = true

		_, err := h.uc.Create(t.Context(), createInput(t, expense1))

		requireAppErr(t, err, apperr.CodeEntitlementRequired)
		assert.Empty(t, h.access.reserved)
		assert.Empty(t, h.taxpayers.checked, "the read-only check precedes the regime check")
	})
	t.Run("trial cap names the expenses counter", func(t *testing.T) {
		h := newHarness(t)
		h.access.capSpent = true

		_, err := h.uc.Create(t.Context(), createInput(t, expense1))

		requireAppErr(t, err, apperr.CodeEntitlementRequired, "expenses_remaining", access.DetailTrialCapReached)
		_, findErr := h.expenses.Find(t.Context(), expense1, usecase.LockNone)
		require.ErrorIs(t, findErr, usecase.ErrExpenseNotFound)
	})
}

func TestCreate_TaxpayerAndRegimeRefusals(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*usecase.CreateInput)
		code   apperr.Code
		field  []string
	}{
		{"foreign taxpayer", func(in *usecase.CreateInput) { in.TaxpayerID = taxpayerB }, apperr.CodeNotFound, nil},
		{"malformed taxpayer id", func(in *usecase.CreateInput) { in.TaxpayerID = "nope" }, apperr.CodeNotFound, nil},
		{"regime ended before paid_on", func(in *usecase.CreateInput) { in.Regime = fiscal.RegimeRESICO },
			apperr.CodeBusinessRuleViolation, []string{"regime_code", "regime_not_active"}},
		{"regime missing", func(in *usecase.CreateInput) { in.Regime = "" },
			apperr.CodeValidationFailed, []string{"regime_code", "required"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			in := createInput(t, expense1)
			c.mutate(&in)

			_, err := h.uc.Create(t.Context(), in)

			requireAppErr(t, err, c.code, c.field...)
			assert.Empty(t, h.access.reserved)
		})
	}
}

func TestCreate_ArchivedTaxpayerIsRefused(t *testing.T) {
	h := newHarness(t)
	h.taxpayers.byID[taxpayerA].archived = true

	_, err := h.uc.Create(t.Context(), createInput(t, expense1))

	requireAppErr(t, err, apperr.CodeBusinessRuleViolation, "taxpayerId", "taxpayer_archived")
}

func TestCreate_ValidationReportsEveryField(t *testing.T) {
	h := newHarness(t)
	in := createInput(t, expense1)
	in.Draft.Amount = 0
	in.Draft.PaymentMethod = ""
	in.Draft.CFDIUUID = uuidOf(t, cfdiUUID)
	in.Draft.HasCFDI = false

	_, err := h.uc.Create(t.Context(), in)

	appErr := appErrOf(t, err, apperr.CodeValidationFailed)
	assert.ElementsMatch(t, []apperr.FieldError{
		{Field: "amount_cents", Code: "must_be_positive", Message: "This amount must be greater than zero."},
		{Field: "payment_method", Code: "required", Message: "This field is required."},
		{Field: "cfdi_uuid", Code: "has_cfdi_mismatch", Message: "A folio fiscal needs has_cfdi to be true."},
	}, appErr.Fields)
	assert.Empty(t, h.access.reserved)
	require.ErrorIs(t, err, domain.ErrHasCFDIMismatch)
}

func TestCreate_UnknownCategoryIsAValidationError(t *testing.T) {
	h := newHarness(t)
	in := createInput(t, expense1)
	in.Draft.Category = "viajes-luna"
	in.Draft.PaymentMethod = "teletransporte"

	_, err := h.uc.Create(t.Context(), in)

	appErr := appErrOf(t, err, apperr.CodeValidationFailed)
	assert.ElementsMatch(t, []string{"category_code:unknown_category", "payment_method:invalid"}, fieldCodes(appErr))
}

func TestCreate_MissingCategoryIsRequired(t *testing.T) {
	h := newHarness(t)
	in := createInput(t, expense1)
	in.Draft.Category = ""

	_, err := h.uc.Create(t.Context(), in)

	requireAppErr(t, err, apperr.CodeValidationFailed, "category_code", "required")
}

func TestCreate_PaidOnOutOfRangeSkipsTheRegimeCheck(t *testing.T) {
	for _, raw := range []string{"2026-10-14", "2021-12-31"} {
		t.Run(raw, func(t *testing.T) {
			h := newHarness(t)
			in := createInput(t, expense1)
			in.Draft.PaidOn = date(t, raw)

			_, err := h.uc.Create(t.Context(), in)

			requireAppErr(t, err, apperr.CodeValidationFailed, "paid_on", "out_of_range")
			assert.Empty(t, h.taxpayers.checked)
		})
	}
}

func TestCreate_DescriptionRules(t *testing.T) {
	t.Run("blank is stored as none", func(t *testing.T) {
		h := newHarness(t)
		in := createInput(t, expense1)
		in.Description = "   "

		got, err := h.uc.Create(t.Context(), in)

		require.NoError(t, err)
		assert.Empty(t, got.Expense.Description)
	})
	t.Run("normalized to NFC", func(t *testing.T) {
		h := newHarness(t)
		in := createInput(t, expense1)
		in.Description = "Café"

		got, err := h.uc.Create(t.Context(), in)

		require.NoError(t, err)
		assert.Equal(t, "Café", got.Expense.Description)
	})
	t.Run("longer than 200 characters", func(t *testing.T) {
		h := newHarness(t)
		in := createInput(t, expense1)
		in.Description = strings.Repeat("ñ", 201)

		_, err := h.uc.Create(t.Context(), in)

		requireAppErr(t, err, apperr.CodeValidationFailed, "description", "too_long")
	})
	t.Run("holding a bidi override", func(t *testing.T) {
		h := newHarness(t)
		in := createInput(t, expense1)
		in.Description = "Caf\u202ee"

		_, err := h.uc.Create(t.Context(), in)

		requireAppErr(t, err, apperr.CodeValidationFailed, "description", "invisible_characters")
	})
}

func TestCreate_BusinessRulesAre422(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*usecase.CreateInput)
		field  string
		detail string
	}{
		{"iva mismatch", func(in *usecase.CreateInput) { in.Draft.IVA = 70_00 }, "iva_cents", "iva_amount_mismatch"},
		{"US-W7: a física names a supplier", func(in *usecase.CreateInput) {
			in.Draft.Supplier = domain.Supplier{Kind: domain.SupplierPM}
		}, "supplier_kind", "supplier_not_applicable"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			in := createInput(t, expense1)
			c.mutate(&in)

			_, err := h.uc.Create(t.Context(), in)

			requireAppErr(t, err, apperr.CodeBusinessRuleViolation, c.field, c.detail)
			assert.Empty(t, h.access.reserved)
		})
	}
}

func TestCreate_CategoryRulesPerRegime(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(*usecase.CreateInput)
		deductible bool
		source     domain.DeductibleSource
		warnings   []string
		cashOver   bool        // stored cash_over_limit
		creditable money.Cents // stored creditable_iva_cents
	}{
		{"612 software by card", func(*usecase.CreateInput) {}, true, domain.SourceCategoryDefault, []string{}, false, 80_00},
		{"RESICO física never deducts", func(in *usecase.CreateInput) {
			in.Regime, in.Draft.PaidOn = fiscal.RegimeRESICO, date(t, "2025-06-10")
		}, false, domain.SourceCategoryDefault, []string{}, false, 80_00},
		{"606 blind 35 %: maintenance does not count", func(in *usecase.CreateInput) {
			in.Regime, in.Draft.Category = fiscal.RegimeRent, "mantenimiento"
		}, false, domain.SourceCategoryDefault, []string{}, false, 0},
		{"606 blind 35 %: predial still counts", func(in *usecase.CreateInput) {
			in.Regime, in.Draft.Category = fiscal.RegimeRent, "predial"
		}, true, domain.SourceCategoryDefault, []string{}, false, 80_00},
		{"cash over the limit", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.PaymentMethod = "papeleria", domain.PaymentCash
			in.Draft.Amount, in.Draft.IVA = 2_000_00, 320_00
		}, false, domain.SourceCategoryDefault, []string{"cash_over_threshold@payment_method"}, true, 0},
		{"fuel in cash, any amount", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.PaymentMethod = "combustible", domain.PaymentCash
			in.Draft.Amount, in.Draft.IVA = 300_00, 48_00
		}, false, domain.SourceCategoryDefault, []string{"cash_over_threshold@payment_method"}, true, 0},
		{"cash over the limit claimed deductible", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.PaymentMethod = "papeleria", domain.PaymentCash
			in.Draft.Amount, in.Draft.IVA = 2_000_00, 320_00
			in.Draft.Deductible = domain.ClaimDeductible
		}, true, domain.SourceUser, []string{"cash_over_threshold@payment_method", "iva_not_creditable_cash_payment@payment_method"}, true, 0},
		{"cash under the limit", func(in *usecase.CreateInput) {
			in.Draft.PaymentMethod = domain.PaymentCash
		}, true, domain.SourceCategoryDefault, []string{}, false, 80_00},
		{"no CFDI", func(in *usecase.CreateInput) { in.Draft.HasCFDI = false },
			false, domain.SourceCategoryDefault, []string{"no_cfdi@has_cfdi"}, false, 0},
		{"partial category overridden", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.Deductible = "restaurantes", domain.ClaimDeductible
		}, true, domain.SourceUser, []string{}, false, 80_00},
		{"not deductible category overridden", func(in *usecase.CreateInput) {
			in.Draft.Category, in.Draft.Deductible = "gastos_personales", domain.ClaimDeductible
		}, true, domain.SourceUser, []string{"category_not_deductible@deductible"}, false, 0},
		{"RESICO física overridden", func(in *usecase.CreateInput) {
			in.Regime, in.Draft.PaidOn = fiscal.RegimeRESICO, date(t, "2025-06-10")
			in.Draft.Deductible = domain.ClaimDeductible
		}, true, domain.SourceUser, []string{"not_isr_deductible_in_regime@deductible"}, false, 80_00},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			in := createInput(t, expense1)
			c.mutate(&in)

			got, err := h.uc.Create(t.Context(), in)

			require.NoError(t, err)
			assert.Equal(t, c.deductible, got.Expense.Deductible)
			assert.Equal(t, c.source, got.Expense.Source)
			assert.Equal(t, c.warnings, warningCodes(got.Expense.Warnings))
			stored, err := h.expenses.Find(ctxAs(t, userA), expense1, usecase.LockNone)
			require.NoError(t, err)
			assert.Equal(t, c.deductible, stored.Deductible)
			assert.Equal(t, c.source == domain.SourceUser, stored.Overridden)
			assert.Equal(t, c.cashOver, stored.CashOverLimit, "the assessed cash rule is stored")
			assert.Equal(t, c.creditable, stored.CreditableIVA, "the assessed creditable-IVA candidate is stored")
		})
	}
}

func TestCreate_ProvisionalParametersWarnOnPaidOn(t *testing.T) {
	h := newHarness(t)
	h.parameters.provisional = true
	h.wire()

	got, err := h.uc.Create(t.Context(), createInput(t, expense1))

	require.NoError(t, err)
	assert.Equal(t, []string{"provisional_parameters@paid_on"}, warningCodes(got.Expense.Warnings))
}

func TestCreate_DateWithoutParametersIsOutOfRange(t *testing.T) {
	h := newHarness(t)
	h.parameters.uncovered = true
	h.wire()

	_, err := h.uc.Create(t.Context(), createInput(t, expense1))

	requireAppErr(t, err, apperr.CodeValidationFailed, "paid_on", "out_of_range")
	assert.Empty(t, h.access.reserved)
}

func TestCreate_DuplicateFolioFiscalWarns(t *testing.T) {
	h := newHarness(t)
	first := createInput(t, expense1)
	first.Draft.CFDIUUID = uuidOf(t, strings.ToLower(cfdiUUID))
	_, err := h.uc.Create(t.Context(), first)
	require.NoError(t, err)
	second := createInput(t, expense2)
	second.Draft.CFDIUUID = uuidOf(t, cfdiUUID)

	got, err := h.uc.Create(t.Context(), second)

	require.NoError(t, err)
	assert.True(t, got.Created)
	assert.Equal(t, []string{"duplicate_cfdi_uuid@cfdi_uuid"}, warningCodes(got.Expense.Warnings))
	assert.False(t, logsMention(h, cfdiUUID), "the folio fiscal is never logged")
}

func TestCreate_InfrastructureFailureIsNotAClientError(t *testing.T) {
	h := newHarness(t)
	h.taxpayers.byID[taxpayerA].regimes = append(h.taxpayers.byID[taxpayerA].regimes, hold{code: "999", start: "2026-01"})
	in := createInput(t, expense1)
	in.Regime = "999"

	_, err := h.uc.Create(t.Context(), in)

	require.ErrorIs(t, err, errCategoryStore)
	_, isAppErr := err.(*apperr.Error) //nolint:errorlint // asserting the concrete type
	assert.False(t, isAppErr)
}

func fieldCodes(e *apperr.Error) []string {
	out := make([]string, 0, len(e.Fields))
	for _, f := range e.Fields {
		out = append(out, f.Field+":"+f.Code)
	}
	return out
}
