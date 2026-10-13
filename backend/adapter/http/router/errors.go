package router

import (
	"log/slog"
	"net/http"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
)

const invalidInputMessage = "One or more fields are invalid."

// errorWriter writes the envelope for errors raised by the generated code
// and by the strict handlers.
type errorWriter struct {
	logger *slog.Logger
}

// badRequest answers a parameter-binding or body-decoding failure in the
// generated code. Contract validation runs first, so this is a backstop.
func (e errorWriter) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	_ = httperr.Write(w, r, apperr.Mapper{}, invalid(err, bindingFieldErrors(err)))
}

// handlerError answers an error returned by a strict handler: a typed
// *apperr.Error keeps its code and message, anything else is a generic 500.
// Server-side failures are logged with their cause.
func (e errorWriter) handlerError(w http.ResponseWriter, r *http.Request, err error) {
	resolved := httperr.Write(w, r, apperr.Mapper{}, err)
	if httperr.Status(resolved.Code) >= http.StatusInternalServerError {
		e.logger.ErrorContext(r.Context(), "request failed", slog.String("code", string(resolved.Code)), slog.Any("error", err))
	}
}

// invalid builds the validation_failed error for a rejected request.
func invalid(cause error, fields []apperr.FieldError) *apperr.Error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, invalidInputMessage).WithFields(fields...)
}
