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
)

// Contract tests of the WH-07 contract step (0.10.1, ADR-0008): derived
// supplier withholdings on createExpense / updateExpense and
// previewExpenseWithholdings, flag ON and OFF, through the real composition
// root over Postgres as cc_app.

const (
	withholdingTaxpayerPM = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7c01" // Ana's persona moral: 601 since 2026-01
	withholdingTaxpayerPF = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7c02" // Ana's física: 612 since 2026-01
)

// usW6Body is US-W6 (WM WH-11): the persona moral pays a 612 consultant
// 40,000.00 + IVA 6,400.00 on 2026-06-15; the operation comes from the
// category. fields override members ("-name" drops one).
func usW6Body(t *testing.T, id string, fields map[string]any) string {
	t.Helper()
	body := map[string]any{
		"regime_code": "601", "category_code": "servicios_profesionales", "amount_cents": 4000000,
		"iva_rate": "rate_16", "iva_cents": 640000, "payment_method": "transfer", "paid_on": "2026-06-15",
		"has_cfdi": true, "supplier_kind": "pf_with_activity", "supplier_regime": "612",
	}
	for k, v := range fields {
		if name, drop := strings.CutPrefix(k, "-"); drop {
			delete(body, name)
			continue
		}
		body[k] = v
	}
	return expenseJSON(t, id, body)
}

func previewTarget(taxpayerID, query string) string {
	return expensesPath(taxpayerID, "/withholding-preview?", query)
}

// usW6Query is US-W6 as previewExpenseWithholdings' query.
const usW6Query = "regime=601&on=2026-06-15&amount_cents=4000000&iva_rate=rate_16&iva_cents=640000" +
	"&category_code=servicios_profesionales&supplier_kind=pf_with_activity&supplier_regime=612"

func (e contractEnv) createWithholdingTaxpayer(t *testing.T, c *contracttest.Contract, token, id, person, rfc, regime string) {
	t.Helper()
	body := fmt.Sprintf(`{"id":%q,"person_type":%q,"rfc":%q,"display_name":"Retenciones","regimes":[{"regime_code":%q,"start_period":"2026-01"}]}`,
		id, person, rfc, regime)
	rec := c.ServeOp(t, contracttest.Op("createTaxpayer"), e.handler, e.send(http.MethodPost, "/v1/taxpayers", body, token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
}

func decodePreview(t *testing.T, rec *httptest.ResponseRecorder) gen.WithholdingPreview {
	t.Helper()
	var p gen.WithholdingPreview
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p), rec.Body.String())
	return p
}

func TestContract_expenseWithholdings(t *testing.T) {
	c := contracttest.Load(t)
	modules := []string{"module_auth", "module_taxpayers", "module_expenses"}

	flagtest.OnOff(t, "module_expenses", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		f["module_auth"], f["module_taxpayers"] = true, true
		env.apply(t, modules, f)
		preview := contracttest.Op("previewExpenseWithholdings")

		if !on {
			flagtest.AssertHidden(t, c.Serve(t, env.handler,
				env.send(http.MethodGet, previewTarget(withholdingTaxpayerPM, usW6Query), "", "any-token")))
			return
		}

		ana := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken
		bea := env.signUp(t, beaEmail, otherSample).Tokens.AccessToken
		env.createWithholdingTaxpayer(t, c, ana, withholdingTaxpayerPM, "moral", rfcMoral, "601")
		env.createWithholdingTaxpayer(t, c, ana, withholdingTaxpayerPF, "fisica", rfcAna, "612")
		serve := func(op contracttest.Operation, method, target, body, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, op, env.handler, env.send(method, target, body, token))
		}
		create, update, get := contracttest.Op("createExpense"), contracttest.Op("updateExpense"), contracttest.Op("getExpense")

		t.Run("listExpenseCategories: the E3 default for a persona moral only (PR #137)", func(t *testing.T) {
			categories := contracttest.Op("listExpenseCategories")
			defaults := func(taxpayer, regime string) map[string]*gen.OperationKind {
				t.Helper()
				rec := serve(categories, http.MethodGet, "/v1/taxpayers/"+taxpayer+"/expense-categories?regime="+regime+"&on=2026-09-12", "", ana)
				require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
				var body gen.ExpenseCategoryList
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				out := map[string]*gen.OperationKind{}
				for _, item := range body.Items {
					out[item.Code] = item.DefaultOperationKind
				}
				return out
			}
			moral := defaults(withholdingTaxpayerPM, "601")
			require.NotNil(t, moral["servicios_profesionales"])
			assert.Equal(t, gen.OperationKind("professional_services"), *moral["servicios_profesionales"])
			require.NotNil(t, moral["fletes"])
			assert.Equal(t, gen.OperationKind("ground_freight"), *moral["fletes"])
			for code, kind := range moral {
				assert.NotNil(t, kind, "%s: every persona moral category has its default", code)
			}
			for code, kind := range defaults(withholdingTaxpayerPF, "612") {
				assert.Nil(t, kind, "%s: a persona física is never asked", code)
			}
		})

		t.Run("US-W6: createExpense derives, stores and serves the withholding", func(t *testing.T) {
			rec := serve(create, http.MethodPost, expensesPath(withholdingTaxpayerPM), usW6Body(t, clientExpenseID(1), nil), ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			e := decodeExpense(t, rec)
			w := e.Withholding
			assert.EqualValues(t, 400000, w.IsrCents)
			assert.EqualValues(t, 426667, w.IvaCents)
			assert.Equal(t, "0.1", w.IsrRate)
			assert.Equal(t, gen.WithholderYou, w.Withholder)
			assert.EqualValues(t, 3813333, w.NetCents)
			require.NotNil(t, w.RemitDueOn)
			assert.Equal(t, "2026-07-17", w.RemitDueOn.String())
			assert.Equal(t, []string{"ISR-5", "IVA-6"}, w.RuleIds)
			assert.Len(t, w.Explanations, 2)
			assert.Equal(t, "2026.1", w.TableVersion)
			assert.EqualValues(t, 400000, e.IsrWithheldCents, "the read-only mirror")
			require.NotNil(t, e.OperationKind)
			assert.Equal(t, gen.OperationKind("professional_services"), *e.OperationKind, "the category's default")
			assert.Equal(t, []string{"deduction_requires_remittance@withholding"}, expenseWarnings(e))
			assert.Equal(t, "Este gasto solo es deducible si enteras las retenciones a más tardar el 17 de julio.",
				e.Warnings[0].Message)

			rec = serve(get, http.MethodGet, expensesPath(withholdingTaxpayerPM, "/", clientExpenseID(1)), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, e.Withholding, decodeExpense(t, rec).Withholding, "reads serve what the write stored")

			rec = serve(create, http.MethodPost, expensesPath(withholdingTaxpayerPM), usW6Body(t, clientExpenseID(1), nil), ana)
			assert.Equal(t, http.StatusOK, rec.Code, "an idempotent retry with the same facts")
		})

		t.Run("updateExpense re-derives: a persona moral supplier withholds nothing", func(t *testing.T) {
			rec := serve(update, http.MethodPatch, expensesPath(withholdingTaxpayerPM, "/", clientExpenseID(1)),
				`{"supplier_kind":"pm"}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			e := decodeExpense(t, rec)
			assert.Nil(t, e.SupplierRegime, "moving off pf_with_activity clears the regime")
			assert.EqualValues(t, 0, e.Withholding.IsrCents+e.Withholding.IvaCents)
			assert.Nil(t, e.Withholding.RemitDueOn)
			assert.Empty(t, e.Warnings)
		})

		t.Run("US-W8: typed amounts are 400 additionalProperties on create and update", func(t *testing.T) {
			rec := serve(create, http.MethodPost, expensesPath(withholdingTaxpayerPM),
				usW6Body(t, "", map[string]any{"isr_withheld_cents": 0}), ana)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"isr_withheld_cents/additionalProperties"}, details(t, rec))

			rec = serve(update, http.MethodPatch, expensesPath(withholdingTaxpayerPM, "/", clientExpenseID(1)),
				`{"iva_withheld_cents":426667}`, ana)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"iva_withheld_cents/additionalProperties"}, details(t, rec))
		})

		t.Run("supplier refusals: 400 required, 422 supplier_not_applicable and residential_rent_exempt", func(t *testing.T) {
			refusals := []struct {
				name, taxpayer, body string
				status               int
				details              []string
			}{
				{"a persona moral without a supplier", withholdingTaxpayerPM,
					usW6Body(t, "", map[string]any{"-supplier_kind": nil, "-supplier_regime": nil}),
					http.StatusBadRequest, []string{"supplier_kind/required"}},
				{"US-W7: a persona física names a supplier", withholdingTaxpayerPF,
					expenseJSON(t, "", map[string]any{"paid_on": "2026-06-15", "supplier_kind": "pm"}),
					http.StatusUnprocessableEntity, []string{"supplier_kind/supplier_not_applicable"}},
				{"a regime for a persona moral supplier", withholdingTaxpayerPM,
					usW6Body(t, "", map[string]any{"supplier_kind": "pm"}),
					http.StatusUnprocessableEntity, []string{"supplier_regime/supplier_not_applicable"}},
				{"housing rent with IVA", withholdingTaxpayerPM,
					usW6Body(t, "", map[string]any{"residential_rent": true}),
					http.StatusUnprocessableEntity, []string{"iva_rate/residential_rent_exempt"}},
			}
			for _, r := range refusals {
				rec := serve(create, http.MethodPost, expensesPath(r.taxpayer), r.body, ana)
				require.Equal(t, r.status, rec.Code, "%s: %s", r.name, rec.Body.String())
				assert.Equal(t, r.details, details(t, rec), r.name)
			}
		})

		t.Run("previewExpenseWithholdings: 200 stores nothing; 400, 422, 404, 401", func(t *testing.T) {
			rec := serve(preview, http.MethodGet, previewTarget(withholdingTaxpayerPM, usW6Query), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			p := decodePreview(t, rec)
			assert.EqualValues(t, 400000, p.Withholding.IsrCents)
			assert.EqualValues(t, 426667, p.Withholding.IvaCents)
			assert.EqualValues(t, 3813333, p.Withholding.NetCents)
			require.Len(t, p.Warnings, 1)
			assert.Equal(t, gen.WarningCode("deduction_requires_remittance"), p.Warnings[0].Code)

			rec = serve(preview, http.MethodGet, previewTarget(withholdingTaxpayerPM,
				strings.Replace(usW6Query, "&supplier_regime=612", "", 1)), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, gen.WarningCode("supplier_regime_unknown"), decodePreview(t, rec).Warnings[0].Code)

			refusals := []struct {
				name, taxpayer, query string
				status                int
				details               []string
			}{
				{"on before the first parameter set", withholdingTaxpayerPM,
					strings.Replace(usW6Query, "on=2026-06-15", "on=2025-06-15", 1), http.StatusBadRequest, []string{"on/out_of_range"}},
				{"an unknown category", withholdingTaxpayerPM,
					strings.Replace(usW6Query, "servicios_profesionales", "yates", 1), http.StatusBadRequest,
					[]string{"category_code/unknown_category"}},
				{"a persona moral without a supplier", withholdingTaxpayerPM,
					strings.Replace(usW6Query, "&supplier_kind=pf_with_activity&supplier_regime=612", "", 1),
					http.StatusBadRequest, []string{"supplier_kind/required"}},
				{"a regime not held", withholdingTaxpayerPM, strings.Replace(usW6Query, "regime=601", "regime=626", 1),
					http.StatusUnprocessableEntity, []string{"regime/regime_not_active"}},
				{"a persona física names a supplier", withholdingTaxpayerPF,
					strings.Replace(usW6Query, "regime=601", "regime=612", 1), http.StatusUnprocessableEntity,
					[]string{"supplier_kind/supplier_not_applicable", "supplier_regime/supplier_not_applicable"}},
			}
			for _, r := range refusals {
				rec := serve(preview, http.MethodGet, previewTarget(r.taxpayer, r.query), "", ana)
				require.Equal(t, r.status, rec.Code, "%s: %s", r.name, rec.Body.String())
				assert.Equal(t, r.details, details(t, rec), r.name)
			}

			rec = serve(preview, http.MethodGet, previewTarget(withholdingTaxpayerPM, usW6Query), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "another user's taxpayer")
			rec = serve(preview, http.MethodGet, previewTarget(withholdingTaxpayerPM, usW6Query), "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)

			rec = serve(contracttest.Op("listExpenses"), http.MethodGet, expensesPath(withholdingTaxpayerPM, "?period=2026-06"), "", ana)
			require.Equal(t, http.StatusOK, rec.Code)
			var page gen.ExpensePage
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			assert.Len(t, page.Items, 1, "previews store nothing")
		})
	})
}
