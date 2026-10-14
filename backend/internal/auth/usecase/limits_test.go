package usecase_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// The per-account limits of MODULE-auth §5. A change here is a change to the
// spec (and the security gate, M1-17, reviews it).
func TestAccountLimits_matchTheSpec(t *testing.T) {
	tests := []struct {
		name   string
		policy ratelimit.Policy
		limit  int
		window time.Duration
	}{
		{name: "login: 5 failures / 15 min per email", policy: usecase.LoginFailuresPolicy, limit: 5, window: 15 * time.Minute},
		{name: "register + resend: 5 / h per email", policy: usecase.SignUpEmailPolicy, limit: 5, window: time.Hour},
		{name: "refresh: 20 rotations / h per family", policy: usecase.RefreshFamilyPolicy, limit: 20, window: time.Hour},
		{name: "deleteMe: 5 attempts, then 1 / 3 min per account", policy: usecase.DeleteAttemptsPolicy, limit: 5, window: 15 * time.Minute},
	}
	scopes := map[string]bool{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.policy.Validate())
			assert.Equal(t, tt.limit, tt.policy.Burst)
			assert.Equal(t, tt.window, tt.policy.Interval*time.Duration(tt.policy.Burst))
			assert.False(t, scopes[tt.policy.Scope], "each limit has its own scope")
			scopes[tt.policy.Scope] = true
		})
	}
}
