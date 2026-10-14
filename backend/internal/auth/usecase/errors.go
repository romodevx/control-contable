package usecase

import (
	"errors"
	"fmt"
	"time"
)

// RateLimitedError reports a key out of budget (MODULE-auth §5): login
// failures of an email, sign-ups of an email, rotations of a token family,
// deletion attempts of an account. RetryAfter is when the next event is
// admitted. It is checked before any credential work, so it reveals nothing
// about the account behind the key.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("auth: rate limited, retry in %s", e.RetryAfter)
}

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
	// ErrWrongPassword reports a current password that does not match the
	// signed-in account's (account deletion re-authentication, D4), or that
	// changed while the deletion ran. Nothing was deleted.
	ErrWrongPassword = errors.New("auth: wrong current password")
	// ErrInvalidAccessToken reports a bearer access token that is malformed,
	// forged, expired or not meant for this API. Verifiers wrap the reason
	// in it for the logs; the client only ever sees the generic 401.
	ErrInvalidAccessToken = errors.New("auth: invalid access token")
	// ErrRefreshRejected is the one answer for a refresh token that cannot
	// be used (US-A2b): malformed, unknown, foreign, expired, revoked,
	// reused, or of an account that may not sign in. The reasons below wrap
	// it for tests and logs; the client only ever sees the generic 401.
	ErrRefreshRejected = errors.New("auth: refresh token rejected")
)

// Refresh refusal reasons. Each wraps ErrRefreshRejected.
var (
	// ErrTokenMalformed: not "<uuid>.<secret>" (domain.ParseRefreshToken).
	ErrTokenMalformed = fmt.Errorf("%w: malformed", ErrRefreshRejected)
	// ErrTokenUnknown: no token with this hash for the token's user.
	ErrTokenUnknown = fmt.Errorf("%w: unknown", ErrRefreshRejected)
	// ErrTokenExpired: past its sliding or absolute expiry, never rotated.
	ErrTokenExpired = fmt.Errorf("%w: expired", ErrRefreshRejected)
	// ErrTokenRevoked: its family was revoked (logout or an earlier reuse).
	ErrTokenRevoked = fmt.Errorf("%w: revoked", ErrRefreshRejected)
	// ErrTokenReused: an already rotated token was presented again; its
	// whole family is revoked.
	ErrTokenReused = fmt.Errorf("%w: reused", ErrRefreshRejected)
	// ErrAccountInactive: the account no longer exists or may not sign in.
	ErrAccountInactive = fmt.Errorf("%w: account may not sign in", ErrRefreshRejected)
)
