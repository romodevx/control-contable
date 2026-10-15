package usecase_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

var (
	fixedClock = clock.NewFixed(time.Date(2026, 10, 13, 18, 30, 0, 0, time.UTC))
	now        = fixedClock.Now() // in America/Mexico_City, as every Clock returns
	errDB      = errors.New("database unavailable")
)

const (
	validEmail     = "Ana.Gomez@Example.MX"
	normalEmail    = "ana.gomez@example.mx"
	acceptedSample = "tlacuache montaña 2026" // made-up sample phrase, not a credential
	commonSample   = "passwordpassword"       // on the embedded common list
	shortSample    = "short"                  // under the 12-character minimum
	newUserID      = "0199d4a2-7c1e-7b3a-9f10-2c5e8a1b4d01"
)

// zeroReader is a deterministic entropy source: every code it draws is
// "000000" and every refresh secret is all zero bytes.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	clear(p)
	return len(p), nil
}

var _ io.Reader = zeroReader{}

// sha256Index stands in for the HMAC blind index: deterministic, 32 bytes.
type sha256Index struct{ calls int }

func (s *sha256Index) Hash(value string) ([]byte, error) {
	s.calls++
	sum := sha256.Sum256([]byte("index:" + value))
	return sum[:], nil
}

func emailHashOf(email string) []byte {
	sum := sha256.Sum256([]byte("index:" + email))
	return sum[:]
}

// prefixHasher stands in for argon2id and counts its calls (timing parity).
type prefixHasher struct{ calls int }

func (p *prefixHasher) Hash(password string) (string, error) {
	p.calls++
	return "hashed:" + password, nil
}

// concatCodes stands in for the HMAC of the code.
type concatCodes struct{}

func (concatCodes) HashCode(emailHash []byte, code domain.VerificationCode) []byte {
	sum := sha256.Sum256(append(bytes.Clone(emailHash), code.Reveal()...))
	return sum[:]
}

func codeHashOf(email, code string) []byte {
	c, err := domain.ParseVerificationCode(code)
	if err != nil {
		panic(err)
	}
	return concatCodes{}.HashCode(emailHashOf(email), c)
}

// fakeLegal serves the current legal document versions.
type fakeLegal struct {
	current map[domain.DocumentKind]string
	err     error
}

func (f fakeLegal) CurrentVersions(context.Context) (map[domain.DocumentKind]string, error) {
	return f.current, f.err
}

var currentDocs = fakeLegal{current: map[domain.DocumentKind]string{
	domain.DocumentTerms: "v1", domain.DocumentPrivacyIntegral: "v1", domain.DocumentPrivacySimplified: "v1",
}}

// fakeSystem runs work inline, as tx.SystemManager would, and records it.
type fakeSystem struct {
	calls int
	err   error
}

func (f *fakeSystem) System(ctx context.Context, work tx.Work) error {
	f.calls++
	if f.err != nil {
		return f.err
	}
	return work(ctx)
}

// fakeTx runs work inline as userID and records the work's own result, so a
// test can tell a commit (nil) from a rollback (error).
type fakeTx struct {
	userID  string
	workErr error
	calls   int
}

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.calls++
	f.userID = userID
	f.workErr = work(context.WithValue(ctx, txUserKey{}, userID))
	return f.workErr
}

// txUserKey marks a context inside fakeTx's unit of work.
type txUserKey struct{}

// txUser is the user of the unit of work ctx runs in; "" outside one.
func txUser(ctx context.Context) string {
	user, _ := ctx.Value(txUserKey{}).(string)
	return user
}

func (f *fakeTx) ReadOnly(context.Context, string, tx.Work) error {
	return errors.New("fakeTx: read-only work is not expected")
}

// startCall records one SignUps.Start.
type startCall struct {
	emailHash, codeHash []byte
	pending             usecase.PendingSignUp
	expiresAt           time.Time
}

// resendCall records one SignUps.Resend.
type resendCall struct {
	emailHash, codeHash []byte
	expiresAt           time.Time
}

// fakeSignUps scripts the definer functions.
type fakeSignUps struct {
	existing      bool
	startOutcome  usecase.CodeOutcome
	resendOutcome usecase.CodeOutcome
	consume       usecase.ConsumeOutcome
	pending       usecase.PendingSignUp
	firstTrial    bool
	err           error

	starts      []startCall
	resends     []resendCall
	consumed    [][2][]byte
	trialClaims [][]byte
}

func (f *fakeSignUps) AccountExists(_ context.Context, emailHash []byte) (bool, error) {
	return f.existing, f.err
}

func (f *fakeSignUps) Start(_ context.Context, emailHash, codeHash []byte, pending usecase.PendingSignUp, expiresAt time.Time) (usecase.CodeOutcome, error) {
	f.starts = append(f.starts, startCall{emailHash, codeHash, pending, expiresAt})
	return f.startOutcome, f.err
}

func (f *fakeSignUps) Resend(_ context.Context, emailHash, codeHash []byte, expiresAt time.Time) (usecase.CodeOutcome, error) {
	f.resends = append(f.resends, resendCall{emailHash, codeHash, expiresAt})
	return f.resendOutcome, f.err
}

func (f *fakeSignUps) Consume(_ context.Context, emailHash, codeHash []byte) (usecase.ConsumeOutcome, usecase.PendingSignUp, error) {
	f.consumed = append(f.consumed, [2][]byte{emailHash, codeHash})
	if f.consume != usecase.ConsumeVerified {
		return f.consume, usecase.PendingSignUp{}, f.err
	}
	return f.consume, f.pending, f.err
}

func (f *fakeSignUps) ClaimTrial(_ context.Context, emailHash []byte) (bool, error) {
	f.trialClaims = append(f.trialClaims, emailHash)
	return f.firstTrial, nil
}

// sent is one email the fake sender accepted.
type sent struct {
	kind string // "code" or "notice"
	to   string
	code string
}

// fakeSender records emails; err makes every send fail.
type fakeSender struct {
	mu   sync.Mutex
	sent []sent
	err  error
}

func (f *fakeSender) SendVerificationCode(_ context.Context, to domain.Email, code domain.VerificationCode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sent{kind: "code", to: to.String(), code: code.Reveal()})
	return f.err
}

func (f *fakeSender) SendSignUpAttemptNotice(_ context.Context, to domain.Email) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, sent{kind: "notice", to: to.String()})
	return f.err
}

// fakeLimiter admits limit events per key (a shared Postgres bucket in
// production), then refuses with a one-minute wait. It records every key
// taken and reset; err makes every call fail.
type fakeLimiter struct {
	mu     sync.Mutex
	limit  int
	err    error
	keys   []string
	resets []string
}

func (f *fakeLimiter) Take(_ context.Context, key string) (bool, time.Duration, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return false, 0, f.err
	}
	f.keys = append(f.keys, key)
	spent := 0
	for _, k := range f.keys {
		if k == key {
			spent++
		}
	}
	if spent > f.limit {
		return false, time.Minute, nil
	}
	return true, 0, nil
}

func (f *fakeLimiter) Reset(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.resets = append(f.resets, key)
	f.keys = slices.DeleteFunc(f.keys, func(k string) bool { return k == key })
	return nil
}

func (f *fakeLimiter) taken() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.keys)
}

// onceThrottle allows each key once.
type onceThrottle struct{ seen map[string]bool }

func (o *onceThrottle) Allow(key string) (bool, time.Duration) {
	if o.seen == nil {
		o.seen = map[string]bool{}
	}
	if o.seen[key] {
		return false, time.Hour
	}
	o.seen[key] = true
	return true, 0
}

// fakeAccounts records the created account.
type fakeAccounts struct {
	created []usecase.NewAccount
	err     error
}

func (f *fakeAccounts) Create(_ context.Context, a usecase.NewAccount) (usecase.Account, error) {
	f.created = append(f.created, a)
	if f.err != nil {
		return usecase.Account{}, f.err
	}
	return usecase.Account{ID: a.ID, Email: a.Email, Name: a.Name, CreatedAt: now}, nil
}

// fakeTrials records started trials.
type fakeTrials struct {
	started []time.Time
	userIDs []domain.UserID
}

func (f *fakeTrials) StartTrial(_ context.Context, userID domain.UserID, startsAt time.Time) error {
	f.userIDs = append(f.userIDs, userID)
	f.started = append(f.started, startsAt)
	return nil
}

// familyCall records one RefreshTokens.StartFamily.
type familyCall struct {
	token             domain.RefreshToken
	sliding, absolute time.Duration
}

type fakeRefresh struct{ families []familyCall }

func (f *fakeRefresh) StartFamily(_ context.Context, token domain.RefreshToken, sliding, absolute time.Duration) error {
	f.families = append(f.families, familyCall{token, sliding, absolute})
	return nil
}

// fakeAccess signs "access:<user>@<unix>".
type fakeAccess struct{}

const accessTTL = 15 * time.Minute

func (fakeAccess) Sign(userID domain.UserID, issuedAt time.Time) (string, time.Duration, error) {
	return "access:" + userID.String() + "@" + issuedAt.UTC().Format(time.RFC3339), accessTTL, nil
}

func fixedUserID() (domain.UserID, error) { return domain.ParseUserID(newUserID) }

// capturedLogs returns a logger writing JSON to the returned buffer.
func capturedLogs() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

func hexOf(b []byte) string { return hex.EncodeToString(b) }
