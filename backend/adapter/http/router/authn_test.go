package router_test

import (
	"bytes"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/adapter/http/router/routertest"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
)

const (
	webhookURL  = "/v1/webhooks/revenuecat"
	webhookBody = `{"api_version":"1.0","event":{"id":"evt-1","type":"TEST","app_user_id":"u-1","event_timestamp_ms":1}}`
)

// credentialAuthn admits only requests whose Authorization header carries its
// own credential, like the real bearer authenticator and the webhook secret
// check: each one knows nothing of the other's credentials.
type credentialAuthn struct {
	credential string
}

func (a credentialAuthn) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+a.credential {
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeUnauthorized, "Authentication required."))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newCredential is generated per run: no credential-looking literal in the
// repository (GitGuardian).
func newCredential() string { return rand.Text() }

// The router picks the authenticator of each operation's own security scheme
// (bearerAuth by default, revenueCatWebhookSecret on the webhook): a user's
// access token never satisfies the webhook, and the webhook secret never
// stands in for a user.
func TestAuthn_eachOperationRunsTheAuthenticatorOfItsScheme(t *testing.T) {
	userToken, webhookSecret := newCredential(), newCredential()
	tests := []struct {
		name, method, target, body, credential string
		wantCall                               string // "" when authentication refuses the request
	}{
		{name: "user operation, user token", method: http.MethodGet, target: incomesURL, credential: userToken, wantCall: "listIncomes"},
		{name: "user operation, webhook secret", method: http.MethodGet, target: incomesURL, credential: webhookSecret},
		{name: "webhook, webhook secret", method: http.MethodPost, target: webhookURL, body: webhookBody, credential: webhookSecret, wantCall: "receiveRevenueCatWebhook"},
		{name: "webhook, user token", method: http.MethodPost, target: webhookURL, body: webhookBody, credential: userToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := gen.GetSpec()
			require.NoError(t, err)
			var calls []string
			handler, err := router.New(fakeAPI{calls: &calls}, spec, router.Options{
				Flags: flagSet{"module_incomes": true, "module_subscriptions": true},
				Authenticators: map[string]router.Authenticator{
					router.SchemeBearer:            credentialAuthn{credential: userToken},
					router.SchemeRevenueCatWebhook: credentialAuthn{credential: webhookSecret},
				},
				Logger: logging.New(&bytes.Buffer{}, nil),
			})
			require.NoError(t, err)
			req := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.credential)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if tt.wantCall == "" {
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
				assert.Empty(t, calls, "the handler never runs")
				return
			}
			assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
			assert.Equal(t, []string{tt.wantCall}, calls)
		})
	}
}

// A built operation whose scheme has no authenticator is a startup error,
// never an unauthenticated route (fail closed).
func TestNew_failsWhenASecuredOperationsSchemeHasNoAuthenticator(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)

	_, err = router.New(fakeAPI{}, spec, router.Options{
		Flags:          flagSet{},
		Authenticators: map[string]router.Authenticator{router.SchemeBearer: routertest.DenyAll{}},
		Logger:         logging.New(&bytes.Buffer{}, nil),
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, "receiveRevenueCatWebhook")
	assert.ErrorContains(t, err, router.SchemeRevenueCatWebhook)
}

// A not-built operation answers the hidden-module 404 before authentication,
// so its scheme needs no authenticator yet.
func TestNew_notBuiltOperationsNeedNoAuthenticator(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)

	_, err = router.New(fakeAPI{}, spec, router.Options{
		Flags:          flagSet{},
		Authenticators: map[string]router.Authenticator{router.SchemeBearer: routertest.DenyAll{}},
		Logger:         logging.New(&bytes.Buffer{}, nil),
		NotBuilt:       []string{"receiveRevenueCatWebhook"},
	})

	assert.NoError(t, err)
}

// An authenticator keyed by a scheme the contract does not declare is a
// typo that would leave the real scheme unauthenticated.
func TestNew_failsOnAnAuthenticatorForAnUndeclaredScheme(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	authenticators := routertest.Authenticators(routertest.DenyAll{})
	authenticators["apiKeyAuth"] = routertest.DenyAll{}

	_, err = router.New(fakeAPI{}, spec, router.Options{
		Flags: flagSet{}, Authenticators: authenticators, Logger: logging.New(&bytes.Buffer{}, nil),
	})

	assert.ErrorContains(t, err, "apiKeyAuth")
}

// The router runs exactly one authenticator per operation; alternative or
// combined requirements would need a policy it does not have, so they are a
// startup error instead of a silent pick.
func TestNew_failsOnAnOperationWithSeveralSecuritySchemes(t *testing.T) {
	tests := []struct {
		name     string
		security openapi3.SecurityRequirements
	}{
		{name: "alternatives", security: openapi3.SecurityRequirements{
			{router.SchemeBearer: {}}, {router.SchemeRevenueCatWebhook: {}},
		}},
		{name: "combined", security: openapi3.SecurityRequirements{
			{router.SchemeBearer: {}, router.SchemeRevenueCatWebhook: {}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := gen.GetSpec()
			require.NoError(t, err)
			spec.Paths.Value("/v1/taxpayers/{taxpayerId}/incomes").Get.Security = &tt.security

			_, err = router.New(fakeAPI{}, spec, router.Options{
				Flags: flagSet{}, Authenticators: routertest.Authenticators(routertest.DenyAll{}), Logger: logging.New(&bytes.Buffer{}, nil),
			})

			assert.ErrorContains(t, err, "listIncomes")
		})
	}
}

func TestRoutertestDenyAll_answersTheUnauthorizedEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	called := false

	routertest.DenyAll{}.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, apperr.CodeUnauthorized, envelope(t, rec).Code)
}
