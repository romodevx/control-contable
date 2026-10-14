package annualgolden

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// Estimator is the consumer-side view of the annual engine: which sections
// are registered, and the estimate of a case's year as a Result. The
// strategies of ANN-08..ANN-11 register sections; ANN-14 wires the
// composition root's estimator into the gate.
type Estimator interface {
	Registered(kind annual.SectionKind) bool
	Estimate(in Input) (Result, error)
}

// Outcome is one case's verdict and why. Section is the first expected
// section that is not registered (the pending key).
type Outcome struct {
	ID         string
	File       string
	Section    annual.SectionKind
	Verdict    golden.Verdict
	Reason     string
	Mismatches []string
}

func (o Outcome) with(v golden.Verdict, reason string) Outcome {
	o.Verdict, o.Reason = v, reason
	return o
}

// Evaluate runs c through est and compares the result to c.Want. A case
// runs when every section it expects is registered; otherwise it is
// pending if marked so and failed if not. A skip_reason wins over both.
func Evaluate(c Case, est Estimator) Outcome {
	o := Outcome{ID: c.ID, File: c.File}
	if c.SkipReason != "" {
		return o.with(golden.VerdictSkipped, "skipped: "+c.SkipReason)
	}
	for _, s := range c.Want.Sections {
		if !est.Registered(s.Kind) {
			o.Section = s.Kind
			break
		}
	}
	switch {
	case o.Section != "" && c.PendingStrategy:
		return o.with(golden.VerdictPending, fmt.Sprintf("pending_strategy: no annual section registered for %s", o.Section))
	case o.Section != "":
		return o.with(golden.VerdictFailed, fmt.Sprintf("no annual section registered for %s and the case is not marked pending_strategy", o.Section))
	case c.PendingStrategy:
		return o.with(golden.VerdictFailed, fmt.Sprintf("sections run: remove pending_strategy from %s", c.File))
	}
	result, err := est.Estimate(c.Input)
	if err != nil {
		return o.with(golden.VerdictFailed, fmt.Sprintf("estimate: %v", err))
	}
	if o.Mismatches = compare(c.Want, result); len(o.Mismatches) > 0 {
		return o.with(golden.VerdictFailed, "estimate differs from the golden")
	}
	return o.with(golden.VerdictPassed, "")
}

// Report turns an outcome into the test's result: a skip with its reason,
// or an error listing every mismatch.
func Report(tb testing.TB, o Outcome) {
	tb.Helper()
	switch o.Verdict {
	case golden.VerdictSkipped, golden.VerdictPending:
		tb.Skip(o.Reason)
	case golden.VerdictFailed:
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
func Run(t *testing.T, cases []Case, est Estimator) Summary {
	t.Helper()
	if len(cases) == 0 {
		t.Fatal(golden.ErrNoCases)
	}
	var s Summary
	for _, c := range cases {
		o := Evaluate(c, est)
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

	skips      []string
	skippedIDs []string
	pending    map[string]int
}

// Add counts o.
func (s *Summary) Add(o Outcome) {
	switch o.Verdict {
	case golden.VerdictPassed:
		s.Passed++
	case golden.VerdictFailed:
		s.Failed++
	case golden.VerdictSkipped:
		s.Skipped++
		s.skippedIDs = append(s.skippedIDs, o.ID)
		s.skips = append(s.skips, fmt.Sprintf("skipped %s: %s", o.ID, strings.TrimPrefix(o.Reason, "skipped: ")))
	case golden.VerdictPending:
		s.Pending++
		if s.pending == nil {
			s.pending = map[string]int{}
		}
		s.pending[string(o.Section)]++
	}
}

// Total is the number of cases counted.
func (s *Summary) Total() int { return s.Passed + s.Failed + s.Skipped + s.Pending }

// Ratchet lists every way the run departs from the baseline (pending
// counts per section kind, skipped ids), as the monthly harness does.
func (s *Summary) Ratchet(b golden.Baseline) []string { return b.Findings(s.pending, s.skippedIDs) }

// String renders the counts, every skip with its reason and the pending
// cases per section.
func (s *Summary) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "annual golden: %d cases: %d passed, %d failed, %d skipped (unverified), %d pending_strategy\n",
		s.Total(), s.Passed, s.Failed, s.Skipped, s.Pending)
	for _, skip := range s.skips {
		b.WriteString("  " + skip + "\n")
	}
	for _, k := range slices.Sorted(maps.Keys(s.pending)) {
		fmt.Fprintf(&b, "  pending_strategy %s: %d cases\n", k, s.pending[k])
	}
	return b.String()
}
