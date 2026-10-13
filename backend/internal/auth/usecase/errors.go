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
	// ErrAccountNotFound reports a user row that does not exist (or that RLS
	// hides from the unit of work's user).
	ErrAccountNotFound = errors.New("auth: account not found")
	// ErrInvalidCredentials is the one answer for a wrong password, an
	// unknown email, a pending (unverified) sign-up or an account that may
	// not sign in (US-A2a): the caller cannot tell which.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	// ErrInvalidAccessToken reports a bearer access token that is malformed,
	// forged, expired or not meant for this API. Verifiers wrap the reason
	// in it for the logs; the client only ever sees the generic 401.
	ErrInvalidAccessToken = errors.New("auth: invalid access token")
)
