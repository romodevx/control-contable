package usecase_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// feesInput is US-W6 (WM WH-11): userA's persona moral 601 pays a 612
// consultant 40,000.00 + IVA 6,400.00 on 2026-06-15; the operation comes
// from the category.
func feesInput(t *testing.T, id string) usecase.CreateInput {
	t.Helper()
	in := createInput(t, id)
	in.TaxpayerID, in.Regime = taxpayerM, "601"
	in.Draft.Category, in.Draft.Amount, in.Draft.IVA = "servicios_profesionales", 40_000_00, 6_400_00
	in.Draft.PaidOn, in.Draft.PaymentMethod = date(t, "2026-06-15"), domain.PaymentTransfer
	in.Draft.Supplier = domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612}
	return in
}

func usW6Record(t *testing.T) fiscal.WithholdingRecord {
	t.Helper()
	r, err := money.ParseRate("0.10")
	require.NoError(t, err)
	return fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67}, ISRRate: r,
		IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
		RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6}, TableVersion: fiscal.WithholdingTableVersion,
	}
}

func TestCreate_USW6StoresTheDerivedWithholding(t *testing.T) {
	h := newHarness(t)

	got, err := h.uc.Create(t.Context(), feesInput(t, expense1))

	require.NoError(t, err)
	stored := h.expenses.rows[expense1].e
	assert.Equal(t, usW6Record(t), stored.Withholding)
	assert.Equal(t, domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
		Operation: fiscal.OperationProfessionalServices}, stored.Supplier)
	assert.Equal(t, money.Cents(2_133_33), stored.CreditableIVA, "creditable share of the IVA paid")
	assert.Equal(t, money.Cents(4_266_67), stored.CreditableIVAWithheld, "creditable share of the IVA withheld")
	assert.Equal(t, usW6Record(t), got.Expense.Withholding)
	assert.Equal(t, "2026-07-17", got.Expense.RemitDueOn.String())
	assert.Equal(t, []string{"deduction_requires_remittance@withholding"}, warningCodes(got.Expense.Warnings))
}

func TestCreate_SupplierRules(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*usecase.CreateInput)
		code   apperr.Code
		field  string
		detail string
	}{
		{"a persona moral must name its supplier", func(in *usecase.CreateInput) { in.Draft.Supplier = domain.Supplier{} },
			apperr.CodeValidationFailed, "supplier_kind", "required"},
		{"a regime of a persona moral supplier", func(in *usecase.CreateInput) { in.Draft.Supplier.Kind = domain.SupplierPM },
			apperr.CodeBusinessRuleViolation, "supplier_regime", "supplier_not_applicable"},
		{"housing rent with IVA", func(in *usecase.CreateInput) { in.Draft.Supplier.ResidentialRent = true },
			apperr.CodeBusinessRuleViolation, "iva_rate", "residential_rent_exempt"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			in := feesInput(t, expense1)
			c.mutate(&in)

			_, err := h.uc.Create(t.Context(), in)

			requireAppErr(t, err, c.code, c.field, c.detail)
			assert.Empty(t, h.expenses.rows)
		})
	}
}

// The idempotent replay compares the facts as sent: an omitted operation
// matches the category default the first create stored, and an omitted
// regime of a persona física supplier matches "No sé".
func TestCreate_ReplayComparesTheSupplierFacts(t *testing.T) {
	h := newHarness(t)
	in := feesInput(t, expense1)
	in.Draft.Supplier.Regime = ""
	_, err := h.uc.Create(t.Context(), in)
	require.NoError(t, err)

	replay, err := h.uc.Create(t.Context(), in)
	require.NoError(t, err)
	assert.False(t, replay.Created)

	explicit := in
	explicit.Draft.Supplier.Operation, explicit.Draft.Supplier.Regime =
		fiscal.OperationProfessionalServices, domain.SupplierRegimeUnknown
	replay, err = h.uc.Create(t.Context(), explicit)
	require.NoError(t, err)
	assert.False(t, replay.Created, "the defaults, sent explicitly, are the same request")

	other := in
	other.Draft.Supplier.Regime = domain.SupplierRegime626
	_, err = h.uc.Create(t.Context(), other)
	requireAppErr(t, err, apperr.CodeConflict, "id", "id_reused")
}

func TestUpdate_ReDerivesAndClearsTheRegimeOffAPersonaFisica(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), feesInput(t, expense1))
	require.NoError(t, err)
	kind := domain.SupplierPM

	got, err := h.uc.Update(t.Context(), usecase.UpdateInput{
		UserID: userA, TaxpayerID: taxpayerM, ExpenseID: expense1, Patch: usecase.Patch{SupplierKind: &kind},
	})

	require.NoError(t, err)
	stored := h.expenses.rows[expense1].e
	assert.Equal(t, domain.Supplier{Kind: domain.SupplierPM, Operation: fiscal.OperationProfessionalServices}, stored.Supplier)
	assert.Equal(t, []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA8}, stored.Withholding.RuleIDs)
	assert.Equal(t, fiscal.Withholdings{}, got.Withholding.Withholdings)
	assert.Equal(t, money.Cents(6_400_00), stored.CreditableIVA)
	assert.Zero(t, stored.CreditableIVAWithheld)
	assert.True(t, got.RemitDueOn.IsZero())
}

func TestUpdate_SupplierFacts(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), feesInput(t, expense1))
	require.NoError(t, err)
	regime, operation, residential := domain.SupplierRegime626, fiscal.OperationCommission, false

	got, err := h.uc.Update(t.Context(), usecase.UpdateInput{
		UserID: userA, TaxpayerID: taxpayerM, ExpenseID: expense1,
		Patch: usecase.Patch{SupplierRegime: &regime, OperationKind: &operation, ResidentialRent: &residential},
	})

	require.NoError(t, err)
	assert.Equal(t, domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: regime, Operation: operation}, got.Supplier)
	assert.Equal(t, []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6}, got.Withholding.RuleIDs)
}

// Restore re-derives with the table in force: what it stores is never older
// than the restore.
func TestRestore_ReDerivesTheWithholding(t *testing.T) {
	h := newHarness(t)
	_, err := h.uc.Create(t.Context(), feesInput(t, expense1))
	require.NoError(t, err)
	require.NoError(t, h.uc.Delete(ctxAs(t, userA), userA, taxpayerM, expense1))
	row := h.expenses.rows[expense1]
	row.e.Withholding = fiscal.WithholdingRecord{}
	row.e.CreditableIVA, row.e.CreditableIVAWithheld = 6_400_00, 0
	h.expenses.rows[expense1] = row

	got, err := h.uc.Restore(t.Context(), userA, taxpayerM, expense1)

	require.NoError(t, err)
	assert.Equal(t, usW6Record(t), h.expenses.rows[expense1].e.Withholding)
	assert.Equal(t, money.Cents(4_266_67), h.expenses.rows[expense1].e.CreditableIVAWithheld)
	assert.False(t, h.expenses.rows[expense1].e.Deleted)
	assert.Equal(t, []string{"deduction_requires_remittance@withholding"}, warningCodes(got.Warnings))
}

// previewInput is US-W6 as previewExpenseWithholdings receives it.
func previewInput(t *testing.T) usecase.PreviewInput {
	t.Helper()
	return usecase.PreviewInput{
		UserID: userA, TaxpayerID: taxpayerM, Regime: "601", On: "2026-06-15", Category: "servicios_profesionales",
		Amount: 40_000_00, IVARate: fiscal.IVARate16, IVA: 6_400_00,
		Supplier: domain.Supplier{Kind: domain.SupplierPFWithActivity},
	}
}

func TestPreview_DerivesWithoutStoring(t *testing.T) {
	h := newHarness(t)
	h.access.readOnly = true

	got, err := h.uc.Preview(t.Context(), previewInput(t))

	require.NoError(t, err)
	assert.Equal(t, usW6Record(t), got.Withholding)
	assert.Equal(t, "2026-07-17", got.RemitDueOn.String())
	assert.Equal(t, []string{"supplier_regime_unknown@supplier_regime", "deduction_requires_remittance@withholding"},
		warningCodes(got.Warnings))
	assert.Equal(t, money.Cents(46_400_00), got.Subtotal+got.IVA)
	assert.Empty(t, h.expenses.rows)
	assert.Equal(t, []string{"ro"}, h.tx.modes)
	assert.Zero(t, h.access.writable, "a read: no access check")
}

func TestPreview_ArchivedTaxpayerAndDefaultDate(t *testing.T) {
	h := newHarness(t)
	h.taxpayers.byID[taxpayerM].archived = true
	in := previewInput(t)
	in.On = ""

	got, err := h.uc.Preview(t.Context(), in)

	require.NoError(t, err)
	assert.Equal(t, "2026-11-17", got.RemitDueOn.String(), "on defaults to today, 2026-10-13")
}

// Only withholding warnings: a preview has no CFDI, payment method or folio.
func TestPreview_PersonaFisicaGetsAZeroDerivation(t *testing.T) {
	h := newHarness(t)
	in := previewInput(t)
	in.TaxpayerID, in.Regime, in.Supplier = taxpayerA, fiscal.RegimeBusiness, domain.Supplier{}

	got, err := h.uc.Preview(t.Context(), in)

	require.NoError(t, err)
	assert.Equal(t, fiscal.Withholdings{}, got.Withholding.Withholdings)
	assert.Equal(t, fiscal.WithholderNone, got.Withholding.Withholder)
	assert.Empty(t, got.Warnings)
}

func TestPreview_Refusals(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*usecase.PreviewInput)
		code   apperr.Code
		field  string
		detail string
	}{
		{"unknown taxpayer", func(in *usecase.PreviewInput) { in.TaxpayerID = taxpayerB }, apperr.CodeNotFound, "", ""},
		{"malformed on", func(in *usecase.PreviewInput) { in.On = "15/06/2026" }, apperr.CodeValidationFailed, "on", "invalid"},
		{"on in the future", func(in *usecase.PreviewInput) { in.On = "2026-10-14" }, apperr.CodeValidationFailed, "on", "out_of_range"},
		{"on before 2022", func(in *usecase.PreviewInput) { in.On = "2021-12-31" }, apperr.CodeValidationFailed, "on", "out_of_range"},
		{"regime not held on the day", func(in *usecase.PreviewInput) { in.Regime = "626" },
			apperr.CodeBusinessRuleViolation, "regime", "regime_not_active"},
		{"unknown category", func(in *usecase.PreviewInput) { in.Category = "yates" },
			apperr.CodeValidationFailed, "category_code", "unknown_category"},
		{"a persona moral without a supplier", func(in *usecase.PreviewInput) { in.Supplier = domain.Supplier{} },
			apperr.CodeValidationFailed, "supplier_kind", "required"},
		{"a persona física names a supplier", func(in *usecase.PreviewInput) {
			in.TaxpayerID, in.Regime = taxpayerA, fiscal.RegimeBusiness
		}, apperr.CodeBusinessRuleViolation, "supplier_kind", "supplier_not_applicable"},
		{"IVA mismatch", func(in *usecase.PreviewInput) { in.IVA = 1_00 },
			apperr.CodeBusinessRuleViolation, "iva_cents", "iva_amount_mismatch"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			in := previewInput(t)
			c.mutate(&in)

			_, err := h.uc.Preview(t.Context(), in)

			if c.field == "" {
				requireAppErr(t, err, c.code)
				return
			}
			requireAppErr(t, err, c.code, c.field, c.detail)
		})
	}
}

func TestPreview_UncoveredDateIsOutOfRange(t *testing.T) {
	h := newHarness(t)
	h.parameters.uncovered = true
	h.wire()

	_, err := h.uc.Preview(t.Context(), previewInput(t))

	requireAppErr(t, err, apperr.CodeValidationFailed, "on", "out_of_range")
}

func TestCategories_CarryTheDefaultOperation(t *testing.T) {
	h := newHarness(t)

	moral, err := h.uc.Categories(t.Context(), usecase.CategoriesInput{UserID: userA, TaxpayerID: taxpayerM, Regime: "601"})
	require.NoError(t, err)
	fisica, err := h.uc.Categories(t.Context(), usecase.CategoriesInput{UserID: userA, TaxpayerID: taxpayerA, Regime: "612"})
	require.NoError(t, err)

	for _, c := range moral.Items {
		if c.Code == "servicios_profesionales" {
			assert.Equal(t, fiscal.OperationProfessionalServices, c.DefaultOperation)
		}
		assert.NotEmpty(t, c.DefaultOperation, c.Code)
	}
	for _, c := range fisica.Items {
		assert.Empty(t, c.DefaultOperation, "a persona física row has none: %s", c.Code)
	}
}
