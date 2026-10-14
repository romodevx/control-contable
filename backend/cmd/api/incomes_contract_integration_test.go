//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// Contract tests of the 7 module_incomes operations (M3-06) through the real
// composition root over Postgres as cc_app, with the flag ON and OFF.
// module_auth and module_taxpayers stay ON: they are how the users sign up
// and get a taxpayer.

const (
	incomeTaxpayerA  = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7c01" // Ana's física: 612 since 2025-01 (backdated), 625 (lodging) since 2026-01
	incomeTaxpayerB  = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7c02" // Bea's física: 612 since 2026-01
	incomeTaxpayerPM = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7c03" // Ana's moral: 601 since 2025-01 (backdated; accrual)
	incomeFolio      = "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F"
	incomeNote       = "Diseño de logotipo"
)

// clientIncomeID is the n-th fixture income id.
func clientIncomeID(n int) string { return fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", n) }

// incomeOps is one request per module_incomes operation, for the OFF sweep;
// the token is a placeholder (the gate answers first).
var incomeOps = []struct{ method, target, body string }{
	{http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09"), ""},
	{http.MethodPost, incomesPath(incomeTaxpayerA), `{}`},
	{http.MethodGet, incomesPath(incomeTaxpayerA, "/defaults?regime=612&counterparty_type=persona_moral"), ""},
	{http.MethodGet, incomesPath(incomeTaxpayerA, "/", clientIncomeID(1)), ""},
	{http.MethodPatch, incomesPath(incomeTaxpayerA, "/", clientIncomeID(1)), `{"description":"x"}`},
	{http.MethodDelete, incomesPath(incomeTaxpayerA, "/", clientIncomeID(1)), ""},
	{http.MethodPost, incomesPath(incomeTaxpayerA, "/", clientIncomeID(1), "/restore"), ""},
}

func incomesPath(taxpayerID string, rest ...string) string {
	return "/v1/taxpayers/" + taxpayerID + "/incomes" + strings.Join(rest, "")
}

// incomeJSON is a create body: US-I1, 10,000.00 + IVA from a persona moral
// to a 612 física with the suggested withholdings. fields override or add
// top-level members ("-name" drops one).
func incomeJSON(t *testing.T, id string, fields map[string]any) string {
	t.Helper()
	body := map[string]any{
		"regime_code": "612", "counterparty_type": "persona_moral", "amount_cents": 1000000, "iva_rate": "rate_16",
		"iva_cents": 160000, "isr_withheld_cents": 100000, "iva_withheld_cents": 106667,
		"invoiced_on": "2026-09-05", "collected_on": "2026-09-10",
	}
	if id != "" {
		body["id"] = id
	}
	for k, v := range fields {
		if name, drop := strings.CutPrefix(k, "-"); drop {
			delete(body, name)
			continue
		}
		body[k] = v
	}
	raw, err := json.Marshal(body)
	require.NoError(t, err)
	return string(raw)
}

func decodeIncome(t *testing.T, rec *httptest.ResponseRecorder) gen.Income {
	t.Helper()
	var i gen.Income
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &i), rec.Body.String())
	return i
}

func incomeWarnings(i gen.Income) []string {
	out := make([]string, 0, len(i.Warnings))
	for _, w := range i.Warnings {
		out = append(out, string(w.Code)+"@"+w.Field)
	}
	return out
}

// incomeActors signs Ana and Bea up. Ana's 612 and 601 are backdated to
// 2025 (the API starts no regime before 2026-01), so a 2025 date reaches the
// parameter check (no set before 2026) rather than the regime window.
type incomeActors struct {
	ana, bea, anaID string
}

func (e contractEnv) incomeActors(t *testing.T, c *contracttest.Contract) incomeActors {
	t.Helper()
	ana := e.signUp(t, signUpEmail, acceptableSample)
	bea := e.signUp(t, beaEmail, otherSample).Tokens.AccessToken
	token := ana.Tokens.AccessToken
	e.createIncomeTaxpayer(t, c, token, incomeTaxpayerA, "fisica", rfcAna,
		`{"regime_code":"612","start_period":"2026-01"}`,
		`{"regime_code":"625","start_period":"2026-01","options":{"activities":["lodging"]}}`)
	e.createIncomeTaxpayer(t, c, token, incomeTaxpayerPM, "moral", rfcMoral, `{"regime_code":"601","start_period":"2026-01"}`)
	e.backdateRegime(t, incomeTaxpayerA, "612", "2025-01")
	e.backdateRegime(t, incomeTaxpayerPM, "601", "2025-01")
	e.createIncomeTaxpayer(t, c, bea, incomeTaxpayerB, "fisica", rfcOther, `{"regime_code":"612","start_period":"2026-01"}`)
	return incomeActors{ana: token, bea: bea, anaID: ana.User.Id.String()}
}

func (e contractEnv) createIncomeTaxpayer(t *testing.T, c *contracttest.Contract, token, id, personType, rfc string, regimes ...string) {
	t.Helper()
	body := fmt.Sprintf(`{"id":%q,"person_type":%q,"rfc":%q,"display_name":"Ingresos","regimes":[%s]}`,
		id, personType, rfc, strings.Join(regimes, ","))
	rec := c.ServeOp(t, contracttest.Op("createTaxpayer"), e.handler, e.send(http.MethodPost, "/v1/taxpayers", body, token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
}

func TestContract_incomes(t *testing.T) {
	c := contracttest.Load(t)
	modules := []string{"module_auth", "module_taxpayers", "module_incomes"}

	flagtest.OnOff(t, "module_incomes", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		f["module_auth"], f["module_taxpayers"] = true, true
		env.apply(t, modules, f)

		if !on {
			for _, op := range incomeOps {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(op.method, op.target, op.body, "any-token")))
			}
			return
		}

		actors := env.incomeActors(t, c)
		ana, bea := actors.ana, actors.bea
		serve := func(op contracttest.Operation, method, target, body, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, op, env.handler, env.send(method, target, body, token))
		}
		defaults := contracttest.Op("getIncomeDefaults")
		list, create, get := contracttest.Op("listIncomes"), contracttest.Op("createIncome"), contracttest.Op("getIncome")
		update, del, restore := contracttest.Op("updateIncome"), contracttest.Op("deleteIncome"), contracttest.Op("restoreIncome")

		t.Run("getIncomeDefaults: rates from the catalog; 400, 422, 404", func(t *testing.T) {
			target := incomesPath(incomeTaxpayerA, "/defaults")
			rec := serve(defaults, http.MethodGet, target+"?regime=612&counterparty_type=persona_moral&on=2026-09-12", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			var body gen.IncomeDefaults
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, []gen.IvaRate{"rate_16", "rate_8", "rate_0", "exempt"}, body.IvaRatesAllowed)
			assert.Equal(t, gen.IvaRate("rate_16"), body.SuggestedIvaRate)
			assert.True(t, body.IsrWithholdingAllowed) //nolint:staticcheck // SA1019: deprecated field tested until WH-06 (ADR-0008)
			require.NotNil(t, body.IsrWithholdingRate)
			assert.Equal(t, "0.1", *body.IsrWithholdingRate, "LISR 106: 10 %")
			require.NotNil(t, body.IvaWithholdingRate)
			assert.Equal(t, "0.666667", *body.IvaWithholdingRate, "RLIVA 3: 2/3 of the IVA")
			assert.Nil(t, body.PlatformActivity)

			rec = serve(defaults, http.MethodGet, target+"?regime=625&counterparty_type=plataforma&on=2026-09-12", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.NotNil(t, body.PlatformActivity)
			assert.Equal(t, gen.PlatformActivity("lodging"), *body.PlatformActivity, "the only activity is the default")
			assert.Equal(t, "0.04", *body.IsrWithholdingRate)
			assert.Equal(t, "0.5", *body.IvaWithholdingRate)

			rec = serve(defaults, http.MethodGet, target+"?regime=612&counterparty_type=publico_general", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, "on defaults to today: %s", rec.Body.String())
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.False(t, body.IsrWithholdingAllowed) //nolint:staticcheck // SA1019: deprecated field tested until WH-06 (ADR-0008)
			assert.Nil(t, body.IsrWithholdingRate)

			refusals := []struct {
				query   string
				status  int
				details []string
			}{
				{"?regime=626&counterparty_type=persona_moral", http.StatusUnprocessableEntity, []string{"regime/regime_not_active"}},
				{"?regime=612&counterparty_type=persona_moral&on=2025-06-01", http.StatusBadRequest, []string{"on/out_of_range"}},
				{"?regime=612&counterparty_type=persona_moral&platform_activity=lodging", http.StatusBadRequest,
					[]string{"platform_activity/not_for_regime"}},
				{"?regime=625&counterparty_type=plataforma&platform_activity=goods_services", http.StatusUnprocessableEntity,
					[]string{"platform_activity/activity_not_held"}},
			}
			for _, r := range refusals {
				rec = serve(defaults, http.MethodGet, target+r.query, "", ana)
				require.Equal(t, r.status, rec.Code, "%s: %s", r.query, rec.Body.String())
				assert.Equal(t, r.details, details(t, rec), r.query)
			}
			rec = serve(defaults, http.MethodGet, target+"?counterparty_type=persona_moral", "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, details(t, rec), "regime/required")
			rec = serve(defaults, http.MethodGet, target+"?regime=612&counterparty_type=persona_moral", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "SE-3: another user's taxpayer")
			rec = serve(defaults, http.MethodGet, target+"?regime=612&counterparty_type=persona_moral", "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("createIncome: 201, then an idempotent 200, then 409s", func(t *testing.T) {
			body := incomeJSON(t, clientIncomeID(1), map[string]any{"cfdi_uuid": strings.ToLower(incomeFolio), "description": "  " + incomeNote + "  "})
			rec := serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), body, ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			created := decodeIncome(t, rec)
			assert.Equal(t, clientIncomeID(1), created.Id.String())
			assert.Equal(t, incomeTaxpayerA, created.TaxpayerId.String())
			assert.False(t, created.PendingCollection)
			assert.Empty(t, created.Warnings)
			require.NotNil(t, created.CfdiUuid)
			assert.Equal(t, incomeFolio, *created.CfdiUuid, "stored upper-case")
			require.NotNil(t, created.Description)
			assert.Equal(t, incomeNote, *created.Description)
			assert.NotContains(t, env.logs.String(), incomeFolio, "S-6: never logged")
			assert.NotContains(t, env.logs.String(), incomeNote, "S-6: never logged")
			assert.Contains(t, env.logs.String(), "income.created")

			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), body, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, clientIncomeID(1), decodeIncome(t, rec).Id.String())

			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA),
				incomeJSON(t, clientIncomeID(1), map[string]any{"amount_cents": 1200000, "iva_cents": 192000}), ana)
			require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
			assert.Equal(t, gen.ErrorCodeConflict, decodeError(t, rec).Error.Code)
			assert.Equal(t, []string{"id/id_reused"}, details(t, rec))

			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerB), incomeJSON(t, clientIncomeID(1), nil), bea)
			require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
			assert.Nil(t, decodeError(t, rec).Error.Details, "another user's id: nothing leaks")
		})

		t.Run("createIncome: warnings in the body", func(t *testing.T) {
			rec := serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, clientIncomeID(2), map[string]any{
				"cfdi_uuid": incomeFolio, "isr_withheld_cents": 90000, "collected_on": "2026-09-15",
			}), ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			i := decodeIncome(t, rec)
			assert.ElementsMatch(t, []string{"withholding_differs_from_rate@isr_withheld_cents", "duplicate_cfdi_uuid@cfdi_uuid"},
				incomeWarnings(i))
			for _, w := range i.Warnings {
				assert.NotEmpty(t, w.Message)
				assert.NotContains(t, w.Message, incomeFolio)
			}
		})

		t.Run("createIncome: an accrual income pending collection", func(t *testing.T) {
			rec := serve(create, http.MethodPost, incomesPath(incomeTaxpayerPM), incomeJSON(t, clientIncomeID(3), map[string]any{
				"regime_code": "601", "-collected_on": nil, "-isr_withheld_cents": nil, "-iva_withheld_cents": nil,
			}), ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			i := decodeIncome(t, rec)
			assert.True(t, i.PendingCollection, "Por cobrar")
			assert.Nil(t, i.CollectedOn)
		})

		t.Run("createIncome: refusals (400 fields, 422, 404)", func(t *testing.T) {
			refusals := []struct {
				name    string
				body    string
				status  int
				details []string
			}{
				{"a date before the first parameter set", incomeJSON(t, "", map[string]any{"invoiced_on": "2025-12-20", "collected_on": "2025-12-31"}),
					http.StatusBadRequest, []string{"collected_on/out_of_range"}},
				{"a cash income without its collection date", incomeJSON(t, "", map[string]any{"-collected_on": nil}),
					http.StatusBadRequest, []string{"collected_on/required"}},
				{"a regime the taxpayer does not hold", incomeJSON(t, "", map[string]any{"regime_code": "626"}),
					http.StatusUnprocessableEntity, []string{"regime_code/regime_not_active"}},
				{"a persona física withholding", incomeJSON(t, "", map[string]any{"counterparty_type": "persona_fisica"}),
					http.StatusUnprocessableEntity, []string{"isr_withheld_cents/withholding_not_applicable", "iva_withheld_cents/withholding_not_applicable"}},
				{"an IVA that does not match", incomeJSON(t, "", map[string]any{"iva_cents": 150000}),
					http.StatusUnprocessableEntity, []string{"iva_cents/iva_amount_mismatch"}},
				{"an activity outside 625", incomeJSON(t, "", map[string]any{"platform_activity": "lodging"}),
					http.StatusBadRequest, []string{"platform_activity/not_for_regime"}},
			}
			for _, r := range refusals {
				rec := serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), r.body, ana)
				require.Equal(t, r.status, rec.Code, "%s: %s", r.name, rec.Body.String())
				assert.ElementsMatch(t, r.details, details(t, rec), r.name)
			}

			rec := serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, "", map[string]any{"-counterparty_type": nil}), ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "the contract's required counterparty_type")
			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, "", map[string]any{"amount_cents": 0}), ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "minimum: 1 at the edge")
			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, "", map[string]any{"user_id": actors.anaID}), ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "additionalProperties: false")
			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, "", nil), bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "another user's taxpayer")
			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, "", nil), "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("getIncome and listIncomes: own data, no warnings, totals", func(t *testing.T) {
			rec := serve(get, http.MethodGet, incomesPath(incomeTaxpayerA, "/", clientIncomeID(2)), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Empty(t, decodeIncome(t, rec).Warnings, "reads carry no warnings")

			rec = serve(get, http.MethodGet, incomesPath(incomeTaxpayerA, "/", clientIncomeID(2)), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "never 403")
			rec = serve(get, http.MethodGet, incomesPath(incomeTaxpayerB, "/", clientIncomeID(2)), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)

			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09&limit=1"), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var page gen.IncomePage
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			require.Len(t, page.Items, 1)
			assert.Equal(t, clientIncomeID(2), page.Items[0].Id.String(), "newest booking date first")
			assert.Equal(t, 2, page.Totals.Count, "totals cover the period, not the page")
			assert.EqualValues(t, 2000000, page.Totals.AmountCents)
			assert.EqualValues(t, 190000, page.Totals.IsrWithheldCents)
			require.NotNil(t, page.NextCursor)

			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09&limit=1&cursor="+*page.NextCursor), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			require.Len(t, page.Items, 1)
			assert.Equal(t, clientIncomeID(1), page.Items[0].Id.String())
			assert.Nil(t, page.NextCursor, "the last page")

			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09&regime=625"), "", ana)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			assert.Empty(t, page.Items)
			assert.Equal(t, 0, page.Totals.Count)

			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerPM, "?period=2026-09"), "", ana)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			require.Len(t, page.Items, 1, "an accrual income books on invoiced_on")
			assert.True(t, page.Items[0].PendingCollection)

			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09&cursor=not-a-cursor"), "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09"), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA), "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, details(t, rec), "period/required")
		})

		t.Run("updateIncome: merge-patch, null clears", func(t *testing.T) {
			target := incomesPath(incomeTaxpayerA, "/", clientIncomeID(1))
			rec := serve(update, http.MethodPatch, target, `{"cfdi_uuid":null,"description":null,"invoiced_on":null}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			i := decodeIncome(t, rec)
			assert.Nil(t, i.CfdiUuid, "null clears the folio")
			assert.Nil(t, i.Description, "null clears the description")
			assert.Nil(t, i.InvoicedOn, "null clears the invoice date of a cash income")
			assert.Contains(t, env.logs.String(), "income.updated")

			rec = serve(update, http.MethodPatch, incomesPath(incomeTaxpayerPM, "/", clientIncomeID(3)), `{"collected_on":"2026-10-02"}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.False(t, decodeIncome(t, rec).PendingCollection, "collected: no longer Por cobrar")

			rec = serve(update, http.MethodPatch, target, `{"collected_on":"2025-12-01"}`, ana)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"collected_on/out_of_range"}, details(t, rec))
			rec = serve(update, http.MethodPatch, target, `{"regime_code":"626"}`, ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			rec = serve(update, http.MethodPatch, target, `{}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "minProperties: 1")
			rec = serve(update, http.MethodPatch, target, `{"id":"`+clientIncomeID(9)+`"}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "id is immutable")
			rec = serve(update, http.MethodPatch, target, `{"description":"x"}`, bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("deleteIncome and restoreIncome: Undo", func(t *testing.T) {
			target := incomesPath(incomeTaxpayerA, "/", clientIncomeID(2))
			rec := serve(del, http.MethodDelete, target, "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(del, http.MethodDelete, target, "", ana)
			require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			rec = serve(del, http.MethodDelete, target, "", ana)
			assert.Equal(t, http.StatusNoContent, rec.Code, "idempotent")
			rec = serve(get, http.MethodGet, target, "", ana)
			assert.Equal(t, http.StatusNotFound, rec.Code, "a deleted income is gone from reads")

			rec = serve(restore, http.MethodPost, target+"/restore", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(restore, http.MethodPost, target+"/restore", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Equal(t, clientIncomeID(2), decodeIncome(t, rec).Id.String())
		})

		t.Run("M2-14 G-1: the taxpayers record lookup reads the incomes", func(t *testing.T) {
			endOp := contracttest.Op("archiveTaxpayerRegime")
			rec := serve(endOp, http.MethodDelete, "/v1/taxpayers/"+incomeTaxpayerA+"/regimes/612?end_period=2026-08", "", ana)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"end_period/records_after_end_period"}, details(t, rec))

			rec = serve(endOp, http.MethodDelete, "/v1/taxpayers/"+incomeTaxpayerA+"/regimes/612?end_period="+
				clock.NewSystem().Now().Format("2006-01"), "", ana)
			require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
			assert.Equal(t, gen.ErrorCodeRegimeHasRecords, decodeError(t, rec).Error.Code)
		})

		t.Run("an archived taxpayer: reads yes, writes 422", func(t *testing.T) {
			rec := serve(contracttest.Op("archiveTaxpayer"), http.MethodPost, "/v1/taxpayers/"+incomeTaxpayerA+"/archive", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			rec = serve(create, http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, "", nil), ana)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, details(t, rec))
			rec = serve(del, http.MethodDelete, incomesPath(incomeTaxpayerA, "/", clientIncomeID(1)), "", ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			rec = serve(list, http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09"), "", ana)
			assert.Equal(t, http.StatusOK, rec.Code)
			rec = serve(defaults, http.MethodGet, incomesPath(incomeTaxpayerA, "/defaults?regime=612&counterparty_type=persona_moral"), "", ana)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	})
}

// MODULE-incomes US-I5: 5 incomes per trial user; the 6th is 402 naming the
// incomes cap, never on an idempotent retry, and Undo still works at 5/5. A
// read-only user (trial over) may read, delete and restore (coordinator
// decision, MODULE-incomes §2.7), not create or edit.
func TestContract_incomesTrialCapAndReadOnly(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers", "module_incomes"},
		flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_incomes": true})
	actors := env.incomeActors(t, c)
	token := actors.ana
	create := contracttest.Op("createIncome")
	post := func(n int) *httptest.ResponseRecorder {
		return c.ServeOp(t, create, env.handler, env.send(http.MethodPost, incomesPath(incomeTaxpayerA),
			incomeJSON(t, clientIncomeID(n), map[string]any{"collected_on": fmt.Sprintf("2026-09-%02d", 10+n)}), token))
	}

	for n := 1; n <= 5; n++ {
		rec := post(n)
		require.Equal(t, http.StatusCreated, rec.Code, "income %d: %s", n, rec.Body.String())
	}
	rec := post(6)
	require.Equal(t, http.StatusPaymentRequired, rec.Code, rec.Body.String())
	assert.Equal(t, gen.ErrorCodeEntitlementRequired, decodeError(t, rec).Error.Code)
	assert.Equal(t, []string{"incomes_remaining/trial_cap_reached"}, details(t, rec))
	assert.Equal(t, http.StatusOK, post(5).Code, "an idempotent retry at the cap is no new slot")

	target := incomesPath(incomeTaxpayerA, "/", clientIncomeID(5))
	rec = c.ServeOp(t, contracttest.Op("deleteIncome"), env.handler, env.send(http.MethodDelete, target, "", token))
	require.Equal(t, http.StatusNoContent, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("restoreIncome"), env.handler, env.send(http.MethodPost, target+"/restore", "", token))
	require.Equal(t, http.StatusOK, rec.Code, "Undo at 5/5 consumes no slot: %s", rec.Body.String())
	rec = post(7)
	assert.Equal(t, http.StatusPaymentRequired, rec.Code, "the cap still holds after the Undo")

	env.makeReadOnly(t, actors.anaID)

	rec = c.ServeOp(t, contracttest.Op("updateIncome"), env.handler, env.send(http.MethodPatch, target, `{"description":"x"}`, token))
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("deleteIncome"), env.handler, env.send(http.MethodDelete, target, "", token))
	assert.Equal(t, http.StatusNoContent, rec.Code, "deleting is allowed read-only")
	rec = c.ServeOp(t, contracttest.Op("restoreIncome"), env.handler, env.send(http.MethodPost, target+"/restore", "", token))
	assert.Equal(t, http.StatusOK, rec.Code, "Undo is allowed read-only (coordinator decision)")
	rec = c.ServeOp(t, create, env.handler, env.send(http.MethodPost, incomesPath(incomeTaxpayerA), incomeJSON(t, clientIncomeID(8), nil), token))
	require.Equal(t, http.StatusPaymentRequired, rec.Code)
	assert.Nil(t, decodeError(t, rec).Error.Details, "read-only: no cap detail")
	rec = c.ServeOp(t, contracttest.Op("listIncomes"), env.handler, env.send(http.MethodGet, incomesPath(incomeTaxpayerA, "?period=2026-09"), "", token))
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("getIncome"), env.handler, env.send(http.MethodGet, incomesPath(incomeTaxpayerA, "/", clientIncomeID(1)), "", token))
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("getIncomeDefaults"), env.handler,
		env.send(http.MethodGet, incomesPath(incomeTaxpayerA, "/defaults?regime=612&counterparty_type=persona_moral"), "", token))
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = c.ServeOp(t, create, env.handler, env.send(http.MethodPost, incomesPath(incomeTaxpayerB), incomeJSON(t, "", nil), actors.bea))
	assert.Equal(t, http.StatusCreated, rec.Code, "Bea's trial is her own")
}

// An income dated in a fiscal year without its own published parameters
// carries provisional_parameters on its dates (the previous year's values).
// The income rules' day is pinned to January 2027; the trial and the
// taxpayer windows keep the system clock.
func TestContract_incomesProvisionalParameters(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{}, withIncomesDay(time.Date(2027, time.January, 15, 12, 0, 0, 0, clock.Location())))
	env.apply(t, []string{"module_auth", "module_taxpayers", "module_incomes"},
		flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_incomes": true})
	actors := env.incomeActors(t, c)

	rec := c.ServeOp(t, contracttest.Op("createIncome"), env.handler, env.send(http.MethodPost, incomesPath(incomeTaxpayerA),
		incomeJSON(t, "", map[string]any{"invoiced_on": "2027-01-08", "collected_on": "2027-01-10"}), actors.ana))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	i := decodeIncome(t, rec)
	assert.Equal(t, []string{"provisional_parameters@collected_on"}, incomeWarnings(i))
	assert.NotEmpty(t, i.Warnings[0].Message)
}
