//go:build integration

package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The M6-03 performance check (MODULE-dashboard §2.4, US-D9): one user
// whose taxpayer holds three regimes (612, 606, 625) with 50,000 incomes
// and 50,000 expenses spread over January 2026..June 2027, so the default
// window (2026-07..2027-06) spans two fiscal years and the read reaches
// back to January 2026 (YTD regimes). About five times the PRD's 10,000
// records a year. It records every statement one getDashboard issues (a
// pgx tracer on the cc_app pool), prints EXPLAIN (ANALYZE, BUFFERS) of each
// as cc_app under RLS, pins the statement count, then measures getDashboard
// end to end through the router (authentication, flag gate, per-user read
// budget, use case, JSON): p95 < 1 s.
//
// It seeds 100k rows, so it runs only on demand:
//
//	CC_PERF=1 go test -tags=integration -count=1 -run TestPerf_dashboardReadPath -v ./cmd/api/
const (
	dashPerfSequential = 300
	dashPerfConcurrent = 200
	dashPerfP95Budget  = time.Second
	// dashPerfReads is the read statements of one engine-ON request: the
	// profile (taxpayer + regimes), one aggregate per table, the tax inputs
	// of each of the two fiscal years, the regime texts and the access
	// summary (entitlements, usage). The window has records, so is_empty
	// needs no EXISTS. It does not grow with months.
	dashPerfReads = 9
	// dashPerfSpanDays is January 1st 2026 .. June 20th 2027.
	dashPerfSpanDays = 536
)

// dashPerfToday is 20 June 2027.
var dashPerfToday = time.Date(2027, time.June, 20, 12, 0, 0, 0, clock.Location())

func TestPerf_dashboardReadPath(t *testing.T) {
	if os.Getenv("CC_PERF") == "" {
		t.Skip("set CC_PERF=1 to seed 100k records and measure the dashboard read path")
	}
	rec := &statementRecorder{}
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(dashPerfToday), func(w *contractWiring) { w.wrapApp = traced(rec) })
	all := flagtest.Static{}
	for _, m := range dashboardModules {
		all[m] = true
	}
	env.apply(t, dashboardModules, all)
	session := env.signUp(t, signUpEmail, acceptableSample)
	user, token := session.User.Id.String(), session.Tokens.AccessToken
	created := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Rendimiento", perfRegimes...), token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	taxpayer := decodeTaxpayer(t, created).Id.String()
	seedTwoYears(t, env, user, taxpayer)

	target := dashboardTarget(taxpayer, "")
	get := func(string) time.Duration {
		start := time.Now()
		r := env.do(env.send(http.MethodGet, target, "", token))
		elapsed := time.Since(start)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		return elapsed
	}
	for range perfWarmup {
		get(target)
	}

	rec.record(true)
	get(target)
	stmts := rec.record(false)
	explainReadPath(t, env.app, user, stmts)
	var reads int
	for _, s := range stmts {
		if readPathStatement(s.SQL) && !strings.Contains(s.SQL, "rate_limit_take") { // the budget is the router's
			reads++
		}
	}
	assert.Equal(t, dashPerfReads, reads, "statements per request, independent of months")

	sequential := make([]time.Duration, 0, dashPerfSequential)
	for range dashPerfSequential {
		sequential = append(sequential, get(target))
	}
	concurrent := measureConcurrentN(get, target, dashPerfConcurrent)
	t.Logf("getDashboard sequential (n=%d): %s", dashPerfSequential, summarize(sequential))
	t.Logf("getDashboard concurrent (n=%d, %d workers): %s", dashPerfConcurrent, perfWorkers, summarize(concurrent))
	assert.Less(t, percentile(sequential, perfPercentile), dashPerfP95Budget, "sequential p95")
	assert.Less(t, percentile(concurrent, perfPercentile), dashPerfP95Budget, "concurrent p95")
}

// seedTwoYears writes perfIncomes incomes and perfExpenses expenses over
// January 2026..June 2027, then analyzes both tables.
func seedTwoYears(t *testing.T, env contractEnv, user, taxpayer string) {
	t.Helper()
	span := fmt.Sprintf("%% %d", dashPerfSpanDays)
	incomes := strings.Replace(perfIncomesSQL, "% 365", span, 1)
	expenses := strings.Replace(perfExpensesSQL, "% 365", span, 1)
	require.NotEqual(t, perfIncomesSQL, incomes, "the incomes seed spans the two years")
	require.NotEqual(t, perfExpensesSQL, expenses, "the expenses seed spans the two years")
	start := time.Now()
	require.NoError(t, pgtest.AsUser(t.Context(), env.app, user, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), incomes, user, taxpayer); err != nil {
			return err
		}
		_, err := tx.ExecContext(t.Context(), expenses, user, taxpayer)
		return err
	}))
	_, err := env.owner.ExecContext(t.Context(), "ANALYZE incomes; ANALYZE expenses")
	require.NoError(t, err)
	t.Logf("seeded %d incomes and %d expenses over 2026-01..2027-06 in %s", perfIncomes, perfExpenses, time.Since(start).Round(time.Millisecond))
}

// measureConcurrentN issues n requests from perfWorkers goroutines.
func measureConcurrentN(get func(string) time.Duration, target string, n int) []time.Duration {
	out := make([]time.Duration, 0, n)
	for len(out) < n {
		batch := measureConcurrent(get, target)
		out = append(out, batch[:min(len(batch), n-len(out))]...)
	}
	return out
}
