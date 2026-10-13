package authhttp

import (
	"context"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
)

// noStore is the Cache-Control of every response carrying tokens (contract
// NoStore header).
const noStore = "no-store"

// Registerer starts a sign-up (usecase.Register).
type Registerer interface {
	Execute(ctx context.Context, in usecase.RegisterInput) error
}

// EmailVerifier confirms a code and signs the new account in (usecase.VerifyEmail).
type EmailVerifier interface {
	Execute(ctx context.Context, email, code string) (usecase.Session, error)
}

// CodeResender resends a verification code (usecase.ResendVerification).
type CodeResender interface {
	Execute(ctx context.Context, email string) error
}

// Handler implements the auth operations of gen.StrictServerInterface that
// M1-05 builds: register, verifyEmail and resendVerification. cmd/api embeds
// it in the API server; the other auth operations stay not built (404).
type Handler struct {
	register Registerer
	verify   EmailVerifier
	resend   CodeResender
}

// NewHandler returns the handler over the use cases.
func NewHandler(register Registerer, verify EmailVerifier, resend CodeResender) *Handler {
	return &Handler{register: register, verify: verify, resend: resend}
}

var verificationSent = gen.VerificationSent{Status: gen.VerificationSentStatusVerificationSent}

// Register implements gen.StrictServerInterface (POST /v1/auth/register).
// Every outcome but a validation error is the same generic 202.
func (h *Handler) Register(ctx context.Context, req gen.RegisterRequestObject) (gen.RegisterResponseObject, error) {
	body := req.Body
	documents := make([]domain.DocumentVersion, 0, len(body.AcceptedDocuments))
	for _, d := range body.AcceptedDocuments {
		documents = append(documents, domain.DocumentVersion{Kind: domain.DocumentKind(d.Kind), Version: d.Version})
	}
	err := h.register.Execute(ctx, usecase.RegisterInput{
		Email:             string(body.Email),
		Password:          body.Password,
		Name:              body.Name,
		AcceptedDocuments: documents,
	})
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.Register202JSONResponse(verificationSent), nil
}

// ResendVerification implements gen.StrictServerInterface
// (POST /v1/auth/verify-email/resend): always the generic 202.
func (h *Handler) ResendVerification(ctx context.Context, req gen.ResendVerificationRequestObject) (gen.ResendVerificationResponseObject, error) {
	if err := h.resend.Execute(ctx, string(req.Body.Email)); err != nil {
		return nil, toAppError(err)
	}
	return gen.ResendVerification202JSONResponse(verificationSent), nil
}

// VerifyEmail implements gen.StrictServerInterface (POST /v1/auth/verify-email).
func (h *Handler) VerifyEmail(ctx context.Context, req gen.VerifyEmailRequestObject) (gen.VerifyEmailResponseObject, error) {
	session, err := h.verify.Execute(ctx, string(req.Body.Email), req.Body.Code)
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.VerifyEmail201JSONResponse{
		Body:    gen.AuthSession{User: toUser(session.Account), Tokens: toTokenPair(session.Tokens)},
		Headers: gen.VerifyEmail201ResponseHeaders{CacheControl: noStore},
	}, nil
}

func toUser(a usecase.Account) gen.User {
	user := gen.User{
		Id:        uuid.UUID(a.ID),
		Email:     openapi_types.Email(a.Email.String()),
		CreatedAt: a.CreatedAt.UTC(),
	}
	if !a.Name.IsZero() {
		name := a.Name.String()
		user.Name = &name
	}
	return user
}

func toTokenPair(t usecase.Tokens) gen.TokenPair {
	return gen.TokenPair{
		AccessToken:      t.Access,
		AccessExpiresIn:  seconds(t.AccessExpiresIn),
		RefreshToken:     t.Refresh.Reveal(),
		RefreshExpiresIn: seconds(t.RefreshExpiresIn),
		TokenType:        gen.Bearer,
	}
}

func seconds(d time.Duration) int { return int(d / time.Second) }
