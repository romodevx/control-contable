package middleware

import (
	"crypto/rand"
	"net/http"

	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

// RequestIDHeader carries the correlation id in requests and responses.
const RequestIDHeader = "X-Request-Id"

const (
	requestIDPrefix = "req_"
	maxRequestIDLen = 64
)

// RequestID puts a correlation id in the request context and echoes it in
// the response header, so the error envelope, the logs and the client report
// the same id. A well-formed inbound id (1-64 of [A-Za-z0-9._-]) is kept, so a
// client or proxy can correlate; anything else is replaced, which keeps
// arbitrary client text out of the logs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !validRequestID(id) {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(requestid.With(r.Context(), id)))
	})
}

// newRequestID returns "req_" and 128 bits of randomness in base32.
func newRequestID() string { return requestIDPrefix + rand.Text() }

func validRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLen {
		return false
	}
	for _, c := range []byte(id) {
		if !isRequestIDByte(c) {
			return false
		}
	}
	return true
}

func isRequestIDByte(c byte) bool {
	switch {
	case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z', '0' <= c && c <= '9':
		return true
	}
	return c == '.' || c == '_' || c == '-'
}
