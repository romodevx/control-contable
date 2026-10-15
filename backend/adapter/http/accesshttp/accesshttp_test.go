package accesshttp_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/accesshttp"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
)

func intPtr(n int) *int { return &n }

// Every AccessSummary on the wire has all five keys (the contract makes
// them required, nullable), and trial_ends_at is UTC to the second whatever
// the stored precision or zone.
func TestSummary(t *testing.T) {
	mexicoCity := time.FixedZone("CST", -6*60*60)
	ends := time.Date(2026, time.October, 20, 10, 0, 0, 123456000, mexicoCity)
	tests := []struct {
		name string
		view access.View
		want string
	}{
		{
			name: "trial",
			view: access.View{Level: access.LevelTrial, TrialEndsAt: &ends, TrialDaysLeft: intPtr(3), IncomesRemaining: intPtr(2), ExpensesRemaining: intPtr(5)},
			want: `{"access":"trial","trial_ends_at":"2026-10-20T16:00:00Z","trial_days_left":3,"incomes_remaining":2,"expenses_remaining":5}`,
		},
		{
			name: "read only",
			view: access.View{Level: access.LevelReadOnly, IncomesRemaining: intPtr(0), ExpensesRemaining: intPtr(0)},
			want: `{"access":"read_only","trial_ends_at":null,"trial_days_left":null,"incomes_remaining":0,"expenses_remaining":0}`,
		},
		{
			name: "pro",
			view: access.View{Level: access.LevelPro},
			want: `{"access":"pro","trial_ends_at":null,"trial_days_left":null,"incomes_remaining":null,"expenses_remaining":null}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(accesshttp.Summary(tt.view))

			require.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestSummary_doesNotAliasTheView(t *testing.T) {
	ends := time.Date(2026, time.October, 20, 16, 0, 0, 0, time.UTC)
	view := access.View{Level: access.LevelTrial, TrialEndsAt: &ends}

	got := accesshttp.Summary(view)
	*got.TrialEndsAt = got.TrialEndsAt.Add(time.Hour)

	assert.Equal(t, time.Date(2026, time.October, 20, 16, 0, 0, 0, time.UTC), ends)
}
