package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// hangingManager is a tx.Manager over a database that never answers: every
// unit of work blocks until its ctx is done (P0-29 FU1 scenario).
type hangingManager struct{}

func (hangingManager) ReadWrite(ctx context.Context, _ string, _ tx.Work) error {
	<-ctx.Done()
	return ctx.Err()
}

func (hangingManager) ReadOnly(ctx context.Context, _ string, _ tx.Work) error {
	<-ctx.Done()
	return ctx.Err()
}

// dbBackedHandler stands in for a handler whose use case runs a unit of work.
func dbBackedHandler(txm tx.Manager) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := txm.ReadWrite(r.Context(), "0190f2a4-7c1e-7b3a-9d2e-4f5a6b7c8d9e", func(context.Context) error { return nil })
		if err != nil {
			_ = httperr.Write(w, r, apperr.Mapper{}, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

// A hung database cannot hold a request past the deadline: the unit of work
// is cancelled and the client gets the 500 envelope at the deadline.
func TestDeadline_cancelsAHangingUnitOfWork(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const timeout = 10 * time.Second
		h := middleware.Deadline(timeout)(dbBackedHandler(hangingManager{}))
		start := time.Now()

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/register", nil))

		assert.Equal(t, timeout, time.Since(start), "answered exactly at the deadline")
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, apperr.CodeInternalError, decodeEnvelope(t, rec).Error.Code)
	})
}

func TestDeadline_setsTheRequestContextDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const timeout = 3 * time.Second
		var got time.Time
		h := middleware.Deadline(timeout)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			var ok bool
			got, ok = r.Context().Deadline()
			require.True(t, ok)
		}))
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
		assert.Equal(t, time.Now().Add(timeout), got)
	})
}

// A client-supplied shorter deadline (an earlier cancellation) still wins.
func TestDeadline_keepsAnEarlierParentDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		var got time.Time
		h := middleware.Deadline(10 * time.Second)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
			got, _ = r.Context().Deadline()
		}))
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(parent, http.MethodGet, "/healthz", nil))
		assert.Equal(t, time.Now().Add(time.Second), got)
	})
}

func TestDeadline_rejectsANonPositiveTimeout(t *testing.T) {
	assert.Panics(t, func() { middleware.Deadline(0) })
}
