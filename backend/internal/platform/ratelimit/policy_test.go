package ratelimit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

func TestPolicy_Validate(t *testing.T) {
	tests := []struct {
		name   string
		policy ratelimit.Policy
		ok     bool
	}{
		{name: "valid", policy: ratelimit.Policy{Scope: "login_account", Burst: 5, Interval: 3 * time.Minute}, ok: true},
		{name: "empty scope", policy: ratelimit.Policy{Burst: 5, Interval: time.Minute}},
		{name: "scope not snake_case", policy: ratelimit.Policy{Scope: "Login-IP", Burst: 5, Interval: time.Minute}},
		{name: "scope starts with a digit", policy: ratelimit.Policy{Scope: "1login", Burst: 5, Interval: time.Minute}},
		{name: "zero burst", policy: ratelimit.Policy{Scope: "login_ip", Interval: time.Minute}},
		{name: "zero interval", policy: ratelimit.Policy{Scope: "login_ip", Burst: 5}},
		{name: "sub-microsecond interval", policy: ratelimit.Policy{Scope: "login_ip", Burst: 5, Interval: time.Nanosecond}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.ok {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, ratelimit.ErrInvalidPolicy)
		})
	}
}

// PerWindow states a limit the way MODULE-auth §5 does ("5 / 15 min"): the
// whole budget is available at once and refills evenly over the window.
func TestPerWindow(t *testing.T) {
	p := ratelimit.PerWindow("login_account", 5, 15*time.Minute)

	assert.Equal(t, ratelimit.Policy{Scope: "login_account", Burst: 5, Interval: 3 * time.Minute}, p)
}

func TestRetryAfterSeconds(t *testing.T) {
	tests := []struct {
		wait time.Duration
		want int
	}{
		{wait: 0, want: 1},
		{wait: time.Millisecond, want: 1},
		{wait: time.Second, want: 1},
		{wait: time.Second + time.Nanosecond, want: 2},
		{wait: 90*time.Second + time.Millisecond, want: 91},
	}
	for _, tt := range tests {
		assert.Equalf(t, tt.want, ratelimit.RetryAfterSeconds(tt.wait), "wait %s", tt.wait)
	}
}
