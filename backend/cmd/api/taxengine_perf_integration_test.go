//go:build integration

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The M5-23 performance check (MODULE-tax-engine §2.10, OQ-T4): one user
// whose taxpayer holds three regimes (612, 606, 625) and has 50,000 incomes
// and 50,000 expenses in fiscal year 2026, about five times the PRD's
// 10,000 records a year. It records every statement the estimate issues
// (a pgx tracer on the cc_app pool), prints EXPLAIN (ANALYZE, BUFFERS) of
// each as cc_app under RLS, then measures getTaxpayerEstimate and
// getEstimateBreakdown end to end through the router (authentication, flag
// gate, per-user read budget, use case, JSON): p95 < 300 ms.
//
// It seeds 100k rows, so it runs only on demand:
//
//	CC_PERF=1 go test -tags=integration -count=1 -run TestPerf_estimateReadPath -v ./cmd/api/
const (
	perfIncomes  = 50_000
	perfExpenses = 50_000
	// perfRequests per operation, sequential then concurrent: within the
	// 600/h per-user read budget (limits.go) with the warm-up.
	perfSequential  = 100
	perfConcurrent  = 80
	perfWorkers     = 4
	perfP95Budget   = 300 * time.Millisecond
	perfWarmup      = 5
	perfPercentile  = 95
	perfPercentBase = 100
)

// perfToday is 20 December 2026: the whole fiscal year is in the past.
var perfToday = time.Date(2026, time.December, 20, 12, 0, 0, 0, clock.Location())

var perfRegimes = []string{
	`{"regime_code":"612","start_period":"2026-01"}`,
	`{"regime_code":"606","start_period":"2026-01","options":{"deduction_mode":"blind_35"}}`,
	`{"regime_code":"625","start_period":"2026-01","options":{"activities":["lodging"]}}`,
}

// perfIncomesSQL writes perfIncomes incomes for user $1 / taxpayer $2 over
// 2026, the three regimes in turn, with withholdings where the
// counterparty allows them; deterministic.
var perfIncomesSQL = fmt.Sprintf(`
WITH g AS (
  SELECT g, (ARRAY['612', '606', '625'])[1 + g %% 3] AS regime, 10000 + (g * 37) %% 500000 AS amount,
         date '2026-01-01' + (g * 7) %% 365 AS collected
  FROM generate_series(1, %d) AS g
), c AS (
  SELECT g, regime, amount, collected,
         CASE regime WHEN '625' THEN 'plataforma' WHEN '606' THEN 'persona_moral'
                     ELSE (ARRAY['persona_moral', 'publico_general'])[1 + g %% 2] END AS counterparty
  FROM g
)
INSERT INTO incomes (user_id, taxpayer_id, regime_code, counterparty_type, amount_cents, iva_rate, iva_cents,
                     isr_withheld_cents, iva_withheld_cents, collected_on, platform_activity)
SELECT $1::uuid, $2::uuid, regime, counterparty, amount, 'rate_16', amount * 16 / 100,
       CASE counterparty WHEN 'publico_general' THEN 0 WHEN 'plataforma' THEN amount * 4 / 100 ELSE amount / 10 END,
       CASE counterparty WHEN 'publico_general' THEN 0 ELSE amount * 16 / 100 / 2 END,
       collected, CASE regime WHEN '625' THEN 'lodging' END
FROM c`, perfIncomes)

// perfExpensesSQL writes perfExpenses expenses of 612 and 606 over 2026;
// one in five is not deductible.
var perfExpensesSQL = fmt.Sprintf(`
WITH g AS (
  SELECT g, (ARRAY['612', '606'])[1 + g %% 2] AS regime, 5000 + (g * 53) %% 200000 AS amount,
         date '2026-01-01' + (g * 11) %% 365 AS paid, g %% 5 <> 0 AS deductible
  FROM generate_series(1, %d) AS g
)
INSERT INTO expenses (user_id, taxpayer_id, regime_code, category_code, amount_cents, iva_rate, iva_cents,
                      payment_method, has_cfdi, deductible, deductible_overridden, cash_over_limit,
                      creditable_iva_cents, isr_withheld_cents, iva_withheld_cents, paid_on)
SELECT $1::uuid, $2::uuid, regime, CASE regime WHEN '612' THEN 'software' ELSE 'mantenimiento' END,
       amount, 'rate_16', amount * 16 / 100, 'card_credit', true, deductible, false, false,
       CASE WHEN deductible THEN amount * 16 / 100 ELSE 0 END, 0, 0, paid
FROM g`, perfExpenses)

// statementRecorder is a pgx tracer that keeps the statements issued while
// it is on.
type statementRecorder struct {
	mu    sync.Mutex
	on    bool
	stmts []pgx.TraceQueryStartData
}

func (r *statementRecorder) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.on {
		r.stmts = append(r.stmts, pgx.TraceQueryStartData{SQL: data.SQL, Args: slices.Clone(data.Args)})
	}
	return ctx
}

func (*statementRecorder) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (r *statementRecorder) record(on bool) []pgx.TraceQueryStartData {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.on = on
	out := r.stmts
	r.stmts = nil
	return out
}

// traced reopens app's database with the recorder as its tracer.
func traced(rec *statementRecorder) func(t *testing.T, app *sql.DB) *sql.DB {
	return func(t *testing.T, app *sql.DB) *sql.DB {
		t.Helper()
		conn, err := app.Conn(t.Context())
		require.NoError(t, err)
		var cfg *pgx.ConnConfig
		require.NoError(t, conn.Raw(func(driverConn any) error {
			cfg = driverConn.(*stdlib.Conn).Conn().Config().Copy()
			return nil
		}))
		require.NoError(t, conn.Close())
		cfg.Tracer = rec
		db := stdlib.OpenDB(*cfg)
		t.Cleanup(func() { _ = db.Close() })
		return db
	}
}

// readPathStatement keeps the estimate's reads: not the session context,
// the transaction control or the rate-limit bucket.
func readPathStatement(sqlText string) bool {
	s := strings.ToUpper(strings.TrimSpace(sqlText))
	if !strings.HasPrefix(s, "SELECT") && !strings.HasPrefix(s, "WITH") {
		return false
	}
	return !strings.Contains(s, "SET_CONFIG") && !strings.Contains(s, "RATE_LIMIT_BUCKETS")
}

func TestPerf_estimateReadPath(t *testing.T) {
	if os.Getenv("CC_PERF") == "" {
		t.Skip("set CC_PERF=1 to seed 100k records and measure the estimate read path")
	}
	rec := &statementRecorder{}
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(perfToday), func(w *contractWiring) { w.wrapApp = traced(rec) })
	modules := []string{"module_auth", "module_taxpayers", "module_tax_engine"}
	env.apply(t, modules, flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_tax_engine": true})
	session := env.signUp(t, signUpEmail, acceptableSample)
	user, token := session.User.Id.String(), session.Tokens.AccessToken
	created := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Rendimiento", perfRegimes...), token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	taxpayer := decodeTaxpayer(t, created).Id.String()

	seedStart := time.Now()
	require.NoError(t, pgtest.AsUser(t.Context(), env.app, user, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(t.Context(), perfIncomesSQL, user, taxpayer); err != nil {
			return err
		}
		_, err := tx.ExecContext(t.Context(), perfExpensesSQL, user, taxpayer)
		return err
	}))
	_, err := env.owner.ExecContext(t.Context(), "ANALYZE incomes; ANALYZE expenses")
	require.NoError(t, err)
	t.Logf("seeded %d incomes and %d expenses for one user in %s", perfIncomes, perfExpenses, time.Since(seedStart).Round(time.Millisecond))

	estimate := estimateTarget(taxpayer, "2026-12")
	breakdown := breakdownTarget(taxpayer, "2026-12", "612")
	get := func(target string) time.Duration {
		start := time.Now()
		r := env.do(env.send(http.MethodGet, target, "", token))
		elapsed := time.Since(start)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String()) // assert: get also runs on the workers
		return elapsed
	}
	for range perfWarmup {
		get(estimate)
	}

	rec.record(true)
	get(estimate)
	explainReadPath(t, env.app, user, rec.record(false))

	for _, target := range []struct{ name, url string }{{"getTaxpayerEstimate", estimate}, {"getEstimateBreakdown", breakdown}} {
		sequential := make([]time.Duration, 0, perfSequential)
		for range perfSequential {
			sequential = append(sequential, get(target.url))
		}
		concurrent := measureConcurrent(get, target.url)
		t.Logf("%s sequential (n=%d): %s", target.name, perfSequential, summarize(sequential))
		t.Logf("%s concurrent (n=%d, %d workers): %s", target.name, perfConcurrent, perfWorkers, summarize(concurrent))
		assert.Lessf(t, percentile(sequential, perfPercentile), perfP95Budget, "%s sequential p95", target.name)
		assert.Lessf(t, percentile(concurrent, perfPercentile), perfP95Budget, "%s concurrent p95", target.name)
	}
}

// measureConcurrent issues perfConcurrent requests from perfWorkers
// goroutines and returns every latency.
func measureConcurrent(get func(string) time.Duration, target string) []time.Duration {
	jobs := make(chan struct{})
	var (
		mu  sync.Mutex
		out = make([]time.Duration, 0, perfConcurrent)
		wg  sync.WaitGroup
	)
	for range perfWorkers {
		wg.Go(func() {
			for range jobs {
				d := get(target)
				mu.Lock()
				out = append(out, d)
				mu.Unlock()
			}
		})
	}
	for range perfConcurrent {
		jobs <- struct{}{}
	}
	close(jobs)
	wg.Wait()
	return out
}

// perfStatementBudgetMS bounds each statement of the read path. With one
// user owning every row, the aggregates scan the whole year: the planner
// rightly picks a sequential scan of the user's 50k rows over the partial
// indexes (selectivity 100 %); the indexes serving a user among many are
// proven by the M3-07/M4-08 ledger EXPLAIN tests.
const perfStatementBudgetMS = 100.0

// explainReadPath prints EXPLAIN (ANALYZE, BUFFERS) of every read the
// estimate issued, as the user under RLS, and bounds each one's execution
// time.
func explainReadPath(t *testing.T, app *sql.DB, user string, stmts []pgx.TraceQueryStartData) {
	t.Helper()
	var reads []pgx.TraceQueryStartData
	for _, s := range stmts {
		if readPathStatement(s.SQL) {
			reads = append(reads, s)
		}
	}
	require.NotEmpty(t, reads, "the estimate issued no read")
	t.Logf("the estimate's read path: %d statements (%d traced in total)", len(reads), len(stmts))
	require.NoError(t, pgtest.AsUser(t.Context(), app, user, func(tx *sql.Tx) error {
		for i, s := range reads {
			plan, err := explain(t.Context(), tx, s)
			if err != nil {
				return fmt.Errorf("statement %d: %w", i+1, err)
			}
			t.Logf("statement %d: %s\n%s", i+1, oneLine(s.SQL), plan)
			var ms float64
			_, err = fmt.Sscanf(plan[strings.LastIndex(plan, "Execution Time:"):], "Execution Time: %f ms", &ms)
			require.NoError(t, err)
			assert.Lessf(t, ms, perfStatementBudgetMS, "statement %d execution time", i+1)
		}
		return nil
	}))
}

func explain(ctx context.Context, tx *sql.Tx, s pgx.TraceQueryStartData) (string, error) {
	rows, err := tx.QueryContext(ctx, "EXPLAIN (ANALYZE, BUFFERS) "+s.SQL, s.Args...)
	if err != nil {
		return "", err
	}
	defer func() { _ = rows.Close() }()
	var lines []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", err
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), rows.Err()
}

func oneLine(s string) string { return strings.Join(strings.Fields(s), " ") }

// percentile is the nearest-rank percentile p of ds.
func percentile(ds []time.Duration, p int) time.Duration {
	sorted := slices.Clone(ds)
	slices.Sort(sorted)
	rank := (p*len(sorted) + perfPercentBase - 1) / perfPercentBase
	return sorted[max(rank-1, 0)]
}

func summarize(ds []time.Duration) string {
	sorted := slices.Clone(ds)
	slices.Sort(sorted)
	return fmt.Sprintf("min %s, p50 %s, p95 %s, p99 %s, max %s",
		sorted[0].Round(time.Microsecond*100), percentile(ds, 50).Round(time.Microsecond*100),
		percentile(ds, perfPercentile).Round(time.Microsecond*100), percentile(ds, 99).Round(time.Microsecond*100),
		sorted[len(sorted)-1].Round(time.Microsecond*100))
}
