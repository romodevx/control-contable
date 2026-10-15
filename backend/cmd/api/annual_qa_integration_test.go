//go:build integration

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/pgtest"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// ANN-23 QA of module_annual through the real composition root over
// PostgreSQL as cc_app (the visible device walk is blocked by the
// environment; see .gstack/qa-reports/qa-report-annual-2026-10-14.md).
// Each test pins a check the module's own suites only proved with fakes or
// for one operation: the "Confirmar todos" flow (US-A9), the disclaimers
// every response carries (§2.8), a persona moral (§5 person_type_not_allowed),
// a read-only user and an archived taxpayer on every write (US-A13) and the
// 500 live deductions a year on the real store.

const (
	pmAnnualEmail = "pm-annual@example.com"
	// annualDeductionsBody is one valid personal deduction of 2026.
	annualDeductionsBody = `{"category":"medical","amount_cents":120000,"paid_on":"2026-05-10","payment_method":"electronic","beneficiary":"self"}`
)

// annualAlwaysDisclaimers are the codes every 2026 annual estimate carries,
// whatever the taxpayer (MODULE-annual §2.8).
var annualAlwaysDisclaimers = []string{"annual_estimate", "rmf_2027_pending", "no_annual_iva"}

func (e contractEnv) annualEstimate(t *testing.T, f estimateFixture) gen.AnnualEstimate {
	t.Helper()
	rec := e.do(e.send(http.MethodGet, annualEstimateTarget(f.taxpayer, 2026), "", f.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var est gen.AnnualEstimate
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &est))
	return est
}

func disclaimerCodes(est gen.AnnualEstimate) []string {
	codes := make([]string, 0, len(est.Disclaimers))
	for _, d := range est.Disclaimers {
		codes = append(codes, string(d.Code))
	}
	return codes
}

// assertAnnualFrame checks what every annual estimate shows next to its
// figures: the frozen legend, annual_estimate first and the always-on codes.
func assertAnnualFrame(t *testing.T, est gen.AnnualEstimate) {
	t.Helper()
	assert.Equal(t, gen.EstoEsUnAproximadoTuContadorTeDaráLasCifrasReales, est.Legend)
	codes := disclaimerCodes(est)
	require.NotEmpty(t, codes)
	assert.Equal(t, "annual_estimate", codes[0], "annual_estimate is first")
	assert.Subset(t, codes, annualAlwaysDisclaimers)
	for _, d := range est.Disclaimers {
		assert.NotEmpty(t, d.Text, "%s has its text", d.Code)
	}
}

func filedTarget(taxpayerID, regime string) string {
	return "/v1/taxpayers/" + taxpayerID + "/regimes/" + regime + "/filed-payments/2026"
}

// confirmAll is the filed-payments screen's "Confirmar todos": every month
// whose amount is shown (effective_cents) is stored as filed, as the app's
// FiledPaymentsFormModel.confirmAll sends it.
func (e contractEnv) confirmAll(t *testing.T, f estimateFixture, regime string) (gen.FiledPayments, bool) {
	t.Helper()
	rec := e.do(e.send(http.MethodGet, filedTarget(f.taxpayer, regime), "", f.token))
	if rec.Code != http.StatusOK {
		return gen.FiledPayments{}, false // a regime without provisional ISR has no payments
	}
	var shown gen.FiledPayments
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &shown))
	type entry struct {
		Month       int   `json:"month"`
		AmountCents int64 `json:"amount_cents"`
	}
	var months []entry
	for _, m := range shown.Months {
		if m.InWindow && m.EffectiveCents != nil {
			months = append(months, entry{Month: m.Month, AmountCents: *m.EffectiveCents})
		}
	}
	body, err := json.Marshal(map[string][]entry{"months": months})
	require.NoError(t, err)
	rec = e.do(e.send(http.MethodPut, filedTarget(f.taxpayer, regime), string(body), f.token))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	var stored gen.FiledPayments
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &stored))
	return stored, true
}

// US-A9: with no filed payment the credits come from the app's monthly
// estimates and payments_estimated shows; "Confirmar todos" stores those
// values as filed, the warning disappears and no figure moves.
func TestAnnualQA_confirmAllKeepsTheFigure(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.apply(t, annualModules, annualAllOn())
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)

	before := env.annualEstimate(t, ana)
	assertAnnualFrame(t, before)
	assert.Contains(t, disclaimerCodes(before), "year_in_progress")
	require.Contains(t, disclaimerCodes(before), "payments_estimated", "no payment is confirmed yet")
	require.NotEmpty(t, before.Sections)
	section := before.Sections[0]
	require.Equal(t, gen.AnnualSectionKind("pf_art152"), section.Kind)
	require.NotNil(t, section.Tax)

	confirmed := 0
	for _, regime := range section.Regimes {
		stored, ok := env.confirmAll(t, ana, regime)
		if !ok {
			continue
		}
		confirmed++
		for _, m := range stored.Months {
			if m.EffectiveCents != nil {
				assert.Equal(t, gen.FiledPaymentSource("filed"), m.Source, "%s %s: stored as filed", regime, m.Period)
			}
		}
	}
	require.Positive(t, confirmed, "at least one regime pays provisional ISR")

	after := env.annualEstimate(t, ana)
	assertAnnualFrame(t, after)
	assert.NotContains(t, disclaimerCodes(after), "payments_estimated", "every payment is confirmed")
	require.NotEmpty(t, after.Sections)
	assert.Equal(t, *section.Tax, *after.Sections[0].Tax, "confirming the shown values moves no figure")
	assert.Equal(t, section.Status, after.Sections[0].Status)
}

// A persona moral: the PM section with its own disclaimers, and the PF-only
// inputs refused (422 person_type_not_allowed) while their lists stay
// readable.
func TestAnnualQA_personaMoral(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.apply(t, annualModules, annualAllOn())
	session := env.signUp(t, pmAnnualEmail, acceptableSample)
	pm := estimateFixture{userID: session.User.Id.String(), token: session.Tokens.AccessToken}
	created := env.do(env.send(http.MethodPost, "/v1/taxpayers", createJSON("", "moral", rfcMoral, "Moral Anual",
		`{"regime_code":"601","start_period":"2026-01","options":{"iva_fully_taxable":true}}`), pm.token))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	pm.taxpayer = decodeTaxpayer(t, created).Id.String()
	base := "/v1/taxpayers/" + pm.taxpayer

	est := env.annualEstimate(t, pm)
	assertAnnualFrame(t, est)
	assert.Equal(t, gen.PersonType("moral"), est.PersonType)
	assert.Contains(t, disclaimerCodes(est), "inflation_adjustment_excluded")
	require.NotEmpty(t, est.Sections)
	assert.Equal(t, gen.AnnualSectionKind("pm_601"), est.Sections[0].Kind)

	for _, write := range []struct{ target, body string }{
		{base + "/personal-deductions", annualDeductionsBody},
		{base + "/salary-statements", `{"fiscal_year":2026,"accumulable_cents":30000000,"exempt_cents":0,"isr_withheld_cents":4000000}`},
	} {
		rec := env.do(env.send(http.MethodPost, write.target, write.body, pm.token))
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "%s: %s", write.target, rec.Body.String())
		assert.True(t, slices.ContainsFunc(details(t, rec), func(d string) bool {
			return strings.HasSuffix(d, "/person_type_not_allowed")
		}), "%s: %v", write.target, details(t, rec))
	}
	rec := env.do(env.send(http.MethodGet, filedTarget(pm.taxpayer, "601"), "", pm.token))
	assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

// annualWriteOps are the writes of annualSteps: everything a read-only user
// or an archived taxpayer must be refused.
var annualWriteOps = []string{
	"putAnnualInputs", "createSalaryStatement", "updateSalaryStatement", "deleteSalaryStatement",
	"createPersonalDeduction", "updatePersonalDeduction", "deletePersonalDeduction", "restorePersonalDeduction",
	"putFiledPayments",
}

// refusedSweep runs annualSteps with ids already created (a deduction and a
// constancia), expecting every read served and every write refused with
// status and, when given, the detail code.
func refusedSweep(t *testing.T, env contractEnv, f estimateFixture, ids map[string]string, status int, detail string) {
	t.Helper()
	fill := func(s string) string {
		for placeholder, id := range ids {
			s = strings.ReplaceAll(s, placeholder, id)
		}
		return s
	}
	writes := 0
	for _, step := range annualSteps(f.taxpayer) {
		rec := env.do(env.send(step.method, fill(step.target), step.body, f.token))
		if !slices.Contains(annualWriteOps, step.op) {
			assert.Equal(t, http.StatusOK, rec.Code, "%s: reads stay open: %s", step.op, rec.Body.String())
			continue
		}
		writes++
		assert.Equal(t, status, rec.Code, "%s: %s", step.op, rec.Body.String())
		if detail != "" && rec.Code == status {
			assert.Contains(t, details(t, rec), detail, step.op)
		}
	}
	assert.Equal(t, len(annualWriteOps), writes, "every write was tried")
}

// annualRecords creates a constancia and a deduction as f and returns the
// ids annualSteps' placeholders take.
func (e contractEnv) annualRecords(t *testing.T, f estimateFixture) map[string]string {
	t.Helper()
	base := "/v1/taxpayers/" + f.taxpayer
	ids := map[string]string{}
	for placeholder, req := range map[string]struct{ target, body string }{
		"{statementId}": {base + "/salary-statements", `{"fiscal_year":2026,"accumulable_cents":30000000,"exempt_cents":0,"isr_withheld_cents":4000000}`},
		"{deductionId}": {base + "/personal-deductions", annualDeductionsBody},
	} {
		ids[placeholder] = e.createdID(t, e.do(e.send(http.MethodPost, req.target, req.body, f.token)))
	}
	return ids
}

func (e contractEnv) createdID(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
	var created struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	require.NotEmpty(t, created.ID)
	return created.ID
}

// US-A13 and §5 on all 15 operations through the stack: a read-only user
// (trial ended) reads everything and every write is 402; an archived
// taxpayer reads everything and every write is 422 taxpayer_archived.
func TestAnnualQA_readOnlyAndArchivedOnEveryWrite(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.apply(t, annualModules, annualAllOn())

	t.Run("read-only user", func(t *testing.T) {
		ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
		ids := env.annualRecords(t, ana)
		env.makeReadOnly(t, ana.userID)

		refusedSweep(t, env, ana, ids, http.StatusPaymentRequired, "")
		assertAnnualFrame(t, env.annualEstimate(t, ana))
	})

	t.Run("archived taxpayer", func(t *testing.T) {
		bea := env.txNet01Taxpayer(t, beaEmail, otherSample, "AAAA010101AA0")
		ids := env.annualRecords(t, bea)
		rec := env.do(env.send(http.MethodPost, "/v1/taxpayers/"+bea.taxpayer+"/archive", "", bea.token))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		refusedSweep(t, env, bea, ids, http.StatusUnprocessableEntity, "taxpayerId/taxpayer_archived")
	})
}

// seedDeductions stores n live 2026 deductions of f straight into the
// table as f (cc_app under RLS), so the bound is reached without n API
// calls and their rate limits.
func (e contractEnv) seedDeductions(t *testing.T, f estimateFixture, n int) {
	t.Helper()
	require.NoError(t, pgtest.AsUser(t.Context(), e.app, f.userID, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(t.Context(), `INSERT INTO personal_deductions
  (user_id, taxpayer_id, category, amount_cents, paid_on, payment_method, beneficiary)
SELECT $1, $2, 'medical', 10000, DATE '2026-02-01', 'electronic', 'self' FROM generate_series(1, $3::int)`,
			f.userID, f.taxpayer, n)
		return err
	}))
}

// The 500 live deductions a year (usecase/annual.MaxDeductionsPerYear) on
// the real store: the 500th is accepted, the 501st refused (422
// limit_reached), a deleted row frees its place, and restoring a deleted
// row into a full year is refused too.
func TestAnnualQA_deductionBoundOnTheRealStore(t *testing.T) {
	env := newContractEnv(t, pingOK{}, withTaxEngineDay(estToday))
	env.apply(t, annualModules, annualAllOn())
	ana := env.txNet01Taxpayer(t, signUpEmail, acceptableSample, rfcAna)
	target := "/v1/taxpayers/" + ana.taxpayer + "/personal-deductions"
	create := func() *httptest.ResponseRecorder {
		return env.do(env.send(http.MethodPost, target, annualDeductionsBody, ana.token))
	}
	limitReached := func(t *testing.T, rec *httptest.ResponseRecorder) {
		t.Helper()
		require.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		assert.Contains(t, details(t, rec), "body/limit_reached")
	}

	env.seedDeductions(t, ana, annual.MaxDeductionsPerYear-1)
	last := env.createdID(t, create())
	limitReached(t, create())

	rec := env.do(env.send(http.MethodDelete, target+"/"+last, "", ana.token))
	require.Equal(t, http.StatusNoContent, rec.Code, rec.Body.String())
	refill := env.createdID(t, create())
	assert.NotEqual(t, last, refill)

	limitReached(t, env.do(env.send(http.MethodPost, target+"/"+last+"/restore", "", ana.token)))
}
