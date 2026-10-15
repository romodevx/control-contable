package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/adapter/http/router"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	reminderstore "github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
)

// signUpNoticeInterval is D3's "at most once per hour" per email for the
// "someone tried to sign up with your email" notice. In process (ADR-0001 O9):
// each instance keeps its own window.
const signUpNoticeInterval = time.Hour

// authWiring is what the auth module needs from the composition root.
type authWiring struct {
	keys config.Auth
	// outbox seals the queued auth emails; the worker opens them with the same key.
	outbox config.AuthOutbox
	pool   *sql.DB         // cc_app, for the shared rate-limit store
	tx     *pgdb.TxManager // the shared cc_app unit of work (appTx)
	logger *slog.Logger
	now    func() time.Time
	// argon2 overrides the password hashing cost (tests); zero means
	// crypto.DefaultArgon2idParams.
	argon2 crypto.Argon2idParams
}

// authModule is the auth module's HTTP side: the handler of its operations
// and the bearer authenticator every secured operation runs behind.
type authModule struct {
	handler  *authhttp.Handler
	account  *authhttp.AccountHandler
	sessions *authhttp.SessionHandler
	authn    *authhttp.Authenticator
	// clientLimits are the per-client limits of the auth operations
	// (MODULE-auth §5), by operationId, for the router.
	clientLimits map[string]router.Limiter
	// limits is the shared rate-limit store (ADR-0005); other modules key
	// their own scopes in it.
	limits *ratelimitpg.Store
}

// newAuthModule wires the auth use cases (M1-05, M1-07) over the shared
// cc_app unit of work and the access-token authenticator.
func newAuthModule(w authWiring) (authModule, error) {
	if w.tx == nil {
		return authModule{}, fmt.Errorf("auth: %w", errNoAppTx)
	}
	txm := w.tx
	keys, err := newAuthCrypto(w)
	if err != nil {
		return authModule{}, err
	}
	clk := clock.NewSystem()
	signUps := infra.NewSignUpStore()
	// Rate-limit buckets shared by every instance (ADR-0005), keyed by HMACs
	// under a key derived from the email pepper.
	limits, err := ratelimitpg.NewStore(w.pool, w.keys.EmailPepper.RevealBytes(), w.now)
	if err != nil {
		return authModule{}, fmt.Errorf("auth: rate limits: %w", err)
	}
	signUpEmails := limits.Bucket(usecase.SignUpEmailPolicy)
	// Register and resend only queue their emails; cmd/worker delivers them
	// (auth_email_delivery), so no request waits on an email provider.
	outbox, err := newOutbox(w.outbox)
	if err != nil {
		return authModule{}, err
	}

	register := usecase.NewRegister(usecase.RegisterDeps{
		System: txm, Clock: clk, Entropy: rand.Reader, Index: keys.index,
		Codes: keys.codes, Legal: infra.NewLegalDocumentStore(), SignUps: signUps, Outbox: outbox,
		Notices: ratelimit.NewEvery(signUpNoticeInterval, 1, w.now), Emails: signUpEmails, Logger: w.logger,
	})
	resend := usecase.NewResendVerification(usecase.ResendDeps{
		System: txm, Clock: clk, Entropy: rand.Reader, Index: keys.index, Codes: keys.codes,
		SignUps: signUps, Outbox: outbox, Emails: signUpEmails, Logger: w.logger,
	})
	verify := usecase.NewVerifyEmail(usecase.VerifyDeps{
		Tx: txm, Clock: clk, Entropy: rand.Reader, NewUserID: newUserID, Index: keys.index,
		Passwords: keys.passwords, Codes: keys.codes, SignUps: signUps, Accounts: infra.NewAccountStore(keys.emails),
		Trials: infra.NewTrialStore(), Refresh: infra.NewRefreshTokenStore(), Access: keys.access,
		Logger: w.logger,
	})
	login, err := usecase.NewLogin(usecase.LoginDeps{
		System: txm, Tx: txm, Clock: clk, Entropy: rand.Reader, Index: keys.index, Passwords: keys.passwords,
		Credentials: infra.NewCredentialStore(), Accounts: infra.NewAccountStore(keys.emails),
		Refresh: infra.NewRefreshTokenStore(), Access: keys.access,
		Failures: limits.Bucket(usecase.LoginFailuresPolicy), Ceiling: limits.Bucket(usecase.LoginCeilingPolicy), Logger: w.logger,
	})
	if err != nil {
		return authModule{}, err
	}
	authn, err := newAuthenticator(w.keys, w.now, w.logger)
	if err != nil {
		return authModule{}, err
	}
	handler := authhttp.NewHandler(authhttp.UseCases{Register: register, Verify: verify, Resend: resend, Login: login})
	return authModule{
		handler:      handler,
		account:      newAccountHandler(txm, keys, limits, w.logger),
		sessions:     newSessionHandler(txm, keys, limits, w.logger),
		authn:        authn,
		clientLimits: clientLimits(limits),
		limits:       limits,
	}, nil
}

// clientLimits builds the router's per-client limits from the auth
// adapter's policies.
func clientLimits(limits *ratelimitpg.Store) map[string]router.Limiter {
	out := make(map[string]router.Limiter, len(authhttp.ClientLimits))
	for op, policy := range authhttp.ClientLimits {
		out[op] = limits.Bucket(policy)
	}
	return out
}

// newSessionHandler wires refreshSession and logout (M1-08).
func newSessionHandler(txm *pgdb.TxManager, keys authCrypto, limits *ratelimitpg.Store, logger *slog.Logger) *authhttp.SessionHandler {
	tokens := infra.NewRefreshTokenStore()
	refresh := usecase.NewRefreshSession(usecase.RefreshDeps{
		Tx: txm, Clock: clock.NewSystem(), Entropy: rand.Reader, Tokens: tokens,
		Accounts: infra.NewAccountStore(keys.emails), Access: keys.access,
		Rotations: limits.Bucket(usecase.RefreshFamilyPolicy), Logger: logger,
	})
	// The push device goes with the session (M9 security gate F1): the
	// reminders store runs in the logout's own unit of work (pgdb.Conn).
	logout := usecase.NewLogout(usecase.LogoutDeps{Tx: txm, Tokens: tokens, Devices: reminderstore.NewDevices(), Logger: logger})
	return authhttp.NewSessionHandler(refresh, logout)
}

// newAuthenticator builds the bearer authenticator over the access-token
// key set: the active key and, during a rotation, the previous one.
func newAuthenticator(keys config.Auth, now func() time.Time, logger *slog.Logger) (*authhttp.Authenticator, error) {
	set, err := accessTokenKeySet(keys)
	if err != nil {
		return nil, err
	}
	verifier, err := infra.NewAccessTokenVerifier(now, set)
	if err != nil {
		return nil, fmt.Errorf("auth: access token verifier: %w", err)
	}
	return authhttp.NewAuthenticator(verifier, logger), nil
}

// accessTokenKeySet is the JWKS-style verification set: kid → public key.
func accessTokenKeySet(keys config.Auth) (map[string]ed25519.PublicKey, error) {
	active, err := infra.PublicKeyFromSeed(keys.AccessTokenKey.RevealBytes())
	if err != nil {
		return nil, fmt.Errorf("auth: access token key: %w", err)
	}
	set := map[string]ed25519.PublicKey{keys.AccessTokenKeyID: active}
	if keys.PreviousAccessTokenKey.IsZero() {
		return set, nil
	}
	previous, err := infra.PublicKeyFromSeed(keys.PreviousAccessTokenKey.RevealBytes())
	if err != nil {
		return nil, fmt.Errorf("auth: previous access token key: %w", err)
	}
	set[keys.PreviousAccessTokenKeyID] = previous
	return set, nil
}

// authCrypto is the auth module's key material, ready to use.
type authCrypto struct {
	index     *crypto.BlindIndex
	emails    *crypto.Keyring
	passwords *crypto.PasswordHasher
	codes     *infra.CodeHasher
	access    *infra.AccessTokenSigner
}

func newAuthCrypto(w authWiring) (authCrypto, error) {
	var (
		c   authCrypto
		err error
	)
	if c.index, err = crypto.NewEmailIndex(w.keys.EmailPepper.RevealBytes()); err != nil {
		return authCrypto{}, fmt.Errorf("auth: email index: %w", err)
	}
	keyID := crypto.KeyID(w.keys.EmailKeyID) // #nosec G115 -- config bounds it to [1, crypto.MaxKeyID]
	if c.emails, err = crypto.NewKeyring(keyID, map[crypto.KeyID][]byte{keyID: w.keys.EmailKey.RevealBytes()}); err != nil {
		return authCrypto{}, fmt.Errorf("auth: email keyring: %w", err)
	}
	params := w.argon2
	if params == (crypto.Argon2idParams{}) {
		params = crypto.DefaultArgon2idParams()
	}
	if c.passwords, err = crypto.NewPasswordHasher(params); err != nil {
		return authCrypto{}, fmt.Errorf("auth: password hasher: %w", err)
	}
	if c.codes, err = infra.NewCodeHasher(w.keys.CodeKey.RevealBytes()); err != nil {
		return authCrypto{}, err
	}
	if c.access, err = infra.NewAccessTokenSigner(w.keys.AccessTokenKey.RevealBytes(), w.keys.AccessTokenKeyID); err != nil {
		return authCrypto{}, err
	}
	return c, nil
}

// newOutbox is the auth email outbox sealed with cfg's key.
func newOutbox(cfg config.AuthOutbox) (*infra.Outbox, error) {
	keyID := crypto.KeyID(cfg.KeyID) // #nosec G115 -- config bounds it to [1, crypto.MaxKeyID]
	keyring, err := crypto.NewKeyring(keyID, map[crypto.KeyID][]byte{keyID: cfg.Key.RevealBytes()})
	if err != nil {
		return nil, fmt.Errorf("auth: outbox keyring: %w", err)
	}
	return infra.NewOutbox(keyring), nil
}

// newUserID draws a UUIDv7: time-ordered, so new users append to the users
// primary key index instead of scattering across it.
func newUserID() (domain.UserID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.UserID{}, fmt.Errorf("uuid v7: %w", err)
	}
	return domain.UserID(id), nil
}
