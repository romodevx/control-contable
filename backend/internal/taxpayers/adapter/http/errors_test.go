package taxpayershttp_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	taxpayershttp "github.com/romodevx/control-contable/backend/internal/taxpayers/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

func TestErrorMapperMapsEveryDomainError(t *testing.T) {
	tests := []struct {
		err  error
		want apperr.Code
	}{
		{err: domain.ErrInvalidPersonType, want: apperr.CodeValidationFailed},
		{err: domain.ErrRFCMalformed, want: apperr.CodeValidationFailed},
		{err: domain.ErrRFCInvalidDate, want: apperr.CodeValidationFailed},
		{err: domain.ErrRFCGeneric, want: apperr.CodeValidationFailed},
		{err: domain.ErrPersonTypeMismatch, want: apperr.CodeValidationFailed},
		{err: domain.ErrOptionUnknown, want: apperr.CodeValidationFailed},
		{err: domain.ErrOptionMissing, want: apperr.CodeValidationFailed},
		{err: domain.ErrOptionType, want: apperr.CodeValidationFailed},
		{err: domain.ErrOptionValue, want: apperr.CodeValidationFailed},
		{err: &domain.OptionError{Key: "deduction_mode", Err: domain.ErrOptionValue}, want: apperr.CodeValidationFailed},
		{err: domain.ErrInvalidOptionDef, want: apperr.CodeInternalError},
		{err: &domain.OptionError{Key: "deduction_mode", Err: domain.ErrInvalidOptionDef}, want: apperr.CodeInternalError},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			resolved := taxpayershttp.ErrorMapper.Resolve(fmt.Errorf("use case: %w", tt.err))
			require.NotNil(t, resolved)
			assert.Equal(t, tt.want, resolved.Code)
			assert.True(t, resolved.Code.Valid())
			assert.NotEmpty(t, resolved.Message)
			assert.ErrorIs(t, resolved, tt.err)
		})
	}
}

// TestErrorMapperMessagesAreClientSafe: messages reach the client, so they
// never echo the domain error text, an option key or an RFC.
func TestErrorMapperMessagesAreClientSafe(t *testing.T) {
	for _, err := range []error{
		domain.ErrRFCMalformed, domain.ErrRFCInvalidDate, domain.ErrRFCGeneric, domain.ErrPersonTypeMismatch,
		&domain.OptionError{Key: "deduction_mode", Err: domain.ErrOptionValue},
	} {
		resolved := taxpayershttp.ErrorMapper.Resolve(err)
		assert.NotContains(t, resolved.Message, "taxpayers:")
		assert.NotContains(t, resolved.Message, "deduction_mode")
	}
}
