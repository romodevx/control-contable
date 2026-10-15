//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of getTaxCalendar (M9-04) through the real composition
// root over Postgres as cc_app, on the M9-03 holidays seed (2026, 2027),
// with module_reminders ON and OFF and its prerequisite module_taxpayers
// OFF.

// remindersModules are the flags these suites set.
var remindersModules = []string{"module_auth", "module_taxpayers", "module_reminders"}

// taxCalendarBurst is tax_calendar_reads' burst (MODULE-reminders §2.7).
const taxCalendarBurst = 60

type remindersEnv struct {
	contractEnv
	c     *contracttest.Contract
	token string
}

func newRemindersEnv(t *testing.T) remindersEnv {
	t.Helper()
	env := newContractEnv(t, pingOK{})
	env.apply(t, remindersModules, allOn(remindersModules))
	return remindersEnv{contractEnv: env, c: contracttest.Load(t), token: env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken}
}

func (e remindersEnv) calendar(t *testing.T, query string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("getTaxCalendar"), e.handler, e.send(http.MethodGet, "/v1/tax-calendar"+query, "", e.token))
}

func decodeCalendar(t *testing.T, rec *httptest.ResponseRecorder) gen.TaxCalendar {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body gen.TaxCalendar
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

// calendarExample is the contract's TaxCalendar example (year 2026).
func calendarExample(t *testing.T) string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	raw, err := json.Marshal(spec.Components.Schemas["TaxCalendar"].Value.Examples[0])
	require.NoError(t, err)
	return string(raw)
}

func TestContract_taxCalendar(t *testing.T) {
	env := newRemindersEnv(t)

	flagtest.OnOff(t, "module_reminders", func(t *testing.T, f flagtest.Static, on bool) {
		env.apply(t, remindersModules, f)
		if !on {
			// OFF answers 404 before any input check, a malformed year included.
			for _, query := range []string{"?year=2026", "?year=dos", ""} {
				flagtest.AssertHidden(t, env.calendar(t, query))
			}
			return
		}

		t.Run("200 is the kernel's rule over the seed, privately cacheable", func(t *testing.T) {
			rec := env.calendar(t, "?year=2026")

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "private, max-age=3600", rec.Header().Get("Cache-Control"))
			assert.JSONEq(t, calendarExample(t), rec.Body.String())
		})
		t.Run("December of the last seeded year is weekends-only and alerts", func(t *testing.T) {
			body := decodeCalendar(t, env.calendar(t, "?year=2027"))

			require.Len(t, body.Entries, 12)
			december := body.Entries[11]
			assert.Equal(t, "2027-12", december.Period)
			assert.Equal(t, "2028-01-17", december.DueDate.String())
			assert.Nil(t, december.MovedReason)
			assert.Equal(t, "2027-04-19", body.Entries[2].DueDate.String(), "March 2027: Saturday 17 April moves")
			assert.Contains(t, env.logs.String(), `"event":"reminders_calendar_unseeded"`)
		})
		t.Run("422 calendar_unavailable for an unseeded year", func(t *testing.T) {
			rec := env.calendar(t, "?year=2028")

			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			body := decodeError(t, rec)
			assert.Equal(t, gen.ErrorCodeBusinessRuleViolation, body.Error.Code)
			require.NotNil(t, body.Error.Details)
			require.Len(t, *body.Error.Details, 1)
			assert.Equal(t, "year", (*body.Error.Details)[0].Field)
			assert.Equal(t, "calendar_unavailable", (*body.Error.Details)[0].Code)
		})
		t.Run("400 for a missing or out-of-range year", func(t *testing.T) {
			for _, query := range []string{"", "?year=2021", "?year=dos"} {
				rec := env.calendar(t, query)

				assert.Equal(t, http.StatusBadRequest, rec.Code, query)
				assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"), query)
			}
		})
		t.Run("401 without a token", func(t *testing.T) {
			rec := env.c.ServeOp(t, contracttest.Op("getTaxCalendar"), env.handler,
				env.send(http.MethodGet, "/v1/tax-calendar?year=2026", "", ""))

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})
	}, "module_auth", "module_taxpayers")
}

// §3: module_taxpayers OFF implies module_reminders OFF.
func TestContract_taxCalendarImpliedOff(t *testing.T) {
	env := newRemindersEnv(t)
	env.apply(t, remindersModules, flagtest.Static{"module_auth": true, "module_reminders": true})

	flagtest.AssertHidden(t, env.calendar(t, "?year=2026"))
}

// §2.7: tax_calendar_reads, burst 60, per user; spent → 429 + Retry-After.
func TestContract_taxCalendarRateLimit(t *testing.T) {
	env := newRemindersEnv(t)
	for i := range taxCalendarBurst {
		require.Equal(t, http.StatusOK, env.do(env.send(http.MethodGet, "/v1/tax-calendar?year=2026", "", env.token)).Code, "read %d", i+1)
	}

	rec := env.calendar(t, "?year=2026")

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
	assert.Equal(t, "rate_limited", string(decodeError(t, rec).Error.Code))
}
