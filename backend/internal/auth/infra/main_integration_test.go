//go:build integration

package infra_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/infra"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// inbox is an EmailSender that keeps what it was asked to send.
type inbox struct {
	mu      sync.Mutex
	codes   map[string][]string // email -> codes, oldest first
	notices map[string]int
}

func newInbox() *inbox { return &inbox{codes: map[string][]string{}, notices: map[string]int{}} }

func (i *inbox) SendVerificationCode(_ context.Context, to domain.Email, code domain.VerificationCode) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.codes[to.String()] = append(i.codes[to.String()], code.Reveal())
	return nil
}

func (i *inbox) SendSignUpAttemptNotice(_ context.Context, to domain.Email) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.notices[to.String()]++
	return nil
}

func (i *inbox) lastCode(t *testing.T, email string) string {
	t.Helper()
	i.mu.Lock()
	defer i.mu.Unlock()
	codes := i.codes[email]
	require.NotEmpty(t, codes, "no code sent to %s", email)
	return codes[len(codes)-1]
}

// stack is the auth use cases over a real database as cc_app, wired as cmd/api
// wires them (only the email sender and the argon2 cost differ).
type stack struct {
	app, owner *sql.DB
	register   *usecase.Register
	resend     *usecase.ResendVerification
	verify     *usecase.VerifyEmail
	inbox      *inbox
	index      *crypto.BlindIndex
	keyring    *crypto.Keyring
	passwords  *crypto.PasswordHasher
}

var (
	testPepper   = bytes.Repeat([]byte{0x11}, 32)
	testEmailKey = bytes.Repeat([]byte{0x22}, 32)
)

// testArgon2 keeps the suite fast; production uses crypto.DefaultArgon2idParams.
var testArgon2 = crypto.Argon2idParams{Memory: 8 * 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}

func newStack(t *testing.T) *stack {
	t.Helper()
	db := pgtest.NewDatabase(t)
	s := &stack{app: db.Connect(t, pgtest.RoleApp), owner: db.Connect(t, pgtest.RoleOwner), inbox: newInbox()}
	txm, err := pgdb.NewTxManager(s.app)
	require.NoError(t, err)
	s.index, err = crypto.NewEmailIndex(testPepper)
	require.NoError(t, err)
	s.keyring, err = crypto.NewKeyring(1, map[crypto.KeyID][]byte{1: testEmailKey})
	require.NoError(t, err)
	s.passwords, err = crypto.NewPasswordHasher(testArgon2)
	require.NoError(t, err)
	codes, err := infra.NewCodeHasher(codeKey)
	require.NoError(t, err)
	access, err := infra.NewAccessTokenSigner(tokenSeed, "k1")
	require.NoError(t, err)
	logger := slog.New(slog.DiscardHandler)
	signUps := infra.NewSignUpStore()

	s.register = usecase.NewRegister(usecase.RegisterDeps{
		System: txm, Clock: clock.NewSystem(), Entropy: rand.Reader, Index: s.index,
		Codes: codes, Legal: infra.NewLegalDocumentStore(), SignUps: signUps, Email: s.inbox,
		Notices: ratelimit.NewEvery(time.Hour, 1, time.Now), Logger: logger,
	})
	s.resend = usecase.NewResendVerification(usecase.ResendDeps{
		System: txm, Clock: clock.NewSystem(), Entropy: rand.Reader, Index: s.index, Codes: codes,
		SignUps: signUps, Email: s.inbox, Logger: logger,
	})
	s.verify = usecase.NewVerifyEmail(usecase.VerifyDeps{
		Tx: txm, Clock: clock.NewSystem(), Entropy: rand.Reader, NewUserID: newUserID, Index: s.index,
		Passwords: s.passwords, Codes: codes, SignUps: signUps, Accounts: infra.NewAccountStore(s.keyring), Trials: infra.NewTrialStore(),
		Refresh: infra.NewRefreshTokenStore(), Access: access, Logger: logger,
	})
	return s
}

func newUserID() (domain.UserID, error) {
	id, err := uuid.NewV7()
	return domain.UserID(id), err
}

// ownerExec runs a fixture statement as cc_owner (it owns the service tables).
func (s *stack) ownerExec(t *testing.T, query string) {
	t.Helper()
	_, err := s.owner.ExecContext(t.Context(), query)
	require.NoError(t, err)
}

// asUser runs fn in a transaction under userID's RLS context, as cc_app.
func (s *stack) asUser(t *testing.T, userID string, fn func(*sql.Tx) error) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), s.app, userID, fn))
}
