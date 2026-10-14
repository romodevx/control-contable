package middleware_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
)

const legalCacheControl = "public, max-age=300"

func TestNoStore(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
	}{
		{
			name:    "a body written without WriteHeader is not stored",
			handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{}")) },
			want:    "no-store",
		},
		{
			name:    "a 204 is not stored",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) },
			want:    "no-store",
		},
		{
			name: "an error envelope is not stored",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeNotFound, "Not found."))
			},
			want: "no-store",
		},
		{
			name: "a success may declare its own caching (public legal documents)",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Cache-Control", legalCacheControl)
				w.WriteHeader(http.StatusOK)
			},
			want: legalCacheControl,
		},
		{
			name: "an error is never cacheable, whatever the handler set first",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Cache-Control", legalCacheControl)
				_ = httperr.Write(w, r, apperr.Mapper{}, errors.New("boom"))
			},
			want: "no-store",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			middleware.NoStore(tt.handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			assert.Equal(t, []string{tt.want}, rec.Header().Values("Cache-Control"))
		})
	}
}

// A panic recovered inside NoStore still answers an uncacheable 500, even
// when the handler had marked the response public before panicking.
func TestNoStore_RecoveredPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	panicking := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", legalCacheControl)
		panic("boom")
	})
	rec := httptest.NewRecorder()
	middleware.Chain(panicking, middleware.NoStore, middleware.Recover(logger)).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

// NoStore keeps the writer reachable for http.ResponseController.
func TestNoStore_Unwrap(t *testing.T) {
	var flushErr error
	h := middleware.NoStore(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flushErr = http.NewResponseController(w).Flush()
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	assert.NoError(t, flushErr)
}
