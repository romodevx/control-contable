// Package authhttp is the auth module's HTTP adapter: it translates between
// the contract (docs/api/openapi.yaml) and the auth use cases.
package authhttp

import (
	"errors"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// Client-safe messages. They never echo the input or name internals.
const (
	msgInvalidInput        = "One or more fields are invalid." // the router's wording for every validation_failed
	msgInvalidEmail        = "Enter a valid email address."
	msgPasswordTooShort    = "Password must be at least 12 characters."
	msgPasswordTooLong     = "Password is too long."
	msgPasswordMalformed   = "Password contains invalid characters."
	msgPasswordCommon      = "This password is too common. Choose a different one."
	msgPasswordPolicy      = "Use at least 12 characters that are not a common or leaked password." //nolint:gosec // G101: a name or message, not a credential
	msgInvalidName         = "Enter a name of 1 to 120 characters."
	msgDocuments           = "Accept the current privacy notice and terms."
	msgInvalidCode         = "Enter the 6-digit code from the email."
	msgCodeRejected        = "The code is invalid or has expired."
	msgInvalidRefreshToken = "Authentication required."
	// msgAuthenticationRequired is the one answer to a missing or refused
	// bearer token, whatever the reason (the contract's Unauthorized example).
	msgAuthenticationRequired = "Authentication required."
	// msgInvalidCredentials never says which of email or password was wrong
	// (the contract's login 401 example).
	msgInvalidCredentials = "Email or password is incorrect."
)

// ErrorMapper maps auth domain and use case errors to contract error codes.
// Errors with no rule (ErrInvalidUserID, ErrMalformedTokenHash,
// ErrRandomSource, any infrastructure failure) are server faults and resolve
// to internal_error.
var ErrorMapper = apperr.NewMapper(
	apperr.Rule{Target: domain.ErrInvalidEmail, Code: apperr.CodeValidationFailed, Message: msgInvalidEmail},
	apperr.Rule{Target: domain.ErrPasswordTooShort, Code: apperr.CodeValidationFailed, Message: msgPasswordTooShort},
	apperr.Rule{Target: domain.ErrPasswordTooLong, Code: apperr.CodeValidationFailed, Message: msgPasswordTooLong},
	apperr.Rule{Target: domain.ErrPasswordMalformed, Code: apperr.CodeValidationFailed, Message: msgPasswordMalformed},
	apperr.Rule{Target: domain.ErrPasswordCommon, Code: apperr.CodeValidationFailed, Message: msgPasswordCommon},
	apperr.Rule{Target: domain.ErrInvalidName, Code: apperr.CodeValidationFailed, Message: msgInvalidName},
	apperr.Rule{Target: domain.ErrDocumentMissing, Code: apperr.CodeValidationFailed, Message: msgDocuments},
	apperr.Rule{Target: domain.ErrDocumentDuplicate, Code: apperr.CodeValidationFailed, Message: msgDocuments},
	apperr.Rule{Target: domain.ErrDocumentUnknown, Code: apperr.CodeValidationFailed, Message: msgDocuments},
	apperr.Rule{Target: domain.ErrDocumentNotCurrent, Code: apperr.CodeValidationFailed, Message: msgDocuments},
	apperr.Rule{Target: domain.ErrInvalidVerificationCode, Code: apperr.CodeValidationFailed, Message: msgInvalidCode},
	apperr.Rule{Target: usecase.ErrVerificationRejected, Code: apperr.CodeUnauthorized, Message: msgCodeRejected},
	apperr.Rule{Target: domain.ErrMalformedRefreshToken, Code: apperr.CodeUnauthorized, Message: msgInvalidRefreshToken},
	apperr.Rule{Target: usecase.ErrInvalidCredentials, Code: apperr.CodeUnauthorized, Message: msgInvalidCredentials},
	apperr.Rule{Target: usecase.ErrInvalidAccessToken, Code: apperr.CodeUnauthorized, Message: msgAuthenticationRequired},
)

// fieldRule names the request field (the router's dotted path, CONTRACT.md)
// and the FieldError code of a validation error.
type fieldRule struct {
	target  error
	field   string
	code    string
	message string
}

// fieldRules attach a FieldError to the validation errors of the request
// bodies, as the contract's examples show (staleDocument, weakPassword). The
// password rules share one code and message so the reason is not echoed.
var fieldRules = []fieldRule{
	{domain.ErrInvalidEmail, "email", "invalid_email", msgInvalidEmail},
	{domain.ErrPasswordTooShort, "password", "password_policy", msgPasswordPolicy},
	{domain.ErrPasswordTooLong, "password", "password_policy", msgPasswordPolicy},
	{domain.ErrPasswordMalformed, "password", "password_policy", msgPasswordPolicy},
	{domain.ErrPasswordCommon, "password", "password_policy", msgPasswordPolicy},
	{domain.ErrInvalidName, "name", "invalid_name", msgInvalidName},
	{domain.ErrDocumentNotCurrent, "accepted_documents", "document_version_not_current", msgDocuments},
	{domain.ErrDocumentMissing, "accepted_documents", "document_missing", msgDocuments},
	{domain.ErrDocumentDuplicate, "accepted_documents", "document_duplicate", msgDocuments},
	{domain.ErrDocumentUnknown, "accepted_documents", "document_unknown", msgDocuments},
}

// toAppError translates a use case error for the router: a field error
// becomes validation_failed naming its field; a mapped error keeps its code;
// anything else is returned as is, so the router logs it and answers the
// generic 500.
func toAppError(err error) error {
	for _, r := range fieldRules {
		if errors.Is(err, r.target) {
			return apperr.Wrap(err, apperr.CodeValidationFailed, msgInvalidInput).
				WithFields(apperr.FieldError{Field: r.field, Code: r.code, Message: r.message})
		}
	}
	if resolved := ErrorMapper.Resolve(err); resolved.Code != apperr.CodeInternalError {
		return resolved
	}
	return err
}
