package annualhttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/contracttest"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	annualhttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// fakeInputs are the input use cases: canned answers, recorded calls.
type fakeInputs struct {
	calls []string
}

func (f *fakeInputs) record(s string) { f.calls = append(f.calls, s) }

func (f *fakeInputs) Get(_ context.Context, userID, taxpayerID string, year int) (usecaseannual.InputsView, error) {
	f.record("facts.get|" + userID + "|" + taxpayerID)
	return inputsView(year), nil
}

func (f *fakeInputs) Put(_ context.Context, userID, taxpayerID string, year int, facts usecaseannual.Facts) (usecaseannual.InputsView, error) {
	f.record("facts.put|" + userID + "|" + taxpayerID)
	v := inputsView(year)
	v.Facts, v.UpdatedAt = facts, &stamp
	return v, nil
}

func inputsView(year int) usecaseannual.InputsView {
	value := money.Cents(5_000_00)
	return usecaseannual.InputsView{FiscalYear: year, RegimeTaxInputs: []usecaseannual.RegimeTaxInput{
		{Regime: "612", Key: domain.InputPTUPaid, Label: "PTU pagada en el ejercicio", Value: &value},
		{Regime: "612", Key: domain.InputPriorYearLosses, Label: "Pérdidas fiscales de ejercicios anteriores"},
	}}
}

func statement() usecaseannual.StoredStatement {
	return usecaseannual.StoredStatement{
		ID: rowID, TaxpayerID: fisicaID, CreatedAt: stamp, UpdatedAt: stamp,
		SalaryFields: usecaseannual.SalaryFields{FiscalYear: 2026, EmployerLabel: "Empresa", Accumulable: 300_000_00, ISRWithheld: 40_000_00, Simultaneous: true},
	}
}

func (f *fakeInputs) List(_ context.Context, userID, taxpayerID string, year int) (usecaseannual.SalaryList, error) {
	f.record("salary.list|" + userID + "|" + taxpayerID)
	return usecaseannual.SalaryList{FiscalYear: year, Items: []usecaseannual.StoredStatement{statement()},
		Totals: usecaseannual.SalaryTotals{Accumulable: 300_000_00, ISRWithheld: 40_000_00}}, nil
}

func (f *fakeInputs) Create(_ context.Context, userID string, in usecaseannual.SalaryInput) (usecaseannual.SalaryResult, error) {
	f.record("salary.create|" + userID + "|" + in.TaxpayerID + "|" + in.ID)
	return usecaseannual.SalaryResult{Statement: statement(), Created: in.ID != rowID,
		Warnings: []usecaseannual.Warning{{Code: usecaseannual.WarningSalaryRegimeNotHeld, Field: "fiscal_year"}}}, nil
}

func (f *fakeInputs) Update(_ context.Context, userID, taxpayerID, id string, p usecaseannual.SalaryPatch) (usecaseannual.SalaryResult, error) {
	label := "absent"
	if p.EmployerLabel.Set {
		label = "null"
		if p.EmployerLabel.Value != nil {
			label = *p.EmployerLabel.Value
		}
	}
	f.record("salary.update|" + userID + "|" + taxpayerID + "|" + id + "|" + label)
	return usecaseannual.SalaryResult{Statement: statement()}, nil
}

func (f *fakeInputs) Delete(_ context.Context, userID, taxpayerID, id string) error {
	f.record("delete|" + userID + "|" + taxpayerID + "|" + id)
	return nil
}

// deductionsFake are the personal-deduction use cases.
func deduction(t *testing.T) usecaseannual.Deduction {
	return usecaseannual.Deduction{
		ID: rowID, TaxpayerID: fisicaID, FiscalYear: 2026, CreatedAt: stamp, UpdatedAt: stamp,
		DeductionFields: usecaseannual.DeductionFields{
			Category: adomain.CategoryMedical, Amount: 1_200_00, PaidOn: mustDate(t, "2026-05-10"),
			PaymentMethod: adomain.PaymentCash, Beneficiary: adomain.BeneficiarySelf,
		},
		Excluded: usecaseannual.ExclusionCashNotDeductible,
	}
}

type deductionsFake struct {
	t     *testing.T
	calls []string
}

func (f *deductionsFake) List(_ context.Context, userID string, q usecaseannual.DeductionQuery) (usecaseannual.DeductionList, error) {
	f.calls = append(f.calls, "list|"+userID+"|"+q.TaxpayerID+"|"+string(q.Category))
	return usecaseannual.DeductionList{FiscalYear: q.FiscalYear, Items: []usecaseannual.Deduction{deduction(f.t)},
		Totals: []usecaseannual.CategoryTotal{{Category: adomain.CategoryMedical, Registered: 1_200_00, Excluded: 1_200_00, Count: 1}}}, nil
}

func (f *deductionsFake) Create(_ context.Context, userID string, in usecaseannual.DeductionInput) (usecaseannual.DeductionResult, error) {
	f.calls = append(f.calls, "create|"+userID+"|"+in.TaxpayerID+"|"+in.PaidOn.String())
	return usecaseannual.DeductionResult{Deduction: deduction(f.t), Created: true,
		Warnings: []usecaseannual.Warning{{Code: usecaseannual.WarningCashNotDeductible, Field: "payment_method"}}}, nil
}

func (f *deductionsFake) Update(_ context.Context, userID, taxpayerID, id string, p usecaseannual.DeductionPatch) (usecaseannual.DeductionResult, error) {
	state := "cfdi:absent"
	if p.CFDIUUID.Set && p.CFDIUUID.Value == nil {
		state = "cfdi:null"
	}
	f.calls = append(f.calls, "update|"+userID+"|"+taxpayerID+"|"+id+"|"+state)
	return usecaseannual.DeductionResult{Deduction: deduction(f.t)}, nil
}

func (f *deductionsFake) Delete(_ context.Context, userID, taxpayerID, id string) error {
	f.calls = append(f.calls, "delete|"+userID+"|"+taxpayerID+"|"+id)
	return nil
}

func (f *deductionsFake) Restore(_ context.Context, userID, taxpayerID, id string) (usecaseannual.DeductionResult, error) {
	f.calls = append(f.calls, "restore|"+userID+"|"+taxpayerID+"|"+id)
	return usecaseannual.DeductionResult{Deduction: deduction(f.t)}, nil
}

// filedView is twelve months of 612, March confirmed.
func filedView(t *testing.T) usecaseannual.FiledView {
	v := usecaseannual.FiledView{Regime: "612", FiscalYear: 2026}
	for m := time.January; m <= time.December; m++ {
		p, err := domain.NewPeriod(2026, m)
		require.NoError(t, err)
		month := usecaseannual.FiledMonthView{Month: m, Period: p, InWindow: true, Closed: true}
		if m == time.March {
			amount := money.Cents(2_000_00)
			month.Filed = &amount
		}
		v.Months = append(v.Months, month)
	}
	return v
}

type filedFake struct {
	t     *testing.T
	calls []string
}

func (f *filedFake) Get(_ context.Context, userID string, req usecaseannual.FiledRequest) (usecaseannual.FiledView, error) {
	f.calls = append(f.calls, "get|"+userID+"|"+req.TaxpayerID+"|"+req.RegimeCode)
	return filedView(f.t), nil
}

func (f *filedFake) Put(_ context.Context, userID string, req usecaseannual.FiledRequest, entries []usecaseannual.FiledEntry) (usecaseannual.FiledView, error) {
	f.calls = append(f.calls, "put|"+userID+"|"+req.TaxpayerID+"|"+req.RegimeCode)
	require.Len(f.t, entries, 1)
	return filedView(f.t), nil
}

type world struct {
	inputs     *fakeInputs
	deductions *deductionsFake
	filed      *filedFake
	rng        *fakeRange
	deps       annualhttp.Deps
}

func newWorld(t *testing.T) *world {
	w := &world{inputs: &fakeInputs{}, deductions: &deductionsFake{t: t}, filed: &filedFake{t: t}, rng: &fakeRange{allowed: true}}
	w.deps = annualhttp.Deps{
		Estimates: realEstimates(t), Facts: w.inputs, Salary: w.inputs, Deductions: w.deductions, Filed: w.filed, Range: w.rng,
	}
	return w
}

func estimateTarget(id string, year int) string {
	return "/v1/taxpayers/" + id + "/annual-estimates/" + strconv.Itoa(year)
}

// US-A1..A8 through the stack: the estimate of a PF (one pf_art152 and the
// RESICO check) and a PM (pm_601 and the 603 explanation) conform to the
// contract; each spends one more estimate read for its range.
func TestGetAnnualEstimate(t *testing.T) {
	c := contracttest.Load(t)
	w := newWorld(t)
	h := newHarness(t, w.deps, annualOn, nil)

	pf := c.ServeOp(t, contracttest.Op("getAnnualEstimate"), h.handler, request(http.MethodGet, estimateTarget(fisicaID, 2026), ""))
	pmRec := c.ServeOp(t, contracttest.Op("getAnnualEstimate"), h.handler, request(http.MethodGet, estimateTarget(moralID, 2026), ""))

	require.Equal(t, http.StatusOK, pf.Code, pf.Body.String())
	require.Equal(t, http.StatusOK, pmRec.Code, pmRec.Body.String())
	assert.Equal(t, "no-store", pf.Header().Get("Cache-Control"))
	var body gen.AnnualEstimate
	require.NoError(t, json.Unmarshal(pf.Body.Bytes(), &body))
	assert.Equal(t, gen.AnnualLegend("Esto es un aproximado; tu contador te dará las cifras reales."), body.Legend)
	assert.Equal(t, []gen.AnnualComputedSectionKind{"pf_art152"}, body.HeadlineSections)
	require.Len(t, body.Sections, 2)
	assert.Equal(t, gen.AnnualSectionKind("pf_art152"), body.Sections[0].Kind)
	assert.Equal(t, gen.AnnualSectionKind("pf_resico_check"), body.Sections[1].Kind)
	assert.NotEmpty(t, body.Sections[0].ObligationReasons)
	require.NoError(t, json.Unmarshal(pmRec.Body.Bytes(), &body))
	require.Len(t, body.Sections, 2)
	require.NotNil(t, body.Sections[0].PmExtras)
	require.NotNil(t, body.Sections[0].PmExtras.SuggestedCuNextYear)
	assert.Len(t, *body.Sections[0].PmExtras.SuggestedCuNextYear, len("0.0000"))
	assert.Equal(t, gen.AnnualSectionKind("explanation"), body.Sections[1].Kind)
	assert.Equal(t, []string{userA, userA}, w.rng.keys, "each read ran the range")
}

// The refusals conform to the contract: a future year (400), a year
// before the parameters (422), another user's or an unknown taxpayer (404).
func TestGetAnnualEstimate_refusals(t *testing.T) {
	c := contracttest.Load(t)
	h := newHarness(t, newWorld(t).deps, annualOn, nil)
	for name, tc := range map[string]struct {
		target string
		status int
	}{
		"future":   {estimateTarget(fisicaID, 2028), http.StatusBadRequest},
		"too old":  {estimateTarget(fisicaID, 2025), http.StatusUnprocessableEntity},
		"unknown":  {estimateTarget(rowID, 2026), http.StatusNotFound},
		"bad year": {"/v1/taxpayers/" + fisicaID + "/annual-estimates/20x6", http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			rec := c.ServeOp(t, contracttest.Op("getAnnualEstimate"), h.handler, request(http.MethodGet, tc.target, ""))

			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
		})
	}
}

// A read whose range finds the estimate_reads budget spent is 429 with
// Retry-After (ADR-0009 "Budgets").
func TestGetAnnualEstimate_rangeBudgetSpent(t *testing.T) {
	c := contracttest.Load(t)
	w := newWorld(t)
	w.rng.allowed = false
	h := newHarness(t, w.deps, annualOn, nil)

	rec := c.ServeOp(t, contracttest.Op("getAnnualEstimate"), h.handler, request(http.MethodGet, estimateTarget(fisicaID, 2026), ""))

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "30", rec.Header().Get("Retry-After"))
}

// Every computed section's trace conforms, every line has a Spanish
// concept and a contract source; a kind the estimate lacks is 422.
func TestGetAnnualBreakdown(t *testing.T) {
	c := contracttest.Load(t)
	h := newHarness(t, newWorld(t).deps, annualOn, nil)
	for _, tc := range []struct{ id, section string }{
		{fisicaID, "pf_art152"}, {fisicaID, "pf_resico_check"}, {moralID, "pm_601"},
	} {
		t.Run(tc.section, func(t *testing.T) {
			target := estimateTarget(tc.id, 2026) + "/breakdown?section=" + tc.section

			rec := c.ServeOp(t, contracttest.Op("getAnnualBreakdown"), h.handler, request(http.MethodGet, target, ""))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			var body gen.AnnualBreakdown
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.NotEmpty(t, body.Lines)
			for _, l := range body.Lines {
				assert.NotEqual(t, "Paso del cálculo", l.Concept, "seq %d has no concept", l.Seq)
			}
		})
	}
	absent := estimateTarget(fisicaID, 2026) + "/breakdown?section=pm_601"
	rec := c.ServeOp(t, contracttest.Op("getAnnualBreakdown"), h.handler, request(http.MethodGet, absent, ""))
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	assert.Contains(t, rec.Body.String(), "section_not_present")
}

// The pf_art152 trace shows the caps, the cash exclusion and each payment
// by its source.
func TestGetAnnualBreakdown_traceTexts(t *testing.T) {
	h := newHarness(t, newWorld(t).deps, annualOn, nil)
	rec := httptest.NewRecorder()
	h.handler.ServeHTTP(rec, request(http.MethodGet, estimateTarget(fisicaID, 2026)+"/breakdown?section=pf_art152", ""))
	require.Equal(t, http.StatusOK, rec.Code)
	var body gen.AnnualBreakdown
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	var concepts, sources []string
	for _, l := range body.Lines {
		concepts, sources = append(concepts, l.Concept), append(sources, l.Source)
	}
	joined := strings.Join(concepts, "\n")
	assert.Contains(t, joined, "Tope aplicado: colegiaturas de Ana, de $30,000.00 a $12,900.00")
	assert.Contains(t, joined, "No deducible por pago en efectivo: gastos médicos, $1,000.00")
	assert.Contains(t, joined, "Pagos provisionales de ISR estimados por la app (12 meses)")
	assert.Contains(t, sources, "estimated_by_app")
	assert.Contains(t, sources, "salary_statements")
	assert.Contains(t, sources, "personal_deductions")
}

// The filed-payments screen: the stored months with every source's
// column; the read spends one more estimate read.
func TestFiledPayments(t *testing.T) {
	c := contracttest.Load(t)
	w := newWorld(t)
	h := newHarness(t, w.deps, annualOn, nil)
	target := "/v1/taxpayers/" + fisicaID + "/regimes/612/filed-payments/2026"

	got := c.ServeOp(t, contracttest.Op("getFiledPayments"), h.handler, request(http.MethodGet, target, ""))
	put := c.ServeOp(t, contracttest.Op("putFiledPayments"), h.handler,
		request(http.MethodPut, target, `{"months":[{"month":3,"amount_cents":200000}]}`))

	require.Equal(t, http.StatusOK, got.Code, got.Body.String())
	require.Equal(t, http.StatusOK, put.Code, put.Body.String())
	var body gen.FiledPayments
	require.NoError(t, json.Unmarshal(got.Body.Bytes(), &body))
	require.Len(t, body.Months, 12)
	assert.Equal(t, gen.FiledPaymentSource("filed"), body.Months[2].Source)
	assert.Equal(t, gen.FiledPaymentSource("estimated_by_app"), body.Months[0].Source)
	assert.Equal(t, []string{userA, userA}, w.rng.keys)
	assert.Equal(t, []string{"get|" + userA + "|" + fisicaID + "|612", "put|" + userA + "|" + fisicaID + "|612"}, w.filed.calls)
}

// A put whose estimate_reads budget is spent is refused before the write:
// a 429 never follows a committed (and audited) replace of the months.
func TestPutFiledPayments_rangeBudgetSpentRefusesBeforeTheWrite(t *testing.T) {
	c := contracttest.Load(t)
	w := newWorld(t)
	w.rng.allowed = false
	h := newHarness(t, w.deps, annualOn, nil)
	target := "/v1/taxpayers/" + fisicaID + "/regimes/612/filed-payments/2026"

	rec := c.ServeOp(t, contracttest.Op("putFiledPayments"), h.handler,
		request(http.MethodPut, target, `{"months":[{"month":3,"amount_cents":200000}]}`))

	assert.Equal(t, http.StatusTooManyRequests, rec.Code, rec.Body.String())
	assert.Empty(t, w.filed.calls, "nothing was written")
}

// The input operations conform to the contract and pass the principal,
// the path ids and the merge-patch null through.
func TestInputOperations(t *testing.T) {
	c := contracttest.Load(t)
	w := newWorld(t)
	h := newHarness(t, w.deps, annualOn, nil)
	base := "/v1/taxpayers/" + fisicaID
	for _, tc := range []struct {
		op                   contracttest.Operation
		method, target, body string
		status               int
	}{
		{contracttest.Op("getAnnualInputs"), http.MethodGet, base + "/annual-inputs/2026", "", http.StatusOK},
		{contracttest.Op("putAnnualInputs"), http.MethodPut, base + "/annual-inputs/2026",
			`{"prior_year_accumulable_income_cents":null,"other_exempt_income_cents":0,"told_employer_will_file":true}`, http.StatusOK},
		{contracttest.Op("listSalaryStatements"), http.MethodGet, base + "/salary-statements?fiscal_year=2026", "", http.StatusOK},
		{contracttest.Op("createSalaryStatement"), http.MethodPost, base + "/salary-statements",
			`{"fiscal_year":2026,"accumulable_cents":30000000,"exempt_cents":0,"isr_withheld_cents":4000000}`, http.StatusCreated},
		{contracttest.Op("createSalaryStatement"), http.MethodPost, base + "/salary-statements",
			`{"id":"` + rowID + `","fiscal_year":2026,"accumulable_cents":30000000,"exempt_cents":0,"isr_withheld_cents":4000000}`, http.StatusOK},
		{contracttest.Op("updateSalaryStatement"), http.MethodPatch, base + "/salary-statements/" + rowID, `{"employer_label":null}`, http.StatusOK},
		{contracttest.Op("deleteSalaryStatement"), http.MethodDelete, base + "/salary-statements/" + rowID, "", http.StatusNoContent},
		{contracttest.Op("listPersonalDeductions"), http.MethodGet, base + "/personal-deductions?fiscal_year=2026&category=medical", "", http.StatusOK},
		{contracttest.Op("createPersonalDeduction"), http.MethodPost, base + "/personal-deductions",
			`{"category":"medical","amount_cents":120000,"paid_on":"2026-05-10","payment_method":"cash","beneficiary":"self"}`, http.StatusCreated},
		{contracttest.Op("updatePersonalDeduction"), http.MethodPatch, base + "/personal-deductions/" + rowID, `{"cfdi_uuid":null}`, http.StatusOK},
		{contracttest.Op("deletePersonalDeduction"), http.MethodDelete, base + "/personal-deductions/" + rowID, "", http.StatusNoContent},
		{contracttest.Op("restorePersonalDeduction"), http.MethodPost, base + "/personal-deductions/" + rowID + "/restore", "", http.StatusOK},
	} {
		t.Run(tc.method+" "+tc.target, func(t *testing.T) {
			rec := c.ServeOp(t, tc.op, h.handler, request(tc.method, tc.target, tc.body))

			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
		})
	}
	assert.Contains(t, w.inputs.calls, "salary.update|"+userA+"|"+fisicaID+"|"+rowID+"|null")
	assert.Contains(t, w.deductions.calls, "update|"+userA+"|"+fisicaID+"|"+rowID+"|cfdi:null")
	assert.Contains(t, w.deductions.calls, "create|"+userA+"|"+fisicaID+"|2026-05-10")
	assert.Contains(t, w.deductions.calls, "list|"+userA+"|"+fisicaID+"|medical")
	assert.Empty(t, w.rng.keys, "the inputs never spend estimate reads")
}

// module_annual OFF (or implied OFF by a prerequisite): every annual route
// is 404 before authentication or validation (US-A13).
func TestAnnualRoutesHiddenWhenOff(t *testing.T) {
	h := newHarness(t, newWorld(t).deps, flagtest.Static{"module_tax_engine": true, "module_taxpayers": true}, nil)
	base := "/v1/taxpayers/" + fisicaID
	for _, r := range []struct{ method, target string }{
		{http.MethodGet, base + "/annual-estimates/2026"},
		{http.MethodGet, base + "/annual-estimates/2026/breakdown?section=pf_art152"},
		{http.MethodGet, base + "/annual-inputs/2026"},
		{http.MethodPut, base + "/annual-inputs/2026"},
		{http.MethodGet, base + "/salary-statements?fiscal_year=2026"},
		{http.MethodPost, base + "/salary-statements"},
		{http.MethodPatch, base + "/salary-statements/" + rowID},
		{http.MethodDelete, base + "/salary-statements/" + rowID},
		{http.MethodGet, base + "/personal-deductions?fiscal_year=2026"},
		{http.MethodPost, base + "/personal-deductions"},
		{http.MethodPatch, base + "/personal-deductions/" + rowID},
		{http.MethodDelete, base + "/personal-deductions/" + rowID},
		{http.MethodPost, base + "/personal-deductions/" + rowID + "/restore"},
		{http.MethodGet, base + "/regimes/612/filed-payments/2026"},
		{http.MethodPut, base + "/regimes/612/filed-payments/2026"},
	} {
		req := httptest.NewRequest(r.method, r.target, nil) // no principal, no body
		rec := httptest.NewRecorder()

		h.handler.ServeHTTP(rec, req)

		flagtest.AssertHidden(t, rec)
	}
}

// Every annual operation has a per-user budget: reads that show the app's
// estimates spend estimate_reads, the other reads annual_reads, writes
// annual_writes.
func TestUserLimits(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	scopes := map[string]string{}
	for _, item := range spec.Paths.Map() {
		for _, op := range item.Operations() {
			if op.Extensions["x-module"] != "module_annual" {
				continue
			}
			policy, ok := annualhttp.UserLimits[op.OperationID]
			require.Truef(t, ok, "%s has no budget", op.OperationID)
			scopes[op.OperationID] = policy.Scope
		}
	}
	assert.Len(t, scopes, 15)
	assert.Equal(t, "estimate_reads", scopes["getAnnualEstimate"])
	assert.Equal(t, "estimate_reads", scopes["putFiledPayments"])
	assert.Equal(t, "annual_reads", scopes["listPersonalDeductions"])
	assert.Equal(t, "annual_writes", scopes["restorePersonalDeduction"])
}
