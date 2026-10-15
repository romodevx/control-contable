package remindershttp_test

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const installationID = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23"

var (
	errDevicesDown  = errors.New("fake: devices down")
	errSettingsDown = errors.New("fake: settings down")
	seenAt          = time.Date(2026, time.October, 15, 15, 2, 0, 0, time.UTC)
)

// pushToken is a token-shaped value generated per test: never a literal.
func pushToken(t *testing.T) string {
	t.Helper()
	return "fcm:" + rand.Text() + rand.Text()
}

type registerCall struct {
	userID string
	reg    usecase.Registration
}

type fakeDevices struct {
	registers   []registerCall
	unregisters []string
	out         usecase.Registered
	err         error
}

func (f *fakeDevices) Register(_ context.Context, userID string, r usecase.Registration) (usecase.Registered, error) {
	f.registers = append(f.registers, registerCall{userID: userID, reg: r})
	return f.out, f.err
}

func (f *fakeDevices) Unregister(_ context.Context, userID, id string) error {
	f.unregisters = append(f.unregisters, userID+"/"+id)
	return f.err
}

type fakeSettings struct {
	gets []string
	puts []usecase.SettingsInput
	out  domain.Settings
	err  error
}

func (f *fakeSettings) Get(_ context.Context, userID string) (domain.Settings, error) {
	f.gets = append(f.gets, userID)
	return f.out, f.err
}

func (f *fakeSettings) Put(_ context.Context, _ string, in usecase.SettingsInput) (domain.Settings, error) {
	f.puts = append(f.puts, in)
	return f.out, f.err
}

func deviceCreate(t *testing.T, token *string) gen.RegisterDeviceRequestObject {
	t.Helper()
	version := "1.2.0"
	return gen.RegisterDeviceRequestObject{Body: &gen.DeviceCreate{
		Id: uuid.MustParse(installationID), Platform: gen.Ios, PushToken: token, AppVersion: &version,
	}}
}

func registered(created bool, version string) usecase.Registered {
	return usecase.Registered{Created: created, Device: domain.Device{
		ID: installationID, Platform: domain.PlatformIOS, AppVersion: version, CreatedAt: seenAt, LastSeenAt: seenAt,
	}}
}

// visit writes a strict response object.
func visit(t *testing.T, resp interface {
	VisitRegisterDeviceResponse(http.ResponseWriter) error
}) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitRegisterDeviceResponse(rec))
	return rec
}

// MODULE-reminders §2.3: 201 for a new or claimed row, 200 for a refresh;
// the body is the Device (never the token); the use case gets the
// caller and the body as sent.
func TestRegisterDevice(t *testing.T) {
	for _, tc := range []struct {
		created bool
		status  int
	}{{true, http.StatusCreated}, {false, http.StatusOK}} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			devices := &fakeDevices{out: registered(tc.created, "1.2.0")}
			token := pushToken(t)

			resp, err := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices}).
				RegisterDevice(asCaller(t), deviceCreate(t, &token))
			require.NoError(t, err)
			rec := visit(t, resp)

			assert.Equal(t, tc.status, rec.Code)
			assert.JSONEq(t, `{"id":"`+installationID+`","platform":"ios","app_version":"1.2.0",
				"created_at":"2026-10-15T15:02:00Z","last_seen_at":"2026-10-15T15:02:00Z"}`, rec.Body.String())
			assert.NotContains(t, rec.Body.String(), token)
			version := "1.2.0"
			assert.Equal(t, []registerCall{{userID: callerID, reg: usecase.Registration{
				InstallationID: installationID, Platform: "ios", Token: token, AppVersion: &version,
			}}}, devices.registers)
		})
	}
}

func TestRegisterDevice_nullAppVersion(t *testing.T) {
	devices := &fakeDevices{out: registered(true, "")}
	token := pushToken(t)
	req := deviceCreate(t, &token)
	req.Body.AppVersion = nil

	resp, err := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices}).RegisterDevice(asCaller(t), req)
	require.NoError(t, err)

	var body map[string]any
	require.NoError(t, json.Unmarshal(visit(t, resp).Body.Bytes(), &body))
	assert.Contains(t, body, "app_version")
	assert.Nil(t, body["app_version"], "app_version is required and nullable")
	assert.Nil(t, devices.registers[0].reg.AppVersion)
}

// PR #162 review: DeviceCreate.PushToken is a *string. A nil token (or a
// nil body) cannot pass the contract's validation; reaching the handler
// anyway is the 400 backstop, never a nil dereference.
func TestRegisterDevice_aMissingTokenIsA400(t *testing.T) {
	for name, req := range map[string]gen.RegisterDeviceRequestObject{
		"nil token": deviceCreate(t, nil),
		"nil body":  {},
	} {
		t.Run(name, func(t *testing.T) {
			devices := &fakeDevices{}

			_, err := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices}).RegisterDevice(asCaller(t), req)

			appErr, ok := errors.AsType[*apperr.Error](err)
			require.True(t, ok, "%v", err)
			assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
			require.Len(t, appErr.Fields, 1)
			assert.Equal(t, "required", appErr.Fields[0].Code)
			assert.Empty(t, devices.registers)
		})
	}
}

func TestRegisterDevice_passesUseCaseErrors(t *testing.T) {
	devices := &fakeDevices{err: errDevicesDown}
	token := pushToken(t)

	_, err := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices}).RegisterDevice(asCaller(t), deviceCreate(t, &token))

	require.ErrorIs(t, err, errDevicesDown)
}

func TestUnregisterDevice(t *testing.T) {
	devices := &fakeDevices{}

	resp, err := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices}).
		UnregisterDevice(asCaller(t), gen.UnregisterDeviceRequestObject{DeviceId: uuid.MustParse(installationID)})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitUnregisterDeviceResponse(rec))

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, []string{callerID + "/" + installationID}, devices.unregisters)
}

func TestUnregisterDevice_passesUseCaseErrors(t *testing.T) {
	devices := &fakeDevices{err: errDevicesDown}

	_, err := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices}).
		UnregisterDevice(asCaller(t), gen.UnregisterDeviceRequestObject{DeviceId: uuid.MustParse(installationID)})

	require.ErrorIs(t, err, errDevicesDown)
}

func savedSettings(t *testing.T) domain.Settings {
	t.Helper()
	at, err := domain.ParseSendTime("08:30")
	require.NoError(t, err)
	s, err := domain.NewSettings(true, domain.NoEarlyReminder, true, at, true)
	require.NoError(t, err)
	return s
}

func TestGetReminderSettings(t *testing.T) {
	for name, tc := range map[string]struct {
		settings domain.Settings
		want     string
	}{
		"defaults (never saved)": {domain.DefaultSettings(),
			`{"enabled":false,"early_day":14,"on_due_date":true,"send_time":"10:00","show_amounts":false}`},
		"saved, no early reminder": {savedSettings(t),
			`{"enabled":true,"early_day":null,"on_due_date":true,"send_time":"08:30","show_amounts":true}`},
	} {
		t.Run(name, func(t *testing.T) {
			settings := &fakeSettings{out: tc.settings}

			resp, err := remindershttp.NewHandler(remindershttp.UseCases{Settings: settings}).
				GetReminderSettings(asCaller(t), gen.GetReminderSettingsRequestObject{})
			require.NoError(t, err)
			rec := httptest.NewRecorder()
			require.NoError(t, resp.VisitGetReminderSettingsResponse(rec))

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, tc.want, rec.Body.String())
			assert.Equal(t, []string{callerID}, settings.gets)
		})
	}
}

func TestPutReminderSettings(t *testing.T) {
	settings := &fakeSettings{out: savedSettings(t)}
	body := gen.ReminderSettings{Enabled: true, EarlyDay: nil, OnDueDate: true, SendTime: "08:30", ShowAmounts: true}

	resp, err := remindershttp.NewHandler(remindershttp.UseCases{Settings: settings}).
		PutReminderSettings(asCaller(t), gen.PutReminderSettingsRequestObject{Body: &body})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitPutReminderSettingsResponse(rec))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled":true,"early_day":null,"on_due_date":true,"send_time":"08:30","show_amounts":true}`, rec.Body.String())
	assert.Equal(t, []usecase.SettingsInput{{Enabled: true, OnDueDate: true, SendTime: "08:30", ShowAmounts: true}}, settings.puts)
}

func TestPutReminderSettings_earlyDayIsCopied(t *testing.T) {
	settings := &fakeSettings{out: domain.DefaultSettings()}
	day := 3
	body := gen.ReminderSettings{Enabled: true, EarlyDay: &day, OnDueDate: false, SendTime: "07:00"}

	_, err := remindershttp.NewHandler(remindershttp.UseCases{Settings: settings}).
		PutReminderSettings(asCaller(t), gen.PutReminderSettingsRequestObject{Body: &body})

	require.NoError(t, err)
	require.NotNil(t, settings.puts[0].EarlyDay)
	assert.Equal(t, 3, *settings.puts[0].EarlyDay)
}

func TestPutReminderSettings_aMissingBodyIsA400(t *testing.T) {
	settings := &fakeSettings{}

	_, err := remindershttp.NewHandler(remindershttp.UseCases{Settings: settings}).
		PutReminderSettings(asCaller(t), gen.PutReminderSettingsRequestObject{})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
	assert.Empty(t, settings.puts)
}

func TestReminderSettings_passUseCaseErrors(t *testing.T) {
	settings := &fakeSettings{err: errSettingsDown}
	h := remindershttp.NewHandler(remindershttp.UseCases{Settings: settings})

	_, err := h.GetReminderSettings(asCaller(t), gen.GetReminderSettingsRequestObject{})
	require.ErrorIs(t, err, errSettingsDown)
	_, err = h.PutReminderSettings(asCaller(t), gen.PutReminderSettingsRequestObject{Body: &gen.ReminderSettings{SendTime: "10:00"}})
	require.ErrorIs(t, err, errSettingsDown)
}

// Every operation answers the contract's 401 without a principal (the
// router authenticates first; a backstop) and calls no use case.
func TestPreferences_withoutPrincipal(t *testing.T) {
	devices, settings := &fakeDevices{}, &fakeSettings{}
	h := remindershttp.NewHandler(remindershttp.UseCases{Devices: devices, Settings: settings})
	token := pushToken(t)
	ctx := t.Context()

	reg, err := h.RegisterDevice(ctx, deviceCreate(t, &token))
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, visit(t, reg).Code)

	unreg, err := h.UnregisterDevice(ctx, gen.UnregisterDeviceRequestObject{})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, unreg.VisitUnregisterDeviceResponse(rec))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	get, err := h.GetReminderSettings(ctx, gen.GetReminderSettingsRequestObject{})
	require.NoError(t, err)
	rec = httptest.NewRecorder()
	require.NoError(t, get.VisitGetReminderSettingsResponse(rec))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	put, err := h.PutReminderSettings(ctx, gen.PutReminderSettingsRequestObject{})
	require.NoError(t, err)
	rec = httptest.NewRecorder()
	require.NoError(t, put.VisitPutReminderSettingsResponse(rec))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	assert.Empty(t, devices.registers)
	assert.Empty(t, devices.unregisters)
	assert.Empty(t, settings.gets)
	assert.Empty(t, settings.puts)
}
