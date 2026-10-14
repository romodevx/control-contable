package taxenginehttp_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

const (
	callerID   = "0199d4a2-0000-7000-8000-00000000000a"
	taxpayerID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
)

// fakeInputs records what it receives and answers the canned view or error.
type fakeInputs struct {
	user   string
	req    usecase.InputsRequest
	values []domain.SubmittedInput
	view   usecase.TaxInputsView
	err    error
	calls  []string
}

func (f *fakeInputs) Get(_ context.Context, userID string, req usecase.InputsRequest) (usecase.TaxInputsView, error) {
	f.user, f.req, f.calls = userID, req, append(f.calls, "get")
	return f.view, f.err
}

func (f *fakeInputs) Put(_ context.Context, userID string, req usecase.InputsRequest, values []domain.SubmittedInput) (usecase.TaxInputsView, error) {
	f.user, f.req, f.values, f.calls = userID, req, values, append(f.calls, "put")
	return f.view, f.err
}

func authed(t *testing.T) context.Context {
	t.Helper()
	return principal.With(t.Context(), principal.Principal{UserID: callerID})
}

func ptr[T any](v T) *T { return &v }

// view601 is the contract example regime601 as the use case returns it.
var view601 = usecase.TaxInputsView{
	Regime: "601", FiscalYear: 2026,
	Definitions: []domain.InputDef{
		{Key: domain.InputCoeficienteUtilidad, Label: "Coeficiente de utilidad", Kind: domain.InputRate, Required: true},
		{Key: domain.InputFirstFiscalYear, Label: "Primer ejercicio fiscal", Kind: domain.InputBoolean},
		{Key: domain.InputPriorProvisionalPayments, Label: "Pagos provisionales anteriores del ejercicio", Kind: domain.InputMoneyMonthly, Required: true},
	},
	Values: []usecase.StoredInput{
		{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: "0.1500"},
		{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: false},
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Cents: 200000},
	},
}

// wantBody601 sends every value field, null when not set (contract
// TaxInputValue: "Responses send every field").
const wantBody601 = `{
  "regime_code": "601",
  "fiscal_year": 2026,
  "definitions": [
    {"key": "coeficiente_utilidad", "label": "Coeficiente de utilidad", "kind": "rate", "required": true},
    {"key": "first_fiscal_year", "label": "Primer ejercicio fiscal", "kind": "boolean", "required": false},
    {"key": "prior_provisional_payments", "label": "Pagos provisionales anteriores del ejercicio", "kind": "money_monthly", "required": true}
  ],
  "values": [
    {"key": "coeficiente_utilidad", "month": null, "value_cents": null, "value_rate": "0.1500", "value_bool": null},
    {"key": "first_fiscal_year", "month": null, "value_cents": null, "value_rate": null, "value_bool": false},
    {"key": "prior_provisional_payments", "month": 1, "value_cents": 200000, "value_rate": null, "value_bool": null}
  ]
}`

func TestGetTaxInputs(t *testing.T) {
	uc := &fakeInputs{view: view601}
	h := taxenginehttp.NewHandler(uc, &fakeEstimates{})

	resp, err := h.GetTaxInputs(authed(t), gen.GetTaxInputsRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "601", Params: gen.GetTaxInputsParams{FiscalYear: 2026},
	})

	require.NoError(t, err)
	assert.Equal(t, callerID, uc.user)
	assert.Equal(t, usecase.InputsRequest{TaxpayerID: taxpayerID, RegimeCode: "601", FiscalYear: 2026}, uc.req)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxInputsResponse(rec))
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, wantBody601, rec.Body.String())
}

func TestGetTaxInputs_emptyValuesAreAnArray(t *testing.T) {
	uc := &fakeInputs{view: usecase.TaxInputsView{Regime: "612", FiscalYear: 2026, Definitions: []domain.InputDef{}, Values: []usecase.StoredInput{}}}
	resp, err := taxenginehttp.NewHandler(uc, &fakeEstimates{}).GetTaxInputs(authed(t), gen.GetTaxInputsRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "612", Params: gen.GetTaxInputsParams{FiscalYear: 2026},
	})
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitGetTaxInputsResponse(rec))
	assert.JSONEq(t, `{"regime_code":"612","fiscal_year":2026,"definitions":[],"values":[]}`, rec.Body.String())
}

func TestPutTaxInputs(t *testing.T) {
	uc := &fakeInputs{view: view601}
	h := taxenginehttp.NewHandler(uc, &fakeEstimates{})

	resp, err := h.PutTaxInputs(authed(t), gen.PutTaxInputsRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "601", Params: gen.PutTaxInputsParams{FiscalYear: 2026},
		Body: &gen.TaxInputsUpdate{Values: []gen.TaxInputValue{
			{Key: "coeficiente_utilidad", ValueRate: ptr("0.1500")},
			{Key: "first_fiscal_year", ValueBool: ptr(false)},
			{Key: "prior_provisional_payments", Month: ptr(1), ValueCents: ptr[int64](200000)},
		}},
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"put"}, uc.calls)
	assert.Equal(t, usecase.InputsRequest{TaxpayerID: taxpayerID, RegimeCode: "601", FiscalYear: 2026}, uc.req)
	assert.Equal(t, []domain.SubmittedInput{
		{Key: domain.InputCoeficienteUtilidad, Rate: ptr("0.1500")},
		{Key: domain.InputFirstFiscalYear, Bool: ptr(false)},
		{Key: domain.InputPriorProvisionalPayments, Month: ptr(1), Cents: ptr[int64](200000)},
	}, uc.values)
	rec := httptest.NewRecorder()
	require.NoError(t, resp.VisitPutTaxInputsResponse(rec))
	assert.Equal(t, 200, rec.Code)
	assert.JSONEq(t, wantBody601, rec.Body.String())
}

func TestPutTaxInputs_withoutBodyIsAValidationError(t *testing.T) {
	uc := &fakeInputs{}

	_, err := taxenginehttp.NewHandler(uc, &fakeEstimates{}).PutTaxInputs(authed(t), gen.PutTaxInputsRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "601", Params: gen.PutTaxInputsParams{FiscalYear: 2026},
	})

	e, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok)
	assert.Equal(t, apperr.CodeValidationFailed, e.Code)
	assert.Empty(t, uc.calls)
}

func TestTaxInputs_useCaseErrorsPassThrough(t *testing.T) {
	refused := apperr.New(apperr.CodeEntitlementRequired, "Read-only.")
	uc := &fakeInputs{err: refused}
	h := taxenginehttp.NewHandler(uc, &fakeEstimates{})

	_, getErr := h.GetTaxInputs(authed(t), gen.GetTaxInputsRequestObject{TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "601"})
	_, putErr := h.PutTaxInputs(authed(t), gen.PutTaxInputsRequestObject{TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "601",
		Body: &gen.TaxInputsUpdate{Values: []gen.TaxInputValue{}}})

	assert.Same(t, refused, getErr)
	assert.Same(t, refused, putErr)
}

func TestTaxInputs_withoutPrincipalIs401(t *testing.T) {
	uc := &fakeInputs{}
	h := taxenginehttp.NewHandler(uc, &fakeEstimates{})

	get, err := h.GetTaxInputs(t.Context(), gen.GetTaxInputsRequestObject{})
	require.NoError(t, err)
	put, err := h.PutTaxInputs(t.Context(), gen.PutTaxInputsRequestObject{})
	require.NoError(t, err)

	for _, visit := range []func(*httptest.ResponseRecorder) error{
		func(r *httptest.ResponseRecorder) error { return get.VisitGetTaxInputsResponse(r) },
		func(r *httptest.ResponseRecorder) error { return put.VisitPutTaxInputsResponse(r) },
	} {
		rec := httptest.NewRecorder()
		require.NoError(t, visit(rec))
		assert.Equal(t, 401, rec.Code)
		assert.Equal(t, `Bearer realm="control-contable"`, rec.Header().Get("WWW-Authenticate"))
	}
	assert.Empty(t, uc.calls)
}

func TestUserLimits(t *testing.T) {
	policy, ok := taxenginehttp.UserLimits["putTaxInputs"]

	require.True(t, ok, "ADR-0005: per-user write budget on putTaxInputs")
	assert.Equal(t, ratelimit.PerWindow("tax_input_writes", 60, time.Hour), policy)
	require.NoError(t, policy.Validate())
	assert.NotContains(t, taxenginehttp.UserLimits, "getTaxInputs", "reads are left to the edge limiter")
}
