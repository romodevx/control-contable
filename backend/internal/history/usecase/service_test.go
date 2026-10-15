package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func mustMonth(t *testing.T, raw string) domain.Month {
	t.Helper()
	m, err := domain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

func requireCode(t *testing.T, err error, code apperr.Code) []apperr.FieldError {
	t.Helper()
	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "want *apperr.Error, got %T: %v", err, err)
	require.Equal(t, code, app.Code)
	return app.Fields
}

func periods(months []usecase.Month) []string {
	out := make([]string, 0, len(months))
	for _, m := range months {
		out = append(out, m.Month.String())
	}
	return out
}

// US-H1: the current year is January .. the current month, newest first,
// read in ONE range estimate and ONE views read, in one read-only unit of
// work (§2.6).
func TestGetCurrentYearInOnePass(t *testing.T) {
	f := newFixture()

	got, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

	require.NoError(t, err)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
	assert.Equal(t, 2026, got.FiscalYear)
	assert.Equal(t, []string{"2026-10", "2026-09", "2026-08", "2026-07", "2026-06", "2026-05", "2026-04", "2026-03",
		"2026-02", "2026-01"}, periods(got.Months))
	assert.True(t, got.Months[0].IsCurrent, "Mes en curso")
	assert.False(t, got.Months[1].IsCurrent)
	assert.Equal(t, []string{"ro"}, f.tx.modes)
	assert.Equal(t, []string{userA + "|" + userA + "|" + taxpayerA + "|2026-01..2026-10"}, f.estimator.calls, "one aggregate pass")
	assert.Equal(t, []string{userA + "|" + taxpayerA + "|2026-01..2026-10"}, f.views.reads, "one views read")
	assert.Equal(t, "legend", got.Legend)
}

// A past year lists all twelve months.
func TestGetPastYear(t *testing.T) {
	f := newFixture()
	f.estimator.byMonth = map[string]usecase.MonthEstimate{}

	svc, err := usecase.NewService(usecase.Deps{
		Tx: f.tx, Flags: f.flags, Estimator: f.estimator, Views: f.views,
		Clock: fixedClock(time.Date(2027, time.March, 2, 9, 0, 0, 0, today.Location())),
	})
	require.NoError(t, err)
	got, err := svc.Get(t.Context(), userA, taxpayerA, 2026)

	require.NoError(t, err)
	require.Len(t, got.Months, 12)
	assert.Equal(t, "2026-12", got.Months[0].Month.String())
	assert.False(t, got.Months[0].IsCurrent, "a past year has no current month")
	assert.Equal(t, []string{userA + "|" + userA + "|" + taxpayerA + "|2026-01..2026-12"}, f.estimator.calls)
}

// §2.2: each month carries the estimate's figures, status, due date, CPA
// flag, fingerprints and versions; tax is null for none and informational.
func TestGetMonthFields(t *testing.T) {
	f := newFixture()
	set := domain.ParameterSet{FiscalYear: 2026, Version: 1}
	tax := domain.Tax{ISRPayable: 15096, IVAPayable: 0, IVAOutcome: "in_favor", IVAInFavor: 200_00, WithholdingsToRemit: 12_50, TotalPayable: 16346}
	f.estimator.byMonth = map[string]usecase.MonthEstimate{
		"2026-09": {Regimes: []domain.Status{domain.StatusIncomplete, domain.StatusComplete}, Tax: tax, DueDate: "2026-10-19",
			CPAReviewPending: true, Seen: seenOf(mustMonth(t, "2026-09"), "9"), ParameterSet: set, HasIncomes: true},
		"2026-08": {Regimes: []domain.Status{domain.StatusInformational}, Tax: domain.Tax{IVAOutcome: "zero"}, DueDate: "2026-09-17",
			Seen: seenOf(mustMonth(t, "2026-08"), "8"), ParameterSet: set},
		"2026-02": {DueDate: "2026-03-17", Seen: seenOf(mustMonth(t, "2026-02"), "2"), ParameterSet: set, HasExpenses: true},
	}

	got, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

	require.NoError(t, err)
	sep, aug, feb := got.Months[1], got.Months[2], got.Months[8]
	assert.Equal(t, domain.StatusIncomplete, sep.Status)
	require.NotNil(t, sep.Tax)
	assert.Equal(t, tax, *sep.Tax)
	assert.Equal(t, "2026-10-19", sep.DueDate)
	assert.True(t, sep.CPAReviewPending)
	assert.True(t, sep.HasRecords)
	assert.Equal(t, hex64("9"), sep.Seen.FiguresFingerprint)
	assert.Equal(t, set, sep.ParameterSet)
	assert.Equal(t, domain.StatusInformational, aug.Status)
	assert.Nil(t, aug.Tax, "every regime informational: no amount")
	assert.False(t, aug.HasRecords)
	assert.Equal(t, domain.StatusNone, feb.Status, "no regime held (Sin régimen en este mes)")
	assert.Nil(t, feb.Tax)
	assert.True(t, feb.HasRecords, "an expense dated in the month")
}

// §3: has_records counts only the record modules that are ON.
func TestGetHasRecordsCountsOnlyEnabledModules(t *testing.T) {
	for _, tc := range []struct {
		incomes, expenses bool
		want              [2]bool // income-only month, expense-only month
	}{
		{true, true, [2]bool{true, true}},
		{false, true, [2]bool{false, true}},
		{true, false, [2]bool{true, false}},
		{false, false, [2]bool{false, false}},
	} {
		f := newFixture()
		f.flags = fakeFlags{usecase.ModuleIncomes: tc.incomes, usecase.ModuleExpenses: tc.expenses}
		f.estimator.byMonth = map[string]usecase.MonthEstimate{
			"2026-10": {HasIncomes: true, Seen: seenOf(mustMonth(t, "2026-10"), "c")},
			"2026-09": {HasExpenses: true, Seen: seenOf(mustMonth(t, "2026-09"), "c")},
		}

		got, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

		require.NoError(t, err)
		assert.Equal(t, tc.want, [2]bool{got.Months[0].HasRecords, got.Months[1].HasRecords}, "%+v", tc)
	}
}

// §2.4 through the use case: a viewed month carries viewed_at and its
// marker; the current month never marks; a never-viewed month is clean.
func TestGetChangedSinceViewed(t *testing.T) {
	f := newFixture()
	viewedAt := time.Date(2026, time.October, 3, 15, 30, 0, 0, time.UTC)
	view := func(raw, figures, engine string) domain.View {
		s := seenOf(mustMonth(t, raw), figures)
		s.EngineVersion = engine
		return domain.View{Month: mustMonth(t, raw), Seen: s, ViewedAt: viewedAt}
	}
	f.views.views = []domain.View{
		view("2026-10", "0", "1.0.1"), // current: never marked
		view("2026-09", "b", "1.0.1"), // same figures
		view("2026-08", "0", "1.0.1"), // records
		view("2026-07", "0", "1.0.0"), // engine
	}

	got, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

	require.NoError(t, err)
	byPeriod := map[string]usecase.Month{}
	for _, m := range got.Months {
		byPeriod[m.Month.String()] = m
	}
	assert.Equal(t, domain.ChangeReason(""), byPeriod["2026-10"].Change)
	require.NotNil(t, byPeriod["2026-10"].ViewedAt, "the current month's view is stored and shown")
	assert.Equal(t, domain.ChangeReason(""), byPeriod["2026-09"].Change)
	assert.Equal(t, viewedAt, *byPeriod["2026-09"].ViewedAt)
	assert.Equal(t, domain.ChangeRecords, byPeriod["2026-08"].Change)
	assert.Equal(t, domain.ChangeEngine, byPeriod["2026-07"].Change)
	assert.Nil(t, byPeriod["2026-06"].ViewedAt, "never viewed")
	assert.Equal(t, domain.ChangeReason(""), byPeriod["2026-06"].Change)
}

// §2.2: year totals sum the shown months (null months add nothing);
// disclaimers are the union of the months'.
func TestGetYearTotalsAndDisclaimers(t *testing.T) {
	f := newFixture()
	cpa := domain.Disclaimer{Code: "cpa_review_pending", Severity: domain.SeverityStrong, Text: "c"}
	notAdvice := domain.Disclaimer{Code: "not_tax_advice", Severity: domain.SeverityInfo, Text: "n"}
	f.estimator.byMonth = map[string]usecase.MonthEstimate{
		"2026-03": {Regimes: []domain.Status{domain.StatusComplete}, Tax: domain.Tax{ISRPayable: 1_00, IVAPayable: 2_00,
			WithholdingsToRemit: 3_00, TotalPayable: 6_00}, Disclaimers: []domain.Disclaimer{notAdvice, cpa}, Seen: seenOf(mustMonth(t, "2026-03"), "b")},
		"2026-02": {Regimes: []domain.Status{domain.StatusInformational}, Tax: domain.Tax{ISRPayable: 9_99}, Seen: seenOf(mustMonth(t, "2026-02"), "b")},
		"2026-01": {Seen: seenOf(mustMonth(t, "2026-01"), "b")},
	}

	got, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

	require.NoError(t, err)
	// Seven default months of 100 ISR (April..October) plus March.
	assert.Equal(t, domain.YearTotals{ISRPayable: 7_00 + 1_00, IVAPayable: 2_00, WithholdingsToRemit: 3_00, TotalPayable: 7_00 + 6_00}, got.Totals)
	assert.Equal(t, []domain.Disclaimer{notAdvice, cpa}, got.Disclaimers)
}

func TestGetTotalsOverflowFails(t *testing.T) {
	f := newFixture()
	f.estimator.byMonth = map[string]usecase.MonthEstimate{
		"2026-02": {Regimes: []domain.Status{domain.StatusComplete}, Tax: domain.Tax{TotalPayable: money.Cents(1<<63 - 1)}, Seen: seenOf(mustMonth(t, "2026-02"), "b")},
	}

	_, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

	assert.ErrorIs(t, err, money.ErrOverflow)
}

// §2.1 and §5: the refusals, before anything is read.
func TestGetRefusesBadInput(t *testing.T) {
	f := newFixture()
	svc := f.service()

	_, err := svc.Get(t.Context(), userA, taxpayerA, 2027)
	fields := requireCode(t, err, apperr.CodeValidationFailed)
	assert.Equal(t, []apperr.FieldError{{Field: "fiscal_year", Code: "out_of_range", Message: "Aún no hay datos de un año futuro."}}, fields)

	_, err = svc.Get(t.Context(), userA, taxpayerA, 2025)
	fields = requireCode(t, err, apperr.CodeBusinessRuleViolation)
	assert.Equal(t, []apperr.FieldError{{Field: "fiscal_year", Code: "parameters_unavailable", Message: "Sólo hay datos desde 2026."}}, fields)
	assert.ErrorIs(t, err, usecase.ErrParametersUnavailable)

	_, err = svc.Get(t.Context(), userA, "not-a-uuid", 2026)
	requireCode(t, err, apperr.CodeNotFound)
	assert.Empty(t, f.tx.modes, "nothing is read for a refused request")
}

// §8: an unknown or another user's taxpayer (the engine's profile 404) is
// the hidden 404, indistinguishable from the module OFF; the cause stays.
func TestGetPassesTheNotFoundOnHidden(t *testing.T) {
	f := newFixture()
	f.estimator.err = notFoundPort

	_, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)

	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeNotFound, app.Code)
	assert.Equal(t, "Not found.", app.Message)
	assert.ErrorIs(t, err, usecase.ErrTaxpayerNotFound)
	assert.Empty(t, f.views.reads, "no view is read for a foreign taxpayer")
}

// §5: no partial mode; any failure fails the whole request.
func TestGetFailsWhole(t *testing.T) {
	f := newFixture()
	f.estimator.err = errPortDown
	_, err := f.service().Get(t.Context(), userA, taxpayerA, 2026)
	assert.ErrorIs(t, err, errPortDown)

	f = newFixture()
	f.views.err = errPortDown
	_, err = f.service().Get(t.Context(), userA, taxpayerA, 2026)
	assert.ErrorIs(t, err, errPortDown)

	f = newFixture()
	f.estimator.short = true
	_, err = f.service().Get(t.Context(), userA, taxpayerA, 2026)
	assert.ErrorIs(t, err, usecase.ErrEstimateCoverage)
}

func viewedRequest() usecase.ViewedRequest {
	return usecase.ViewedRequest{
		Period: "2026-03", RecordsFingerprint: hex64("a"), FiguresFingerprint: hex64("b"), EngineVersion: "1.0.1",
		ParamsYear: 2026, ParamsVersion: 1,
	}
}

// §2.5: one upsert of what the user saw, stamped with the server clock, in
// one read-write unit of work; the current month is accepted.
func TestMarkViewed(t *testing.T) {
	for _, period := range []string{"2026-03", "2026-10"} {
		f := newFixture()
		req := viewedRequest()
		req.Period = period

		require.NoError(t, f.service().MarkViewed(t.Context(), userA, taxpayerA, req))

		assert.Equal(t, []string{"rw"}, f.tx.modes)
		assert.Equal(t, []string{userA + "|" + userA + "|" + taxpayerA + "|" + period + "|a|b|1.0.1|2026/v1|" +
			today.Format(time.RFC3339)}, f.views.saved)
		assert.Empty(t, f.estimator.calls, "no estimate is computed (§2.5)")
	}
}

// §2.5 and §5: format, range and ownership refusals.
func TestMarkViewedRefusals(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*usecase.ViewedRequest)
		code   apperr.Code
		field  string
		detail string
	}{
		{"malformed period", func(r *usecase.ViewedRequest) { r.Period = "2026-3" }, apperr.CodeValidationFailed, "period", "pattern"},
		{"future period", func(r *usecase.ViewedRequest) { r.Period = "2026-11" }, apperr.CodeValidationFailed, "period", "out_of_range"},
		{"before 2026", func(r *usecase.ViewedRequest) { r.Period = "2025-12" }, apperr.CodeBusinessRuleViolation, "period", "parameters_unavailable"},
		{"records fingerprint", func(r *usecase.ViewedRequest) { r.RecordsFingerprint = "ABC" }, apperr.CodeValidationFailed, "records_fingerprint", "pattern"},
		{"figures fingerprint", func(r *usecase.ViewedRequest) { r.FiguresFingerprint = hex64("g") }, apperr.CodeValidationFailed, "figures_fingerprint", "pattern"},
		{"engine version", func(r *usecase.ViewedRequest) { r.EngineVersion = "1 0" }, apperr.CodeValidationFailed, "engine_version", "pattern"},
		{"parameter set", func(r *usecase.ViewedRequest) { r.ParamsVersion = 0 }, apperr.CodeValidationFailed, "parameter_set", "pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture()
			req := viewedRequest()
			tt.mutate(&req)

			err := f.service().MarkViewed(t.Context(), userA, taxpayerA, req)

			fields := requireCode(t, err, tt.code)
			require.Len(t, fields, 1)
			assert.Equal(t, tt.field, fields[0].Field)
			assert.Equal(t, tt.detail, fields[0].Code)
			assert.Empty(t, f.tx.modes)
		})
	}
}

// §8: a taxpayer that is not the caller's (the foreign key) and a
// malformed id are the hidden 404; a store failure passes on.
func TestMarkViewedNotFoundAndFailures(t *testing.T) {
	f := newFixture()
	f.views.saveErr = usecase.ErrTaxpayerNotFound
	err := f.service().MarkViewed(t.Context(), userA, taxpayerA, viewedRequest())
	requireCode(t, err, apperr.CodeNotFound)
	assert.ErrorIs(t, err, usecase.ErrTaxpayerNotFound)

	err = newFixture().service().MarkViewed(t.Context(), userA, "nope", viewedRequest())
	requireCode(t, err, apperr.CodeNotFound)

	f = newFixture()
	f.views.saveErr = errPortDown
	err = f.service().MarkViewed(t.Context(), userA, taxpayerA, viewedRequest())
	assert.ErrorIs(t, err, errPortDown)
	_, isApp := errors.AsType[*apperr.Error](err)
	assert.False(t, isApp, "an infrastructure failure is the generic 500")
}

// The id is canonical lower case wherever it goes (ids compare as strings).
func TestCanonicalTaxpayerID(t *testing.T) {
	f := newFixture()
	upper := "0199D4A2-1111-7000-8000-00000000E00A"

	got, err := f.service().Get(t.Context(), userA, upper, 2026)
	require.NoError(t, err)
	assert.Equal(t, taxpayerA, got.TaxpayerID)
	require.NoError(t, f.service().MarkViewed(t.Context(), userA, upper, viewedRequest()))
	assert.Contains(t, f.views.saved[0], "|"+taxpayerA+"|")
}

func TestNewServiceRefusesMissingPorts(t *testing.T) {
	_, err := usecase.NewService(usecase.Deps{})
	assert.ErrorIs(t, err, usecase.ErrMissingDependency)
}
