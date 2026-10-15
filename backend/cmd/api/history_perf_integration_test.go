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

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The M8-03/M8-04 performance check (MODULE-history §2.6, US-H10): one
// user whose taxpayer holds three regimes (612, 606, 625) with 50,000
// incomes and 50,000 expenses in 2026 (five times the PRD's 10,000 a
// year), read on 15 January 2027 so getHistory?fiscal_year=2026 lists all
// twelve months, every one of them viewed. It records every statement one
// getHistory issues (a pgx tracer on the cc_app pool), prints EXPLAIN
// (ANALYZE, BUFFERS) of each as cc_app under RLS, pins the statement count,
// compares the one aggregate pass with twelve getTaxpayerEstimate calls,
// then measures getHistory end to end through the router (authentication,
// flag gate, per-user read budget, use case, JSON): p95 < 1 s.
//
// It seeds 100k rows, so it runs only on demand:
//
//	CC_PERF=1 go test -tags=integration -count=1 -run TestPerf_historyReadPath -v ./cmd/api/
const (
	histPerfSequential = 300
	histPerfConcurrent = 200
	histPerfP95Budget  = time.Second
	// histPerfNaive is how many times the twelve-estimates baseline runs.
	histPerfNaive = 10
	// histPerfReads is the read statements of one getHistory: the profile
	// (taxpayer, regimes), one aggregate per table for the fiscal year, the
	// tax inputs (one statement), the regime texts and the year's views
	// (one primary-key range scan). It does not grow with months
	// (MODULE-history §2.6 targets at most 7). The due date's holidays are
	// a boot snapshot (M9-04): no statement here.
	histPerfReads = 7
	// histCountRecords of each table seed the statement-count check.
	histCountRecords = 400
	// histMonths is a past fiscal year's months.
	histMonths = 12
)

// histPerfToday is 15 January 2027: fiscal year 2026 is wholly past.
var histPerfToday = time.Date(2027, time.January, 15, 12, 0, 0, 0, clock.Location())

// historyPerfEnv is a traced contractEnv with one perf taxpayer.
type historyPerfEnv struct {
	historyEnv
	rec            *statementRecorder
	user, taxpayer string
	token          string
	fixture        estimateFixture
	target         string
}

func newHistoryPerfEnv(t *testing.T, records int) historyPerfEnv {
	t.Helper()
	rec := &statementRecorder{}
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(histPerfToday), func(w *contractWiring) { w.wrapApp = traced(rec) })
	env.apply(t, historyModules, allOn(historyModules))
	session := env.signUp(t, signUpEmail, acceptableSample)
	user, token := session.User.Id.String(), session.Tokens.AccessToken
	created := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Historial", perfRegimes...), token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	taxpayer := decodeTaxpayer(t, created).Id.String()
	seedOneYear(t, env, user, taxpayer, records)
	h := historyPerfEnv{
		historyEnv: historyEnv{contractEnv: env, c: contracttest.Load(t)},
		rec:        rec, user: user, taxpayer: taxpayer, token: token,
		fixture: estimateFixture{userID: user, token: token, taxpayer: taxpayer},
		target:  historyTarget(taxpayer, "fiscal_year=2026"),
	}
	for m := 1; m <= histMonths; m++ { // every month viewed: the marker is computed for each
		h.open(t, h.fixture, fmt.Sprintf("2026-%02d", m))
	}
	return h
}

// seedOneYear writes n incomes and n expenses over 2026, then analyzes.
func seedOneYear(t *testing.T, env contractEnv, user, taxpayer string, n int) {
	t.Helper()
	series := func(statement string, rows int) string {
		from := fmt.Sprintf("generate_series(1, %d)", rows)
		require.Contains(t, statement, from)
		return strings.Replace(statement, from, fmt.Sprintf("generate_series(1, %d)", n), 1)
	}
	start := time.Now()
	require.NoError(t, pgtest.AsUser(t.Context(), env.app, user, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), series(perfIncomesSQL, perfIncomes), user, taxpayer); err != nil {
			return err
		}
		_, err := tx.ExecContext(t.Context(), series(perfExpensesSQL, perfExpenses), user, taxpayer)
		return err
	}))
	_, err := env.owner.ExecContext(t.Context(), "ANALYZE incomes; ANALYZE expenses; ANALYZE estimate_views")
	require.NoError(t, err)
	t.Logf("seeded %d incomes and %d expenses over 2026 in %s", n, n, time.Since(start).Round(time.Millisecond))
}

// get times one request of target; it asserts (not requires) so it may run
// on the concurrent workers.
func (h historyPerfEnv) get(t *testing.T) func(string) time.Duration {
	return func(target string) time.Duration {
		start := time.Now()
		r := h.do(h.send(http.MethodGet, target, "", h.token))
		elapsed := time.Since(start)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
		return elapsed
	}
}

func TestPerf_historyReadPath(t *testing.T) {
	if os.Getenv("CC_PERF") == "" {
		t.Skip("set CC_PERF=1 to seed 100k records and measure the history read path")
	}
	h := newHistoryPerfEnv(t, perfIncomes)
	get := h.get(t)
	for range perfWarmup {
		get(h.target)
	}

	stmts := recordOne(h.rec, func() { get(h.target) })
	explainReadPath(t, h.app, h.user, stmts)
	assert.Equal(t, histPerfReads, readCount(stmts), "statements per request, independent of months")

	// The batched path against the naive one: twelve getTaxpayerEstimate
	// calls read the ledger twelve times.
	naive := make([]time.Duration, 0, histPerfNaive)
	for range histPerfNaive {
		var total time.Duration
		for m := 1; m <= histMonths; m++ {
			total += get(estimateTarget(h.taxpayer, fmt.Sprintf("2026-%02d", m)))
		}
		naive = append(naive, total)
	}

	sequential := make([]time.Duration, 0, histPerfSequential)
	for range histPerfSequential {
		sequential = append(sequential, get(h.target))
	}
	concurrent := measureConcurrentN(get, h.target, histPerfConcurrent)
	t.Logf("12 x getTaxpayerEstimate (n=%d): %s", histPerfNaive, summarize(naive))
	t.Logf("getHistory sequential (n=%d): %s", histPerfSequential, summarize(sequential))
	t.Logf("getHistory concurrent (n=%d, %d workers): %s", histPerfConcurrent, perfWorkers, summarize(concurrent))
	assert.Less(t, percentile(sequential, perfPercentile), histPerfP95Budget, "sequential p95")
	assert.Less(t, percentile(concurrent, perfPercentile), histPerfP95Budget, "concurrent p95")
	assert.Less(t, percentile(sequential, perfPercentile), percentile(naive, perfPercentile), "one pass beats twelve")
}

// TestContract_historyStatementCount pins the read statements of one
// getHistory (MODULE-history §2.6, US-H10) in the regular integration
// suite: twelve viewed months, the same count as one.
func TestContract_historyStatementCount(t *testing.T) {
	h := newHistoryPerfEnv(t, histCountRecords)

	for _, target := range []string{h.target, historyTarget(h.taxpayer, "fiscal_year=2027")} {
		stmts := recordOne(h.rec, func() {
			r := h.do(h.send(http.MethodGet, target, "", h.token))
			require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		})
		assert.Equal(t, histPerfReads, readCount(stmts), "%s: statements per request", target)
	}
	stmts := recordOne(h.rec, func() { h.open(t, h.fixture, "2026-06") })
	var writes int
	for _, s := range stmts {
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(s.SQL)), "INSERT INTO ESTIMATE_VIEWS") {
			writes++
		}
	}
	assert.Equal(t, 1, writes, "markHistoryViewed is one upsert")
}
