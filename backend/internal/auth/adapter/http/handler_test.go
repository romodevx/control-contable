package authhttp_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

const userIDText = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"

var errBoom = errors.New("database unavailable")

type fakeRegister struct {
	got usecase.RegisterInput
	err error
}

func (f *fakeRegister) Execute(_ context.Context, in usecase.RegisterInput) error {
	f.got = in
	return f.err
}

type fakeVerify struct {
	got     usecase.VerifyInput
	session usecase.Session
	err     error
}

func (f *fakeVerify) Execute(_ context.Context, in usecase.VerifyInput) (usecase.Session, error) {
	f.got = in
	return f.session, f.err
}

type fakeResend struct {
	email string
	err   error
}

func (f *fakeResend) Execute(_ context.Context, email string) error {
	f.email = email
	return f.err
}

type fakeLogin struct {
	got     usecase.LoginInput
	session usecase.Session
	err     error
}

func (f *fakeLogin) Execute(_ context.Context, in usecase.LoginInput) (usecase.Session, error) {
	f.got = in
	return f.session, f.err
}

func newHandler(r *fakeRegister, v *fakeVerify, s *fakeResend) *authhttp.Handler {
	return authhttp.NewHandler(authhttp.UseCases{Register: r, Verify: v, Resend: s, Login: &fakeLogin{}})
}

func loginHandler(l *fakeLogin) *authhttp.Handler {
	return authhttp.NewHandler(authhttp.UseCases{Register: &fakeRegister{}, Verify: &fakeVerify{}, Resend: &fakeResend{}, Login: l})
}

func loginBody() *gen.LoginJSONRequestBody {
	return &gen.LoginJSONRequestBody{Email: "Ana.Gomez@example.mx", Password: "made-up sample phrase"}
}

func TestLogin_answers200WithANoStoreSession(t *testing.T) {
	userID, err := domain.ParseUserID(userIDText)
	require.NoError(t, err)
	email, err := domain.NewEmail("ana.gomez@example.mx")
	require.NoError(t, err)
	refresh, err := domain.NewRefreshToken(userID, zeroes{})
	require.NoError(t, err)
	createdAt := time.Date(2026, 10, 13, 12, 30, 0, 0, time.FixedZone("CST", -6*3600))
	login := &fakeLogin{session: usecase.Session{
		Account: usecase.Account{ID: userID, Email: email, CreatedAt: createdAt},
		Tokens:  usecase.Tokens{Access: "jwt", AccessExpiresIn: 15 * time.Minute, Refresh: refresh, RefreshExpiresIn: 720 * time.Hour},
	}}

	resp, err := loginHandler(login).Login(t.Context(), gen.LoginRequestObject{Body: loginBody()})

	require.NoError(t, err)
	assert.Equal(t, usecase.LoginInput{Email: "Ana.Gomez@example.mx", Password: "made-up sample phrase"}, login.got,
		"normalization is the use case's job")
	got, ok := resp.(gen.Login200JSONResponse)
	require.True(t, ok)
	assert.Equal(t, "no-store", got.Headers.CacheControl)
	assert.Equal(t, userIDText, got.Body.User.Id.String())
	assert.Equal(t, openapi_types.Email("ana.gomez@example.mx"), got.Body.User.Email)
	assert.Nil(t, got.Body.User.Name)
	assert.Equal(t, createdAt.UTC(), got.Body.User.CreatedAt)
	assert.Equal(t, gen.TokenPair{
		AccessToken: "jwt", AccessExpiresIn: 900,
		RefreshToken: refresh.Reveal(), RefreshExpiresIn: 2592000, TokenType: gen.Bearer,
	}, got.Body.Tokens)
}

func TestLogin_invalidCredentialsAreTheGeneric401(t *testing.T) {
	h := loginHandler(&fakeLogin{err: usecase.ErrInvalidCredentials})

	_, err := h.Login(t.Context(), gen.LoginRequestObject{Body: loginBody()})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeUnauthorized, appErr.Code)
	assert.Equal(t, "Email or password is incorrect.", appErr.Message)
	assert.Empty(t, appErr.Fields, "never which of the two was wrong")
}

func TestLogin_serverFaultsStayInternal(t *testing.T) {
	_, err := loginHandler(&fakeLogin{err: errBoom}).Login(t.Context(), gen.LoginRequestObject{Body: loginBody()})

	require.ErrorIs(t, err, errBoom)
	assert.Equal(t, apperr.CodeInternalError, apperr.Mapper{}.Resolve(err).Code)
}

func registerBody() *gen.RegisterJSONRequestBody {
	return &gen.RegisterJSONRequestBody{
		Email: openapi_types.Email("ana.gomez@example.mx"),
		AcceptedDocuments: []gen.DocumentAcceptance{
			{Kind: gen.PrivacyNoticeIntegral, Version: "v1"},
			{Kind: gen.Terms, Version: "v1"},
		},
	}
}

func TestRegister_answersTheGeneric202(t *testing.T) {
	reg := &fakeRegister{}

	resp, err := newHandler(reg, &fakeVerify{}, &fakeResend{}).Register(t.Context(), gen.RegisterRequestObject{Body: registerBody()})

	require.NoError(t, err)
	assert.Equal(t, gen.Register202JSONResponse{Status: gen.VerificationSentStatusVerificationSent}, resp)
	assert.Equal(t, "ana.gomez@example.mx", reg.got.Email)
	assert.Equal(t, []domain.DocumentVersion{
		{Kind: domain.DocumentPrivacyIntegral, Version: "v1"},
		{Kind: domain.DocumentTerms, Version: "v1"},
	}, reg.got.AcceptedDocuments)
}

func TestRegister_validationErrorsNameTheField(t *testing.T) {
	tests := []struct {
		err       error
		field     string
		fieldCode string
	}{
		{domain.ErrInvalidEmail, "email", "invalid_email"},
		{domain.ErrDocumentNotCurrent, "accepted_documents", "document_version_not_current"},
		{domain.ErrDocumentMissing, "accepted_documents", "document_missing"},
		{domain.ErrDocumentDuplicate, "accepted_documents", "document_duplicate"},
		{domain.ErrDocumentUnknown, "accepted_documents", "document_unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			h := newHandler(&fakeRegister{err: fmt.Errorf("register: %w", tt.err)}, &fakeVerify{}, &fakeResend{})

			_, err := h.Register(t.Context(), gen.RegisterRequestObject{Body: registerBody()})

			appErr, ok := errors.AsType[*apperr.Error](err)
			require.True(t, ok, "a typed error the router renders as is")
			assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
			assert.Equal(t, "One or more fields are invalid.", appErr.Message)
			require.Len(t, appErr.Fields, 1)
			assert.Equal(t, tt.field, appErr.Fields[0].Field)
			assert.Equal(t, tt.fieldCode, appErr.Fields[0].Code)
			assert.NotEmpty(t, appErr.Fields[0].Message)
			assert.NotContains(t, appErr.Fields[0].Message, "auth:", "client-safe text only")
			assert.ErrorIs(t, err, tt.err, "the cause stays wrapped for logs")
		})
	}
}

func TestRegister_serverFaultsStayInternal(t *testing.T) {
	h := newHandler(&fakeRegister{err: errBoom}, &fakeVerify{}, &fakeResend{})

	_, err := h.Register(t.Context(), gen.RegisterRequestObject{Body: registerBody()})

	require.ErrorIs(t, err, errBoom)
	assert.Equal(t, apperr.CodeInternalError, apperr.Mapper{}.Resolve(err).Code)
}

func TestResendVerification(t *testing.T) {
	resend := &fakeResend{}
	h := newHandler(&fakeRegister{}, &fakeVerify{}, resend)

	resp, err := h.ResendVerification(t.Context(), gen.ResendVerificationRequestObject{
		Body: &gen.ResendVerificationJSONRequestBody{Email: "ana.gomez@example.mx"},
	})

	require.NoError(t, err)
	assert.Equal(t, gen.ResendVerification202JSONResponse{Status: gen.VerificationSentStatusVerificationSent}, resp)
	assert.Equal(t, "ana.gomez@example.mx", resend.email)
}

func TestResendVerification_invalidEmailNamesTheField(t *testing.T) {
	h := newHandler(&fakeRegister{}, &fakeVerify{}, &fakeResend{err: domain.ErrInvalidEmail})

	_, err := h.ResendVerification(t.Context(), gen.ResendVerificationRequestObject{
		Body: &gen.ResendVerificationJSONRequestBody{Email: "x@y"},
	})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok)
	assert.Equal(t, "email", appErr.Fields[0].Field)
}

func TestVerifyEmail_answers201WithTheSession(t *testing.T) {
	userID, err := domain.ParseUserID(userIDText)
	require.NoError(t, err)
	email, err := domain.NewEmail("ana.gomez@example.mx")
	require.NoError(t, err)
	name, err := domain.NewName("Ana Gómez")
	require.NoError(t, err)
	refresh, err := domain.NewRefreshToken(userID, zeroes{})
	require.NoError(t, err)
	createdAt := time.Date(2026, 10, 13, 12, 30, 0, 0, time.FixedZone("CST", -6*3600))
	verify := &fakeVerify{session: usecase.Session{
		Account: usecase.Account{ID: userID, Email: email, Name: name, CreatedAt: createdAt},
		Tokens:  usecase.Tokens{Access: "jwt", AccessExpiresIn: 15 * time.Minute, Refresh: refresh, RefreshExpiresIn: 720 * time.Hour},
	}}

	resp, err := newHandler(&fakeRegister{}, verify, &fakeResend{}).VerifyEmail(t.Context(), gen.VerifyEmailRequestObject{
		Body: verifyBody(),
	})

	require.NoError(t, err)
	require.NotNil(t, verify.got.Name)
	assert.Equal(t, usecase.VerifyInput{
		Email: "ana.gomez@example.mx", Code: "482913", Password: "made-up sample phrase", Name: verify.got.Name,
	}, verify.got, "the password travels with the code (D8)")
	assert.Equal(t, "Ana Gómez", *verify.got.Name)
	got, ok := resp.(gen.VerifyEmail201JSONResponse)
	require.True(t, ok)
	assert.Equal(t, "no-store", got.Headers.CacheControl)
	assert.Equal(t, userIDText, got.Body.User.Id.String())
	assert.Equal(t, openapi_types.Email("ana.gomez@example.mx"), got.Body.User.Email)
	require.NotNil(t, got.Body.User.Name)
	assert.Equal(t, "Ana Gómez", *got.Body.User.Name)
	assert.Equal(t, createdAt.UTC(), got.Body.User.CreatedAt)
	assert.Equal(t, gen.TokenPair{
		AccessToken: "jwt", AccessExpiresIn: 900,
		RefreshToken: refresh.Reveal(), RefreshExpiresIn: 2592000, TokenType: gen.Bearer,
	}, got.Body.Tokens)
}

func TestVerifyEmail_withoutANameOmitsIt(t *testing.T) {
	userID, err := domain.ParseUserID(userIDText)
	require.NoError(t, err)
	verify := &fakeVerify{session: usecase.Session{Account: usecase.Account{ID: userID}}}

	body := verifyBody()
	body.Name = nil
	resp, err := newHandler(&fakeRegister{}, verify, &fakeResend{}).VerifyEmail(t.Context(), gen.VerifyEmailRequestObject{Body: body})

	require.NoError(t, err)
	assert.Nil(t, verify.got.Name)
	assert.Nil(t, resp.(gen.VerifyEmail201JSONResponse).Body.User.Name)
}

func TestVerifyEmail_rejectionIsTheGeneric401(t *testing.T) {
	h := newHandler(&fakeRegister{}, &fakeVerify{err: usecase.ErrVerificationRejected}, &fakeResend{})

	_, err := h.VerifyEmail(t.Context(), gen.VerifyEmailRequestObject{
		Body: verifyBody(),
	})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeUnauthorized, appErr.Code)
	assert.Equal(t, "The code is invalid or has expired.", appErr.Message)
	assert.Empty(t, appErr.Fields)
}

func TestVerifyEmail_serverFaultsStayInternal(t *testing.T) {
	h := newHandler(&fakeRegister{}, &fakeVerify{err: errBoom}, &fakeResend{})

	_, err := h.VerifyEmail(t.Context(), gen.VerifyEmailRequestObject{
		Body: verifyBody(),
	})

	assert.Equal(t, apperr.CodeInternalError, apperr.Mapper{}.Resolve(err).Code)
}

func verifyBody() *gen.VerifyEmailJSONRequestBody {
	name := "Ana Gómez"
	return &gen.VerifyEmailJSONRequestBody{
		Email: "ana.gomez@example.mx", Code: "482913", Password: "made-up sample phrase", Name: &name,
	}
}

// The password policy and the name are verifyEmail's validation errors now
// (D8): each names its field, never the value.
func TestVerifyEmail_validationErrorsNameTheField(t *testing.T) {
	tests := []struct {
		err       error
		field     string
		fieldCode string
	}{
		{domain.ErrPasswordCommon, "password", "password_policy"},
		{domain.ErrPasswordTooShort, "password", "password_policy"},
		{domain.ErrPasswordTooLong, "password", "password_policy"},
		{domain.ErrPasswordMalformed, "password", "password_policy"},
		{domain.ErrInvalidName, "name", "invalid_name"},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			h := newHandler(&fakeRegister{}, &fakeVerify{err: fmt.Errorf("verify: %w", tt.err)}, &fakeResend{})

			_, err := h.VerifyEmail(t.Context(), gen.VerifyEmailRequestObject{Body: verifyBody()})

			appErr, ok := errors.AsType[*apperr.Error](err)
			require.True(t, ok, "a typed error the router renders as is")
			assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
			require.Len(t, appErr.Fields, 1)
			assert.Equal(t, tt.field, appErr.Fields[0].Field)
			assert.Equal(t, tt.fieldCode, appErr.Fields[0].Code)
			assert.NotContains(t, appErr.Fields[0].Message, "made-up sample phrase")
		})
	}
}

// zeroes is a deterministic entropy source.
type zeroes struct{}

func (zeroes) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}
