//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// pm601IVAModules are the modules the app uses to enter PM-601-IVA-01.
var pm601IVAModules = []string{"module_auth", "module_taxpayers", "module_incomes", "module_expenses", "module_tax_engine"}

// TestGolden_pm601IVAEnteredThroughTheAPI (M5-30, the M5-29 QA note "601 IVA
// was not reproduced"): PM-601-IVA-01 entered with the same operations and
// fields the app sends, never with SQL. The income is the 601 form's
// invoice date plus its "Más detalles" collection date (an accrual income
// counts for IVA only once collected, LIVA 1-B); the lawyer's fees carry
// the withholdings the expense form prefills (ISR 10 % and 2/3 of the IVA);
// the merchandise brings the creditable IVA to the golden's 200,000.00.
//
// The golden isolates the IVA: its withheld_to_remit is the IVA withheld
// only. The app also records the ISR the persona moral withholds from the
// lawyer (LISR 106), so the regime's withholdings to remit add 5,000.00.
func TestGolden_pm601IVAEnteredThroughTheAPI(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday), withIncomesDay(estToday), withExpensesDay(estToday))
	on := flagtest.Static{}
	for _, m := range pm601IVAModules {
		on[m] = true
	}
	env.apply(t, pm601IVAModules, on)
	token := env.signUp(t, signUpEmail, acceptableSample).Tokens.AccessToken

	created := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "moral", rfcMoral, "General Moral",
		`{"regime_code":"601","start_period":"2026-01","options":{"iva_fully_taxable":true}}`), token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	taxpayer := decodeTaxpayer(t, created).Id.String()

	post := func(target, body string) {
		t.Helper()
		rec := env.do(env.send(http.MethodPost, target, body, token))
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	}
	post(incomesPath(taxpayer), incomeJSON(t, "", map[string]any{
		"regime_code": "601", "amount_cents": 200_000_000, "iva_cents": 32_000_000,
		"isr_withheld_cents": 0, "iva_withheld_cents": 0,
		"invoiced_on": "2026-06-10", "collected_on": "2026-06-20",
	}))
	post(expensesPath(taxpayer), expenseJSON(t, "", map[string]any{
		"regime_code": "601", "category_code": "servicios_profesionales", "amount_cents": 5_000_000, "iva_cents": 800_000,
		"isr_withheld_cents": 500_000, "iva_withheld_cents": 533_333, "paid_on": "2026-06-15",
	}))
	post(expensesPath(taxpayer), expenseJSON(t, "", map[string]any{
		"regime_code": "601", "category_code": "mercancia", "amount_cents": 120_000_000, "iva_cents": 19_200_000,
		"paid_on": "2026-06-15",
	}))
	inputs := env.do(env.send(http.MethodPut, taxInputsTarget(taxpayer, "601", 2026), pm601IVAInputs, token))
	require.Equal(t, http.StatusOK, inputs.Code, inputs.Body.String())

	rec := env.do(env.send(http.MethodGet, estimateTarget(taxpayer, "2026-06"), "", token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var got gen.TaxpayerEstimate
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Len(t, got.Regimes, 1)
	regime := got.Regimes[0]
	assert.Equal(t, gen.RegimeEstimateStatusComplete, regime.Status)
	require.NotNil(t, regime.IvaComponents)
	assert.EqualValues(t, 32_000_000, regime.IvaComponents.ChargedCollectedCents, "charged_collected")
	assert.EqualValues(t, 20_000_000, regime.IvaComponents.CreditablePaidCents, "creditable_paid")
	assert.EqualValues(t, 0, regime.IvaComponents.WithheldFromTaxpayerCents, "withheld_from_taxpayer")
	assert.EqualValues(t, 12_000_000, got.Totals.Iva.AmountCents, "IVA net")
	assert.EqualValues(t, 533_333+500_000, regime.WithheldToRemitCents, "IVA 5,333.33 (golden) + ISR 5,000.00")

	breakdown := env.do(env.send(http.MethodGet, breakdownTarget(taxpayer, "2026-06", "601"), "", token))
	require.Equal(t, http.StatusOK, breakdown.Code, breakdown.Body.String())
	var lines gen.Breakdown
	require.NoError(t, json.Unmarshal(breakdown.Body.Bytes(), &lines))
	withholdings := map[string]gen.MoneyCents{}
	for _, l := range lines.Lines {
		if l.Tax == gen.Withholding {
			withholdings[l.Concept] = l.AmountCents
		}
	}
	assert.Equal(t, map[string]gen.MoneyCents{
		"IVA que retuviste a tus proveedores (a enterar)": 533_333, // the golden's withholding.iva_to_remit
		"ISR que retuviste a tus proveedores (a enterar)": 500_000,
		"Retenciones a enterar":                           1_033_333,
	}, withholdings)
}

// pm601IVAInputs are the golden's tax inputs: coeficiente 0.15 and no prior
// payments January through May.
var pm601IVAInputs = `{"values":[{"key":"coeficiente_utilidad","value_rate":"0.1500"},` + strings.Join([]string{
	`{"key":"prior_provisional_payments","month":1,"value_cents":0}`,
	`{"key":"prior_provisional_payments","month":2,"value_cents":0}`,
	`{"key":"prior_provisional_payments","month":3,"value_cents":0}`,
	`{"key":"prior_provisional_payments","month":4,"value_cents":0}`,
	`{"key":"prior_provisional_payments","month":5,"value_cents":0}`,
}, ",") + `]}`
