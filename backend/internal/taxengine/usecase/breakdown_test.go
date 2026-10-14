package usecase_test

import (
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

// bdLine is a strategy line of the fake estimates: only what ordering and
// numbering read.
func bdLine(seq int, tax domain.TaxKind, code domain.LineCode) domain.Line {
	return domain.Line{Seq: seq, Tax: tax, Code: code, Label: domain.LabelKey(code), Operation: domain.OpInput}
}

func lineCodes(lines []domain.Line) []domain.LineCode {
	out := make([]domain.LineCode, 0, len(lines))
	for _, l := range lines {
		out = append(out, l.Code)
	}
	return out
}

func lineSeqs(lines []domain.Line) []int {
	out := make([]int, 0, len(lines))
	for _, l := range lines {
		out = append(out, l.Seq)
	}
	return out
}

// bdNetting is TX-NET-01: 612 and 606 both contribute IVA in March; 612
// also has a withholding to remit, to see where those lines go.
func bdNetting() (*estFixture, []domain.Strategy) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull))
	s612 := &estStrategy{key: est612, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {
			Coverage: domain.TreatmentFull, ISR: estISR(150_96), IVA: estIVA612,
			Lines: []domain.Line{
				bdLine(1, domain.TaxISR, "isr.income_ytd"),
				bdLine(2, domain.TaxISR, "isr.payable"),
				bdLine(3, domain.TaxIVA, domain.LineIVACharged),
				bdLine(4, domain.TaxWithholding, domain.LineIVAWithheldToRemit),
				bdLine(5, domain.TaxIVA, domain.LineIVANet),
			},
		},
	}}
	s606 := &estStrategy{key: est606, treatment: domain.TreatmentFull, byMonth: map[time.Month]domain.RegimeEstimate{
		time.March: {Coverage: domain.TreatmentFull, ISR: estISR(-924_00), IVA: estIVA606},
	}}
	return f, []domain.Strategy{s612, s606}
}

// ADR-0002 §1 / contract getEstimateBreakdown: the regime's ISR and IVA
// lines, then the taxpayer's consolidated IVA lines, then the withholdings
// to remit, renumbered 1..n; the header is the estimate's.
func TestBreakdownOrdersTheLines(t *testing.T) {
	f, strategies := bdNetting()

	got, err := f.estimator(strategies...).Breakdown(t.Context(), userA, taxpayerA, "2026-03", "612")

	require.NoError(t, err)
	assert.Equal(t, []domain.LineCode{
		"isr.income_ytd", "isr.payable", domain.LineIVACharged, domain.LineIVANet,
		domain.LineIVATotalCharged, domain.LineIVATotalCreditable, domain.LineIVATotalWithheld, domain.LineIVAConsolidated,
		domain.LineIVAWithheldToRemit,
	}, lineCodes(got.Lines))
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, lineSeqs(got.Lines))
	assert.True(t, got.Lines[7].NeedsCPAReview, "two regimes contribute: the consolidated line is under CPA review")
	assert.Equal(t, money.Cents(200_00), got.Lines[7].Amount)

	assert.Equal(t, fiscal.RegimeBusiness, got.Regime.Key.Regime)
	assert.Equal(t, []domain.CPATopic{domain.TopicCrossRegimeIVANetting}, got.Regime.CPATopics)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
	assert.Equal(t, "2026-03", got.Period.String())
	assert.Equal(t, domain.Legend, got.Legend)
	assert.Equal(t, domain.EngineVersion, got.EngineVersion)
	assert.Equal(t, taxparams.SetID{FiscalYear: 2026, Version: 3}, got.ParameterSet)
	assert.Equal(t, fiscal.PersonFisica, got.PersonType)
	assert.Equal(t, []domain.Disclaimer{
		{Code: domain.DisclaimerNotTaxAdvice, Severity: domain.SeverityInfo},
		{Code: domain.DisclaimerCPAReviewPending, Severity: domain.SeverityStrong, Topics: []domain.CPATopic{domain.TopicCrossRegimeIVANetting}},
	}, got.Disclaimers)
}

// A regime with no IVA shows no consolidated lines and none of the
// consolidated IVA's disclaimers; another regime's status does not leak
// into it (contract Breakdown.disclaimers: those bearing on these lines).
func TestBreakdownOfARegimeWithoutIVA(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica,
		estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull),
		estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull),
		estHold("605", "2026-01", domain.TreatmentInformational))
	_, strategies := bdNetting()

	got, err := f.estimator(strategies...).Breakdown(t.Context(), userA, taxpayerA, "2026-03", "605")

	require.NoError(t, err)
	assert.Empty(t, got.Lines, "an informational regime answers its explanation with no lines")
	assert.Equal(t, domain.StatusInformational, got.Regime.Status)
	assert.Equal(t, []domain.DisclaimerCode{domain.DisclaimerNotTaxAdvice, domain.DisclaimerInformationalRegime},
		disclaimerCodesOf(got.Disclaimers))
}

// An incomplete regime's breakdown carries incomplete_inputs; the
// provisional parameters travel with every breakdown.
func TestBreakdownDisclaimersOfTheRegime(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-02", domain.TreatmentFull))
	f.params = estParams{provisional: true}
	s612 := &estStrategy{
		key: est612, treatment: domain.TreatmentFull,
		required: []domain.RequiredInput{{
			Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: []time.Month{time.February},
		}},
		byMonth: map[time.Month]domain.RegimeEstimate{time.March: {Coverage: domain.TreatmentFull, ISR: estISR(4_150_96)}},
	}

	got, err := f.estimator(s612).Breakdown(t.Context(), userA, taxpayerA, "2026-03", "612")

	require.NoError(t, err)
	assert.Equal(t, domain.StatusIncomplete, got.Regime.Status)
	assert.Equal(t, []domain.DisclaimerCode{
		domain.DisclaimerNotTaxAdvice, domain.DisclaimerIncompleteInputs, domain.DisclaimerProvisionalParameters,
	}, disclaimerCodesOf(got.Disclaimers))
}

// The refusals: the estimate's (404 taxpayer, 400 period, 422 parameters)
// first, then 422 regime_not_active on the regime query parameter for a
// regime not held in the period.
func TestBreakdownRefusals(t *testing.T) {
	ended := estHold(fiscal.RegimeRent, "2026-01", domain.TreatmentFull)
	ended.End = mustPeriod("2026-02")
	tests := []struct {
		name       string
		taxpayer   string
		period     string
		regime     string
		wantCode   apperr.Code
		wantField  string
		wantDetail string
	}{
		{name: "unknown taxpayer id", taxpayer: "not-a-uuid", period: "2026-03", regime: "612", wantCode: apperr.CodeNotFound},
		{
			name: "future month", taxpayer: taxpayerA, period: "2026-11", regime: "612",
			wantCode: apperr.CodeValidationFailed, wantField: "period", wantDetail: usecase.DetailOutOfRange,
		},
		{
			name: "before 2026", taxpayer: taxpayerA, period: "2025-12", regime: "612",
			wantCode: apperr.CodeBusinessRuleViolation, wantField: "period", wantDetail: usecase.DetailParametersUnavailable,
		},
		{
			name: "never held", taxpayer: taxpayerA, period: "2026-03", regime: "626",
			wantCode: apperr.CodeBusinessRuleViolation, wantField: "regime", wantDetail: usecase.DetailRegimeNotActive,
		},
		{
			name: "held, but not in that month", taxpayer: taxpayerA, period: "2026-03", regime: "606",
			wantCode: apperr.CodeBusinessRuleViolation, wantField: "regime", wantDetail: usecase.DetailRegimeNotActive,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull), ended)

			_, err := f.estimator(&estStrategy{key: est612, treatment: domain.TreatmentFull}).
				Breakdown(t.Context(), userA, tt.taxpayer, tt.period, tt.regime)

			fields := requireCode(t, err, tt.wantCode)
			if tt.wantDetail == "" {
				return
			}
			require.Len(t, fields, 1)
			assert.Equal(t, tt.wantField, fields[0].Field)
			assert.Equal(t, tt.wantDetail, fields[0].Code)
			assert.NotEmpty(t, fields[0].Message)
		})
	}
}

// A port failure is the estimate's: a wrapped infrastructure error.
func TestBreakdownPortFailure(t *testing.T) {
	f := newEstFixture(fiscal.PersonFisica, estHold(fiscal.RegimeBusiness, "2026-01", domain.TreatmentFull))
	f.ledger.err = errEstPortDown

	_, err := f.estimator(&estStrategy{key: est612, treatment: domain.TreatmentFull}).
		Breakdown(t.Context(), userA, taxpayerA, "2026-03", "612")

	assert.ErrorIs(t, err, errEstPortDown)
}
