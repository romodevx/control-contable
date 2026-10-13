// Package flaggate is the server-side feature-flag gate (P0-15, ADR-0001,
// feature-flags rule): a request for a module whose flag is OFF answers 404
// with the standard ErrorResponse, before the generated wrapper binds path or
// query parameters and before the strict handler decodes the body. A 400 from
// either would reveal that the disabled module exists.
//
// The gate therefore sits at route registration: Mux is the gen.ServeMux handed
// to the generated server, and wraps every handler by the pattern it registers
// (gen.RouteModule). Middleware outside the mux sees no r.Pattern yet, and
// gen.MiddlewareFunc / strict middlewares run after binding or decoding
// (pinned by adapter/http/gen/gate_placement_test.go).
package flaggate

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
)

// ErrUnknownRoute is a registration of a pattern the generated flag map does
// not know (e.g. a non-empty BaseURL). It is a startup error: such a route is
// never served, gated or not.
var ErrUnknownRoute = errors.New("flaggate: route has no x-module")

const notFoundMessage = "Not found."

// Checker reports whether a module's flag is ON. *flags.Service implements it
// and fails closed.
type Checker interface {
	Enabled(ctx context.Context, module string) bool
}

// Mux is a gen.ServeMux that gates each route by its module at registration
// and answers every unmatched request with the same 404 envelope (never the
// ServeMux's 405 + Allow, which would list a disabled module's methods).
type Mux struct {
	mux      *http.ServeMux
	flags    Checker
	failures []error
}

// NewMux returns an empty gating Mux.
func NewMux(flags Checker) *Mux {
	return &Mux{mux: http.NewServeMux(), flags: flags}
}

// HandleFunc implements gen.ServeMux. A pattern without an x-module is
// recorded (see Err) and not registered.
func (m *Mux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	module, ok := gen.RouteModule(pattern)
	if !ok {
		m.failures = append(m.failures, fmt.Errorf("%w: %q", ErrUnknownRoute, pattern))
		return
	}
	m.mux.Handle(pattern, Wrap(m.flags, module, http.HandlerFunc(handler)))
}

// Err returns the registration failures; the composition root must refuse to
// start when it is non-nil.
func (m *Mux) Err() error { return errors.Join(m.failures...) }

// ServeHTTP implements http.Handler.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, pattern := m.mux.Handler(r); pattern == "" {
		WriteNotFound(w, r)
		return
	}
	m.mux.ServeHTTP(w, r)
}

// Wrap gates next behind module's flag. Platform routes are never gated.
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

// WriteNotFound writes the standard 404 ErrorResponse.
func WriteNotFound(w http.ResponseWriter, r *http.Request) {
	_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeNotFound, notFoundMessage))
}

// Handler registers every operation of si on a gating Mux and returns it, or
// the registration failures.
func Handler(si gen.ServerInterface, flags Checker) (http.Handler, error) {
	mux := NewMux(flags)
	handler := gen.HandlerFromMux(si, mux)
	if err := mux.Err(); err != nil {
		return nil, err
	}
	return handler, nil
}
