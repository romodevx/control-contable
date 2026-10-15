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

// us1Draft is US-I1 with derived withholdings: 10,000.00 + IVA 1,600.00
// billed by a 612 física to a persona moral for professional services.
func us1Draft(t *testing.T) domain.Draft {
	t.Helper()
	return domain.Draft{
		Amount:       10_000_00,
		IVARate:      fiscal.IVARate16,
		IVA:          1_600_00,
		Counterparty: fiscal.ReceptorPM,
		Operation:    fiscal.OperationProfessionalServices,
		Dates:        domain.Dates{CollectedOn: date(t, "2026-09-12")},
	}
}

func assess(t *testing.T, d domain.Draft, p domain.Payee) (domain.Assessment, error) {
	t.Helper()
	return domain.Assess(d, p, date(t, today), withholdingCatalog2026(t))
}

func TestAssessUS1DerivesTheWithholdings(t *testing.T) {
	got, err := assess(t, us1Draft(t), pf612)
	require.NoError(t, err)

	assert.Equal(t, "2026-09-12", got.BookedOn.String())
	assert.False(t, got.PendingCollection)
	assert.Empty(t, got.PlatformActivity)
	assert.Empty(t, got.Warnings)
	assert.Equal(t, fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67},
		ISRRate:      rate(t, "0.10"), IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
		RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6}, TableVersion: fiscal.WithholdingTableVersion,
	}, got.Withholding)
}

// TestAssessDerivesTheUserStories are the MODULE-withholdings §8 incomes
// stories to the centavo (WM goldens WH-01..WH-12).
func TestAssessDerivesTheUserStories(t *testing.T) {
	tests := []struct {
		name         string
		payee        domain.Payee
		edit         func(*domain.Draft)
		wantISR      money.Cents
		wantIVA      money.Cents
		wantNet      money.Cents
		wantRules    []fiscal.RuleID
		wantWarnings []domain.Warning
	}{
		{
			name: "US-W1 (WH-01) owner: RESICO PF 100,000.00 of services to a PM", payee: pf626,
			edit:    func(d *domain.Draft) { d.Amount, d.IVA = 100_000_00, 16_000_00 },
			wantISR: 1_250_00, wantIVA: 10_666_67, wantNet: 104_083_33,
			wantRules: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6},
		},
		{
			name: "US-W2 (WH-10) 612 to the general public", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Counterparty = 5_000_00, 800_00, fiscal.ReceptorPFNoActivityOrPublic
			},
			wantNet: 5_800_00, wantRules: []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA0},
		},
		{
			name: "US-W2 (WH-04) 612 to a persona física with activity", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Counterparty = 20_000_00, 3_200_00, fiscal.ReceptorPFWithActivity
			},
			wantNet: 23_200_00, wantRules: []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA1},
		},
		{
			name: "US-W3 (WH-02) 612 honorarios 50,000.00 to a company", payee: pf612,
			edit:    func(d *domain.Draft) { d.Amount, d.IVA = 50_000_00, 8_000_00 },
			wantISR: 5_000_00, wantIVA: 5_333_33, wantNet: 47_666_67,
			wantRules: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6},
		},
		{
			name: "US-W3 (WH-03) 606 commercial rent", payee: pf606,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Operation = 30_000_00, 4_800_00, fiscal.OperationRentRealEstate
			},
			wantISR: 3_000_00, wantIVA: 3_200_00, wantNet: 28_600_00,
			wantRules: []fiscal.RuleID{fiscal.RuleISR6, fiscal.RuleIVA6},
		},
		{
			name: "US-W3 (WH-03b) 606 housing rent is exempt", payee: pf606,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVARate, d.IVA, d.Operation = 30_000_00, fiscal.IVAExempt, 0, fiscal.OperationRentRealEstate
				d.ResidentialRent = true
			},
			wantISR: 3_000_00, wantNet: 27_000_00, wantRules: []fiscal.RuleID{fiscal.RuleISR6, fiscal.RuleIVANoTax},
		},
		{
			name: "US-W4 (WH-05) 612 sells goods: no withholding", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Operation = 50_000_00, 8_000_00, fiscal.OperationSaleOfGoods
			},
			wantNet: 58_000_00, wantRules: []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA8},
		},
		{
			name: "US-W4 (WH-05b) RESICO sells goods: ISR only", payee: pf626,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Operation = 50_000_00, 8_000_00, fiscal.OperationSaleOfGoods
			},
			wantISR: 625_00, wantNet: 57_375_00, wantRules: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA8},
		},
		{
			name: "US-W4 (WH-06) PM 601 freight to a company: 4 % of the price", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Operation = fiscal.OperationGroundFreight
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12"), CollectedOn: date(t, "2026-09-12")}
			},
			wantIVA: 400_00, wantNet: 11_200_00, wantRules: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA4},
		},
		{
			name: "US-W4 (WH-07) PM 601 specialized services: no withholding", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Operation = 100_000_00, 16_000_00, fiscal.OperationBusinessServices
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12")}
			},
			wantNet: 116_000_00, wantRules: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA8},
		},
		{
			name: "US-W4 (WH-08) 612 scrap to a company: all the IVA", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Operation = 20_000_00, 3_200_00, fiscal.OperationScrap
			},
			wantIVA: 3_200_00, wantNet: 20_000_00, wantRules: []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA5},
		},
		{
			name: "US-W4 (WH-09) 625 goods via the platform, RFC given", payee: pf625goods,
			edit: func(d *domain.Draft) {
				d.IVA, d.ViaPlatform, d.Counterparty = 1_600_00, true, fiscal.ReceptorPFNoActivityOrPublic
			},
			wantISR: 250_00, wantIVA: 800_00, wantNet: 10_550_00,
			wantRules: []fiscal.RuleID{fiscal.RuleISR3, fiscal.RuleIVA3},
		},
		{
			name: "625 via the platform without RFC: 20 % and all the IVA", payee: pf625goodsNoRFC,
			edit:    func(d *domain.Draft) { d.ViaPlatform = true },
			wantISR: 2_000_00, wantIVA: 1_600_00, wantNet: 8_000_00,
			wantRules: []fiscal.RuleID{fiscal.RuleISR3, fiscal.RuleIVA3},
		},
		{
			name: "a persona moral via a platform: 2.5 % ISR", payee: pm601,
			edit: func(d *domain.Draft) {
				d.ViaPlatform = true
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12"), CollectedOn: date(t, "2026-09-12")}
			},
			wantISR: 250_00, wantNet: 11_350_00, wantRules: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA8},
		},
		{
			name: "US-W5 (WH-12) 612 honorarios to the federal government warn", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Counterparty = 30_000_00, 4_800_00, fiscal.ReceptorGovFederal
			},
			wantISR: 3_000_00, wantIVA: 3_200_00, wantNet: 28_600_00,
			wantRules:    []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA7},
			wantWarnings: []domain.Warning{{Code: domain.WarningWithholdingCPAUnverified, Field: domain.FieldWithholding}},
		},
		{
			name: "US-W5 a state government withholds no IVA", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Amount, d.IVA, d.Counterparty = 30_000_00, 4_800_00, fiscal.ReceptorGovStateMunicipal
			},
			wantISR: 3_000_00, wantNet: 31_800_00,
			wantRules:    []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA2},
			wantWarnings: []domain.Warning{{Code: domain.WarningWithholdingCPAUnverified, Field: domain.FieldWithholding}},
		},
		{
			name: "a customer abroad: nothing derived, warned", payee: pf612,
			edit:      func(d *domain.Draft) { d.Counterparty = fiscal.ReceptorForeign },
			wantNet:   11_600_00,
			wantRules: []fiscal.RuleID{fiscal.RuleISR1, fiscal.RuleIVA0},
			wantWarnings: []domain.Warning{
				{Code: domain.WarningForeignCounterpartyNotSupported, Field: domain.FieldCounterpartyKind},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := us1Draft(t)
			tt.edit(&d)
			got, err := assess(t, d, tt.payee)
			require.NoError(t, err)
			assert.Equal(t, tt.wantISR, got.Withholding.ISR, "ISR")
			assert.Equal(t, tt.wantIVA, got.Withholding.IVA, "IVA")
			net, err := got.Withholding.Net(d.Amount, d.IVA)
			require.NoError(t, err)
			assert.Equal(t, tt.wantNet, net, "net")
			assert.Equal(t, tt.wantRules, got.Withholding.RuleIDs)
			assert.Equal(t, fiscal.WithholdingTableVersion, got.Withholding.TableVersion)
			assert.Equal(t, tt.wantWarnings, got.Warnings)
			assert.Equal(t, len(tt.wantWarnings) > 0 && tt.wantWarnings[0].Code == domain.WarningWithholdingCPAUnverified,
				got.Withholding.CPAUnverified)
		})
	}
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
	}{
		{
			name: "601 invoiced 2026-09-05, not collected: Por cobrar", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Counterparty = fiscal.ReceptorPFNoActivityOrPublic
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-05")}
			},
			wantBooked: "2026-09-05", wantPending: true,
		},
		{
			name: "601 collected later keeps the invoice month", payee: pm601,
			edit: func(d *domain.Draft) {
				d.Counterparty = fiscal.ReceptorPFNoActivityOrPublic
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
				d.ViaPlatform = true
			},
			wantBooked: "2026-09-12", wantActivity: domain.ActivityLodging,
		},
		{
			name: "advance collected before the invoice", payee: pf612,
			edit: func(d *domain.Draft) {
				d.Dates = domain.Dates{InvoicedOn: date(t, "2026-10-13"), CollectedOn: date(t, "2026-01-02")}
			},
			wantBooked: "2026-01-02",
		},
		{
			name: "informational regime", payee: payee(fiscal.PersonFisica, "605", domain.BasisCash),
			edit: func(d *domain.Draft) {
				d.IVARate, d.IVA = fiscal.IVAExempt, 0
				d.Operation = fiscal.OperationOther
			},
			wantBooked: "2026-09-12",
		},
		{
			name: "housing rent on another operation keeps nothing to withhold", payee: pf612,
			edit: func(d *domain.Draft) {
				d.IVARate, d.IVA, d.ResidentialRent = fiscal.IVAExempt, 0, true
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
			assert.True(t, got.Withholding.Derived())
		})
	}
}

// TestAssessRejects covers the 400 and 422 rows of MODULE-incomes §4 and
// MODULE-withholdings §7. Validation (400) runs first: a draft with
// schema-level problems never reaches the business rules.
func TestAssessRejects(t *testing.T) {
	twoActivities := pf625
	twoActivities.PlatformActivities = []domain.PlatformActivity{domain.ActivityTransportDelivery, domain.ActivityGoodsServices}
	platform := func(a domain.PlatformActivity) func(*domain.Draft) {
		return func(d *domain.Draft) {
			d.ViaPlatform = true
			d.PlatformActivity = a
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
		{"IVA above the maximum", pf612, func(d *domain.Draft) { d.IVA = math.MaxInt64 }, fiscal.ClassValidation,
			violations(domain.FieldIVACents, fiscal.ErrOutOfRange)},
		{"missing IVA rate", pf612, func(d *domain.Draft) { d.IVARate = "" }, fiscal.ClassValidation,
			violations(domain.FieldIVARate, fiscal.ErrRequired)},
		{"unknown IVA rate", pf612, func(d *domain.Draft) { d.IVARate = "rate_20" }, fiscal.ClassValidation,
			violations(domain.FieldIVARate, fiscal.ErrInvalid)},
		{"missing counterparty kind", pf612, func(d *domain.Draft) { d.Counterparty = "" }, fiscal.ClassValidation,
			violations(domain.FieldCounterpartyKind, fiscal.ErrRequired)},
		{"the old counterparty vocabulary", pf612, func(d *domain.Draft) { d.Counterparty = "persona_moral" }, fiscal.ClassValidation,
			violations(domain.FieldCounterpartyKind, fiscal.ErrInvalid)},
		{"missing operation kind", pf612, func(d *domain.Draft) { d.Operation = "" }, fiscal.ClassValidation,
			violations(domain.FieldOperationKind, fiscal.ErrRequired)},
		{"platform is not an operation on the wire", pf612, func(d *domain.Draft) { d.Operation = fiscal.OperationPlatform },
			fiscal.ClassValidation, violations(domain.FieldOperationKind, fiscal.ErrInvalid)},
		{"subcontracting was repealed", pf612, func(d *domain.Draft) { d.Operation = "subcontracting" }, fiscal.ClassValidation,
			violations(domain.FieldOperationKind, fiscal.ErrInvalid)},
		{"via a platform on a 612 income", pf612, func(d *domain.Draft) { d.ViaPlatform = true }, fiscal.ClassValidation,
			violations(domain.FieldViaPlatform, domain.ErrNotForRegime)},
		{"US-I1: 612 without collected_on", pf612, func(d *domain.Draft) { d.Dates = domain.Dates{} }, fiscal.ClassValidation,
			violations(domain.FieldCollectedOn, fiscal.ErrRequired)},
		{"601 without invoiced_on", pm601, func(*domain.Draft) {}, fiscal.ClassValidation,
			violations(domain.FieldInvoicedOn, fiscal.ErrRequired)},
		{"US-I1: collected tomorrow", pf612, func(d *domain.Draft) { d.Dates.CollectedOn = date(t, "2026-10-14") }, fiscal.ClassValidation,
			violations(domain.FieldCollectedOn, fiscal.ErrOutOfRange)},
		{"collected before 2022-01-01 (the floor; 2026 is the parameter check)", pf612, func(d *domain.Draft) { d.Dates.CollectedOn = date(t, "2021-12-31") }, fiscal.ClassValidation,
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
			d.Counterparty = ""
		}, fiscal.ClassValidation, fiscal.Violations{
			{Field: domain.FieldAmountCents, Err: fiscal.ErrMustBePositive},
			{Field: domain.FieldCounterpartyKind, Err: fiscal.ErrRequired},
			{Field: domain.FieldCollectedOn, Err: fiscal.ErrRequired},
		}},
		{"US-I1: exempt with IVA 1.00", pf612, func(d *domain.Draft) { d.IVARate, d.IVA = fiscal.IVAExempt, 1_00 },
			fiscal.ClassBusinessRule, violations(domain.FieldIVACents, fiscal.ErrIVAAmountMismatch)},
		{"US-I1: 16 % of 10,000 entered as 1,500", pf612, func(d *domain.Draft) { d.IVA = 1_500_00 },
			fiscal.ClassBusinessRule, violations(domain.FieldIVACents, fiscal.ErrIVAAmountMismatch)},
		{"housing rent with IVA (LIVA 20 fr. II)", pf606, func(d *domain.Draft) {
			d.Operation, d.ResidentialRent = fiscal.OperationRentRealEstate, true
		}, fiscal.ClassBusinessRule, violations(domain.FieldIVARate, domain.ErrResidentialRentExempt)},
		{"US-I1: 625 lodging not among the taxpayer's activities", twoActivities, platform(domain.ActivityLodging), fiscal.ClassBusinessRule,
			violations(domain.FieldPlatformActivity, domain.ErrActivityNotHeld)},
		{"every business rule is reported", pf606, func(d *domain.Draft) {
			d.IVA, d.Operation, d.ResidentialRent = 1_500_00, fiscal.OperationRentRealEstate, true
		}, fiscal.ClassBusinessRule, fiscal.Violations{
			{Field: domain.FieldIVACents, Err: fiscal.ErrIVAAmountMismatch},
			{Field: domain.FieldIVARate, Err: domain.ErrResidentialRentExempt},
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

// A persona moral may sell through a platform (TR §B.2.3).
func TestAssessViaPlatformIsForPlatformsAndPersonasMorales(t *testing.T) {
	for _, p := range []domain.Payee{pf625goods, pm601, pm626} {
		d := us1Draft(t)
		d.ViaPlatform = true
		d.Dates = domain.Dates{InvoicedOn: date(t, "2026-09-12"), CollectedOn: date(t, "2026-09-12")}
		_, err := assess(t, d, p)
		assert.NoError(t, err, p.Regime)
	}
}

// TestAssessRateNotValidOnDate is US-I1: a rate_8 income collected on
// 2027-01-05 with no 2027 parameter extending the border decree.
func TestAssessRateNotValidOnDate(t *testing.T) {
	d := us1Draft(t)
	d.IVARate, d.IVA = fiscal.IVARate8, 800_00
	d.Counterparty = fiscal.ReceptorPFNoActivityOrPublic
	d.Dates.CollectedOn = date(t, "2027-01-05")

	_, err := domain.Assess(d, pf612, date(t, "2027-02-01"), withholdingCatalog2026(t))
	require.ErrorIs(t, err, fiscal.ErrRateNotValidOnDate)
	got, ok := errors.AsType[fiscal.Violations](err)
	require.True(t, ok)
	assert.Equal(t, violations(domain.FieldIVARate, fiscal.ErrRateNotValidOnDate), got)
}

// TestAssessPendingIncomeRateDate: an accrual income still pending
// collection validates its IVA rate, and derives, on the invoice date
// (§2.5, MODULE-withholdings §4.1).
func TestAssessPendingIncomeRateDate(t *testing.T) {
	d := us1Draft(t)
	d.IVARate, d.IVA = fiscal.IVARate8, 800_00
	d.Counterparty = fiscal.ReceptorPFNoActivityOrPublic
	d.Dates = domain.Dates{InvoicedOn: date(t, "2026-12-31")}

	got, err := domain.Assess(d, pm601, date(t, "2027-02-01"), withholdingCatalog2026(t))
	require.NoError(t, err)
	assert.True(t, got.PendingCollection)

	d.Dates.CollectedOn = date(t, "2027-01-05")
	_, err = domain.Assess(d, pm601, date(t, "2027-02-01"), withholdingCatalog2026(t))
	require.ErrorIs(t, err, fiscal.ErrRateNotValidOnDate, "once collected, the collection date rules")
}

func TestAssessServerFaults(t *testing.T) {
	absurdIVA := withholdingCatalog2026(t)
	absurdIVA.fakeCatalog = catalog2026(t)
	absurdIVA.fakeCatalog[fiscal.KeyIVAGeneral] = []datedRate{{from: date(t, "2022-01-01"), rate: math.MaxInt64}}
	absurdISR := withholdingCatalog2026(t)
	absurdISR.fakeCatalog = catalog2026(t)
	absurdISR.fakeCatalog[fiscal.KeyISRWithholdingFees] = []datedRate{{from: date(t, "2022-01-01"), rate: math.MaxInt64}}
	noFeesRate := withholdingCatalog2026(t)
	noFeesRate.fakeCatalog = catalog2026(t)
	delete(noFeesRate.fakeCatalog, fiscal.KeyISRWithholdingFees)
	badBasis := pf612
	badBasis.Basis = "weekly"
	badPerson := pf612
	badPerson.PersonType = "trust"
	noRegime := pf612
	noRegime.Regime = ""
	tests := []struct {
		name    string
		payee   domain.Payee
		today   fiscal.Date
		catalog fiscal.WithholdingCatalog
	}{
		{"catalog IVA rate overflows", pf612, date(t, today), absurdIVA},
		{"catalog ISR rate overflows", pf612, date(t, today), absurdISR},
		{"no withholding rate on the date", pf612, date(t, today), noFeesRate},
		{"unknown income basis", badBasis, date(t, today), withholdingCatalog2026(t)},
		{"unknown person type", badPerson, date(t, today), withholdingCatalog2026(t)},
		{"a persona física payee without a regime", noRegime, date(t, today), withholdingCatalog2026(t)},
		{"no today from the clock", pf612, fiscal.Date{}, withholdingCatalog2026(t)},
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

// TestDefaultOperation is MODULE-withholdings §2.1 Q2's first-time default.
func TestDefaultOperation(t *testing.T) {
	tests := []struct {
		person fiscal.PersonType
		regime fiscal.RegimeCode
		want   fiscal.OperationKind
	}{
		{fiscal.PersonFisica, "612", fiscal.OperationProfessionalServices},
		{fiscal.PersonFisica, "606", fiscal.OperationRentRealEstate},
		{fiscal.PersonFisica, "626", fiscal.OperationProfessionalServices},
		{fiscal.PersonFisica, "621", fiscal.OperationSaleOfGoods},
		{fiscal.PersonFisica, "625", fiscal.OperationOther},
		{fiscal.PersonMoral, "601", fiscal.OperationBusinessServices},
		{fiscal.PersonMoral, "626", fiscal.OperationBusinessServices},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, domain.DefaultOperation(tt.person, tt.regime), "%s %s", tt.person, tt.regime)
	}
}

func TestDefaultViaPlatform(t *testing.T) {
	assert.True(t, domain.DefaultViaPlatform(fiscal.RegimePlatforms))
	assert.False(t, domain.DefaultViaPlatform(fiscal.RegimeBusiness))
	assert.False(t, domain.DefaultViaPlatform("601"))
}

// TestLegacyCounterparty is MODULE-withholdings §5.2 read backwards: the
// counterparty_type column (dropped by WH-08) mirrors the new facts, and
// only a payer that may withhold maps to persona_moral or plataforma (the
// incomes_withholding_needs_payer CHECK).
func TestLegacyCounterparty(t *testing.T) {
	tests := []struct {
		kind fiscal.ReceptorKind
		via  bool
		want domain.Counterparty
	}{
		{fiscal.ReceptorPM, false, domain.CounterpartyMoral},
		{fiscal.ReceptorGovFederal, false, domain.CounterpartyMoral},
		{fiscal.ReceptorGovStateMunicipal, false, domain.CounterpartyMoral},
		{fiscal.ReceptorPFWithActivity, false, domain.CounterpartyFisica},
		{fiscal.ReceptorPFNoActivityOrPublic, false, domain.CounterpartyPublic},
		{fiscal.ReceptorForeign, false, domain.CounterpartyPublic},
		{fiscal.ReceptorPFNoActivityOrPublic, true, domain.CounterpartyPlatform},
		{fiscal.ReceptorPM, true, domain.CounterpartyPlatform},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, domain.LegacyCounterparty(tt.kind, tt.via), "%s via=%v", tt.kind, tt.via)
	}
}

func TestMaxCents(t *testing.T) {
	assert.Equal(t, money.Cents(99_999_999_999_999), fiscal.MaxCents)
}

func violations(field string, err fiscal.Error) fiscal.Violations {
	return fiscal.Violations{{Field: field, Err: err}}
}
