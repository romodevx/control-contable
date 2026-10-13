// Package router builds the HTTP handler of the generated strict server and
// attaches the per-operation pipeline at route-registration time:
//
//	flag gate → authentication → contract validation → generated handler
//
// ADR-0001 (P0-06 notes) explains the placement: a middleware outside the
// mux sees no route yet, and the generated wrapper binds path and query
// parameters (and the strict handler decodes the body) before any
// gen.MiddlewareFunc or strict middleware runs, so either placement would
// answer 400 for a disabled module and reveal it. Wrapping each handler as
// the generated code registers it puts the gate first. Every request that
// matches no route (unknown path, or another method on a known path) gets the
// same 404 envelope as a disabled module.
//
// The gate runs before authentication, unlike the C4 listing, so an anonymous
// caller cannot tell a disabled module (404) from an enabled one (401).
package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"

	"github.com/romodevx/control-contable/backend/adapter/http/flaggate"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
)

// DefaultMaxBodyBytes caps request bodies read for validation. The largest
// request in the contract (an income or expense) is well under 64 KiB.
const DefaultMaxBodyBytes int64 = 1 << 20

// FlagGate reports whether a module's feature flag is ON. The flags service
// (P0-15) implements it; it must fail closed (OFF) when it cannot tell.
type FlagGate interface {
	Enabled(ctx context.Context, module string) bool
}

// AllModulesOff is a FlagGate with every module OFF (flags default OFF), for
// tests. cmd/api wires the flags service (*flags.Service).
type AllModulesOff struct{}

// Enabled implements FlagGate.
func (AllModulesOff) Enabled(context.Context, string) bool { return false }

// Authenticator guards operations that declare a security requirement. It
// writes the 401 envelope itself or calls next with the principal in the
// request context. The JWT authenticator (M1-07) implements it.
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

// NoAuthentication is the Authenticator until M1-07: it lets every request
// through. No secured operation is reachable before then because their
// modules' flags are OFF.
type NoAuthentication struct{}

// Authenticate implements Authenticator.
func (NoAuthentication) Authenticate(next http.Handler) http.Handler { return next }

// Options configures New. Flags, Authn and Logger are required.
type Options struct {
	Flags  FlagGate
	Authn  Authenticator
	Logger *slog.Logger
	// MaxBodyBytes caps the request body; 0 means DefaultMaxBodyBytes.
	MaxBodyBytes int64
}

// New returns the handler serving every operation of spec through si. It
// fails when a route the generated server registers has no x-module or no
// operation in spec: an unknown route is a startup error, never an ungated
// or unvalidated one.
func New(si gen.StrictServerInterface, spec *openapi3.T, opts Options) (http.Handler, error) {
	if opts.MaxBodyBytes == 0 {
		opts.MaxBodyBytes = DefaultMaxBodyBytes
	}
	errs := errorWriter{logger: opts.Logger}
	strict := gen.NewStrictHandlerWithOptions(si, nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  errs.badRequest,
		ResponseErrorHandlerFunc: errs.handlerError,
	})
	reg := &registrar{mux: http.NewServeMux(), spec: spec, opts: opts, errs: errs}
	handler := gen.HandlerWithOptions(strict, gen.StdHTTPServerOptions{
		BaseRouter:       reg,
		ErrorHandlerFunc: errs.badRequest,
	})
	if err := errors.Join(reg.failures...); err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}
	return handler, nil
}

// registrar is the gen.ServeMux handed to the generated server: it wraps each
// handler with the operation pipeline as the handler is registered.
type registrar struct {
	mux      *http.ServeMux
	spec     *openapi3.T
	opts     Options
	errs     errorWriter
	failures []error
}

// HandleFunc implements gen.ServeMux.
func (g *registrar) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	module, ok := gen.RouteModule(pattern)
	if !ok {
		g.failures = append(g.failures, fmt.Errorf("route %q has no x-module", pattern))
		return
	}
	route, err := findRoute(g.spec, pattern)
	if err != nil {
		g.failures = append(g.failures, err)
		return
	}
	var next http.Handler = http.HandlerFunc(handler)
	next = validate(route, g.opts.MaxBodyBytes, next)
	if !isPublic(route) {
		next = g.opts.Authn.Authenticate(next)
	}
	g.mux.Handle(pattern, flaggate.Wrap(g.opts.Flags, module, next))
}

// ServeHTTP implements http.Handler. A request that matches no pattern gets
// the 404 envelope; the mux's own 404 text and its 405 + Allow (which would
// list a disabled module's methods) are never sent.
func (g *registrar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, pattern := g.mux.Handler(r); pattern == "" {
		flaggate.WriteNotFound(w, r)
		return
	}
	g.mux.ServeHTTP(w, r)
}

// findRoute resolves a ServeMux pattern ("GET /v1/x/{id}") to its operation.
func findRoute(spec *openapi3.T, pattern string) (*routers.Route, error) {
	method, path, ok := strings.Cut(pattern, " ")
	if !ok {
		return nil, fmt.Errorf("route %q is not METHOD /path", pattern)
	}
	item := spec.Paths.Value(path)
	if item == nil || item.GetOperation(method) == nil {
		return nil, fmt.Errorf("route %q has no operation in the contract", pattern)
	}
	return &routers.Route{Spec: spec, Path: path, PathItem: item, Method: method, Operation: item.GetOperation(method)}, nil
}

// isPublic reports whether the operation opts out of security with
// `security: []` (the contract pairs it with x-public: true).
func isPublic(route *routers.Route) bool {
	security := route.Operation.Security
	return security != nil && len(*security) == 0
}
