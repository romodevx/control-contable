//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Contract tests of the estimate and breakdown operations (M5-23) through
// the real composition root over Postgres as cc_app, with module_tax_engine
// ON and OFF, and with its prerequisite module_taxpayers OFF. The estimate's
// "today" is estToday (20 October 2026); the records are TX-NET-01's
// (MODULE-tax-engine §3), entered as SQL rows under the user's RLS context
// and, for the prior payments, through putTaxInputs.

func estimateTarget(taxpayerID, period string) string {
	if period == "" {
		return "/v1/taxpayers/" + taxpayerID + "/estimates"
	}
	return "/v1/taxpayers/" + taxpayerID + "/estimates?period=" + period
}

func breakdownTarget(taxpayerID, period, regime string) string {
	return fmt.Sprintf("/v1/taxpayers/%s/estimates/%s/breakdown?regime=%s", taxpayerID, period, regime)
}

// txNet01Regimes is TX-NET-01's taxpayer: 612 and 606 (blind deduction)
// since January 2026.
var txNet01Regimes = []string{
	`{"regime_code":"612","start_period":"2026-01"}`,
	`{"regime_code":"606","start_period":"2026-01","options":{"deduction_mode":"blind_35"}}`,
}

// estimateFixture is a signed-up user with TX-NET-01's taxpayer.
type estimateFixture struct {
	userID, token, taxpayer string
}

func (e contractEnv) txNet01Taxpayer(t *testing.T, email, phrase, rfc string) estimateFixture {
	t.Helper()
	session := e.signUp(t, email, phrase)
	f := estimateFixture{userID: session.User.Id.String(), token: session.Tokens.AccessToken}
	rec := e.do(e.send(http.MethodPost, "/v1/taxpayers", createJSON("", "fisica", rfc, "Estimación", txNet01Regimes...), f.token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	f.taxpayer = decodeTaxpayer(t, rec).Id.String()
	(&estimateEnv{app: e.app}).txNet01Records(t, f.userID, f.taxpayer)
	return f
}

// priorPayments stores 612's January and February prior payments through
// the API.
func (e contractEnv) priorPayments(t *testing.T, f estimateFixture) {
	t.Helper()
	rec := e.do(e.send(http.MethodPut, taxInputsTarget(f.taxpayer, "612", 2026),
		`{"values":[{"key":"prior_provisional_payments","month":1,"value_cents":200000},`+
			`{"key":"prior_provisional_payments","month":2,"value_cents":200000}]}`, f.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

// withoutIdentity drops the fields that name this run's taxpayer, records
// and engine build, so a body compares with the contract example (whose
// engine_version is illustrative: assertEngineVersion checks it). Regime names are
// the catalog's (the SAT spelling, migration 20261013180000), which the
// examples abbreviate: they are compared on their own.
// assertEngineVersion checks that a body names the engine that computed it.
func assertEngineVersion(t *testing.T, body []byte) {
	t.Helper()
	var head struct {
		EngineVersion string `json:"engine_version"`
	}
	require.NoError(t, json.Unmarshal(body, &head))
	assert.Equal(t, domain.EngineVersion, head.EngineVersion)
}

func withoutIdentity(t *testing.T, body string) map[string]json.RawMessage {
	t.Helper()
	var m map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(body), &m), body)
	delete(m, "taxpayer_id")
	delete(m, "records_fingerprint")
	delete(m, "figures_fingerprint")
	delete(m, "engine_version")
	if raw, ok := m["regimes"]; ok {
		var regimes []map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &regimes))
		for _, r := range regimes {
			delete(r, "regime_name")
		}
		m["regimes"] = mustJSON(t, regimes)
	}
	if raw, ok := m["regime"]; ok {
		var regime map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(raw, &regime))
		delete(regime, "regime_name")
		m["regime"] = mustJSON(t, regime)
	}
	return m
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	return raw
}

// regimeNames are the regime_name values of an estimate body, in order.
func regimeNames(t *testing.T, rec *httptest.ResponseRecorder) []string {
	t.Helper()
	var body gen.TaxpayerEstimate
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	var names []string
	for _, r := range body.Regimes {
		names = append(names, r.RegimeName)
	}
	return names
}

// exampleOf is the named 200 example of operationID in the served contract.
func exampleOf(t *testing.T, operationID, name string) string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	for _, item := range spec.Paths.Map() {
		for _, op := range item.Operations() {
			if op.OperationID != operationID {
				continue
			}
			ex := op.Responses.Status(http.StatusOK).Value.Content.Get("application/json").Examples[name]
			require.NotNilf(t, ex, "%s has no example %q", operationID, name)
			raw, err := json.Marshal(ex.Value.Value)
			require.NoError(t, err)
			return string(raw)
		}
	}
	t.Fatalf("no operation %s", operationID)
	return ""
}

func TestContract_estimates(t *testing.T) {
	c := contracttest.Load(t)
	estimateOp, breakdownOp := contracttest.Op("getTaxpayerEstimate"), contracttest.Op("getEstimateBreakdown")

	flagtest.OnOff(t, "module_tax_engine", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
		env.apply(t, []string{"module_auth", "module_taxpayers", "module_tax_engine"}, f)
		ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)

		if !on {
			flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, estimateTarget(ana.taxpayer, "2026-03"), "", ana.token)))
			flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, breakdownTarget(ana.taxpayer, "2026-03", "612"), "", ana.token)))
			return
		}
		env.priorPayments(t, ana)

		estimate := func(target, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, estimateOp, env.handler, env.send(http.MethodGet, target, "", token))
		}
		breakdown := func(target, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, breakdownOp, env.handler, env.send(http.MethodGet, target, "", token))
		}

		t.Run("getTaxpayerEstimate: TX-NET-01 is the contract's crossRegimeIva example", func(t *testing.T) {
			rec := estimate(estimateTarget(ana.taxpayer, "2026-03"), ana.token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"), "S-4")
			assertEngineVersion(t, rec.Body.Bytes())
			got := withoutIdentity(t, rec.Body.String())
			want := withoutIdentity(t, exampleOf(t, "getTaxpayerEstimate", "crossRegimeIva"))
			for field, value := range want {
				assert.JSONEqf(t, string(value), string(got[field]), "field %s", field)
			}
			assert.Len(t, got, len(want))
			assert.Equal(t, []string{"Arrendamiento", "Personas Físicas con Actividades Empresariales y Profesionales"},
				regimeNames(t, rec), "the catalog's names")
		})

		t.Run("getTaxpayerEstimate: no period is the current month", func(t *testing.T) {
			rec := estimate(estimateTarget(ana.taxpayer, ""), ana.token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var body gen.TaxpayerEstimate
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "2026-10", body.Period)
			assert.Equal(t, gen.Legend("Esto es un aproximado. No todos los gastos aplican para tu giro comercial."), body.Legend)
		})

		t.Run("getEstimateBreakdown: the 612 trace, then the consolidated IVA", func(t *testing.T) {
			rec := breakdown(breakdownTarget(ana.taxpayer, "2026-03", "612"), ana.token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"), "S-4")
			assertEngineVersion(t, rec.Body.Bytes())
			got := withoutIdentity(t, rec.Body.String())
			want := withoutIdentity(t, exampleOf(t, "getEstimateBreakdown", "crossRegimeIva612"))
			for _, field := range []string{"period", "regime", "legend", "disclaimers"} {
				assert.JSONEqf(t, string(want[field]), string(got[field]), "field %s", field)
			}
			var lines []gen.BreakdownLine
			require.NoError(t, json.Unmarshal(got["lines"], &lines))
			require.NotEmpty(t, lines)
			assert.Equal(t, "Ingresos cobrados acumulados ene–mar", lines[0].Concept)
			assert.Equal(t, int64(9_000_000), lines[0].AmountCents)
			last := lines[len(lines)-1]
			assert.Equal(t, "IVA a cargo del mes (todos tus regímenes)", last.Concept)
			assert.Equal(t, int64(20_000), last.AmountCents)
			assert.True(t, last.NeedsCpaReview, "two regimes contribute IVA")
			for i, l := range lines {
				assert.Equal(t, i+1, l.Seq)
				assert.NotContains(t, l.Concept, "isr.", "a concept, not a key")
				assert.NotEqual(t, "Paso del cálculo", l.Concept)
				if l.Operation != "info" && l.Operation != "rounding_adjustment" {
					assert.NotEmptyf(t, l.LegalBasis, "line %d (%s) cites its basis", l.Seq, l.Concept)
				}
			}
		})

		t.Run("getEstimateBreakdown: a regime not held in the month is 422", func(t *testing.T) {
			rec := breakdown(breakdownTarget(ana.taxpayer, "2026-03", "626"), ana.token)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, gen.ErrorCodeBusinessRuleViolation, decodeError(t, rec).Error.Code)
			assert.Equal(t, []string{"regime/regime_not_active"}, details(t, rec))
		})

		t.Run("estimates: a future month is 400 out_of_range, before 2026 is 422 parameters_unavailable", func(t *testing.T) {
			for _, rec := range []*httptest.ResponseRecorder{
				estimate(estimateTarget(ana.taxpayer, "2026-11"), ana.token),
				breakdown(breakdownTarget(ana.taxpayer, "2026-11", "612"), ana.token),
			} {
				assert.Equal(t, http.StatusBadRequest, rec.Code)
				assert.Equal(t, []string{"period/out_of_range"}, details(t, rec))
			}
			for _, rec := range []*httptest.ResponseRecorder{
				estimate(estimateTarget(ana.taxpayer, "2025-12"), ana.token),
				breakdown(breakdownTarget(ana.taxpayer, "2025-12", "612"), ana.token),
			} {
				assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
				assert.Equal(t, []string{"period/parameters_unavailable"}, details(t, rec))
			}
			assert.Equal(t, http.StatusBadRequest, estimate(estimateTarget(ana.taxpayer, "2026-13"), ana.token).Code, "schema: Period")
			assert.Equal(t, http.StatusBadRequest, breakdown(breakdownTarget(ana.taxpayer, "2026-03", "61"), ana.token).Code,
				"schema: RegimeCode")
		})

		t.Run("estimates: another user's or an unknown taxpayer is 404", func(t *testing.T) {
			bea := env.txNet01Taxpayer(t, beaEmail, otherSample, "AAAA010101AA0")
			for _, rec := range []*httptest.ResponseRecorder{
				estimate(estimateTarget(ana.taxpayer, "2026-03"), bea.token),
				breakdown(breakdownTarget(ana.taxpayer, "2026-03", "612"), bea.token),
				estimate(estimateTarget(unknownTaxpayer, "2026-03"), ana.token),
				breakdown(breakdownTarget(unknownTaxpayer, "2026-03", "612"), ana.token),
			} {
				assert.Equal(t, http.StatusNotFound, rec.Code)
				assert.Equal(t, gen.ErrorCodeNotFound, decodeError(t, rec).Error.Code)
			}
			rec := estimate(estimateTarget(bea.taxpayer, "2026-03"), bea.token)
			require.Equal(t, http.StatusOK, rec.Code)
			var own gen.TaxpayerEstimate
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &own))
			assert.Equal(t, int64(415_096), own.Totals.Isr.AmountCents,
				"S-2: Bea's own records only (no prior payments of hers), never Ana's inputs")
		})

		t.Run("estimates: no session is 401", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, estimate(estimateTarget(ana.taxpayer, "2026-03"), "").Code)
			assert.Equal(t, http.StatusUnauthorized, breakdown(breakdownTarget(ana.taxpayer, "2026-03", "612"), "").Code)
		})

		t.Run("estimates: an archived taxpayer and a read-only user still read", func(t *testing.T) {
			rec := env.do(env.send(http.MethodPost, "/v1/taxpayers/"+ana.taxpayer+"/archive", "", ana.token))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			env.makeReadOnly(t, ana.userID)

			assert.Equal(t, http.StatusOK, estimate(estimateTarget(ana.taxpayer, "2026-03"), ana.token).Code)
			assert.Equal(t, http.StatusOK, breakdown(breakdownTarget(ana.taxpayer, "2026-03", "606"), ana.token).Code)
		})

		t.Run("estimates: amounts are never logged (S-3)", func(t *testing.T) {
			assert.False(t, strings.Contains(env.logs.String(), "415096") || strings.Contains(env.logs.String(), "35096"))
		})
	}, "module_auth", "module_taxpayers")
}

// MODULE-tax-engine §7: module_tax_engine ON with its prerequisite
// module_taxpayers OFF hides both operations (implied OFF).
func TestContract_estimatesNeedTheTaxpayersFlag(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	modules := []string{"module_auth", "module_taxpayers", "module_tax_engine"}
	env.apply(t, modules, flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_tax_engine": true})
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)

	env.apply(t, modules, flagtest.Static{"module_auth": true, "module_taxpayers": false, "module_tax_engine": true})

	flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, estimateTarget(ana.taxpayer, "2026-03"), "", ana.token)))
	flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, breakdownTarget(ana.taxpayer, "2026-03", "612"), "", ana.token)))
}
