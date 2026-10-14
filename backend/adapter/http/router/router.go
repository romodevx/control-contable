// Package router builds the HTTP handler of the generated strict server and
// attaches the per-operation pipeline at route-registration time:
//
//	flag gate → client rate limit → authentication → user rate limit → contract validation → generated handler
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
// caller cannot tell a disabled module (404) from an enabled one (401). The
// per-operation client limits (MODULE-auth §5) run right after the gate: a
// disabled module spends no budget and never answers 429, and a flood of
// anonymous or malformed requests is limited before any authentication or
// validation work. The per-user limits (MODULE-taxpayers S-10) need the
// principal, so they run right after authentication.
package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
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
// request context. The bearer JWT authenticator (authhttp.Authenticator,
// M1-07) implements it; there is deliberately no pass-through one.
type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

// Options configures New. Flags, Authn and Logger are required.
type Options struct {
	Flags  FlagGate
	Authn  Authenticator
	Logger *slog.Logger
	// MaxBodyBytes caps the request body; 0 means DefaultMaxBodyBytes.
	MaxBodyBytes int64
	// NotBuilt lists the operationIds whose handler does not exist yet
	// (P0-29 FU2). They answer the hidden-module 404 whatever their flag,
	// before authentication, validation or the handler, so turning a module
	// ON before it is built shows nothing new and never reaches a stub.
	NotBuilt []string
	// OperationLimits are per-client limits by operationId (MODULE-auth §5),
	// keyed by ClientKey, which is required when any is set.
	OperationLimits map[string]Limiter
	// ClientKey derives the client key of a request (the client address).
	ClientKey func(*http.Request) string
	// UserLimits are per-user limits by operationId (MODULE-taxpayers
	// S-10), keyed by the authenticated user's id. They run right after
	// authentication, so only secured operations may have one.
	UserLimits map[string]Limiter
}

// New returns the handler serving every operation of spec through si. It
// fails when a route the generated server registers has no x-module or no
// operation in spec: an unknown route is a startup error, never an ungated
// or unvalidated one.
func New(si gen.StrictServerInterface, spec *openapi3.T, opts Options) (http.Handler, error) {
	if opts.Authn == nil {
		return nil, errors.New("router: no authenticator: secured operations would run unauthenticated")
	}
	if opts.MaxBodyBytes == 0 {
		opts.MaxBodyBytes = DefaultMaxBodyBytes
	}
	errs := errorWriter{logger: opts.Logger}
	strict := gen.NewStrictHandlerWithOptions(si, nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  errs.badRequest,
		ResponseErrorHandlerFunc: errs.handlerError,
	})
	declared := declaredOperations(spec)
	notBuilt, err := knownOperations(declared, opts.NotBuilt)
	if err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}
	if err := checkLimits(opts, declared, publicOperations(spec)); err != nil {
		return nil, fmt.Errorf("router: %w", err)
	}
	reg := &registrar{mux: http.NewServeMux(), spec: spec, opts: opts, errs: errs, notBuilt: notBuilt}
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
	notBuilt map[string]bool // operationIds answered with the hidden-module 404
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
	if g.notBuilt[route.Operation.OperationID] {
		g.mux.Handle(pattern, http.HandlerFunc(flaggate.WriteNotFound))
		return
	}
	var next http.Handler = http.HandlerFunc(handler)
	next = validate(route, g.opts.MaxBodyBytes, next)
	if limiter, ok := g.opts.UserLimits[route.Operation.OperationID]; ok {
		next = limit(limiter, principalKey, g.opts.Logger, next)
	}
	if !isPublic(route) {
		next = g.opts.Authn.Authenticate(next)
	}
	if limiter, ok := g.opts.OperationLimits[route.Operation.OperationID]; ok {
		next = limit(limiter, g.opts.ClientKey, g.opts.Logger, next)
	}
	g.mux.Handle(pattern, flaggate.Wrap(g.opts.Flags, module, next))
}

// ServeHTTP implements http.Handler. A request that matches no pattern gets
// the 404 envelope; the mux's own 404 text and its 405 + Allow (which would
// list a disabled module's methods) are never sent. Neither is the mux's 307
// to the cleaned path ("/v1//auth/login" -> "/v1/auth/login"): ServeMux sends
// it only when the cleaned path matches a route, so it would reveal disabled
// routes too. A non-canonical path is a 404 for every route.
func (g *registrar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !isCanonicalPath(r.URL.EscapedPath()) {
		flaggate.WriteNotFound(w, r)
		return
	}
	if _, pattern := g.mux.Handler(r); pattern == "" {
		flaggate.WriteNotFound(w, r)
		return
	}
	g.mux.ServeHTTP(w, r)
}

// isCanonicalPath reports whether p is already in the form ServeMux cleans
// paths to (rooted; no "//", "." or ".." elements; a trailing slash kept), so
// the mux never answers with a redirect.
func isCanonicalPath(p string) bool {
	if !strings.HasPrefix(p, "/") {
		return false
	}
	clean := path.Clean(p)
	if strings.HasSuffix(p, "/") && clean != "/" {
		clean += "/"
	}
	return clean == p
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

// declaredOperations is the set of the contract's operationIds.
func declaredOperations(spec *openapi3.T) map[string]bool {
	declared := map[string]bool{}
	for _, item := range spec.Paths.Map() {
		for _, op := range item.Operations() {
			declared[op.OperationID] = true
		}
	}
	return declared
}

// publicOperations returns the operationIds with no security requirement.
func publicOperations(spec *openapi3.T) map[string]bool {
	public := map[string]bool{}
	for path, item := range spec.Paths.Map() {
		for method, op := range item.Operations() {
			public[op.OperationID] = isPublic(&routers.Route{Spec: spec, Path: path, PathItem: item, Method: method, Operation: op})
		}
	}
	return public
}

// knownOperations returns ids as a set, failing on an id the contract does
// not declare: a typo would otherwise leave a stub reachable.
func knownOperations(declared map[string]bool, ids []string) (map[string]bool, error) {
	set := make(map[string]bool, len(ids))
	var unknown []error
	for _, id := range ids {
		if !declared[id] {
			unknown = append(unknown, fmt.Errorf("not-built operation %q is not in the contract", id))
		}
		set[id] = true
	}
	return set, errors.Join(unknown...)
}

// isPublic reports whether the operation opts out of security with
// `security: []` (the contract pairs it with x-public: true).
func isPublic(route *routers.Route) bool {
	security := route.Operation.Security
	return security != nil && len(*security) == 0
}
