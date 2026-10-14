package taxpayershttp_test

import (
	"context"
	"encoding/json"
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
	taxpayershttp "github.com/romodevx/control-contable/backend/internal/taxpayers/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const (
	callerID   = "0199d4a2-0000-7000-8000-00000000000a"
	taxpayerID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	noStore    = "no-store"
)

var created = time.Date(2026, time.October, 13, 16, 5, 0, 0, time.UTC)

// fakeTaxpayers records the inputs it receives and answers the canned
// result or error.
type fakeTaxpayers struct {
	user       string
	create     usecase.CreateInput
	createRes  usecase.CreateResult
	update     usecase.UpdateInput
	ids        []string
	archived   bool
	confirm    bool
	taxpayer   usecase.Taxpayer
	list       []usecase.Taxpayer
	err        error
	operations []string
}

func (f *fakeTaxpayers) record(ctx context.Context, op string) {
	f.user, _ = principal.UserID(ctx)
	f.operations = append(f.operations, op)
}

func (f *fakeTaxpayers) Create(ctx context.Context, userID string, in usecase.CreateInput) (usecase.CreateResult, error) {
	f.record(ctx, "create")
	f.user, f.create = userID, in
	return f.createRes, f.err
}

func (f *fakeTaxpayers) List(_ context.Context, userID string, includeArchived bool) ([]usecase.Taxpayer, error) {
	f.user, f.archived = userID, includeArchived
	return f.list, f.err
}

func (f *fakeTaxpayers) Get(_ context.Context, userID, id string) (usecase.Taxpayer, error) {
	f.user, f.ids = userID, append(f.ids, id)
	return f.taxpayer, f.err
}

func (f *fakeTaxpayers) Update(_ context.Context, userID, id string, in usecase.UpdateInput) (usecase.Taxpayer, error) {
	f.user, f.ids, f.update = userID, append(f.ids, id), in
	return f.taxpayer, f.err
}

func (f *fakeTaxpayers) Archive(_ context.Context, userID, id string) (usecase.Taxpayer, error) {
	f.user, f.ids = userID, append(f.ids, id)
	f.operations = append(f.operations, "archive")
	return f.taxpayer, f.err
}

func (f *fakeTaxpayers) Unarchive(_ context.Context, userID, id string) (usecase.Taxpayer, error) {
	f.user, f.ids = userID, append(f.ids, id)
	f.operations = append(f.operations, "unarchive")
	return f.taxpayer, f.err
}

func (f *fakeTaxpayers) Delete(_ context.Context, userID, id string, confirm bool) error {
	f.user, f.ids, f.confirm = userID, append(f.ids, id), confirm
	return f.err
}

type fakeRegimes struct {
	user    string
	id      string
	add     usecase.RegimeInput
	code    string
	patch   map[string]domain.OptionValue
	end     usecase.EndRegimeInput
	regime  usecase.TaxpayerRegime
	err     error
	catalog []usecase.CatalogEntry
	pt      string
}

func (f *fakeRegimes) Add(_ context.Context, userID, id string, in usecase.RegimeInput) (usecase.TaxpayerRegime, error) {
	f.user, f.id, f.add = userID, id, in
	return f.regime, f.err
}

func (f *fakeRegimes) UpdateOptions(_ context.Context, userID, id, code string, patch map[string]domain.OptionValue) (usecase.TaxpayerRegime, error) {
	f.user, f.id, f.code, f.patch = userID, id, code, patch
	return f.regime, f.err
}

func (f *fakeRegimes) End(_ context.Context, userID, id string, in usecase.EndRegimeInput) (usecase.TaxpayerRegime, error) {
	f.user, f.id, f.end = userID, id, in
	return f.regime, f.err
}

func (f *fakeRegimes) List(_ context.Context, userID, personType string) ([]usecase.CatalogEntry, error) {
	f.user, f.pt = userID, personType
	return f.catalog, f.err
}

func newHandler() (*taxpayershttp.Handler, *fakeTaxpayers, *fakeRegimes) {
	tp, rg := &fakeTaxpayers{}, &fakeRegimes{}
	return taxpayershttp.NewHandler(taxpayershttp.UseCases{Taxpayers: tp, Regimes: rg, Catalog: rg}), tp, rg
}

func signedIn(t *testing.T) context.Context {
	t.Helper()
	return principal.With(t.Context(), principal.Principal{UserID: callerID})
}

func mustPeriod(t *testing.T, raw string) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod(raw)
	require.NoError(t, err)
	return p
}

func sampleRegime(t *testing.T) usecase.TaxpayerRegime {
	t.Helper()
	return usecase.TaxpayerRegime{
		Code: "625", Name: "Plataformas", Treatment: "full",
		Options: map[string]domain.OptionValue{
			"activities":        domain.ListOption([]string{"goods_services"}),
			"definitive_option": domain.BoolOption(false),
			"mode":              domain.StringOption("blind_35"),
		},
		Window: domain.Window{Start: mustPeriod(t, "2026-01")},
	}
}

func sampleTaxpayer(t *testing.T) usecase.Taxpayer {
	t.Helper()
	ended := sampleRegime(t)
	ended.Code, ended.Window.End = "606", mustPeriod(t, "2026-06")
	return usecase.Taxpayer{
		ID: taxpayerID, PersonType: domain.PersonTypeFisica, DisplayName: "Ana (personal)",
		RFCMasked: "PEXJ******HK8", Regimes: []usecase.TaxpayerRegime{ended, sampleRegime(t)}, CreatedAt: created,
	}
}

// encode renders a response object the way the strict server would.
func encode(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func options(t *testing.T, raw string) *gen.RegimeOptions {
	t.Helper()
	var o gen.RegimeOptions
	require.NoError(t, json.Unmarshal([]byte(raw), &o))
	return &o
}

func TestCreateTaxpayer_mapsTheBodyAndAnswers201WithWarnings(t *testing.T) {
	h, tp, _ := newHandler()
	result := sampleTaxpayer(t)
	result.Warnings = []usecase.Warning{{Code: usecase.WarningRFCCheckDigitMismatch, Field: "rfc", Message: "Revisa tu RFC."}}
	tp.createRes = usecase.CreateResult{Taxpayer: result, Created: true}
	clientID := uuid.MustParse(taxpayerID)
	start := "2026-01"

	resp, err := h.CreateTaxpayer(signedIn(t), gen.CreateTaxpayerRequestObject{Body: &gen.TaxpayerCreate{
		Id: &clientID, PersonType: "fisica", Rfc: "pexj-900215 hk8", DisplayName: "Ana (personal)",
		Regimes: []gen.TaxpayerRegimeCreate{
			{RegimeCode: "625", Options: options(t, `{"activities":["goods_services"],"definitive_option":false}`), StartPeriod: &start},
			{RegimeCode: "612"},
		},
	}})

	require.NoError(t, err)
	got, ok := resp.(gen.CreateTaxpayer201JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, noStore, got.Headers.CacheControl)
	assert.Equal(t, callerID, tp.user, "the user comes from the principal")
	assert.Equal(t, usecase.CreateInput{
		ID: taxpayerID, PersonType: "fisica", RFC: "pexj-900215 hk8", DisplayName: "Ana (personal)",
		Regimes: []usecase.RegimeInput{
			{Code: "625", Start: "2026-01", Options: map[string]domain.OptionValue{
				"activities": domain.ListOption([]string{"goods_services"}), "definitive_option": domain.BoolOption(false),
			}},
			{Code: "612"},
		},
	}, tp.create)
	body := encode(t, got.Body)
	assert.Equal(t, []any{map[string]any{"code": "rfc_check_digit_mismatch", "field": "rfc", "message": "Revisa tu RFC."}}, body["warnings"])
	assert.Nil(t, body["archived_at"])
	assert.Contains(t, body, "archived_at", "archived_at is required, null while active")
	assert.Equal(t, "2026-10-13T16:05:00Z", body["created_at"])
	regimes := body["regimes"].([]any)
	require.Len(t, regimes, 2)
	assert.Equal(t, "2026-06", regimes[0].(map[string]any)["end_period"])
	open := regimes[1].(map[string]any)
	assert.Contains(t, open, "end_period")
	assert.Nil(t, open["end_period"])
	assert.NotContains(t, open, "warnings", "never inside Taxpayer.regimes")
	assert.Equal(t, map[string]any{"activities": []any{"goods_services"}, "definitive_option": false, "mode": "blind_35"}, open["options"])
}

func TestCreateTaxpayer_idempotentRetryAnswers200WithoutWarnings(t *testing.T) {
	h, tp, _ := newHandler()
	tp.createRes = usecase.CreateResult{Taxpayer: sampleTaxpayer(t)}

	resp, err := h.CreateTaxpayer(signedIn(t), gen.CreateTaxpayerRequestObject{Body: &gen.TaxpayerCreate{PersonType: "moral"}})

	require.NoError(t, err)
	got, ok := resp.(gen.CreateTaxpayer200JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, noStore, got.Headers.CacheControl)
	assert.NotContains(t, encode(t, got.Body), "warnings", "present only when non-empty")
	assert.Empty(t, tp.create.ID, "no client id: the server generates one")
}

// The router authenticates every secured operation first; a request that
// still reaches a handler without a principal gets the contract's 401 with
// its challenge, never a unit of work with an empty user.
func TestHandler_withoutAPrincipalAnswers401WithTheChallenge(t *testing.T) {
	h, tp, rg := newHandler()
	ctx := t.Context()
	calls := []func() (any, error){
		func() (any, error) { return h.ListRegimes(ctx, gen.ListRegimesRequestObject{}) },
		func() (any, error) { return h.ListTaxpayers(ctx, gen.ListTaxpayersRequestObject{}) },
		func() (any, error) {
			return h.CreateTaxpayer(ctx, gen.CreateTaxpayerRequestObject{Body: &gen.TaxpayerCreate{}})
		},
		func() (any, error) { return h.GetTaxpayer(ctx, gen.GetTaxpayerRequestObject{}) },
		func() (any, error) {
			return h.UpdateTaxpayer(ctx, gen.UpdateTaxpayerRequestObject{Body: &gen.TaxpayerUpdate{}})
		},
		func() (any, error) { return h.DeleteTaxpayer(ctx, gen.DeleteTaxpayerRequestObject{}) },
		func() (any, error) { return h.ArchiveTaxpayer(ctx, gen.ArchiveTaxpayerRequestObject{}) },
		func() (any, error) { return h.UnarchiveTaxpayer(ctx, gen.UnarchiveTaxpayerRequestObject{}) },
		func() (any, error) {
			return h.AddTaxpayerRegime(ctx, gen.AddTaxpayerRegimeRequestObject{Body: &gen.TaxpayerRegimeCreate{}})
		},
		func() (any, error) {
			return h.UpdateTaxpayerRegime(ctx, gen.UpdateTaxpayerRegimeRequestObject{Body: &gen.TaxpayerRegimeUpdate{}})
		},
		func() (any, error) { return h.ArchiveTaxpayerRegime(ctx, gen.ArchiveTaxpayerRegimeRequestObject{}) },
	}
	for i, call := range calls {
		resp, err := call()

		require.NoError(t, err, i)
		rec := httptest.NewRecorder()
		writeResponse(t, resp, rec)
		assert.Equal(t, 401, rec.Code, "operation %d", i)
		assert.Equal(t, `Bearer realm="control-contable"`, rec.Header().Get("WWW-Authenticate"), "operation %d", i)
	}
	assert.Empty(t, tp.user)
	assert.Empty(t, tp.operations)
	assert.Empty(t, rg.user)
}

// writeResponse runs the response object's generated Visit method.
func writeResponse(t *testing.T, resp any, rec *httptest.ResponseRecorder) {
	t.Helper()
	switch r := resp.(type) {
	case gen.ListRegimesResponseObject:
		require.NoError(t, r.VisitListRegimesResponse(rec))
	case gen.ListTaxpayersResponseObject:
		require.NoError(t, r.VisitListTaxpayersResponse(rec))
	case gen.CreateTaxpayerResponseObject:
		require.NoError(t, r.VisitCreateTaxpayerResponse(rec))
	case gen.GetTaxpayerResponseObject:
		require.NoError(t, r.VisitGetTaxpayerResponse(rec))
	case gen.UpdateTaxpayerResponseObject:
		require.NoError(t, r.VisitUpdateTaxpayerResponse(rec))
	case gen.DeleteTaxpayerResponseObject:
		require.NoError(t, r.VisitDeleteTaxpayerResponse(rec))
	case gen.ArchiveTaxpayerResponseObject:
		require.NoError(t, r.VisitArchiveTaxpayerResponse(rec))
	case gen.UnarchiveTaxpayerResponseObject:
		require.NoError(t, r.VisitUnarchiveTaxpayerResponse(rec))
	case gen.AddTaxpayerRegimeResponseObject:
		require.NoError(t, r.VisitAddTaxpayerRegimeResponse(rec))
	case gen.UpdateTaxpayerRegimeResponseObject:
		require.NoError(t, r.VisitUpdateTaxpayerRegimeResponse(rec))
	case gen.ArchiveTaxpayerRegimeResponseObject:
		require.NoError(t, r.VisitArchiveTaxpayerRegimeResponse(rec))
	default:
		t.Fatalf("unexpected response %T", resp)
	}
}

// Use case errors are client-ready *apperr.Error values: the handler
// returns them as is and the router writes the envelope and the status.
func TestHandler_passesUseCaseErrorsThrough(t *testing.T) {
	h, tp, rg := newHandler()
	duplicate := apperr.Wrap(usecase.ErrRFCAlreadyRegistered, apperr.CodeRFCAlreadyRegistered, "This RFC is already registered in your account.")
	tp.err, rg.err = duplicate, duplicate

	_, err := h.CreateTaxpayer(signedIn(t), gen.CreateTaxpayerRequestObject{Body: &gen.TaxpayerCreate{}})
	assert.Same(t, duplicate, err)
	_, err = h.GetTaxpayer(signedIn(t), gen.GetTaxpayerRequestObject{TaxpayerId: uuid.MustParse(taxpayerID)})
	assert.Same(t, duplicate, err)
	_, err = h.AddTaxpayerRegime(signedIn(t), gen.AddTaxpayerRegimeRequestObject{Body: &gen.TaxpayerRegimeCreate{RegimeCode: "612"}})
	assert.Same(t, duplicate, err)

	infra := errors.New("connection reset")
	tp.err = infra
	_, err = h.ListTaxpayers(signedIn(t), gen.ListTaxpayersRequestObject{})
	assert.ErrorIs(t, err, infra, "an infrastructure failure is the router's generic 500")
}

func TestListTaxpayers_includeArchivedDefaultsToFalse(t *testing.T) {
	h, tp, _ := newHandler()
	tp.list = []usecase.Taxpayer{sampleTaxpayer(t)}

	resp, err := h.ListTaxpayers(signedIn(t), gen.ListTaxpayersRequestObject{})
	require.NoError(t, err)
	got := resp.(gen.ListTaxpayers200JSONResponse)
	assert.False(t, tp.archived)
	assert.Equal(t, noStore, got.Headers.CacheControl)
	require.Len(t, got.Body.Items, 1)
	assert.Nil(t, got.Body.Items[0].Warnings, "never on reads")

	include := true
	_, err = h.ListTaxpayers(signedIn(t), gen.ListTaxpayersRequestObject{Params: gen.ListTaxpayersParams{IncludeArchived: &include}})
	require.NoError(t, err)
	assert.True(t, tp.archived)
}

func TestListTaxpayers_emptyIsAnEmptyArray(t *testing.T) {
	h, _, _ := newHandler()

	resp, err := h.ListTaxpayers(signedIn(t), gen.ListTaxpayersRequestObject{})

	require.NoError(t, err)
	assert.Equal(t, map[string]any{"items": []any{}}, encode(t, resp.(gen.ListTaxpayers200JSONResponse).Body))
}

func TestArchiveAndUnarchive(t *testing.T) {
	h, tp, _ := newHandler()
	archived := sampleTaxpayer(t)
	archived.ArchivedAt = created
	tp.taxpayer = archived
	id := uuid.MustParse(taxpayerID)

	resp, err := h.ArchiveTaxpayer(signedIn(t), gen.ArchiveTaxpayerRequestObject{TaxpayerId: id})
	require.NoError(t, err)
	got := resp.(gen.ArchiveTaxpayer200JSONResponse)
	require.NotNil(t, got.Body.ArchivedAt)
	assert.Equal(t, created, *got.Body.ArchivedAt)
	assert.Equal(t, noStore, got.Headers.CacheControl)

	_, err = h.UnarchiveTaxpayer(signedIn(t), gen.UnarchiveTaxpayerRequestObject{TaxpayerId: id})
	require.NoError(t, err)
	assert.Equal(t, []string{"archive", "unarchive"}, tp.operations)
	assert.Equal(t, []string{taxpayerID, taxpayerID}, tp.ids, "the canonical id string")
}

func TestUpdateTaxpayer_passesOnlyTheFieldsSent(t *testing.T) {
	h, tp, _ := newHandler()
	result := sampleTaxpayer(t)
	result.Warnings = []usecase.Warning{{Code: usecase.WarningRFCCheckDigitMismatch, Field: "rfc", Message: "m"}}
	tp.taxpayer = result
	name := "Ana (negocio)"

	resp, err := h.UpdateTaxpayer(signedIn(t), gen.UpdateTaxpayerRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), Body: &gen.TaxpayerUpdate{DisplayName: &name},
	})

	require.NoError(t, err)
	got := resp.(gen.UpdateTaxpayer200JSONResponse)
	assert.Equal(t, usecase.UpdateInput{DisplayName: &name}, tp.update)
	require.NotNil(t, got.Body.Warnings)
	assert.Len(t, *got.Body.Warnings, 1)
}

func TestDeleteTaxpayer_confirmDefaultsToFalseAndAnswers204(t *testing.T) {
	h, tp, _ := newHandler()
	id := uuid.MustParse(taxpayerID)

	_, err := h.DeleteTaxpayer(signedIn(t), gen.DeleteTaxpayerRequestObject{TaxpayerId: id})
	require.NoError(t, err)
	assert.False(t, tp.confirm, "no confirm: the use case answers 400 confirm_required")

	confirm := true
	resp, err := h.DeleteTaxpayer(signedIn(t), gen.DeleteTaxpayerRequestObject{TaxpayerId: id, Params: gen.DeleteTaxpayerParams{Confirm: &confirm}})
	require.NoError(t, err)
	assert.IsType(t, gen.DeleteTaxpayer204Response{}, resp)
	assert.True(t, tp.confirm)
}

func TestAddTaxpayerRegime_answers201WithTheRegimeWarnings(t *testing.T) {
	h, _, rg := newHandler()
	regime := sampleRegime(t)
	regime.Warnings = []usecase.Warning{{Code: usecase.WarningRegimeNeedsCPAReview, Field: "regime_code", Message: "Pendiente."}}
	rg.regime = regime

	resp, err := h.AddTaxpayerRegime(signedIn(t), gen.AddTaxpayerRegimeRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID),
		Body:       &gen.TaxpayerRegimeCreate{RegimeCode: "625", Options: options(t, `{"definitive_option":true}`)},
	})

	require.NoError(t, err)
	got := resp.(gen.AddTaxpayerRegime201JSONResponse)
	assert.Equal(t, noStore, got.Headers.CacheControl)
	assert.Equal(t, usecase.RegimeInput{Code: "625", Options: map[string]domain.OptionValue{"definitive_option": domain.BoolOption(true)}}, rg.add)
	assert.Equal(t, taxpayerID, rg.id)
	body := encode(t, got.Body)
	assert.Equal(t, []any{map[string]any{"code": "regime_combination_needs_cpa_review", "field": "regime_code", "message": "Pendiente."}}, body["warnings"])
	assert.Equal(t, "2026-01", body["start_period"])
}

func TestUpdateTaxpayerRegime_decodesEveryOptionKind(t *testing.T) {
	h, _, rg := newHandler()
	rg.regime = sampleRegime(t)

	resp, err := h.UpdateTaxpayerRegime(signedIn(t), gen.UpdateTaxpayerRegimeRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "625",
		Body: &gen.TaxpayerRegimeUpdate{Options: *options(t, `{"a":"x","b":true,"c":["p","q"],"d":[]}`)},
	})

	require.NoError(t, err)
	assert.IsType(t, gen.UpdateTaxpayerRegime200JSONResponse{}, resp)
	assert.Equal(t, "625", rg.code)
	assert.Equal(t, map[string]domain.OptionValue{
		"a": domain.StringOption("x"), "b": domain.BoolOption(true),
		"c": domain.ListOption([]string{"p", "q"}), "d": domain.ListOption([]string{}),
	}, rg.patch)
}

// Contract validation rejects any other JSON type first; the handler still
// refuses one (a number, null) as a field error, never a 500.
func TestUpdateTaxpayerRegime_refusesAnOptionTheContractCannotCarry(t *testing.T) {
	h, _, rg := newHandler()

	_, err := h.UpdateTaxpayerRegime(signedIn(t), gen.UpdateTaxpayerRegimeRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "625",
		Body: &gen.TaxpayerRegimeUpdate{Options: *options(t, `{"rate":0.16}`)},
	})

	app, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeValidationFailed, app.Code)
	require.Len(t, app.Fields, 1)
	assert.Equal(t, apperr.FieldError{Field: "options.rate", Code: usecase.DetailInvalidOption, Message: app.Fields[0].Message}, app.Fields[0])
	assert.Empty(t, rg.user, "the use case is not reached")
}

func TestArchiveTaxpayerRegime_mapsTheQuery(t *testing.T) {
	h, _, rg := newHandler()
	regime := sampleRegime(t)
	regime.Window.End = mustPeriod(t, "2026-09")
	rg.regime = regime
	confirm := true

	resp, err := h.ArchiveTaxpayerRegime(signedIn(t), gen.ArchiveTaxpayerRegimeRequestObject{
		TaxpayerId: uuid.MustParse(taxpayerID), RegimeCode: "625",
		Params: gen.ArchiveTaxpayerRegimeParams{EndPeriod: "2026-09", Confirm: &confirm},
	})

	require.NoError(t, err)
	got := resp.(gen.ArchiveTaxpayerRegime200JSONResponse)
	assert.Equal(t, usecase.EndRegimeInput{Code: "625", End: "2026-09", Confirm: true}, rg.end)
	require.NotNil(t, got.Body.EndPeriod)
	assert.Equal(t, "2026-09", *got.Body.EndPeriod)
}

func TestListRegimes_mapsTheCatalog(t *testing.T) {
	h, _, rg := newHandler()
	rg.catalog = []usecase.CatalogEntry{
		{
			ListedRegime: usecase.ListedRegime{
				Code: "606", PersonType: domain.PersonTypeFisica, Name: "Arrendamiento", Treatment: "full",
				OptionDefs: []usecase.OptionDefView{
					{Key: "deduction_mode", Label: "Tipo de deducción", Type: domain.OptionTypeEnum,
						AllowedValues: []string{"blind_35", "actual"}, Default: domain.StringOption("blind_35"), Required: true},
					{Key: "vat_rate", Label: "Tasa", Type: domain.OptionTypeRate, Min: "0", Max: "0.16", Required: false},
					{Key: "flag", Label: "Bandera", Type: domain.OptionTypeBoolean, Default: domain.BoolOption(true), Required: true},
				},
				RequiredInputs: []usecase.TaxInputDefView{{Key: "coeficiente_utilidad", Label: "Coeficiente", Kind: "rate", Required: true}},
			},
			IncompatibleWith:  []usecase.RegimeLink{{Code: "626", Reason: "RESICO."}},
			ReviewPendingWith: []usecase.RegimeLink{},
		},
		{
			ListedRegime: usecase.ListedRegime{
				Code: "605", PersonType: domain.PersonTypeFisica, Name: "Sueldos", Treatment: "informational",
				Explanation: "Tu patrón retiene.", OptionDefs: []usecase.OptionDefView{}, RequiredInputs: []usecase.TaxInputDefView{},
			},
			IncompatibleWith:  []usecase.RegimeLink{},
			ReviewPendingWith: []usecase.RegimeLink{{Code: "625", Reason: "Pendiente."}},
		},
	}

	resp, err := h.ListRegimes(signedIn(t), gen.ListRegimesRequestObject{Params: gen.ListRegimesParams{PersonType: "fisica"}})

	require.NoError(t, err)
	assert.Equal(t, "fisica", rg.pt)
	assert.Equal(t, callerID, rg.user)
	body := encode(t, resp.(gen.ListRegimes200JSONResponse))
	items := body["items"].([]any)
	require.Len(t, items, 2)
	arrendamiento := items[0].(map[string]any)
	assert.Nil(t, arrendamiento["explanation"])
	assert.Contains(t, arrendamiento, "explanation", "required, null for a full regime")
	assert.Equal(t, []any{map[string]any{"code": "626", "reason": "RESICO."}}, arrendamiento["incompatible_with"])
	assert.Equal(t, []any{}, arrendamiento["review_pending_with"], "always sent")
	defs := arrendamiento["option_defs"].([]any)
	assert.Equal(t, map[string]any{
		"key": "deduction_mode", "label": "Tipo de deducción", "type": "enum",
		"allowed_values": []any{"blind_35", "actual"}, "default": "blind_35", "required": true,
	}, defs[0])
	assert.Equal(t, map[string]any{
		"key": "vat_rate", "label": "Tasa", "type": "rate", "min": "0", "max": "0.16", "default": nil, "required": false,
	}, defs[1], "rate bounds as decimal strings; no default is null")
	assert.Equal(t, true, defs[2].(map[string]any)["default"])
	assert.Equal(t, []any{map[string]any{"key": "coeficiente_utilidad", "label": "Coeficiente", "kind": "rate", "required": true}},
		arrendamiento["required_inputs"])
	sueldos := items[1].(map[string]any)
	assert.Equal(t, "Tu patrón retiene.", sueldos["explanation"])
	assert.Equal(t, []any{map[string]any{"code": "625", "reason": "Pendiente."}}, sueldos["review_pending_with"])
}
