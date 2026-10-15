package usecase_test

import (
	"context"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const (
	userA      = "0199d4a2-0000-7000-8000-00000000d00a"
	deviceA    = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23"
	appVersion = "1.2.0"
)

var registeredAt = time.Date(2026, time.October, 15, 15, 2, 0, 0, time.UTC)

// pushToken is a token-shaped value generated per test: never a literal.
func pushToken(t *testing.T) string {
	t.Helper()
	return "fcm:" + rand.Text() + rand.Text()
}

// userKey carries the fake unit of work's user, standing in for RLS.
type userKey struct{}

func userOf(ctx context.Context) string {
	user, _ := ctx.Value(userKey{}).(string)
	return user
}

// fakeTx runs work inline, records the modes used and stands in for the
// SET LOCAL of the user.
type fakeTx struct {
	modes []string
	users []string
}

func (f *fakeTx) ReadWrite(ctx context.Context, userID string, work tx.Work) error {
	f.modes, f.users = append(f.modes, "rw"), append(f.users, userID)
	return work(context.WithValue(ctx, userKey{}, userID))
}

func (f *fakeTx) ReadOnly(ctx context.Context, userID string, work tx.Work) error {
	f.modes, f.users = append(f.modes, "ro"), append(f.users, userID)
	return work(context.WithValue(ctx, userKey{}, userID))
}

// sealCall is one Seal: the token and the row it was bound to.
type sealCall struct {
	token   string
	binding usecase.TokenBinding
}

// fakeSealer seals by tagging; it records what it was asked.
type fakeSealer struct {
	calls []sealCall
	err   error
}

func (f *fakeSealer) Seal(token domain.PushToken, b usecase.TokenBinding) (usecase.SealedToken, error) {
	f.calls = append(f.calls, sealCall{token: token.Reveal(), binding: b})
	if f.err != nil {
		return usecase.SealedToken{}, f.err
	}
	return usecase.SealedToken{Ciphertext: []byte("sealed"), KeyID: 1, HMAC: []byte("hmac")}, nil
}

// claimCall is one Claim as the store saw it, with the unit of work's user.
type claimCall struct {
	txUser string
	userID string
	device usecase.SealedDevice
}

// fakeDevices is the devices table of one user: claims, reads, deletes.
type fakeDevices struct {
	claims   []claimCall
	created  bool
	claimErr []error // returned by successive claims; nil when exhausted
	rows     map[string]domain.Device
	readErr  error
	deleted  []string
	delErr   error
}

func newFakeDevices() *fakeDevices { return &fakeDevices{rows: map[string]domain.Device{}} }

func (f *fakeDevices) Claim(ctx context.Context, userID string, d usecase.SealedDevice) (bool, error) {
	f.claims = append(f.claims, claimCall{txUser: userOf(ctx), userID: userID, device: d})
	if len(f.claimErr) > 0 {
		err := f.claimErr[0]
		f.claimErr = f.claimErr[1:]
		if err != nil {
			return false, err
		}
	}
	f.rows[d.InstallationID] = domain.Device{
		ID: d.InstallationID, Platform: d.Platform, AppVersion: d.AppVersion, CreatedAt: registeredAt, LastSeenAt: registeredAt,
	}
	return f.created, nil
}

func (f *fakeDevices) Device(_ context.Context, id string) (domain.Device, error) {
	if f.readErr != nil {
		return domain.Device{}, f.readErr
	}
	d, ok := f.rows[id]
	if !ok {
		return domain.Device{}, usecase.ErrDeviceNotFound
	}
	return d, nil
}

func (f *fakeDevices) DeleteDevice(_ context.Context, id string) (bool, error) {
	if f.delErr != nil {
		return false, f.delErr
	}
	_, ok := f.rows[id]
	delete(f.rows, id)
	if ok {
		f.deleted = append(f.deleted, id)
	}
	return ok, nil
}

// fakeSettings is the reminder_settings table of one user.
type fakeSettings struct {
	saved    *domain.Settings
	savedFor string
	readFor  string
	txUser   string
	readErr  error
	saveErr  error
}

func (f *fakeSettings) Settings(ctx context.Context, userID string) (domain.Settings, bool, error) {
	f.txUser, f.readFor = userOf(ctx), userID
	if f.readErr != nil {
		return domain.Settings{}, false, f.readErr
	}
	if f.saved == nil {
		return domain.Settings{}, false, nil
	}
	return *f.saved, true, nil
}

func (f *fakeSettings) SaveSettings(ctx context.Context, userID string, s domain.Settings) error {
	f.txUser = userOf(ctx)
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved, f.savedFor = &s, userID
	return nil
}

// requireAppError asserts err is the client-ready error with code and, when
// field is set, that one field error.
func requireAppError(t *testing.T, err error, code apperr.Code, field, detail string) {
	t.Helper()
	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "a client-ready error, got %v", err)
	require.Equal(t, code, appErr.Code)
	if field != "" {
		require.Len(t, appErr.Fields, 1)
		require.Equal(t, field, appErr.Fields[0].Field)
		require.Equal(t, detail, appErr.Fields[0].Code)
	}
}
