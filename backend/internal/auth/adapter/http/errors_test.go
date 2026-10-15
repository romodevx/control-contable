package authhttp_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

func TestErrorMapperMapsEveryDomainError(t *testing.T) {
	tests := []struct {
		err  error
		want apperr.Code
	}{
		{err: domain.ErrInvalidEmail, want: apperr.CodeValidationFailed},
		{err: domain.ErrPasswordTooShort, want: apperr.CodeValidationFailed},
		{err: domain.ErrPasswordTooLong, want: apperr.CodeValidationFailed},
		{err: domain.ErrPasswordMalformed, want: apperr.CodeValidationFailed},
		{err: domain.ErrPasswordCommon, want: apperr.CodeValidationFailed},
		{err: domain.ErrInvalidVerificationCode, want: apperr.CodeValidationFailed},
		{err: domain.ErrInvalidName, want: apperr.CodeValidationFailed},
		{err: domain.ErrNameInvisible, want: apperr.CodeValidationFailed},
		{err: domain.ErrDocumentMissing, want: apperr.CodeValidationFailed},
		{err: domain.ErrDocumentDuplicate, want: apperr.CodeValidationFailed},
		{err: domain.ErrDocumentUnknown, want: apperr.CodeValidationFailed},
		{err: domain.ErrDocumentNotCurrent, want: apperr.CodeValidationFailed},
		{err: usecase.ErrVerificationRejected, want: apperr.CodeUnauthorized},
		{err: domain.ErrMalformedRefreshToken, want: apperr.CodeUnauthorized},
		{err: usecase.ErrInvalidCredentials, want: apperr.CodeUnauthorized},
		{err: usecase.ErrInvalidAccessToken, want: apperr.CodeUnauthorized},
		{err: usecase.ErrRefreshRejected, want: apperr.CodeUnauthorized},
		{err: usecase.ErrTokenReused, want: apperr.CodeUnauthorized},
		{err: usecase.ErrTokenExpired, want: apperr.CodeUnauthorized},
		{err: usecase.ErrAccountInactive, want: apperr.CodeUnauthorized},
		{err: domain.ErrInvalidUserID, want: apperr.CodeInternalError},
		{err: domain.ErrMalformedTokenHash, want: apperr.CodeInternalError},
		{err: domain.ErrRandomSource, want: apperr.CodeInternalError},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			wrapped := fmt.Errorf("use case: %w", tt.err)
			resolved := authhttp.ErrorMapper.Resolve(wrapped)
			require.NotNil(t, resolved)
			assert.Equal(t, tt.want, resolved.Code)
			assert.True(t, resolved.Code.Valid())
			assert.NotEmpty(t, resolved.Message)
			assert.ErrorIs(t, resolved, tt.err)
		})
	}
}

// TestErrorMapperMessagesAreClientSafe: messages reach the client, so they
// must not echo the domain error text (which may name internals).
func TestErrorMapperMessagesAreClientSafe(t *testing.T) {
	for _, err := range []error{domain.ErrPasswordCommon, domain.ErrMalformedRefreshToken} {
		resolved := authhttp.ErrorMapper.Resolve(err)
		assert.NotContains(t, resolved.Message, "auth:")
	}
}

// US-A2b: every refresh refusal is one answer, the contract's example, so
// the client cannot tell reuse from expiry from a suspended account.
func TestErrorMapper_everyRefreshRefusalIsTheSameMessage(t *testing.T) {
	for _, err := range []error{
		usecase.ErrTokenMalformed, usecase.ErrTokenUnknown, usecase.ErrTokenExpired,
		usecase.ErrTokenRevoked, usecase.ErrTokenReused, usecase.ErrAccountInactive,
	} {
		resolved := authhttp.ErrorMapper.Resolve(err)
		assert.Equal(t, apperr.CodeUnauthorized, resolved.Code, err)
		assert.Equal(t, "Your session has ended. Sign in again.", resolved.Message, err)
	}
}

func TestErrorMapperLeavesUnknownErrorsInternal(t *testing.T) {
	resolved := authhttp.ErrorMapper.Resolve(errors.New("db down"))
	assert.Equal(t, apperr.CodeInternalError, resolved.Code)
	assert.Equal(t, apperr.InternalMessage, resolved.Message)
}
