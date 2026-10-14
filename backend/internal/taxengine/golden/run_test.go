package golden_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

var key612 = domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}

// fakeStrategy returns a canned estimate and records what it was given.
type fakeStrategy struct {
	estimate  domain.RegimeEstimate
	err       error
	required  []domain.RequiredInput
	called    bool
	gotFacts  domain.Facts
	gotParams *taxparams.Params
}

func (f *fakeStrategy) Key() domain.StrategyKey     { return key612 }
func (f *fakeStrategy) Treatment() domain.Treatment { return domain.TreatmentFull }
func (f *fakeStrategy) RequiredInputs(_ domain.Options, _ domain.Period, facts domain.Facts) []domain.RequiredInput {
	f.gotFacts = facts
	return f.required
}

func (f *fakeStrategy) Estimate(_ domain.Input, p *taxparams.Params) (domain.RegimeEstimate, error) {
	f.called, f.gotParams = true, p
	return f.estimate, f.err
}

// registryOf is the real M5-10 registry over the given strategies.
func registryOf(t *testing.T, strategies ...domain.Strategy) domain.Registry {
	t.Helper()
	r, err := domain.NewRegistry(strategies...)
	require.NoError(t, err)
	return r
}

type fakeSelector struct {
	params *taxparams.Params
	err    error
	year   int
}

func (s *fakeSelector) Select(year int) (taxparams.Selection, error) {
	s.year = year
	return taxparams.Selection{Params: s.params, RequestedYear: year}, s.err
}

var cite = []domain.Citation{{Law: "LISR", Article: "106"}}

// matching is an estimate that satisfies baseCase's expectations.
func matching() domain.RegimeEstimate {
	isr := domain.NewTaxResult(money.Exact(15096) * money.ExactPerCent)
	return domain.RegimeEstimate{
		Key: key612, Coverage: domain.TreatmentFull, ISR: &isr,
		IVA: &domain.IVAComponents{ChargedCollected: 480000, CreditablePaid: 160000, WithheldFromTaxpayer: 106667},
		Lines: []domain.Line{
			{Seq: 1, Tax: domain.TaxISR, Code: "isr.income_ytd", Operation: domain.OpInput, Amount: 9000000, Basis: cite},
			{Seq: 2, Tax: domain.TaxISR, Code: "isr.note", Operation: domain.OpInfo, Amount: 0},
			{Seq: 3, Tax: domain.TaxIVA, Code: "iva.net", Operation: domain.OpResult, Amount: 213333, Basis: cite},
		},
		Disclaimers: []domain.DisclaimerCode{domain.DisclaimerDeductibilityClaimed},
	}
}

func baseCase(t *testing.T) golden.Case {
	t.Helper()
	c, err := golden.Decode([]byte(validCase), "612-pf/PF-612-ISR-01.yaml")
	require.NoError(t, err)
	c.PendingStrategy = false
	c.Want.Lines = []golden.LineWant{
		{Tax: domain.TaxISR, Code: "isr.income_ytd", Op: domain.OpInput, Amount: 9000000},
		{Tax: domain.TaxISR, Code: "isr.note", Op: domain.OpInfo, Amount: 0},
		{Tax: domain.TaxIVA, Code: "iva.net", Op: domain.OpResult, Amount: 213333},
	}
	return c
}

func evaluate(t *testing.T, c golden.Case, s *fakeStrategy) golden.Outcome {
	t.Helper()
	reg := registryOf(t)
	if s != nil {
		reg = registryOf(t, s)
	}
	return golden.Evaluate(c, reg, &fakeSelector{params: &taxparams.Params{}})
}

func TestEvaluatePassesAMatchingEstimate(t *testing.T) {
	s := &fakeStrategy{estimate: matching()}
	sel := &fakeSelector{params: &taxparams.Params{}}

	o := golden.Evaluate(baseCase(t), registryOf(t, s), sel)

	assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Mismatches)
	assert.Empty(t, o.Mismatches)
	assert.Equal(t, 2026, sel.year, "parameters are selected for the period's fiscal year")
	assert.Same(t, sel.params, s.gotParams)
}

func TestEvaluateSkipsAnUnverifiedCaseWithItsReason(t *testing.T) {
	c := baseCase(t)
	c.SkipReason = "⚠️ 18-M effect not confirmed"
	s := &fakeStrategy{estimate: matching()}

	o := evaluate(t, c, s)

	assert.Equal(t, golden.VerdictSkipped, o.Verdict)
	assert.Contains(t, o.Reason, "⚠️ 18-M effect not confirmed")
	assert.False(t, s.called, "a skipped case never runs")
}

func TestEvaluateStrategyRegistrationRules(t *testing.T) {
	tests := []struct {
		name       string
		pending    bool
		generated  bool
		registered bool
		want       golden.Verdict
		reason     string
	}{
		{"marked pending, no strategy", true, false, false, golden.VerdictPending, "pending_strategy: no strategy registered for 612/fisica"},
		{"not marked, no strategy: nothing would gate it", false, false, false, golden.VerdictFailed, "not marked pending_strategy"},
		{"marked pending, strategy registered: stale marker", true, false, true, golden.VerdictFailed, "remove pending_strategy"},
		{"generated, no strategy", false, true, false, golden.VerdictPending, "pending_strategy"},
		{"generated, strategy registered", false, true, true, golden.VerdictPassed, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := baseCase(t)
			c.PendingStrategy, c.Generated = tt.pending, tt.generated
			var s *fakeStrategy
			if tt.registered {
				s = &fakeStrategy{estimate: matching()}
			}

			o := evaluate(t, c, s)

			assert.Equal(t, tt.want, o.Verdict, o.Reason)
			assert.Contains(t, o.Reason, tt.reason)
		})
	}
}

func TestEvaluateComparesTheExplanationWhenGiven(t *testing.T) {
	informational := func(t *testing.T, explanation domain.LabelKey) golden.Case {
		t.Helper()
		c := baseCase(t)
		c.Input.Key = domain.StrategyKey{Regime: "605", PersonType: fiscal.PersonFisica}
		c.Want = golden.Expected{
			Status: domain.StatusInformational, Taxes: []domain.TaxKind{domain.TaxISR, domain.TaxIVA},
			Lines: []golden.LineWant{}, Explanation: explanation,
		}
		return c
	}

	t.Run("match", func(t *testing.T) {
		o := evaluate(t, informational(t, "regime_explanation:605/fisica"), nil)

		assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Reason, o.Mismatches)
	})
	t.Run("mismatch", func(t *testing.T) {
		o := evaluate(t, informational(t, "regime_explanation:606/fisica"), nil)

		assert.Equal(t, golden.VerdictFailed, o.Verdict)
		assert.Contains(t, o.Mismatches,
			"explanation: want regime_explanation:606/fisica, got regime_explanation:605/fisica")
	})
}

func TestEvaluateRunsAnInformationalCaseThroughTheFallback(t *testing.T) {
	c := baseCase(t)
	c.Input.Key = domain.StrategyKey{Regime: "605", PersonType: fiscal.PersonFisica}
	c.Want = golden.Expected{
		Status: domain.StatusInformational, Taxes: []domain.TaxKind{domain.TaxISR, domain.TaxIVA},
		Lines: []golden.LineWant{}, MissingInputs: []domain.MissingInput{},
		CPATopics: []domain.CPATopic{}, Disclaimers: []domain.DisclaimerCode{},
	}

	o := evaluate(t, c, nil)

	assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Reason, o.Mismatches)
}

func TestEvaluateFailsAnEstimateForAnotherKey(t *testing.T) {
	got := matching()
	got.Key = domain.StrategyKey{Regime: "606", PersonType: fiscal.PersonFisica}

	o := evaluate(t, baseCase(t), &fakeStrategy{estimate: got})

	assert.Equal(t, golden.VerdictFailed, o.Verdict)
	assert.Contains(t, o.Reason, string(domain.ErrStrategyMismatch))
}

func TestEvaluateFailsOnEngineErrors(t *testing.T) {
	t.Run("parameters", func(t *testing.T) {
		s := &fakeStrategy{estimate: matching()}
		o := golden.Evaluate(baseCase(t), registryOf(t, s), &fakeSelector{err: taxparams.ErrNoParameters})

		assert.Equal(t, golden.VerdictFailed, o.Verdict)
		assert.Contains(t, o.Reason, "no parameters")
		assert.False(t, s.called)
	})
	t.Run("estimate", func(t *testing.T) {
		o := evaluate(t, baseCase(t), &fakeStrategy{err: errors.New("boom")})

		assert.Equal(t, golden.VerdictFailed, o.Verdict)
		assert.Contains(t, o.Reason, "boom")
	})
}

func TestEvaluateReportsEveryMismatch(t *testing.T) {
	got := matching()
	other := domain.NewTaxResult(money.Exact(15097) * money.ExactPerCent)
	got.ISR = &other
	got.IVA = &domain.IVAComponents{ChargedCollected: 480000, CreditablePaid: 160000}
	got.Lines[0].Amount = 9000001
	got.Lines = append(got.Lines, domain.Line{Seq: 4, Tax: domain.TaxIVA, Code: "iva.extra", Operation: domain.OpAdd, Amount: 1})
	got.Disclaimers = nil
	got.CPATopics = []domain.CPATopic{domain.TopicCrossRegimeIVANetting}
	got.Coverage = domain.TreatmentPartial

	o := evaluate(t, baseCase(t), &fakeStrategy{estimate: got})

	require.Equal(t, golden.VerdictFailed, o.Verdict)
	report := strings.Join(o.Mismatches, "\n")
	for _, want := range []string{
		"status: want complete, got partial",
		"isr: want 150.96, got 150.97",
		"iva: want",
		"line 1: want isr isr.income_ytd input 90000.00, got isr isr.income_ytd input 90000.01",
		"line 4: unexpected iva iva.extra add 0.01",
		"disclaimers: want [deductibility_claimed], got []",
		"cpa_topics: want [], got [cross_regime_iva_netting]",
		"line 4 iva.extra: no legal basis",
	} {
		assert.Contains(t, report, want)
	}
}

func TestEvaluateReportsAMissingLine(t *testing.T) {
	got := matching()
	got.Lines = got.Lines[:2]

	o := evaluate(t, baseCase(t), &fakeStrategy{estimate: got})

	assert.Contains(t, strings.Join(o.Mismatches, "\n"), "line 3: missing iva iva.net result 2133.33")
}

func TestEvaluateComparesTheOperationOfALine(t *testing.T) {
	got := matching()
	got.Lines[0].Operation = domain.OpAdd

	o := evaluate(t, baseCase(t), &fakeStrategy{estimate: got})

	assert.Contains(t, strings.Join(o.Mismatches, "\n"),
		"line 1: want isr isr.income_ytd input 90000.00, got isr isr.income_ytd add 90000.00")
}

func TestEvaluateComparesOnlyTheAssertedTaxes(t *testing.T) {
	c := baseCase(t)
	c.Want.Taxes = []domain.TaxKind{domain.TaxISR}
	c.Want.IVA = nil
	c.Want.Lines = c.Want.Lines[:2]
	got := matching()
	got.IVA = &domain.IVAComponents{ChargedCollected: 1}
	got.Lines[2].Amount = 42
	got.WithheldToRemit = 99

	o := evaluate(t, c, &fakeStrategy{estimate: got})

	assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Mismatches)
}

func TestEvaluateAssertsAbsentFiguresAndWithholdings(t *testing.T) {
	c := baseCase(t)
	c.Want.Taxes = append(c.Want.Taxes, domain.TaxWithholding)
	c.Want.ISR = nil
	c.Want.WithheldToRemit = 533333
	got := matching()

	o := evaluate(t, c, &fakeStrategy{estimate: got})

	report := strings.Join(o.Mismatches, "\n")
	assert.Contains(t, report, "isr: want none, got 150.96")
	assert.Contains(t, report, "withheld_to_remit: want 5333.33, got 0.00")
}

func TestEvaluateLeavesNilListsUnasserted(t *testing.T) {
	c := baseCase(t)
	c.Generated = true
	c.Want.Lines, c.Want.Disclaimers, c.Want.CPATopics, c.Want.MissingInputs = nil, nil, nil, nil
	got := matching()
	got.Lines = nil
	got.Disclaimers = []domain.DisclaimerCode{domain.DisclaimerNotTaxAdvice}

	o := evaluate(t, c, &fakeStrategy{estimate: got})

	assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Mismatches)
}

func TestEvaluateDerivesMissingInputsFromRequiredInputs(t *testing.T) {
	c := baseCase(t) // has prior payments for January only, coeficiente and first_fiscal_year
	c.ActiveMonths = []time.Month{time.January, time.February, time.March}
	c.Want.Status = domain.StatusIncomplete
	c.Want.MissingInputs = []domain.MissingInput{
		{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.February}},
		{Key: domain.InputPTUPaid},
	}
	s := &fakeStrategy{estimate: matching(), required: []domain.RequiredInput{
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Months: []time.Month{time.January, time.February}},
		{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate},
		{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean},
		{Key: domain.InputPTUPaid, Kind: domain.InputMoney},
	}}

	o := evaluate(t, c, s)

	assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Mismatches)
	assert.Equal(t, c.ActiveMonths, s.gotFacts.ActiveMonths)
	_, ok := s.gotFacts.Inputs.Rate(domain.InputCoeficienteUtilidad)
	assert.True(t, ok, "facts carry the case's inputs")
}

// recordingTB captures what Report does; any other testing.TB method panics.
type recordingTB struct {
	testing.TB
	skipped, errored string
}

func (r *recordingTB) Helper()                           {}
func (r *recordingTB) Skip(args ...any)                  { r.skipped = fmt.Sprint(args...) }
func (r *recordingTB) Errorf(format string, args ...any) { r.errored = fmt.Sprintf(format, args...) }

func TestReportSkipsWithTheReasonAndFailsWithEveryMismatch(t *testing.T) {
	tests := []struct {
		name             string
		o                golden.Outcome
		skipped, errored string
	}{
		{"passed", golden.Outcome{Verdict: golden.VerdictPassed}, "", ""},
		{"skipped", golden.Outcome{Verdict: golden.VerdictSkipped, Reason: "⚠️ x"}, "⚠️ x", ""},
		{"pending", golden.Outcome{Verdict: golden.VerdictPending, Reason: "pending_strategy: y"}, "pending_strategy: y", ""},
		{"failed", golden.Outcome{
			ID: "PF-1", File: "612-pf/PF-1.yaml", Verdict: golden.VerdictFailed, Reason: "mismatch",
			Mismatches: []string{"a", "b"},
		}, "", "PF-1 (612-pf/PF-1.yaml): mismatch\n  a\n  b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb := &recordingTB{}

			golden.Report(tb, tt.o)

			assert.Equal(t, tt.skipped, tb.skipped)
			assert.Equal(t, tt.errored, tb.errored)
		})
	}
}

func TestSummaryCountsAndListsEverySkip(t *testing.T) {
	var s golden.Summary
	for _, o := range []golden.Outcome{
		{ID: "A", Verdict: golden.VerdictPassed},
		{ID: "B", Verdict: golden.VerdictFailed},
		{ID: "C", Verdict: golden.VerdictSkipped, Reason: "⚠️ rif reduction"},
		{ID: "D", Key: key612, Verdict: golden.VerdictPending},
		{ID: "E", Key: key612, Verdict: golden.VerdictPending},
	} {
		s.Add(o)
	}

	assert.Equal(t, 5, s.Total())
	assert.Equal(t, 1, s.Passed)
	assert.Equal(t, 1, s.Failed)
	assert.Equal(t, 1, s.Skipped)
	assert.Equal(t, 2, s.Pending)
	assert.Equal(t,
		"golden: 5 cases: 1 passed, 1 failed, 1 skipped (unverified), 2 pending_strategy\n"+
			"  skipped C: ⚠️ rif reduction\n"+
			"  pending_strategy 612/fisica: 2 cases\n",
		s.String())
}
