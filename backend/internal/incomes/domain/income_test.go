package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

const today = "2026-10-13"

// us1Draft is US-I1: 10,000.00 + IVA 1,600.00 billed by a 612 física to a
// persona moral, with the suggested withholdings entered.
func us1Draft(t *testing.T) domain.Draft {
	t.Helper()
	return domain.Draft{
		Amount:       10_000_00,
		IVARate:      fiscal.IVARate16,
		IVA:          1_600_00,
		Counterparty: domain.CounterpartyMoral,
		Withheld:     fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67},
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
	assert.Equal(t, fiscal.Suggestion{Cents: 1_000_00, Known: true}, got.Suggested.ISR)
	assert.Equal(t, fiscal.Suggestion{Cents: 1_066_67, Known: true}, got.Suggested.IVA)
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
				d.Withheld = fiscal.Withholdings{}
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-05")}
			},
			wantBooked: "2026-09-05", wantPending: true,
		},
		{
			name: "601 collected later keeps the invoice month", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Counterparty = domain.CounterpartyPublic
				d.Withheld = fiscal.Withholdings{}
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
				d.Withheld = fiscal.Withholdings{ISR: 600_00, IVA: 1_200_00}
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
			edit:       func(d *domain.Draft) { d.Withheld = fiscal.Withholdings{ISR: 500_00, IVA: 500_00} },
			wantBooked: "2026-09-12",
			wantWarnings: []domain.Warning{
				{Code: domain.WarningWithholdingDiffersFromRate, Field: domain.FieldISRWithheldCents},
				{Code: domain.WarningWithholdingDiffersFromRate, Field: domain.FieldIVAWithheldCents},
			},
		},
		{
			name: "PM withheld 4 % freight IVA by a PM: no suggestion, no warning", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Withheld = fiscal.Withholdings{IVA: 400_00}
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
			name: "informational regime with no withholding", payee: payee(fiscal.PersonFisica, "605", domain.BasisCash),
			edit: func(d *domain.Draft) {
				d.IVARate, d.IVA = fiscal.IVAExempt, 0
				d.Withheld = fiscal.Withholdings{}
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
			d.Withheld = fiscal.Withholdings{}
		}
	}
	tests := []struct {
		name      string
		payee     domain.Payee
		edit      func(*domain.Draft)
		wantClass fiscal.Class
		want      fiscal.Violations
	}{
		{"amount 0", pf612, func(d *domain.Draft) { d.Amount = 0 }, fiscal.ClassValidation,
			violations(domain.FieldAmountCents, fiscal.ErrMustBePositive)},
		{"negative amount", pf612, func(d *domain.Draft) { d.Amount = -1 }, fiscal.ClassValidation,
			violations(domain.FieldAmountCents, fiscal.ErrMustBePositive)},
		{"amount above the contract maximum", pf612, func(d *domain.Draft) { d.Amount = fiscal.MaxCents + 1 }, fiscal.ClassValidation,
			violations(domain.FieldAmountCents, fiscal.ErrOutOfRange)},
		{"negative IVA", pf612, func(d *domain.Draft) { d.IVA = -1 }, fiscal.ClassValidation,
			violations(domain.FieldIVACents, fiscal.ErrOutOfRange)},
		{"negative ISR withheld", pf612, func(d *domain.Draft) { d.Withheld.ISR = -1 }, fiscal.ClassValidation,
			violations(domain.FieldISRWithheldCents, fiscal.ErrOutOfRange)},
		{"IVA withheld above the maximum", pf612, func(d *domain.Draft) { d.Withheld.IVA = math.MaxInt64 }, fiscal.ClassValidation,
			violations(domain.FieldIVAWithheldCents, fiscal.ErrOutOfRange)},
		{"missing IVA rate", pf612, func(d *domain.Draft) { d.IVARate = "" }, fiscal.ClassValidation,
			violations(domain.FieldIVARate, fiscal.ErrRequired)},
		{"unknown IVA rate", pf612, func(d *domain.Draft) { d.IVARate = "rate_20" }, fiscal.ClassValidation,
			violations(domain.FieldIVARate, fiscal.ErrInvalid)},
		{"missing counterparty", pf612, func(d *domain.Draft) { d.Counterparty = "" }, fiscal.ClassValidation,
			violations(domain.FieldCounterpartyType, fiscal.ErrRequired)},
		{"unknown counterparty", pf612, func(d *domain.Draft) { d.Counterparty = "gobierno" }, fiscal.ClassValidation,
			violations(domain.FieldCounterpartyType, fiscal.ErrInvalid)},
		{"US-I1: 612 without collected_on", pf612, func(d *domain.Draft) { d.Dates = domain.Dates{} }, fiscal.ClassValidation,
			violations(domain.FieldCollectedOn, fiscal.ErrRequired)},
		{"601 without invoiced_on", pm601, func(d *domain.Draft) { d.Withheld = fiscal.Withholdings{} }, fiscal.ClassValidation,
			violations(domain.FieldInvoicedOn, fiscal.ErrRequired)},
		{"US-I1: collected tomorrow", pf612, func(d *domain.Draft) { d.Dates.CollectedOn = date(t, "2026-10-14") }, fiscal.ClassValidation,
			violations(domain.FieldCollectedOn, fiscal.ErrOutOfRange)},
		{"US-I1: collected before 2022-01-01", pf612, func(d *domain.Draft) { d.Dates.CollectedOn = date(t, "2021-12-31") }, fiscal.ClassValidation,
			violations(domain.FieldCollectedOn, fiscal.ErrOutOfRange)},
		{"optional invoice date in the future", pf612, func(d *domain.Draft) { d.Dates.InvoicedOn = date(t, "2027-01-01") }, fiscal.ClassValidation,
			violations(domain.FieldInvoicedOn, fiscal.ErrOutOfRange)},
		{"US-I1: 625 with two activities and none sent", twoActivities, platform(""), fiscal.ClassValidation,
			violations(domain.FieldPlatformActivity, fiscal.ErrRequired)},
		{"625 with an unknown activity", twoActivities, platform("mining"), fiscal.ClassValidation,
			violations(domain.FieldPlatformActivity, fiscal.ErrInvalid)},
		{"platform_activity on a 612 income", pf612, func(d *domain.Draft) { d.PlatformActivity = domain.ActivityLodging }, fiscal.ClassValidation,
			violations(domain.FieldPlatformActivity, domain.ErrNotForRegime)},
		{"validation errors are all reported, business rules wait", pf612, func(d *domain.Draft) {
			d.Amount = 0
			d.Dates = domain.Dates{}
			d.Counterparty = domain.CounterpartyFisica
		}, fiscal.ClassValidation, fiscal.Violations{
			{Field: domain.FieldAmountCents, Err: fiscal.ErrMustBePositive},
			{Field: domain.FieldCollectedOn, Err: fiscal.ErrRequired},
		}},
		{"US-I1: ISR withheld by a persona física", pf612, func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyFisica
			d.Withheld = fiscal.Withholdings{ISR: 5_00}
		}, fiscal.ClassBusinessRule, violations(domain.FieldISRWithheldCents, fiscal.ErrWithholdingNotApplicable)},
		{"IVA withheld by the general public", pf612, func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyPublic
			d.Withheld = fiscal.Withholdings{IVA: 1_00}
		}, fiscal.ClassBusinessRule, violations(domain.FieldIVAWithheldCents, fiscal.ErrWithholdingNotApplicable)},
		{"IVA withheld from a PM by a platform", pm601, func(d *domain.Draft) {
			d.Counterparty = domain.CounterpartyPlatform
			d.Withheld = fiscal.Withholdings{ISR: 250_00, IVA: 1_00}
			d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12")}
		}, fiscal.ClassBusinessRule, violations(domain.FieldIVAWithheldCents, fiscal.ErrWithholdingNotApplicable)},
		{"ISR withheld above the amount", pf612, func(d *domain.Draft) { d.Withheld.ISR = 10_000_01 }, fiscal.ClassBusinessRule,
			violations(domain.FieldISRWithheldCents, fiscal.ErrWithholdingExceedsBase)},
		{"IVA withheld above the IVA", pf612, func(d *domain.Draft) { d.Withheld.IVA = 1_600_01 }, fiscal.ClassBusinessRule,
			violations(domain.FieldIVAWithheldCents, fiscal.ErrWithholdingExceedsBase)},
		{"US-I1: exempt with IVA 1.00", pf612, func(d *domain.Draft) {
			d.IVARate, d.IVA = fiscal.IVAExempt, 1_00
			d.Withheld.IVA = 0
		}, fiscal.ClassBusinessRule, violations(domain.FieldIVACents, fiscal.ErrIVAAmountMismatch)},
		{"US-I1: 16 % of 10,000 entered as 1,500", pf612, func(d *domain.Draft) {
			d.IVA = 1_500_00
			d.Withheld.IVA = 1_000_00
		}, fiscal.ClassBusinessRule, violations(domain.FieldIVACents, fiscal.ErrIVAAmountMismatch)},
		{"US-I1: 625 lodging not among the taxpayer's activities", twoActivities, platform(domain.ActivityLodging), fiscal.ClassBusinessRule,
			violations(domain.FieldPlatformActivity, domain.ErrActivityNotHeld)},
		{"every business rule is reported", pf612, func(d *domain.Draft) {
			d.IVA = 1_500_00
			d.Withheld = fiscal.Withholdings{ISR: 10_000_01, IVA: 1_500_01}
		}, fiscal.ClassBusinessRule, fiscal.Violations{
			{Field: domain.FieldIVACents, Err: fiscal.ErrIVAAmountMismatch},
			{Field: domain.FieldISRWithheldCents, Err: fiscal.ErrWithholdingExceedsBase},
			{Field: domain.FieldIVAWithheldCents, Err: fiscal.ErrWithholdingExceedsBase},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := us1Draft(t)
			tt.edit(&d)
			_, err := assess(t, d, tt.payee)
			got, ok := errors.AsType[fiscal.Violations](err)
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
	d.IVARate, d.IVA = fiscal.IVARate8, 800_00
	d.Withheld = fiscal.Withholdings{}
	d.Counterparty = domain.CounterpartyPublic
	d.Dates.CollectedOn = date(t, "2027-01-05")

	_, err := domain.Assess(d, pf612, date(t, "2027-02-01"), catalog2026(t))
	require.ErrorIs(t, err, fiscal.ErrRateNotValidOnDate)
	got, ok := errors.AsType[fiscal.Violations](err)
	require.True(t, ok)
	assert.Equal(t, violations(domain.FieldIVARate, fiscal.ErrRateNotValidOnDate), got)
}

// TestAssessPendingIncomeRateDate: an accrual income still pending
// collection validates its IVA rate on the invoice date (§2.5).
func TestAssessPendingIncomeRateDate(t *testing.T) {
	d := us1Draft(t)
	d.IVARate, d.IVA = fiscal.IVARate8, 800_00
	d.Counterparty, d.Withheld = domain.CounterpartyPublic, fiscal.Withholdings{}
	d.Dates = domain.Dates{InvoicedOn: date(t, "2026-12-31")}

	got, err := domain.Assess(d, pm601, date(t, "2027-02-01"), catalog2026(t))
	require.NoError(t, err)
	assert.True(t, got.PendingCollection)

	d.Dates.CollectedOn = date(t, "2027-01-05")
	_, err = domain.Assess(d, pm601, date(t, "2027-02-01"), catalog2026(t))
	require.ErrorIs(t, err, fiscal.ErrRateNotValidOnDate, "once collected, the collection date rules")
}

func TestAssessServerFaults(t *testing.T) {
	absurdIVA := catalog2026(t)
	absurdIVA[fiscal.KeyIVAGeneral] = []datedRate{{from: date(t, "2022-01-01"), rate: math.MaxInt64}}
	absurdISR := catalog2026(t)
	absurdISR[fiscal.KeyISRWithholdingFees] = []datedRate{{from: date(t, "2022-01-01"), rate: math.MaxInt64}}
	badBasis := pf612
	badBasis.Basis = "weekly"
	badPerson := pf612
	badPerson.PersonType = "trust"
	tests := []struct {
		name    string
		payee   domain.Payee
		today   fiscal.Date
		catalog fiscal.RateCatalog
	}{
		{"catalog IVA rate overflows", pf612, date(t, today), absurdIVA},
		{"catalog ISR rate overflows", pf612, date(t, today), absurdISR},
		{"unknown income basis", badBasis, date(t, today), catalog2026(t)},
		{"unknown person type", badPerson, date(t, today), catalog2026(t)},
		{"no today from the clock", pf612, fiscal.Date{}, catalog2026(t)},
		{"no catalog", pf612, date(t, today), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.Assess(us1Draft(t), tt.payee, tt.today, tt.catalog)
			require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
			_, isViolation := errors.AsType[fiscal.Violations](err)
			assert.False(t, isViolation, "a server fault is never the client's violation")
		})
	}
}

func TestMaxCents(t *testing.T) {
	assert.Equal(t, money.Cents(99_999_999_999_999), fiscal.MaxCents)
}

func violations(field string, err fiscal.Error) fiscal.Violations {
	return fiscal.Violations{{Field: field, Err: err}}
}
