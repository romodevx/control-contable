// Package dashboardhttp is the dashboard's HTTP adapter: it translates
// between the contract (docs/api/openapi.yaml, module_dashboard) and the
// dashboard use case.
package dashboardhttp

import (
	"context"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const (
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
)

// DashboardReader is the dashboard use case (usecase.Service). D is the
// tax engine's presented estimate, encoded as is.
type DashboardReader[D any] interface {
	Get(ctx context.Context, userID string, req usecase.Request) (usecase.Dashboard[D], error)
}

// Handler implements getDashboard of gen.StrictServerInterface (M6-03).
// cmd/api embeds it in the API server. The user id comes from the
// principal the router's authenticator verified; use case errors are
// client-ready *apperr.Error values returned as is (the router writes the
// envelope, an infrastructure failure becomes the generic 500).
type Handler[D any] struct {
	dashboards DashboardReader[D]
}

// NewHandler returns the handler over the use case.
func NewHandler[D any](dashboards DashboardReader[D]) *Handler[D] {
	return &Handler[D]{dashboards: dashboards}
}

// GetDashboard implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/dashboard). The router has validated the
// path id and the query against the contract; no period is the current
// month and no months the default, which the use case resolves.
func (h *Handler[D]) GetDashboard(ctx context.Context, req gen.GetDashboardRequestObject) (gen.GetDashboardResponseObject, error) {
	userID, err := principal.UserID(ctx)
	if err != nil {
		return gen.GetDashboard401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	in := usecase.Request{TaxpayerID: req.TaxpayerId.String()}
	if req.Params.Period != nil {
		in.Period = *req.Params.Period
	}
	if req.Params.Months != nil {
		in.Months = *req.Params.Months
	}
	d, err := h.dashboards.Get(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	return okResponse{body: toDashboard(d)}, nil
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
