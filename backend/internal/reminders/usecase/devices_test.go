package usecase_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

type devicesFixture struct {
	tx     *fakeTx
	store  *fakeDevices
	sealer *fakeSealer
	uc     *usecase.Devices
}

func newDevicesFixture(t *testing.T) *devicesFixture {
	t.Helper()
	f := &devicesFixture{tx: &fakeTx{}, store: newFakeDevices(), sealer: &fakeSealer{}}
	uc, err := usecase.NewDevices(usecase.DevicesDeps{Tx: f.tx, Store: f.store, Sealer: f.sealer})
	require.NoError(t, err)
	f.uc = uc
	return f
}

func registration(t *testing.T) usecase.Registration {
	t.Helper()
	version := appVersion
	return usecase.Registration{InstallationID: deviceA, Platform: "ios", Token: pushToken(t), AppVersion: &version}
}

func TestNewDevices_needsEveryPort(t *testing.T) {
	for name, deps := range map[string]usecase.DevicesDeps{
		"Tx":     {Store: newFakeDevices(), Sealer: &fakeSealer{}},
		"Store":  {Tx: &fakeTx{}, Sealer: &fakeSealer{}},
		"Sealer": {Tx: &fakeTx{}, Store: newFakeDevices()},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := usecase.NewDevices(deps)

			require.ErrorIs(t, err, usecase.ErrMissingDependency)
		})
	}
}

// MODULE-reminders §2.3: the token is sealed bound to (installation, user),
// claimed in one read-write unit of work as the caller, and the answer is
// the caller's row (never the token).
func TestRegister_sealsAndClaimsAsTheCaller(t *testing.T) {
	for _, created := range []bool{true, false} {
		t.Run(fmt.Sprintf("created=%v", created), func(t *testing.T) {
			f := newDevicesFixture(t)
			f.store.created = created
			reg := registration(t)

			got, err := f.uc.Register(t.Context(), userA, reg)

			require.NoError(t, err)
			assert.Equal(t, created, got.Created, "201 when created or claimed, 200 when refreshed")
			assert.Equal(t, domain.Device{
				ID: deviceA, Platform: domain.PlatformIOS, AppVersion: appVersion, CreatedAt: registeredAt, LastSeenAt: registeredAt,
			}, got.Device)
			require.Len(t, f.sealer.calls, 1)
			assert.Equal(t, sealCall{token: reg.Token, binding: usecase.TokenBinding{InstallationID: deviceA, UserID: userA}}, f.sealer.calls[0])
			require.Len(t, f.store.claims, 1)
			claim := f.store.claims[0]
			assert.Equal(t, userA, claim.txUser, "the claim runs as the caller (SET LOCAL)")
			assert.Equal(t, userA, claim.userID)
			assert.Equal(t, usecase.SealedDevice{
				InstallationID: deviceA, Platform: domain.PlatformIOS, AppVersion: appVersion,
				Token: usecase.SealedToken{Ciphertext: []byte("sealed"), KeyID: 1, HMAC: []byte("hmac")},
			}, claim.device)
			assert.Equal(t, []string{"rw"}, f.tx.modes)
		})
	}
}

func TestRegister_withoutAppVersion(t *testing.T) {
	f := newDevicesFixture(t)
	reg := registration(t)
	reg.AppVersion = nil

	got, err := f.uc.Register(t.Context(), userA, reg)

	require.NoError(t, err)
	assert.Empty(t, got.Device.AppVersion)
	assert.Empty(t, f.store.claims[0].device.AppVersion)
}

// The installation id is canonical (lower-case) before it is bound into
// the ciphertext or claimed.
func TestRegister_canonicalizesTheInstallationID(t *testing.T) {
	f := newDevicesFixture(t)
	reg := registration(t)
	reg.InstallationID = strings.ToUpper(deviceA)

	_, err := f.uc.Register(t.Context(), userA, reg)

	require.NoError(t, err)
	assert.Equal(t, deviceA, f.store.claims[0].device.InstallationID)
	assert.Equal(t, deviceA, f.sealer.calls[0].binding.InstallationID)
}

// The router validates the body against the contract first; these are the
// use case's backstops. None of them echoes the token.
func TestRegister_refusesInvalidInput(t *testing.T) {
	tooLong := strings.Repeat("v", domain.MaxAppVersionLength+1)
	tests := []struct {
		name   string
		mutate func(*usecase.Registration)
		field  string
		detail string
	}{
		{"installation id not a uuid", func(r *usecase.Registration) { r.InstallationID = "nope" }, "id", "format"},
		{"unknown platform", func(r *usecase.Registration) { r.Platform = "web" }, "platform", "enum"},
		{"token with a space", func(r *usecase.Registration) { r.Token += " x" }, "push_token", "pattern"},
		{"empty token", func(r *usecase.Registration) { r.Token = "" }, "push_token", "pattern"},
		{"app version too long", func(r *usecase.Registration) { r.AppVersion = &tooLong }, "app_version", "pattern"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newDevicesFixture(t)
			reg := registration(t)
			tc.mutate(&reg)

			_, err := f.uc.Register(t.Context(), userA, reg)

			requireAppError(t, err, apperr.CodeValidationFailed, tc.field, tc.detail)
			if reg.Token != "" {
				assert.NotContains(t, err.Error(), reg.Token, "the token never reaches an error")
			}
			assert.Empty(t, f.store.claims, "nothing is written")
			assert.Empty(t, f.tx.modes)
		})
	}
}

// claim_push_token's documented races (40001 or 23505 from a concurrent
// claim) retry the whole unit of work, up to three attempts.
func TestRegister_retriesAClaimRace(t *testing.T) {
	f := newDevicesFixture(t)
	f.store.claimErr = []error{usecase.ErrClaimContention, usecase.ErrClaimContention}

	_, err := f.uc.Register(t.Context(), userA, registration(t))

	require.NoError(t, err)
	assert.Len(t, f.store.claims, 3)
	assert.Equal(t, []string{"rw", "rw", "rw"}, f.tx.modes, "each attempt is its own transaction")
	assert.Len(t, f.sealer.calls, 1, "sealed once")
}

func TestRegister_givesUpAfterThreeRaces(t *testing.T) {
	f := newDevicesFixture(t)
	f.store.claimErr = []error{usecase.ErrClaimContention, usecase.ErrClaimContention, usecase.ErrClaimContention}

	_, err := f.uc.Register(t.Context(), userA, registration(t))

	require.ErrorIs(t, err, usecase.ErrClaimContention)
	assert.Len(t, f.store.claims, 3)
}

func TestRegister_failures(t *testing.T) {
	tests := map[string]func(*devicesFixture){
		"sealer":  func(f *devicesFixture) { f.sealer.err = errPortDown },
		"claim":   func(f *devicesFixture) { f.store.claimErr = []error{errPortDown} },
		"re-read": func(f *devicesFixture) { f.store.readErr = errPortDown },
	}
	for name, breakIt := range tests {
		t.Run(name, func(t *testing.T) {
			f := newDevicesFixture(t)
			breakIt(f)
			reg := registration(t)

			_, err := f.uc.Register(t.Context(), userA, reg)

			require.ErrorIs(t, err, errPortDown)
			assert.NotContains(t, err.Error(), reg.Token)
			assert.LessOrEqual(t, len(f.store.claims), 1, "an infrastructure failure is not retried")
		})
	}
}

// §2.3, §8: a Registration carries the token, so it never prints it.
func TestRegistration_neverPrintsTheToken(t *testing.T) {
	reg := registration(t)
	var logs strings.Builder
	slog.New(slog.NewJSONHandler(&logs, nil)).Info("x", slog.Any("registration", reg))

	printed := fmt.Sprintf("%v %+v %#v %s", reg, reg, reg, reg)

	assert.NotContains(t, printed, reg.Token)
	assert.NotContains(t, logs.String(), reg.Token)
	assert.Contains(t, printed, deviceA, "the non-secret fields still help support")
}

func TestUnregister_deletesTheCallersDevice(t *testing.T) {
	f := newDevicesFixture(t)
	_, err := f.uc.Register(t.Context(), userA, registration(t))
	require.NoError(t, err)

	err = f.uc.Unregister(t.Context(), userA, strings.ToUpper(deviceA))

	require.NoError(t, err)
	assert.Equal(t, []string{deviceA}, f.store.deleted)
	assert.Equal(t, []string{userA, userA}, f.tx.users)
	assert.Equal(t, "rw", f.tx.modes[1])
}

// §2.3, US-R10: another user's, an unknown or a malformed installation id
// is the hidden 404 (the same as the module OFF); so is a second call.
func TestUnregister_unknownIsTheHidden404(t *testing.T) {
	for name, id := range map[string]string{"unknown (or another user's: RLS)": deviceA, "malformed": "nope"} {
		t.Run(name, func(t *testing.T) {
			f := newDevicesFixture(t)

			err := f.uc.Unregister(t.Context(), userA, id)

			requireAppError(t, err, apperr.CodeNotFound, "", "")
		})
	}
}

func TestUnregister_storeFailure(t *testing.T) {
	f := newDevicesFixture(t)
	f.store.delErr = errPortDown

	err := f.uc.Unregister(t.Context(), userA, deviceA)

	require.ErrorIs(t, err, errPortDown)
}
