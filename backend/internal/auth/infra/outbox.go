package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// The outbox definer functions (migration 20261023100000_auth_email_outbox).
// cc_app may run enqueue only; cc_worker claim, finish and retry only.
const (
	enqueueEmailSQL = "SELECT auth_email_outbox_enqueue(?, ?, ?)"
	claimEmailsSQL  = "SELECT id::text AS id, payload, attempts, claim_token::text AS claim_token FROM auth_email_outbox_claim(?, ?)"
	finishEmailSQL  = "SELECT auth_email_outbox_finish(?, ?)"
	retryEmailSQL   = "SELECT auth_email_outbox_retry(?, ?, ?)"
)

// outboxVersion versions the sealed payload, so a later shape change can read
// rows written by this one (they live at most usecase.CodeTTL).
const outboxVersion = 1

// outboxPlaintextSize is the padded size of every sealed plaintext, so the
// ciphertext length tells neither the kind (a code or a notice: whether the
// address holds an account) nor the address length. The longest payload (a
// 254-byte address and a code) is far below it.
const outboxPlaintextSize = 384

// outboxAADPrefix binds a ciphertext to its purpose and row.
const outboxAADPrefix = "auth_email_outbox:"

// Payload kinds, never stored in clear.
const (
	kindCode   = "code"
	kindNotice = "notice"
)

// outboxJSON is the sealed payload: what the worker needs to send, nothing more.
type outboxJSON struct {
	Version int    `json:"v"`
	Kind    string `json:"kind"`
	To      string `json:"to"`
	Code    string `json:"code"`
}

// Outbox implements usecase.EmailOutbox (the API, cc_app) and
// usecase.EmailQueue (the worker, cc_worker) over the outbox definer
// functions, in the unit of work carried by ctx. Payloads are sealed with
// the AUTH_OUTBOX_KEY keyring, which the API and the worker share.
type Outbox struct {
	keyring *crypto.Keyring
}

var (
	_ usecase.EmailOutbox = (*Outbox)(nil)
	_ usecase.EmailQueue  = (*Outbox)(nil)
)

// NewOutbox returns the adapter over keyring.
func NewOutbox(keyring *crypto.Keyring) *Outbox { return &Outbox{keyring: keyring} }

// Enqueue implements usecase.EmailOutbox.
func (o *Outbox) Enqueue(ctx context.Context, email usecase.OutboundEmail, expiresAt time.Time) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("auth infra: outbox id: %w", err)
	}
	sealed, err := o.seal(id, email)
	if err != nil {
		return err
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Exec(enqueueEmailSQL, id.String(), bytea(sealed), expiresAt).Error; err != nil {
		return fmt.Errorf("auth infra: enqueue email: %w", err)
	}
	return nil
}

// claimRow is one row of auth_email_outbox_claim.
type claimRow struct {
	ID         string
	Payload    []byte
	Attempts   int
	ClaimToken string
}

// Claim implements usecase.EmailQueue. A payload that does not open is
// returned Unreadable, never as an error: one bad row must not stop the queue.
func (o *Outbox) Claim(ctx context.Context, limit int, lease time.Duration) ([]usecase.ClaimedEmail, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []claimRow
	if err := db.Raw(claimEmailsSQL, limit, wholeSeconds(lease)).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("auth infra: claim emails: %w", err)
	}
	claimed := make([]usecase.ClaimedEmail, len(rows))
	for i, r := range rows {
		email, ok := o.openRow(r)
		claimed[i] = usecase.ClaimedEmail{
			ID: r.ID, Claim: r.ClaimToken, Attempt: r.Attempts, Email: email, Unreadable: !ok,
		}
	}
	return claimed, nil
}

func (o *Outbox) openRow(r claimRow) (usecase.OutboundEmail, bool) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return usecase.OutboundEmail{}, false
	}
	return o.open(id, r.Payload)
}

// Finish implements usecase.EmailQueue.
func (o *Outbox) Finish(ctx context.Context, id, claim string) (bool, error) {
	return callHeld(ctx, finishEmailSQL, id, claim)
}

// Retry implements usecase.EmailQueue.
func (o *Outbox) Retry(ctx context.Context, id, claim string, delay time.Duration) (bool, error) {
	return callHeld(ctx, retryEmailSQL, id, claim, wholeSeconds(delay))
}

func callHeld(ctx context.Context, query string, args ...any) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var held bool
	if err := db.Raw(query, args...).Scan(&held).Error; err != nil {
		return false, fmt.Errorf("auth infra: settle email: %w", err)
	}
	return held, nil
}

// wholeSeconds rounds d up to a whole number of seconds, at least one.
func wholeSeconds(d time.Duration) int {
	return max(1, int((d+time.Second-1)/time.Second))
}

func outboxAAD(id uuid.UUID) []byte { return []byte(outboxAADPrefix + id.String()) }

// seal encodes email, pads it to outboxPlaintextSize and encrypts it bound to id.
func (o *Outbox) seal(id uuid.UUID, email usecase.OutboundEmail) ([]byte, error) {
	doc := outboxJSON{Version: outboxVersion, To: email.To.String()}
	switch email.Kind {
	case usecase.EmailVerificationCode:
		doc.Kind, doc.Code = kindCode, email.Code.Reveal()
	case usecase.EmailSignUpAttemptNotice:
		doc.Kind = kindNotice
	default:
		return nil, fmt.Errorf("auth infra: unknown email kind %d", email.Kind)
	}
	plain, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("auth infra: encode email: %w", err)
	}
	if len(plain) > outboxPlaintextSize {
		return nil, fmt.Errorf("auth infra: email payload of %d bytes exceeds %d", len(plain), outboxPlaintextSize)
	}
	padded := bytes.Repeat([]byte{' '}, outboxPlaintextSize)
	copy(padded, plain)
	sealed, err := o.keyring.Encrypt(padded, outboxAAD(id))
	if err != nil {
		return nil, fmt.Errorf("auth infra: seal email: %w", err)
	}
	return sealed, nil
}

// open decrypts and validates a payload sealed by seal for id. It re-checks
// what the domain owns: the payload crossed a trust boundary (the database).
func (o *Outbox) open(id uuid.UUID, sealed []byte) (usecase.OutboundEmail, bool) {
	plain, err := o.keyring.Decrypt(sealed, outboxAAD(id))
	if err != nil {
		return usecase.OutboundEmail{}, false
	}
	dec := json.NewDecoder(bytes.NewReader(plain))
	dec.DisallowUnknownFields()
	var doc outboxJSON
	if err := dec.Decode(&doc); err != nil || doc.Version != outboxVersion {
		return usecase.OutboundEmail{}, false
	}
	to, err := domain.NewEmail(doc.To)
	if err != nil {
		return usecase.OutboundEmail{}, false
	}
	switch {
	case doc.Kind == kindNotice && doc.Code == "":
		return usecase.OutboundEmail{Kind: usecase.EmailSignUpAttemptNotice, To: to}, true
	case doc.Kind == kindCode:
		code, err := domain.ParseVerificationCode(doc.Code)
		if err != nil {
			return usecase.OutboundEmail{}, false
		}
		return usecase.OutboundEmail{Kind: usecase.EmailVerificationCode, To: to, Code: code}, true
	default:
		return usecase.OutboundEmail{}, false
	}
}
