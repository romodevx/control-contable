package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit/ratelimitpg"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	remindersusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// stubRemindersModule is the reminders handler with no use case: the flag
// gate or the authenticator answers before it in these tests.
func stubRemindersModule() remindersModule {
	return remindersModule{handler: remindershttp.NewHandler(nil)}
}

func TestNewHandler_refusesToServeWithoutTheRemindersHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t),
		taxpayers: stubTaxpayersModule(), incomes: stubIncomesModule(), expenses: stubExpensesModule(),
		taxEngine: stubTaxEngineModule(), dashboard: stubDashboardModule(), history: stubHistoryModule(),
		annual: stubAnnualModule(), now: time.Now, logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no reminders handler")
}

// noTx is a unit of work that is never run (the wiring test only builds).
type noTx struct{}

func (noTx) ReadWrite(context.Context, string, tx.Work) error { return nil }
func (noTx) ReadOnly(context.Context, string, tx.Work) error  { return nil }

func TestNewRemindersModule_needsEveryPort(t *testing.T) {
	logger := logging.New(io.Discard, nil)
	limits := &ratelimitpg.Store{}
	for name, w := range map[string]remindersWiring{
		"tx":     {limits: limits, logger: logger},
		"limits": {tx: noTx{}, logger: logger},
		"logger": {tx: noTx{}, limits: limits},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := newRemindersModule(w)

			require.ErrorIs(t, err, remindersusecase.ErrMissingDependency)
		})
	}
}

func TestNewRemindersModule_registersTheCalendarBudget(t *testing.T) {
	m, err := newRemindersModule(remindersWiring{tx: noTx{}, limits: &ratelimitpg.Store{}, logger: logging.New(io.Discard, nil)})

	require.NoError(t, err)
	assert.NotNil(t, m.handler)
	assert.Contains(t, m.userLimits, "getTaxCalendar")
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

// MODULE-reminders §3, US-R10: getTaxCalendar is hidden while
// module_reminders is OFF or implied OFF (module_taxpayers OFF), a
// malformed year included; with both ON it is routed past the gate.
func TestTaxCalendarRoute_needsRemindersAndTaxpayersFlags(t *testing.T) {
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
			h := newFlaggedHandler(t, testConfig(), pingOK{}, flags.NewService(keysReader{keys: tc.enabled}))
			for _, target := range []string{"/v1/tax-calendar?year=2026", "/v1/tax-calendar?year=nope"} {
				rec := serveRecorded(h, newRequest(http.MethodGet, target, ""))
				if !tc.visible {
					flagtest.AssertHidden(t, rec)
				} else {
					assert.Equal(t, http.StatusUnauthorized, rec.Code, "routed past the gate")
				}
			}
		})
	}
}
