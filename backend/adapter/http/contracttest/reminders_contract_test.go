package contracttest_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

// M9-01 (contract 0.13.0, MODULE-reminders §10): the response shapes the M9
// handlers must send, pinned against the frozen contract before they are
// built. These validate canned responses only, so they claim no coverage:
// the five operations stay in pendingOperations until M9-04 (getTaxCalendar)
// and M9-05 (devices, settings) serve them from the real router with ServeOp.

const (
	deviceID         = "0199d4a4-3c4d-7f5e-8a61-7c8d9e0f1a23"
	deviceBody       = `{"id":"` + deviceID + `","platform":"android","app_version":null,"created_at":"2026-10-15T15:02:00Z","last_seen_at":"2026-10-15T15:02:00Z"}`
	settingsBody     = `{"enabled":true,"early_day":null,"on_due_date":true,"send_time":"10:00","show_amounts":false}`
	calendarCache    = "private, max-age=3600"
	monthsInYear     = 12
	calendarYear     = 2026
	ruleViolationFmt = `{"error":{"code":"business_rule_violation","message":"m","request_id":"req_1","details":[{"field":%q,"code":%q,"message":"m"}]}}`
)

// calendarBody is a TaxCalendar of calendarYear with n entries (weekends are
// not moved: the contract checks shapes, the kernel golden checks dates).
func calendarBody(n int) string {
	entries := make([]string, 0, n)
	for month := 1; month <= n; month++ {
		dueYear, dueMonth := calendarYear, month+1
		if dueMonth > monthsInYear {
			dueYear, dueMonth = calendarYear+1, 1
		}
		due := fmt.Sprintf("%d-%02d-17", dueYear, dueMonth)
		entries = append(entries, fmt.Sprintf(
			`{"period":"%d-%02d","obligation":"monthly_isr_iva","nominal_due_date":%q,"due_date":%q,"moved_reason":null,"subject_to_confirmation":true}`,
			calendarYear, month, due, due))
	}
	return fmt.Sprintf(`{"year":%d,"entries":[%s]}`, calendarYear, strings.Join(entries, ","))
}

func ruleViolation(field, code string) string { return fmt.Sprintf(ruleViolationFmt, field, code) }

func registerDevice() *http.Request {
	return httptest.NewRequest(http.MethodPost, "/v1/devices", nil)
}

func taxCalendar() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/v1/tax-calendar?year=2026", nil)
}

func TestRemindersContract_acceptsWhatTheHandlersWillSend(t *testing.T) {
	c := contracttest.Load(t)
	tests := []struct {
		name string
		req  *http.Request
		h    http.Handler
	}{
		{"registerDevice 201 (new or claimed)", registerDevice(),
			respond(http.StatusCreated, jsonType, deviceBody, noStore...)},
		{"registerDevice 200 (refreshed)", registerDevice(),
			respond(http.StatusOK, jsonType, deviceBody, noStore...)},
		{"registerDevice 429 with Retry-After", registerDevice(),
			respond(http.StatusTooManyRequests, jsonType, rateLimitBody, "Cache-Control", "no-store", "Retry-After", "60")},
		{"unregisterDevice 204", httptest.NewRequest(http.MethodDelete, "/v1/devices/"+deviceID, nil),
			respond(http.StatusNoContent, "", "", noStore...)},
		{"unregisterDevice 404 (foreign, unknown or flag OFF)", httptest.NewRequest(http.MethodDelete, "/v1/devices/"+deviceID, nil),
			respond(http.StatusNotFound, jsonType, notFoundBody, noStore...)},
		{"getReminderSettings 200 with early_day null", httptest.NewRequest(http.MethodGet, "/v1/reminder-settings", nil),
			respond(http.StatusOK, jsonType, settingsBody, noStore...)},
		{"putReminderSettings 422 no_reminder_selected", httptest.NewRequest(http.MethodPut, "/v1/reminder-settings", nil),
			respond(http.StatusUnprocessableEntity, jsonType, ruleViolation("body", "no_reminder_selected"), noStore...)},
		{"getTaxCalendar 200, private cache", taxCalendar(),
			respond(http.StatusOK, jsonType, calendarBody(monthsInYear), "Cache-Control", calendarCache)},
		{"getTaxCalendar 422 calendar_unavailable", taxCalendar(),
			respond(http.StatusUnprocessableEntity, jsonType, ruleViolation("year", "calendar_unavailable"), noStore...)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spy := run(t, func(tb testing.TB) { c.Serve(tb, tc.h, tc.req) })
			assert.Empty(t, spy.report())
		})
	}
}

func TestRemindersContract_rejectsWhatTheSpecForbids(t *testing.T) {
	c := contracttest.Load(t)
	tokenLeak := strings.Replace(deviceBody, `"platform"`, `"push_token":"x","platform"`, 1)
	noEarlyDay := strings.Replace(settingsBody, `"early_day":null,`, "", 1)
	tests := []struct {
		name string
		req  *http.Request
		h    http.Handler
	}{
		// §2.3: the token is write-only, never returned.
		{"a Device carrying the push token", registerDevice(),
			respond(http.StatusCreated, jsonType, tokenLeak, noStore...)},
		// §10 item 3: early_day is required (nullable), never implied.
		{"settings without early_day", httptest.NewRequest(http.MethodGet, "/v1/reminder-settings", nil),
			respond(http.StatusOK, jsonType, noEarlyDay, noStore...)},
		// §10 item 4: the calendar's 200 is the private-cache exception.
		{"a calendar sent no-store", taxCalendar(),
			respond(http.StatusOK, jsonType, calendarBody(monthsInYear), noStore...)},
		{"a calendar missing a period", taxCalendar(),
			respond(http.StatusOK, jsonType, calendarBody(monthsInYear-1), "Cache-Control", calendarCache)},
		// The API-wide no-409 exception: a claim is never a conflict.
		{"registerDevice answering 409", registerDevice(),
			respond(http.StatusConflict, jsonType, notFoundBody, noStore...)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spy := run(t, func(tb testing.TB) { c.Serve(tb, tc.h, tc.req) })
			assert.NotEmpty(t, spy.report(), "the contract must reject it")
		})
	}
}
