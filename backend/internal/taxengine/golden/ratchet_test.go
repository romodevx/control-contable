package golden_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

var key626 = domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonFisica}

func summaryOf(outcomes ...golden.Outcome) golden.Summary {
	var s golden.Summary
	for _, o := range outcomes {
		s.Add(o)
	}
	return s
}

func TestReadBaselineIsStrict(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    golden.Baseline
		wantErr error
	}{
		{
			name: "skipped ids and pending counts",
			yaml: "skipped: [PF-621-01]\npending:\n  612/fisica: 3\n",
			want: golden.Baseline{Skipped: []string{"PF-621-01"}, Pending: map[string]int{"612/fisica": 3}},
		},
		{name: "unknown field", yaml: "skipped: []\npending: {}\nslack: 2\n", wantErr: golden.ErrSyntax},
		{name: "negative count", yaml: "pending:\n  612/fisica: -1\n", wantErr: golden.ErrInvalidCase},
		{name: "repeated skipped id", yaml: "skipped: [A, A]\n", wantErr: golden.ErrInvalidCase},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := golden.ReadBaseline([]byte(tt.yaml))

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRatchetHoldsSkipsAndPendingAtTheBaseline(t *testing.T) {
	baseline := golden.Baseline{Skipped: []string{"S1"}, Pending: map[string]int{"612/fisica": 2}}
	tests := []struct {
		name     string
		outcomes []golden.Outcome
		want     []string
	}{
		{
			name: "at the baseline",
			outcomes: []golden.Outcome{
				{ID: "S1", Verdict: golden.VerdictSkipped},
				{ID: "P1", Key: key612, Verdict: golden.VerdictPending},
				{ID: "P2", Key: key612, Verdict: golden.VerdictPending},
			},
		},
		{
			name: "a new skip",
			outcomes: []golden.Outcome{
				{ID: "S1", Verdict: golden.VerdictSkipped},
				{ID: "S2", Verdict: golden.VerdictSkipped},
				{ID: "P1", Key: key612, Verdict: golden.VerdictPending},
				{ID: "P2", Key: key612, Verdict: golden.VerdictPending},
			},
			want: []string{"skipped S2 is not in the baseline: verify the rule instead of skipping it"},
		},
		{
			name: "more pending for a strategy, and a strategy not in the baseline",
			outcomes: []golden.Outcome{
				{ID: "S1", Verdict: golden.VerdictSkipped},
				{ID: "P1", Key: key612, Verdict: golden.VerdictPending},
				{ID: "P2", Key: key612, Verdict: golden.VerdictPending},
				{ID: "P3", Key: key612, Verdict: golden.VerdictPending},
				{ID: "P4", Key: key626, Verdict: golden.VerdictPending},
			},
			want: []string{
				"pending_strategy 612/fisica: 3 cases, baseline 2: register the strategy instead of adding pending cases",
				"pending_strategy 626/fisica: 1 cases, baseline 0: register the strategy instead of adding pending cases",
			},
		},
		{
			name: "fewer than the baseline must tighten it",
			outcomes: []golden.Outcome{
				{ID: "P1", Key: key612, Verdict: golden.VerdictPending},
			},
			want: []string{
				"pending_strategy 612/fisica: 1 cases, baseline 2: lower the baseline to 1",
				"skipped S1 no longer skips: remove it from the baseline",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := summaryOf(tt.outcomes...)

			got := s.Ratchet(baseline)

			assert.Equal(t, tt.want, got)
		})
	}
}
