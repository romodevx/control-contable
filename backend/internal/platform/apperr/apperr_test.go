package apperr_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

var errRFCTaken = errors.New("taxpayer: rfc already registered")

func TestError_messageAndUnwrap(t *testing.T) {
	cause := errors.New("db: unique violation")
	tests := []struct {
		name      string
		err       *apperr.Error
		wantText  string
		wantCause error
	}{
		{
			name:     "without cause",
			err:      apperr.New(apperr.CodeNotFound, "Not found."),
			wantText: "not_found: Not found.",
		},
		{
			name:      "with cause",
			err:       apperr.Wrap(cause, apperr.CodeConflict, "Conflict."),
			wantText:  "conflict: Conflict.: db: unique violation",
			wantCause: cause,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantText, tt.err.Error())
			assert.Equal(t, tt.wantCause, errors.Unwrap(tt.err))
			if tt.wantCause != nil {
				assert.ErrorIs(t, tt.err, tt.wantCause)
			}
		})
	}
}

func TestError_WithFieldsDoesNotMutateReceiver(t *testing.T) {
	base := apperr.New(apperr.CodeValidationFailed, "One or more fields are invalid.")
	field := apperr.FieldError{Field: "amount_cents", Code: "minimum", Message: "Must be >= 0."}

	got := base.WithFields(field)

	assert.Equal(t, []apperr.FieldError{field}, got.Fields)
	assert.Empty(t, base.Fields)
	assert.Equal(t, base.Code, got.Code)
}

func TestMapper_Resolve(t *testing.T) {
	typed := apperr.New(apperr.CodeEntitlementRequired, "Your trial has ended.")
	mapper := apperr.NewMapper(
		apperr.Rule{Target: errRFCTaken, Code: apperr.CodeRFCAlreadyRegistered, Message: "This RFC is already registered."},
		apperr.Rule{Target: context.DeadlineExceeded, Code: apperr.CodeUpstreamUnavailable, Message: "Service temporarily unavailable."},
	)
	tests := []struct {
		name        string
		err         error
		wantCode    apperr.Code
		wantMessage string
	}{
		{
			name:        "typed error passes through",
			err:         typed,
			wantCode:    apperr.CodeEntitlementRequired,
			wantMessage: "Your trial has ended.",
		},
		{
			name:        "wrapped typed error is found",
			err:         fmt.Errorf("record income: %w", typed),
			wantCode:    apperr.CodeEntitlementRequired,
			wantMessage: "Your trial has ended.",
		},
		{
			name:        "domain sentinel matched by rule",
			err:         fmt.Errorf("create taxpayer: %w", errRFCTaken),
			wantCode:    apperr.CodeRFCAlreadyRegistered,
			wantMessage: "This RFC is already registered.",
		},
		{
			name:        "second rule matches",
			err:         fmt.Errorf("query: %w", context.DeadlineExceeded),
			wantCode:    apperr.CodeUpstreamUnavailable,
			wantMessage: "Service temporarily unavailable.",
		},
		{
			name:        "unknown error hides its text",
			err:         errors.New("pq: password authentication failed for user x@y.com"),
			wantCode:    apperr.CodeInternalError,
			wantMessage: apperr.InternalMessage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.Resolve(tt.err)

			require.NotNil(t, got)
			assert.Equal(t, tt.wantCode, got.Code)
			assert.Equal(t, tt.wantMessage, got.Message)
			assert.ErrorIs(t, got, tt.err, "the resolved error keeps the original for logging")
		})
	}
}

func TestMapper_wrappedTypedErrorKeepsFields(t *testing.T) {
	field := apperr.FieldError{Field: "rfc", Code: "pattern", Message: "Invalid RFC."}
	typed := apperr.New(apperr.CodeValidationFailed, "Invalid.").WithFields(field)

	got := apperr.NewMapper().Resolve(fmt.Errorf("create taxpayer: %w", typed))

	assert.Equal(t, []apperr.FieldError{field}, got.Fields)
	assert.Equal(t, "validation_failed: Invalid.: create taxpayer: validation_failed: Invalid.", got.Error())
}

func TestMapper_ResolveNil(t *testing.T) {
	assert.Nil(t, apperr.NewMapper().Resolve(nil))
}

func TestMapper_typedErrorWinsOverRules(t *testing.T) {
	typed := apperr.Wrap(errRFCTaken, apperr.CodeConflict, "Conflict.")
	mapper := apperr.NewMapper(apperr.Rule{Target: errRFCTaken, Code: apperr.CodeRFCAlreadyRegistered, Message: "x"})

	assert.Equal(t, apperr.CodeConflict, mapper.Resolve(typed).Code)
}

func TestCode_Valid(t *testing.T) {
	tests := []struct {
		code apperr.Code
		want bool
	}{
		{apperr.CodeValidationFailed, true},
		{apperr.CodeInternalError, true},
		{apperr.Code(""), false},
		{apperr.Code("forbidden"), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.code.Valid())
		})
	}
}

// TestCodes_matchContract keeps the Go codes and the ErrorCode enum in
// docs/api/openapi.yaml identical: the contract is the source of truth.
func TestCodes_matchContract(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "docs", "api", "openapi.yaml"))
	require.NoError(t, err)
	var doc struct {
		Components struct {
			Schemas struct {
				ErrorCode struct {
					Enum []string `yaml:"enum"`
				} `yaml:"ErrorCode"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &doc))
	want := doc.Components.Schemas.ErrorCode.Enum
	require.NotEmpty(t, want, "ErrorCode enum not found in the contract")

	got := make([]string, 0, len(apperr.Codes()))
	for _, c := range apperr.Codes() {
		got = append(got, string(c))
	}
	slices.Sort(got)
	slices.Sort(want)

	assert.Equal(t, want, got)
}
