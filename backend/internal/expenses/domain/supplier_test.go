package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// feesDraft is US-W6 (WM WH-11): a persona moral pays a 612 consultant
// 40,000.00 + IVA 6,400.00 on 2026-06-15.
func feesDraft(t *testing.T) domain.Draft {
	t.Helper()
	d := softwareDraft(t)
	d.Category, d.Amount, d.IVA, d.PaidOn = domain.CategoryProfessionalFees, 40_000_00, 6_400_00, date(t, "2026-06-15")
	d.PaymentMethod = domain.PaymentTransfer
	d.Supplier = domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
		Operation: fiscal.OperationProfessionalServices}
	return d
}

// moral is a persona moral rule row: it carries the E3 default operation
// (category_regime_rules.default_operation_kind, MODULE-withholdings §2.2).
func moral(r domain.CategoryRule, op fiscal.OperationKind) domain.CategoryRule {
	r.DefaultOperation = op
	return r
}

var (
	feesMoral     = moral(feesBusiness, fiscal.OperationProfessionalServices)
	freightMoral  = moral(rule(domain.CategoryFreight, domain.Deductible, true), fiscal.OperationGroundFreight)
	rentMoral     = moral(rule(domain.CategoryRent, domain.Deductible, true), fiscal.OperationRentRealEstate)
	softwareMoral = moral(softwareBusiness, fiscal.OperationBusinessServices)
)

func TestAssessUSW6PersonaMoralWithholdsFromAConsultant(t *testing.T) {
	got, err := assess(t, feesDraft(t), pm601, feesMoral)
	require.NoError(t, err)

	assert.Equal(t, fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67},
		ISRRate:      rate(t, "0.10"), IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
		RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6}, TableVersion: fiscal.WithholdingTableVersion,
	}, got.Withholding)
	assert.Equal(t, domain.Totals{Subtotal: 40_000_00, IVA: 6_400_00, Total: 46_400_00, Withheld: 8_266_67, Net: 38_133_33}, got.Totals)
	assert.Equal(t, money.Cents(2_133_33), got.CreditableIVA, "creditable share of the IVA paid")
	assert.Equal(t, money.Cents(4_266_67), got.CreditableIVAWithheld, "credited the month after remittance")
	assert.Equal(t, "2026-07-17", got.RemitDueOn.String())
	assert.Equal(t, []domain.Warning{{
		Code: domain.WarningDeductionRequiresRemittance, Field: domain.FieldWithholding, DueOn: date(t, "2026-07-17"),
	}}, got.Warnings)
	assert.Equal(t, feesDraft(t).Supplier, got.Supplier)
}

func TestAssessSupplierWithholdings(t *testing.T) {
	tests := []struct {
		name            string
		rule            domain.CategoryRule
		edit            func(*domain.Draft)
		wantSupplier    domain.Supplier
		wantWithheld    fiscal.Withholdings
		wantRules       []fiscal.RuleID
		wantCPA         bool
		wantRemit       string
		wantCreditable  [2]money.Cents
		wantWarningCode []domain.WarningCode
	}{
		{
			name: "\"No sé\" derives as 612 and warns", rule: feesMoral,
			edit: func(d *domain.Draft) { d.Supplier.Regime = "" },
			wantSupplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegimeUnknown,
				Operation: fiscal.OperationProfessionalServices},
			wantWithheld: fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67}, wantRules: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6},
			wantRemit: "2026-07-17", wantCreditable: [2]money.Cents{2_133_33, 4_266_67},
			wantWarningCode: []domain.WarningCode{domain.WarningSupplierRegimeUnknown, domain.WarningDeductionRequiresRemittance},
		},
		{
			name: "RESICO supplier: 1.25 %", rule: feesMoral,
			edit: func(d *domain.Draft) { d.Supplier.Regime = domain.SupplierRegime626 },
			wantSupplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime626,
				Operation: fiscal.OperationProfessionalServices},
			wantWithheld: fiscal.Withholdings{ISR: 500_00, IVA: 4_266_67}, wantRules: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6},
			wantRemit: "2026-07-17", wantCreditable: [2]money.Cents{2_133_33, 4_266_67},
			wantWarningCode: []domain.WarningCode{domain.WarningDeductionRequiresRemittance},
		},
		{
			name: "WH-06: a persona moral carrier, operation from the category", rule: freightMoral,
			edit: func(d *domain.Draft) {
				d.Category, d.Amount, d.IVA = domain.CategoryFreight, 10_000_00, 1_600_00
				d.Supplier = domain.Supplier{Kind: domain.SupplierPM}
			},
			wantSupplier: domain.Supplier{Kind: domain.SupplierPM, Operation: fiscal.OperationGroundFreight},
			wantWithheld: fiscal.Withholdings{IVA: 400_00}, wantRules: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA4},
			wantRemit: "2026-07-17", wantCreditable: [2]money.Cents{1_200_00, 400_00},
			wantWarningCode: []domain.WarningCode{domain.WarningDeductionRequiresRemittance},
		},
		{
			name: "a persona moral supplier of software: nothing withheld", rule: softwareMoral,
			edit: func(d *domain.Draft) {
				d.Category, d.Amount, d.IVA = "software", 500_00, 80_00
				d.Supplier = domain.Supplier{Kind: domain.SupplierPM}
			},
			wantSupplier: domain.Supplier{Kind: domain.SupplierPM, Operation: fiscal.OperationBusinessServices},
			wantRules:    []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA8}, wantCreditable: [2]money.Cents{80_00, 0},
		},
		{
			name: "foreign supplier: zero and warned", rule: feesMoral,
			edit:         func(d *domain.Draft) { d.Supplier = domain.Supplier{Kind: domain.SupplierForeign} },
			wantSupplier: domain.Supplier{Kind: domain.SupplierForeign, Operation: fiscal.OperationProfessionalServices},
			wantRules:    []fiscal.RuleID{fiscal.RuleISR1, fiscal.RuleIVA0}, wantCreditable: [2]money.Cents{6_400_00, 0},
			wantWarningCode: []domain.WarningCode{domain.WarningForeignSupplier},
		},
		{
			name: "commission from a persona física: CPA unverified", rule: feesMoral,
			edit: func(d *domain.Draft) { d.Supplier.Operation = fiscal.OperationCommission },
			wantSupplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
				Operation: fiscal.OperationCommission},
			wantWithheld: fiscal.Withholdings{IVA: 4_266_67}, wantRules: []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA6},
			wantCPA: true, wantRemit: "2026-07-17", wantCreditable: [2]money.Cents{2_133_33, 4_266_67},
			wantWarningCode: []domain.WarningCode{domain.WarningCPAUnverified, domain.WarningDeductionRequiresRemittance},
		},
		{
			name: "housing rent: exempt, ISR only", rule: rentMoral,
			edit: func(d *domain.Draft) {
				d.Category, d.IVARate, d.IVA = domain.CategoryRent, fiscal.IVAExempt, 0
				d.Supplier = domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime606, ResidentialRent: true}
			},
			wantSupplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime606,
				Operation: fiscal.OperationRentRealEstate, ResidentialRent: true},
			wantWithheld: fiscal.Withholdings{ISR: 4_000_00}, wantRules: []fiscal.RuleID{fiscal.RuleISR6, fiscal.RuleIVANoTax},
			wantRemit: "2026-07-17", wantWarningCode: []domain.WarningCode{domain.WarningDeductionRequiresRemittance},
		},
		{
			name: "paid in December: remit by January 17 of the next year", rule: feesMoral,
			edit: func(d *domain.Draft) { d.PaidOn = date(t, "2025-12-31") },
			wantSupplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
				Operation: fiscal.OperationProfessionalServices},
			wantWithheld: fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67}, wantRules: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6},
			wantRemit: "2026-01-17", wantCreditable: [2]money.Cents{2_133_33, 4_266_67},
			wantWarningCode: []domain.WarningCode{domain.WarningDeductionRequiresRemittance},
		},
		{
			name: "no CFDI: nothing creditable, the withholding stands", rule: feesMoral,
			edit: func(d *domain.Draft) { d.HasCFDI = false },
			wantSupplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
				Operation: fiscal.OperationProfessionalServices},
			wantWithheld: fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67}, wantRules: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6},
			wantRemit:       "2026-07-17",
			wantWarningCode: []domain.WarningCode{domain.WarningNoCFDI, domain.WarningDeductionRequiresRemittance},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := feesDraft(t)
			tt.edit(&d)
			got, err := domain.Assess(d, pm601, tt.rule, date(t, today), catalog2026(t), withholdings2026(t))
			require.NoError(t, err)
			assert.Equal(t, tt.wantSupplier, got.Supplier, "supplier")
			assert.Equal(t, tt.wantWithheld, got.Withholding.Withholdings, "withheld")
			assert.Equal(t, tt.wantRules, got.Withholding.RuleIDs, "rules")
			assert.Equal(t, tt.wantCPA, got.Withholding.CPAUnverified, "cpa")
			assert.Equal(t, tt.wantRemit, got.RemitDueOn.String(), "remit due")
			assert.Equal(t, tt.wantCreditable, [2]money.Cents{got.CreditableIVA, got.CreditableIVAWithheld}, "creditable")
			codes := make([]domain.WarningCode, 0, len(got.Warnings))
			for _, w := range got.Warnings {
				codes = append(codes, w.Code)
			}
			if tt.wantWarningCode == nil {
				tt.wantWarningCode = []domain.WarningCode{}
			}
			assert.Equal(t, tt.wantWarningCode, codes, "warnings")
		})
	}
}

// A persona física never withholds from a domestic supplier (WM F1, F2):
// nothing is asked, nothing stored, a zero derivation.
func TestAssessPersonaFisicaWithholdsNothing(t *testing.T) {
	got, err := assess(t, softwareDraft(t), pf612, softwareBusiness)
	require.NoError(t, err)

	assert.Equal(t, domain.Supplier{}, got.Supplier)
	assert.Equal(t, fiscal.WithholdingRecord{
		IVABasis: fiscal.IVABasisNone, Withholder: fiscal.WithholderNone,
		RuleIDs: []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA1}, TableVersion: fiscal.WithholdingTableVersion,
	}, got.Withholding)
	assert.True(t, got.RemitDueOn.IsZero())
}

func TestAssessRejectsSupplierFacts(t *testing.T) {
	tests := []struct {
		name  string
		payer domain.Payer
		edit  func(*domain.Draft)
		want  fiscal.Violations
	}{
		{"US-W7: a persona física names a supplier", pf612, func(d *domain.Draft) {
			d.Supplier = domain.Supplier{Kind: domain.SupplierPM}
		}, violations(domain.FieldSupplierKind, domain.ErrSupplierNotApplicable)},
		{"a persona física sends every supplier fact", pf612, func(d *domain.Draft) {
			d.Supplier = domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
				Operation: fiscal.OperationRentRealEstate, ResidentialRent: true}
			d.IVARate, d.IVA = fiscal.IVAExempt, 0
		}, fiscal.Violations{
			{Field: domain.FieldSupplierKind, Err: domain.ErrSupplierNotApplicable},
			{Field: domain.FieldSupplierRegime, Err: domain.ErrSupplierNotApplicable},
			{Field: domain.FieldOperationKind, Err: domain.ErrSupplierNotApplicable},
			{Field: domain.FieldResidentialRent, Err: domain.ErrSupplierNotApplicable},
		}},
		{"a regime for a persona moral supplier", pm601, func(d *domain.Draft) {
			d.Supplier = domain.Supplier{Kind: domain.SupplierPM, Regime: domain.SupplierRegime612}
		}, violations(domain.FieldSupplierRegime, domain.ErrSupplierNotApplicable)},
		{"housing rent with IVA", pm601, func(d *domain.Draft) {
			d.Supplier.ResidentialRent = true
		}, violations(domain.FieldIVARate, domain.ErrResidentialRentExempt)},
		{"a persona moral without a supplier", pm601, func(d *domain.Draft) {
			d.Supplier = domain.Supplier{}
		}, violations(domain.FieldSupplierKind, fiscal.ErrRequired)},
		{"unknown supplier kind", pm601, func(d *domain.Draft) {
			d.Supplier.Kind = "trust"
		}, violations(domain.FieldSupplierKind, fiscal.ErrInvalid)},
		{"unknown supplier regime", pm601, func(d *domain.Draft) {
			d.Supplier.Regime = "601"
		}, violations(domain.FieldSupplierRegime, fiscal.ErrInvalid)},
		{"platform is not an expense operation", pm601, func(d *domain.Draft) {
			d.Supplier.Operation = fiscal.OperationPlatform
		}, violations(domain.FieldOperationKind, fiscal.ErrInvalid)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := feesDraft(t)
			tt.edit(&d)
			_, err := domain.Assess(d, tt.payer, feesMoral, date(t, today), catalog2026(t), withholdings2026(t))
			got, ok := errors.AsType[fiscal.Violations](err)
			require.True(t, ok, "want Violations, got %v", err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemitDueOn(t *testing.T) {
	for paid, want := range map[string]string{
		"2026-06-15": "2026-07-17", "2026-06-30": "2026-07-17", "2026-01-31": "2026-02-17", "2026-12-01": "2027-01-17",
	} {
		assert.Equal(t, want, domain.RemitDueOn(date(t, paid)).String(), paid)
	}
}

func TestSupplierEnums(t *testing.T) {
	for _, k := range []domain.SupplierKind{domain.SupplierPM, domain.SupplierPFWithActivity, domain.SupplierForeign} {
		assert.True(t, k.Valid(), k)
	}
	assert.False(t, domain.SupplierKind("persona_moral").Valid())
	for _, r := range []domain.SupplierRegime{
		domain.SupplierRegime612, domain.SupplierRegime606, domain.SupplierRegime626,
		domain.SupplierRegime625, domain.SupplierRegime621, domain.SupplierRegimeUnknown,
	} {
		assert.True(t, r.Valid(), r)
	}
	assert.False(t, domain.SupplierRegime("601").Valid())
	got, err := domain.ParseSupplierKind("pm")
	require.NoError(t, err)
	assert.Equal(t, domain.SupplierPM, got)
	_, err = domain.ParseSupplierRegime("x")
	assert.ErrorIs(t, err, fiscal.ErrInvalid)
}
