// Package authhttp is the auth module's HTTP adapter: it translates between
// the contract (docs/api/openapi.yaml) and the auth domain.
package authhttp

import (
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// Client-safe messages. They never echo the input or name internals.
const (
	msgInvalidEmail        = "Enter a valid email address."
	msgPasswordTooShort    = "Password must be at least 12 characters."
	msgPasswordTooLong     = "Password is too long."
	msgPasswordMalformed   = "Password contains invalid characters."
	msgPasswordCommon      = "This password is too common. Choose a different one."
	msgInvalidCode         = "Enter the 6-digit code from the email."
	msgInvalidRefreshToken = "Authentication required."
)

// ErrorMapper maps auth domain errors to contract error codes. Errors with no
// rule (ErrInvalidUserID, ErrMalformedTokenHash, ErrRandomSource) are server
// faults and resolve to internal_error.
var ErrorMapper = apperr.NewMapper(
	apperr.Rule{Target: domain.ErrInvalidEmail, Code: apperr.CodeValidationFailed, Message: msgInvalidEmail},
	apperr.Rule{Target: domain.ErrPasswordTooShort, Code: apperr.CodeValidationFailed, Message: msgPasswordTooShort},
	apperr.Rule{Target: domain.ErrPasswordTooLong, Code: apperr.CodeValidationFailed, Message: msgPasswordTooLong},
	apperr.Rule{Target: domain.ErrPasswordMalformed, Code: apperr.CodeValidationFailed, Message: msgPasswordMalformed},
	apperr.Rule{Target: domain.ErrPasswordCommon, Code: apperr.CodeValidationFailed, Message: msgPasswordCommon},
	apperr.Rule{Target: domain.ErrInvalidVerificationCode, Code: apperr.CodeValidationFailed, Message: msgInvalidCode},
	apperr.Rule{Target: domain.ErrMalformedRefreshToken, Code: apperr.CodeUnauthorized, Message: msgInvalidRefreshToken},
)
