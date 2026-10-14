//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Contract tests of getDashboard (M6-03) through the real composition root
// over Postgres as cc_app, with module_dashboard ON and OFF, every
// combination of the engine and the record modules, and the prerequisite
// module_taxpayers OFF. "Today" is estToday (20 October 2026): the default
// window is 2026-01..2026-10. The records are TX-NET-01's (January..March)
// plus one October income, one soft-deleted income and Bea's own records.

// dashboardModules are the flags these suites set.
var dashboardModules = []string{
	"module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine", "module_dashboard",
}

func dashboardTarget(taxpayerID, query string) string {
	target := "/v1/taxpayers/" + taxpayerID + "/dashboard"
	if query != "" {
		target += "?" + query
	}
	return target
}

// dashboardBody is the 200 body with the parts these tests read.
type dashboardBody struct {
	TaxpayerID string `json:"taxpayer_id"`
	Period     string `json:"period"`
	IsEmpty    bool   `json:"is_empty"`
	Month      struct {
		IncomeCents  *int64 `json:"income_cents"`
		ExpenseCents *int64 `json:"expense_cents"`
	} `json:"month"`
	Estimate json.RawMessage `json:"estimate"`
	Trend    []struct {
		Period       string `json:"period"`
		IncomeCents  *int64 `json:"income_cents"`
		ExpenseCents *int64 `json:"expense_cents"`
		Tax          *struct {
			TotalPayableRoundedCents int64  `json:"total_payable_rounded_cents"`
			IVAOutcome               string `json:"iva_outcome"`
			Status                   string `json:"status"`
		} `json:"tax"`
	} `json:"trend"`
	Access struct {
		Access           string `json:"access"`
		TrialDaysLeft    *int   `json:"trial_days_left"`
		IncomesRemaining *int   `json:"incomes_remaining"`
	} `json:"access"`
	Legend *string `json:"legend"`
}

func decodeDashboard(t *testing.T, rec *httptest.ResponseRecorder) dashboardBody {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body dashboardBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

// dashboardRecords adds, to TX-NET-01, an October income, a soft-deleted
// March income and, for Bea, her own March income: none but the first may
// ever show in Ana's figures.
func (e contractEnv) dashboardRecords(t *testing.T, ana, bea estimateFixture) {
	t.Helper()
	env := &estimateEnv{app: e.app}
	env.seed(t, ana.userID, ana.taxpayer, []estIncome{
		{regime: "612", counterparty: "publico_general", rate: "rate_16", amount: 1_234_56, iva: 197_53, on: "2026-10-05"},
		{regime: "612", counterparty: "publico_general", rate: "rate_16", amount: 777_00, iva: 124_32, on: "2026-03-11"},
	}, nil)
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, ana.userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), "UPDATE incomes SET deleted_at = now() WHERE taxpayer_id = $1 AND amount_cents = 77700", ana.taxpayer)
		return err
	}))
	env.seed(t, bea.userID, bea.taxpayer, []estIncome{
		{regime: "612", counterparty: "publico_general", rate: "rate_16", amount: 9_999_00, iva: 1_599_84, on: "2026-03-12"},
	}, nil)
}

// txNet01Monthly are TX-NET-01's incomes collected and expenses paid per
// month, plus the October income: the exact figures every view must show.
var txNet01Monthly = map[string][2]int64{
	"2026-01": {30_000_00, 10_000_00},
	"2026-02": {30_000_00, 10_000_00},
	"2026-03": {50_000_00, 28_750_00},
	"2026-10": {1_234_56, 0},
}

func assertExactTrend(t *testing.T, body dashboardBody, incomes, expenses bool) {
	t.Helper()
	require.Len(t, body.Trend, 10, "2026-01..2026-10")
	for i, p := range body.Trend {
		assert.Equal(t, fmt.Sprintf("2026-%02d", i+1), p.Period, "oldest first, one per month")
		want := txNet01Monthly[p.Period]
		if incomes {
			require.NotNil(t, p.IncomeCents, p.Period)
			assert.Equal(t, want[0], *p.IncomeCents, "%s income collected", p.Period)
		} else {
			assert.Nil(t, p.IncomeCents, p.Period)
		}
		if expenses {
			require.NotNil(t, p.ExpenseCents, p.Period)
			assert.Equal(t, want[1], *p.ExpenseCents, "%s expenses paid", p.Period)
		} else {
			assert.Nil(t, p.ExpenseCents, p.Period)
		}
	}
	assert.Equal(t, body.Trend[9].IncomeCents, body.Month.IncomeCents, "month is October's")
	assert.Equal(t, body.Trend[9].ExpenseCents, body.Month.ExpenseCents)
}

func TestContract_dashboard(t *testing.T) {
	c := contracttest.Load(t)
	op := contracttest.Op("getDashboard")
	estimateOp := contracttest.Op("getTaxpayerEstimate")

	flagtest.OnOff(t, "module_dashboard", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
		env.apply(t, dashboardModules, f)
		ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
		bea := env.txNet01Taxpayer(t, beaEmail, otherSample, "AAAA010101AA0")
		env.dashboardRecords(t, ana, bea)

		if !on {
			flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, dashboardTarget(ana.taxpayer, ""), "", ana.token)))
			// OFF answers 404 before any input check, a malformed request included.
			flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, dashboardTarget("not-a-uuid", "months=99"), "", ana.token)))
			return
		}
		get := func(target, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, op, env.handler, env.send(http.MethodGet, target, "", token))
		}

		t.Run("exact sums of the default window, engine ON", func(t *testing.T) {
			rec := get(dashboardTarget(ana.taxpayer, ""), ana.token)
			body := decodeDashboard(t, rec)
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Equal(t, "2026-10", body.Period, "no period is the current month in Mexico City")
			assert.Equal(t, ana.taxpayer, body.TaxpayerID)
			assert.False(t, body.IsEmpty)
			assertExactTrend(t, body, true, true)
			require.NotNil(t, body.Legend)
			assert.Equal(t, "Esto es un aproximado. No todos los gastos aplican para tu giro comercial.", *body.Legend)
			assert.Equal(t, "trial", body.Access.Access)
			require.NotNil(t, body.Access.TrialDaysLeft, "server-computed (OQ-D6)")
			assert.GreaterOrEqual(t, *body.Access.TrialDaysLeft, 1, "the trial started at sign-up; the dashboard's today is pinned")
			for _, p := range body.Trend {
				require.NotNil(t, p.Tax, p.Period)
				assert.NotEqual(t, "none", p.Tax.Status, "%s: both regimes are held since January", p.Period)
			}
		})

		t.Run("the estimate is getTaxpayerEstimate's, and each month's tax its totals", func(t *testing.T) {
			body := decodeDashboard(t, get(dashboardTarget(ana.taxpayer, "period=2026-03"), ana.token))
			estimate := c.ServeOp(t, estimateOp, env.handler, env.send(http.MethodGet, estimateTarget(ana.taxpayer, "2026-03"), "", ana.token))
			require.Equal(t, http.StatusOK, estimate.Code)
			assert.JSONEq(t, estimate.Body.String(), string(body.Estimate), "same body to the centavo")
			var totals struct {
				Totals struct {
					TotalPayableRoundedCents int64 `json:"total_payable_rounded_cents"`
				} `json:"totals"`
			}
			require.NoError(t, json.Unmarshal(estimate.Body.Bytes(), &totals))
			require.Len(t, body.Trend, 3)
			assert.Equal(t, totals.Totals.TotalPayableRoundedCents, body.Trend[2].Tax.TotalPayableRoundedCents)
			for _, p := range body.Trend[:2] {
				month := c.ServeOp(t, estimateOp, env.handler, env.send(http.MethodGet, estimateTarget(ana.taxpayer, p.Period), "", ana.token))
				require.NoError(t, json.Unmarshal(month.Body.Bytes(), &totals))
				assert.Equal(t, totals.Totals.TotalPayableRoundedCents, p.Tax.TotalPayableRoundedCents, p.Period)
			}
		})

		t.Run("months=6 is the last six months", func(t *testing.T) {
			body := decodeDashboard(t, get(dashboardTarget(ana.taxpayer, "months=6"), ana.token))
			require.Len(t, body.Trend, 6)
			assert.Equal(t, "2026-05", body.Trend[0].Period)
			assert.Equal(t, "2026-10", body.Trend[5].Period)
		})

		t.Run("400 and 422", func(t *testing.T) {
			rec := get(dashboardTarget(ana.taxpayer, "period=2026-11"), ana.token)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"period/out_of_range"}, details(t, rec))
			rec = get(dashboardTarget(ana.taxpayer, "period=2025-12"), ana.token)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"period/parameters_unavailable"}, details(t, rec))
			for _, target := range []string{
				dashboardTarget(ana.taxpayer, "period=2026-13"),
				dashboardTarget(ana.taxpayer, "months=13"),
				dashboardTarget(ana.taxpayer, "months=5"),
				dashboardTarget(ana.taxpayer, "months=doce"),
				dashboardTarget(ana.taxpayer, "fiscal_year=2026"),
				dashboardTarget("not-a-uuid", ""),
			} {
				assert.Equal(t, http.StatusBadRequest, get(target, ana.token).Code, target)
			}
		})

		t.Run("another user's or an unknown taxpayer is the hidden 404", func(t *testing.T) {
			for _, rec := range []*httptest.ResponseRecorder{
				get(dashboardTarget(ana.taxpayer, ""), bea.token),
				get(dashboardTarget(unknownTaxpayer, ""), ana.token),
			} {
				flagtest.AssertHidden(t, rec)
				assert.Equal(t, "Not found.", decodeError(t, rec).Error.Message, "indistinguishable from a module OFF")
			}
			own := decodeDashboard(t, get(dashboardTarget(bea.taxpayer, "period=2026-03"), bea.token))
			assert.Equal(t, int64(50_000_00+9_999_00), *own.Month.IncomeCents,
				"RLS: Bea sees her own records only")
		})

		t.Run("no session is 401", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, get(dashboardTarget(ana.taxpayer, ""), "").Code)
		})

		t.Run("an archived taxpayer and a read-only user still read", func(t *testing.T) {
			rec := env.do(env.send(http.MethodPost, "/v1/taxpayers/"+ana.taxpayer+"/archive", "", ana.token))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			env.makeReadOnly(t, ana.userID)

			body := decodeDashboard(t, get(dashboardTarget(ana.taxpayer, ""), ana.token))
			assert.Equal(t, "read_only", body.Access.Access)
			require.NotNil(t, body.Access.IncomesRemaining)
			assert.Equal(t, 0, *body.Access.IncomesRemaining)
			assert.Nil(t, body.Access.TrialDaysLeft)
			assertExactTrend(t, body, true, true)
		})

		t.Run("amounts are never logged", func(t *testing.T) {
			assert.False(t, strings.Contains(env.logs.String(), "123456") || strings.Contains(env.logs.String(), "5000000"))
		})
	}, "module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine")
}

// MODULE-dashboard §3: the engine and the record modules degrade the
// response; nothing is computed, nor sent, for a module that is OFF.
func TestContract_dashboardFlagMatrix(t *testing.T) {
	c := contracttest.Load(t)
	op := contracttest.Op("getDashboard")
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	all := flagtest.Static{}
	for _, m := range dashboardModules {
		all[m] = true
	}
	env.apply(t, dashboardModules, all)
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
	bea := env.txNet01Taxpayer(t, beaEmail, otherSample, "AAAA010101AA0")
	env.dashboardRecords(t, ana, bea)

	for _, tc := range []struct{ tax, incomes, expenses bool }{
		{true, true, true}, {true, false, true}, {true, true, false}, {true, false, false},
		{false, true, true}, {false, false, true}, {false, true, false}, {false, false, false},
	} {
		t.Run(fmt.Sprintf("tax=%v incomes=%v expenses=%v", tc.tax, tc.incomes, tc.expenses), func(t *testing.T) {
			f := flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_dashboard": true,
				"module_tax_engine": tc.tax, "module_incomes": tc.incomes, "module_expenses": tc.expenses}
			env.apply(t, dashboardModules, f)

			body := decodeDashboard(t, c.ServeOp(t, op, env.handler, env.send(http.MethodGet, dashboardTarget(ana.taxpayer, ""), "", ana.token)))

			assertExactTrend(t, body, tc.incomes, tc.expenses)
			assert.Equal(t, tc.tax, string(body.Estimate) != "null", "estimate iff the engine is ON")
			assert.Equal(t, tc.tax, body.Legend != nil, "legend iff estimate")
			for _, p := range body.Trend {
				assert.Equal(t, tc.tax, p.Tax != nil, "%s tax iff the engine is ON", p.Period)
			}
			assert.False(t, body.IsEmpty, "records exist (or no record module is ON)")
		})
	}

	t.Run("taxpayers OFF hides the dashboard (prerequisite)", func(t *testing.T) {
		env.apply(t, dashboardModules, flagtest.Static{"module_auth": true, "module_dashboard": true, "module_tax_engine": true})
		flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, dashboardTarget(ana.taxpayer, ""), "", ana.token)))
	})
}

// US-D5: a taxpayer with no live record is the empty state; a deleted
// record does not count.
func TestContract_dashboardEmptyState(t *testing.T) {
	c := contracttest.Load(t)
	op := contracttest.Op("getDashboard")
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	all := flagtest.Static{}
	for _, m := range dashboardModules {
		all[m] = true
	}
	env.apply(t, dashboardModules, all)
	session := env.signUp(t, signUpEmail, acceptableSample)
	user, token := session.User.Id.String(), session.Tokens.AccessToken
	rec := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfcAna, "Vacío", txNet01Regimes...), token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	taxpayer := decodeTaxpayer(t, rec).Id.String()
	get := func() dashboardBody {
		t.Helper()
		return decodeDashboard(t, c.ServeOp(t, op, env.handler, env.send(http.MethodGet, dashboardTarget(taxpayer, ""), "", token)))
	}

	body := get()
	assert.True(t, body.IsEmpty)
	assert.Equal(t, int64(0), *body.Month.IncomeCents, "a point of zeros, never a hole")

	(&estimateEnv{app: env.app}).seed(t, user, taxpayer, []estIncome{
		{regime: "612", counterparty: "publico_general", rate: "rate_16", amount: 100_00, iva: 16_00, on: "2026-02-01"},
	}, nil)
	assert.False(t, get().IsEmpty)
	require.NoError(t, pgtest.AsUser(t.Context(), env.app, user, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), "UPDATE incomes SET deleted_at = now() WHERE taxpayer_id = $1", taxpayer)
		return err
	}))
	assert.True(t, get().IsEmpty, "deleting the only record shows the empty state again")
}
