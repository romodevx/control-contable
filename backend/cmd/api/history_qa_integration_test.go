//go:build integration

package main

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
)

// M8-07 QA: the "Cambió desde que lo viste" marker must appear only when
// the figures shown change (MODULE-history §2.4, OQ-H4). History reads the
// year with one range estimate while the app marks a month with the
// single-month estimate it rendered, so the two must agree to the hash on
// every month, for a persona física and a persona moral, including the
// months before the first regime (status none) and a month that carries an
// IVA credit from the month before (WH-09). Otherwise opening a month would
// mark it at once (a false mark).

// historyPMRegimeStart is when the persona moral's 601 starts: January ..
// March are months with no regime.
const historyPMRegimeStart = "2026-04"

// pmHistoryTaxpayer is a persona moral, 601 from April, with the
// PM-601-IVA-01 June records entered through the API (incomes, a 612
// lawyer whose withheld IVA is credited in July, merchandise) and its tax
// inputs.
func (e historyEnv) pmHistoryTaxpayer(t *testing.T) estimateFixture {
	t.Helper()
	session := e.signUp(t, "pm-history@example.com", acceptableSample)
	f := estimateFixture{userID: session.User.Id.String(), token: session.Tokens.AccessToken}
	created := e.do(e.send(http.MethodPost, "/v1/taxpayers", createJSON("", "moral", rfcMoral, "Moral Historial",
		`{"regime_code":"601","start_period":"`+historyPMRegimeStart+`","options":{"iva_fully_taxable":true}}`), f.token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	f.taxpayer = decodeTaxpayer(t, created).Id.String()

	e.postRecord(t, f, incomesPath(f.taxpayer), incomeJSON(t, "", map[string]any{
		"regime_code": "601", "amount_cents": 200_000_000, "iva_cents": 32_000_000,
		"operation_kind": "business_services", "invoiced_on": "2026-06-10", "collected_on": "2026-06-20",
	}))
	e.postRecord(t, f, expensesPath(f.taxpayer), expenseJSON(t, "", map[string]any{
		"regime_code": "601", "category_code": "servicios_profesionales", "amount_cents": 5_000_000, "iva_cents": 800_000,
		"supplier_kind": "pf_with_activity", "supplier_regime": "612", "paid_on": "2026-06-15",
	}))
	e.postRecord(t, f, expensesPath(f.taxpayer), expenseJSON(t, "", map[string]any{
		"regime_code": "601", "category_code": "mercancia", "amount_cents": 123_333_331, "iva_cents": 19_733_333,
		"supplier_kind": "pm", "paid_on": "2026-06-15",
	}))
	inputs := e.do(e.send(http.MethodPut, taxInputsTarget(f.taxpayer, "601", 2026), `{"values":[{"key":"coeficiente_utilidad","value_rate":"0.1500"}]}`, f.token))
	require.Equal(t, http.StatusOK, inputs.Code, inputs.Body.String())
	return f
}

func (e historyEnv) postRecord(t *testing.T, f estimateFixture, target, body string) {
	t.Helper()
	rec := e.do(e.send(http.MethodPost, target, body, f.token))
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
}

// assertMatchesEstimates checks every month of the year against the
// single-month estimate the app renders when the month is opened.
func (e historyEnv) assertMatchesEstimates(t *testing.T, f estimateFixture, year historyBody) {
	t.Helper()
	for _, m := range year.Months {
		est := e.estimate(t, f, m.Period)
		assert.Equal(t, est.FiguresFingerprint, m.FiguresFingerprint, "%s: figures hash", m.Period)
		assert.Equal(t, est.RecordsFingerprint, m.RecordsFingerprint, "%s: records hash", m.Period)
		assert.Equal(t, est.EngineVersion, m.EngineVersion, m.Period)
		assert.JSONEq(t, string(est.ParameterSet), string(m.ParameterSet), m.Period)
		assert.Equal(t, est.DueDate, m.DueDate, m.Period)
		if m.Tax != nil {
			assert.Equal(t, est.Totals.TotalPayableRoundedCents, m.Tax.TotalPayableRoundedCents, "%s: to the centavo", m.Period)
		}
	}
}

// openEveryPastMonth is the user opening each month of the year but the
// current one, then reading History again.
func (e historyEnv) openEveryPastMonth(t *testing.T, f estimateFixture, year historyBody) historyBody {
	t.Helper()
	for _, m := range year.Months {
		if !m.IsCurrent {
			e.open(t, f, m.Period)
		}
	}
	return decodeHistory(t, e.history(t, f, "fiscal_year=2026"))
}

func TestHistoryQA_noFalseMarks(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday), withIncomesDay(estToday), withExpensesDay(estToday))
	env.apply(t, historyModules, allOn(historyModules))
	h := historyEnv{contractEnv: env, c: contracttest.Load(t)}

	t.Run("persona física (TX-NET-01, 612 + 606): opening every month marks none", func(t *testing.T) {
		pf := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
		year := decodeHistory(t, h.history(t, pf, "fiscal_year=2026"))
		h.assertMatchesEstimates(t, pf, year)

		seen := h.openEveryPastMonth(t, pf, year)
		for _, m := range seen.Months {
			assert.False(t, m.Changed, "%s: nothing changed since it was opened", m.Period)
			assert.Nil(t, m.Reason, m.Period)
			assert.Equal(t, !m.IsCurrent, m.ViewedAt != nil, "%s: viewed iff opened", m.Period)
		}
	})

	t.Run("persona moral (601 from April): months with no regime, then no false mark", func(t *testing.T) {
		pm := h.pmHistoryTaxpayer(t)
		year := decodeHistory(t, h.history(t, pm, "fiscal_year=2026"))
		require.Len(t, year.Months, 10, "October .. January")
		var sum int64
		for _, m := range year.Months {
			if m.Period < historyPMRegimeStart {
				assert.Equal(t, "none", m.Status, "%s: Sin régimen en este mes", m.Period)
				assert.Nil(t, m.Tax, "%s: no amount without a regime", m.Period)
				assert.False(t, m.HasRecords, m.Period)
				continue
			}
			assert.NotEqual(t, "none", m.Status, m.Period)
			require.NotNil(t, m.Tax, m.Period)
			sum += m.Tax.TotalPayableRoundedCents
		}
		assert.Equal(t, sum, year.YearTotals.TotalPayableRoundedCents, "Suma de estimados mensuales")
		assert.True(t, year.month(t, "2026-06").HasRecords)
		assert.False(t, year.month(t, "2026-07").HasRecords, "July only carries June's credit")
		h.assertMatchesEstimates(t, pm, year)

		seen := h.openEveryPastMonth(t, pm, year)
		for _, m := range seen.Months {
			assert.False(t, m.Changed, "%s: nothing changed since it was opened", m.Period)
			assert.Equal(t, !m.IsCurrent, m.ViewedAt != nil, "%s: viewed iff opened", m.Period)
		}

		// A late June expense moves June's figures: June is marked for its
		// records; the months before it are not.
		h.postRecord(t, pm, expensesPath(pm.taxpayer), expenseJSON(t, "", map[string]any{
			"regime_code": "601", "category_code": "mercancia", "amount_cents": 1_000_000, "iva_cents": 160_000,
			"supplier_kind": "pm", "paid_on": "2026-06-28",
		}))
		after := decodeHistory(t, h.history(t, pm, "fiscal_year=2026"))
		june := after.month(t, "2026-06")
		assert.True(t, june.Changed, "June's figures moved")
		require.NotNil(t, june.Reason)
		assert.Equal(t, "records", *june.Reason)
		for _, p := range []string{"2026-01", "2026-03", "2026-04", "2026-05"} {
			assert.False(t, after.month(t, p).Changed, "%s: before the late expense", p)
		}
		assert.False(t, after.month(t, "2026-10").Changed, "the current month")
	})
}
