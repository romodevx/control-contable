package usecase

import "errors"

var (
	// ErrVerificationRejected is the one answer for a wrong, expired,
	// consumed, locked or unknown code or email (US-A1b): the caller cannot
	// tell which.
	ErrVerificationRejected = errors.New("auth: verification rejected")
	// ErrAccountExists reports a user already holding the email hash
	// (Accounts.Create lost a race to another verification).
	ErrAccountExists = errors.New("auth: an account with this email already exists")
)
