package middleware

import "net/http"

const (
	cacheControlHeader = "Cache-Control"
	noStore            = "no-store"
)

// NoStore is the single owner of the API's caching policy (contract
// components.headers.NoStore; S-11, M3/M4 security gate G-1): every response
// carries Cache-Control: no-store, so no cache keeps tokens, account data,
// financial records or error envelopes. A successful response may declare its
// own value (the public legal documents: public, max-age=300); an error
// status (>= 400) is always no-store, whatever the handler set before failing.
func NoStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(cacheControlHeader, noStore)
		next.ServeHTTP(&noStoreWriter{ResponseWriter: w}, r)
	})
}

// noStoreWriter forces no-store on an error status when the status is sent.
type noStoreWriter struct {
	http.ResponseWriter
}

func (w *noStoreWriter) WriteHeader(code int) {
	if code >= http.StatusBadRequest {
		w.Header().Set(cacheControlHeader, noStore)
	}
	w.ResponseWriter.WriteHeader(code)
}

// Unwrap lets http.ResponseController reach Flush, deadlines, etc.
func (w *noStoreWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
