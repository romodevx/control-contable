package access_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/access"
)

func intPtr(n int) *int { return &n }

// OQ-D6: whole days left, rounded up, at least 1 while the trial runs.
func TestTrialDaysLeft(t *testing.T) {
	ends := time.Date(2026, time.October, 20, 16, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		left time.Duration
		want int
	}{
		{7 * 24 * time.Hour, 7},
		{3*24*time.Hour - time.Second, 3},
		{2*24*time.Hour + time.Second, 3},
		{time.Hour, 1},
		{time.Nanosecond, 1},
		{0, 1},
		{-time.Hour, 1},
	} {
		assert.Equal(t, tc.want, access.TrialDaysLeft(ends.Add(-tc.left), ends), "%s left", tc.left)
	}
}

// The contract AccessSummary per level: trial fields only while the trial
// runs; counters for trial and read-only; pro and comp all null.
func TestSummaryView(t *testing.T) {
	now := time.Date(2026, time.October, 14, 18, 0, 0, 0, time.UTC)
	ends := now.Add(3*24*time.Hour + 2*time.Hour)
	tests := []struct {
		name    string
		summary access.Summary
		want    access.View
	}{
		{
			name: "trial",
			summary: access.Summary{
				Status: access.Status{Level: access.LevelTrial, TrialEndsAt: ends}, Usage: access.Usage{Incomes: 3, Expenses: 0},
			},
			want: access.View{
				Level: access.LevelTrial, TrialEndsAt: &ends, TrialDaysLeft: intPtr(4), IncomesRemaining: intPtr(2), ExpensesRemaining: intPtr(5),
			},
		},
		{
			name:    "read only: counters 0, no trial fields",
			summary: access.Summary{Status: access.Status{Level: access.LevelReadOnly, TrialEndsAt: now.AddDate(0, 0, -30)}},
			want:    access.View{Level: access.LevelReadOnly, IncomesRemaining: intPtr(0), ExpensesRemaining: intPtr(0)},
		},
		{name: "pro", summary: access.Summary{Status: access.Status{Level: access.LevelPro}}, want: access.View{Level: access.LevelPro}},
		{name: "comp", summary: access.Summary{Status: access.Status{Level: access.LevelComp}}, want: access.View{Level: access.LevelComp}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.summary.View(now))
		})
	}
}

// Policy.View is Summary.View on the policy's own clock, in one read-only
// unit of work of the user.
func TestPolicyView(t *testing.T) {
	now := registeredAt.Add(time.Hour)
	policy, txm := newPolicy(&fakeStore{entitlements: []access.Entitlement{trial()}, usage: access.Usage{Incomes: 1}}, now)

	got, err := policy.View(t.Context(), userID)

	require.NoError(t, err)
	assert.Equal(t, access.View{
		Level: access.LevelTrial, TrialEndsAt: &trialEnd, TrialDaysLeft: intPtr(7), IncomesRemaining: intPtr(4), ExpensesRemaining: intPtr(5),
	}, got)
	assert.Equal(t, []string{"ro"}, txm.modes)
}

func TestPolicyView_PropagatesFailures(t *testing.T) {
	boom := errors.New("boom")
	policy, _ := newPolicy(&fakeStore{entErr: boom}, registeredAt)

	_, err := policy.View(t.Context(), userID)

	require.ErrorIs(t, err, boom)
}
