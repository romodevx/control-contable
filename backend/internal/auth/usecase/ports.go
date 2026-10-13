// Package usecase holds the auth module's application logic (MODULE-auth):
// register, verifyEmail, resendVerification, login, refreshSession and
// logout. It owns its ports,
// which the infra adapters implement and cmd/api wires.
//
// Every database access runs in a unit of work: register and resend through
// tx.SystemManager (no user exists yet; they reach only SECURITY DEFINER
// functions), verifyEmail through tx.Manager as the account it creates, and
// login through both: the credential lookup (a definer function) with no
// user context, then the session as the user the password proved. Refresh
// runs as the user the token's own prefix names (no RLS bypass) and logout
// as the bearer token's user.
package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
)

// Policy values (MODULE-auth §2, ADR-0001 O6).
const (
	// CodeTTL is how long a verification code (and its pending sign-up) lives.
	CodeTTL = 15 * time.Minute
	// RefreshSlidingWindow is a refresh token's lifetime from issue (or rotation).
	RefreshSlidingWindow = 720 * time.Hour
	// RefreshAbsoluteLifetime caps a token family, whatever its rotations.
	RefreshAbsoluteLifetime = 2160 * time.Hour
)

// EmailIndex computes the email_hash blind index (crypto.BlindIndex).
type EmailIndex interface {
	Hash(value string) ([]byte, error)
}

// PasswordHasher hashes a new password (crypto.PasswordHasher, argon2id).
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// PasswordVerifier checks a login password against a stored argon2id hash
// and keeps the stored hashes on the current policy (crypto.PasswordHasher).
type PasswordVerifier interface {
	PasswordHasher
	// Verify reports whether password matches encoded, in constant time.
	Verify(password, encoded string) (bool, error)
	// NeedsRehash reports whether encoded was made with parameters other
	// than the current policy's.
	NeedsRehash(encoded string) (bool, error)
}

// CodeHasher computes code_hash: a keyed HMAC of the email hash and the code,
// never a plain digest (a 6-digit space is brute-forced instantly).
type CodeHasher interface {
	HashCode(emailHash []byte, code domain.VerificationCode) []byte
}

// LegalDocuments reads the current version of each legal document kind.
type LegalDocuments interface {
	CurrentVersions(ctx context.Context) (map[domain.DocumentKind]string, error)
}

// PendingSignUp is what register keeps until the email is verified. It holds
// no password and no name: verifyEmail takes both from whoever presents the
// emailed code, so pre-registering someone else's address sets nothing of
// theirs (D8).
type PendingSignUp struct {
	Documents  []domain.DocumentVersion
	AcceptedAt time.Time
}

// CodeOutcome is what a start or resend did to the pending sign-up.
type CodeOutcome int

// The code outcomes (email_verification_start / _resend).
const (
	// CodeUnchanged: nothing was written and nothing must be sent (no pending
	// sign-up, an expired one on resend, or the 60 s cooldown).
	CodeUnchanged CodeOutcome = iota
	// CodeStarted: a new pending sign-up holds the new code.
	CodeStarted
	// CodeResent: a live pending sign-up got a new code; its payload is kept.
	CodeResent
)

// ConsumeOutcome is the result of checking a code.
type ConsumeOutcome int

// The consume outcomes (email_verification_consume).
const (
	ConsumeInvalid ConsumeOutcome = iota
	ConsumeVerified
	ConsumeExpired
	ConsumeLocked
)

func (o ConsumeOutcome) String() string {
	switch o {
	case ConsumeVerified:
		return "verified"
	case ConsumeExpired:
		return "expired"
	case ConsumeLocked:
		return "locked"
	default:
		return "invalid"
	}
}

// SignUpStarter is what register needs from the pending sign-up store.
type SignUpStarter interface {
	// AccountExists reports whether a verified account holds the email.
	AccountExists(ctx context.Context, emailHash []byte) (bool, error)
	// Start starts a sign-up, replaces an expired one, or (for a live one)
	// acts as a resend that keeps its payload (D7).
	Start(ctx context.Context, emailHash, codeHash []byte, pending PendingSignUp, expiresAt time.Time) (CodeOutcome, error)
}

// CodeResender is what resendVerification needs from the store.
type CodeResender interface {
	Resend(ctx context.Context, emailHash, codeHash []byte, expiresAt time.Time) (CodeOutcome, error)
}

// SignUpVerifier is what verifyEmail needs from the store.
type SignUpVerifier interface {
	// Consume checks a code; on ConsumeVerified it returns the payload and
	// the pending row is gone (restored if the unit of work rolls back).
	Consume(ctx context.Context, emailHash, codeHash []byte) (ConsumeOutcome, PendingSignUp, error)
	// ClaimTrial reports true exactly once per email hash, ever.
	ClaimTrial(ctx context.Context, emailHash []byte) (bool, error)
}

// NewAccount is a verified sign-up, ready to become a user.
type NewAccount struct {
	ID           domain.UserID
	Email        domain.Email
	EmailHash    []byte
	PasswordHash string
	Name         domain.Name
	Consents     []domain.DocumentVersion
	AcceptedAt   time.Time
	VerifiedAt   time.Time
}

// Account is a user as the API shows it.
type Account struct {
	ID        domain.UserID
	Email     domain.Email
	Name      domain.Name
	CreatedAt time.Time
}

// Accounts creates users: the users row, its trial-cap counters and its
// consents, in the caller's unit of work. A unique email_hash clash is
// ErrAccountExists.
type Accounts interface {
	Create(ctx context.Context, account NewAccount) (Account, error)
}

// Credentials is what login needs of an account (a row of
// auth_lookup_credentials), never its email or name.
type Credentials struct {
	UserID       domain.UserID
	PasswordHash string
	// Active is false for an account that may not sign in (suspended, or a
	// status this code does not know: fail closed).
	Active bool
}

// CredentialLookup finds the credentials of the account holding an email
// hash, across users, through the SECURITY DEFINER lookup.
type CredentialLookup interface {
	// Lookup returns found = false when no account holds emailHash.
	Lookup(ctx context.Context, emailHash []byte) (creds Credentials, found bool, err error)
}

// LoginAccounts is what login needs of the users table, in the unit of
// work of the user signing in.
type LoginAccounts interface {
	// Find reads the account; ErrAccountNotFound when it does not exist.
	Find(ctx context.Context, id domain.UserID) (Account, error)
	// ReplacePasswordHash stores newHash only while the stored hash is still
	// oldHash, so a concurrent password change is never overwritten. No
	// matching row is not an error.
	ReplacePasswordHash(ctx context.Context, id domain.UserID, oldHash, newHash string) error
}

// TrialStarter starts a user's trial entitlement at startsAt.
type TrialStarter interface {
	StartTrial(ctx context.Context, userID domain.UserID, startsAt time.Time) error
}

// RefreshTokens stores refresh tokens (only their hash).
type RefreshTokens interface {
	// StartFamily stores token as the first of a new family, expiring after
	// sliding and at most absolute after issue (database clock).
	StartFamily(ctx context.Context, token domain.RefreshToken, sliding, absolute time.Duration) error
}

// RevokeReason is why a token family was revoked (refresh_tokens.revoke_reason).
type RevokeReason string

// The revocation reasons the auth use cases record.
const (
	RevokeLogout        RevokeReason = "logout"
	RevokeReuseDetected RevokeReason = "reuse_detected"
)

// StoredRefreshToken is a refresh token row as rotation sees it, locked
// until the unit of work ends.
type StoredRefreshToken struct {
	// ID is the store's opaque row id.
	ID string
	// Hash is the stored token hash; the use case compares it with the
	// presented token's in constant time.
	Hash    domain.TokenHash
	Rotated bool
	Revoked bool
	// Expired is true past the sliding or the absolute expiry, by the
	// database clock.
	Expired bool
}

// FamilyRevoker revokes token families (logout, reuse detection).
type FamilyRevoker interface {
	// RevokeFamily revokes every token not yet revoked in the family of the
	// unit of work user's token with hash; earlier revocations keep their
	// reason. No such token is not an error (logout is idempotent).
	RevokeFamily(ctx context.Context, hash domain.TokenHash, reason RevokeReason) error
}

// RefreshRotation is what refreshing a session needs of refresh_tokens, in
// the unit of work of the token's user.
type RefreshRotation interface {
	FamilyRevoker
	// Lock reads the unit of work user's token with hash and locks its row
	// (SELECT ... FOR UPDATE) until the unit of work ends, so concurrent
	// refreshes of one token run one after the other. found = false when
	// there is no such token.
	Lock(ctx context.Context, hash domain.TokenHash) (token StoredRefreshToken, found bool, err error)
	// Rotate marks parent rotated and stores successor in parent's family
	// with the family's absolute expiry. It returns the successor's
	// lifetime: sliding, or less when the absolute expiry comes first.
	Rotate(ctx context.Context, parent StoredRefreshToken, successor domain.RefreshToken, sliding time.Duration) (time.Duration, error)
}

// AccountStatus tells whether an account may still hold a session.
type AccountStatus interface {
	// CanSignIn reports whether the account exists and its status lets it
	// sign in (fail closed on any other status).
	CanSignIn(ctx context.Context, id domain.UserID) (bool, error)
}

// AccessTokens signs short-lived access tokens (EdDSA JWT).
type AccessTokens interface {
	Sign(userID domain.UserID, issuedAt time.Time) (token string, expiresIn time.Duration, err error)
}

// AccessClaims is what a verified access token asserts.
type AccessClaims struct {
	UserID domain.UserID
	// TokenID is the token's jti.
	TokenID   string
	ExpiresAt time.Time
}

// AccessTokenVerifier verifies a bearer access token. Every failure wraps
// ErrInvalidAccessToken.
type AccessTokenVerifier interface {
	Verify(token string) (AccessClaims, error)
}

// EmailSender delivers the auth emails. Implementations never put the
// address or the code in their errors.
type EmailSender interface {
	SendVerificationCode(ctx context.Context, to domain.Email, code domain.VerificationCode) error
	// SendSignUpAttemptNotice tells an account owner that someone tried to
	// register with their address (D3).
	SendSignUpAttemptNotice(ctx context.Context, to domain.Email) error
}

// Throttle admits at most a configured number of events per key and window
// (platform/ratelimit implements it).
type Throttle interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}
