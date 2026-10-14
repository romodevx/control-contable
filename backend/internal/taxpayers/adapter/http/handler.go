package taxpayershttp

import (
	"context"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const (
	// noStore is the Cache-Control of every response carrying taxpayer data
	// (components.headers.NoStore).
	noStore = "no-store"
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
)

// TaxpayerManager is the taxpayer use case (usecase.Taxpayers).
type TaxpayerManager interface {
	Create(ctx context.Context, userID string, in usecase.CreateInput) (usecase.CreateResult, error)
	List(ctx context.Context, userID string, includeArchived bool) ([]usecase.Taxpayer, error)
	Get(ctx context.Context, userID, id string) (usecase.Taxpayer, error)
	Update(ctx context.Context, userID, id string, in usecase.UpdateInput) (usecase.Taxpayer, error)
	Archive(ctx context.Context, userID, id string) (usecase.Taxpayer, error)
	Unarchive(ctx context.Context, userID, id string) (usecase.Taxpayer, error)
	Delete(ctx context.Context, userID, id string, confirm bool) error
}

// RegimeManager is the regime assignment use case (usecase.Regimes).
type RegimeManager interface {
	Add(ctx context.Context, userID, taxpayerID string, in usecase.RegimeInput) (usecase.TaxpayerRegime, error)
	UpdateOptions(ctx context.Context, userID, taxpayerID, code string, patch map[string]domain.OptionValue) (usecase.TaxpayerRegime, error)
	End(ctx context.Context, userID, taxpayerID string, in usecase.EndRegimeInput) (usecase.TaxpayerRegime, error)
}

// CatalogLister lists the regime catalog (usecase.Catalog).
type CatalogLister interface {
	List(ctx context.Context, userID, personType string) ([]usecase.CatalogEntry, error)
}

// UseCases are the use cases Handler serves. All are required.
type UseCases struct {
	Taxpayers TaxpayerManager
	Regimes   RegimeManager
	Catalog   CatalogLister
}

// Handler implements the module_taxpayers operations of
// gen.StrictServerInterface (M2-07). cmd/api embeds it in the API server.
//
// Every operation is secured: the user id comes from the principal the
// router's authenticator verified, never from the request. Use case errors
// are client-ready *apperr.Error values and are returned as is; the router
// writes the envelope (an infrastructure failure becomes the generic 500).
type Handler struct {
	taxpayers TaxpayerManager
	regimes   RegimeManager
	catalog   CatalogLister
}

// NewHandler returns the handler over the use cases.
func NewHandler(uc UseCases) *Handler {
	return &Handler{taxpayers: uc.Taxpayers, regimes: uc.Regimes, catalog: uc.Catalog}
}

// ListRegimes implements gen.StrictServerInterface (GET /v1/regimes).
func (h *Handler) ListRegimes(ctx context.Context, req gen.ListRegimesRequestObject) (gen.ListRegimesResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListRegimes401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	entries, err := h.catalog.List(ctx, userID, string(req.Params.PersonType))
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.ListRegimes200JSONResponse(toRegimeList(entries)), nil
}

// ListTaxpayers implements gen.StrictServerInterface (GET /v1/taxpayers).
func (h *Handler) ListTaxpayers(ctx context.Context, req gen.ListTaxpayersRequestObject) (gen.ListTaxpayersResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListTaxpayers401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	includeArchived := req.Params.IncludeArchived != nil && *req.Params.IncludeArchived
	list, err := h.taxpayers.List(ctx, userID, includeArchived)
	if err != nil {
		return nil, toAppError(err)
	}
	items := make([]gen.Taxpayer, 0, len(list))
	for _, t := range list {
		item, err := toTaxpayer(t)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return gen.ListTaxpayers200JSONResponse{
		Body: gen.TaxpayerList{Items: items}, Headers: gen.ListTaxpayers200ResponseHeaders{CacheControl: noStore},
	}, nil
}

// CreateTaxpayer implements gen.StrictServerInterface (POST /v1/taxpayers):
// 201 when created, 200 for an idempotent retry of a client id the caller
// already owns (the stored taxpayer, whatever the retry's body).
func (h *Handler) CreateTaxpayer(ctx context.Context, req gen.CreateTaxpayerRequestObject) (gen.CreateTaxpayerResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.CreateTaxpayer401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	in, err := toCreateInput(req.Body)
	if err != nil {
		return nil, err
	}
	result, err := h.taxpayers.Create(ctx, userID, in)
	if err != nil {
		return nil, toAppError(err)
	}
	body, err := toTaxpayer(result.Taxpayer)
	if err != nil {
		return nil, err
	}
	if !result.Created {
		return gen.CreateTaxpayer200JSONResponse{Body: body, Headers: gen.CreateTaxpayer200ResponseHeaders{CacheControl: noStore}}, nil
	}
	return gen.CreateTaxpayer201JSONResponse{Body: body, Headers: gen.CreateTaxpayer201ResponseHeaders{CacheControl: noStore}}, nil
}

// GetTaxpayer implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}).
func (h *Handler) GetTaxpayer(ctx context.Context, req gen.GetTaxpayerRequestObject) (gen.GetTaxpayerResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetTaxpayer401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	t, err := h.taxpayers.Get(ctx, userID, req.TaxpayerId.String())
	if err != nil {
		return nil, toAppError(err)
	}
	body, err := toTaxpayer(t)
	if err != nil {
		return nil, err
	}
	return gen.GetTaxpayer200JSONResponse{Body: body, Headers: gen.GetTaxpayer200ResponseHeaders{CacheControl: noStore}}, nil
}

// UpdateTaxpayer implements gen.StrictServerInterface
// (PATCH /v1/taxpayers/{taxpayerId}): a rename and/or an RFC correction.
func (h *Handler) UpdateTaxpayer(ctx context.Context, req gen.UpdateTaxpayerRequestObject) (gen.UpdateTaxpayerResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UpdateTaxpayer401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	var in usecase.UpdateInput
	if req.Body != nil {
		in = usecase.UpdateInput{DisplayName: req.Body.DisplayName, RFC: req.Body.Rfc}
	}
	t, err := h.taxpayers.Update(ctx, userID, req.TaxpayerId.String(), in)
	if err != nil {
		return nil, toAppError(err)
	}
	body, err := toTaxpayer(t)
	if err != nil {
		return nil, err
	}
	return gen.UpdateTaxpayer200JSONResponse{Body: body, Headers: gen.UpdateTaxpayer200ResponseHeaders{CacheControl: noStore}}, nil
}

// DeleteTaxpayer implements gen.StrictServerInterface
// (DELETE /v1/taxpayers/{taxpayerId}?confirm=true): 204 once deleted.
func (h *Handler) DeleteTaxpayer(ctx context.Context, req gen.DeleteTaxpayerRequestObject) (gen.DeleteTaxpayerResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.DeleteTaxpayer401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err := h.taxpayers.Delete(ctx, userID, req.TaxpayerId.String(), confirmed(req.Params.Confirm)); err != nil {
		return nil, toAppError(err)
	}
	return gen.DeleteTaxpayer204Response{}, nil
}

// ArchiveTaxpayer implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/archive).
func (h *Handler) ArchiveTaxpayer(ctx context.Context, req gen.ArchiveTaxpayerRequestObject) (gen.ArchiveTaxpayerResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ArchiveTaxpayer401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	t, err := h.taxpayers.Archive(ctx, userID, req.TaxpayerId.String())
	if err != nil {
		return nil, toAppError(err)
	}
	body, err := toTaxpayer(t)
	if err != nil {
		return nil, err
	}
	return gen.ArchiveTaxpayer200JSONResponse{Body: body, Headers: gen.ArchiveTaxpayer200ResponseHeaders{CacheControl: noStore}}, nil
}

// UnarchiveTaxpayer implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/unarchive).
func (h *Handler) UnarchiveTaxpayer(ctx context.Context, req gen.UnarchiveTaxpayerRequestObject) (gen.UnarchiveTaxpayerResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UnarchiveTaxpayer401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	t, err := h.taxpayers.Unarchive(ctx, userID, req.TaxpayerId.String())
	if err != nil {
		return nil, toAppError(err)
	}
	body, err := toTaxpayer(t)
	if err != nil {
		return nil, err
	}
	return gen.UnarchiveTaxpayer200JSONResponse{Body: body, Headers: gen.UnarchiveTaxpayer200ResponseHeaders{CacheControl: noStore}}, nil
}

// AddTaxpayerRegime implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/regimes).
func (h *Handler) AddTaxpayerRegime(ctx context.Context, req gen.AddTaxpayerRegimeRequestObject) (gen.AddTaxpayerRegimeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.AddTaxpayerRegime401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	in, err := toRegimeInput(req.Body, optionsField)
	if err != nil {
		return nil, err
	}
	regime, err := h.regimes.Add(ctx, userID, req.TaxpayerId.String(), in)
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.AddTaxpayerRegime201JSONResponse{Body: toTaxpayerRegime(regime), Headers: gen.AddTaxpayerRegime201ResponseHeaders{CacheControl: noStore}}, nil
}

// UpdateTaxpayerRegime implements gen.StrictServerInterface
// (PATCH /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}): merge semantics.
func (h *Handler) UpdateTaxpayerRegime(ctx context.Context, req gen.UpdateTaxpayerRegimeRequestObject) (gen.UpdateTaxpayerRegimeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UpdateTaxpayerRegime401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	var patch map[string]domain.OptionValue
	if req.Body != nil {
		var err error
		if patch, err = fromOptions(&req.Body.Options, optionsField); err != nil {
			return nil, err
		}
	}
	regime, err := h.regimes.UpdateOptions(ctx, userID, req.TaxpayerId.String(), req.RegimeCode, patch)
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.UpdateTaxpayerRegime200JSONResponse{Body: toTaxpayerRegime(regime), Headers: gen.UpdateTaxpayerRegime200ResponseHeaders{CacheControl: noStore}}, nil
}

// ArchiveTaxpayerRegime implements gen.StrictServerInterface
// (DELETE /v1/taxpayers/{taxpayerId}/regimes/{regimeCode}?end_period=…):
// ends the regime; its records are kept.
func (h *Handler) ArchiveTaxpayerRegime(ctx context.Context, req gen.ArchiveTaxpayerRegimeRequestObject) (gen.ArchiveTaxpayerRegimeResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ArchiveTaxpayerRegime401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	regime, err := h.regimes.End(ctx, userID, req.TaxpayerId.String(), usecase.EndRegimeInput{
		Code: req.RegimeCode, End: req.Params.EndPeriod, Confirm: confirmed(req.Params.Confirm),
	})
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.ArchiveTaxpayerRegime200JSONResponse{Body: toTaxpayerRegime(regime), Headers: gen.ArchiveTaxpayerRegime200ResponseHeaders{CacheControl: noStore}}, nil
}

// caller returns the authenticated user's id.
func caller(ctx context.Context) (string, bool) {
	userID, err := principal.UserID(ctx)
	return userID, err == nil
}

func confirmed(confirm *bool) bool { return confirm != nil && *confirm }

// unauthorized is the contract's Unauthorized response: the generic 401
// with the bearer challenge.
func unauthorized(ctx context.Context) gen.UnauthorizedJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeUnauthorized
	body.Error.Message = msgAuthenticationRequired
	body.Error.RequestId = requestid.From(ctx)
	return gen.UnauthorizedJSONResponse{Body: body, Headers: gen.UnauthorizedResponseHeaders{WWWAuthenticate: bearerChallenge}}
}
