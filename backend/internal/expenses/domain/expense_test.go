package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

const today = "2026-10-13"

// softwareDraft is US-E1: software 500.00 + IVA 80.00 paid by debit card,
// with a CFDI.
func softwareDraft(t *testing.T) domain.Draft {
	t.Helper()
	return domain.Draft{
		Category:      "software",
		Amount:        500_00,
		IVARate:       domain.IVARate16,
		IVA:           80_00,
		PaymentMethod: domain.PaymentCardDebit,
		PaidOn:        date(t, "2026-09-12"),
		HasCFDI:       true,
	}
}

func assess(t *testing.T, d domain.Draft, p domain.Payer, r domain.CategoryRule) (domain.Assessment, error) {
	t.Helper()
	return domain.Assess(d, p, r, date(t, today), catalog2026(t))
}

func TestAssessUS1Software(t *testing.T) {
	got, err := assess(t, softwareDraft(t), pf612, softwareBusiness)
	require.NoError(t, err)

	assert.Equal(t, domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}, got.Hint)
	assert.True(t, got.Deductible)
	assert.Equal(t, domain.SourceCategoryDefault, got.Source)
	assert.False(t, got.CashOverLimit)
	assert.Equal(t, money.Cents(80_00), got.CreditableIVA)
	assert.Nil(t, got.Warnings)
	assert.Equal(t, domain.Suggested{}, got.Suggested)
	assert.Equal(t, domain.Totals{Subtotal: 500_00, IVA: 80_00, Total: 580_00, Net: 580_00}, got.Totals)
	assert.Equal(t, "2026-09-12", got.BookedOn.String())
}

// TestAssessDeductible is §2.4 and §2.5 through the US-E1 stories.
func TestAssessDeductible(t *testing.T) {
	cashOver := domain.Warning{Code: domain.WarningCashOverThreshold, Field: domain.FieldPaymentMethod}
	noCFDI := domain.Warning{Code: domain.WarningNoCFDI, Field: domain.FieldHasCFDI}
	tests := []struct {
		name           string
		payer          domain.Payer
		rule           domain.CategoryRule
		edit           func(*domain.Draft)
		wantDeductible bool
		wantSource     domain.DeductibleSource
		wantCashOver   bool
		wantCreditable money.Cents
		wantWarnings   []domain.Warning
	}{
		{
			name: "US-E1: RESICO PF stores IVA but no deduction", payer: pf626, rule: softwareRESICOPF,
			edit:       func(*domain.Draft) {},
			wantSource: domain.SourceCategoryDefault, wantCreditable: 80_00,
		},
		{
			name: "PF-626-IVA-01: RESICO PF IVA 2,000 is creditable", payer: pf626, rule: softwareRESICOPF,
			edit:       func(d *domain.Draft) { d.Amount, d.IVA = 12_500_00, 2_000_00 },
			wantSource: domain.SourceCategoryDefault, wantCreditable: 2_000_00,
		},
		{
			name: "US-E1: 606 blind 35 %, maintenance", payer: pf606Blind, rule: mantenimiento606,
			edit:       func(d *domain.Draft) { d.Category = "mantenimiento" },
			wantSource: domain.SourceCategoryDefault,
		},
		{
			name: "US-E1: 606 blind 35 %, predial", payer: pf606Blind, rule: predial606,
			edit:           func(d *domain.Draft) { d.Category = domain.CategoryPredial; d.IVARate, d.IVA = domain.IVAExempt, 0 },
			wantDeductible: true, wantSource: domain.SourceCategoryDefault,
		},
		{
			name: "US-E1: 2,320.00 in cash for papelería", payer: pf612, rule: papeleriaBusiness,
			edit: func(d *domain.Draft) {
				d.Category, d.PaymentMethod = "papeleria", domain.PaymentCash
				d.Amount, d.IVA = 2_000_00, 320_00
			},
			wantSource: domain.SourceCategoryDefault, wantCashOver: true, wantWarnings: []domain.Warning{cashOver},
		},
		{
			name: "exactly $2,000.00 in cash is not over the limit", payer: pf612, rule: papeleriaBusiness,
			edit: func(d *domain.Draft) {
				d.Category, d.PaymentMethod = "papeleria", domain.PaymentCash
				d.Amount, d.IVA = 1_724_14, 275_86
			},
			wantDeductible: true, wantSource: domain.SourceCategoryDefault, wantCreditable: 275_86,
		},
		{
			name: "US-E1: 300.00 of fuel in cash", payer: pf612, rule: fuelBusiness,
			edit: func(d *domain.Draft) {
				d.Category, d.PaymentMethod = "combustible", domain.PaymentCash
				d.Amount, d.IVA = 258_62, 41_38
			},
			wantSource: domain.SourceCategoryDefault, wantCashOver: true, wantWarnings: []domain.Warning{cashOver},
		},
		{
			name: "fuel by card, any amount", payer: pf612, rule: fuelBusiness,
			edit:           func(d *domain.Draft) { d.Category, d.Amount, d.IVA = "combustible", 5_000_00, 800_00 },
			wantDeductible: true, wantSource: domain.SourceCategoryDefault, wantCreditable: 800_00,
		},
		{
			name: "RESICO PF in cash over the limit: IVA not creditable", payer: pf626, rule: softwareRESICOPF,
			edit: func(d *domain.Draft) {
				d.PaymentMethod = domain.PaymentCash
				d.Amount, d.IVA = 5_000_00, 800_00
			},
			wantSource: domain.SourceCategoryDefault, wantCashOver: true, wantWarnings: []domain.Warning{cashOver},
		},
		{
			name: "US-E1: no CFDI", payer: pf612, rule: softwareBusiness,
			edit:       func(d *domain.Draft) { d.HasCFDI = false },
			wantSource: domain.SourceCategoryDefault, wantWarnings: []domain.Warning{noCFDI},
		},
		{
			name: "US-E1: switch on for restaurants", payer: pf612, rule: diningBusiness,
			edit:           func(d *domain.Draft) { d.Category, d.Deductible = diningCategory, domain.ClaimDeductible },
			wantDeductible: true, wantSource: domain.SourceUser, wantCreditable: 80_00,
		},
		{
			name: "US-E1: switch on for personal expenses", payer: pf612, rule: personalBusiness,
			edit:           func(d *domain.Draft) { d.Category, d.Deductible = "gastos_personales", domain.ClaimDeductible },
			wantDeductible: true, wantSource: domain.SourceUser, wantCreditable: 80_00,
			wantWarnings: []domain.Warning{{Code: domain.WarningCategoryNotDeductible, Field: domain.FieldDeductible}},
		},
		{
			name: "switch on in RESICO PF", payer: pf626, rule: softwareRESICOPF,
			edit:           func(d *domain.Draft) { d.Deductible = domain.ClaimDeductible },
			wantDeductible: true, wantSource: domain.SourceUser, wantCreditable: 80_00,
			wantWarnings: []domain.Warning{{Code: domain.WarningNotISRDeductibleInRegime, Field: domain.FieldDeductible}},
		},
		{
			name: "switch off for software", payer: pf612, rule: softwareBusiness,
			edit:       func(d *domain.Draft) { d.Deductible = domain.ClaimNotDeductible },
			wantSource: domain.SourceUser,
		},
		{
			name: "the override wins over the cash rule, the warning stays", payer: pf612, rule: papeleriaBusiness,
			edit: func(d *domain.Draft) {
				d.Category, d.PaymentMethod, d.Deductible = "papeleria", domain.PaymentCash, domain.ClaimDeductible
				d.Amount, d.IVA = 2_000_00, 320_00
			},
			wantDeductible: true, wantSource: domain.SourceUser, wantCashOver: true, wantCreditable: 320_00,
			wantWarnings: []domain.Warning{cashOver},
		},
		{
			name: "no CFDI in cash over the limit: both warnings", payer: pf612, rule: papeleriaBusiness,
			edit: func(d *domain.Draft) {
				d.Category, d.PaymentMethod, d.HasCFDI = "papeleria", domain.PaymentCash, false
				d.Amount, d.IVA = 2_000_00, 320_00
			},
			wantSource: domain.SourceCategoryDefault, wantCashOver: true, wantWarnings: []domain.Warning{cashOver, noCFDI},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := softwareDraft(t)
			tt.edit(&d)
			got, err := assess(t, d, tt.payer, tt.rule)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDeductible, got.Deductible, "deductible")
			assert.Equal(t, tt.wantSource, got.Source, "source")
			assert.Equal(t, tt.wantCashOver, got.CashOverLimit, "cash over limit")
			assert.Equal(t, tt.wantCreditable, got.CreditableIVA, "creditable IVA")
			assert.Equal(t, tt.wantWarnings, got.Warnings, "warnings")
		})
	}
}

// TestAssessWithholdingsMade: US-E1 and PM-601-IVA-01, a persona moral
// withholding from a PF lawyer.
func TestAssessWithholdingsMade(t *testing.T) {
	tests := []struct {
		name         string
		amount, iva  money.Cents
		withheld     domain.Withholdings
		wantTotals   domain.Totals
		wantSuggests domain.Suggested
	}{
		{
			name: "US-E1: fees 10,000 + 1,600", amount: 10_000_00, iva: 1_600_00,
			withheld:   domain.Withholdings{ISR: 1_000_00, IVA: 1_066_67},
			wantTotals: domain.Totals{Subtotal: 10_000_00, IVA: 1_600_00, Total: 11_600_00, Withheld: 2_066_67, Net: 9_533_33},
			wantSuggests: domain.Suggested{
				ISR: domain.Suggestion{Cents: 1_000_00, Known: true}, IVA: domain.Suggestion{Cents: 1_066_67, Known: true},
			},
		},
		{
			name: "PM-601-IVA-01: fees 50,000 + 8,000", amount: 50_000_00, iva: 8_000_00,
			withheld:   domain.Withholdings{ISR: 5_000_00, IVA: 5_333_33},
			wantTotals: domain.Totals{Subtotal: 50_000_00, IVA: 8_000_00, Total: 58_000_00, Withheld: 10_333_33, Net: 47_666_67},
			wantSuggests: domain.Suggested{
				ISR: domain.Suggestion{Cents: 5_000_00, Known: true}, IVA: domain.Suggestion{Cents: 5_333_33, Known: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := softwareDraft(t)
			d.Category, d.Amount, d.IVA, d.Withheld = domain.CategoryProfessionalFees, tt.amount, tt.iva, tt.withheld
			got, err := assess(t, d, pm601, feesBusiness)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTotals, got.Totals)
			assert.Equal(t, tt.wantSuggests, got.Suggested)
			assert.Nil(t, got.Warnings, "suggestions are UX only: no warning")
		})
	}
}

func TestAssessRejects(t *testing.T) {
	someUUID, err := domain.ParseCFDIUUID("6F1C2A3B-4D5E-4F60-8A9B-0C1D2E3F4A5B")
	require.NoError(t, err)
	tests := []struct {
		name      string
		payer     domain.Payer
		rule      domain.CategoryRule
		edit      func(*domain.Draft)
		wantClass domain.Class
		want      domain.Violations
	}{
		{"amount 0", pf612, softwareBusiness, func(d *domain.Draft) { d.Amount = 0 }, domain.ClassValidation,
			violations(domain.FieldAmountCents, domain.ErrMustBePositive)},
		{"amount above the maximum", pf612, softwareBusiness, func(d *domain.Draft) { d.Amount = domain.MaxCents + 1 }, domain.ClassValidation,
			violations(domain.FieldAmountCents, domain.ErrOutOfRange)},
		{"negative IVA", pf612, softwareBusiness, func(d *domain.Draft) { d.IVA = -1 }, domain.ClassValidation,
			violations(domain.FieldIVACents, domain.ErrOutOfRange)},
		{"negative ISR withheld", pm601, softwareBusiness, func(d *domain.Draft) { d.Withheld.ISR = -1 }, domain.ClassValidation,
			violations(domain.FieldISRWithheldCents, domain.ErrOutOfRange)},
		{"IVA withheld above the maximum", pm601, softwareBusiness, func(d *domain.Draft) { d.Withheld.IVA = math.MaxInt64 }, domain.ClassValidation,
			violations(domain.FieldIVAWithheldCents, domain.ErrOutOfRange)},
		{"missing category", pf612, rule("", domain.Deductible, true), func(d *domain.Draft) { d.Category = "" }, domain.ClassValidation,
			violations(domain.FieldCategoryCode, domain.ErrRequired)},
		{"missing IVA rate", pf612, softwareBusiness, func(d *domain.Draft) { d.IVARate = "" }, domain.ClassValidation,
			violations(domain.FieldIVARate, domain.ErrRequired)},
		{"unknown IVA rate", pf612, softwareBusiness, func(d *domain.Draft) { d.IVARate = "rate_20" }, domain.ClassValidation,
			violations(domain.FieldIVARate, domain.ErrInvalid)},
		{"missing payment method", pf612, softwareBusiness, func(d *domain.Draft) { d.PaymentMethod = "" }, domain.ClassValidation,
			violations(domain.FieldPaymentMethod, domain.ErrRequired)},
		{"unknown payment method", pf612, softwareBusiness, func(d *domain.Draft) { d.PaymentMethod = "crypto" }, domain.ClassValidation,
			violations(domain.FieldPaymentMethod, domain.ErrInvalid)},
		{"missing paid_on", pf612, softwareBusiness, func(d *domain.Draft) { d.PaidOn = domain.Date{} }, domain.ClassValidation,
			violations(domain.FieldPaidOn, domain.ErrRequired)},
		{"paid tomorrow", pf612, softwareBusiness, func(d *domain.Draft) { d.PaidOn = date(t, "2026-10-14") }, domain.ClassValidation,
			violations(domain.FieldPaidOn, domain.ErrOutOfRange)},
		{"paid before 2022", pf612, softwareBusiness, func(d *domain.Draft) { d.PaidOn = date(t, "2021-12-31") }, domain.ClassValidation,
			violations(domain.FieldPaidOn, domain.ErrOutOfRange)},
		{"US-E1: cfdi_uuid without a CFDI", pf612, softwareBusiness, func(d *domain.Draft) {
			d.HasCFDI, d.CFDIUUID = false, someUUID
		}, domain.ClassValidation, violations(domain.FieldCFDIUUID, domain.ErrHasCFDIMismatch)},
		{"unknown override", pf612, softwareBusiness, func(d *domain.Draft) { d.Deductible = domain.Override(7) }, domain.ClassValidation,
			violations(domain.FieldDeductible, domain.ErrInvalid)},
		{"every validation error is reported", pf612, softwareBusiness, func(d *domain.Draft) {
			d.Amount, d.PaidOn = 0, domain.Date{}
			d.Withheld.ISR = 1_00
		}, domain.ClassValidation, domain.Violations{
			{Field: domain.FieldAmountCents, Err: domain.ErrMustBePositive},
			{Field: domain.FieldPaidOn, Err: domain.ErrRequired},
		}},
		{"US-E1: a física withholds ISR", pf612, feesBusiness, func(d *domain.Draft) {
			d.Category, d.Amount, d.IVA = domain.CategoryProfessionalFees, 10_000_00, 1_600_00
			d.Withheld = domain.Withholdings{ISR: 1_000_00}
		}, domain.ClassBusinessRule, violations(domain.FieldISRWithheldCents, domain.ErrWithholdingNotApplicable)},
		{"a física withholds IVA", pf612, feesBusiness, func(d *domain.Draft) {
			d.Category, d.Amount, d.IVA = domain.CategoryProfessionalFees, 10_000_00, 1_600_00
			d.Withheld = domain.Withholdings{IVA: 1_066_67}
		}, domain.ClassBusinessRule, violations(domain.FieldIVAWithheldCents, domain.ErrWithholdingNotApplicable)},
		{"ISR withheld above the amount", pm601, feesBusiness, func(d *domain.Draft) {
			d.Category = domain.CategoryProfessionalFees
			d.Withheld = domain.Withholdings{ISR: 500_01, IVA: 80_01}
		}, domain.ClassBusinessRule, domain.Violations{
			{Field: domain.FieldISRWithheldCents, Err: domain.ErrWithholdingExceedsBase},
			{Field: domain.FieldIVAWithheldCents, Err: domain.ErrWithholdingExceedsBase},
		}},
		{"IVA amount mismatch", pf612, softwareBusiness, func(d *domain.Draft) { d.IVA = 60_00 }, domain.ClassBusinessRule,
			violations(domain.FieldIVACents, domain.ErrIVAAmountMismatch)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := softwareDraft(t)
			tt.edit(&d)
			_, err := assess(t, d, tt.payer, tt.rule)
			got, ok := errors.AsType[domain.Violations](err)
			require.True(t, ok, "want Violations, got %v", err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantClass, got.Class())
		})
	}
}

func TestAssessRateNotValidOnDate(t *testing.T) {
	d := softwareDraft(t)
	d.IVARate, d.IVA, d.PaidOn = domain.IVARate8, 40_00, date(t, "2027-01-05")
	_, err := domain.Assess(d, pf612, softwareBusiness, date(t, "2027-02-01"), catalog2026(t))
	got, ok := errors.AsType[domain.Violations](err)
	require.True(t, ok)
	assert.Equal(t, violations(domain.FieldIVARate, domain.ErrRateNotValidOnDate), got)
}

func TestAssessServerFaults(t *testing.T) {
	absurd := func(key domain.RateKey) fakeCatalog {
		c := catalog2026(t)
		c.rates[key] = []dated[money.Rate]{{from: date(t, "2022-01-01"), value: math.MaxInt64}}
		return c
	}
	noCashLimit := catalog2026(t)
	delete(noCashLimit.amounts, domain.KeyCashPaymentLimit)
	cash := func(d *domain.Draft) { d.PaymentMethod = domain.PaymentCash }
	fees := func(d *domain.Draft) { d.Category = domain.CategoryProfessionalFees }
	none := func(*domain.Draft) {}
	badRule := softwareBusiness
	badRule.Deductibility = "maybe"
	tests := []struct {
		name    string
		payer   domain.Payer
		rule    domain.CategoryRule
		edit    func(*domain.Draft)
		today   string
		catalog domain.RateCatalog
	}{
		{"unknown person type", domain.Payer{PersonType: "trust", Regime: "612"}, softwareBusiness, none, today, catalog2026(t)},
		{"unknown deductibility in the rule", pf612, badRule, none, today, catalog2026(t)},
		{"rule of another category", pf612, papeleriaBusiness, none, today, catalog2026(t)},
		{"no today from the clock", pf612, softwareBusiness, none, "", catalog2026(t)},
		{"no catalog", pf612, softwareBusiness, none, today, nil},
		{"no cash-payment limit in force", pf612, softwareBusiness, cash, today, noCashLimit},
		{"catalog IVA rate overflows", pf612, softwareBusiness, none, today, absurd(domain.KeyIVAGeneral)},
		{"catalog ISR withholding rate overflows", pm601, feesBusiness, fees, today, absurd(domain.KeyISRWithholdingFees)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := softwareDraft(t)
			tt.edit(&d)
			var now domain.Date
			if tt.today != "" {
				now = date(t, tt.today)
			}
			_, err := domain.Assess(d, tt.payer, tt.rule, now, tt.catalog)
			require.ErrorIs(t, err, domain.ErrInvalidParameter)
			_, isViolation := errors.AsType[domain.Violations](err)
			assert.False(t, isViolation)
		})
	}
}

func violations(field string, err domain.Error) domain.Violations {
	return domain.Violations{{Field: field, Err: err}}
}
