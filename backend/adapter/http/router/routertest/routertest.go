// Package routertest holds router.Options helpers for the tests of packages
// that build the real router over the whole contract.
package routertest

import (
	"net/http"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
)

// DenyAll is a router.Authenticator that refuses every request with the 401
// envelope. It stands in for a scheme a test does not exercise.
type DenyAll struct{}

// Authenticate implements router.Authenticator.
func (DenyAll) Authenticate(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeUnauthorized, "Authentication required."))
	})
}

// Authenticators returns router.Options.Authenticators with bearer for the
// user scheme (bearerAuth) and DenyAll for every other contract scheme, so a
// test of user operations still builds every route.
func Authenticators(bearer router.Authenticator) map[string]router.Authenticator {
	return map[string]router.Authenticator{
		router.SchemeBearer:            bearer,
		router.SchemeRevenueCatWebhook: DenyAll{},
	}
}
