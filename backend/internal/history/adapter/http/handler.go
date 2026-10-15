// Package historyhttp is History's HTTP adapter: it translates between the
// contract (docs/api/openapi.yaml, module_history) and the history use
// case.
package historyhttp

import (
	"context"
	"errors"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const (
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
	// msgInvalidInput is the router's wording for every validation_failed.
	msgInvalidInput = "One or more fields are invalid."
)

// errMissingBody reports a markHistoryViewed without its required body
// (the router refuses it first; this is the backstop).
var errMissingBody = errors.New("history: missing request body")

// History is the history use case (usecase.Service).
type History interface {
	Get(ctx context.Context, userID, taxpayerID string, fiscalYear int) (usecase.Year, error)
	MarkViewed(ctx context.Context, userID, taxpayerID string, req usecase.ViewedRequest) error
}

// Handler implements getHistory and markHistoryViewed of
// gen.StrictServerInterface (M8-04). cmd/api embeds it in the API server.
// The user id comes from the principal the router's authenticator
// verified; use case errors are client-ready *apperr.Error values returned
// as is (the router writes the envelope, an infrastructure failure becomes
// the generic 500).
type Handler struct {
	history History
}

// NewHandler returns the handler over the use case.
func NewHandler(history History) *Handler { return &Handler{history: history} }

// GetHistory implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/history). The router has validated the
// path id and fiscal_year against the contract.
func (h *Handler) GetHistory(ctx context.Context, req gen.GetHistoryRequestObject) (gen.GetHistoryResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.GetHistory401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	year, err := h.history.Get(ctx, userID, req.TaxpayerId.String(), req.Params.FiscalYear)
	if err != nil {
		return nil, err
	}
	return okResponse{body: toYear(year)}, nil
}

// MarkHistoryViewed implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/history/{period}/viewed): 204 once the
// view is stored.
func (h *Handler) MarkHistoryViewed(ctx context.Context, req gen.MarkHistoryViewedRequestObject) (gen.MarkHistoryViewedResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.MarkHistoryViewed401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, apperr.Wrap(errMissingBody, apperr.CodeValidationFailed, msgInvalidInput).
			WithFields(apperr.FieldError{Field: "body", Code: "required", Message: "Falta el cuerpo de la solicitud."})
	}
	b := req.Body
	err = h.history.MarkViewed(ctx, userID, req.TaxpayerId.String(), usecase.ViewedRequest{
		Period: req.Period, RecordsFingerprint: b.RecordsFingerprint, FiguresFingerprint: b.FiguresFingerprint,
		EngineVersion: b.EngineVersion, ParamsYear: b.ParameterSet.FiscalYear, ParamsVersion: b.ParameterSet.Version,
	})
	if err != nil {
		return nil, err
	}
	// Cache-Control: no-store is written by middleware.NoStore, as on every response.
	return gen.MarkHistoryViewed204Response{}, nil
}

// unauthorized is the contract's Unauthorized response: the generic 401
// with the bearer challenge.
func unauthorized(ctx context.Context) gen.UnauthorizedJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeUnauthorized
	body.Error.Message = msgAuthenticationRequired
	body.Error.RequestId = requestid.From(ctx)
	return gen.UnauthorizedJSONResponse{Body: body, Headers: gen.UnauthorizedResponseHeaders{WWWAuthenticate: bearerChallenge}}
}
