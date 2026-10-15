// Package store is the reminders module's GORM adapter of the devices and
// reminder_settings tables (migration 20261021101000, M9-02). It holds no
// pool: every method runs on the unit of work in ctx (cc_app, the caller's
// app.current_user_id set per transaction), so RLS scopes every statement
// to the caller. Devices are written only through the claim_push_token
// definer function; the token columns are never read (cc_app has no grant
// on them).
package store

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// SQLSTATEs of claim_push_token's documented races: a concurrent claim of
// the same installation by another user (serialization_failure) or of the
// same token (unique_violation). Both mean "retry".
const (
	sqlstateSerializationFailure = "40001"
	sqlstateUniqueViolation      = "23505"
)

const (
	claimSQL = `SELECT claim_push_token(CAST(? AS uuid), CAST(? AS uuid), CAST(? AS text), CAST(? AS bytea),
       CAST(? AS smallint), CAST(? AS bytea), CAST(? AS text)) AS created`
	// The device read and delete also name the caller (defense in depth,
	// like every user-scoped query): RLS already limits them to the unit of
	// work's user, and an installation id alone never reaches another user's
	// row should a policy ever be loosened.
	deviceSQL = `SELECT id::text AS id, platform, app_version, created_at, last_seen_at
FROM devices WHERE id = CAST(? AS uuid) AND user_id = (SELECT app_user_id())`
	deleteDeviceSQL = `DELETE FROM devices WHERE id = CAST(? AS uuid) AND user_id = (SELECT app_user_id())`
	settingsSQL     = `SELECT enabled, early_day, on_due_date, to_char(send_time, 'HH24:MI') AS send_time, show_amounts
FROM reminder_settings WHERE user_id = CAST(? AS uuid)`
	saveSettingsSQL = `INSERT INTO reminder_settings (user_id, enabled, early_day, on_due_date, send_time, show_amounts)
VALUES (CAST(? AS uuid), ?, ?, ?, CAST(? AS time), ?)
ON CONFLICT (user_id) DO UPDATE SET enabled = EXCLUDED.enabled, early_day = EXCLUDED.early_day,
  on_due_date = EXCLUDED.on_due_date, send_time = EXCLUDED.send_time, show_amounts = EXCLUDED.show_amounts`
)

// Devices implements usecase.DeviceStore.
type Devices struct{}

// Settings implements usecase.SettingsStore.
type Settings struct{}

var (
	_ usecase.DeviceStore   = Devices{}
	_ usecase.SettingsStore = Settings{}
)

// NewDevices returns the devices store.
func NewDevices() Devices { return Devices{} }

// NewSettings returns the settings store.
func NewSettings() Settings { return Settings{} }

type claimRow struct {
	Created bool `gorm:"column:created"`
}

// Claim implements usecase.DeviceStore.
func (Devices) Claim(ctx context.Context, userID string, d usecase.SealedDevice) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var row claimRow
	err = db.Raw(claimSQL, userID, d.InstallationID, string(d.Platform), bytea(d.Token.Ciphertext),
		d.Token.KeyID, bytea(d.Token.HMAC), nullable(d.AppVersion)).Scan(&row).Error
	if err != nil {
		return false, translateClaim(err)
	}
	return row.Created, nil
}

type deviceRow struct {
	ID         string    `gorm:"column:id"`
	Platform   string    `gorm:"column:platform"`
	AppVersion *string   `gorm:"column:app_version"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	LastSeenAt time.Time `gorm:"column:last_seen_at"`
}

// Device implements usecase.DeviceStore.
func (Devices) Device(ctx context.Context, installationID string) (domain.Device, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return domain.Device{}, err
	}
	var rows []deviceRow
	if err := db.Raw(deviceSQL, installationID).Scan(&rows).Error; err != nil {
		return domain.Device{}, fmt.Errorf("reminders store: device: %w", err)
	}
	if len(rows) == 0 {
		return domain.Device{}, usecase.ErrDeviceNotFound
	}
	r := rows[0]
	platform, err := domain.ParsePlatform(r.Platform)
	if err != nil {
		return domain.Device{}, fmt.Errorf("reminders store: device: %w", err)
	}
	device := domain.Device{ID: r.ID, Platform: platform, CreatedAt: r.CreatedAt.UTC(), LastSeenAt: r.LastSeenAt.UTC()}
	if r.AppVersion != nil {
		device.AppVersion = *r.AppVersion
	}
	return device, nil
}

// DeleteDevice implements usecase.DeviceStore.
func (Devices) DeleteDevice(ctx context.Context, installationID string) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	res := db.Exec(deleteDeviceSQL, installationID)
	if res.Error != nil {
		return false, fmt.Errorf("reminders store: delete device: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

type settingsRow struct {
	Enabled     bool   `gorm:"column:enabled"`
	EarlyDay    *int   `gorm:"column:early_day"`
	OnDueDate   bool   `gorm:"column:on_due_date"`
	SendTime    string `gorm:"column:send_time"`
	ShowAmounts bool   `gorm:"column:show_amounts"`
}

// Settings implements usecase.SettingsStore.
func (Settings) Settings(ctx context.Context, userID string) (domain.Settings, bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return domain.Settings{}, false, err
	}
	var rows []settingsRow
	if err := db.Raw(settingsSQL, userID).Scan(&rows).Error; err != nil {
		return domain.Settings{}, false, fmt.Errorf("reminders store: settings: %w", err)
	}
	if len(rows) == 0 {
		return domain.Settings{}, false, nil
	}
	s, err := toSettings(rows[0])
	if err != nil {
		return domain.Settings{}, false, fmt.Errorf("reminders store: stored settings: %w", err)
	}
	return s, true, nil
}

func toSettings(r settingsRow) (domain.Settings, error) {
	sendTime, err := domain.ParseSendTime(r.SendTime)
	if err != nil {
		return domain.Settings{}, err
	}
	earlyDay := domain.NoEarlyReminder
	if r.EarlyDay != nil {
		earlyDay = *r.EarlyDay
	}
	return domain.NewSettings(r.Enabled, earlyDay, r.OnDueDate, sendTime, r.ShowAmounts)
}

// SaveSettings implements usecase.SettingsStore.
func (Settings) SaveSettings(ctx context.Context, userID string, s domain.Settings) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	var earlyDay *int
	if s.HasEarlyReminder() {
		earlyDay = &s.EarlyDay
	}
	err = db.Exec(saveSettingsSQL, userID, s.Enabled, earlyDay, s.OnDueDate, s.SendTime.String(), s.ShowAmounts).Error
	if err != nil {
		return fmt.Errorf("reminders store: save settings: %w", err)
	}
	return nil
}

// translateClaim maps claim_push_token's documented races to
// usecase.ErrClaimContention, keeping the driver error (never the token:
// the statement binds only the ciphertext and the HMAC).
func translateClaim(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
		(pgErr.Code == sqlstateSerializationFailure || pgErr.Code == sqlstateUniqueViolation) {
		return fmt.Errorf("%w (reminders store: claim: %w)", usecase.ErrClaimContention, err)
	}
	return fmt.Errorf("reminders store: claim: %w", err)
}

// nullable is NULL for "".
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// bytea passes a []byte as one bytea parameter: GORM expands a bare []byte
// argument of Raw/Exec into one placeholder per byte.
type bytea []byte

// Value implements driver.Valuer.
func (b bytea) Value() (driver.Value, error) { return []byte(b), nil }
