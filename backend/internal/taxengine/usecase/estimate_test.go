package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

var (
	est612 = estKey(fiscal.RegimeBusiness, fiscal.PersonFisica)
	est606 = estKey(fiscal.RegimeRent, fiscal.PersonFisica)
	est601 = estKey("601", fiscal.PersonMoral)
	est605 = estKey("605", fiscal.PersonFisica)
	est621 = estKey("621", fiscal.PersonFisica)
)

// TX-NET-01 components (MODULE-tax-engine §3): 612 nets +2,133.33, 606
// (commercial rent with creditable IVA) −1,933.33.
var (
	estIVA612 = &domain.IVAComponents{ChargedCollected: 4_800_00, CreditablePaid: 1_600_00, WithheldFromTaxpayer: 1_066_67}
	estIVA606 = &domain.IVAComponents{ChargedCollected: 3_200_00, CreditablePaid: 3_000_00, WithheldFromTaxpayer: 2_133_33}
)

func disclaimerCodesOf(ds []domain.Disclaimer) []domain.DisclaimerCode {
	out := make([]domain.DisclaimerCode, 0, len(ds))
	for _, d := range ds {
		out = append(out, d.Code)
	}
	return out
}

func requireCode(t *testing.T, err error, code apperr.Code) []apperr.FieldError {
	t.Helper()
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, app.Code)
	return app.Fields
}

// A PF with 612 and 606 in March: one consolidated IVA (OQ-T2) with the
// cross-regime warning, the 606 favor balance kept out of the ISR total
// (OQ-T10), regimes in catalog (code) order.
func TestEstimateMonthNetsIVAAcrossRegimes(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull))
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, ISR: estISR(150_96), IVA: estIVA612},
	}}
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, ISR: estISR(-924_00), IVA: estIVA606},
	}}

	got, err := f.estimator(s612, s606).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Equal(t, "2026-03", got.Period.String())
	require.Len(t, got.Regimes, 2)
	assert.Equal(t, fiscal.RegimeRent, got.Regimes[0].Key.Regime, "catalog order: 606 before 612")
	assert.Equal(t, fiscal.RegimeBusiness, got.Regimes[1].Key.Regime)
	for _, r := range got.Regimes {
		assert.Equal(t, domain.StatusComplete, r.Status)
		assert.Equal(t, domain.TreatmentFull, r.Treatment)
		assert.Equal(t, []domain.CPATopic{domain.TopicCrossRegimeIVANetting}, r.CPATopics,
			"each contributing regime carries the netting topic (contract TX-NET-01)")
	}
	assert.Equal(t, money.Cents(200_00), got.Totals.IVA.Amount)
	assert.Equal(t, money.Cents(150_96), got.Totals.ISR.Amount)
	assert.Equal(t, money.Cents(350_96), got.Totals.TotalPayable)
	assert.Equal(t, 2, got.IVA.Contributors)
	assert.Equal(t, []domain.Disclaimer{
		{Code: domain.DisclaimerNotTaxAdvice, Severity: domain.SeverityInfo},
		{Code: domain.DisclaimerCPAReviewPending, Severity: domain.SeverityStrong, Topics: []domain.CPATopic{domain.TopicCrossRegimeIVANetting}},
	}, got.Disclaimers)
	assert.Equal(t, domain.Legend, got.Legend)
	assert.Equal(t, domain.EngineVersion, got.EngineVersion)
	assert.Equal(t, taxparams.SetID{FiscalYear: 2026, Version: 3}, got.ParameterSet)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
	assert.Equal(t, fiscal.PersonFisica, got.PersonType)
}

// Each regime carries its catalog texts (name, explanation, input labels)
// for the presenter; a regime the catalog read misses gets none.
func TestEstimateMonthAttachesTheCatalogTexts(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold("605", "2026-01", domain.TreatmentInformational))
	text612 := usecase.RegimeText{
		Name:        "Actividades Empresariales y Profesionales",
		InputLabels: map[domain.InputKey]string{domain.InputPriorProvisionalPayments: "Pagos provisionales anteriores del ejercicio"},
	}
	f.texts.texts = map[fiscal.RegimeCode]usecase.RegimeText{fiscal.RegimeBusiness: text612}
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull}

	got, err := f.estimator(s612).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	require.Len(t, got.Regimes, 2)
	assert.Equal(t, usecase.RegimeText{}, got.Regimes[0].Text)
	assert.Equal(t, text612, got.Regimes[1].Text)
}

// A single regime contributing IVA is not netted: no topic is added.
func TestEstimateMonthOneIVARegimeHasNoNettingTopic(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, ISR: estISR(150_96), IVA: estIVA612},
	}}

	got, err := f.estimator(s612).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Empty(t, got.Regimes[0].CPATopics)
	assert.False(t, got.Regimes[0].CPAReviewPending())
}

// TX-601-REMIT-01: the PM's withholdings to remit are counted once, apart.
func TestEstimateMonthWithholdingsToRemit(t *testing.T) {
	f := newEstFixture(fiscal.PersonMoral, estHold("601", "2026-01", domain.TreatmentFull))
	s601 := &estStrategy{key: est601, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.June: {
			Coverage: domain.TreatmentFull, ISR: estISR(36_000_00),
			IVA: &domain.IVAComponents{ChargedCollected: 320_000_00, CreditablePaid: 200_000_00}, WithheldToRemit: 10_333_33,
		},
	}}

	got, err := f.estimator(s601).EstimateMonth(t.Context(), userA, taxpayerA, "2026-06")

	require.NoError(t, err)
	assert.Equal(t, money.Cents(120_000_00), got.Totals.IVA.Amount)
	assert.Equal(t, money.Cents(36_000_00), got.Totals.ISR.Amount)
	assert.Equal(t, money.Cents(10_333_33), got.Totals.WithholdingsToRemit)
	assert.Equal(t, money.Cents(166_333_33), got.Totals.TotalPayable)
}

// PR #78: an incomplete regime still shows its figures, names what is
// missing, and the estimate carries incomplete_inputs ("se usaron ceros").
// The strategy sees the months the regime was held up to the period.
func TestEstimateMonthIncompleteInputs(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-02", domain.TreatmentFull))
	s612 := &estStrategy{
		key: est612, treatment: domain.TreatmentFull,
		required: []domain.RequiredInput{{
			Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: []time.Month{time.February},
		}},
		byMonth: map[time.Month]domain.RegimeEstimate{time.March: {Coverage: domain.TreatmentFull, ISR: estISR(4_150_96)}},
	}

	got, err := f.estimator(s612).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	require.Len(t, got.Regimes, 1)
	r := got.Regimes[0]
	assert.Equal(t, domain.StatusIncomplete, r.Status)
	assert.Equal(t, []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.February}}}, r.MissingInputs)
	assert.Equal(t, money.Cents(4_150_96), r.ISR.Amount, "computed with zero for the missing input")
	assert.Contains(t, disclaimerCodesOf(got.Disclaimers), domain.DisclaimerIncompleteInputs)
	assert.Equal(t, []time.Month{time.February, time.March}, s612.facts[0].ActiveMonths)
}

// The tax inputs of each regime reach its strategy; another regime's do
// not.
func TestEstimateMonthPassesEachRegimeItsInputs(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	mine, err := domain.NewTaxInputs(domain.TaxInputValue{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: 7_00})
	require.NoError(t, err)
	theirs, err := domain.NewTaxInputs(domain.TaxInputValue{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: 9_00})
	require.NoError(t, err)
	f.inputs.values = map[fiscal.RegimeCode]domain.TaxInputs{fiscal.RegimeBusiness: mine, fiscal.RegimeRent: theirs}
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull}

	_, err = f.estimator(s612).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	ptu, ok := s612.ran[0].Inputs.Money(domain.InputPTUPaid)
	assert.True(t, ok)
	assert.Equal(t, money.Cents(7_00), ptu)
}

// The strategy reads its regime's sums, January..the period; a regime with
// no records gets zero sums that still cover the year to date.
func TestEstimateMonthPassesEachRegimeItsSums(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull))
	f.ledger.incomes = []usecase.IncomeMonthSums{
		{Regime: fiscal.RegimeBusiness, Month: time.January, Sums: domain.IncomeSums{Cash: 1_00}},
		{Regime: fiscal.RegimeBusiness, Month: time.March, Sums: domain.IncomeSums{Cash: 3_00}},
	}
	f.ledger.expenses = []usecase.ExpenseMonthSums{
		{Regime: fiscal.RegimeBusiness, Month: time.March, Sums: domain.ExpenseSums{DeductiblePaid: 2_00}},
	}
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull}
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull}

	_, err := f.estimator(s612, s606).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Equal(t, usecase.LedgerQuery{TaxpayerID: taxpayerA, FiscalYear: 2026, Through: time.March}, f.ledger.queries[0])
	ytd, err := s612.ran[0].Sums.Sum(time.January, time.March)
	require.NoError(t, err)
	assert.Equal(t, money.Cents(4_00), ytd.Income.Cash)
	assert.Equal(t, money.Cents(2_00), ytd.Expenses.DeductiblePaid)
	empty, err := s606.ran[0].Sums.Sum(time.January, time.March)
	require.NoError(t, err, "a regime without records still has January..March")
	assert.Equal(t, domain.MonthSums{}, empty)
}

// US-X1: a regime without a strategy, or that the catalog marks
// informational, shows its explanation and no number; records in it never
// change the totals (MODULE-incomes OQ-4).
func TestEstimateMonthInformationalRegimes(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold("605", "2026-01", domain.TreatmentInformational),
		estHold(fiscal.RegimePlatforms, "2026-01", domain.TreatmentFull))
	f.ledger.incomes = []usecase.IncomeMonthSums{{Regime: "605", Month: time.March, Sums: domain.IncomeSums{
		Cash: 50_000_00, ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 50_000_00, IVA: 8_000_00}},
	}}}
	// A strategy registered for 605 is ignored: the catalog says informational.
	s605 := &estStrategy{key: est605, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, ISR: estISR(1_00), IVA: estIVA612},
	}}
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, ISR: estISR(150_96), IVA: estIVA612},
	}}

	got, err := f.estimator(s612, s605).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	require.Len(t, got.Regimes, 3)
	info605, full612, fallback625 := got.Regimes[0], got.Regimes[1], got.Regimes[2]
	assert.Equal(t, domain.StatusInformational, info605.Status)
	assert.Nil(t, info605.ISR)
	assert.Nil(t, info605.IVA)
	assert.Equal(t, domain.RegimeExplanation(est605), info605.Explanation)
	assert.Empty(t, s605.ran, "the catalog's informational treatment wins")
	assert.Equal(t, domain.StatusComplete, full612.Status)
	assert.Equal(t, domain.StatusInformational, fallback625.Status, "no 625 strategy registered: the fallback")
	assert.Equal(t, domain.TreatmentFull, fallback625.Treatment, "the catalog treatment is reported as is")
	assert.Equal(t, money.Cents(2_133_33), got.Totals.IVA.Amount)
	assert.Equal(t, money.Cents(150_96), got.Totals.ISR.Amount)
	assert.Contains(t, disclaimerCodesOf(got.Disclaimers), domain.DisclaimerInformationalRegime)
}

// A partial regime (622, 603, 621) is partial and raises partial_regime.
func TestEstimateMonthPartialRegime(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold("621", "2026-01", domain.TreatmentPartial))
	s621 := &estStrategy{key: est621, treatment: domain.TreatmentPartial, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentPartial, CPATopics: []domain.CPATopic{domain.TopicRIFReduction}},
	}}

	got, err := f.estimator(s621).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Equal(t, domain.StatusPartial, got.Regimes[0].Status)
	assert.Equal(t, []domain.DisclaimerCode{
		domain.DisclaimerNotTaxAdvice, domain.DisclaimerCPAReviewPending, domain.DisclaimerPartialRegime,
	}, disclaimerCodesOf(got.Disclaimers))
}

// §2.1: only regimes whose window contains the period; an archived
// taxpayer's past months stay readable; no regime in the period is an
// empty estimate, not an error.
func TestEstimateMonthOnlyRegimesHeldInThePeriod(t *testing.T) {
	ended := estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull)
	ended.End = mustPeriod("2026-02")
	later := estHold(fiscal.RegimeBusiness, "2026-04", domain.TreatmentFull)
	f := newEstFixture(fiscal.PersonFisica, ended, later)
	f.profiles.profile.Archived = true
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull}
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull}

	got, err := f.estimator(s612, s606).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Empty(t, got.Regimes)
	assert.Equal(t, domain.OutcomeZero, got.Totals.IVA.Outcome())
	assert.Zero(t, got.Totals.TotalPayable)
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerNotTaxAdvice}, disclaimerCodesOf(got.Disclaimers))
	assert.Empty(t, got.IVA.Lines)
}

// OQ-T5: the previous month's saldo a favor is not carried forward; the
// consolidated IVA says so in an info line and the result is unchanged.
func TestEstimateMonthPreviousMonthInFavorIsNotCarried(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull))
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.February: {Coverage: domain.TreatmentFull, IVA: estIVA606},
		time.March:    {Coverage: domain.TreatmentFull, IVA: estIVA612},
	}}

	got, err := f.estimator(s606).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Equal(t, money.Cents(2_133_33), got.Totals.IVA.Amount)
	last := got.IVA.Lines[len(got.IVA.Lines)-1]
	assert.Equal(t, domain.LineIVAPreviousInFavor, last.Code)
	assert.Equal(t, money.Cents(1_933_33), last.Amount)
	assert.Equal(t, time.February, s606.ran[len(s606.ran)-1].Period.Month(), "February ran on its own year to date")
	assert.Equal(t, time.February, s606.ran[len(s606.ran)-1].Sums.Through())
}

// January has no previous month in the fiscal year: no info line.
func TestEstimateMonthJanuaryHasNoPreviousMonth(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull))
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.January: {Coverage: domain.TreatmentFull, IVA: estIVA612},
	}}

	got, err := f.estimator(s606).EstimateMonth(t.Context(), userA, taxpayerA, "2026-01")

	require.NoError(t, err)
	assert.Len(t, s606.ran, 1)
	for _, l := range got.IVA.Lines {
		assert.NotEqual(t, domain.LineIVAPreviousInFavor, l.Code)
	}
}

// PR #79: a taxed regime crediting IVA plus another with exempt
// collections is mixed activity at the taxpayer level.
func TestEstimateMonthMixedActivityAcrossRegimes(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull))
	f.ledger.incomes = []usecase.IncomeMonthSums{{Regime: fiscal.RegimeRent, Month: time.March, Sums: domain.IncomeSums{
		Cash: 15_000_00, ByRate: domain.IVAByRate{ExemptBase: 15_000_00},
	}}}
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, IVA: estIVA612},
	}}
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, IVA: &domain.IVAComponents{}},
	}}

	got, err := f.estimator(s612, s606).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Contains(t, disclaimerCodesOf(got.Disclaimers), domain.DisclaimerIVAMixedActivity)
}

// ADR-0002 §3: a fiscal year without its own set falls back, provisional.
func TestEstimateMonthProvisionalParameters(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica)
	f.params.provisional = true

	got, err := f.estimator().EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	require.NoError(t, err)
	assert.Contains(t, disclaimerCodesOf(got.Disclaimers), domain.DisclaimerProvisionalParameters)
}

// §2.10: no period is the current month in Mexico City (14 October 2026).
func TestEstimateMonthDefaultsToTheCurrentMonth(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull}

	got, err := f.estimator(s612).EstimateMonth(t.Context(), userA, taxpayerA, "")

	require.NoError(t, err)
	assert.Equal(t, "2026-10", got.Period.String())
	assert.Equal(t, time.October, f.ledger.queries[0].Through)
}

// §2.10, §5: the refusals, in order, before any port is read: the taxpayer
// id (404), the period's format (400 invalid), a future month (400
// out_of_range), a month before the first published year (422
// parameters_unavailable).
func TestEstimateMonthRefusals(t *testing.T) {
	tests := []struct {
		name       string
		taxpayer   string
		period     string
		params     estParams
		wantCode   apperr.Code
		wantDetail string
	}{
		{name: "malformed taxpayer id", taxpayer: "not-a-uuid", period: "2026-03", wantCode: apperr.CodeNotFound},
		{name: "malformed period", taxpayer: taxpayerA, period: "2026-13", wantCode: apperr.CodeValidationFailed, wantDetail: usecase.DetailInvalid},
		{name: "period with a day", taxpayer: taxpayerA, period: "2026-03-01", wantCode: apperr.CodeValidationFailed, wantDetail: usecase.DetailInvalid},
		{name: "next month", taxpayer: taxpayerA, period: "2026-11", wantCode: apperr.CodeValidationFailed, wantDetail: usecase.DetailOutOfRange},
		{name: "next year", taxpayer: taxpayerA, period: "2027-01", wantCode: apperr.CodeValidationFailed, wantDetail: usecase.DetailOutOfRange},
		{
			name: "before 2026", taxpayer: taxpayerA, period: "2025-12",
			wantCode: apperr.CodeBusinessRuleViolation, wantDetail: usecase.DetailParametersUnavailable,
		},
		{
			name: "no parameter set within reach", taxpayer: taxpayerA, period: "2026-03", params: estParams{err: taxparams.ErrNoParameters},
			wantCode: apperr.CodeBusinessRuleViolation, wantDetail: usecase.DetailParametersUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEstFixture(fiscal.PersonFisica)
			f.params = tt.params

			_, err := f.estimator().EstimateMonth(t.Context(), userA, tt.taxpayer, tt.period)

			fields := requireCode(t, err, tt.wantCode)
			if tt.wantDetail != "" {
				require.Len(t, fields, 1)
				assert.Equal(t, "period", fields[0].Field)
				assert.Equal(t, tt.wantDetail, fields[0].Code)
				assert.NotEmpty(t, fields[0].Message)
			}
			assert.Empty(t, f.tx.modes, "refused before the unit of work")
		})
	}
}

// The current, still-open month is allowed (§2.10).
func TestEstimateMonthAllowsTheCurrentMonth(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica)

	_, err := f.estimator().EstimateMonth(t.Context(), userA, taxpayerA, "2026-10")

	assert.NoError(t, err)
}

// S-2: every port runs inside the caller's read-only unit of work, as the
// caller; the taxpayer id is canonical.
func TestEstimateMonthReadsInOneReadOnlyUnitOfWork(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull}
	upper := "0199D4A2-1111-7000-8000-00000000000A"

	got, err := f.estimator(s612).EstimateMonth(t.Context(), userA, upper, "2026-03")

	require.NoError(t, err)
	assert.Equal(t, []string{"ro"}, f.tx.modes)
	assert.Equal(t, []string{userA + "|" + userA + "|" + taxpayerA}, f.profiles.users)
	assert.Equal(t, []string{userA, userA}, f.ledger.users)
	assert.Equal(t, []string{userA}, f.inputs.users)
	assert.Equal(t, []string{userA}, f.texts.users)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
}

// A port's failure: a client error (the profile's 404) passes as is;
// anything else is a wrapped infrastructure failure.
func TestEstimateMonthPortFailures(t *testing.T) {
	notFound := apperr.New(apperr.CodeNotFound, "Not found.")
	tests := []struct {
		name      string
		breakPort func(*estFixture)
		want      error
	}{
		{"profile 404", func(f *estFixture) { f.profiles.err = notFound }, notFound},
		{"ledger", func(f *estFixture) { f.ledger.err = errEstPortDown }, errEstPortDown},
		{"tax inputs", func(f *estFixture) { f.inputs.err = errEstPortDown }, errEstPortDown},
		{"regime texts", func(f *estFixture) { f.texts.err = errEstPortDown }, errEstPortDown},
		{"holidays", func(f *estFixture) { f.holidays.err = errEstPortDown }, errEstPortDown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
			tt.breakPort(f)

			_, err := f.estimator(&estStrategy{key: est612, treatment: domain.TreatmentFull}).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

			assert.ErrorIs(t, err, tt.want)
			if errors.Is(tt.want, errEstPortDown) {
				_, isApp := errors.AsType[*apperr.Error](err)
				assert.False(t, isApp, "an infrastructure failure is a 500")
			}
		})
	}
}

func TestEstimateMonthStrategyFailure(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull, err: errEstPortDown}

	_, err := f.estimator(s612).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")

	assert.ErrorIs(t, err, errEstPortDown)
}

// §2.10: day 17 of the next month, moved past weekends and holidays (CFF 12
// ⚠️); after the bimester when the taxpayer only holds 621.
func TestEstimateMonthDueDate(t *testing.T) {
	tests := []struct {
		name     string
		period   string
		regime   usecase.RegimeHold
		holidays map[string]bool
		want     string
	}{
		{"a weekday", "2026-03", estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull), nil, "2026-04-17"},
		{"Saturday moves to Monday", "2026-09", estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull), nil, "2026-10-19"},
		{
			"a holiday Monday moves to Tuesday", "2026-09", estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
			map[string]bool{"2026-10-19": true}, "2026-10-20",
		},
		{"January pays in February", "2026-01", estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull), nil, "2026-02-17"},
		{"621, the bimester's first month", "2026-09", estHold("621", "2026-01", domain.TreatmentPartial), nil, "2026-11-17"},
		{"621, the bimester's second month", "2026-10", estHold("621", "2026-01", domain.TreatmentPartial), nil, "2026-11-17"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEstFixture(fiscal.PersonFisica, tt.regime)
			f.holidays.days = tt.holidays

			got, err := f.estimator().EstimateMonth(t.Context(), userA, taxpayerA, tt.period)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.DueDate.String())
		})
	}
}

// ADR-0002 §5: the same records, inputs, parameters and engine give the
// same fingerprint; a changed record moves it.
func TestEstimateMonthRecordsFingerprint(t *testing.T) {
	run := func(cash money.Cents) string {
		f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
		f.ledger.incomes = []usecase.IncomeMonthSums{{Regime: fiscal.RegimeBusiness, Month: time.March, Sums: domain.IncomeSums{Cash: cash}}}
		got, err := f.estimator(&estStrategy{key: est612, treatment: domain.TreatmentFull}).EstimateMonth(t.Context(), userA, taxpayerA, "2026-03")
		require.NoError(t, err)
		return got.RecordsFingerprint
	}

	first := run(1_00)

	assert.Regexp(t, `^[0-9a-f]{64}$`, first)
	assert.Equal(t, first, run(1_00))
	assert.NotEqual(t, first, run(1_01))
}
