// Package infra is the PostgreSQL adapter of the key re-seal job (ADR-0012,
// migration 20261024100000): one Store per sealed column and the key_usage
// reader, all on the unit of work in ctx.
//
// Pending calls the column's reseal_pending_* SECURITY DEFINER function
// (owner cc_worker_definer, EXECUTE cc_worker) in a System unit of work: ids
// only. Lock and Replace run in the row owner's unit of work, so RLS (the
// cc_worker per-user policies) scopes them: a ref naming another user's row
// finds nothing and updates nothing.
package infra

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/keyrotation/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// Targets name the sealed columns as key_usage does.
const (
	TargetUserEmail   = "users.email_encrypted"
	TargetTaxpayerRFC = "taxpayers.rfc_encrypted"
	TargetDeviceToken = "devices.token_ciphertext"
)

// ErrRowNotUpdated reports a Replace that matched no row: gone, or not the
// unit of work's user's.
var ErrRowNotUpdated = errors.New("keyrotation infra: sealed row not updated")

// columnSQL is one sealed column's statements. The value is read FOR UPDATE
// so a concurrent re-seal of the same row (worker split brain) waits and
// then sees the row on the active key; the update writes the value and its
// key id together (the framing CHECK keeps them in step).
type columnSQL struct {
	target  string
	pending string
	lock    string
	replace string
}

var (
	userEmailSQL = columnSQL{
		target:  TargetUserEmail,
		pending: "SELECT id::text AS id, user_id::text AS user_id FROM reseal_pending_users(?::smallint, ?::uuid, ?)",
		lock:    "SELECT email_encrypted AS value FROM users WHERE id = ? FOR UPDATE",
		replace: "UPDATE users SET email_encrypted = ?, email_key_id = ? WHERE id = ?",
	}
	taxpayerRFCSQL = columnSQL{
		target:  TargetTaxpayerRFC,
		pending: "SELECT id::text AS id, user_id::text AS user_id FROM reseal_pending_taxpayers(?::smallint, ?::uuid, ?)",
		lock:    "SELECT rfc_encrypted AS value FROM taxpayers WHERE id = ? FOR UPDATE",
		replace: "UPDATE taxpayers SET rfc_encrypted = ?, rfc_key_id = ? WHERE id = ?",
	}
	deviceTokenSQL = columnSQL{
		target:  TargetDeviceToken,
		pending: "SELECT id::text AS id, user_id::text AS user_id FROM reseal_pending_devices(?::smallint, ?::uuid, ?)",
		lock:    "SELECT token_ciphertext AS value FROM devices WHERE id = ? FOR UPDATE",
		replace: "UPDATE devices SET token_ciphertext = ?, token_key_id = ? WHERE id = ?",
	}
)

// Store implements usecase.Store over one sealed column.
type Store struct {
	sql columnSQL
}

var _ usecase.Store = (*Store)(nil)

// NewUserEmailStore is the store of users.email_encrypted.
func NewUserEmailStore() *Store { return &Store{sql: userEmailSQL} }

// NewTaxpayerRFCStore is the store of taxpayers.rfc_encrypted (archived
// taxpayers included).
func NewTaxpayerRFCStore() *Store { return &Store{sql: taxpayerRFCSQL} }

// NewDeviceTokenStore is the store of devices.token_ciphertext.
func NewDeviceTokenStore() *Store { return &Store{sql: deviceTokenSQL} }

// bytea passes a []byte as one bytea parameter: GORM expands a bare []byte
// in Raw/Exec into one placeholder per byte.
type bytea []byte

// Value implements driver.Valuer.
func (b bytea) Value() (driver.Value, error) { return []byte(b), nil }

// refRow is one reseal_pending_* row.
type refRow struct {
	ID     string `gorm:"column:id"`
	UserID string `gorm:"column:user_id"`
}

// Pending implements usecase.Store.
func (s *Store) Pending(ctx context.Context, active crypto.KeyID, after string, limit int) ([]usecase.Ref, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var cursor any
	if after != "" {
		cursor = after
	}
	var rows []refRow
	if err := db.Raw(s.sql.pending, int16(active), cursor, limit).Scan(&rows).Error; err != nil { // #nosec G115 -- crypto.MaxKeyID fits int16
		return nil, fmt.Errorf("keyrotation infra: list %s: %w", s.sql.target, err)
	}
	refs := make([]usecase.Ref, len(rows))
	for i, r := range rows {
		refs[i] = usecase.Ref{ID: r.ID, UserID: r.UserID}
	}
	return refs, nil
}

// valueRow is the locked sealed value.
type valueRow struct {
	Value []byte `gorm:"column:value"`
}

// Lock implements usecase.Store.
func (s *Store) Lock(ctx context.Context, ref usecase.Ref) ([]byte, bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, false, err
	}
	var rows []valueRow
	if err := db.Raw(s.sql.lock, ref.ID).Scan(&rows).Error; err != nil {
		return nil, false, fmt.Errorf("keyrotation infra: lock %s row: %w", s.sql.target, err)
	}
	if len(rows) == 0 {
		return nil, false, nil
	}
	return rows[0].Value, true, nil
}

// Replace implements usecase.Store.
func (s *Store) Replace(ctx context.Context, ref usecase.Ref, ciphertext []byte, keyID crypto.KeyID) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	res := db.Exec(s.sql.replace, bytea(ciphertext), int16(keyID), ref.ID) // #nosec G115 -- crypto.MaxKeyID fits int16
	if res.Error != nil {
		return fmt.Errorf("keyrotation infra: rewrite %s row: %w", s.sql.target, res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("%w: %s", ErrRowNotUpdated, s.sql.target)
	}
	return nil
}

// UsageReader implements usecase.UsageReader over key_usage().
type UsageReader struct{}

var _ usecase.UsageReader = UsageReader{}

// NewUsageReader returns the adapter.
func NewUsageReader() UsageReader { return UsageReader{} }

// usageSQL reads every target's rows per key id or pepper version.
const usageSQL = "SELECT target, version, row_count FROM key_usage()"

type usageRow struct {
	Target   string `gorm:"column:target"`
	Version  int    `gorm:"column:version"`
	RowCount int64  `gorm:"column:row_count"`
}

// KeyUsage implements usecase.UsageReader, normally in a System unit of work.
func (UsageReader) KeyUsage(ctx context.Context) ([]usecase.Usage, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []usageRow
	if err := db.Raw(usageSQL).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("keyrotation infra: key usage: %w", err)
	}
	out := make([]usecase.Usage, len(rows))
	for i, r := range rows {
		out[i] = usecase.Usage{Target: r.Target, Version: r.Version, Rows: r.RowCount}
	}
	return out, nil
}
