// Package incomeshttp is the incomes module's HTTP adapter: it translates
// between the contract (docs/api/openapi.yaml) and the income use cases.
package incomeshttp

import (
	"context"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const (
	// noStore is the Cache-Control of every response carrying income data
	// (components.headers.NoStore, MODULE-incomes S-11).
	noStore = "no-store"
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
)

// IncomeManager is the income use cases (usecase.Incomes, M3-04/M3-05 and
// the defaults query).
type IncomeManager interface {
	Create(ctx context.Context, userID string, in usecase.CreateInput) (usecase.WriteResult, error)
	Update(ctx context.Context, userID, taxpayerID, incomeID string, p usecase.Patch) (usecase.WriteResult, error)
	Get(ctx context.Context, userID, taxpayerID, incomeID string) (usecase.Income, error)
	List(ctx context.Context, userID string, q usecase.ListQuery) (usecase.Page, error)
	Delete(ctx context.Context, userID, taxpayerID, incomeID string) error
	Restore(ctx context.Context, userID, taxpayerID, incomeID string) (usecase.WriteResult, error)
	Defaults(ctx context.Context, userID string, q usecase.DefaultsQuery) (usecase.Defaults, error)
}

// Handler implements the module_incomes operations of
// gen.StrictServerInterface (M3-06). cmd/api embeds it in the API server.
//
// Every operation is secured: the user id comes from the principal the
// router's authenticator verified and the taxpayer id from the path, never
// from the body. Use case errors are client-ready *apperr.Error values (402
// naming the incomes cap, 409 id_reused, 422 regime_not_active, 404 across
// users, 400 with field details) and are returned as is; the router writes
// the envelope, and any other failure becomes the generic 500.
type Handler struct {
	incomes IncomeManager
}

// NewHandler returns the handler over the use cases.
func NewHandler(incomes IncomeManager) *Handler {
	return &Handler{incomes: incomes}
}

// ListIncomes implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/incomes).
func (h *Handler) ListIncomes(ctx context.Context, req gen.ListIncomesRequestObject) (gen.ListIncomesResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListIncomes401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	page, err := h.incomes.List(ctx, userID, toListQuery(req))
	if err != nil {
		return nil, err
	}
	body, err := toPage(page)
	if err != nil {
		return nil, err
	}
	return gen.ListIncomes200JSONResponse{Body: body, Headers: gen.ListIncomes200ResponseHeaders{CacheControl: noStore}}, nil
}

// CreateIncome implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/incomes): 201 when created, 200 for an
// idempotent retry of a client id the caller already owns.
func (h *Handler) CreateIncome(ctx context.Context, req gen.CreateIncomeRequestObject) (gen.CreateIncomeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.CreateIncome401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	in, err := toCreateInput(req.TaxpayerId.String(), req.Body)
	if err != nil {
		return nil, err
	}
	result, err := h.incomes.Create(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	body, err := toIncome(result.Income, result.Warnings)
	if err != nil {
		return nil, err
	}
	if !result.Created {
		return gen.CreateIncome200JSONResponse{Body: body, Headers: gen.CreateIncome200ResponseHeaders{CacheControl: noStore}}, nil
	}
	return gen.CreateIncome201JSONResponse{Body: body, Headers: gen.CreateIncome201ResponseHeaders{CacheControl: noStore}}, nil
}

// GetIncomeDefaults implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/incomes/defaults).
func (h *Handler) GetIncomeDefaults(ctx context.Context, req gen.GetIncomeDefaultsRequestObject) (gen.GetIncomeDefaultsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetIncomeDefaults401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	defaults, err := h.incomes.Defaults(ctx, userID, toDefaultsQuery(req))
	if err != nil {
		return nil, err
	}
	return gen.GetIncomeDefaults200JSONResponse{
		Body: toDefaults(defaults), Headers: gen.GetIncomeDefaults200ResponseHeaders{CacheControl: noStore},
	}, nil
}

// GetIncome implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/incomes/{incomeId}).
func (h *Handler) GetIncome(ctx context.Context, req gen.GetIncomeRequestObject) (gen.GetIncomeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetIncome401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	income, err := h.incomes.Get(ctx, userID, req.TaxpayerId.String(), req.IncomeId.String())
	if err != nil {
		return nil, err
	}
	body, err := toIncome(income, nil)
	if err != nil {
		return nil, err
	}
	return gen.GetIncome200JSONResponse{Body: body, Headers: gen.GetIncome200ResponseHeaders{CacheControl: noStore}}, nil
}

// UpdateIncome implements gen.StrictServerInterface
// (PATCH /v1/taxpayers/{taxpayerId}/incomes/{incomeId}): merge-patch, where
// null clears a date, the 625 activity, the folio or the description.
func (h *Handler) UpdateIncome(ctx context.Context, req gen.UpdateIncomeRequestObject) (gen.UpdateIncomeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UpdateIncome401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	patch, err := toPatch(req.Body)
	if err != nil {
		return nil, err
	}
	result, err := h.incomes.Update(ctx, userID, req.TaxpayerId.String(), req.IncomeId.String(), patch)
	if err != nil {
		return nil, err
	}
	body, err := toIncome(result.Income, result.Warnings)
	if err != nil {
		return nil, err
	}
	return gen.UpdateIncome200JSONResponse{Body: body, Headers: gen.UpdateIncome200ResponseHeaders{CacheControl: noStore}}, nil
}

// DeleteIncome implements gen.StrictServerInterface
// (DELETE /v1/taxpayers/{taxpayerId}/incomes/{incomeId}): a soft delete,
// 204 also when already deleted.
func (h *Handler) DeleteIncome(ctx context.Context, req gen.DeleteIncomeRequestObject) (gen.DeleteIncomeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.DeleteIncome401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err := h.incomes.Delete(ctx, userID, req.TaxpayerId.String(), req.IncomeId.String()); err != nil {
		return nil, err
	}
	return gen.DeleteIncome204Response{}, nil
}

// RestoreIncome implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore): Undo,
// never spending a trial slot.
func (h *Handler) RestoreIncome(ctx context.Context, req gen.RestoreIncomeRequestObject) (gen.RestoreIncomeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.RestoreIncome401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	result, err := h.incomes.Restore(ctx, userID, req.TaxpayerId.String(), req.IncomeId.String())
	if err != nil {
		return nil, err
	}
	body, err := toIncome(result.Income, result.Warnings)
	if err != nil {
		return nil, err
	}
	return gen.RestoreIncome200JSONResponse{Body: body, Headers: gen.RestoreIncome200ResponseHeaders{CacheControl: noStore}}, nil
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
