package main

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	remindersusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// stubRemindersModule is the reminders handler with no use case: the flag
// gate or the authenticator answers before it in these tests.
func stubRemindersModule() remindersModule {
	return remindersModule{handler: remindershttp.NewHandler(remindershttp.UseCases{})}
}

// lazyTx is a cc_app unit of work over a pool that never connects
// (sql.Open is lazy): enough to wire the module without a database.
func lazyTx(t *testing.T) *pgdb.TxManager {
	t.Helper()
	pool, err := sql.Open("pgx", "postgres://127.0.0.1:1/unused")
	require.NoError(t, err)
	t.Cleanup(func() { _ = pool.Close() })
	txm, err := pgdb.NewTxManager(pool)
	require.NoError(t, err)
	return txm
}

// testReminderKeys are fixture keys (one byte each), distinct from the
// auth and RFC fixtures, never real ones.
func testReminderKeys() config.Reminders {
	key := func(b byte) config.Secret { return config.NewSecret(string(bytes.Repeat([]byte{b}, crypto.KeySize))) }
	return config.Reminders{PushTokenKey: key(0x71), PushTokenKeyID: 1, PushTokenIndexPepper: key(0x72)}
}

// withReminderKeys adds fixture push token keys (hex, all one byte),
// distinct from withAuthKeys's and withTaxpayersKeys's.
func withReminderKeys(env map[string]string) map[string]string {
	env["PUSH_TOKEN_ENC_KEY"] = strings.Repeat("7c", crypto.KeySize)
	env["PUSH_TOKEN_INDEX_PEPPER"] = strings.Repeat("7d", crypto.MinPepperSize)
	return env
}

// One secret, one purpose: the process refuses to start when a push token
// key is missing outside dev or repeats another key, in every environment.
func TestRun_refusesMissingOrSharedReminderKeys(t *testing.T) {
	base := func(appEnv string) map[string]string {
		return withReminderKeys(withTaxpayersKeys(withAuthKeys(map[string]string{
			"APP_ENV": appEnv, "TRUSTED_PROXY_HOPS": "0", "DATABASE_URL": "postgres://u@127.0.0.1:1/db",
		})))
	}
	tests := []struct {
		name    string
		appEnv  string
		mutate  func(map[string]string)
		wantErr string
	}{
		{"missing outside dev", "staging", func(e map[string]string) { delete(e, "PUSH_TOKEN_ENC_KEY") }, "PUSH_TOKEN_ENC_KEY"},
		{"the RFC key reused (dev)", "dev", func(e map[string]string) { e["PUSH_TOKEN_ENC_KEY"] = e["RFC_ENC_KEY"] },
			"PUSH_TOKEN_ENC_KEY must differ from RFC_ENC_KEY"},
		{"the pepper is the key (production)", "production",
			func(e map[string]string) { e["PUSH_TOKEN_INDEX_PEPPER"] = e["PUSH_TOKEN_ENC_KEY"] },
			"PUSH_TOKEN_INDEX_PEPPER must differ from PUSH_TOKEN_ENC_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := base(tt.appEnv)
			tt.mutate(env)

			err := runWith(env)

			require.Error(t, err)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func remindersTestWiring(t *testing.T) remindersWiring {
	t.Helper()
	return remindersWiring{
		env: config.EnvDev, keys: testReminderKeys(), tx: lazyTx(t),
		holidays: holidaySnapshot{}, limits: &ratelimitpg.Store{}, logger: logging.New(io.Discard, nil),
	}
}

func TestNewHandler_refusesToServeWithoutTheRemindersHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t),
		taxpayers: stubTaxpayersModule(), incomes: stubIncomesModule(), expenses: stubExpensesModule(),
		taxEngine: stubTaxEngineModule(), dashboard: stubDashboardModule(), history: stubHistoryModule(),
		annual: stubAnnualModule(), now: time.Now, logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no reminders handler")
}

func TestNewRemindersModule_needsEveryPort(t *testing.T) {
	for name, drop := range map[string]func(*remindersWiring){
		"holidays": func(w *remindersWiring) { w.holidays = nil },
		"limits":   func(w *remindersWiring) { w.limits = nil },
		"logger":   func(w *remindersWiring) { w.logger = nil },
		"tx":       func(w *remindersWiring) { w.tx = nil },
	} {
		t.Run(name, func(t *testing.T) {
			w := remindersTestWiring(t)
			drop(&w)

			_, err := newRemindersModule(w)

			require.ErrorIs(t, err, remindersusecase.ErrMissingDependency)
		})
	}
}

// MODULE-reminders §2.7: the calendar read budget and the two write
// budgets (registerDevice and unregisterDevice share devices_writes).
func TestNewRemindersModule_registersTheBudgets(t *testing.T) {
	m, err := newRemindersModule(remindersTestWiring(t))

	require.NoError(t, err)
	assert.NotNil(t, m.handler)
	assert.ElementsMatch(t, []string{"getTaxCalendar", "registerDevice", "unregisterDevice", "putReminderSettings"},
		mapKeys(m.userLimits))
}

func mapKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// §2.3, OQ-R7: in dev a missing push token key becomes an ephemeral one,
// said loudly in the log (tokens sealed with it cannot be opened after a
// restart: the next registerDevice re-seals them).
func TestPushTokenVault_devFillsMissingKeysWithEphemeralOnes(t *testing.T) {
	var logs bytes.Buffer
	w := remindersTestWiring(t)
	w.keys, w.logger = config.Reminders{PushTokenKeyID: 1}, logging.New(&logs, nil)

	vault, err := newPushTokenVault(w)

	require.NoError(t, err)
	tok, err := domain.NewPushToken("fcm:" + t.Name())
	require.NoError(t, err)
	sealed, err := vault.Seal(tok, remindersusecase.TokenBinding{
		InstallationID: "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23", UserID: "0199d4a2-0000-7000-8000-00000000000a",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, sealed.Ciphertext)
	assert.Contains(t, logs.String(), "PUSH_TOKEN_ENC_KEY")
	assert.Contains(t, logs.String(), "PUSH_TOKEN_INDEX_PEPPER")
	assert.Contains(t, logs.String(), "ephemeral")
}

func TestPushTokenVault_neverEphemeralOutsideDev(t *testing.T) {
	w := remindersTestWiring(t)
	w.env, w.keys = config.EnvProduction, config.Reminders{PushTokenKeyID: 1}

	_, err := newPushTokenVault(w)

	require.ErrorIs(t, err, errMissingKey)
	assert.ErrorContains(t, err, "PUSH_TOKEN_ENC_KEY")
}

// MODULE-reminders §2.1, §9: an unseeded next year is a structured warning
// the reminders_calendar_unseeded alert matches, carrying the year only.
func TestCalendarAlerts_logTheUnseededYear(t *testing.T) {
	var buf syncBuffer
	alerts := calendarAlerts{logger: logging.New(&buf, nil)}

	alerts.CalendarUnseeded(t.Context(), 2028)

	assert.Contains(t, buf.String(), `"event":"reminders_calendar_unseeded"`)
	assert.Contains(t, buf.String(), `"year":2028`)
	assert.Contains(t, buf.String(), `"level":"WARN"`)
}

// remindersRoutes are the five module_reminders routes, each with a body
// the router would otherwise validate.
var remindersRoutes = []struct{ method, target, body string }{
	{http.MethodGet, "/v1/tax-calendar?year=2026", ""},
	{http.MethodGet, "/v1/tax-calendar?year=nope", ""},
	{http.MethodPost, "/v1/devices", `{"id":"nope"}`},
	{http.MethodDelete, "/v1/devices/0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23", ""},
	{http.MethodDelete, "/v1/devices/nope", ""},
	{http.MethodGet, "/v1/reminder-settings", ""},
	{http.MethodPut, "/v1/reminder-settings", `{}`},
}

// MODULE-reminders §3, US-R10: every route is hidden while module_reminders
// is OFF or implied OFF (module_taxpayers OFF), malformed input included;
// with both ON each is routed past the gate (here to the authenticator).
func TestRemindersRoutes_needRemindersAndTaxpayersFlags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		enabled []string
		visible bool
	}{
		{"both ON", []string{"module_taxpayers", "module_reminders"}, true},
		{"reminders OFF", []string{"module_taxpayers"}, false},
		{"taxpayers OFF", []string{"module_reminders"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := testConfig()
			cfg.RateLimit.Burst = len(remindersRoutes) // the edge limiter stays out of the way
			h := newFlaggedHandler(t, cfg, pingOK{}, flags.NewService(keysReader{keys: tc.enabled}))
			for _, route := range remindersRoutes {
				rec := serveRecorded(h, newRequest(route.method, route.target, route.body))
				if !tc.visible {
					flagtest.AssertHidden(t, rec)
				} else {
					assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s %s routed past the gate", route.method, route.target)
				}
			}
		})
	}
}
