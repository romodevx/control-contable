package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

// signUpNoticeInterval is D3's "at most once per hour" per email for the
// "someone tried to sign up with your email" notice. In process (ADR-0001 O9):
// each instance keeps its own window.
const signUpNoticeInterval = time.Hour

// authWiring is what the auth module needs from the composition root.
type authWiring struct {
	env    config.Environment
	keys   config.Auth
	pool   *sql.DB // cc_app
	logger *slog.Logger
	now    func() time.Time
	// argon2 overrides the password hashing cost (tests); zero means
	// crypto.DefaultArgon2idParams.
	argon2 crypto.Argon2idParams
}

// newAuthHandler wires the auth use cases (M1-05) over the cc_app pool.
func newAuthHandler(w authWiring) (*authhttp.Handler, error) {
	txm, err := pgdb.NewTxManager(w.pool)
	if err != nil {
		return nil, err
	}
	keys, err := newAuthCrypto(w)
	if err != nil {
		return nil, err
	}
	clk := clock.NewSystem()
	signUps := infra.NewSignUpStore()
	sender := infra.NewLogSender(w.logger, w.env) // M1-06 replaces it with the provider adapter

	register := usecase.NewRegister(usecase.RegisterDeps{
		System: txm, Clock: clk, Entropy: rand.Reader, Index: keys.index, Passwords: keys.passwords,
		Codes: keys.codes, Legal: infra.NewLegalDocumentStore(), SignUps: signUps, Email: sender,
		Notices: ratelimit.NewEvery(signUpNoticeInterval, 1, w.now), Logger: w.logger,
	})
	resend := usecase.NewResendVerification(usecase.ResendDeps{
		System: txm, Clock: clk, Entropy: rand.Reader, Index: keys.index, Codes: keys.codes,
		SignUps: signUps, Email: sender, Logger: w.logger,
	})
	verify := usecase.NewVerifyEmail(usecase.VerifyDeps{
		Tx: txm, Clock: clk, Entropy: rand.Reader, NewUserID: newUserID, Index: keys.index,
		Codes: keys.codes, SignUps: signUps, Accounts: infra.NewAccountStore(keys.emails),
		Trials: infra.NewTrialStore(), Refresh: infra.NewRefreshTokenStore(), Access: keys.access,
		Logger: w.logger,
	})
	return authhttp.NewHandler(register, verify, resend), nil
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

// newUserID draws a UUIDv7: time-ordered, so new users append to the users
// primary key index instead of scattering across it.
func newUserID() (domain.UserID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.UserID{}, fmt.Errorf("uuid v7: %w", err)
	}
	return domain.UserID(id), nil
}
