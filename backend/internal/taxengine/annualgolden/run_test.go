package annualgolden_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// fakeEstimator is the consumer-side Estimator: the sections it knows and a
// canned result.
type fakeEstimator struct {
	registered []annual.SectionKind
	result     annualgolden.Result
	err        error
	got        *annualgolden.Input
}

func (f *fakeEstimator) Registered(k annual.SectionKind) bool {
	return slices.Contains(f.registered, k)
}

func (f *fakeEstimator) Estimate(in annualgolden.Input) (annualgolden.Result, error) {
	f.got = &in
	return f.result, f.err
}

func caseOf(t *testing.T) annualgolden.Case {
	t.Helper()
	c, err := annualgolden.Decode([]byte(validCase), "pf/AN-02.yaml")
	require.NoError(t, err)
	c.PendingStrategy = false
	return c
}

// matching is a result that satisfies validCase.
func matching(t *testing.T) annualgolden.Result {
	t.Helper()
	return annualgolden.Result{
		InformativeIncomeOver500k: true,
		Disclaimers:               []annual.DisclaimerCode{annual.DisclaimerDeductionsAsEntered, annual.DisclaimerAnnualEstimate},
		Sections: []annualgolden.SectionResult{{
			Kind: annual.SectionPFArt152, Regimes: []fiscal.RegimeCode{"612", "606", "605"},
			Status: annual.StatusComplete, Obligation: annual.ObligationObligated,
			CPATopics: []annual.CPATopic{annual.TopicUMAForMinimumWage},
			Tax: &annualgolden.TaxFigures{
				AccumulableIncome: cents(t, "760000.00"), PersonalDeductions: cents(t, "167900.00"),
				Base: cents(t, "592100.00"), ISR: cents(t, "107435.78"),
				Payments: cents(t, "38000.00"), Withholdings: cents(t, "15000.00"), SalaryISR: cents(t, "54235.91"),
				Balance: annualgolden.Balance{Kind: annual.BalanceACargo, Amount: cents(t, "199.87")},
			},
			Lines: []annualgolden.Line{
				{Code: "isr.accumulable_income", Op: domain.OpInput, Amount: cents(t, "760000.00")},
				{Code: "deductions.cap_applied", Op: domain.OpInfo, Amount: cents(t, "12900.00")},
				{Code: "isr.balance", Op: domain.OpResult, Amount: cents(t, "199.87")},
			},
		}},
	}
}

func TestEvaluatePassesAMatchingResult(t *testing.T) {
	est := &fakeEstimator{registered: []annual.SectionKind{annual.SectionPFArt152}, result: matching(t)}
	c := caseOf(t)

	o := annualgolden.Evaluate(c, est)

	assert.Equal(t, golden.VerdictPassed, o.Verdict, o.Mismatches)
	require.NotNil(t, est.got)
	assert.Equal(t, c.Input, *est.got)
}

func TestEvaluateReportsEveryMismatch(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*annualgolden.Result)
		want string
	}{
		{"isr", func(r *annualgolden.Result) { r.Sections[0].Tax.ISR++ }, "pf_art152 tax.isr"},
		{"balance", func(r *annualgolden.Result) { r.Sections[0].Tax.Balance.Kind = annual.BalanceAFavor }, "pf_art152 tax.balance"},
		{"credits", func(r *annualgolden.Result) { r.Sections[0].Tax.SalaryISR = 0 }, "pf_art152 tax.salary_isr"},
		{"status", func(r *annualgolden.Result) { r.Sections[0].Status = annual.StatusIncomplete }, "pf_art152 status"},
		{"obligation", func(r *annualgolden.Result) { r.Sections[0].Obligation = annual.ObligationOptional }, "pf_art152 obligation"},
		{"regimes", func(r *annualgolden.Result) { r.Sections[0].Regimes = []fiscal.RegimeCode{"612"} }, "pf_art152 regimes"},
		{"topics", func(r *annualgolden.Result) { r.Sections[0].CPATopics = nil }, "pf_art152 cpa_topics"},
		{"missing", func(r *annualgolden.Result) { r.Sections[0].MissingInputs = []string{"salary_statement"} }, "pf_art152 missing_inputs"},
		{"no tax", func(r *annualgolden.Result) { r.Sections[0].Tax = nil }, "pf_art152 tax"},
		{"line amount", func(r *annualgolden.Result) { r.Sections[0].Lines[1].Amount = 1 }, "pf_art152 lines"},
		{"line missing", func(r *annualgolden.Result) { r.Sections[0].Lines = r.Sections[0].Lines[:1] }, "pf_art152 lines"},
		{"disclaimers", func(r *annualgolden.Result) { r.Disclaimers = nil }, "disclaimers"},
		{"informative income", func(r *annualgolden.Result) { r.InformativeIncomeOver500k = false }, "informative_income_over_500k"},
		{"sections", func(r *annualgolden.Result) {
			r.Sections = append(r.Sections, annualgolden.SectionResult{Kind: annual.SectionExplanation})
		}, "sections"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matching(t)
			tt.mut(&result)
			est := &fakeEstimator{registered: []annual.SectionKind{annual.SectionPFArt152}, result: result}

			o := annualgolden.Evaluate(caseOf(t), est)

			assert.Equal(t, golden.VerdictFailed, o.Verdict)
			require.NotEmpty(t, o.Mismatches)
			assert.Contains(t, o.Mismatches[0], tt.want)
		})
	}
}

func TestEvaluateComparesPMExtrasAndTheDifference(t *testing.T) {
	doc := `# golden: hand-authored
id: AN-06
source: research:AN-06
input:
  fiscal_year: 2026
  person_type: moral
  regimes:
    - {regime: "601", income: "10000000.00"}
expected:
  sections:
    - kind: pm_601
      status: complete
      pm_extras: {ptu_base: "1500000.00", suggested_cu_next_year: "0.1380"}
    - kind: pf_resico_check
      status: informational
      difference: "-1320.00"
`
	c, err := annualgolden.Decode([]byte(doc), "pm/AN-06.yaml")
	require.NoError(t, err)
	ok := annualgolden.Result{Sections: []annualgolden.SectionResult{
		{Kind: annual.SectionPM601, Status: annual.StatusComplete,
			PMExtras: &annualgolden.PMExtrasResult{PTUBase: cents(t, "1500000.00"), PTUToDistribute: 1, SuggestedCU: 138000}},
		{Kind: annual.SectionPFRESICOCheck, Status: annual.StatusInformational, Difference: centsPtr(t, "-1320.00")},
	}}
	est := &fakeEstimator{registered: []annual.SectionKind{annual.SectionPM601, annual.SectionPFRESICOCheck}, result: ok}

	assert.Equal(t, golden.VerdictPassed, annualgolden.Evaluate(c, est).Verdict)

	ok.Sections[0].PMExtras.SuggestedCU = 138100
	ok.Sections[1].Difference = nil
	o := annualgolden.Evaluate(c, est)
	assert.Equal(t, golden.VerdictFailed, o.Verdict)
	assert.Len(t, o.Mismatches, 2)
}

func TestEvaluateMarksAndPolicesPendingAndSkippedCases(t *testing.T) {
	tests := []struct {
		name       string
		mut        func(*annualgolden.Case)
		registered []annual.SectionKind
		err        error
		want       golden.Verdict
		reason     string
	}{
		{name: "skip_reason wins", mut: func(c *annualgolden.Case) { c.SkipReason = "⚠️ CPA" },
			want: golden.VerdictSkipped, reason: "skipped: ⚠️ CPA"},
		{name: "pending strategy", mut: func(c *annualgolden.Case) { c.PendingStrategy = true },
			want: golden.VerdictPending, reason: "pending_strategy: no annual section registered for pf_art152"},
		{name: "unmarked unregistered section fails", mut: func(*annualgolden.Case) {},
			want: golden.VerdictFailed, reason: "no annual section registered for pf_art152 and the case is not marked pending_strategy"},
		{name: "stale pending marker fails", mut: func(c *annualgolden.Case) { c.PendingStrategy = true },
			registered: []annual.SectionKind{annual.SectionPFArt152},
			want:       golden.VerdictFailed, reason: "sections run: remove pending_strategy from pf/AN-02.yaml"},
		{name: "estimate error fails", mut: func(*annualgolden.Case) {},
			registered: []annual.SectionKind{annual.SectionPFArt152}, err: errors.New("boom"),
			want: golden.VerdictFailed, reason: "estimate: boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caseOf(t)
			tt.mut(&c)
			est := &fakeEstimator{registered: tt.registered, result: matching(t), err: tt.err}

			o := annualgolden.Evaluate(c, est)

			assert.Equal(t, tt.want, o.Verdict)
			assert.Equal(t, tt.reason, o.Reason)
		})
	}
}

func TestSummaryCountsAndRatchetsByPendingSection(t *testing.T) {
	var s annualgolden.Summary
	s.Add(annualgolden.Outcome{ID: "AN-01", Section: annual.SectionPFArt152, Verdict: golden.VerdictPending})
	s.Add(annualgolden.Outcome{ID: "AN-02", Section: annual.SectionPFArt152, Verdict: golden.VerdictPending})
	s.Add(annualgolden.Outcome{ID: "AN-04", Verdict: golden.VerdictSkipped, Reason: "skipped: ⚠️ platform"})
	s.Add(annualgolden.Outcome{ID: "AN-06", Verdict: golden.VerdictPassed})
	s.Add(annualgolden.Outcome{ID: "AN-07", Verdict: golden.VerdictFailed})

	assert.Equal(t, 5, s.Total())
	assert.Equal(t, "annual golden: 5 cases: 1 passed, 1 failed, 1 skipped (unverified), 2 pending_strategy\n"+
		"  skipped AN-04: ⚠️ platform\n"+
		"  pending_strategy pf_art152: 2 cases\n", s.String())
	assert.Empty(t, s.Ratchet(golden.Baseline{Skipped: []string{"AN-04"}, Pending: map[string]int{"pf_art152": 2}}))
	assert.Equal(t, []string{
		"pending_strategy pf_art152: 2 cases, baseline 1: register the strategy instead of adding pending cases",
		"skipped AN-04 is not in the baseline: verify the rule instead of skipping it",
	}, s.Ratchet(golden.Baseline{Pending: map[string]int{"pf_art152": 1}}))
}

func TestRunMakesOneSubtestPerCase(t *testing.T) {
	c := caseOf(t)
	c.PendingStrategy = true

	s := annualgolden.Run(t, []annualgolden.Case{c}, &fakeEstimator{})

	assert.Equal(t, 1, s.Pending)
}
