package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// errPanic marks a recovered panic; its text reaches the logs only.
var errPanic = errors.New("panic recovered")

// Recover turns a panic into the contract's 500 internal_error envelope. The
// panic value and the stack go to the (redacting) logger, never to the
// client. If the handler already started the response, nothing more is
// written. http.ErrAbortHandler is re-raised, as net/http expects.
func Recover(logger *slog.Logger) Func {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tw := &trackingWriter{ResponseWriter: w}
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler { //nolint:errorlint // net/http compares the sentinel by identity
					panic(v)
				}
				logger.ErrorContext(r.Context(), "panic recovered",
					slog.Any("panic", v), slog.String("stack", string(debug.Stack())))
				if tw.started {
					return
				}
				_ = httperr.Write(w, r, apperr.Mapper{}, fmt.Errorf("%w: %v", errPanic, v))
			}()
			next.ServeHTTP(tw, r)
		})
	}
}

// trackingWriter records whether the response has started.
type trackingWriter struct {
	http.ResponseWriter
	started bool
}

func (w *trackingWriter) WriteHeader(code int) {
	w.started = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *trackingWriter) Write(b []byte) (int, error) {
	w.started = true
	return w.ResponseWriter.Write(b)
}

// Unwrap lets http.ResponseController reach Flush, deadlines, etc.
func (w *trackingWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
