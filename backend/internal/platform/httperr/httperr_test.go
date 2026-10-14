package httperr_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

func TestStatus(t *testing.T) {
	tests := []struct {
		code apperr.Code
		want int
	}{
		{apperr.CodeValidationFailed, http.StatusBadRequest},
		{apperr.CodeUnauthorized, http.StatusUnauthorized},
		{apperr.CodeEntitlementRequired, http.StatusPaymentRequired},
		{apperr.CodeNotFound, http.StatusNotFound},
		{apperr.CodeConflict, http.StatusConflict},
		{apperr.CodeRFCAlreadyRegistered, http.StatusConflict},
		{apperr.CodeRegimeHasRecords, http.StatusConflict},
		{apperr.CodeInvalidRegimeCombination, http.StatusUnprocessableEntity},
		{apperr.CodeBusinessRuleViolation, http.StatusUnprocessableEntity},
		{apperr.CodeTaxpayerLimitReached, http.StatusUnprocessableEntity},
		{apperr.CodeRateLimited, http.StatusTooManyRequests},
		{apperr.CodeUpstreamUnavailable, http.StatusServiceUnavailable},
		{apperr.CodeInternalError, http.StatusInternalServerError},
		{apperr.CodeInvalidPassword, http.StatusForbidden},
		{apperr.Code("not_in_contract"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.want, httperr.Status(tt.code))
		})
	}
}

// TestStatus_coversEveryCode fails when a contract code is added without
// deciding its HTTP status.
func TestStatus_coversEveryCode(t *testing.T) {
	for _, code := range apperr.Codes() {
		if code == apperr.CodeInternalError {
			continue
		}
		assert.NotEqual(t, http.StatusInternalServerError, httperr.Status(code), "no status for %s", code)
	}
}

var errRFCTaken = errors.New("taxpayer: rfc already registered")

func TestWrite(t *testing.T) {
	mapper := apperr.NewMapper(apperr.Rule{
		Target: errRFCTaken, Code: apperr.CodeRFCAlreadyRegistered, Message: "This RFC is already registered in your account.",
	})
	field := apperr.FieldError{Field: "amount_cents", Code: "minimum", Message: "Must be greater than or equal to 0."}
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "typed error with details",
			err:        apperr.New(apperr.CodeValidationFailed, "One or more fields are invalid.").WithFields(field),
			wantStatus: http.StatusBadRequest,
			wantBody: `{"error":{"code":"validation_failed","message":"One or more fields are invalid.",` +
				`"request_id":"req_01J9Z6Q4X8","details":[{"field":"amount_cents","code":"minimum",` +
				`"message":"Must be greater than or equal to 0."}]}}`,
		},
		{
			name:       "domain sentinel via mapper, details omitted",
			err:        fmt.Errorf("create taxpayer: %w", errRFCTaken),
			wantStatus: http.StatusConflict,
			wantBody: `{"error":{"code":"rfc_already_registered",` +
				`"message":"This RFC is already registered in your account.","request_id":"req_01J9Z6Q4X8"}}`,
		},
		{
			name:       "unexpected error never leaks its text",
			err:        errors.New("dial tcp 10.0.0.3:5432: user ana@example.com"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"internal_error","message":"Something went wrong.","request_id":"req_01J9Z6Q4X8"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/taxpayers", nil)
			req = req.WithContext(requestid.With(req.Context(), "req_01J9Z6Q4X8"))
			rec := httptest.NewRecorder()

			got := httperr.Write(rec, req, mapper, tt.err)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
			require.NotNil(t, got)
			assert.ErrorIs(t, got, tt.err, "Write returns the resolved error so the caller can log the cause")
		})
	}
}

// TestNewResponse_shape pins the wire format to the contract's ErrorResponse:
// exactly one top-level key and only the declared keys inside it.
// Regression (PR #9 follow-up 1): a nil error dereferenced a nil
// *apperr.Error and panicked. It is a programming error, answered as a 500
// internal_error instead of crashing the request.
func TestWrite_nilErrorIsInternalError(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/x", nil)
	w := httptest.NewRecorder()

	var resolved *apperr.Error
	require.NotPanics(t, func() { resolved = httperr.Write(w, r, apperr.Mapper{}, nil) })

	require.NotNil(t, resolved)
	assert.Equal(t, apperr.CodeInternalError, resolved.Code)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var got httperr.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, apperr.CodeInternalError, got.Error.Code)
}

func TestNewResponse_shape(t *testing.T) {
	resp := httperr.NewResponse(apperr.New(apperr.CodeNotFound, "Not found."), "req_1")

	raw, err := json.Marshal(resp)
	require.NoError(t, err)
	var top map[string]map[string]any
	require.NoError(t, json.Unmarshal(raw, &top))

	require.Len(t, top, 1)
	assert.Equal(t, map[string]any{"code": "not_found", "message": "Not found.", "request_id": "req_1"}, top["error"])
}
