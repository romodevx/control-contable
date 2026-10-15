package remindershttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	remindershttp "github.com/romodevx/control-contable/backend/internal/reminders/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const callerID = "0199d4a2-0000-7000-8000-00000000c4aa"

var errCalendarDown = errors.New("fake: calendar down")

type fakeCalendar struct {
	calls []string
	year  usecase.Year
	err   error
}

func (f *fakeCalendar) Year(_ context.Context, userID string, year int) (usecase.Year, error) {
	f.calls = append(f.calls, userID)
	f.year.Year = year
	return f.year, f.err
}

func asCaller(t *testing.T) context.Context {
	t.Helper()
	return principal.With(t.Context(), principal.Principal{UserID: callerID})
}

func date(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

// calendar2026 is the contract's TaxCalendar example as the use case
// answers it (the kernel's rule over the 2026-2027 seed).
func calendar2026(t *testing.T) usecase.Year {
	t.Helper()
	moved := map[string]string{"2026-04": "2026-05-18", "2026-09": "2026-10-19", "2026-12": "2027-01-18"}
	y := usecase.Year{Year: 2026}
	for m := time.January; m <= time.December; m++ {
		period := time.Date(2026, m, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")
		nominal := time.Date(2026, m+1, fiscal.MonthlyDueDay, 0, 0, 0, 0, time.UTC).Format(time.DateOnly)
		e := usecase.Entry{Period: period, Nominal: date(t, nominal), Due: date(t, nominal), SubjectToConfirmation: true}
		if due, ok := moved[period]; ok {
			e.Due, e.MovedReason = date(t, due), fiscal.MovedWeekend
		}
		y.Entries = append(y.Entries, e)
	}
	return y
}

// schemaExample is the first example of a component schema.
func schemaExample(t *testing.T, schema string) string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	ref := spec.Components.Schemas[schema]
	require.NotNil(t, ref, schema)
	require.NotEmpty(t, ref.Value.Examples, schema)
	raw, err := json.Marshal(ref.Value.Examples[0])
	require.NoError(t, err)
	return string(raw)
}

// 200: the contract's example body, Cache-Control private, max-age=3600
// (the second exception to no-store, contract 0.13.0).
func TestGetTaxCalendar(t *testing.T) {
	cal := &fakeCalendar{year: calendar2026(t)}
	h := remindershttp.NewHandler(cal)

	resp, err := h.GetTaxCalendar(asCaller(t), gen.GetTaxCalendarRequestObject{Params: gen.GetTaxCalendarParams{Year: 2026}})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxCalendarResponse(rec))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "private, max-age=3600", rec.Header().Get("Cache-Control"))
	assert.JSONEq(t, schemaExample(t, "TaxCalendar"), rec.Body.String())
	assert.Equal(t, []string{callerID}, cal.calls)
}

func TestGetTaxCalendarWithoutPrincipal(t *testing.T) {
	cal := &fakeCalendar{}

	resp, err := remindershttp.NewHandler(cal).GetTaxCalendar(t.Context(), gen.GetTaxCalendarRequestObject{})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxCalendarResponse(rec))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, `Bearer realm="control-contable"`, rec.Header().Get("WWW-Authenticate"))
	assert.Empty(t, cal.calls)
}

// A use case error (422 calendar_unavailable, 500) passes as is: the
// router writes the envelope, no-store.
func TestGetTaxCalendarPassesUseCaseErrors(t *testing.T) {
	cal := &fakeCalendar{err: errCalendarDown}

	_, err := remindershttp.NewHandler(cal).GetTaxCalendar(asCaller(t), gen.GetTaxCalendarRequestObject{Params: gen.GetTaxCalendarParams{Year: 2026}})

	require.ErrorIs(t, err, errCalendarDown)
}

// An entry without a date is a use case bug: a 500, never "0001-01-01".
func TestGetTaxCalendarRefusesAZeroDate(t *testing.T) {
	for name, entry := range map[string]usecase.Entry{
		"nominal": {Period: "2026-01", Due: date(t, "2026-02-17")},
		"due":     {Period: "2026-01", Nominal: date(t, "2026-02-17")},
	} {
		t.Run(name, func(t *testing.T) {
			cal := &fakeCalendar{year: usecase.Year{Entries: []usecase.Entry{entry}}}

			_, err := remindershttp.NewHandler(cal).GetTaxCalendar(asCaller(t), gen.GetTaxCalendarRequestObject{Params: gen.GetTaxCalendarParams{Year: 2026}})

			require.Error(t, err)
		})
	}
}

// MODULE-reminders §2.7: tax_calendar_reads, 120 per hour, burst 60.
func TestUserLimits(t *testing.T) {
	assert.Equal(t, map[string]ratelimit.Policy{
		"getTaxCalendar": {Scope: "tax_calendar_reads", Burst: 60, Interval: 30 * time.Second},
	}, remindershttp.UserLimits)
	for _, p := range remindershttp.UserLimits {
		require.NoError(t, p.Validate())
	}
}
