package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Deadline bounds every request with a context deadline (P0-29 FU1).
// http.Server.WriteTimeout does not cancel r.Context(), so without it a hung
// database would hold a handler (and its pooled connection) until the client
// gives up. Database work honours the context, so a unit of work still
// running at the deadline is rolled back and the handler answers the 500
// envelope. Keep timeout below the server's WriteTimeout, so the client
// receives that answer. An earlier deadline already on the request wins.
// It panics on a non-positive timeout: config validates it first.
func Deadline(timeout time.Duration) Func {
	if timeout <= 0 {
		panic(fmt.Sprintf("middleware: deadline %s must be positive", timeout))
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
