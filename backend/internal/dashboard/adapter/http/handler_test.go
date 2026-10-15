package dashboardhttp_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	dashboardhttp "github.com/romodevx/control-contable/backend/internal/dashboard/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	"github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
)

const (
	callerID   = "0199d4a2-0000-7000-8000-00000000d0aa"
	exampleTax = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80" // the contract examples' taxpayer
)

// The hero passes through as raw JSON: the handler never looks inside it.
type hero = json.RawMessage

type fakeDashboards struct {
	calls []string
	out   usecase.Dashboard[hero]
	err   error
}

func (f *fakeDashboards) Get(_ context.Context, userID string, req usecase.Request) (usecase.Dashboard[hero], error) {
	f.calls = append(f.calls, userID+"|"+req.TaxpayerID+"|"+req.Period+"|"+strconv.Itoa(req.Months))
	return f.out, f.err
}

func asCaller(t *testing.T) context.Context {
	t.Helper()
	return principal.With(t.Context(), principal.Principal{UserID: callerID})
}

// contractExample is the named 200 example of getDashboard.
func contractExample(t *testing.T, name string) map[string]json.RawMessage {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	op := spec.Paths.Find("/v1/taxpayers/{taxpayerId}/dashboard").Get
	ex := op.Responses.Status(200).Value.Content.Get("application/json").Examples[name]
	require.NotNil(t, ex, name)
	raw, err := json.Marshal(ex.Value.Value)
	require.NoError(t, err)
	var out map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func cents(c money.Cents) *money.Cents { return &c }

func month(t *testing.T, raw string) domain.Month {
	t.Helper()
	m, err := domain.ParseMonth(raw)
	require.NoError(t, err)
	return m
}

func serve(t *testing.T, h *dashboardhttp.Handler[hero], req gen.GetDashboardRequestObject) (*httptest.ResponseRecorder, error) {
	t.Helper()
	resp, err := h.GetDashboard(asCaller(t), req)
	if err != nil {
		return nil, err
	}
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetDashboardResponse(rec))
	return rec, nil
}

func assertExample(t *testing.T, rec *httptest.ResponseRecorder, name string) {
	t.Helper()
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	want := contractExample(t, name)
	assert.Len(t, got, len(want), "exactly the example's fields: %s", rec.Body.String())
	for field, value := range want {
		assert.JSONEqf(t, string(value), string(got[field]), "field %s", field)
	}
}

func readOnlyEngineOff(t *testing.T) usecase.Dashboard[hero] {
	t.Helper()
	zero := 0
	return usecase.Dashboard[hero]{
		TaxpayerID: exampleTax, Period: month(t, "2026-02"),
		Current: usecase.Figures{Income: cents(2_500_000), Expense: cents(2_900_000)},
		Trend: []usecase.TrendPoint{
			{Month: month(t, "2026-01"), Figures: usecase.Figures{Income: cents(0), Expense: cents(0)}},
			{Month: month(t, "2026-02"), Figures: usecase.Figures{Income: cents(2_500_000), Expense: cents(2_900_000)}},
		},
		Access: access.View{Level: access.LevelReadOnly, IncomesRemaining: &zero, ExpensesRemaining: &zero},
	}
}

func TestGetDashboard_readOnlyEngineOffExample(t *testing.T) {
	uc := &fakeDashboards{out: readOnlyEngineOff(t)}
	period, months := "2026-02", 6

	rec, err := serve(t, dashboardhttp.NewHandler[hero](uc), gen.GetDashboardRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax), Params: gen.GetDashboardParams{Period: &period, Months: &months},
	})

	require.NoError(t, err)
	assertExample(t, rec, "readOnlyEngineOff")
	assert.Equal(t, []string{callerID + "|" + exampleTax + "|2026-02|6"}, uc.calls)
}

func TestGetDashboard_incomesAndEngineOffExample(t *testing.T) {
	out := readOnlyEngineOff(t)
	out.Current.Income = nil
	for i := range out.Trend {
		out.Trend[i].Income = nil
	}
	out.Access = access.View{Level: access.LevelPro}

	rec, err := serve(t, dashboardhttp.NewHandler[hero](&fakeDashboards{out: out}), gen.GetDashboardRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax),
	})

	require.NoError(t, err)
	assertExample(t, rec, "incomesAndEngineOff")
}

// Engine ON: the hero passes through as the use case gave it, the legend
// next to it, each trend point with its tax.
func TestGetDashboard_emptyTrialExample(t *testing.T) {
	example := contractExample(t, "emptyTrial")
	ends := time.Date(2026, time.January, 20, 16, 0, 0, 0, time.UTC)
	days, incomes, expenses := 3, 2, 5
	legend := "Esto es un aproximado. No todos los gastos aplican para tu giro comercial."
	estimate := example["estimate"]
	out := usecase.Dashboard[hero]{
		TaxpayerID: exampleTax, Period: month(t, "2026-01"), IsEmpty: true,
		Current:  usecase.Figures{Income: cents(0), Expense: cents(0)},
		Estimate: &estimate,
		Trend: []usecase.TrendPoint{{
			Month: month(t, "2026-01"), Figures: usecase.Figures{Income: cents(0), Expense: cents(0)},
			Tax: &usecase.TrendTax{IVAOutcome: usecase.OutcomeZero, Status: domain.StatusComplete},
		}},
		Access: access.View{
			Level: access.LevelTrial, TrialEndsAt: &ends, TrialDaysLeft: &days, IncomesRemaining: &incomes, ExpensesRemaining: &expenses,
		},
		Legend: legend,
	}

	rec, err := serve(t, dashboardhttp.NewHandler[hero](&fakeDashboards{out: out}), gen.GetDashboardRequestObject{
		TaxpayerId: uuid.MustParse(exampleTax),
	})

	require.NoError(t, err)
	assertExample(t, rec, "emptyTrial")
}

func TestGetDashboard_trendTaxFigures(t *testing.T) {
	out := readOnlyEngineOff(t)
	out.Trend[1].Tax = &usecase.TrendTax{
		WithholdingsToRemit: 12_500, TotalPayable: 12_500, IVAOutcome: usecase.OutcomeInFavor, Status: domain.StatusIncomplete,
	}

	rec, err := serve(t, dashboardhttp.NewHandler[hero](&fakeDashboards{out: out}), gen.GetDashboardRequestObject{})

	require.NoError(t, err)
	var body struct {
		Trend []struct {
			Tax json.RawMessage `json:"tax"`
		} `json:"trend"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.JSONEq(t, `null`, string(body.Trend[0].Tax))
	assert.JSONEq(t, `{"iva_payable_rounded_cents":0,"isr_payable_rounded_cents":0,"withholdings_to_remit_cents":12500,`+
		`"total_payable_rounded_cents":12500,"iva_outcome":"in_favor","status":"incomplete"}`, string(body.Trend[1].Tax))
}

func TestGetDashboard_passesUseCaseErrors(t *testing.T) {
	boom := apperr.New(apperr.CodeNotFound, "Not found.")
	_, err := serve(t, dashboardhttp.NewHandler[hero](&fakeDashboards{err: boom}), gen.GetDashboardRequestObject{})
	require.ErrorIs(t, err, boom)
}

func TestGetDashboard_withoutPrincipalIs401(t *testing.T) {
	uc := &fakeDashboards{}
	resp, err := dashboardhttp.NewHandler[hero](uc).GetDashboard(t.Context(), gen.GetDashboardRequestObject{})
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetDashboardResponse(rec))
	assert.Equal(t, 401, rec.Code)
	assert.Equal(t, `Bearer realm="control-contable"`, rec.Header().Get("WWW-Authenticate"))
	assert.Empty(t, uc.calls)
}

// OQ-D5: own budget dashboard_reads, 600 per hour, separate from
// estimate_reads.
func TestUserLimits(t *testing.T) {
	assert.Equal(t, map[string]ratelimit.Policy{
		"getDashboard": ratelimit.PerWindow("dashboard_reads", 600, time.Hour),
	}, dashboardhttp.UserLimits)
	require.NoError(t, dashboardhttp.UserLimits["getDashboard"].Validate())
}
