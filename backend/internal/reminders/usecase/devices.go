package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

// claimAttempts bounds the retries of a claim that lost a race in
// claim_push_token (40001 or 23505, migration 20261021101000).
const claimAttempts = 3

// DevicesDeps are the ports of Devices. Store must run on Tx's units of
// work.
type DevicesDeps struct {
	Tx     tx.Manager
	Store  DeviceStore
	Sealer TokenSealer
}

// Devices answers registerDevice and unregisterDevice (MODULE-reminders
// §2.3). Both are account preferences, not records: no access policy, no
// 402 for a read-only user (§4). It is safe for concurrent use.
type Devices struct {
	deps DevicesDeps
}

// NewDevices returns the use case, or ErrMissingDependency.
func NewDevices(deps DevicesDeps) (*Devices, error) {
	switch {
	case deps.Tx == nil:
		return nil, fmt.Errorf("%w: DevicesDeps.Tx", ErrMissingDependency)
	case deps.Store == nil:
		return nil, fmt.Errorf("%w: DevicesDeps.Store", ErrMissingDependency)
	case deps.Sealer == nil:
		return nil, fmt.Errorf("%w: DevicesDeps.Sealer", ErrMissingDependency)
	}
	return &Devices{deps: deps}, nil
}

// Registration is registerDevice's body. It carries the push token, so it
// never prints it: every fmt verb and slog render the token redacted.
type Registration struct {
	InstallationID string
	Platform       string
	Token          string
	// AppVersion is nil when the app sent none.
	AppVersion *string
}

// Format renders the registration with the token redacted.
func (r Registration) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, fmt.Sprintf("{InstallationID:%s Platform:%s Token:REDACTED}", r.InstallationID, r.Platform))
}

// LogValue renders the registration in slog with the token redacted.
func (r Registration) LogValue() slog.Value {
	return slog.GroupValue(slog.String("installation_id", r.InstallationID), slog.String("platform", r.Platform))
}

// Registered is registerDevice's answer: the caller's row and whether it
// was created (201) or refreshed (200).
type Registered struct {
	Device  domain.Device
	Created bool
}

// validRegistration is a Registration past every check.
type validRegistration struct {
	installationID string
	platform       domain.Platform
	token          domain.PushToken
	appVersion     string
}

// Register registers the caller's installation (§2.3): the token is sealed
// bound to (installation, caller), then claim_push_token moves any other
// row holding the installation or the token to the caller and the row is
// read back, in one read-write unit of work as the caller. A claim that
// loses a race is retried in a new unit of work. Whether the token had
// another owner never shows (no oracle).
func (d *Devices) Register(ctx context.Context, userID string, r Registration) (Registered, error) {
	in, err := validateRegistration(r)
	if err != nil {
		return Registered{}, err
	}
	sealed, err := d.deps.Sealer.Seal(in.token, TokenBinding{InstallationID: in.installationID, UserID: userID})
	if err != nil {
		return Registered{}, fmt.Errorf("reminders: register device: seal token: %w", err)
	}
	device := SealedDevice{InstallationID: in.installationID, Platform: in.platform, Token: sealed, AppVersion: in.appVersion}
	for attempt := 1; ; attempt++ {
		out, err := d.claim(ctx, userID, device)
		if !errors.Is(err, ErrClaimContention) || attempt == claimAttempts {
			return out, err
		}
	}
}

func (d *Devices) claim(ctx context.Context, userID string, device SealedDevice) (Registered, error) {
	var out Registered
	err := d.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		created, err := d.deps.Store.Claim(ctx, userID, device)
		if err != nil {
			return err
		}
		row, err := d.deps.Store.Device(ctx, device.InstallationID)
		if err != nil {
			return err
		}
		out = Registered{Device: row, Created: created}
		return nil
	})
	if err != nil {
		return Registered{}, fmt.Errorf("reminders: register device: %w", err)
	}
	return out, nil
}

// validateRegistration is the backstop of the contract's DeviceCreate
// schema (the router answers first). The errors never quote the token.
func validateRegistration(r Registration) (validRegistration, error) {
	id, err := uuid.Parse(r.InstallationID)
	if err != nil {
		return validRegistration{}, invalidField(err, fieldInstallationID, detailFormat, msgInvalidFormat)
	}
	platform, err := domain.ParsePlatform(r.Platform)
	if err != nil {
		return validRegistration{}, invalidField(err, fieldPlatform, detailEnum, msgInvalidPlatform)
	}
	token, err := domain.NewPushToken(r.Token)
	if err != nil {
		return validRegistration{}, invalidField(err, fieldPushToken, detailPattern, msgInvalidFormat)
	}
	var version string
	if r.AppVersion != nil {
		if version, err = domain.ParseAppVersion(*r.AppVersion); err != nil {
			return validRegistration{}, invalidField(err, fieldAppVersion, detailPattern, msgInvalidFormat)
		}
	}
	return validRegistration{installationID: id.String(), platform: platform, token: token, appVersion: version}, nil
}

// Unregister removes the caller's installation (§2.3): another user's, an
// unknown or a malformed id is the hidden 404, so is a second call (the
// app treats 404 as done).
func (d *Devices) Unregister(ctx context.Context, userID, installationID string) error {
	id, err := uuid.Parse(installationID)
	if err != nil {
		return notFound(err)
	}
	var deleted bool
	err = d.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		deleted, err = d.deps.Store.DeleteDevice(ctx, id.String())
		return err
	})
	if err != nil {
		return fmt.Errorf("reminders: unregister device: %w", err)
	}
	if !deleted {
		return notFound(nil)
	}
	return nil
}
