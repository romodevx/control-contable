package router_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/adapter/http/router/routertest"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

// userHeader names the fake caller: principalAuthn admits it as the
// principal, and refuses a request without it.
const userHeader = "X-Test-User"

type principalAuthn struct{}

func (principalAuthn) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get(userHeader)
		if user == "" {
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeUnauthorized, "Authentication required."))
			return
		}
		next.ServeHTTP(w, r.WithContext(principal.With(r.Context(), principal.Principal{UserID: user})))
	})
}

type userLimitedHarness struct {
	handler http.Handler
	calls   []string
	logs    bytes.Buffer
}

func newUserLimitedHarness(t *testing.T, flags flagSet, limits map[string]router.Limiter) *userLimitedHarness {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	h := &userLimitedHarness{}
	h.handler, err = router.New(fakeAPI{calls: &h.calls}, spec, router.Options{
		Flags:          flags,
		Authenticators: routertest.Authenticators(principalAuthn{}),
		Logger:         logging.New(&h.logs, nil),
		UserLimits:     limits,
	})
	require.NoError(t, err)
	return h
}

func (h *userLimitedHarness) listIncomes(user string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, incomesURL, nil)
	if user != "" {
		req.Header.Set(userHeader, user)
	}
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, req)
	return rec
}

var incomesOn = flagSet{"module_incomes": true}

// MODULE-taxpayers S-10: a per-user budget, keyed by the authenticated
// user, so one account cannot spend another's and a shared address (a
// carrier NAT) does not throttle distinct users.
func TestUserLimits_answer429WithRetryAfterOnceTheUserIsOut(t *testing.T) {
	limiter := &fakeLimiter{budget: 1, wait: 30 * time.Minute}
	h := newUserLimitedHarness(t, incomesOn, map[string]router.Limiter{"listIncomes": limiter})

	require.NotEqual(t, http.StatusTooManyRequests, h.listIncomes("user-a").Code)
	rec := h.listIncomes("user-a")

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "1800", rec.Header().Get("Retry-After"))
	assert.Equal(t, apperr.CodeRateLimited, envelope(t, rec).Code)
	assert.Equal(t, []string{"listIncomes"}, h.calls, "the limited request never reaches the handler")
	assert.Equal(t, []string{"user-a", "user-a"}, limiter.taken(), "keyed by the principal's user id")
}

// The user is known only after authentication: an anonymous request is
// refused with 401 and spends nobody's budget.
func TestUserLimits_runAfterAuthentication(t *testing.T) {
	limiter := &fakeLimiter{}
	h := newUserLimitedHarness(t, incomesOn, map[string]router.Limiter{"listIncomes": limiter})

	rec := h.listIncomes("")

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, limiter.taken())
}

func TestUserLimits_runAfterTheFlagGate(t *testing.T) {
	limiter := &fakeLimiter{}
	h := newUserLimitedHarness(t, flagSet{}, map[string]router.Limiter{"listIncomes": limiter})

	rec := h.listIncomes("user-a")

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, limiter.taken(), "a disabled operation spends no budget")
}

func TestUserLimits_failClosedWhenTheLimiterFails(t *testing.T) {
	limiter := &fakeLimiter{err: errors.New("database down")}
	h := newUserLimitedHarness(t, incomesOn, map[string]router.Limiter{"listIncomes": limiter})

	rec := h.listIncomes("user-a")

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, h.calls)
	assert.Contains(t, h.logs.String(), "database down")
}

func TestNew_rejectsBadUserLimits(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	base := router.Options{Flags: flagSet{}, Authenticators: routertest.Authenticators(principalAuthn{}), Logger: logging.New(&bytes.Buffer{}, nil)}

	t.Run("an operation the contract lacks", func(t *testing.T) {
		opts := base
		opts.UserLimits = map[string]router.Limiter{"noSuchOperation": &fakeLimiter{}}
		_, err := router.New(fakeAPI{calls: new([]string)}, spec, opts)
		assert.ErrorContains(t, err, "noSuchOperation")
	})
	t.Run("a public operation has no user", func(t *testing.T) {
		opts := base
		opts.UserLimits = map[string]router.Limiter{"login": &fakeLimiter{}}
		_, err := router.New(fakeAPI{calls: new([]string)}, spec, opts)
		assert.ErrorContains(t, err, `user-limited operation "login" is public`)
	})
	// The webhook secret authenticates RevenueCat, not a user: a per-user
	// limit there would key every call by "" (one shared bucket).
	t.Run("an operation secured by a non-user scheme has no user", func(t *testing.T) {
		opts := base
		opts.UserLimits = map[string]router.Limiter{"receiveRevenueCatWebhook": &fakeLimiter{}}
		_, err := router.New(fakeAPI{calls: new([]string)}, spec, opts)
		assert.ErrorContains(t, err, `user-limited operation "receiveRevenueCatWebhook"`)
		assert.ErrorContains(t, err, router.SchemeRevenueCatWebhook)
	})
}
