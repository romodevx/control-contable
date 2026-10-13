package domain

// Error is a domain error of the auth module. Its values are constants, so
// callers match them with errors.Is and nothing can reassign them. The
// adapter maps each one to a contract ErrorCode (authhttp.ErrorMapper).
type Error string

func (e Error) Error() string { return string(e) }

// Domain errors. None of them ever carries the offending input.
const (
	// ErrInvalidEmail reports an address that is not a valid ASCII
	// dot-atom address within the RFC 5321 length limits.
	ErrInvalidEmail Error = "auth: invalid email"
	// ErrPasswordTooShort reports fewer than MinPasswordLength characters.
	ErrPasswordTooShort Error = "auth: password too short"
	// ErrPasswordTooLong reports more than MaxPasswordBytes bytes, before or
	// after normalization.
	ErrPasswordTooLong Error = "auth: password too long"
	// ErrPasswordMalformed reports a password that is not valid UTF-8.
	ErrPasswordMalformed Error = "auth: password is not valid UTF-8"
	// ErrPasswordCommon reports a commonly used or easily guessed password.
	ErrPasswordCommon Error = "auth: password is too common"
	// ErrInvalidUserID reports a user id that is not a canonical, non-nil UUID.
	ErrInvalidUserID Error = "auth: invalid user id"
	// ErrMalformedRefreshToken reports a refresh token that is not
	// "<uuid>.<secret>" with a canonical 256-bit base64url secret.
	ErrMalformedRefreshToken Error = "auth: malformed refresh token" //nolint:gosec // G101: an error message, not a credential
	// ErrMalformedTokenHash reports a stored token hash of the wrong length.
	ErrMalformedTokenHash Error = "auth: malformed token hash" //nolint:gosec // G101: an error message, not a credential
	// ErrInvalidVerificationCode reports a code that is not 6 ASCII digits.
	ErrInvalidVerificationCode Error = "auth: invalid verification code"
	// ErrRandomSource reports a failure of the injected entropy source.
	ErrRandomSource Error = "auth: random source failed"
	// ErrInvalidName reports a display name that is blank, too long, not
	// valid UTF-8 or holds control characters.
	ErrInvalidName Error = "auth: invalid name"
	// ErrDocumentMissing reports an acceptance without the terms or without
	// a privacy notice.
	ErrDocumentMissing Error = "auth: a required legal document was not accepted"
	// ErrDocumentDuplicate reports a legal document kind accepted twice.
	ErrDocumentDuplicate Error = "auth: a legal document kind was accepted twice"
	// ErrDocumentUnknown reports a legal document kind that does not exist.
	ErrDocumentUnknown Error = "auth: unknown legal document kind"
	// ErrDocumentNotCurrent reports an accepted version that is not the
	// current version of its document.
	ErrDocumentNotCurrent Error = "auth: accepted legal document version is not current"
)
