package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

const today = "2026-10-13"

// us1Draft is US-I1: 10,000.00 + IVA 1,600.00 billed by a 612 física to a
// persona moral, with the suggested withholdings entered.
func us1Draft(t *testing.T) domain.Draft {
	t.Helper()
	return domain.Draft{
		Amount:       10_000_00,
		IVARate:      domain.IVARate16,
		IVA:          1_600_00,
		Counterparty: domain.CounterpartyMoral,
		Withheld:     domain.Withholdings{ISR: 1_000_00, IVA: 1_066_67},
		Dates:        domain.Dates{CollectedOn: date(t, "2026-09-12")},
	}
}

func assess(t *testing.T, d domain.Draft, p domain.Payee) (domain.Assessment, error) {
	t.Helper()
	return domain.Assess(d, p, date(t, today), catalog2026(t))
}

func TestAssessUS1(t *testing.T) {
	got, err := assess(t, us1Draft(t), pf612)
	require.NoError(t, err)

	assert.Equal(t, "2026-09-12", got.BookedOn.String())
	assert.False(t, got.PendingCollection)
	assert.Empty(t, got.PlatformActivity)
	assert.Empty(t, got.Warnings)
	assert.Equal(t, domain.Suggestion{Cents: 1_000_00, Known: true}, got.Suggested.ISR)
	assert.Equal(t, domain.Suggestion{Cents: 1_066_67, Known: true}, got.Suggested.IVA)
	assert.Equal(t, domain.Totals{
		Subtotal: 10_000_00,
		IVA:      1_600_00,
		Total:    11_600_00,
		Withheld: 2_066_67,
		Net:      9_533_33, // "Neto a recibir 9,533.33"
	}, got.Totals)
}

func TestAssessAccepts(t *testing.T) {
	lodgingOnly := pf625
	lodgingOnly.PlatformActivities = []domain.PlatformActivity{domain.ActivityLodging}
	tests := []struct {
		name         string
		payee        domain.Payee
		edit         func(*domain.Draft)
		wantBooked   string
		wantPending  bool
		wantActivity domain.PlatformActivity
		wantWarnings []domain.Warning
	}{
		{
			name: "601 invoiced 2026-09-05, not collected: Por cobrar", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Counterparty = domain.CounterpartyPublic
				d.Withheld = domain.Withholdings{}
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-05")}
			},
			wantBooked: "2026-09-05", wantPending: true,
		},
		{
			name: "601 collected later keeps the invoice month", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Counterparty = domain.CounterpartyPublic
				d.Withheld = domain.Withholdings{}
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-05"), CollectedOn: date(t, "2026-10-01")}
			},
			wantBooked: "2026-09-05",
		},
		{
			name: "IVA 1,600.40 is inside the tolerance", payee: pf612,
			edit:       func(d *domain.Draft) { d.IVA = 1_600_40 },
			wantBooked: "2026-09-12",
		},
		{
			name: "625 with one activity defaults to it (PF-625-02)", payee: lodgingOnly,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA = 15_000_00, 2_400_00
				d.Counterparty = domain.CounterpartyPlatform
				d.Withheld = domain.Withholdings{ISR: 600_00, IVA: 1_200_00}
			},
			wantBooked: "2026-09-12", wantActivity: domain.ActivityLodging,
		},
		{
			name: "withholding inside the tolerance of the suggestion", payee: pf612,
			edit:       func(d *domain.Draft) { d.Withheld.ISR = 1_001_00 },
			wantBooked: "2026-09-12",
		},
		{
			name: "CFDI wins: ISR withheld differs from 10 %", payee: pf612,
			edit:       func(d *domain.Draft) { d.Withheld.ISR = 0 },
			wantBooked: "2026-09-12",
			wantWarnings: []domain.Warning{
				{Code: domain.WarningWithholdingDiffersFromRate, Field: domain.FieldISRWithheldCents},
			},
		},
		{
			name: "both withholdings differ", payee: pf612,
			edit:       func(d *domain.Draft) { d.Withheld = domain.Withholdings{ISR: 500_00, IVA: 500_00} },
			wantBooked: "2026-09-12",
			wantWarnings: []domain.Warning{
				{Code: domain.WarningWithholdingDiffersFromRate, Field: domain.FieldISRWithheldCents},
				{Code: domain.WarningWithholdingDiffersFromRate, Field: domain.FieldIVAWithheldCents},
			},
		},
		{
			name: "PM withheld 4 % freight IVA by a PM: no suggestion, no warning", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Withheld = domain.Withholdings{IVA: 400_00}
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12"), CollectedOn: date(t, "2026-09-12")}
			},
			wantBooked: "2026-09-12",
		},
		{
			name: "advance collected before the invoice", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-10-13"), CollectedOn: date(t, "2022-01-01")}
			},
			wantBooked: "2022-01-01",
		},
		{
			name: "informational regime with no withholding", payee: payee(domain.PersonFisica, "605", domain.BasisCash),
			edit: func(d *domain.Draft) {
				d.IVARate, d.IVA = domain.IVAExempt, 0
				d.Withheld = domain.Withholdings{}
			},
			wantBooked: "2026-09-12",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := us1Draft(t)
			tt.edit(&d)
			got, err := assess(t, d, tt.payee)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBooked, got.BookedOn.String())
			assert.Equal(t, tt.wantPending, got.PendingCollection)
			assert.Equal(t, tt.wantActivity, got.PlatformActivity)
			assert.Equal(t, tt.wantWarnings, got.Warnings)
		})
	}
}

// TestAssessRejects covers the 400 and 422 rows of MODULE-incomes §4 and
// the US-I1 refusals. Validation (400) runs first: a draft with schema-level
// problems never reaches the business rules.
func TestAssessRejects(t *testing.T) {
	twoActivities := pf625
	twoActivities.PlatformActivities = []domain.PlatformActivity{domain.ActivityTransportDelivery, domain.ActivityGoodsServices}
	platform := func(a domain.PlatformActivity) func(*domain.Draft) {
		return func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyPlatform
			d.PlatformActivity = a
			d.Withheld = domain.Withholdings{}
		}
	}
	tests := []struct {
		name      string
		payee     domain.Payee
		edit      func(*domain.Draft)
		wantClass domain.Class
		want      domain.Violations
	}{
		{"amount 0", pf612, func(d *domain.Draft) { d.Amount = 0 }, domain.ClassValidation,
			violations(domain.FieldAmountCents, domain.ErrMustBePositive)},
		{"negative amount", pf612, func(d *domain.Draft) { d.Amount = -1 }, domain.ClassValidation,
			violations(domain.FieldAmountCents, domain.ErrMustBePositive)},
		{"amount above the contract maximum", pf612, func(d *domain.Draft) { d.Amount = domain.MaxCents + 1 }, domain.ClassValidation,
			violations(domain.FieldAmountCents, domain.ErrOutOfRange)},
		{"negative IVA", pf612, func(d *domain.Draft) { d.IVA = -1 }, domain.ClassValidation,
			violations(domain.FieldIVACents, domain.ErrOutOfRange)},
		{"negative ISR withheld", pf612, func(d *domain.Draft) { d.Withheld.ISR = -1 }, domain.ClassValidation,
			violations(domain.FieldISRWithheldCents, domain.ErrOutOfRange)},
		{"IVA withheld above the maximum", pf612, func(d *domain.Draft) { d.Withheld.IVA = math.MaxInt64 }, domain.ClassValidation,
			violations(domain.FieldIVAWithheldCents, domain.ErrOutOfRange)},
		{"missing IVA rate", pf612, func(d *domain.Draft) { d.IVARate = "" }, domain.ClassValidation,
			violations(domain.FieldIVARate, domain.ErrRequired)},
		{"unknown IVA rate", pf612, func(d *domain.Draft) { d.IVARate = "rate_20" }, domain.ClassValidation,
			violations(domain.FieldIVARate, domain.ErrInvalid)},
		{"missing counterparty", pf612, func(d *domain.Draft) { d.Counterparty = "" }, domain.ClassValidation,
			violations(domain.FieldCounterpartyType, domain.ErrRequired)},
		{"unknown counterparty", pf612, func(d *domain.Draft) { d.Counterparty = "gobierno" }, domain.ClassValidation,
			violations(domain.FieldCounterpartyType, domain.ErrInvalid)},
		{"US-I1: 612 without collected_on", pf612, func(d *domain.Draft) { d.Dates = domain.Dates{} }, domain.ClassValidation,
			violations(domain.FieldCollectedOn, domain.ErrRequired)},
		{"601 without invoiced_on", pm601, func(d *domain.Draft) { d.Withheld = domain.Withholdings{} }, domain.ClassValidation,
			violations(domain.FieldInvoicedOn, domain.ErrRequired)},
		{"US-I1: collected tomorrow", pf612, func(d *domain.Draft) { d.Dates.CollectedOn = date(t, "2026-10-14") }, domain.ClassValidation,
			violations(domain.FieldCollectedOn, domain.ErrOutOfRange)},
		{"US-I1: collected before 2022-01-01", pf612, func(d *domain.Draft) { d.Dates.CollectedOn = date(t, "2021-12-31") }, domain.ClassValidation,
			violations(domain.FieldCollectedOn, domain.ErrOutOfRange)},
		{"optional invoice date in the future", pf612, func(d *domain.Draft) { d.Dates.InvoicedOn = date(t, "2027-01-01") }, domain.ClassValidation,
			violations(domain.FieldInvoicedOn, domain.ErrOutOfRange)},
		{"US-I1: 625 with two activities and none sent", twoActivities, platform(""), domain.ClassValidation,
			violations(domain.FieldPlatformActivity, domain.ErrRequired)},
		{"625 with an unknown activity", twoActivities, platform("mining"), domain.ClassValidation,
			violations(domain.FieldPlatformActivity, domain.ErrInvalid)},
		{"platform_activity on a 612 income", pf612, func(d *domain.Draft) { d.PlatformActivity = domain.ActivityLodging }, domain.ClassValidation,
			violations(domain.FieldPlatformActivity, domain.ErrNotForRegime)},
		{"validation errors are all reported, business rules wait", pf612, func(d *domain.Draft) {
			d.Amount = 0
			d.Dates = domain.Dates{}
			d.Counterparty = domain.CounterpartyFisica
		}, domain.ClassValidation, domain.Violations{
			{Field: domain.FieldAmountCents, Err: domain.ErrMustBePositive},
			{Field: domain.FieldCollectedOn, Err: domain.ErrRequired},
		}},
		{"US-I1: ISR withheld by a persona física", pf612, func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyFisica
			d.Withheld = domain.Withholdings{ISR: 5_00}
		}, domain.ClassBusinessRule, violations(domain.FieldISRWithheldCents, domain.ErrWithholdingNotApplicable)},
		{"IVA withheld by the general public", pf612, func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyPublic
			d.Withheld = domain.Withholdings{IVA: 1_00}
		}, domain.ClassBusinessRule, violations(domain.FieldIVAWithheldCents, domain.ErrWithholdingNotApplicable)},
		{"IVA withheld from a PM by a platform", pm601, func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyPlatform
			d.Withheld = domain.Withholdings{ISR: 250_00, IVA: 1_00}
			d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12")}
		}, domain.ClassBusinessRule, violations(domain.FieldIVAWithheldCents, domain.ErrWithholdingNotApplicable)},
		{"ISR withheld above the amount", pf612, func(d *domain.Draft) { d.Withheld.ISR = 10_000_01 }, domain.ClassBusinessRule,
			violations(domain.FieldISRWithheldCents, domain.ErrWithholdingExceedsBase)},
		{"IVA withheld above the IVA", pf612, func(d *domain.Draft) { d.Withheld.IVA = 1_600_01 }, domain.ClassBusinessRule,
			violations(domain.FieldIVAWithheldCents, domain.ErrWithholdingExceedsBase)},
		{"US-I1: exempt with IVA 1.00", pf612, func(d *domain.Draft) {
			d.IVARate, d.IVA = domain.IVAExempt, 1_00
			d.Withheld.IVA = 0
		}, domain.ClassBusinessRule, violations(domain.FieldIVACents, domain.ErrIVAAmountMismatch)},
		{"US-I1: 16 % of 10,000 entered as 1,500", pf612, func(d *domain.Draft) {
			d.IVA = 1_500_00
			d.Withheld.IVA = 1_000_00
		}, domain.ClassBusinessRule, violations(domain.FieldIVACents, domain.ErrIVAAmountMismatch)},
		{"US-I1: 625 lodging not among the taxpayer's activities", twoActivities, platform(domain.ActivityLodging), domain.ClassBusinessRule,
			violations(domain.FieldPlatformActivity, domain.ErrActivityNotHeld)},
		{"every business rule is reported", pf612, func(d *domain.Draft) {
			d.IVA = 1_500_00
			d.Withheld = domain.Withholdings{ISR: 10_000_01, IVA: 1_500_01}
		}, domain.ClassBusinessRule, domain.Violations{
			{Field: domain.FieldIVACents, Err: domain.ErrIVAAmountMismatch},
			{Field: domain.FieldISRWithheldCents, Err: domain.ErrWithholdingExceedsBase},
			{Field: domain.FieldIVAWithheldCents, Err: domain.ErrWithholdingExceedsBase},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := us1Draft(t)
			tt.edit(&d)
			_, err := assess(t, d, tt.payee)
			got, ok := errors.AsType[domain.Violations](err)
			require.True(t, ok, "want Violations, got %v", err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantClass, got.Class())
		})
	}
}

// TestAssessRateNotValidOnDate is US-I1: a rate_8 income collected on
// 2027-01-05 with no 2027 parameter extending the border decree.
func TestAssessRateNotValidOnDate(t *testing.T) {
	d := us1Draft(t)
	d.IVARate, d.IVA = domain.IVARate8, 800_00
	d.Withheld = domain.Withholdings{}
	d.Counterparty = domain.CounterpartyPublic
	d.Dates.CollectedOn = date(t, "2027-01-05")

	_, err := domain.Assess(d, pf612, date(t, "2027-02-01"), catalog2026(t))
	require.ErrorIs(t, err, domain.ErrRateNotValidOnDate)
	got, ok := errors.AsType[domain.Violations](err)
	require.True(t, ok)
	assert.Equal(t, violations(domain.FieldIVARate, domain.ErrRateNotValidOnDate), got)
}

// TestAssessPendingIncomeRateDate: an accrual income still pending
// collection validates its IVA rate on the invoice date (§2.5).
func TestAssessPendingIncomeRateDate(t *testing.T) {
	d := us1Draft(t)
	d.IVARate, d.IVA = domain.IVARate8, 800_00
	d.Counterparty, d.Withheld = domain.CounterpartyPublic, domain.Withholdings{}
	d.Dates = domain.Dates{InvoicedOn: date(t, "2026-12-31")}

	got, err := domain.Assess(d, pm601, date(t, "2027-02-01"), catalog2026(t))
	require.NoError(t, err)
	assert.True(t, got.PendingCollection)

	d.Dates.CollectedOn = date(t, "2027-01-05")
	_, err = domain.Assess(d, pm601, date(t, "2027-02-01"), catalog2026(t))
	require.ErrorIs(t, err, domain.ErrRateNotValidOnDate, "once collected, the collection date rules")
}

func TestAssessServerFaults(t *testing.T) {
	absurdIVA := catalog2026(t)
	absurdIVA[domain.KeyIVAGeneral] = []datedRate{{from: date(t, "2022-01-01"), rate: math.MaxInt64}}
	absurdISR := catalog2026(t)
	absurdISR[domain.KeyISRWithholdingFees] = []datedRate{{from: date(t, "2022-01-01"), rate: math.MaxInt64}}
	badBasis := pf612
	badBasis.Basis = "weekly"
	badPerson := pf612
	badPerson.PersonType = "trust"
	tests := []struct {
		name    string
		payee   domain.Payee
		today   domain.Date
		catalog domain.RateCatalog
	}{
		{"catalog IVA rate overflows", pf612, date(t, today), absurdIVA},
		{"catalog ISR rate overflows", pf612, date(t, today), absurdISR},
		{"unknown income basis", badBasis, date(t, today), catalog2026(t)},
		{"unknown person type", badPerson, date(t, today), catalog2026(t)},
		{"no today from the clock", pf612, domain.Date{}, catalog2026(t)},
		{"no catalog", pf612, date(t, today), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.Assess(us1Draft(t), tt.payee, tt.today, tt.catalog)
			require.ErrorIs(t, err, domain.ErrInvalidParameter)
			_, isViolation := errors.AsType[domain.Violations](err)
			assert.False(t, isViolation, "a server fault is never the client's violation")
		})
	}
}

func TestMaxCents(t *testing.T) {
	assert.Equal(t, money.Cents(99_999_999_999_999), domain.MaxCents)
}

func violations(field string, err domain.Error) domain.Violations {
	return domain.Violations{{Field: field, Err: err}}
}
