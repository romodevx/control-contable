// Package middleware holds the process-wide HTTP middlewares that run before
// routing (docs/architecture/c4.md): request id → recover → rate limit.
// Per-operation concerns (flag gate, authentication, contract validation) run
// after routing, in adapter/http/router, because they depend on the matched
// operation.
package middleware

import (
	"net/http"
	"slices"
)

// Func is a standard net/http middleware.
type Func func(http.Handler) http.Handler

// Chain wraps h so that mws run in the order given: mws[0] is the outermost.
func Chain(h http.Handler, mws ...Func) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}
