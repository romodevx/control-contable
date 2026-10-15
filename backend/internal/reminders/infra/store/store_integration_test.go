//go:build integration

package store_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/store"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	"github.com/romodevx/control-contable/backend/migrations"
)

func TestMain(m *testing.M) {
	os.Exit(pgtest.Main(m, pgtest.WithMigrator(migrations.Up)))
}

// The devices and reminder_settings stores over the real migrations, as
// cc_app under RLS, with the real token vault. Fixture ids are this suite's
// own; every test gets its own cloned database.
const (
	userA    = "0199d4a2-0000-7000-8000-0000000009aa"
	userB    = "0199d4a2-0000-7000-8000-0000000009bb"
	device1  = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f0901"
	device2  = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f0902"
	tokenKey = crypto.KeyID(1)
)

// registerSQL is sign-up as cc_app does it under the new user's context;
// every stored value derives from the id (no literal looks like a secret).
const registerSQL = `
WITH v(id, digest) AS (SELECT $1::uuid, sha256(convert_to($1::text, 'UTF8')))
INSERT INTO users (id, email_hash, email_encrypted, email_key_id, password_hash, name)
SELECT id, digest, '\x0001'::bytea || digest || digest, 1,
       '$argon2id$v=19$m=8,t=1,p=1$' || rtrim(encode(digest, 'base64'), '=') || '$' || rtrim(encode(digest, 'base64'), '='),
       'Fixture'
FROM v`

type env struct {
	db       *pgtest.Database
	app      *sql.DB
	txm      *pgdb.TxManager
	vault    *pushtoken.Vault
	devices  store.Devices
	settings store.Settings
}

func newEnv(t *testing.T) *env {
	t.Helper()
	db := pgtest.NewDatabase(t)
	e := &env{db: db, app: db.Connect(t, pgtest.RoleApp), devices: store.NewDevices(), settings: store.NewSettings()}
	var err error
	e.txm, err = pgdb.NewTxManager(e.app)
	require.NoError(t, err)
	for _, user := range []string{userA, userB} {
		require.NoError(t, pgtest.AsUser(t.Context(), e.app, user, func(tx *sql.Tx) error {
			_, err := tx.ExecContext(t.Context(), registerSQL, user)
			return err
		}))
	}
	key, pepper := make([]byte, crypto.KeySize), make([]byte, crypto.MinPepperSize)
	_, _ = rand.Read(key)
	_, _ = rand.Read(pepper)
	keyring, err := crypto.NewKeyring(tokenKey, map[crypto.KeyID][]byte{tokenKey: key})
	require.NoError(t, err)
	index, err := crypto.NewTokenIndex(pepper)
	require.NoError(t, err)
	e.vault = pushtoken.NewVault(keyring, index)
	return e
}

// newToken is a token-shaped value generated per call: never a literal.
func newToken(t *testing.T) domain.PushToken {
	t.Helper()
	tok, err := domain.NewPushToken("fcm:" + rand.Text() + rand.Text())
	require.NoError(t, err)
	return tok
}

func (e *env) sealed(t *testing.T, user, installation string, tok domain.PushToken, version string) usecase.SealedDevice {
	t.Helper()
	s, err := e.vault.Seal(tok, usecase.TokenBinding{InstallationID: installation, UserID: user})
	require.NoError(t, err)
	return usecase.SealedDevice{InstallationID: installation, Platform: domain.PlatformAndroid, Token: s, AppVersion: version}
}

func (e *env) claim(t *testing.T, user string, d usecase.SealedDevice) bool {
	t.Helper()
	var created bool
	require.NoError(t, e.txm.ReadWrite(t.Context(), user, func(ctx context.Context) error {
		var err error
		created, err = e.devices.Claim(ctx, user, d)
		return err
	}))
	return created
}

func (e *env) device(t *testing.T, user, id string) (domain.Device, error) {
	t.Helper()
	var d domain.Device
	err := e.txm.ReadOnly(t.Context(), user, func(ctx context.Context) error {
		var err error
		d, err = e.devices.Device(ctx, id)
		return err
	})
	return d, err
}

func (e *env) remove(t *testing.T, user, id string) bool {
	t.Helper()
	var deleted bool
	require.NoError(t, e.txm.ReadWrite(t.Context(), user, func(ctx context.Context) error {
		var err error
		deleted, err = e.devices.DeleteDevice(ctx, id)
		return err
	}))
	return deleted
}

// MODULE-reminders §2.3: a new installation is created (201), the same
// caller and installation again is a refresh (200) that moves last_seen_at
// and keeps created_at.
func TestDevices_claimCreatesThenRefreshes(t *testing.T) {
	e := newEnv(t)

	created := e.claim(t, userA, e.sealed(t, userA, device1, newToken(t), "1.2.0"))
	require.True(t, created)
	first, err := e.device(t, userA, device1)
	require.NoError(t, err)
	assert.Equal(t, domain.Device{
		ID: device1, Platform: domain.PlatformAndroid, AppVersion: "1.2.0", CreatedAt: first.CreatedAt, LastSeenAt: first.LastSeenAt,
	}, first)
	assert.WithinDuration(t, time.Now(), first.CreatedAt, time.Minute)

	refreshed := e.claim(t, userA, e.sealed(t, userA, device1, newToken(t), ""))

	assert.False(t, refreshed)
	second, err := e.device(t, userA, device1)
	require.NoError(t, err)
	assert.Empty(t, second.AppVersion, "an omitted version is NULL")
	assert.Equal(t, first.CreatedAt, second.CreatedAt)
	assert.False(t, second.LastSeenAt.Before(first.LastSeenAt))
}

// US-R6: B registering the token A's installation holds moves it to B (a
// 201 for B), and A's row is gone; B learns nothing about A.
func TestDevices_claimMovesAnotherUsersToken(t *testing.T) {
	e := newEnv(t)
	tok := newToken(t)
	e.claim(t, userA, e.sealed(t, userA, device1, tok, "1.2.0"))

	created := e.claim(t, userB, e.sealed(t, userB, device2, tok, "1.3.0"))

	assert.True(t, created)
	_, err := e.device(t, userA, device1)
	require.ErrorIs(t, err, usecase.ErrDeviceNotFound, "A lost the token")
	got, err := e.device(t, userB, device2)
	require.NoError(t, err)
	assert.Equal(t, "1.3.0", got.AppVersion)
}

// A shared phone changes hands: B registers A's installation id with a new
// token; the row becomes B's.
func TestDevices_claimMovesAnotherUsersInstallation(t *testing.T) {
	e := newEnv(t)
	e.claim(t, userA, e.sealed(t, userA, device1, newToken(t), ""))

	created := e.claim(t, userB, e.sealed(t, userB, device1, newToken(t), ""))

	assert.True(t, created)
	_, err := e.device(t, userA, device1)
	require.ErrorIs(t, err, usecase.ErrDeviceNotFound)
	_, err = e.device(t, userB, device1)
	require.NoError(t, err)
}

// US-R10: B cannot read or delete A's device (RLS: 404 upstream); A can,
// once.
func TestDevices_deleteIsOwnOnly(t *testing.T) {
	e := newEnv(t)
	e.claim(t, userA, e.sealed(t, userA, device1, newToken(t), ""))

	_, err := e.device(t, userB, device1)
	require.ErrorIs(t, err, usecase.ErrDeviceNotFound)
	assert.False(t, e.remove(t, userB, device1), "another user's device")
	_, err = e.device(t, userA, device1)
	require.NoError(t, err, "A's row is intact")

	assert.True(t, e.remove(t, userA, device1))
	assert.False(t, e.remove(t, userA, device1), "a second call matches nothing")
	assert.False(t, e.remove(t, userA, device2), "unknown")
}

// errRollback ends a unit of work that must leave nothing behind.
var errRollback = errors.New("rollback")

// Defense in depth: the device read and delete name the caller themselves
// (user_id = app_user_id()), so another user's installation stays out of
// reach even where RLS does not apply. Proved as the table owner with FORCE
// RLS lifted inside a unit of work that is rolled back: without the
// predicate, B would read and delete A's device here.
func TestDevices_readAndDeleteAreOwnOnlyWithoutRLS(t *testing.T) {
	e := newEnv(t)
	e.claim(t, userA, e.sealed(t, userA, device1, newToken(t), ""))
	owner, err := pgdb.NewTxManager(e.db.Connect(t, pgtest.RoleOwner))
	require.NoError(t, err)

	err = owner.ReadWrite(t.Context(), userB, func(ctx context.Context) error {
		conn, err := pgdb.Conn(ctx)
		require.NoError(t, err)
		require.NoError(t, conn.Exec("ALTER TABLE devices NO FORCE ROW LEVEL SECURITY").Error)

		_, err = e.devices.Device(ctx, device1)
		require.ErrorIs(t, err, usecase.ErrDeviceNotFound, "B reads A's device")
		deleted, err := e.devices.DeleteDevice(ctx, device1)
		require.NoError(t, err)
		assert.False(t, deleted, "B deletes A's device")

		_, err = e.devices.Device(ctx, device1)
		require.ErrorIs(t, err, usecase.ErrDeviceNotFound)
		return errRollback
	})

	require.ErrorIs(t, err, errRollback)
	_, err = e.device(t, userA, device1)
	require.NoError(t, err, "A's row is intact")
	require.NoError(t, pgtest.CheckRLSForced(t.Context(), e.db.Connect(t, pgtest.RoleOwner), pgtest.Table{Name: "devices", OwnerColumn: "user_id"}))
}

func (e *env) readSettings(t *testing.T, user string) (domain.Settings, bool) {
	t.Helper()
	return e.readSettingsOf(t, user, user)
}

// readSettingsOf reads owner's settings in caller's unit of work.
func (e *env) readSettingsOf(t *testing.T, caller, owner string) (domain.Settings, bool) {
	t.Helper()
	var (
		s     domain.Settings
		found bool
	)
	require.NoError(t, e.txm.ReadOnly(t.Context(), caller, func(ctx context.Context) error {
		var err error
		s, found, err = e.settings.Settings(ctx, owner)
		return err
	}))
	return s, found
}

func (e *env) saveSettings(t *testing.T, user string, s domain.Settings) {
	t.Helper()
	require.NoError(t, e.txm.ReadWrite(t.Context(), user, func(ctx context.Context) error {
		return e.settings.SaveSettings(ctx, user, s)
	}))
}

func settings(t *testing.T, enabled bool, earlyDay int, onDue bool, at string, show bool) domain.Settings {
	t.Helper()
	st, err := domain.ParseSendTime(at)
	require.NoError(t, err)
	s, err := domain.NewSettings(enabled, earlyDay, onDue, st, show)
	require.NoError(t, err)
	return s
}

// §2.2: no row until the first save; PUT upserts the whole object (last
// write wins), early_day none is NULL; B never sees A's row.
func TestSettings_upsertAndIsolation(t *testing.T) {
	e := newEnv(t)
	_, found := e.readSettings(t, userA)
	require.False(t, found)

	first := settings(t, true, 14, true, "08:30", true)
	e.saveSettings(t, userA, first)
	got, found := e.readSettings(t, userA)
	require.True(t, found)
	assert.Equal(t, first, got)

	second := settings(t, true, domain.NoEarlyReminder, true, "23:59", false)
	e.saveSettings(t, userA, second)
	got, _ = e.readSettings(t, userA)
	assert.Equal(t, second, got)

	_, found = e.readSettingsOf(t, userB, userA)
	assert.False(t, found, "RLS: B never reads A's settings")
}
