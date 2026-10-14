package golden

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"go.yaml.in/yaml/v3"
)

// Baseline is the checked-in ceiling of the cases that do not run: the
// skipped (⚠️) case ids and the pending_strategy count per strategy key
// ("612/fisica"). It is edited by hand, only downwards: a strategy PR
// lowers its count, a CPA sign-off removes a skip (PLAN M5-07 ratchet).
type Baseline struct {
	Skipped []string       `yaml:"skipped"`
	Pending map[string]int `yaml:"pending"`
}

// ReadBaseline decodes a baseline strictly: unknown fields, negative counts
// and repeated ids are errors, so a typo cannot loosen the ratchet.
func ReadBaseline(data []byte) (Baseline, error) {
	var b Baseline
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&b); err != nil {
		return Baseline{}, fmt.Errorf("golden: baseline: %w: %w", ErrSyntax, err)
	}
	for key, n := range b.Pending {
		if n < 0 {
			return Baseline{}, fmt.Errorf("golden: baseline: pending %s is %d: %w", key, n, ErrInvalidCase)
		}
	}
	sorted := slices.Sorted(slices.Values(b.Skipped))
	if len(slices.Compact(sorted)) != len(b.Skipped) {
		return Baseline{}, fmt.Errorf("golden: baseline: a skipped id is repeated: %w", ErrInvalidCase)
	}
	return b, nil
}

// Ratchet lists every way the run departs from the baseline: a skip or a
// pending count above it (the suite may not quietly stop gating), and one
// below it (the baseline must be lowered, or the slack lets it grow back).
// Pending findings come first, by strategy key; then skips, by id.
func (s *Summary) Ratchet(b Baseline) []string {
	counts := make(map[string]int, len(s.pending))
	for k, n := range s.pending {
		counts[k.String()] = n
	}
	return b.Findings(counts, s.skippedIDs)
}

// Findings compares a run's pending counts per key and skipped ids to the
// baseline, as Summary.Ratchet does; the annual harness, whose keys are
// section kinds, shares it (PLAN ANN-06).
func (b Baseline) Findings(pending map[string]int, skippedIDs []string) []string {
	return slices.Concat(pendingRatchet(pending, b.Pending), skipRatchet(skippedIDs, b.Skipped))
}

func pendingRatchet(counts, ceiling map[string]int) []string {
	keys := slices.Concat(slices.Collect(maps.Keys(counts)), slices.Collect(maps.Keys(ceiling)))
	slices.Sort(keys)
	keys = slices.Compact(keys)
	var out []string
	for _, k := range keys {
		got, want := counts[k], ceiling[k]
		switch {
		case got > want:
			out = append(out, fmt.Sprintf("pending_strategy %s: %d cases, baseline %d: register the strategy instead of adding pending cases", k, got, want))
		case got < want:
			out = append(out, fmt.Sprintf("pending_strategy %s: %d cases, baseline %d: lower the baseline to %d", k, got, want, got))
		}
	}
	return out
}

func skipRatchet(skippedIDs, allowed []string) []string {
	var out []string
	for _, id := range slices.Sorted(slices.Values(skippedIDs)) {
		if !slices.Contains(allowed, id) {
			out = append(out, fmt.Sprintf("skipped %s is not in the baseline: verify the rule instead of skipping it", id))
		}
	}
	for _, id := range slices.Sorted(slices.Values(allowed)) {
		if !slices.Contains(skippedIDs, id) {
			out = append(out, fmt.Sprintf("skipped %s no longer skips: remove it from the baseline", id))
		}
	}
	return out
}
