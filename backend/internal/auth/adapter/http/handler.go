package authhttp

import (
	"context"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clientnet"
)

// Registerer starts a sign-up (usecase.Register).
type Registerer interface {
	Execute(ctx context.Context, in usecase.RegisterInput) error
}

// EmailVerifier confirms a code, sets the password and signs the new account
// in (usecase.VerifyEmail).
type EmailVerifier interface {
	Execute(ctx context.Context, in usecase.VerifyInput) (usecase.Session, error)
}

// CodeResender resends a verification code (usecase.ResendVerification).
type CodeResender interface {
	Execute(ctx context.Context, email string) error
}

// SignIn checks email and password and starts a session (usecase.Login).
type SignIn interface {
	Execute(ctx context.Context, in usecase.LoginInput) (usecase.Session, error)
}

// UseCases are the auth use cases the handler serves. All are required.
type UseCases struct {
	Register Registerer
	Verify   EmailVerifier
	Resend   CodeResender
	Login    SignIn
}

// Handler implements the auth operations of gen.StrictServerInterface built
// so far: register, verifyEmail and resendVerification (M1-05) and login
// (M1-07). cmd/api embeds it in the API server; the other auth operations
// stay not built (404).
type Handler struct {
	register Registerer
	verify   EmailVerifier
	resend   CodeResender
	login    SignIn
}

// NewHandler returns the handler over the use cases.
func NewHandler(uc UseCases) *Handler {
	return &Handler{register: uc.Register, verify: uc.Verify, resend: uc.Resend, login: uc.Login}
}

// Login implements gen.StrictServerInterface (POST /v1/auth/login). Every
// credential failure is the one generic 401 (US-A2a).
func (h *Handler) Login(ctx context.Context, req gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	session, err := h.login.Execute(ctx, usecase.LoginInput{
		Email: string(req.Body.Email), Password: req.Body.Password, Client: clientnet.From(ctx),
	})
	if limited, ok := tooManyRequests(ctx, err); ok {
		return gen.Login429JSONResponse{TooManyRequestsJSONResponse: limited}, nil
	}
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.Login200JSONResponse{
		Body: gen.AuthSession{User: toUser(session.Account), Tokens: toTokenPair(session.Tokens)},
	}, nil
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
		AcceptedDocuments: documents,
	})
	if limited, ok := tooManyRequests(ctx, err); ok {
		return gen.Register429JSONResponse{TooManyRequestsJSONResponse: limited}, nil
	}
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.Register202JSONResponse{Body: verificationSent}, nil
}

// ResendVerification implements gen.StrictServerInterface
// (POST /v1/auth/verify-email/resend): always the generic 202.
func (h *Handler) ResendVerification(ctx context.Context, req gen.ResendVerificationRequestObject) (gen.ResendVerificationResponseObject, error) {
	if err := h.resend.Execute(ctx, string(req.Body.Email)); err != nil {
		return nil, toAppError(err)
	}
	return gen.ResendVerification202JSONResponse{Body: verificationSent}, nil
}

// VerifyEmail implements gen.StrictServerInterface (POST /v1/auth/verify-email):
// the password and the optional name are set here, by whoever holds the code (D8).
func (h *Handler) VerifyEmail(ctx context.Context, req gen.VerifyEmailRequestObject) (gen.VerifyEmailResponseObject, error) {
	body := req.Body
	session, err := h.verify.Execute(ctx, usecase.VerifyInput{
		Email:    string(body.Email),
		Code:     body.Code,
		Password: body.Password,
		Name:     body.Name,
	})
	if err != nil {
		return nil, toAppError(err)
	}
	return gen.VerifyEmail201JSONResponse{
		Body: gen.AuthSession{User: toUser(session.Account), Tokens: toTokenPair(session.Tokens)},
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
