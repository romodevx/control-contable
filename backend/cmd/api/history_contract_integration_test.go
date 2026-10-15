//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
)

// Contract tests of getHistory and markHistoryViewed (M8-04) through the
// real composition root over Postgres as cc_app, with module_history ON
// and OFF, its prerequisites (module_tax_engine, module_taxpayers) OFF,
// and the record modules OFF. "Today" is estToday (20 October 2026): the
// 2026 list is October .. January. The records are TX-NET-01's (January ..
// March, 612 + 606).

// historyModules are the flags these suites set.
var historyModules = []string{
	"module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine", "module_history",
}

func historyTarget(taxpayerID, query string) string {
	target := "/v1/taxpayers/" + taxpayerID + "/history"
	if query != "" {
		target += "?" + query
	}
	return target
}

func viewedTarget(taxpayerID, period string) string {
	return "/v1/taxpayers/" + taxpayerID + "/history/" + period + "/viewed"
}

// historyMonthBody is one month of the 200 body, the parts these tests read.
type historyMonthBody struct {
	Period     string `json:"period"`
	Status     string `json:"status"`
	HasRecords bool   `json:"has_records"`
	Tax        *struct {
		TotalPayableRoundedCents int64 `json:"total_payable_rounded_cents"`
		ISRPayableRoundedCents   int64 `json:"isr_payable_rounded_cents"`
		IVAPayableRoundedCents   int64 `json:"iva_payable_rounded_cents"`
	} `json:"tax"`
	DueDate            string          `json:"due_date"`
	IsCurrent          bool            `json:"is_current"`
	RecordsFingerprint string          `json:"records_fingerprint"`
	FiguresFingerprint string          `json:"figures_fingerprint"`
	EngineVersion      string          `json:"engine_version"`
	ParameterSet       json.RawMessage `json:"parameter_set"`
	ViewedAt           *string         `json:"viewed_at"`
	Changed            bool            `json:"changed_since_viewed"`
	Reason             *string         `json:"change_reason"`
}

type historyBody struct {
	TaxpayerID string             `json:"taxpayer_id"`
	FiscalYear int                `json:"fiscal_year"`
	Months     []historyMonthBody `json:"months"`
	YearTotals struct {
		TotalPayableRoundedCents int64 `json:"total_payable_rounded_cents"`
	} `json:"year_totals"`
	Legend      string            `json:"legend"`
	Disclaimers []json.RawMessage `json:"disclaimers"`
}

func decodeHistory(t *testing.T, rec *httptest.ResponseRecorder) historyBody {
	t.Helper()
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var body historyBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

func (b historyBody) month(t *testing.T, period string) historyMonthBody {
	t.Helper()
	for _, m := range b.Months {
		if m.Period == period {
			return m
		}
	}
	require.Failf(t, "missing month", "%s", period)
	return historyMonthBody{}
}

// estimateSeen is the estimate body's part the app posts to markHistoryViewed.
type estimateSeen struct {
	RecordsFingerprint string          `json:"records_fingerprint"`
	FiguresFingerprint string          `json:"figures_fingerprint"`
	EngineVersion      string          `json:"engine_version"`
	ParameterSet       json.RawMessage `json:"parameter_set"`
	Totals             struct {
		TotalPayableRoundedCents int64 `json:"total_payable_rounded_cents"`
	} `json:"totals"`
	DueDate string `json:"due_date"`
}

func (s estimateSeen) body(t *testing.T) string {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"records_fingerprint": s.RecordsFingerprint, "figures_fingerprint": s.FiguresFingerprint,
		"engine_version": s.EngineVersion, "parameter_set": s.ParameterSet,
	})
	require.NoError(t, err)
	return string(raw)
}

// historyEnv is a contractEnv with every module ON and Ana's and Bea's
// TX-NET-01 taxpayers.
type historyEnv struct {
	contractEnv
	c        *contracttest.Contract
	ana, bea estimateFixture
}

func newHistoryEnv(t *testing.T) historyEnv {
	t.Helper()
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.apply(t, historyModules, allOn(historyModules))
	return historyEnv{
		contractEnv: env, c: contracttest.Load(t),
		ana: env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna),
		bea: env.txNet01Taxpayer(t, beaEmail, otherSample, "AAAA010101AA0"),
	}
}

func allOn(modules []string) flagtest.Static {
	f := flagtest.Static{}
	for _, m := range modules {
		f[m] = true
	}
	return f
}

func (e historyEnv) history(t *testing.T, f estimateFixture, query string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("getHistory"), e.handler, e.send(http.MethodGet, historyTarget(f.taxpayer, query), "", f.token))
}

func (e historyEnv) markViewed(t *testing.T, f estimateFixture, taxpayer, period, body string) *httptest.ResponseRecorder {
	t.Helper()
	return e.c.ServeOp(t, contracttest.Op("markHistoryViewed"), e.handler, e.send(http.MethodPost, viewedTarget(taxpayer, period), body, f.token))
}

// estimate is getTaxpayerEstimate's body for period, as the app renders it.
func (e historyEnv) estimate(t *testing.T, f estimateFixture, period string) estimateSeen {
	t.Helper()
	rec := e.c.ServeOp(t, contracttest.Op("getTaxpayerEstimate"), e.handler, e.send(http.MethodGet, estimateTarget(f.taxpayer, period), "", f.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var s estimateSeen
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &s))
	return s
}

// open is the app opening a month: render the estimate, then mark it viewed.
func (e historyEnv) open(t *testing.T, f estimateFixture, period string) {
	t.Helper()
	rec := e.markViewed(t, f, f.taxpayer, period, e.estimate(t, f, period).body(t))
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}

// lateIncome is a 612 income collected in March, entered after March was viewed.
func (e historyEnv) lateIncome(t *testing.T, f estimateFixture) {
	t.Helper()
	(&estimateEnv{app: e.app}).seed(t, f.userID, f.taxpayer, []estIncome{
		{regime: "612", counterparty: "publico_general", rate: "rate_16", amount: 10_000_07, iva: 1_600_01, on: "2026-03-25"},
	}, nil)
}

func (e historyEnv) deleteLateIncome(t *testing.T, f estimateFixture) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, f.userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), "UPDATE incomes SET deleted_at = now() WHERE taxpayer_id = $1 AND amount_cents = 1000007", f.taxpayer)
		return err
	}))
}

func TestContract_history(t *testing.T) {
	c := contracttest.Load(t)

	flagtest.OnOff(t, "module_history", func(t *testing.T, f flagtest.Static, on bool) {
		env := newHistoryEnv(t)
		env.apply(t, historyModules, f)
		ana := env.ana
		if !on {
			for _, req := range []*http.Request{
				env.send(http.MethodGet, historyTarget(ana.taxpayer, "fiscal_year=2026"), "", ana.token),
				env.send(http.MethodPost, viewedTarget(ana.taxpayer, "2026-03"), env.estimate(t, ana, "2026-03").body(t), ana.token),
				// OFF answers 404 before any input check, a malformed request included.
				env.send(http.MethodGet, historyTarget("not-a-uuid", "fiscal_year=dos"), "", ana.token),
				env.send(http.MethodPost, viewedTarget("not-a-uuid", "2026-13"), `{"x":1}`, ana.token),
			} {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, req))
			}
			return
		}

		t.Run("the current year, newest first, each month the estimate's totals", func(t *testing.T) {
			rec := env.history(t, ana, "fiscal_year=2026")
			body := decodeHistory(t, rec)
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Equal(t, ana.taxpayer, body.TaxpayerID)
			assert.Equal(t, 2026, body.FiscalYear)
			require.Len(t, body.Months, 10, "October .. January")
			var sum int64
			for i, m := range body.Months {
				assert.Equal(t, fmt.Sprintf("2026-%02d", 10-i), m.Period, "newest first")
				assert.Equal(t, i == 0, m.IsCurrent, m.Period)
				assert.Nil(t, m.ViewedAt, "never viewed")
				assert.False(t, m.Changed)
				require.NotNil(t, m.Tax, m.Period)
				est := env.estimate(t, ana, m.Period)
				assert.Equal(t, est.Totals.TotalPayableRoundedCents, m.Tax.TotalPayableRoundedCents, "%s: to the centavo", m.Period)
				assert.Equal(t, est.FiguresFingerprint, m.FiguresFingerprint, m.Period)
				assert.Equal(t, est.RecordsFingerprint, m.RecordsFingerprint, m.Period)
				assert.Equal(t, est.EngineVersion, m.EngineVersion, m.Period)
				assert.JSONEq(t, string(est.ParameterSet), string(m.ParameterSet), m.Period)
				assert.Equal(t, est.DueDate, m.DueDate, m.Period)
				sum += m.Tax.TotalPayableRoundedCents
			}
			assert.Equal(t, sum, body.YearTotals.TotalPayableRoundedCents, "Suma de estimados mensuales")
			assert.True(t, body.month(t, "2026-03").HasRecords)
			assert.False(t, body.month(t, "2026-07").HasRecords, "Sin registros en este mes")
			assert.Equal(t, "Esto es un aproximado. No todos los gastos aplican para tu giro comercial.", body.Legend)
			assert.NotEmpty(t, body.Disclaimers)
		})

		t.Run("viewed, then a late income marks it (records), deleting it clears it", func(t *testing.T) {
			env.open(t, ana, "2026-03")
			env.open(t, ana, "2026-04")
			march := decodeHistory(t, env.history(t, ana, "fiscal_year=2026")).month(t, "2026-03")
			require.NotNil(t, march.ViewedAt)
			assert.False(t, march.Changed)
			assert.Nil(t, march.Reason)

			env.lateIncome(t, ana)
			body := decodeHistory(t, env.history(t, ana, "fiscal_year=2026"))
			for _, p := range []string{"2026-03", "2026-04"} {
				m := body.month(t, p)
				assert.True(t, m.Changed, "%s: 612 ISR is cumulative", p)
				require.NotNil(t, m.Reason, p)
				assert.Equal(t, "records", *m.Reason, p)
			}
			assert.False(t, body.month(t, "2026-05").Changed, "never viewed")
			assert.False(t, body.month(t, "2026-10").Changed, "the current month")

			env.open(t, ana, "2026-03")
			body = decodeHistory(t, env.history(t, ana, "fiscal_year=2026"))
			assert.False(t, body.month(t, "2026-03").Changed, "opening clears it")
			assert.True(t, body.month(t, "2026-04").Changed, "until April is opened")

			env.deleteLateIncome(t, ana)
			body = decodeHistory(t, env.history(t, ana, "fiscal_year=2026"))
			assert.False(t, body.month(t, "2026-04").Changed, "figures back to the viewed ones")
			assert.True(t, body.month(t, "2026-03").Changed, "March was viewed with the income")
			env.open(t, ana, "2026-03")
		})

		t.Run("an engine or parameter change that moved the figures is the engine reason", func(t *testing.T) {
			seen := env.estimate(t, ana, "2026-02")
			seen.EngineVersion, seen.FiguresFingerprint = "0.9.9", strings.Repeat("0", 64)
			require.Equal(t, http.StatusNoContent, env.markViewed(t, ana, ana.taxpayer, "2026-02", seen.body(t)).Code)

			feb := decodeHistory(t, env.history(t, ana, "fiscal_year=2026")).month(t, "2026-02")
			assert.True(t, feb.Changed)
			require.NotNil(t, feb.Reason)
			assert.Equal(t, "engine", *feb.Reason)

			same := env.estimate(t, ana, "2026-02")
			same.EngineVersion, same.RecordsFingerprint = "0.9.9", strings.Repeat("0", 64)
			require.Equal(t, http.StatusNoContent, env.markViewed(t, ana, ana.taxpayer, "2026-02", same.body(t)).Code)
			feb = decodeHistory(t, env.history(t, ana, "fiscal_year=2026")).month(t, "2026-02")
			assert.False(t, feb.Changed, "same figures: an engine or input change alone never marks (OQ-H4)")
		})

		t.Run("the current month is stored but never marked", func(t *testing.T) {
			seen := env.estimate(t, ana, "2026-10")
			seen.FiguresFingerprint = strings.Repeat("0", 64)
			require.Equal(t, http.StatusNoContent, env.markViewed(t, ana, ana.taxpayer, "2026-10", seen.body(t)).Code)
			oct := decodeHistory(t, env.history(t, ana, "fiscal_year=2026")).month(t, "2026-10")
			require.NotNil(t, oct.ViewedAt)
			assert.False(t, oct.Changed)
		})

		t.Run("marking is idempotent", func(t *testing.T) {
			body := env.estimate(t, ana, "2026-01").body(t)
			for range 2 {
				assert.Equal(t, http.StatusNoContent, env.markViewed(t, ana, ana.taxpayer, "2026-01", body).Code)
			}
		})

		t.Run("400 and 422", func(t *testing.T) {
			rec := env.history(t, ana, "fiscal_year=2027")
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"fiscal_year/out_of_range"}, details(t, rec))
			rec = env.history(t, ana, "fiscal_year=2025")
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"fiscal_year/parameters_unavailable"}, details(t, rec))
			for _, query := range []string{"", "fiscal_year=dos", "fiscal_year=2026&period=2026-03"} {
				assert.Equal(t, http.StatusBadRequest, env.history(t, ana, query).Code, query)
			}
			assert.Equal(t, http.StatusBadRequest, env.c.ServeOp(t, contracttest.Op("getHistory"), env.handler,
				env.send(http.MethodGet, historyTarget("not-a-uuid", "fiscal_year=2026"), "", ana.token)).Code)

			good := env.estimate(t, ana, "2026-03")
			rec = env.markViewed(t, ana, ana.taxpayer, "2026-11", good.body(t))
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"period/out_of_range"}, details(t, rec))
			rec = env.markViewed(t, ana, ana.taxpayer, "2025-12", good.body(t))
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"period/parameters_unavailable"}, details(t, rec))
			bad := good
			bad.FiguresFingerprint = strings.ToUpper(good.FiguresFingerprint)
			badVersion := good
			badVersion.EngineVersion = "1.0 1"
			for name, body := range map[string]string{
				"malformed fingerprint": bad.body(t),
				"malformed version":     badVersion.body(t),
				"unknown field":         strings.Replace(good.body(t), "{", `{"amount_cents":1,`, 1),
				"missing field":         `{"records_fingerprint":"` + good.RecordsFingerprint + `"}`,
				"no body":               "",
			} {
				assert.Equal(t, http.StatusBadRequest, env.markViewed(t, ana, ana.taxpayer, "2026-03", body).Code, name)
			}
			assert.Equal(t, http.StatusBadRequest, env.markViewed(t, ana, ana.taxpayer, "2026-3", good.body(t)).Code)
		})

		t.Run("another user's or an unknown taxpayer is the hidden 404", func(t *testing.T) {
			bea := env.bea
			body := env.estimate(t, ana, "2026-03").body(t)
			for _, rec := range []*httptest.ResponseRecorder{
				env.history(t, estimateFixture{token: bea.token, taxpayer: ana.taxpayer}, "fiscal_year=2026"),
				env.history(t, estimateFixture{token: ana.token, taxpayer: unknownTaxpayer}, "fiscal_year=2026"),
				env.markViewed(t, bea, ana.taxpayer, "2026-03", body),
				env.markViewed(t, ana, unknownTaxpayer, "2026-03", body),
			} {
				flagtest.AssertHidden(t, rec)
				assert.Equal(t, "Not found.", decodeError(t, rec).Error.Message, "indistinguishable from a module OFF")
			}
			own := decodeHistory(t, env.history(t, bea, "fiscal_year=2026"))
			assert.Nil(t, own.month(t, "2026-03").ViewedAt, "RLS: Bea never sees Ana's views")
		})

		t.Run("no session is 401", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, env.history(t, estimateFixture{taxpayer: ana.taxpayer}, "fiscal_year=2026").Code)
			assert.Equal(t, http.StatusUnauthorized, env.markViewed(t, estimateFixture{}, ana.taxpayer, "2026-03", "{}").Code)
		})

		t.Run("an archived taxpayer and a read-only user still read and mark", func(t *testing.T) {
			rec := env.do(env.send(http.MethodPost, "/v1/taxpayers/"+ana.taxpayer+"/archive", "", ana.token))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			env.makeReadOnly(t, ana.userID)

			assert.Len(t, decodeHistory(t, env.history(t, ana, "fiscal_year=2026")).Months, 10)
			env.open(t, ana, "2026-05")
		})

		t.Run("amounts and fingerprints are never logged", func(t *testing.T) {
			logs := env.logs.String()
			march := env.estimate(t, ana, "2026-03")
			assert.NotContains(t, logs, march.FiguresFingerprint)
			assert.NotContains(t, logs, strconv.FormatInt(march.Totals.TotalPayableRoundedCents, 10))
		})
	}, "module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine")
}

// MODULE-history §3: (history, tax_engine) ∈ {ON/ON, ON/OFF, OFF/ON}, the
// taxpayers prerequisite, and the record modules: OFF never changes a
// figure, only has_records.
func TestContract_historyFlagMatrix(t *testing.T) {
	env := newHistoryEnv(t)
	ana := env.ana
	get := func() *httptest.ResponseRecorder {
		return env.c.Serve(t, env.handler, env.send(http.MethodGet, historyTarget(ana.taxpayer, "fiscal_year=2026"), "", ana.token))
	}
	mark := func() *httptest.ResponseRecorder {
		return env.c.Serve(t, env.handler, env.send(http.MethodPost, viewedTarget(ana.taxpayer, "2026-03"),
			`{"records_fingerprint":"`+strings.Repeat("a", 64)+`","figures_fingerprint":"`+strings.Repeat("b", 64)+
				`","engine_version":"1.0.1","parameter_set":{"fiscal_year":2026,"version":1}}`, ana.token))
	}
	all := decodeHistory(t, get())

	for _, tc := range []struct {
		name    string
		off     []string
		visible bool
	}{
		{"history and engine ON", nil, true},
		{"engine OFF implies history OFF", []string{"module_tax_engine"}, false},
		{"history OFF", []string{"module_history"}, false},
		{"taxpayers OFF implies both OFF", []string{"module_taxpayers"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := allOn(historyModules)
			for _, m := range tc.off {
				f[m] = false
			}
			env.apply(t, historyModules, f)
			if !tc.visible {
				flagtest.AssertHidden(t, get())
				flagtest.AssertHidden(t, mark())
				return
			}
			assert.Equal(t, http.StatusOK, get().Code)
			assert.Equal(t, http.StatusNoContent, mark().Code)
		})
	}

	for _, tc := range []struct {
		name              string
		incomes, expenses bool
		marchHasRecords   bool
	}{
		{"incomes OFF", false, true, true},  // March has expenses
		{"expenses OFF", true, false, true}, // and incomes
		{"both OFF", false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := allOn(historyModules)
			f["module_incomes"], f["module_expenses"] = tc.incomes, tc.expenses
			env.apply(t, historyModules, f)

			body := decodeHistory(t, env.c.ServeOp(t, contracttest.Op("getHistory"), env.handler,
				env.send(http.MethodGet, historyTarget(ana.taxpayer, "fiscal_year=2026"), "", ana.token)))

			assert.Equal(t, tc.marchHasRecords, body.month(t, "2026-03").HasRecords)
			for i, m := range body.Months {
				assert.Equal(t, all.Months[i].Tax, m.Tax, "%s: the figures never depend on the record flags", m.Period)
				assert.Equal(t, all.Months[i].FiguresFingerprint, m.FiguresFingerprint, m.Period)
			}
		})
	}
}

// §2.7: markHistoryViewed has its own budget (history_views, burst 120);
// spending it never touches getHistory's (history_reads).
func TestContract_historyRateLimits(t *testing.T) {
	env := newHistoryEnv(t)
	ana := env.ana
	body := env.estimate(t, ana, "2026-03").body(t)
	for i := range historyViewBurst {
		rec := env.do(env.send(http.MethodPost, viewedTarget(ana.taxpayer, "2026-03"), body, ana.token))
		require.Equal(t, http.StatusNoContent, rec.Code, "view %d", i+1)
	}

	rec := env.markViewed(t, ana, ana.taxpayer, "2026-03", body)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("Retry-After"))
	assert.Equal(t, "rate_limited", string(decodeError(t, rec).Error.Code))
	assert.Equal(t, http.StatusOK, env.history(t, ana, "fiscal_year=2026").Code, "another budget")
	assert.Equal(t, http.StatusNoContent, env.markViewed(t, env.bea, env.bea.taxpayer, "2026-03",
		env.estimate(t, env.bea, "2026-03").body(t)).Code, "another user's budget")
}

// historyViewBurst is history_views' burst (MODULE-history §2.7).
const historyViewBurst = 120
