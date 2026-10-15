//go:build integration

package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// The annual performance check (MODULE-annual §2.11, ANN-15): a persona
// física with three regimes (612, 606, 625), 10,000 records of fiscal year
// 2026 and 200 personal deductions, the year closed and nothing filed, so
// every read also runs the 12-month Estimator range for the default
// payments. It prints EXPLAIN (ANALYZE, BUFFERS) of every statement the
// estimate issues, as cc_app under RLS, and measures getAnnualEstimate and
// getAnnualBreakdown end to end through the router: p95 < 500 ms. It is
// small enough to run in CI on every push (no CC_PERF switch).
const (
	annualPerfIncomes    = 6_000
	annualPerfExpenses   = 4_000
	annualPerfDeductions = 200
	// annualPerfRequests per operation: each spends two estimate reads
	// (the range), well inside the 600/h budget with the warm-up.
	annualPerfRequests = 40
	annualPerfP95      = 500 * time.Millisecond
)

// annualPerfToday is 10 February 2027: fiscal year 2026 is closed.
var annualPerfToday = time.Date(2027, time.February, 10, 12, 0, 0, 0, clock.Location())

// annualPerfIncomesSQL writes n incomes of user $1 / taxpayer $2 over 2026,
// the three regimes in turn; deterministic.
func annualPerfIncomesSQL(n int) string {
	return fmt.Sprintf(`
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
FROM c`, n)
}

// annualPerfExpensesSQL writes n expenses of 612 and 606 over 2026.
func annualPerfExpensesSQL(n int) string {
	return fmt.Sprintf(`
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
FROM g`, n)
}

// annualPerfDeductionsSQL writes n medical deductions of 2026, one in ten
// paid in cash (excluded).
func annualPerfDeductionsSQL(n int) string {
	return fmt.Sprintf(`
INSERT INTO personal_deductions (user_id, taxpayer_id, category, amount_cents, paid_on, payment_method, beneficiary)
SELECT $1::uuid, $2::uuid, 'medical', 50000 + (g * 97) %% 100000, date '2026-01-01' + (g * 13) %% 365,
       CASE WHEN g %% 10 = 0 THEN 'cash' ELSE 'electronic' END, 'self'
FROM generate_series(1, %d) AS g`, n)
}

func TestPerf_annualReadPath(t *testing.T) {
	rec := &statementRecorder{}
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(annualPerfToday), func(w *contractWiring) { w.wrapApp = traced(rec) })
	env.registerAnnualFlag(t)
	env.apply(t, annualModules, annualAllOn())
	session := env.signUp(t, signUpEmail, acceptableSample)
	user, token := session.User.Id.String(), session.Tokens.AccessToken
	created := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Rendimiento", perfRegimes...), token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	taxpayer := decodeTaxpayer(t, created).Id.String()

	seedStart := time.Now()
	require.NoError(t, pgtest.AsUser(t.Context(), env.app, user, func(tx *sql.Tx) error {
		for _, stmt := range []string{
			annualPerfIncomesSQL(annualPerfIncomes), annualPerfExpensesSQL(annualPerfExpenses), annualPerfDeductionsSQL(annualPerfDeductions),
		} {
			if _, err := tx.ExecContext(t.Context(), stmt, user, taxpayer); err != nil {
				return err
			}
		}
		return nil
	}))
	_, err := env.owner.ExecContext(t.Context(), "ANALYZE incomes; ANALYZE expenses; ANALYZE personal_deductions")
	require.NoError(t, err)
	t.Logf("seeded %d incomes, %d expenses and %d personal deductions in %s",
		annualPerfIncomes, annualPerfExpenses, annualPerfDeductions, time.Since(seedStart).Round(time.Millisecond))

	estimate := annualEstimateTarget(taxpayer, 2026)
	breakdown := estimate + "/breakdown?section=pf_art152"
	get := func(target string) time.Duration {
		start := time.Now()
		r := env.do(env.send(http.MethodGet, target, "", token))
		elapsed := time.Since(start)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		return elapsed
	}
	for range perfWarmup {
		get(estimate)
	}

	rec.record(true)
	get(estimate)
	explainReadPath(t, env.app, user, rec.record(false))

	for _, target := range []struct{ name, url string }{{"getAnnualEstimate", estimate}, {"getAnnualBreakdown", breakdown}} {
		latencies := make([]time.Duration, 0, annualPerfRequests)
		for range annualPerfRequests {
			latencies = append(latencies, get(target.url))
		}
		t.Logf("ANNUAL-PERF %s (n=%d, %d records, %d deductions, range on): %s",
			target.name, annualPerfRequests, annualPerfIncomes+annualPerfExpenses, annualPerfDeductions, summarize(latencies))
		assert.Lessf(t, percentile(latencies, perfPercentile), annualPerfP95, "%s p95", target.name)
	}
}
