// Package flaggate is the server-side feature-flag gate (P0-15, ADR-0001,
// feature-flags rule): a request for a module whose flag is OFF answers 404
// with the standard ErrorResponse. It is the security boundary; the client's
// use of GET /v1/flags is UX only.
//
// Where the gate sits matters as much as what it does: adapter/http/router
// applies Wrap to each route as the generated server registers it (keyed by
// gen.RouteModule(pattern)), so the 404 comes before the generated wrapper
// binds path or query parameters and before the strict handler decodes the
// body. A 400 from either would reveal that the disabled module exists
// (ADR-0001 P0-06 note, adapter/http/gen/gate_placement_test.go).
package flaggate

import (
	"context"
	"net/http"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

const notFoundMessage = "Not found."

// Checker reports whether a module's flag is ON. *flags.Service implements it
// and fails closed.
type Checker interface {
	Enabled(ctx context.Context, module string) bool
}

// Wrap gates next behind module's flag: OFF answers the 404 envelope and next
// never runs. Platform routes (gen.ModulePlatform) are never gated.
func Wrap(flags Checker, module string, next http.Handler) http.Handler {
	if module == gen.ModulePlatform {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !flags.Enabled(r.Context(), module) {
			WriteNotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// WriteNotFound writes the standard 404 ErrorResponse, the same answer for a
// disabled module and for a route that does not exist.
func WriteNotFound(w http.ResponseWriter, r *http.Request) {
	_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeNotFound, notFoundMessage))
}
