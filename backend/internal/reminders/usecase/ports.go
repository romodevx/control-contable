package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

// DeviceStore is the devices table (M9-02) under the caller's RLS. Every
// method runs on the unit of work in ctx.
type DeviceStore interface {
	// Claim registers the installation for userID through the
	// claim_push_token definer function: it removes every other row holding
	// the installation id or the token HMAC (whoever owns it) and upserts
	// the caller's. created is true for a new or claimed row (201), false
	// for a refresh of the caller's own (200). A concurrent claim that wins
	// the race is ErrClaimContention: retry in a new unit of work.
	Claim(ctx context.Context, userID string, d SealedDevice) (created bool, err error)
	// Device reads one of the caller's devices; another user's or an
	// unknown id is ErrDeviceNotFound.
	Device(ctx context.Context, installationID string) (domain.Device, error)
	// DeleteDevice removes one of the caller's devices; deleted is false
	// when none matched (another user's or unknown).
	DeleteDevice(ctx context.Context, installationID string) (deleted bool, err error)
}

// SettingsStore is the reminder_settings table (M9-02) under the caller's
// RLS.
type SettingsStore interface {
	// Settings reads userID's saved settings; found is false when the user
	// never saved.
	Settings(ctx context.Context, userID string) (s domain.Settings, found bool, err error)
	// SaveSettings upserts the caller's whole settings row.
	SaveSettings(ctx context.Context, userID string, s domain.Settings) error
}

// TokenSealer seals a push token at rest (§2.3, OQ-R7): AES-256-GCM bound
// to its row, plus the keyed HMAC the claim looks it up by.
type TokenSealer interface {
	Seal(token domain.PushToken, b TokenBinding) (SealedToken, error)
}

// TokenBinding is the row a sealed token belongs to: a ciphertext moved to
// another installation or user does not open.
type TokenBinding struct {
	InstallationID string
	UserID         string
}

// SealedToken is a push token as stored: never the token itself.
type SealedToken struct {
	// Ciphertext is the crypto.Keyring frame (key id, nonce, ciphertext,
	// tag).
	Ciphertext []byte
	// KeyID repeats the frame's key id (devices.token_key_id).
	KeyID int16
	// HMAC is the keyed token_hmac, the claim's lookup key.
	HMAC []byte
}

// SealedDevice is what Claim writes.
type SealedDevice struct {
	InstallationID string
	Platform       domain.Platform
	Token          SealedToken
	// AppVersion is "" when the app sent none (NULL).
	AppVersion string
}
