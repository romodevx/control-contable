package taxenginehttp_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// fakeEstimates records the call and answers the canned estimate,
// breakdown or error.
type fakeEstimates struct {
	calls     []string
	estimate  usecase.TaxpayerEstimate
	breakdown usecase.RegimeBreakdown
	err       error
}

func (f *fakeEstimates) EstimateMonth(_ context.Context, userID, taxpayerID, period string) (usecase.TaxpayerEstimate, error) {
	f.calls = append(f.calls, "estimate|"+userID+"|"+taxpayerID+"|"+period)
	return f.estimate, f.err
}

func (f *fakeEstimates) Breakdown(_ context.Context, userID, taxpayerID, period, regime string) (usecase.RegimeBreakdown, error) {
	f.calls = append(f.calls, "breakdown|"+userID+"|"+taxpayerID+"|"+period+"|"+regime)
	return f.breakdown, f.err
}

// exampleEngineVersion is the engine_version the contract examples show.
// The handler passes the use case's version through, so the fakes answer
// the examples' one rather than domain.EngineVersion, which moves with
// every engine release.
const exampleEngineVersion = "1.0.0"

// contractExample is the named 200 example of operationID in the served
// contract, re-encoded as JSON.
func contractExample(t *testing.T, operationID, name string) string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	for _, item := range spec.Paths.Map() {
		for _, op := range item.Operations() {
			if op.OperationID != operationID {
				continue
			}
			ex := op.Responses.Status(200).Value.Content.Get("application/json").Examples[name]
			require.NotNilf(t, ex, "%s has no example %q", operationID, name)
			raw, err := json.Marshal(ex.Value.Value)
			require.NoError(t, err)
			return string(raw)
		}
	}
	t.Fatalf("no operation %s", operationID)
	return ""
}

func mustDate(t *testing.T, s string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(s)
	require.NoError(t, err)
	return d
}

func taxResult(c money.Cents) *domain.TaxResult {
	r := domain.NewTaxResult(money.Exact(c) * money.ExactPerCent)
	return &r
}

var (
	name612 = usecase.RegimeText{Name: "Actividades Empresariales y Profesionales"}
	name606 = usecase.RegimeText{Name: "Arrendamiento"}
	netting = []domain.CPATopic{domain.TopicCrossRegimeIVANetting}
)

// regime612Net is the 612 regime of TX-NET-01 as the use case returns it.
func regime612Net() usecase.RegimeResult {
	return usecase.RegimeResult{
		RegimeEstimate: domain.RegimeEstimate{
			Key: domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}, Coverage: domain.TreatmentFull,
			ISR:       taxResult(15096),
			IVA:       &domain.IVAComponents{ChargedCollected: 480000, CreditablePaid: 160000, WithheldFromTaxpayer: 106667},
			CPATopics: netting,
		},
		Treatment: domain.TreatmentFull, Status: domain.StatusComplete, Text: name612,
	}
}

var nettingDisclaimers = []domain.Disclaimer{
	{Code: domain.DisclaimerNotTaxAdvice, Severity: domain.SeverityInfo},
	{Code: domain.DisclaimerCPAReviewPending, Severity: domain.SeverityStrong, Topics: netting},
	{Code: domain.DisclaimerDeductibilityClaimed, Severity: domain.SeverityInfo},
}

// estimateNet is the contract's crossRegimeIva example (TX-NET-01) as the
// use case returns it.
func estimateNet(t *testing.T) usecase.TaxpayerEstimate {
	t.Helper()
	regime606 := usecase.RegimeResult{
		RegimeEstimate: domain.RegimeEstimate{
			Key: domain.StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica}, Coverage: domain.TreatmentFull,
			ISR:       taxResult(-92400),
			IVA:       &domain.IVAComponents{ChargedCollected: 320000, CreditablePaid: 300000, WithheldFromTaxpayer: 213333},
			CPATopics: netting,
		},
		Treatment: domain.TreatmentFull, Status: domain.StatusComplete, Text: name606,
	}
	iva := *taxResult(20000)
	return usecase.TaxpayerEstimate{
		TaxpayerID: "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80", PersonType: fiscal.PersonFisica,
		Period: march(t), DueDate: mustDate(t, "2026-04-17"), Legend: domain.Legend,
		Disclaimers: nettingDisclaimers,
		Regimes:     []usecase.RegimeResult{regime606, regime612Net()},
		IVA:         domain.ConsolidatedIVA{Result: iva},
		Totals: domain.Totals{
			IVA: iva, ISR: *taxResult(15096), TotalPayable: 35096,
		},
		ParameterSet:       taxparams.SetID{FiscalYear: 2026, Version: 1},
		RecordsFingerprint: "5e2a8c1f4b7d0e3a6c9f2b5e8d1a4c7f0b3e6a9d2c5f8b1e4a7d0c3f6b9e2a5d",
		EngineVersion:      exampleEngineVersion,
	}
}

// The body is the contract example, byte for byte as JSON: legend,
// disclaimers with their texts, nulls, the per-regime parameter set.
func TestGetTaxpayerEstimate_isTheContractExample(t *testing.T) {
	uc := &fakeEstimates{estimate: estimateNet(t)}
	period := "2026-03"
	id := uuid.MustParse(taxpayerID)

	resp, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetTaxpayerEstimate(authed(t), gen.GetTaxpayerEstimateRequestObject{
		TaxpayerId: id, Params: gen.GetTaxpayerEstimateParams{Period: &period},
	})

	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxpayerEstimateResponse(rec))
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, contractExample(t, "getTaxpayerEstimate", "crossRegimeIva"), rec.Body.String())
	assert.Equal(t, []string{"estimate|" + callerID + "|" + taxpayerID + "|2026-03"}, uc.calls)
}

// No period is the current month: the use case gets "".
func TestGetTaxpayerEstimate_withoutPeriod(t *testing.T) {
	uc := &fakeEstimates{estimate: estimateNet(t)}

	_, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetTaxpayerEstimate(authed(t), gen.GetTaxpayerEstimateRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID),
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"estimate|" + callerID + "|" + taxpayerID + "|"}, uc.calls)
}

// Empty lists are arrays, missing inputs carry the catalog label and their
// months, an informational regime has null figures and its explanation.
func TestGetTaxpayerEstimate_regimeShapes(t *testing.T) {
	est := estimateNet(t)
	incomplete := regime612Net()
	incomplete.CPATopics, incomplete.Status = nil, domain.StatusIncomplete
	incomplete.MissingInputs = []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments, Months: []time.Month{time.January, time.February}}}
	incomplete.Text.InputLabels = map[domain.InputKey]string{domain.InputPriorProvisionalPayments: "Pagos provisionales anteriores del ejercicio"}
	key605 := domain.StrategyKey{Regime: "605", PersonType: fiscal.PersonFisica}
	info := usecase.RegimeResult{
		RegimeEstimate: domain.RegimeEstimate{Key: key605, Coverage: domain.TreatmentInformational, Explanation: domain.RegimeExplanation(key605)},
		Treatment:      domain.TreatmentInformational, Status: domain.StatusInformational,
		Text: usecase.RegimeText{Name: "Sueldos y Salarios", Explanation: "Tu patrón calcula y retiene tu ISR."},
	}
	est.Regimes = []usecase.RegimeResult{info, incomplete}
	uc := &fakeEstimates{estimate: est}

	resp, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetTaxpayerEstimate(authed(t), gen.GetTaxpayerEstimateRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID),
	})

	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxpayerEstimateResponse(rec))
	var body struct {
		Regimes []json.RawMessage `json:"regimes"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Regimes, 2)
	assert.JSONEq(t, `{"regime_code":"605","regime_name":"Sueldos y Salarios","treatment":"informational","status":"informational",
		"missing_inputs":[],"explanation":"Tu patrón calcula y retiene tu ISR.","isr":null,"iva_components":null,
		"withheld_to_remit_cents":0,"cpa_review_pending":false,"cpa_topics":[],"parameter_set":{"fiscal_year":2026,"version":1}}`,
		string(body.Regimes[0]))
	var regime612 struct {
		MissingInputs json.RawMessage `json:"missing_inputs"`
		CPATopics     json.RawMessage `json:"cpa_topics"`
	}
	require.NoError(t, json.Unmarshal(body.Regimes[1], &regime612))
	assert.JSONEq(t, `[{"key":"prior_provisional_payments","label":"Pagos provisionales anteriores del ejercicio",
		"months":["2026-01","2026-02"]}]`, string(regime612.MissingInputs))
	assert.JSONEq(t, `[]`, string(regime612.CPATopics))
}

// A use case error passes as is: the router writes the envelope.
func TestGetTaxpayerEstimate_passesUseCaseErrors(t *testing.T) {
	want := apperr.New(apperr.CodeNotFound, "Not found.")
	uc := &fakeEstimates{err: want}

	_, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetTaxpayerEstimate(authed(t), gen.GetTaxpayerEstimateRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID),
	})

	assert.ErrorIs(t, err, want)
}

// Without a principal both reads answer the contract's 401 and never reach
// the use case.
func TestEstimateReads_withoutPrincipalAre401(t *testing.T) {
	uc := &fakeEstimates{}
	h := taxenginehttp.NewHandler(&fakeInputs{}, uc)

	est, err := h.GetTaxpayerEstimate(t.Context(), gen.GetTaxpayerEstimateRequestObject{TaxpayerId: uuid.MustParse(taxpayerID)})
	require.NoError(t, err)
	assert.IsType(t, gen.GetTaxpayerEstimate401JSONResponse{}, est)
	bd, err := h.GetEstimateBreakdown(t.Context(), gen.GetEstimateBreakdownRequestObject{TaxpayerId: uuid.MustParse(taxpayerID)})
	require.NoError(t, err)
	assert.IsType(t, gen.GetEstimateBreakdown401JSONResponse{}, bd)
	assert.Empty(t, uc.calls)
}

// breakdown612 is the contract's crossRegimeIva612 example up to the
// regime's own IVA result (seq 1..13), as the use case returns it.
func breakdown612(t *testing.T) usecase.RegimeBreakdown {
	t.Helper()
	cite := func(law, article, note string) domain.Citation {
		return domain.Citation{Law: law, Article: article, Note: note}
	}
	tariff := domain.SourceParams("isr.art106.accumulated") // gitleaks:allow (a public parameter key)
	rate, err := money.ParseRate("0.2136")
	require.NoError(t, err)
	lines := []domain.Line{
		{Tax: domain.TaxISR, Label: "isr.income_ytd", Args: []string{"2026-01", "2026-03"}, Operation: domain.OpInput, Amount: 9000000,
			Basis: []domain.Citation{cite("LISR", "102", "")}, Source: domain.SourceRecords()},
		{Tax: domain.TaxISR, Label: "isr.deductions_ytd", Args: []string{"2026-01", "2026-03"}, Operation: domain.OpSubtract, Amount: 3000000,
			Basis: []domain.Citation{cite("LISR", "103", ""), cite("LISR", "105", "")}, Source: domain.SourceRecords()},
		{Tax: domain.TaxISR, Label: "isr.base", Operation: domain.OpResult, Amount: 6000000,
			Basis: []domain.Citation{cite("LISR", "106", "párrafo primero")}, Source: domain.SourceComputed()},
		{Tax: domain.TaxISR, Label: "isr.tariff_row", Args: []string{"2026-03"}, Row: 6, Operation: domain.OpLookup, Amount: 557052,
			Basis: []domain.Citation{cite("Anexo 8 RMF 2026", "B.VI", "")}, Source: tariff},
		{Tax: domain.TaxISR, Label: "isr.tariff_excess", Operation: domain.OpMultiply, Amount: 158044, Rate: rate,
			Basis: []domain.Citation{cite("Anexo 8 RMF 2026", "B.VI", "")}, Source: tariff},
		{Tax: domain.TaxISR, Label: "isr.caused", Operation: domain.OpResult, Amount: 715096,
			Basis: []domain.Citation{cite("LISR", "106", "párrafo primero")}, Source: domain.SourceComputed()},
		{Tax: domain.TaxISR, Label: "isr.prior_payments", Operation: domain.OpSubtract, Amount: 400000,
			Basis: []domain.Citation{cite("LISR", "106", "párrafo cuarto")}, Source: domain.SourceTaxInput(domain.InputPriorProvisionalPayments)},
		{Tax: domain.TaxISR, Label: "isr.withheld_by_moral", Operation: domain.OpSubtract, Amount: 300000,
			Basis: []domain.Citation{cite("LISR", "106", "párrafo quinto")}, Source: domain.SourceRecords()},
		{Tax: domain.TaxISR, Label: "isr.payable", Operation: domain.OpResult, Amount: 15096,
			Basis: []domain.Citation{cite("LISR", "106", "")}, Source: domain.SourceComputed()},
		{Tax: domain.TaxIVA, Label: "iva.charged", Operation: domain.OpInput, Amount: 480000,
			Basis: []domain.Citation{cite("LIVA", "1", ""), cite("LIVA", "1-B", "")}, Source: domain.SourceRecords()},
		{Tax: domain.TaxIVA, Label: "iva.creditable", Operation: domain.OpSubtract, Amount: 160000,
			Basis: []domain.Citation{cite("LIVA", "5", "fracciones I a V")}, Source: domain.SourceRecords()},
		{Tax: domain.TaxIVA, Label: "iva.withheld.moral", Operation: domain.OpSubtract, Amount: 106667,
			Basis: []domain.Citation{cite("LIVA", "1-A", ""), cite("RLIVA", "3", "")}, Source: domain.SourceRecords()},
		{Tax: domain.TaxIVA, Label: "iva.regime_net", Operation: domain.OpResult, Amount: 213333,
			Basis: []domain.Citation{cite("LIVA", "5-D", "")}, Source: domain.SourceComputed()},
	}
	for i := range lines {
		lines[i].Seq = i + 1
	}
	return usecase.RegimeBreakdown{
		TaxpayerID: "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80", PersonType: fiscal.PersonFisica, Period: march(t),
		Regime: regime612Net(), ParameterSet: taxparams.SetID{FiscalYear: 2026, Version: 1},
		Lines: lines, Legend: domain.Legend, Disclaimers: nettingDisclaimers, EngineVersion: exampleEngineVersion,
	}
}

// The breakdown body matches the contract example field by field; its
// lines 1..13 (the regime's own) are the example's exactly.
func TestGetEstimateBreakdown_isTheContractExample(t *testing.T) {
	uc := &fakeEstimates{breakdown: breakdown612(t)}

	resp, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetEstimateBreakdown(authed(t), gen.GetEstimateBreakdownRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), Period: "2026-03", Params: gen.GetEstimateBreakdownParams{Regime: "612"},
	})

	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetEstimateBreakdownResponse(rec))
	assert.Equal(t, 200, rec.Code)
	var got, want map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.NoError(t, json.Unmarshal([]byte(contractExample(t, "getEstimateBreakdown", "crossRegimeIva612")), &want))
	for _, field := range []string{"taxpayer_id", "period", "regime", "legend", "disclaimers", "engine_version"} {
		assert.JSONEqf(t, string(want[field]), string(got[field]), "field %s", field)
	}
	var gotLines, wantLines []json.RawMessage
	require.NoError(t, json.Unmarshal(got["lines"], &gotLines))
	require.NoError(t, json.Unmarshal(want["lines"], &wantLines))
	require.Len(t, gotLines, 13)
	for i := range gotLines {
		assert.JSONEqf(t, string(wantLines[i]), string(gotLines[i]), "line %d", i+1)
	}
	assert.Equal(t, []string{"breakdown|" + callerID + "|" + taxpayerID + "|2026-03|612"}, uc.calls)
}

// The breakdown's own error passes as is.
func TestGetEstimateBreakdown_passesUseCaseErrors(t *testing.T) {
	want := apperr.New(apperr.CodeBusinessRuleViolation, "The taxpayer does not hold this regime in that period.")
	uc := &fakeEstimates{err: want}

	_, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetEstimateBreakdown(authed(t), gen.GetEstimateBreakdownRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), Period: "2026-03", Params: gen.GetEstimateBreakdownParams{Regime: "626"},
	})

	assert.ErrorIs(t, err, want)
}

// Disclaimer texts with variable parts: the provisional year, the RESICO
// cap by person type, several CPA topics joined the Spanish way.
func TestGetTaxpayerEstimate_variableDisclaimerTexts(t *testing.T) {
	est := estimateNet(t)
	est.PersonType = fiscal.PersonMoral
	est.Period = mustPeriod(t, "2027-01")
	est.ParameterSet = taxparams.SetID{FiscalYear: 2026, Version: 1}
	est.Disclaimers = []domain.Disclaimer{
		{Code: domain.DisclaimerNotTaxAdvice, Severity: domain.SeverityInfo},
		{Code: domain.DisclaimerCPAReviewPending, Severity: domain.SeverityStrong, Topics: []domain.CPATopic{
			domain.TopicCrossRegimeIVANetting, domain.TopicRIFBimonthlyIVA, domain.TopicRIFReduction,
		}},
		{Code: domain.DisclaimerProvisionalParameters, Severity: domain.SeverityWarning},
		{Code: domain.DisclaimerRESICOCapExceeded, Severity: domain.SeverityWarning},
	}
	uc := &fakeEstimates{estimate: est}

	resp, err := taxenginehttp.NewHandler(&fakeInputs{}, uc).GetTaxpayerEstimate(authed(t), gen.GetTaxpayerEstimateRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID),
	})

	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxpayerEstimateResponse(rec))
	var body struct {
		Disclaimers []gen.Disclaimer `json:"disclaimers"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Disclaimers, 4)
	assert.Equal(t, "Este cálculo usa reglas que un contador aún no valida: compensación del IVA entre regímenes, "+
		"el IVA bimestral del RIF y la reducción del RIF. Tómalo con mayor reserva.", body.Disclaimers[1].Text)
	assert.Equal(t, "Aún no se publican las tarifas de 2027; usamos las de 2026.", body.Disclaimers[2].Text)
	assert.Contains(t, body.Disclaimers[3].Text, "$35,000,000", "the PM cap")
}

func mustPeriod(t *testing.T, s string) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod(s)
	require.NoError(t, err)
	return p
}
