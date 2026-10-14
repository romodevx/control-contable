//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of the tax-inputs operations (M5-21) through the real
// composition root over Postgres as cc_app, with module_tax_engine ON and
// OFF. module_auth and module_taxpayers stay ON: they are how the users sign
// up and create their taxpayers (module_taxpayers OFF hiding these routes
// is TestTaxEngineRoutes_needTheTaxpayersFlagToo and the platform sweep).

func taxInputsTarget(taxpayerID, regime string, fiscalYear int) string {
	return fmt.Sprintf("/v1/taxpayers/%s/regimes/%s/tax-inputs?fiscal_year=%d", taxpayerID, regime, fiscalYear)
}

func decodeTaxInputs(t *testing.T, rec *httptest.ResponseRecorder) gen.TaxInputs {
	t.Helper()
	var got gen.TaxInputs
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got), rec.Body.String())
	return got
}

func TestContract_taxInputs(t *testing.T) {
	c := contracttest.Load(t)
	getOp, putOp := contracttest.Op("getTaxInputs"), contracttest.Op("putTaxInputs")

	flagtest.OnOff(t, "module_tax_engine", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		env.apply(t, []string{"module_auth", "module_taxpayers", "module_tax_engine"}, f)
		ana := env.signUp(t, signUpEmail, acceptableSample)
		token := ana.Tokens.AccessToken
		created := env.do(env.send(http.MethodPost, "/v1/taxpayers",
			createJSON("", "fisica", rfcAna, "Ana (negocio)", regimeJSON("612")), token))
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		fisica := decodeTaxpayer(t, created).Id.String()

		if !on {
			flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodGet, taxInputsTarget(fisica, "612", 2026), "", token)))
			flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(http.MethodPut, taxInputsTarget(fisica, "612", 2026), `{"values":[]}`, token)))
			return
		}

		get := func(target, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, getOp, env.handler, env.send(http.MethodGet, target, "", token))
		}
		put := func(target, body, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, putOp, env.handler, env.send(http.MethodPut, target, body, token))
		}

		t.Run("getTaxInputs: the definitions and no values yet", func(t *testing.T) {
			rec := get(taxInputsTarget(fisica, "612", 2026), token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			got := decodeTaxInputs(t, rec)
			assert.Equal(t, "612", got.RegimeCode)
			assert.Equal(t, 2026, got.FiscalYear)
			require.NotEmpty(t, got.Definitions)
			assert.Equal(t, "prior_provisional_payments", got.Definitions[0].Key)
			assert.Equal(t, gen.TaxInputDefKind("money_monthly"), got.Definitions[0].Kind)
			assert.Empty(t, got.Values)
		})

		t.Run("putTaxInputs: replace, then read back with every field", func(t *testing.T) {
			body := `{"values":[{"key":"ptu_paid","value_cents":0},` +
				`{"key":"prior_provisional_payments","month":2,"value_cents":200000},` +
				`{"key":"prior_provisional_payments","month":1,"value_cents":200000,"value_rate":null}]}`
			rec := put(taxInputsTarget(fisica, "612", 2026), body, token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			want := `[{"key":"prior_provisional_payments","month":1,"value_cents":200000,"value_rate":null,"value_bool":null},` +
				`{"key":"prior_provisional_payments","month":2,"value_cents":200000,"value_rate":null,"value_bool":null},` +
				`{"key":"ptu_paid","month":null,"value_cents":0,"value_rate":null,"value_bool":null}]`
			assert.JSONEq(t, want, valuesJSON(t, rec))

			rec = get(taxInputsTarget(fisica, "612", 2026), token)
			require.Equal(t, http.StatusOK, rec.Code)
			assert.JSONEq(t, want, valuesJSON(t, rec))
			assert.Contains(t, env.logs.String(), "tax_inputs.updated", "S-8: the audit event")
			assert.NotContains(t, env.logs.String(), "200000", "S-3: values are never logged")
		})

		t.Run("putTaxInputs: value rules name dotted fields", func(t *testing.T) {
			rec := put(taxInputsTarget(fisica, "612", 2026),
				`{"values":[{"key":"coeficiente_utilidad","value_rate":"0.5"},{"key":"ptu_paid","month":3,"value_cents":1}]}`, token)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, gen.ErrorCodeValidationFailed, decodeError(t, rec).Error.Code)
			assert.Equal(t, []string{"values.0.key/unknown_input", "values.1.month/invalid"}, details(t, rec))

			rec = put(taxInputsTarget(fisica, "612", 2026), `{"values":[{"key":"ptu_paid","value_cents":-1}]}`, token)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "schema: NonNegativeCents")
			rec = put(taxInputsTarget(fisica, "612", 2026), `{"values":[],"extra":1}`, token)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "S-7: closed body")
		})

		t.Run("putTaxInputs: the coeficiente de utilidad is a rate in [0, 1] kept as given", func(t *testing.T) {
			rec := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "moral", rfcMoral, "Empresa", regimeJSON("601")), token))
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			moral := decodeTaxpayer(t, rec).Id.String()

			rec = put(taxInputsTarget(moral, "601", 2026), `{"values":[{"key":"coeficiente_utilidad","value_rate":"1.5"}]}`, token)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"values.0.value_rate/out_of_range"}, details(t, rec))

			rec = put(taxInputsTarget(moral, "601", 2026),
				`{"values":[{"key":"coeficiente_utilidad","value_rate":"0.1500"},{"key":"first_fiscal_year","value_bool":false}]}`, token)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.JSONEq(t, `[{"key":"coeficiente_utilidad","month":null,"value_cents":null,"value_rate":"0.1500","value_bool":null},`+
				`{"key":"first_fiscal_year","month":null,"value_cents":null,"value_rate":null,"value_bool":false}]`, valuesJSON(t, rec))
		})

		t.Run("tax inputs: fiscal year and regime rules", func(t *testing.T) {
			rec := get(taxInputsTarget(fisica, "626", 2026), token)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"regimeCode/regime_not_active"}, details(t, rec), "a regime the taxpayer does not hold")
			rec = put(taxInputsTarget(fisica, "626", 2026), `{"values":[]}`, token)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)

			rec = get(taxInputsTarget(fisica, "612", 2100), token)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Equal(t, []string{"fiscal_year/out_of_range"}, details(t, rec), "a future year")

			rec = get(taxInputsTarget(fisica, "612", 2021), token)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "schema: FiscalYear.minimum")
		})

		t.Run("tax inputs: another user's taxpayer is 404 on both operations", func(t *testing.T) {
			bea := env.signUp(t, beaEmail, otherSample).Tokens.AccessToken
			rec := get(taxInputsTarget(fisica, "612", 2026), bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = put(taxInputsTarget(fisica, "612", 2026), `{"values":[]}`, bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = get(taxInputsTarget(unknownTaxpayer, "612", 2026), token)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("tax inputs: no session is 401", func(t *testing.T) {
			assert.Equal(t, http.StatusUnauthorized, get(taxInputsTarget(fisica, "612", 2026), "").Code)
			assert.Equal(t, http.StatusUnauthorized, put(taxInputsTarget(fisica, "612", 2026), `{"values":[]}`, "").Code)
		})

		t.Run("putTaxInputs: an archived taxpayer is 422, still readable", func(t *testing.T) {
			rec := env.do(env.send(http.MethodPost, "/v1/taxpayers/"+fisica+"/archive", "", token))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			rec = put(taxInputsTarget(fisica, "612", 2026), `{"values":[]}`, token)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, details(t, rec))
			rec = get(taxInputsTarget(fisica, "612", 2026), token)
			assert.Equal(t, http.StatusOK, rec.Code)
		})

		t.Run("putTaxInputs: a read-only user gets 402; GET still works", func(t *testing.T) {
			env.makeReadOnly(t, ana.User.Id.String())

			rec := put(taxInputsTarget(fisica, "612", 2026), `{"values":[]}`, token)
			assert.Equal(t, http.StatusPaymentRequired, rec.Code)
			assert.Equal(t, gen.ErrorCodeEntitlementRequired, decodeError(t, rec).Error.Code)
			rec = get(taxInputsTarget(fisica, "612", 2026), token)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}, "module_auth", "module_taxpayers")
}

// valuesJSON returns the raw "values" array of a TaxInputs body.
func valuesJSON(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Values json.RawMessage `json:"values"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
	return string(body.Values)
}
