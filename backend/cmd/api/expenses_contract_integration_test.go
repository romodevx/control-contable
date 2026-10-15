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

// Contract tests of the 7 module_expenses operations (M4-07) through the
// real composition root over Postgres as cc_app, with the flag ON and OFF.
// module_auth and module_taxpayers stay ON: they are how the users sign up
// and get a taxpayer.

const (
	expenseTaxpayerA = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7b01" // Ana's física: 612 since 2025-01 (backdated), 606 since 2026-01
	expenseTaxpayerB = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7b02" // Bea's física: 612 since 2026-01
	expenseFolio     = "3A9C7E21-5B4D-4F6A-8E2C-9D1B7A3F5C60"
	expenseLegend    = "Esto es un aproximado. No todos los gastos aplican para tu giro comercial."
)

// clientExpenseID is the n-th fixture expense id.
func clientExpenseID(n int) string { return fmt.Sprintf("0199d4a3-0000-7000-8000-%012d", n) }

// expenseOps is one request per module_expenses operation, for the OFF
// sweep; the token is a placeholder (the gate answers first).
var expenseOps = []struct{ method, target, body string }{
	{http.MethodGet, "/v1/taxpayers/" + expenseTaxpayerA + "/expense-categories?regime=612", ""},
	{http.MethodGet, "/v1/taxpayers/" + expenseTaxpayerA + "/expenses?period=2026-09", ""},
	{http.MethodPost, "/v1/taxpayers/" + expenseTaxpayerA + "/expenses", `{}`},
	{http.MethodGet, "/v1/taxpayers/" + expenseTaxpayerA + "/expenses/" + clientExpenseID(1), ""},
	{http.MethodPatch, "/v1/taxpayers/" + expenseTaxpayerA + "/expenses/" + clientExpenseID(1), `{"deductible":true}`},
	{http.MethodDelete, "/v1/taxpayers/" + expenseTaxpayerA + "/expenses/" + clientExpenseID(1), ""},
	{http.MethodPost, "/v1/taxpayers/" + expenseTaxpayerA + "/expenses/" + clientExpenseID(1) + "/restore", ""},
}

// expenseJSON is a create body; fields override or add top-level members
// ("-name" drops one).
func expenseJSON(t *testing.T, id string, fields map[string]any) string {
	t.Helper()
	body := map[string]any{
		"regime_code": "612", "category_code": "software", "amount_cents": 50000, "iva_rate": "rate_16",
		"iva_cents": 8000, "payment_method": "card_debit", "paid_on": "2026-09-12", "has_cfdi": true,
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

func decodeExpense(t *testing.T, rec *httptest.ResponseRecorder) gen.Expense {
	t.Helper()
	var e gen.Expense
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &e), rec.Body.String())
	return e
}

func expenseWarnings(e gen.Expense) []string {
	out := make([]string, 0, len(e.Warnings))
	for _, w := range e.Warnings {
		out = append(out, string(w.Code)+"@"+w.Field)
	}
	return out
}

// expenseActors signs Ana and Bea up and gives each a física taxpayer. Ana's
// 612 is backdated to 2025 (the API starts no regime before 2026-01), so a
// 2025 date reaches the parameter check (no set before 2026) rather than the
// regime window.
type expenseActors struct {
	ana, bea, anaID string
}

func (e contractEnv) expenseActors(t *testing.T, c *contracttest.Contract) expenseActors {
	t.Helper()
	ana := e.signUp(t, signUpEmail, acceptableSample)
	bea := e.signUp(t, beaEmail, otherSample).Tokens.AccessToken
	e.createExpenseTaxpayer(t, c, ana.Tokens.AccessToken, expenseTaxpayerA, rfcAna,
		`{"regime_code":"612","start_period":"2026-01"}`,
		`{"regime_code":"606","start_period":"2026-01","options":{"deduction_mode":"actual"}}`)
	e.createExpenseTaxpayer(t, c, bea, expenseTaxpayerB, rfcOther, `{"regime_code":"612","start_period":"2026-01"}`)
	e.backdateRegime(t, expenseTaxpayerA, "612", "2025-01")
	return expenseActors{ana: ana.Tokens.AccessToken, bea: bea, anaID: ana.User.Id.String()}
}

func (e contractEnv) createExpenseTaxpayer(t *testing.T, c *contracttest.Contract, token, id, rfc string, regimes ...string) {
	t.Helper()
	body := fmt.Sprintf(`{"id":%q,"person_type":"fisica","rfc":%q,"display_name":"Gastos","regimes":[%s]}`,
		id, rfc, strings.Join(regimes, ","))
	rec := c.ServeOp(t, contracttest.Op("createTaxpayer"), e.handler, e.send(http.MethodPost, "/v1/taxpayers", body, token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
}

func expensesPath(taxpayerID string, rest ...string) string {
	return "/v1/taxpayers/" + taxpayerID + "/expenses" + strings.Join(rest, "")
}

func TestContract_expenses(t *testing.T) {
	c := contracttest.Load(t)
	modules := []string{"module_auth", "module_taxpayers", "module_expenses"}

	flagtest.OnOff(t, "module_expenses", func(t *testing.T, f flagtest.Static, on bool) {
		env := newContractEnv(t, pingOK{})
		f["module_auth"], f["module_taxpayers"] = true, true
		env.apply(t, modules, f)

		if !on {
			for _, op := range expenseOps {
				flagtest.AssertHidden(t, c.Serve(t, env.handler, env.send(op.method, op.target, op.body, "any-token")))
			}
			return
		}

		actors := env.expenseActors(t, c)
		ana, bea := actors.ana, actors.bea
		serve := func(op contracttest.Operation, method, target, body, token string) *httptest.ResponseRecorder {
			t.Helper()
			return c.ServeOp(t, op, env.handler, env.send(method, target, body, token))
		}
		categories := contracttest.Op("listExpenseCategories")
		list, create, get := contracttest.Op("listExpenses"), contracttest.Op("createExpense"), contracttest.Op("getExpense")
		update, del, restore := contracttest.Op("updateExpense"), contracttest.Op("deleteExpense"), contracttest.Op("restoreExpense")

		t.Run("listExpenseCategories: resolved for the regime, with the legend; 422, 400, 404", func(t *testing.T) {
			target := "/v1/taxpayers/" + expenseTaxpayerA + "/expense-categories"
			rec := serve(categories, http.MethodGet, target+"?regime=612&on=2026-09-12", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			var body gen.ExpenseCategoryList
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, gen.Legend(expenseLegend), body.Legend, "MODULE-expenses §2.7: the server's legend")
			assert.Equal(t, []gen.IvaRate{"rate_16", "rate_8", "rate_0", "exempt"}, body.IvaRatesAllowed)
			assert.EqualValues(t, 200000, body.CashPaymentLimitCents)
			byCode := map[string]gen.ExpenseCategory{}
			for _, item := range body.Items {
				byCode[item.Code] = item
			}
			require.Contains(t, byCode, "software")
			assert.Equal(t, gen.Deductibility("deductible"), byCode["software"].Deductibility)

			rec = serve(categories, http.MethodGet, target+"?regime=612", "", ana)
			assert.Equal(t, http.StatusOK, rec.Code, "on defaults to today")

			rec = serve(categories, http.MethodGet, target+"?regime=626&on=2026-09-12", "", ana)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			assert.Equal(t, gen.ErrorCodeBusinessRuleViolation, decodeError(t, rec).Error.Code)
			assert.Equal(t, []string{"regime/regime_not_active"}, details(t, rec))

			rec = serve(categories, http.MethodGet, target+"?regime=612&on=2025-06-01", "", ana)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"on/out_of_range"}, details(t, rec), "no parameters before 2026 (PR #82)")

			rec = serve(categories, http.MethodGet, target, "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, details(t, rec), "regime/required")

			rec = serve(categories, http.MethodGet, target+"?regime=612", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "SE-3: another user's taxpayer")
			rec = serve(categories, http.MethodGet, target+"?regime=612", "", "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("createExpense: 201, then an idempotent 200, then 409s", func(t *testing.T) {
			body := expenseJSON(t, clientExpenseID(1), map[string]any{"cfdi_uuid": strings.ToLower(expenseFolio), "description": "  Suscripción  "})
			rec := serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), body, ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			created := decodeExpense(t, rec)
			assert.Equal(t, clientExpenseID(1), created.Id.String())
			assert.Equal(t, expenseTaxpayerA, created.TaxpayerId.String())
			assert.True(t, created.Deductible)
			assert.Equal(t, gen.DeductibleSource("category_default"), created.DeductibleSource)
			assert.Empty(t, created.Warnings)
			assert.Equal(t, gen.Legend(expenseLegend), created.Legend)
			require.NotNil(t, created.CfdiUuid)
			assert.Equal(t, expenseFolio, *created.CfdiUuid, "stored upper-case")
			require.NotNil(t, created.Description)
			assert.Equal(t, "Suscripción", *created.Description)
			assert.NotContains(t, env.logs.String(), expenseFolio, "S-6: never logged")
			assert.NotContains(t, env.logs.String(), "Suscripción", "S-6: never logged")
			assert.Contains(t, env.logs.String(), "expense.created")

			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), body, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, clientExpenseID(1), decodeExpense(t, rec).Id.String())

			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerA),
				expenseJSON(t, clientExpenseID(1), map[string]any{"amount_cents": 60000, "iva_cents": 9600}), ana)
			require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
			assert.Equal(t, gen.ErrorCodeConflict, decodeError(t, rec).Error.Code)
			assert.Equal(t, []string{"id/id_reused"}, details(t, rec))

			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerB), expenseJSON(t, clientExpenseID(1), nil), bea)
			require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
			assert.Nil(t, decodeError(t, rec).Error.Details, "another user's id: nothing leaks")
		})

		t.Run("createExpense: warnings in the body", func(t *testing.T) {
			rec := serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), expenseJSON(t, clientExpenseID(2), map[string]any{
				"category_code": "papeleria", "amount_cents": 200000, "iva_cents": 32000, "payment_method": "cash",
				"paid_on": "2026-09-15", "deductible": true,
			}), ana)
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			e := decodeExpense(t, rec)
			assert.True(t, e.Deductible)
			assert.Equal(t, gen.DeductibleSource("user"), e.DeductibleSource)
			assert.ElementsMatch(t, []string{"cash_over_threshold@payment_method", "iva_not_creditable_cash_payment@payment_method"},
				expenseWarnings(e))
			for _, w := range e.Warnings {
				assert.NotEmpty(t, w.Message)
			}
		})

		t.Run("createExpense: refusals (400 fields, 422, 404)", func(t *testing.T) {
			refusals := []struct {
				name    string
				body    string
				status  int
				details []string
			}{
				{"a date before the first parameter set", expenseJSON(t, "", map[string]any{"paid_on": "2025-12-31"}),
					http.StatusBadRequest, []string{"paid_on/out_of_range"}},
				{"an unknown category", expenseJSON(t, "", map[string]any{"category_code": "viajes-luna"}),
					http.StatusBadRequest, []string{"category_code/unknown_category"}},
				{"a folio without a CFDI", expenseJSON(t, "", map[string]any{"has_cfdi": false, "cfdi_uuid": expenseFolio}),
					http.StatusBadRequest, []string{"cfdi_uuid/has_cfdi_mismatch"}},
				{"a regime the taxpayer does not hold", expenseJSON(t, "", map[string]any{"regime_code": "626"}),
					http.StatusUnprocessableEntity, []string{"regime_code/regime_not_active"}},
				{"US-W8: a typed withholding", expenseJSON(t, "", map[string]any{"isr_withheld_cents": 1000}),
					http.StatusBadRequest, []string{"isr_withheld_cents/additionalProperties"}},
				{"US-W7: a física names a supplier", expenseJSON(t, "", map[string]any{"supplier_kind": "pm"}),
					http.StatusUnprocessableEntity, []string{"supplier_kind/supplier_not_applicable"}},
			}
			for _, r := range refusals {
				rec := serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), r.body, ana)
				require.Equal(t, r.status, rec.Code, "%s: %s", r.name, rec.Body.String())
				assert.Equal(t, r.details, details(t, rec), r.name)
			}

			rec := serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), expenseJSON(t, "", map[string]any{"-paid_on": nil}), ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "the contract's required paid_on")
			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), expenseJSON(t, "", map[string]any{"user_id": actors.anaID}), ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "additionalProperties: false")
			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), expenseJSON(t, "", nil), bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "another user's taxpayer")
			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), expenseJSON(t, "", nil), "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code)
		})

		t.Run("getExpense and listExpenses: own data, no warnings, totals and the legend", func(t *testing.T) {
			rec := serve(get, http.MethodGet, expensesPath(expenseTaxpayerA, "/", clientExpenseID(2)), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.Empty(t, decodeExpense(t, rec).Warnings, "reads carry no warnings")

			rec = serve(get, http.MethodGet, expensesPath(expenseTaxpayerA, "/", clientExpenseID(2)), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code, "never 403")
			rec = serve(get, http.MethodGet, expensesPath(expenseTaxpayerB, "/", clientExpenseID(2)), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)

			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09&limit=1"), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var page gen.ExpensePage
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			require.Len(t, page.Items, 1)
			assert.Equal(t, clientExpenseID(2), page.Items[0].Id.String(), "newest paid_on first")
			assert.Equal(t, gen.Legend(expenseLegend), page.Legend)
			assert.Equal(t, 2, page.Totals.Count, "totals cover the period, not the page")
			assert.EqualValues(t, 250000, page.Totals.AmountCents)
			require.NotNil(t, page.NextCursor)

			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09&limit=1&cursor="+*page.NextCursor), "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			require.Len(t, page.Items, 1)
			assert.Equal(t, clientExpenseID(1), page.Items[0].Id.String())
			assert.Nil(t, page.NextCursor, "the last page")

			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09&regime=606"), "", ana)
			require.Equal(t, http.StatusOK, rec.Code)
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &page))
			assert.Empty(t, page.Items)
			assert.Equal(t, 0, page.Totals.Count)

			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09&cursor=not-a-cursor"), "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09"), "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA), "", ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, details(t, rec), "period/required")
		})

		t.Run("updateExpense: merge-patch, null clears, override and back to the default", func(t *testing.T) {
			target := expensesPath(expenseTaxpayerA, "/", clientExpenseID(1))
			rec := serve(update, http.MethodPatch, target, `{"cfdi_uuid":null,"description":null,"deductible":false}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			e := decodeExpense(t, rec)
			assert.Nil(t, e.CfdiUuid, "null clears the folio")
			assert.Nil(t, e.Description, "null clears the description")
			assert.False(t, e.Deductible)
			assert.Equal(t, gen.DeductibleSource("user"), e.DeductibleSource)
			assert.Contains(t, env.logs.String(), "expense.updated")

			rec = serve(update, http.MethodPatch, target, `{"deductible":null}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			e = decodeExpense(t, rec)
			assert.True(t, e.Deductible, "the default applies again")
			assert.Equal(t, gen.DeductibleSource("category_default"), e.DeductibleSource)

			rec = serve(update, http.MethodPatch, target, `{"has_cfdi":false}`, ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"no_cfdi@has_cfdi"}, expenseWarnings(decodeExpense(t, rec)))

			rec = serve(update, http.MethodPatch, target, `{"paid_on":"2025-12-01"}`, ana)
			require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"paid_on/out_of_range"}, details(t, rec))
			rec = serve(update, http.MethodPatch, target, `{"regime_code":"626"}`, ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			rec = serve(update, http.MethodPatch, target, `{}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "minProperties: 1")
			rec = serve(update, http.MethodPatch, target, `{"id":"`+clientExpenseID(9)+`"}`, ana)
			assert.Equal(t, http.StatusBadRequest, rec.Code, "id is immutable")
			rec = serve(update, http.MethodPatch, target, `{"deductible":true}`, bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
		})

		t.Run("deleteExpense and restoreExpense: Undo", func(t *testing.T) {
			target := expensesPath(expenseTaxpayerA, "/", clientExpenseID(2))
			rec := serve(del, http.MethodDelete, target, "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(del, http.MethodDelete, target, "", ana)
			require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
			rec = serve(del, http.MethodDelete, target, "", ana)
			assert.Equal(t, http.StatusNoContent, rec.Code, "idempotent")
			rec = serve(get, http.MethodGet, target, "", ana)
			assert.Equal(t, http.StatusNotFound, rec.Code, "a deleted expense is gone from reads")

			rec = serve(restore, http.MethodPost, target+"/restore", "", bea)
			assert.Equal(t, http.StatusNotFound, rec.Code)
			rec = serve(restore, http.MethodPost, target+"/restore", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
			assert.ElementsMatch(t, []string{"cash_over_threshold@payment_method", "iva_not_creditable_cash_payment@payment_method"},
				expenseWarnings(decodeExpense(t, rec)), "warnings are computed as on create")
		})

		t.Run("M2-14 G-1: the taxpayers record lookup reads the expenses", func(t *testing.T) {
			target := "/v1/taxpayers/" + expenseTaxpayerA + "/regimes/612"
			endOp := contracttest.Op("archiveTaxpayerRegime")
			rec := serve(endOp, http.MethodDelete, target+"?end_period=2026-08", "", ana)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"end_period/records_after_end_period"}, details(t, rec))

			rec = serve(endOp, http.MethodDelete, target+"?end_period="+clock.NewSystem().Now().Format("2006-01"), "", ana)
			require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
			assert.Equal(t, gen.ErrorCodeRegimeHasRecords, decodeError(t, rec).Error.Code)
		})

		t.Run("an archived taxpayer: reads yes, writes 422", func(t *testing.T) {
			rec := serve(contracttest.Op("archiveTaxpayer"), http.MethodPost, "/v1/taxpayers/"+expenseTaxpayerA+"/archive", "", ana)
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

			rec = serve(create, http.MethodPost, expensesPath(expenseTaxpayerA), expenseJSON(t, "", nil), ana)
			require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
			assert.Equal(t, []string{"taxpayerId/taxpayer_archived"}, details(t, rec))
			rec = serve(del, http.MethodDelete, expensesPath(expenseTaxpayerA, "/", clientExpenseID(1)), "", ana)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
			rec = serve(list, http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09"), "", ana)
			assert.Equal(t, http.StatusOK, rec.Code)
			rec = serve(categories, http.MethodGet, "/v1/taxpayers/"+expenseTaxpayerA+"/expense-categories?regime=612", "", ana)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	})
}

// MODULE-expenses US-E3: 5 expenses per trial user; the 6th is 402 naming
// the expenses cap, never on an idempotent retry, and Undo still works at
// 5/5. A read-only user (trial over) may read, delete and restore (DECIDED,
// coordinator, MODULE-incomes §2.7; see usecase.Expenses.Restore), not
// create or edit.
func TestContract_expensesTrialCapAndReadOnly(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth", "module_taxpayers", "module_expenses"},
		flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_expenses": true})
	actors := env.expenseActors(t, c)
	token := actors.ana
	create := contracttest.Op("createExpense")
	post := func(n int) *httptest.ResponseRecorder {
		return c.ServeOp(t, create, env.handler, env.send(http.MethodPost, expensesPath(expenseTaxpayerA),
			expenseJSON(t, clientExpenseID(n), map[string]any{"paid_on": fmt.Sprintf("2026-09-%02d", n)}), token))
	}

	for n := 1; n <= 5; n++ {
		rec := post(n)
		require.Equal(t, http.StatusCreated, rec.Code, "expense %d: %s", n, rec.Body.String())
	}
	rec := post(6)
	require.Equal(t, http.StatusPaymentRequired, rec.Code, rec.Body.String())
	assert.Equal(t, gen.ErrorCodeEntitlementRequired, decodeError(t, rec).Error.Code)
	assert.Equal(t, []string{"expenses_remaining/trial_cap_reached"}, details(t, rec))
	assert.Equal(t, http.StatusOK, post(5).Code, "an idempotent retry at the cap is no new slot")

	target := expensesPath(expenseTaxpayerA, "/", clientExpenseID(5))
	rec = c.ServeOp(t, contracttest.Op("deleteExpense"), env.handler, env.send(http.MethodDelete, target, "", token))
	require.Equal(t, http.StatusNoContent, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("restoreExpense"), env.handler, env.send(http.MethodPost, target+"/restore", "", token))
	require.Equal(t, http.StatusOK, rec.Code, "Undo at 5/5 consumes no slot: %s", rec.Body.String())
	rec = post(7)
	assert.Equal(t, http.StatusPaymentRequired, rec.Code, "the cap still holds after the Undo")

	env.makeReadOnly(t, actors.anaID)

	rec = c.ServeOp(t, contracttest.Op("updateExpense"), env.handler, env.send(http.MethodPatch, target, `{"deductible":false}`, token))
	assert.Equal(t, http.StatusPaymentRequired, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("deleteExpense"), env.handler, env.send(http.MethodDelete, target, "", token))
	assert.Equal(t, http.StatusNoContent, rec.Code, "deleting is allowed read-only")
	rec = c.ServeOp(t, contracttest.Op("restoreExpense"), env.handler, env.send(http.MethodPost, target+"/restore", "", token))
	assert.Equal(t, http.StatusOK, rec.Code, "Undo is allowed read-only (coordinator decision)")
	rec = c.ServeOp(t, create, env.handler, env.send(http.MethodPost, expensesPath(expenseTaxpayerA),
		expenseJSON(t, clientExpenseID(8), nil), token))
	require.Equal(t, http.StatusPaymentRequired, rec.Code)
	assert.Nil(t, decodeError(t, rec).Error.Details, "read-only: no cap detail")
	rec = c.ServeOp(t, contracttest.Op("listExpenses"), env.handler, env.send(http.MethodGet, expensesPath(expenseTaxpayerA, "?period=2026-09"), "", token))
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("getExpense"), env.handler, env.send(http.MethodGet, expensesPath(expenseTaxpayerA, "/", clientExpenseID(1)), "", token))
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = c.ServeOp(t, contracttest.Op("listExpenseCategories"), env.handler,
		env.send(http.MethodGet, "/v1/taxpayers/"+expenseTaxpayerA+"/expense-categories?regime=612", "", token))
	assert.Equal(t, http.StatusOK, rec.Code)
	rec = c.ServeOp(t, create, env.handler, env.send(http.MethodPost, expensesPath(expenseTaxpayerB), expenseJSON(t, "", nil), actors.bea))
	assert.Equal(t, http.StatusCreated, rec.Code, "Bea's trial is her own")
}

// A payment in a fiscal year without its own published parameters carries
// provisional_parameters on paid_on (the previous year's values). The
// expense rules' day is pinned to January 2027; the trial and the taxpayer
// windows keep the system clock.
func TestContract_expensesProvisionalParameters(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{}, withExpensesDay(time.Date(2027, time.January, 15, 12, 0, 0, 0, clock.Location())))
	env.apply(t, []string{"module_auth", "module_taxpayers", "module_expenses"},
		flagtest.Static{"module_auth": true, "module_taxpayers": true, "module_expenses": true})
	actors := env.expenseActors(t, c)

	rec := c.ServeOp(t, contracttest.Op("createExpense"), env.handler, env.send(http.MethodPost, expensesPath(expenseTaxpayerA),
		expenseJSON(t, "", map[string]any{"paid_on": "2027-01-10"}), actors.ana))

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	e := decodeExpense(t, rec)
	assert.Equal(t, []string{"provisional_parameters@paid_on"}, expenseWarnings(e))
	assert.NotEmpty(t, e.Warnings[0].Message)
}
