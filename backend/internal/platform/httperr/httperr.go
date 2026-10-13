// Package httperr is the HTTP edge of platform/apperr: it maps contract error
// codes to HTTP statuses and writes the single error envelope, ErrorResponse in
// docs/api/openapi.yaml.
package httperr

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

// statusByCode follows the response descriptions of the contract:
// 400 BadRequest, 401 Unauthorized, 402 PaymentRequired, 404 NotFound,
// 409 Conflict (state), 422 UnprocessableEntity (business rule), 429, 503.
var statusByCode = map[apperr.Code]int{
	apperr.CodeValidationFailed:         http.StatusBadRequest,
	apperr.CodeUnauthorized:             http.StatusUnauthorized,
	apperr.CodeEntitlementRequired:      http.StatusPaymentRequired,
	apperr.CodeNotFound:                 http.StatusNotFound,
	apperr.CodeConflict:                 http.StatusConflict,
	apperr.CodeRFCAlreadyRegistered:     http.StatusConflict,
	apperr.CodeRegimeHasRecords:         http.StatusConflict,
	apperr.CodeInvalidRegimeCombination: http.StatusUnprocessableEntity,
	apperr.CodeBusinessRuleViolation:    http.StatusUnprocessableEntity,
	apperr.CodeTaxpayerLimitReached:     http.StatusUnprocessableEntity,
	apperr.CodeRateLimited:              http.StatusTooManyRequests,
	apperr.CodeUpstreamUnavailable:      http.StatusServiceUnavailable,
	apperr.CodeInternalError:            http.StatusInternalServerError,
}

// Status returns the HTTP status for code; an unknown code is a 500.
func Status(code apperr.Code) int {
	if status, ok := statusByCode[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// Response is the contract's ErrorResponse.
type Response struct {
	Error Body `json:"error"`
}

// Body is the "error" object of ErrorResponse.
type Body struct {
	Code      apperr.Code  `json:"code"`
	Message   string       `json:"message"`
	RequestID string       `json:"request_id"`
	Details   []FieldError `json:"details,omitempty"`
}

// FieldError is the contract's FieldError.
type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewResponse builds the wire envelope for e. Only the client-safe parts of e
// (code, message, fields) are copied; the cause never reaches the wire.
func NewResponse(e *apperr.Error, requestID string) Response {
	body := Body{Code: e.Code, Message: e.Message, RequestID: requestID}
	for _, f := range e.Fields {
		body.Details = append(body.Details, FieldError(f))
	}
	return Response{Error: body}
}

// errNilWrite stands in for a nil error passed to Write: a programming error
// that is answered as internal_error rather than a panic.
var errNilWrite = errors.New("httperr: Write called with a nil error")

// Write resolves err with mapper and writes the envelope with its status and
// the request id from r's context. It returns the resolved error, which still
// wraps err, so the caller can log the cause. A nil err is a caller bug and
// is answered as internal_error. A 429
// needs a Retry-After header, which the caller sets before calling Write
// (middleware.RateLimit does).
func Write(w http.ResponseWriter, r *http.Request, mapper apperr.Mapper, err error) *apperr.Error {
	if err == nil {
		err = errNilWrite
	}
	resolved := mapper.Resolve(err)
	header := w.Header()
	header.Set("Content-Type", "application/json")
	header.Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(Status(resolved.Code))
	// The status line is already sent; a failed body write (client gone)
	// cannot be reported to the client, and the caller logs resolved anyway.
	_ = json.NewEncoder(w).Encode(NewResponse(resolved, requestid.From(r.Context())))
	return resolved
}
