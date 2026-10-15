package historyhttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	historyhttp "github.com/romodevx/control-contable/backend/internal/history/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

const (
	callerID   = "0199d4a2-0000-7000-8000-00000000e0aa"
	exampleTax = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80" // the contract examples' taxpayer
)

type fakeHistory struct {
	calls  []string
	year   usecase.Year
	viewed []usecase.ViewedRequest
	err    error
}

func (f *fakeHistory) Get(_ context.Context, userID, taxpayerID string, fiscalYear int) (usecase.Year, error) {
	f.calls = append(f.calls, userID+"|"+taxpayerID+"|"+strconv.Itoa(fiscalYear))
	return f.year, f.err
}

func (f *fakeHistory) MarkViewed(_ context.Context, userID, taxpayerID string, req usecase.ViewedRequest) error {
	f.calls = append(f.calls, userID+"|"+taxpayerID)
	f.viewed = append(f.viewed, req)
	return f.err
}

func asCaller(t *testing.T) context.Context {
	t.Helper()
	return principal.With(t.Context(), principal.Principal{UserID: callerID})
}

// schemaExample is the first example of a component schema.
func schemaExample(t *testing.T, schema string) string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	ref := spec.Components.Schemas[schema]
	require.NotNil(t, ref, schema)
	require.NotEmpty(t, ref.Value.Examples, schema)
	raw, err := json.Marshal(ref.Value.Examples[0])
	require.NoError(t, err)
	return string(raw)
}

func month(t *testing.T, raw string) domain.Month {
	t.Helper()
	m, err := domain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

func hex64(tag string) string { return strings.Repeat(tag[:1], 58) + tag[1:] }

// juneChanged is the contract's HistoryMonth example: June 2026, viewed on
// 20 September, changed by records.
func juneChanged(t *testing.T) usecase.Month {
	viewed := time.Date(2026, time.September, 20, 15, 30, 0, 0, time.UTC)
	return usecase.Month{
		Month: month(t, "2026-06"), Status: domain.StatusComplete, HasRecords: true,
		Tax:     &domain.Tax{ISRPayable: 18000, IVAPayable: 96000, IVAOutcome: "payable", TotalPayable: 114000},
		DueDate: "2026-07-17", ParameterSet: domain.ParameterSet{FiscalYear: 2026, Version: 1},
		Seen:     domain.Seen{RecordsFingerprint: hex64("a202606"), FiguresFingerprint: hex64("b202606"), EngineVersion: "1.0.1"},
		ViewedAt: &viewed, Change: domain.ChangeRecords,
	}
}

func get(t *testing.T, h *historyhttp.Handler, year int) (*httptest.ResponseRecorder, error) {
	t.Helper()
	resp, err := h.GetHistory(asCaller(t), gen.GetHistoryRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax), Params: gen.GetHistoryParams{FiscalYear: year},
	})
	if err != nil {
		return nil, err
	}
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetHistoryResponse(rec))
	return rec, nil
}

type historyBody struct {
	TaxpayerID  string            `json:"taxpayer_id"`
	FiscalYear  int               `json:"fiscal_year"`
	Months      []json.RawMessage `json:"months"`
	YearTotals  json.RawMessage   `json:"year_totals"`
	Legend      string            `json:"legend"`
	Disclaimers json.RawMessage   `json:"disclaimers"`
}

// Each month is the contract's HistoryMonth field for field; a never
// viewed, no-regime month sends its nulls; the year carries its totals,
// legend and disclaimers.
func TestGetHistory_body(t *testing.T) {
	none := usecase.Month{
		Month: month(t, "2026-01"), Status: domain.StatusNone, DueDate: "2026-02-17",
		ParameterSet: domain.ParameterSet{FiscalYear: 2026, Version: 1},
		Seen:         domain.Seen{RecordsFingerprint: hex64("a202601"), FiguresFingerprint: hex64("b202601"), EngineVersion: "1.0.1"},
	}
	uc := &fakeHistory{year: usecase.Year{
		TaxpayerID: exampleTax, FiscalYear: 2026, Months: []usecase.Month{juneChanged(t), none},
		Totals: domain.YearTotals{ISRPayable: 91096, IVAPayable: 675333, WithholdingsToRemit: 12500, TotalPayable: 778929},
		Legend: "legend", Disclaimers: []domain.Disclaimer{{Code: "not_tax_advice", Severity: domain.SeverityInfo, Text: "n"}},
	}}

	rec, err := get(t, historyhttp.NewHandler(uc), 2026)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.Equal(t, []string{callerID + "|" + exampleTax + "|2026"}, uc.calls)
	var body historyBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, exampleTax, body.TaxpayerID)
	assert.Equal(t, 2026, body.FiscalYear)
	require.Len(t, body.Months, 2)
	assert.JSONEq(t, schemaExample(t, "HistoryMonth"), string(body.Months[0]))
	assert.JSONEq(t, `{"period":"2026-01","status":"none","has_records":false,"tax":null,"due_date":"2026-02-17",
		"cpa_review_pending":false,"is_current":false,"records_fingerprint":"`+hex64("a202601")+`",
		"figures_fingerprint":"`+hex64("b202601")+`","engine_version":"1.0.1","parameter_set":{"fiscal_year":2026,"version":1},
		"viewed_at":null,"changed_since_viewed":false,"change_reason":null}`, string(body.Months[1]))
	assert.JSONEq(t, schemaExample(t, "HistoryYearTotals"), string(body.YearTotals))
	assert.Equal(t, "legend", body.Legend)
	assert.JSONEq(t, `[{"code":"not_tax_advice","severity":"info","text":"n"}]`, string(body.Disclaimers))
}

// An in-favor month sends its saldo a favor; the engine reason.
func TestGetHistory_inFavorAndEngine(t *testing.T) {
	m := juneChanged(t)
	m.Tax = &domain.Tax{ISRPayable: 1, IVAOutcome: "in_favor", IVAInFavor: 5000, TotalPayable: 1}
	m.Change = domain.ChangeEngine
	uc := &fakeHistory{year: usecase.Year{TaxpayerID: exampleTax, FiscalYear: 2026, Months: []usecase.Month{m}}}

	rec, err := get(t, historyhttp.NewHandler(uc), 2026)

	require.NoError(t, err)
	var body struct {
		Months []struct {
			Tax struct {
				IVAOutcome string `json:"iva_outcome"`
				InFavor    int64  `json:"iva_in_favor_cents"`
			} `json:"tax"`
			Changed bool   `json:"changed_since_viewed"`
			Reason  string `json:"change_reason"`
		} `json:"months"`
		Disclaimers []json.RawMessage `json:"disclaimers"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "in_favor", body.Months[0].Tax.IVAOutcome)
	assert.Equal(t, int64(5000), body.Months[0].Tax.InFavor)
	assert.True(t, body.Months[0].Changed)
	assert.Equal(t, "engine", body.Months[0].Reason)
	assert.NotNil(t, body.Disclaimers, "an array, never null")
}

func TestGetHistory_passesErrors(t *testing.T) {
	want := apperr.New(apperr.CodeNotFound, "Not found.")
	_, err := get(t, historyhttp.NewHandler(&fakeHistory{err: want}), 2026)
	assert.ErrorIs(t, err, want)
}

func viewedBody() *gen.MarkHistoryViewedJSONRequestBody {
	return &gen.MarkHistoryViewedJSONRequestBody{
		RecordsFingerprint: hex64("a202603"), FiguresFingerprint: hex64("b202603"), EngineVersion: "1.0.1",
		ParameterSet: gen.HistoryViewedParameterSet{FiscalYear: 2026, Version: 1},
	}
}

func TestMarkHistoryViewed_204(t *testing.T) {
	uc := &fakeHistory{}

	resp, err := historyhttp.NewHandler(uc).MarkHistoryViewed(asCaller(t), gen.MarkHistoryViewedRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax), Period: "2026-03", Body: viewedBody(),
	})

	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitMarkHistoryViewedResponse(rec))
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	assert.Equal(t, []string{callerID + "|" + exampleTax}, uc.calls)
	assert.Equal(t, []usecase.ViewedRequest{{
		Period: "2026-03", RecordsFingerprint: hex64("a202603"), FiguresFingerprint: hex64("b202603"),
		EngineVersion: "1.0.1", ParamsYear: 2026, ParamsVersion: 1,
	}}, uc.viewed)
}

func TestMarkHistoryViewed_errors(t *testing.T) {
	want := apperr.New(apperr.CodeNotFound, "Not found.")
	_, err := historyhttp.NewHandler(&fakeHistory{err: want}).MarkHistoryViewed(asCaller(t), gen.MarkHistoryViewedRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax), Period: "2026-03", Body: viewedBody(),
	})
	assert.ErrorIs(t, err, want)

	uc := &fakeHistory{}
	_, err = historyhttp.NewHandler(uc).MarkHistoryViewed(asCaller(t), gen.MarkHistoryViewedRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax), Period: "2026-03",
	})
	app, ok := err.(*apperr.Error) //nolint:errorlint // the handler returns the *apperr.Error itself
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeValidationFailed, app.Code, "a missing body is the router's 400, backstopped")
	assert.Empty(t, uc.calls)
}

// No principal is the contract's 401 with the bearer challenge (the router
// authenticates first; this is the backstop).
func TestHandlers_unauthorized(t *testing.T) {
	h := historyhttp.NewHandler(&fakeHistory{})

	resp, err := h.GetHistory(t.Context(), gen.GetHistoryRequestObject{TaxpayerId: uuid.MustParse(exampleTax)})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetHistoryResponse(rec))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, `Bearer realm="control-contable"`, rec.Header().Get("WWW-Authenticate"))

	viewed, err := h.MarkHistoryViewed(t.Context(), gen.MarkHistoryViewedRequestObject{TaxpayerId: uuid.MustParse(exampleTax), Body: viewedBody()})
	require.NoError(t, err)
	rec = httptest.NewRecorder()
	require.NoError(t, viewed.VisitMarkHistoryViewedResponse(rec))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// §2.7: own per-user budgets, separate from estimate_reads.
func TestUserLimits(t *testing.T) {
	assert.Equal(t, map[string]ratelimit.Policy{
		"getHistory":        {Scope: "history_reads", Burst: 600, Interval: 6 * time.Second},
		"markHistoryViewed": {Scope: "history_views", Burst: 120, Interval: 3 * time.Second},
	}, historyhttp.UserLimits)
	for _, p := range historyhttp.UserLimits {
		require.NoError(t, p.Validate())
	}
}
