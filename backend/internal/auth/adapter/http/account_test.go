package authhttp_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const challenge = `Bearer realm="control-contable"`

type fakeProfile struct {
	gotID   domain.UserID
	gotName string
	account usecase.Account
	err     error
}

func (f *fakeProfile) Get(_ context.Context, id domain.UserID) (usecase.Account, error) {
	f.gotID = id
	return f.account, f.err
}

func (f *fakeProfile) Rename(_ context.Context, id domain.UserID, name string) (usecase.Account, error) {
	f.gotID, f.gotName = id, name
	return f.account, f.err
}

type fakeDeleter struct {
	gotID       domain.UserID
	gotPassword string
	calls       int
	err         error
}

func (f *fakeDeleter) Execute(_ context.Context, id domain.UserID, password string) error {
	f.calls++
	f.gotID, f.gotPassword = id, password
	return f.err
}

type fakeLegalList struct {
	docs []usecase.LegalDocument
	err  error
}

func (f fakeLegalList) Execute(context.Context) ([]usecase.LegalDocument, error) {
	return f.docs, f.err
}

type accountFixture struct {
	profile *fakeProfile
	deleter *fakeDeleter
	legal   *fakeLegalList
	handler *authhttp.AccountHandler
}

func newAccountFixture(t *testing.T) accountFixture {
	t.Helper()
	id, err := domain.ParseUserID(userIDText)
	require.NoError(t, err)
	email, err := domain.NewEmail("ana.gomez@example.mx")
	require.NoError(t, err)
	name, err := domain.NewName("Ana Gómez")
	require.NoError(t, err)
	f := accountFixture{
		profile: &fakeProfile{account: usecase.Account{
			ID: id, Email: email, Name: name,
			CreatedAt: time.Date(2026, 10, 13, 12, 30, 0, 0, time.FixedZone("CST", -6*3600)),
		}},
		deleter: &fakeDeleter{},
		legal:   &fakeLegalList{},
	}
	f.handler = authhttp.NewAccountHandler(authhttp.AccountUseCases{Profile: f.profile, Delete: f.deleter, Legal: f.legal})
	return f
}

// signedIn is a request context as the authenticator leaves it.
func signedIn(t *testing.T) context.Context {
	t.Helper()
	ctx := principal.With(t.Context(), principal.Principal{UserID: userIDText, TokenID: "jti"})
	return requestid.With(ctx, "req_test")
}

func TestGetMe_answersTheCallersProfileNoStore(t *testing.T) {
	f := newAccountFixture(t)

	resp, err := f.handler.GetMe(signedIn(t), gen.GetMeRequestObject{})

	require.NoError(t, err)
	got, ok := resp.(gen.GetMe200JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, "no-store", got.Headers.CacheControl)
	assert.Equal(t, userIDText, got.Body.Id.String())
	assert.Equal(t, "ana.gomez@example.mx", string(got.Body.Email))
	require.NotNil(t, got.Body.Name)
	assert.Equal(t, "Ana Gómez", *got.Body.Name)
	assert.Equal(t, time.UTC, got.Body.CreatedAt.Location())
	assert.Equal(t, userIDText, f.profile.gotID.String(), "the principal's user, nothing from the request")
}

func TestGetMe_aNamelessAccountHasANullName(t *testing.T) {
	f := newAccountFixture(t)
	f.profile.account.Name = domain.Name{}

	resp, err := f.handler.GetMe(signedIn(t), gen.GetMeRequestObject{})

	require.NoError(t, err)
	assert.Nil(t, resp.(gen.GetMe200JSONResponse).Body.Name)
}

// A deleted account's access token lives on for up to 15 minutes: every call
// it makes is the generic 401 with the bearer challenge.
func TestGetMe_aDeletedAccountIsTheGeneric401(t *testing.T) {
	f := newAccountFixture(t)
	f.profile.err = fmt.Errorf("read: %w", usecase.ErrAccountNotFound)

	resp, err := f.handler.GetMe(signedIn(t), gen.GetMeRequestObject{})

	require.NoError(t, err)
	got, ok := resp.(gen.GetMe401JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, challenge, got.Headers.WWWAuthenticate)
	assert.Equal(t, gen.ErrorCodeUnauthorized, got.Body.Error.Code)
	assert.Equal(t, "Authentication required.", got.Body.Error.Message)
	assert.Equal(t, "req_test", got.Body.Error.RequestId)
}

func TestGetMe_noPrincipalIsTheGeneric401(t *testing.T) {
	f := newAccountFixture(t)

	resp, err := f.handler.GetMe(t.Context(), gen.GetMeRequestObject{})

	require.NoError(t, err)
	_, ok := resp.(gen.GetMe401JSONResponse)
	assert.True(t, ok, "%T", resp)
}

func TestGetMe_aFailureIsReturnedForThe500(t *testing.T) {
	f := newAccountFixture(t)
	f.profile.err = errBoom

	_, err := f.handler.GetMe(signedIn(t), gen.GetMeRequestObject{})

	assert.ErrorIs(t, err, errBoom)
}

func TestUpdateMe_renamesAndAnswersTheProfileNoStore(t *testing.T) {
	f := newAccountFixture(t)
	name := "Ana Gómez de la Torre"

	resp, err := f.handler.UpdateMe(signedIn(t), gen.UpdateMeRequestObject{Body: &gen.UpdateMeJSONRequestBody{Name: &name}})

	require.NoError(t, err)
	got, ok := resp.(gen.UpdateMe200JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, "no-store", got.Headers.CacheControl)
	assert.Equal(t, name, f.profile.gotName)
	assert.Equal(t, userIDText, f.profile.gotID.String())
}

func TestUpdateMe_anInvalidNameNamesTheField(t *testing.T) {
	f := newAccountFixture(t)
	f.profile.err = domain.ErrInvalidName
	blank := "   "

	_, err := f.handler.UpdateMe(signedIn(t), gen.UpdateMeRequestObject{Body: &gen.UpdateMeJSONRequestBody{Name: &blank}})

	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, "name", appErr.Fields[0].Field)
}

func TestUpdateMe_aDeletedAccountIsTheGeneric401(t *testing.T) {
	f := newAccountFixture(t)
	f.profile.err = usecase.ErrAccountNotFound
	name := "Ana"

	resp, err := f.handler.UpdateMe(signedIn(t), gen.UpdateMeRequestObject{Body: &gen.UpdateMeJSONRequestBody{Name: &name}})

	require.NoError(t, err)
	got, ok := resp.(gen.UpdateMe401JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, challenge, got.Headers.WWWAuthenticate)
}

func TestDeleteMe_answers204(t *testing.T) {
	f := newAccountFixture(t)

	resp, err := f.handler.DeleteMe(signedIn(t), gen.DeleteMeRequestObject{Body: &gen.DeleteMeJSONRequestBody{Password: "made-up sample phrase"}})

	require.NoError(t, err)
	assert.IsType(t, gen.DeleteMe204Response{}, resp)
	assert.Equal(t, userIDText, f.deleter.gotID.String())
	assert.Equal(t, "made-up sample phrase", f.deleter.gotPassword)
}

func TestDeleteMe_aWrongPasswordIsA403InvalidPassword(t *testing.T) {
	f := newAccountFixture(t)
	f.deleter.err = fmt.Errorf("delete: %w", usecase.ErrWrongPassword)

	_, err := f.handler.DeleteMe(signedIn(t), gen.DeleteMeRequestObject{Body: &gen.DeleteMeJSONRequestBody{Password: "x"}})

	// Not a 401: the bearer was valid, and a client refreshes on a 401.
	var appErr *apperr.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, apperr.CodeInvalidPassword, appErr.Code)
	assert.Equal(t, "The password is incorrect.", appErr.Message)
}

func TestDeleteMe_bearerFailuresAreTheGeneric401WithTheChallenge(t *testing.T) {
	t.Run("a deleted account", func(t *testing.T) {
		f := newAccountFixture(t)
		f.deleter.err = fmt.Errorf("delete: %w", usecase.ErrAccountNotFound)

		resp, err := f.handler.DeleteMe(signedIn(t), gen.DeleteMeRequestObject{Body: &gen.DeleteMeJSONRequestBody{Password: "x"}})

		require.NoError(t, err)
		got, ok := resp.(gen.DeleteMe401JSONResponse)
		require.True(t, ok, "%T", resp)
		assert.Equal(t, challenge, got.Headers.WWWAuthenticate)
		assert.Equal(t, gen.ErrorCodeUnauthorized, got.Body.Error.Code)
		assert.Equal(t, "Authentication required.", got.Body.Error.Message)
	})

	t.Run("no principal deletes nothing", func(t *testing.T) {
		f := newAccountFixture(t)

		resp, err := f.handler.DeleteMe(t.Context(), gen.DeleteMeRequestObject{Body: &gen.DeleteMeJSONRequestBody{Password: "x"}})

		require.NoError(t, err)
		got, ok := resp.(gen.DeleteMe401JSONResponse)
		require.True(t, ok, "%T", resp)
		assert.Equal(t, challenge, got.Headers.WWWAuthenticate)
		assert.Zero(t, f.deleter.calls)
	})
}

func TestDeleteMe_outOfAttemptsIsA429WithRetryAfter(t *testing.T) {
	f := newAccountFixture(t)
	f.deleter.err = fmt.Errorf("delete: %w", &usecase.TooManyAttemptsError{RetryAfter: 90*time.Second + time.Millisecond})

	resp, err := f.handler.DeleteMe(signedIn(t), gen.DeleteMeRequestObject{Body: &gen.DeleteMeJSONRequestBody{Password: "x"}})

	require.NoError(t, err)
	got, ok := resp.(gen.DeleteMe429JSONResponse)
	require.True(t, ok, "%T", resp)
	require.NotNil(t, got.Headers.RetryAfter)
	assert.Equal(t, 91, *got.Headers.RetryAfter, "whole seconds, rounded up")
	assert.Equal(t, gen.ErrorCodeRateLimited, got.Body.Error.Code)
	assert.Equal(t, "req_test", got.Body.Error.RequestId)
}

func TestDeleteMe_anInfrastructureFailureIsReturnedForThe500(t *testing.T) {
	f := newAccountFixture(t)
	f.deleter.err = errBoom

	_, err := f.handler.DeleteMe(signedIn(t), gen.DeleteMeRequestObject{Body: &gen.DeleteMeJSONRequestBody{Password: "x"}})

	assert.ErrorIs(t, err, errBoom)
}

func TestListLegalDocuments_answersTheCurrentDocumentsInUTC(t *testing.T) {
	f := newAccountFixture(t)
	published := time.Date(2026, 10, 1, 0, 0, 0, 0, time.FixedZone("CST", -6*3600))
	f.legal.docs = []usecase.LegalDocument{
		{Kind: domain.DocumentPrivacyIntegral, Version: "v1", URL: "https://example.mx/aviso", PublishedAt: published},
		{Kind: domain.DocumentTerms, Version: "2026-10-01", URL: "https://example.mx/terminos", PublishedAt: published},
	}

	resp, err := f.handler.ListLegalDocuments(t.Context(), gen.ListLegalDocumentsRequestObject{})

	require.NoError(t, err)
	got, ok := resp.(gen.ListLegalDocuments200JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, "public, max-age=300", got.Headers.CacheControl, "public and the same for everyone")
	require.Len(t, got.Body.Items, 2)
	assert.Equal(t, gen.LegalDocument{
		Kind: gen.PrivacyNoticeIntegral, Version: "v1", Url: "https://example.mx/aviso", PublishedAt: published.UTC(),
	}, got.Body.Items[0])
	assert.Equal(t, gen.Terms, got.Body.Items[1].Kind)
	assert.Equal(t, time.UTC, got.Body.Items[1].PublishedAt.Location())
}

func TestListLegalDocuments_aFailureIsReturnedForThe500(t *testing.T) {
	f := newAccountFixture(t)
	f.legal.err = errBoom

	_, err := f.handler.ListLegalDocuments(t.Context(), gen.ListLegalDocumentsRequestObject{})

	assert.ErrorIs(t, err, errBoom)
}
