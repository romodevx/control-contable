package authhttp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
)

// The per-client limits of MODULE-auth §5, by operationId. getMe, updateMe
// and logout need a valid bearer and are bounded by the edge limiter;
// deleteMe is limited per account (usecase.DeleteAttemptsPolicy).
func TestClientLimits_matchTheSpec(t *testing.T) {
	want := map[string]struct {
		limit  int
		window time.Duration
	}{
		"register":           {limit: 10, window: time.Hour},
		"verifyEmail":        {limit: 20, window: 15 * time.Minute},
		"resendVerification": {limit: 10, window: time.Hour},
		"login":              {limit: 20, window: 15 * time.Minute},
		"refreshSession":     {limit: 60, window: time.Minute},
		"listLegalDocuments": {limit: 60, window: time.Minute},
	}

	assert.Len(t, authhttp.ClientLimits, len(want))
	scopes := map[string]bool{}
	for op, policy := range authhttp.ClientLimits {
		w, ok := want[op]
		if !assert.Truef(t, ok, "%s is not limited per client in §5", op) {
			continue
		}
		assert.NoError(t, policy.Validate(), op)
		assert.Equal(t, w.limit, policy.Burst, op)
		assert.Equal(t, w.window, policy.Interval*time.Duration(policy.Burst), op)
		assert.False(t, scopes[policy.Scope], "%s shares a scope", op)
		scopes[policy.Scope] = true
	}
}
