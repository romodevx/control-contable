package infra

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// bytea binds a []byte as one bytea parameter. GORM expands a bare []byte
// passed to Raw/Exec into one placeholder per byte; a driver.Valuer is
// bound as a single value.
type bytea []byte

// Value implements driver.Valuer.
func (b bytea) Value() (driver.Value, error) { return []byte(b), nil }

// ErrCorruptPending reports a pending sign-up payload this code cannot read.
var ErrCorruptPending = errors.New("auth infra: unreadable pending sign-up")

// errUnknownOutcome reports a definer function answering an outcome this
// code does not know: the migration and the adapter disagree.
var errUnknownOutcome = errors.New("auth infra: unknown outcome")

// The definer functions (migrations 20261013150000_auth_service and
// 20261013160000_auth_signup). They are the only way cc_app reaches
// email_verifications, trial_registry and other users' credentials.
const (
	lookupCredentialsSQL = "SELECT count(*) FROM auth_lookup_credentials(?)" //nolint:gosec // G101: a name or message, not a credential
	startSignUpSQL       = "SELECT email_verification_start(?, ?, ?::jsonb, ?)"
	resendCodeSQL        = "SELECT email_verification_resend(?, ?, ?)"
	consumeCodeSQL       = "SELECT outcome, pending::text AS pending FROM email_verification_consume(?, ?)"
	claimTrialSQL        = "SELECT trial_registry_check_and_insert(?)"
)

var codeOutcomes = map[string]usecase.CodeOutcome{
	"started":   usecase.CodeStarted,
	"resent":    usecase.CodeResent,
	"unchanged": usecase.CodeUnchanged,
}

var consumeOutcomes = map[string]usecase.ConsumeOutcome{
	"verified": usecase.ConsumeVerified,
	"invalid":  usecase.ConsumeInvalid,
	"expired":  usecase.ConsumeExpired,
	"locked":   usecase.ConsumeLocked,
}

// SignUpStore implements the sign-up ports over the definer functions, in
// the unit of work carried by ctx.
type SignUpStore struct{}

// NewSignUpStore returns the store.
func NewSignUpStore() SignUpStore { return SignUpStore{} }

var (
	_ usecase.SignUpStarter  = SignUpStore{}
	_ usecase.CodeResender   = SignUpStore{}
	_ usecase.SignUpVerifier = SignUpStore{}
)

// AccountExists implements usecase.SignUpStarter.
func (SignUpStore) AccountExists(ctx context.Context, emailHash []byte) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var matches int64
	if err := db.Raw(lookupCredentialsSQL, bytea(emailHash)).Scan(&matches).Error; err != nil {
		return false, fmt.Errorf("auth infra: look up credentials: %w", err)
	}
	return matches > 0, nil
}

// Start implements usecase.SignUpStarter.
func (SignUpStore) Start(ctx context.Context, emailHash, codeHash []byte, pending usecase.PendingSignUp, expiresAt time.Time) (usecase.CodeOutcome, error) {
	payload, err := encodePending(pending)
	if err != nil {
		return usecase.CodeUnchanged, err
	}
	return callCodeFunction(ctx, startSignUpSQL, bytea(emailHash), bytea(codeHash), payload, expiresAt)
}

// Resend implements usecase.CodeResender.
func (SignUpStore) Resend(ctx context.Context, emailHash, codeHash []byte, expiresAt time.Time) (usecase.CodeOutcome, error) {
	return callCodeFunction(ctx, resendCodeSQL, bytea(emailHash), bytea(codeHash), expiresAt)
}

func callCodeFunction(ctx context.Context, query string, args ...any) (usecase.CodeOutcome, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.CodeUnchanged, err
	}
	var raw string
	if err := db.Raw(query, args...).Scan(&raw).Error; err != nil {
		return usecase.CodeUnchanged, fmt.Errorf("auth infra: issue code: %w", err)
	}
	outcome, ok := codeOutcomes[raw]
	if !ok {
		return usecase.CodeUnchanged, fmt.Errorf("%w %q", errUnknownOutcome, raw)
	}
	return outcome, nil
}

// consumeRow is one row of email_verification_consume.
type consumeRow struct {
	Outcome string
	Pending sql.NullString
}

// Consume implements usecase.SignUpVerifier.
func (SignUpStore) Consume(ctx context.Context, emailHash, codeHash []byte) (usecase.ConsumeOutcome, usecase.PendingSignUp, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.ConsumeInvalid, usecase.PendingSignUp{}, err
	}
	var row consumeRow
	if err := db.Raw(consumeCodeSQL, bytea(emailHash), bytea(codeHash)).Scan(&row).Error; err != nil {
		return usecase.ConsumeInvalid, usecase.PendingSignUp{}, fmt.Errorf("auth infra: consume code: %w", err)
	}
	outcome, ok := consumeOutcomes[row.Outcome]
	if !ok {
		return usecase.ConsumeInvalid, usecase.PendingSignUp{}, fmt.Errorf("%w %q", errUnknownOutcome, row.Outcome)
	}
	if outcome != usecase.ConsumeVerified {
		return outcome, usecase.PendingSignUp{}, nil
	}
	pending, err := decodePending(row.Pending.String)
	if err != nil {
		return usecase.ConsumeInvalid, usecase.PendingSignUp{}, err
	}
	return outcome, pending, nil
}

// ClaimTrial implements usecase.SignUpVerifier.
func (SignUpStore) ClaimTrial(ctx context.Context, emailHash []byte) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var first bool
	if err := db.Raw(claimTrialSQL, bytea(emailHash)).Scan(&first).Error; err != nil {
		return false, fmt.Errorf("auth infra: claim trial: %w", err)
	}
	return first, nil
}

// pendingVersion versions the pending payload, so a later shape change can
// read rows written by this one (they live at most CodeTTL). Version 1 held a
// password hash and a name chosen at register; it is refused, never turned
// into an account (D8).
const pendingVersion = 2

// pendingJSON is email_verifications.pending: the accepted documents only.
// No password (hash or not), no name (D8) and no email (the row is keyed by
// its hash).
type pendingJSON struct {
	Version    int            `json:"v"`
	Documents  []documentJSON `json:"accepted_documents"`
	AcceptedAt time.Time      `json:"accepted_at"`
}

type documentJSON struct {
	Kind    string `json:"kind"`
	Version string `json:"version"`
}

func encodePending(p usecase.PendingSignUp) (string, error) {
	doc := pendingJSON{Version: pendingVersion, AcceptedAt: p.AcceptedAt.UTC()}
	for _, d := range p.Documents {
		doc.Documents = append(doc.Documents, documentJSON{Kind: string(d.Kind), Version: d.Version})
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return "", fmt.Errorf("auth infra: encode pending sign-up: %w", err)
	}
	return string(raw), nil
}

// decodePending reads a payload back, re-validating what the domain owns:
// it crossed a trust boundary (the database).
func decodePending(raw string) (usecase.PendingSignUp, error) {
	dec := json.NewDecoder(bytes.NewReader([]byte(raw)))
	dec.DisallowUnknownFields()
	var doc pendingJSON
	if err := dec.Decode(&doc); err != nil || doc.Version != pendingVersion {
		return usecase.PendingSignUp{}, ErrCorruptPending
	}
	pending := usecase.PendingSignUp{AcceptedAt: doc.AcceptedAt}
	documents := make([]domain.DocumentVersion, 0, len(doc.Documents))
	for _, d := range doc.Documents {
		documents = append(documents, domain.DocumentVersion{Kind: domain.DocumentKind(d.Kind), Version: d.Version})
	}
	acceptance, err := domain.NewAcceptance(documents)
	if err != nil {
		return usecase.PendingSignUp{}, ErrCorruptPending
	}
	pending.Documents = acceptance.Documents()
	return pending, nil
}
