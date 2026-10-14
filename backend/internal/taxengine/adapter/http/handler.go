// Package taxenginehttp is the tax engine's HTTP adapter: it translates
// between the contract (docs/api/openapi.yaml, module_tax_engine) and the
// tax engine use cases.
package taxenginehttp

import (
	"context"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

const (
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
	// msgInvalidInput is the router's wording for every validation_failed.
	msgInvalidInput = "One or more fields are invalid."
	msgBodyRequired = "Send the values to store."
	fieldValues     = "values"
	detailRequired  = "required"
)

// InputsManager is the tax-inputs use case (usecase.Inputs).
type InputsManager interface {
	Get(ctx context.Context, userID string, req usecase.InputsRequest) (usecase.TaxInputsView, error)
	Put(ctx context.Context, userID string, req usecase.InputsRequest, values []domain.SubmittedInput) (usecase.TaxInputsView, error)
}

// EstimateReader is the estimate use case (usecase.Estimator).
type EstimateReader interface {
	EstimateMonth(ctx context.Context, userID, taxpayerID, rawPeriod string) (usecase.TaxpayerEstimate, error)
	Breakdown(ctx context.Context, userID, taxpayerID, rawPeriod, regimeCode string) (usecase.RegimeBreakdown, error)
}

// Handler implements the module_tax_engine operations of
// gen.StrictServerInterface: getTaxInputs, putTaxInputs (M5-21),
// getTaxpayerEstimate and getEstimateBreakdown (M5-23). cmd/api embeds it in
// the API server.
//
// Every operation is secured: the user id comes from the principal the
// router's authenticator verified, never from the request. Use case errors
// are client-ready *apperr.Error values and are returned as is; the router
// writes the envelope (an infrastructure failure becomes the generic 500).
type Handler struct {
	inputs    InputsManager
	estimates EstimateReader
}

// NewHandler returns the handler over the use cases.
func NewHandler(inputs InputsManager, estimates EstimateReader) *Handler {
	return &Handler{inputs: inputs, estimates: estimates}
}

// GetTaxpayerEstimate implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/estimates). No period is the current
// month, which the use case resolves.
func (h *Handler) GetTaxpayerEstimate(ctx context.Context, req gen.GetTaxpayerEstimateRequestObject) (gen.GetTaxpayerEstimateResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetTaxpayerEstimate401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	var period string
	if req.Params.Period != nil {
		period = *req.Params.Period
	}
	est, err := h.estimates.EstimateMonth(ctx, userID, req.TaxpayerId.String(), period)
	if err != nil {
		return nil, err
	}
	return okResponse{body: toEstimate(est)}, nil
}

// GetEstimateBreakdown implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown).
func (h *Handler) GetEstimateBreakdown(ctx context.Context, req gen.GetEstimateBreakdownRequestObject) (gen.GetEstimateBreakdownResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetEstimateBreakdown401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	b, err := h.estimates.Breakdown(ctx, userID, req.TaxpayerId.String(), req.Period, req.Params.Regime)
	if err != nil {
		return nil, err
	}
	return okResponse{body: toBreakdown(b)}, nil
}

// GetTaxInputs implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs).
func (h *Handler) GetTaxInputs(ctx context.Context, req gen.GetTaxInputsRequestObject) (gen.GetTaxInputsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetTaxInputs401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	view, err := h.inputs.Get(ctx, userID, inputsRequest(req.TaxpayerId.String(), req.RegimeCode, req.Params.FiscalYear))
	if err != nil {
		return nil, err
	}
	return taxInputsResponse{body: toTaxInputs(view)}, nil
}

// PutTaxInputs implements gen.StrictServerInterface
// (PUT /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs).
func (h *Handler) PutTaxInputs(ctx context.Context, req gen.PutTaxInputsRequestObject) (gen.PutTaxInputsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.PutTaxInputs401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil { // contract validation answers first; a backstop
		return nil, apperr.New(apperr.CodeValidationFailed, msgInvalidInput).
			WithFields(apperr.FieldError{Field: fieldValues, Code: detailRequired, Message: msgBodyRequired})
	}
	view, err := h.inputs.Put(ctx, userID, inputsRequest(req.TaxpayerId.String(), req.RegimeCode, req.Params.FiscalYear),
		toSubmitted(req.Body.Values))
	if err != nil {
		return nil, err
	}
	return taxInputsResponse{body: toTaxInputs(view)}, nil
}

func inputsRequest(taxpayerID, regimeCode string, fiscalYear int) usecase.InputsRequest {
	return usecase.InputsRequest{TaxpayerID: taxpayerID, RegimeCode: regimeCode, FiscalYear: fiscalYear}
}

func toSubmitted(values []gen.TaxInputValue) []domain.SubmittedInput {
	out := make([]domain.SubmittedInput, 0, len(values))
	for _, v := range values {
		out = append(out, domain.SubmittedInput{
			Key: domain.InputKey(v.Key), Month: v.Month, Cents: v.ValueCents, Rate: v.ValueRate, Bool: v.ValueBool,
		})
	}
	return out
}

// caller returns the authenticated user's id.
func caller(ctx context.Context) (string, bool) {
	userID, err := principal.UserID(ctx)
	return userID, err == nil
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
