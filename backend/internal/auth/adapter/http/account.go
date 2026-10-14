package authhttp

import (
	"context"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const (
	// msgWrongPassword answers a wrong current password on account deletion.
	msgWrongPassword = "The password is incorrect."
	// legalDocumentsCacheControl lets shared caches keep the public legal
	// document list briefly: a new version takes effect within 5 minutes.
	legalDocumentsCacheControl = "public, max-age=300"
)

// ProfileEditor reads and renames the signed-in account (usecase.Profile).
type ProfileEditor interface {
	Get(ctx context.Context, userID domain.UserID) (usecase.Account, error)
	Rename(ctx context.Context, userID domain.UserID, name string) (usecase.Account, error)
}

// AccountDeleter deletes the signed-in account after re-checking its
// password (usecase.DeleteAccount).
type AccountDeleter interface {
	Execute(ctx context.Context, userID domain.UserID, password string) error
}

// LegalDocumentLister lists the current legal documents
// (usecase.ListLegalDocuments).
type LegalDocumentLister interface {
	Execute(ctx context.Context) ([]usecase.LegalDocument, error)
}

// AccountUseCases are the account use cases AccountHandler serves. All are
// required.
type AccountUseCases struct {
	Profile ProfileEditor
	Delete  AccountDeleter
	Legal   LegalDocumentLister
}

// AccountHandler implements the account operations of
// gen.StrictServerInterface (M1-09): getMe, updateMe, deleteMe and
// listLegalDocuments. cmd/api embeds it in the API server.
//
// The secured operations act only on the principal's own user: the id comes
// from the access token the authenticator verified, never from the request.
type AccountHandler struct {
	profile ProfileEditor
	remove  AccountDeleter
	legal   LegalDocumentLister
}

// NewAccountHandler returns the handler over the use cases.
func NewAccountHandler(uc AccountUseCases) *AccountHandler {
	return &AccountHandler{profile: uc.Profile, remove: uc.Delete, legal: uc.Legal}
}

// errNoPrincipal reports a secured operation reached without an
// authenticated user (the router's authenticator always sets one).
var errNoPrincipal = fmt.Errorf("%w: no principal in the request context", usecase.ErrInvalidAccessToken)

// GetMe implements gen.StrictServerInterface (GET /v1/me).
func (h *AccountHandler) GetMe(ctx context.Context, _ gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	userID, err := caller(ctx)
	if err != nil {
		return gen.GetMe401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	account, err := h.profile.Get(ctx, userID)
	if errors.Is(err, usecase.ErrAccountNotFound) {
		return gen.GetMe401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.GetMe200JSONResponse{Body: toUser(account), Headers: gen.GetMe200ResponseHeaders{CacheControl: noStore}}, nil
}

// UpdateMe implements gen.StrictServerInterface (PATCH /v1/me). The contract
// makes name required in practice (minProperties 1, the only property) and
// non-nullable: an explicit null or an empty body is refused by request
// validation before this runs (ADR-0003, "Null on the wire").
func (h *AccountHandler) UpdateMe(ctx context.Context, req gen.UpdateMeRequestObject) (gen.UpdateMeResponseObject, error) {
	userID, err := caller(ctx)
	if err != nil {
		return gen.UpdateMe401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil || req.Body.Name == nil {
		return nil, toAppError(domain.ErrInvalidName)
	}
	account, err := h.profile.Rename(ctx, userID, *req.Body.Name)
	if errors.Is(err, usecase.ErrAccountNotFound) {
		return gen.UpdateMe401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.UpdateMe200JSONResponse{Body: toUser(account), Headers: gen.UpdateMe200ResponseHeaders{CacheControl: noStore}}, nil
}

// DeleteMe implements gen.StrictServerInterface (DELETE /v1/me): 204 once the
// account and everything it owns are gone. A 401 always means the bearer
// (with the challenge, the client's cue to refresh); a wrong password is a
// 403 invalid_password, and an account out of attempts a 429.
func (h *AccountHandler) DeleteMe(ctx context.Context, req gen.DeleteMeRequestObject) (gen.DeleteMeResponseObject, error) {
	userID, err := caller(ctx)
	if err != nil {
		return gen.DeleteMe401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	var password string
	if req.Body != nil {
		password = req.Body.Password
	}
	err = h.remove.Execute(ctx, userID, password)
	if limited, ok := tooManyRequests(ctx, err); ok {
		return gen.DeleteMe429JSONResponse{TooManyRequestsJSONResponse: limited}, nil
	}
	switch {
	case err == nil:
		return gen.DeleteMe204Response{}, nil
	case errors.Is(err, usecase.ErrAccountNotFound):
		return gen.DeleteMe401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	case errors.Is(err, usecase.ErrWrongPassword):
		return nil, apperr.Wrap(err, apperr.CodeInvalidPassword, msgWrongPassword)
	default:
		return nil, toAppError(err)
	}
}

// ListLegalDocuments implements gen.StrictServerInterface
// (GET /v1/legal/documents): public, the current version of each kind.
func (h *AccountHandler) ListLegalDocuments(ctx context.Context, _ gen.ListLegalDocumentsRequestObject) (gen.ListLegalDocumentsResponseObject, error) {
	docs, err := h.legal.Execute(ctx)
	if err != nil {
		return nil, toAppError(err)
	}
	items := make([]gen.LegalDocument, 0, len(docs))
	for _, d := range docs {
		items = append(items, gen.LegalDocument{
			Kind: gen.LegalDocumentKind(d.Kind), Version: d.Version, Url: d.URL, PublishedAt: d.PublishedAt.UTC(),
		})
	}
	return gen.ListLegalDocuments200JSONResponse{
		Body:    gen.LegalDocumentList{Items: items},
		Headers: gen.ListLegalDocuments200ResponseHeaders{CacheControl: legalDocumentsCacheControl},
	}, nil
}

// caller is the authenticated user of the request.
func caller(ctx context.Context) (domain.UserID, error) {
	id, err := principal.UserID(ctx)
	if err != nil {
		return domain.UserID{}, errNoPrincipal
	}
	userID, err := domain.ParseUserID(id)
	if err != nil {
		return domain.UserID{}, fmt.Errorf("%w: principal user id: %w", usecase.ErrInvalidAccessToken, err)
	}
	return userID, nil
}

// unauthorized is the contract's Unauthorized response: the generic 401 and
// the bearer challenge, as the authenticator writes it.
func unauthorized(ctx context.Context) gen.UnauthorizedJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeUnauthorized
	body.Error.Message = msgAuthenticationRequired
	body.Error.RequestId = requestid.From(ctx)
	return gen.UnauthorizedJSONResponse{Body: body, Headers: gen.UnauthorizedResponseHeaders{WWWAuthenticate: bearerChallenge}}
}
