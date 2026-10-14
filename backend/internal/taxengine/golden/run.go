package golden

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Registry is the consumer-side view of domain.Registry (M5-10): Lookup
// tells a registered strategy from the informational fallback, and
// Estimate derives the missing inputs exactly as the use case does.
type Registry interface {
	Lookup(key domain.StrategyKey) (domain.Strategy, bool)
	Estimate(in domain.Input, activeMonths []time.Month, p *taxparams.Params) (domain.RegimeEstimate, error)
}

// ParamsSelector selects the parameter set of a fiscal year;
// *taxparams.Catalog implements it.
type ParamsSelector interface {
	Select(fiscalYear int) (taxparams.Selection, error)
}

// Verdict is the result of one case.
type Verdict int

// The verdicts. Skipped and Pending are reported, never silent.
const (
	VerdictPassed  Verdict = iota + 1
	VerdictFailed          // a mismatch, an engine error or a broken marker
	VerdictSkipped         // skip_reason: an unverified (⚠️) rule
	VerdictPending         // pending_strategy: its strategy is not registered yet
)

// Outcome is one case's verdict and why.
type Outcome struct {
	ID         string
	File       string
	Key        domain.StrategyKey
	Verdict    Verdict
	Reason     string
	Mismatches []string
}

// Evaluate runs c through the registry with the parameters of its fiscal
// year and compares the estimate to c.Want. A case runs when its strategy
// is registered, or when it expects an informational regime (the
// registry's fallback is then the strategy under test); otherwise it is
// pending if marked so (or generated) and failed if not.
func Evaluate(c Case, reg Registry, sel ParamsSelector) Outcome {
	o := Outcome{ID: c.ID, File: c.File, Key: c.Input.Key}
	if c.SkipReason != "" {
		return o.with(VerdictSkipped, "skipped: "+c.SkipReason)
	}
	_, registered := reg.Lookup(c.Input.Key)
	runnable := registered || c.Want.Status == domain.StatusInformational
	switch {
	case !runnable && (c.PendingStrategy || c.Generated):
		return o.with(VerdictPending, fmt.Sprintf("pending_strategy: no strategy registered for %s", c.Input.Key))
	case !runnable:
		return o.with(VerdictFailed, fmt.Sprintf("no strategy registered for %s and the case is not marked pending_strategy", c.Input.Key))
	case c.PendingStrategy:
		return o.with(VerdictFailed, fmt.Sprintf("strategy %s runs: remove pending_strategy from %s", c.Input.Key, c.File))
	}
	selection, err := sel.Select(c.Input.Period.Year())
	if err != nil {
		return o.with(VerdictFailed, fmt.Sprintf("parameters for %d: %v", c.Input.Period.Year(), err))
	}
	estimate, err := reg.Estimate(c.Input, slices.Clone(c.ActiveMonths), selection.Params)
	if err != nil {
		return o.with(VerdictFailed, fmt.Sprintf("estimate: %v", err))
	}
	if o.Mismatches = compare(c.Want, estimate); len(o.Mismatches) > 0 {
		return o.with(VerdictFailed, "estimate differs from the golden")
	}
	return o.with(VerdictPassed, "")
}

func (o Outcome) with(v Verdict, reason string) Outcome {
	o.Verdict, o.Reason = v, reason
	return o
}

// Report turns an outcome into the test's result: a skip with its reason,
// or an error listing every mismatch.
func Report(tb testing.TB, o Outcome) {
	tb.Helper()
	switch o.Verdict {
	case VerdictSkipped, VerdictPending:
		tb.Skip(o.Reason)
	case VerdictFailed:
		var b strings.Builder
		fmt.Fprintf(&b, "%s (%s): %s", o.ID, o.File, o.Reason)
		for _, m := range o.Mismatches {
			b.WriteString("\n  " + m)
		}
		tb.Errorf("%s", b.String())
	}
}

// Run evaluates every case as a subtest named by its id and logs the
// Summary. An empty list fails: an empty suite gates nothing.
func Run(t *testing.T, cases []Case, reg Registry, sel ParamsSelector) Summary {
	t.Helper()
	if len(cases) == 0 {
		t.Fatal(ErrNoCases)
	}
	var s Summary
	for _, c := range cases {
		o := Evaluate(c, reg, sel)
		s.Add(o)
		t.Run(c.ID, func(t *testing.T) { Report(t, o) })
	}
	t.Log(s.String())
	return s
}

// Summary counts the verdicts of a run and keeps every skip reason, so the
// skipped and pending cases are always visible in the test log.
type Summary struct {
	Passed, Failed, Skipped, Pending int

	skips   []string
	pending map[domain.StrategyKey]int
}

// Add counts o.
func (s *Summary) Add(o Outcome) {
	switch o.Verdict {
	case VerdictPassed:
		s.Passed++
	case VerdictFailed:
		s.Failed++
	case VerdictSkipped:
		s.Skipped++
		s.skips = append(s.skips, fmt.Sprintf("skipped %s: %s", o.ID, strings.TrimPrefix(o.Reason, "skipped: ")))
	case VerdictPending:
		s.Pending++
		if s.pending == nil {
			s.pending = map[domain.StrategyKey]int{}
		}
		s.pending[o.Key]++
	}
}

// Total is the number of cases counted.
func (s *Summary) Total() int { return s.Passed + s.Failed + s.Skipped + s.Pending }

// String renders the counts, every skip with its reason and the pending
// cases per strategy.
func (s *Summary) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "golden: %d cases: %d passed, %d failed, %d skipped (unverified), %d pending_strategy\n",
		s.Total(), s.Passed, s.Failed, s.Skipped, s.Pending)
	for _, skip := range s.skips {
		b.WriteString("  " + skip + "\n")
	}
	keys := make([]domain.StrategyKey, 0, len(s.pending))
	for k := range s.pending {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b domain.StrategyKey) int { return strings.Compare(a.String(), b.String()) })
	for _, k := range keys {
		fmt.Fprintf(&b, "  pending_strategy %s: %d cases\n", k, s.pending[k])
	}
	return b.String()
}
