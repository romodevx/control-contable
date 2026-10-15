//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// WH-13 QA (MODULE-withholdings §9): the WM goldens WH-01..WH-12 were only
// pinned in the kernel (fiscal.DeriveWithholdings) and, through the API, US-W1
// and US-W6 alone. These tests enter every golden with the operations and
// fields the app sends, through the real composition root over Postgres as
// cc_app, and check that createIncome / createExpense store the golden, that
// the preview the form shows derives the same, and that the estimate reads
// the derived amounts (WH-01-EST, WH-11-JUN, WH-11-JUL). The same requests were
// run against the QA's local stack (qa-report-withholdings-2026-10-14.md).

// withholdingsQAModules are the modules the app uses to enter the goldens.
var withholdingsQAModules = []string{"module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine"}

// grantComp gives userID the operator's comp entitlement (data-model.md:
// demo and QA accounts), so the goldens are not cut by the 5-income trial.
// The owner's entitlements policy covers comp rows only, under the user's
// context, exactly as the operator grants it.
func (e contractEnv) grantComp(t *testing.T, userID string) {
	t.Helper()
	tx, err := e.owner.BeginTx(t.Context(), nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(t.Context(), "SELECT set_config('app.current_user_id', $1, true)", userID)
	require.NoError(t, err)
	_, err = tx.ExecContext(t.Context(),
		"INSERT INTO entitlements (user_id, kind, source, starts_at) VALUES ($1, 'comp', 'admin', now())", userID)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
}

// incomePreviewQuery is previewIncomeWithholdings' query for a create body:
// the same facts, dated by the withholding date (collected_on, else invoiced_on).
func incomePreviewQuery(t *testing.T, body string) string {
	t.Helper()
	var fields map[string]any
	require.NoError(t, json.Unmarshal([]byte(body), &fields))
	q := url.Values{}
	q.Set("regime", fmt.Sprint(fields["regime_code"]))
	on, ok := fields["collected_on"]
	if !ok {
		on = fields["invoiced_on"]
	}
	q.Set("on", fmt.Sprint(on))
	for _, k := range []string{"amount_cents", "iva_rate", "iva_cents", "counterparty_kind", "operation_kind",
		"residential_rent", "via_platform", "platform_activity"} {
		if v, ok := fields[k]; ok {
			q.Set(k, fmt.Sprint(v))
		}
	}
	return "/withholding-preview?" + q.Encode()
}

type withholdingGolden struct {
	name          string
	taxpayer      string
	fields        map[string]any
	isr, iva, net int64
	rules         []string
	cpaUnverified bool
	warnings      []string
}

func TestWithholdingsQA_goldensThroughTheAPI(t *testing.T) {
	c := contracttest.Load(t)
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday), withIncomesDay(estToday), withExpensesDay(estToday))
	on := flagtest.Static{}
	for _, m := range withholdingsQAModules {
		on[m] = true
	}
	env.apply(t, withholdingsQAModules, on)
	session := env.signUp(t, signUpEmail, acceptableSample)
	token := session.Tokens.AccessToken
	env.grantComp(t, session.User.Id.String())

	newTaxpayer := func(person, rfc string, regimes ...string) string {
		t.Helper()
		rec := c.ServeOp(t, contracttest.Op("createTaxpayer"), env.handler,
			env.send(http.MethodPost, "/v1/taxpayers", createJSON("", person, rfc, "Retenciones QA", regimes...), token))
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		return decodeTaxpayer(t, rec).Id.String()
	}
	fisica := newTaxpayer("fisica", rfcAna,
		`{"regime_code":"612","start_period":"2026-01"}`,
		`{"regime_code":"606","start_period":"2026-01","options":{"deduction_mode":"actual"}}`,
		`{"regime_code":"625","start_period":"2026-01","options":{"activities":["goods_services"],"rfc_given_to_platform":true}}`)
	resico := newTaxpayer("fisica", rfcOther, `{"regime_code":"626","start_period":"2026-01"}`)
	moral := newTaxpayer("moral", rfcMoral, `{"regime_code":"601","start_period":"2026-01"}`)

	july := map[string]any{"invoiced_on": "2026-07-10", "collected_on": "2026-07-10"}
	with := func(extra map[string]any) map[string]any {
		out := map[string]any{}
		for k, v := range july {
			out[k] = v
		}
		for k, v := range extra {
			out[k] = v
		}
		return out
	}
	goldens := []withholdingGolden{
		{name: "WH-01 owner: RESICO PF services to a PM", taxpayer: resico,
			fields: with(map[string]any{"regime_code": "626", "amount_cents": 10_000_000, "iva_cents": 1_600_000}),
			isr:    125_000, iva: 1_066_667, net: 10_408_333, rules: []string{"ISR-4", "IVA-6"}},
		{name: "WH-02 612 honorarios to a PM", taxpayer: fisica,
			fields: with(map[string]any{"amount_cents": 5_000_000, "iva_cents": 800_000}),
			isr:    500_000, iva: 533_333, net: 4_766_667, rules: []string{"ISR-5", "IVA-6"}},
		{name: "WH-03 606 commercial rent to a PM", taxpayer: fisica,
			fields: with(map[string]any{"regime_code": "606", "operation_kind": "rent_real_estate", "amount_cents": 3_000_000, "iva_cents": 480_000}),
			isr:    300_000, iva: 320_000, net: 2_860_000, rules: []string{"ISR-6", "IVA-6"}},
		{name: "WH-03b 606 residential rent to a PM: exempt", taxpayer: fisica,
			fields: with(map[string]any{"regime_code": "606", "operation_kind": "rent_real_estate", "residential_rent": true,
				"iva_rate": "exempt", "amount_cents": 3_000_000, "iva_cents": 0}),
			isr: 300_000, iva: 0, net: 2_700_000},
		{name: "WH-04 612 to a PF with activity: nobody withholds", taxpayer: fisica,
			fields: with(map[string]any{"counterparty_kind": "pf_with_activity", "-operation_kind": nil, "amount_cents": 2_000_000, "iva_cents": 320_000}),
			net:    2_320_000, rules: []string{"ISR-0", "IVA-1"}},
		{name: "WH-05 612 sells goods to a PM", taxpayer: fisica,
			fields: with(map[string]any{"operation_kind": "sale_of_goods", "amount_cents": 5_000_000, "iva_cents": 800_000}),
			net:    5_800_000, rules: []string{"ISR-7", "IVA-8"}},
		{name: "WH-05b RESICO PF sells goods to a PM (F5), in August", taxpayer: resico,
			fields: map[string]any{"regime_code": "626", "operation_kind": "sale_of_goods", "amount_cents": 5_000_000, "iva_cents": 800_000,
				"invoiced_on": "2026-08-10", "collected_on": "2026-08-10"},
			isr: 62_500, net: 5_737_500, rules: []string{"ISR-4", "IVA-8"}},
		{name: "WH-06 PM carrier to a PM: freight 4 % of the price", taxpayer: moral,
			fields: with(map[string]any{"regime_code": "601", "operation_kind": "ground_freight"}),
			iva:    40_000, net: 1_120_000, rules: []string{"ISR-2", "IVA-4"}},
		{name: "WH-06b PF 612 carrier to a PM", taxpayer: fisica,
			fields: with(map[string]any{"operation_kind": "ground_freight"}),
			iva:    40_000, net: 1_120_000, rules: []string{"ISR-7", "IVA-4"}},
		{name: "WH-07 PM to PM specialized services: the repealed 6 % is gone", taxpayer: moral,
			fields: with(map[string]any{"regime_code": "601", "operation_kind": "business_services", "amount_cents": 10_000_000, "iva_cents": 1_600_000}),
			net:    11_600_000, rules: []string{"ISR-2", "IVA-8"}},
		{name: "WH-08 PF sells scrap to a PM: 100 % of the IVA", taxpayer: fisica,
			fields: with(map[string]any{"operation_kind": "scrap", "amount_cents": 2_000_000, "iva_cents": 320_000}),
			iva:    320_000, net: 2_000_000, rules: []string{"ISR-7", "IVA-5"}},
		{name: "WH-09 625 PF sells goods via a platform, RFC given", taxpayer: fisica,
			fields: with(map[string]any{"regime_code": "625", "via_platform": true, "-operation_kind": nil, "platform_activity": "goods_services"}),
			isr:    25_000, iva: 80_000, net: 1_055_000, rules: []string{"ISR-3", "IVA-3"}},
		{name: "WH-10 612 honorarios to público en general", taxpayer: fisica,
			fields: with(map[string]any{"counterparty_kind": "pf_no_activity_or_public", "-operation_kind": nil, "amount_cents": 500_000, "iva_cents": 80_000}),
			net:    580_000, rules: []string{"ISR-0", "IVA-0"}},
		{name: "WH-12 612 honorarios to the Federación: MVP defaults, flagged", taxpayer: fisica,
			fields: with(map[string]any{"counterparty_kind": "gov_federal", "amount_cents": 3_000_000, "iva_cents": 480_000}),
			isr:    300_000, iva: 320_000, net: 2_860_000, rules: []string{"ISR-5", "IVA-7"},
			cpaUnverified: true, warnings: []string{"withholding_cpa_unverified@counterparty_kind"}},
	}
	create, preview := contracttest.Op("createIncome"), contracttest.Op("previewIncomeWithholdings")
	for _, g := range goldens {
		t.Run(g.name, func(t *testing.T) {
			body := incomeJSON(t, "", g.fields)
			rec := c.ServeOp(t, create, env.handler, env.send(http.MethodPost, incomesPath(g.taxpayer), body, token))
			require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
			created := decodeIncome(t, rec)
			w := created.Withholding
			require.NotNil(t, w, "a new income is always derived")
			assert.EqualValues(t, g.isr, w.IsrCents, "ISR withheld")
			assert.EqualValues(t, g.iva, w.IvaCents, "IVA withheld")
			assert.EqualValues(t, g.net, w.NetCents, "net")
			if g.rules != nil {
				assert.Equal(t, g.rules, w.RuleIds)
			}
			assert.Equal(t, g.cpaUnverified, w.CpaUnverified, "cpa_unverified")
			if g.warnings != nil {
				assert.Subset(t, incomeWarnings(created), g.warnings)
			}
			assert.Equal(t, "2026.1", w.TableVersion)

			rec = c.ServeOp(t, preview, env.handler,
				env.send(http.MethodGet, incomesPath(g.taxpayer, incomePreviewQuery(t, body)), "", token))
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var p gen.WithholdingPreview
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p), rec.Body.String())
			assert.Equal(t, *w, p.Withholding, "the card the form shows is what the create stored")
		})
	}

	t.Run("WH-11 expense side: PM 601 pays a PF 612 consultant, create and preview agree", func(t *testing.T) {
		rec := c.ServeOp(t, contracttest.Op("createExpense"), env.handler,
			env.send(http.MethodPost, expensesPath(moral), usW6Body(t, "", nil), token))
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		w := decodeExpense(t, rec).Withholding
		require.NotNil(t, w)
		assert.EqualValues(t, 400_000, w.IsrCents)
		assert.EqualValues(t, 426_667, w.IvaCents)
		assert.EqualValues(t, 3_813_333, w.NetCents)
		assert.Equal(t, []string{"ISR-5", "IVA-6"}, w.RuleIds)
		require.NotNil(t, w.RemitDueOn)
		assert.Equal(t, "2026-07-17", w.RemitDueOn.String())

		rec = c.ServeOp(t, contracttest.Op("previewExpenseWithholdings"), env.handler,
			env.send(http.MethodGet, previewTarget(moral, usW6Query), "", token))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Equal(t, *w, decodePreview(t, rec).Withholding)
	})

	estimate := func(t *testing.T, taxpayer, period string) gen.TaxpayerEstimate {
		t.Helper()
		rec := env.do(env.send(http.MethodGet, estimateTarget(taxpayer, period), "", token))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var got gen.TaxpayerEstimate
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
		require.Len(t, got.Regimes, 1)
		require.NotNil(t, got.Regimes[0].IvaComponents)
		return got
	}

	t.Run("WH-01-EST: July of the RESICO taxpayer, ISR 750.00 and IVA 5,333.33 to pay", func(t *testing.T) {
		got := estimate(t, resico, "2026-07")
		assert.EqualValues(t, 75_000, got.Totals.Isr.AmountCents, "2.00 % × 100,000.00 − 1,250.00 withheld")
		assert.EqualValues(t, 533_333, got.Totals.Iva.AmountCents, "16,000.00 − 10,666.67 withheld")
		iva := got.Regimes[0].IvaComponents
		assert.EqualValues(t, 1_600_000, iva.ChargedCollectedCents)
		assert.EqualValues(t, 1_066_667, iva.WithheldFromTaxpayerCents)
	})

	t.Run("WH-05b in August: RESICO ISR below the withholding shows 0 to pay", func(t *testing.T) {
		got := estimate(t, resico, "2026-08")
		assert.EqualValues(t, 0, got.Totals.Isr.PayableRoundedCents, "1.10 % × 50,000.00 = 550.00 < 625.00 withheld")
	})

	t.Run("WH-11-JUN and WH-11-JUL: remit in June, credit the withheld IVA in July", func(t *testing.T) {
		june := estimate(t, moral, "2026-06")
		assert.EqualValues(t, 826_667, june.Regimes[0].WithheldToRemitCents, "ISR 4,000.00 + IVA 4,266.67")
		assert.EqualValues(t, 213_333, june.Regimes[0].IvaComponents.CreditablePaidCents, "6,400.00 − 4,266.67 paid in June")
		jul := estimate(t, moral, "2026-07")
		assert.EqualValues(t, 0, jul.Regimes[0].WithheldToRemitCents)
		// July also holds WH-06 and WH-07 (the PM's incomes): 400.00 + 16,000.00 charged.
		assert.EqualValues(t, 426_667, jul.Regimes[0].IvaComponents.CreditablePaidCents,
			"June's withheld IVA, credited in July (LIVA 5 fr. IV)")
	})
}
