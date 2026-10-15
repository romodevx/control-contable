package incomeshttp_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	incomeshttp "github.com/romodevx/control-contable/backend/internal/incomes/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

const (
	callerID   = "0199d4a2-0000-7000-8000-00000000000a"
	taxpayerID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	incomeID   = "0199d4a3-1a2b-7d3c-8e4f-5a6b7c8d9e01"
	folio      = "6F3B2C1D-8A4E-4B7F-9C2D-1E5A7B9C3D4F"
)

var stamp = time.Date(2026, time.September, 10, 17, 0, 0, 0, time.UTC)

// fakeIncomes records what the handler passes and answers the canned
// values or err.
type fakeIncomes struct {
	userID, taxpayerID, incomeID string
	create                       usecase.CreateInput
	patch                        usecase.Patch
	list                         usecase.ListQuery
	defaults                     usecase.DefaultsQuery
	result                       usecase.WriteResult
	income                       usecase.Income
	page                         usecase.Page
	prefill                      usecase.Defaults
	previewQuery                 usecase.PreviewQuery
	preview                      usecase.Preview
	deleted                      bool
	err                          error
}

func (f *fakeIncomes) Create(_ context.Context, userID string, in usecase.CreateInput) (usecase.WriteResult, error) {
	f.userID, f.create = userID, in
	return f.result, f.err
}

func (f *fakeIncomes) Update(_ context.Context, userID, taxpayerID, incomeID string, p usecase.Patch) (usecase.WriteResult, error) {
	f.userID, f.taxpayerID, f.incomeID, f.patch = userID, taxpayerID, incomeID, p
	return f.result, f.err
}

func (f *fakeIncomes) Get(_ context.Context, userID, taxpayerID, incomeID string) (usecase.Income, error) {
	f.userID, f.taxpayerID, f.incomeID = userID, taxpayerID, incomeID
	return f.income, f.err
}

func (f *fakeIncomes) List(_ context.Context, userID string, q usecase.ListQuery) (usecase.Page, error) {
	f.userID, f.list = userID, q
	return f.page, f.err
}

func (f *fakeIncomes) Delete(_ context.Context, userID, taxpayerID, incomeID string) error {
	f.userID, f.taxpayerID, f.incomeID, f.deleted = userID, taxpayerID, incomeID, true
	return f.err
}

func (f *fakeIncomes) Restore(_ context.Context, userID, taxpayerID, incomeID string) (usecase.WriteResult, error) {
	f.userID, f.taxpayerID, f.incomeID = userID, taxpayerID, incomeID
	return f.result, f.err
}

func (f *fakeIncomes) Defaults(_ context.Context, userID string, q usecase.DefaultsQuery) (usecase.Defaults, error) {
	f.userID, f.defaults = userID, q
	return f.prefill, f.err
}

func (f *fakeIncomes) Preview(_ context.Context, userID string, q usecase.PreviewQuery) (usecase.Preview, error) {
	f.userID, f.previewQuery = userID, q
	return f.preview, f.err
}

func authed(t *testing.T) context.Context {
	t.Helper()
	return principal.With(t.Context(), principal.Principal{UserID: callerID})
}

func uid(raw string) openapi_types.UUID { return uuid.MustParse(raw) }

func day(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

func genDate(year int, month time.Month, d int) openapi_types.Date {
	return openapi_types.Date{Time: time.Date(year, month, d, 0, 0, 0, 0, time.UTC)}
}

func rate(t *testing.T, raw string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(raw)
	require.NoError(t, err)
	return r
}

// sample is the contract's Income example: every optional field set.
func sample(t *testing.T) usecase.Income {
	t.Helper()
	folioValue, err := fiscal.ParseCFDIUUID(folio)
	require.NoError(t, err)
	return usecase.Income{
		ID: incomeID, TaxpayerID: taxpayerID,
		Fields: usecase.Fields{
			Regime: "612", Counterparty: fiscal.ReceptorPM, Operation: fiscal.OperationProfessionalServices,
			Amount: 10_000_00, IVARate: fiscal.IVARate16, IVA: 1_600_00,
			Dates:    domain.Dates{InvoicedOn: day(t, "2026-09-05"), CollectedOn: day(t, "2026-09-10")},
			CFDIUUID: folioValue, Description: "Diseño de logotipo",
		},
		Withholding: fiscal.WithholdingRecord{
			Withholdings: fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67}, ISRRate: rate(t, "0.10"),
			IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
			RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6}, TableVersion: "2026.1",
		},
		BookedOn: day(t, "2026-09-10"), CreatedAt: stamp, UpdatedAt: stamp.Add(time.Hour),
	}
}

func encode(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestCreateIncome_mapsTheRequestAndAnswers201(t *testing.T) {
	fake := &fakeIncomes{result: usecase.WriteResult{
		Income: sample(t), Created: true,
		Warnings: []domain.Warning{{Code: domain.WarningDuplicateCFDIUUID, Field: domain.FieldCFDIUUID}},
	}}
	h := incomeshttp.NewHandler(fake)
	id := uid(incomeID)
	kind, operation, via := gen.CounterpartyKind("pm"), gen.OperationKind("professional_services"), false
	folioLower, description := "6f3b2c1d-8a4e-4b7f-9c2d-1e5a7b9c3d4f", "  Diseño  "
	invoiced, collected := genDate(2026, time.September, 5), genDate(2026, time.September, 10)

	resp, err := h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{
		TaxpayerId: uid(taxpayerID),
		Body: &gen.IncomeCreate{
			Id: &id, RegimeCode: "612", AmountCents: 10_000_00,
			CounterpartyKind: kind, OperationKind: &operation, ViaPlatform: &via,
			IvaRate: "rate_16", IvaCents: 1_600_00,
			InvoicedOn: &invoiced, CollectedOn: &collected, CfdiUuid: &folioLower, Description: &description,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, callerID, fake.userID, "the user comes from the principal")
	assert.Equal(t, usecase.CreateInput{
		ID: incomeID, TaxpayerID: taxpayerID,
		Income: usecase.IncomeInput{
			Regime: "612", Counterparty: fiscal.ReceptorPM, Operation: fiscal.OperationProfessionalServices,
			ViaPlatform: &via, Amount: 10_000_00, IVARate: fiscal.IVARate16, IVA: 1_600_00,
			Dates:    domain.Dates{InvoicedOn: day(t, "2026-09-05"), CollectedOn: day(t, "2026-09-10")},
			CFDIUUID: folioLower, Description: description,
		},
	}, fake.create, "raw text: the use case normalizes the folio and the description")
	created, ok := resp.(gen.CreateIncome201JSONResponse)
	require.True(t, ok, "%T", resp)
	body := encode(t, created.Body)
	assert.Equal(t, map[string]any{
		"id": incomeID, "taxpayer_id": taxpayerID, "regime_code": "612",
		"counterparty_kind": "pm", "operation_kind": "professional_services", "residential_rent": false,
		"via_platform": false,
		"withholding": map[string]any{
			"isr_cents": float64(100_000), "iva_cents": float64(106_667), "isr_rate": "0.1",
			"iva_basis": "two_thirds_of_iva", "withholder": "counterparty", "net_cents": float64(953_333),
			"remit_due_on": nil, "rule_ids": []any{"ISR-5", "IVA-6"},
			"explanations": []any{
				map[string]any{"rule_id": "ISR-5", "text": "La empresa te retiene 10 % de ISR por honorarios (LISR art. 106)."},
				map[string]any{"rule_id": "IVA-6", "text": "Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3)."},
			},
			"table_version": "2026.1", "cpa_unverified": false,
		},
		"amount_cents": float64(1_000_000), "iva_rate": "rate_16", "iva_cents": float64(160_000),
		"isr_withheld_cents": float64(100_000), "iva_withheld_cents": float64(106_667),
		"invoiced_on": "2026-09-05", "collected_on": "2026-09-10", "pending_collection": false,
		"platform_activity": nil, "cfdi_uuid": folio, "description": "Diseño de logotipo",
		"created_at": "2026-09-10T17:00:00Z", "updated_at": "2026-09-10T18:00:00Z",
		"warnings": []any{map[string]any{
			"code": "duplicate_cfdi_uuid", "field": "cfdi_uuid", "message": "Ya tienes otro registro con este folio fiscal.",
		}},
	}, body)
}

func TestCreateIncome_idempotentRetryAnswers200AndOptionalsAbsent(t *testing.T) {
	stored := sample(t)
	stored.Dates.InvoicedOn, stored.CFDIUUID, stored.Description = fiscal.Date{}, fiscal.CFDIUUID{}, ""
	stored.Regime, stored.PlatformActivity = "625", domain.ActivityLodging
	fake := &fakeIncomes{result: usecase.WriteResult{Income: stored}}
	h := incomeshttp.NewHandler(fake)

	resp, err := h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{
		TaxpayerId: uid(taxpayerID),
		Body: &gen.IncomeCreate{
			RegimeCode: "625", CounterpartyKind: "pm", AmountCents: 100, IvaRate: "exempt",
		},
	})

	require.NoError(t, err)
	assert.Equal(t, usecase.CreateInput{TaxpayerID: taxpayerID, Income: usecase.IncomeInput{
		Regime: "625", Counterparty: fiscal.ReceptorPM, Amount: 100, IVARate: fiscal.IVAExempt,
	}}, fake.create, "no id: the server generates one; absent fields are zero")
	retried, ok := resp.(gen.CreateIncome200JSONResponse)
	require.True(t, ok, "%T", resp)
	body := encode(t, retried.Body)
	assert.Nil(t, body["invoiced_on"])
	assert.Nil(t, body["cfdi_uuid"])
	assert.Nil(t, body["description"])
	assert.Equal(t, "lodging", body["platform_activity"])
	assert.Equal(t, []any{}, body["warnings"], "warnings is [] when there are none")
}

func TestCreateIncome_pendingCollection(t *testing.T) {
	stored := sample(t)
	stored.Dates.CollectedOn, stored.PendingCollection = fiscal.Date{}, true
	h := incomeshttp.NewHandler(&fakeIncomes{result: usecase.WriteResult{Income: stored, Created: true}})

	resp, err := h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{TaxpayerId: uid(taxpayerID), Body: &gen.IncomeCreate{}})

	require.NoError(t, err)
	body := encode(t, resp.(gen.CreateIncome201JSONResponse).Body)
	assert.Equal(t, true, body["pending_collection"])
	assert.Nil(t, body["collected_on"])
}

func TestHandler_everyOperationRequiresTheCaller(t *testing.T) {
	fake := &fakeIncomes{}
	h := incomeshttp.NewHandler(fake)
	ctx := t.Context() // no principal
	tp, id := uid(taxpayerID), uid(incomeID)

	responses := []any{}
	r1, err := h.ListIncomes(ctx, gen.ListIncomesRequestObject{TaxpayerId: tp})
	require.NoError(t, err)
	responses = append(responses, r1)
	r2, err := h.CreateIncome(ctx, gen.CreateIncomeRequestObject{TaxpayerId: tp})
	require.NoError(t, err)
	responses = append(responses, r2)
	r3, err := h.GetIncomeDefaults(ctx, gen.GetIncomeDefaultsRequestObject{TaxpayerId: tp})
	require.NoError(t, err)
	responses = append(responses, r3)
	r4, err := h.GetIncome(ctx, gen.GetIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	require.NoError(t, err)
	responses = append(responses, r4)
	r5, err := h.UpdateIncome(ctx, gen.UpdateIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	require.NoError(t, err)
	responses = append(responses, r5)
	r6, err := h.DeleteIncome(ctx, gen.DeleteIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	require.NoError(t, err)
	responses = append(responses, r6)
	r7, err := h.RestoreIncome(ctx, gen.RestoreIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	require.NoError(t, err)
	responses = append(responses, r7)

	for _, resp := range responses {
		body := encode(t, resp)
		require.Contains(t, body, "Body", "%T", resp)
		assert.Equal(t, "unauthorized", body["Body"].(map[string]any)["error"].(map[string]any)["code"], "%T", resp)
		assert.Equal(t, `Bearer realm="control-contable"`, body["Headers"].(map[string]any)["WWWAuthenticate"], "%T", resp)
	}
	assert.Empty(t, fake.userID, "no use case ran")
}

func TestHandler_passesUseCaseErrorsThrough(t *testing.T) {
	refusal := apperr.New(apperr.CodeEntitlementRequired, "Cap.")
	fake := &fakeIncomes{err: refusal}
	h := incomeshttp.NewHandler(fake)
	ctx, tp, id := authed(t), uid(taxpayerID), uid(incomeID)

	_, err := h.ListIncomes(ctx, gen.ListIncomesRequestObject{TaxpayerId: tp})
	assert.ErrorIs(t, err, refusal)
	_, err = h.CreateIncome(ctx, gen.CreateIncomeRequestObject{TaxpayerId: tp, Body: &gen.IncomeCreate{}})
	assert.ErrorIs(t, err, refusal)
	_, err = h.GetIncomeDefaults(ctx, gen.GetIncomeDefaultsRequestObject{TaxpayerId: tp})
	assert.ErrorIs(t, err, refusal)
	_, err = h.GetIncome(ctx, gen.GetIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	assert.ErrorIs(t, err, refusal)
	_, err = h.UpdateIncome(ctx, gen.UpdateIncomeRequestObject{TaxpayerId: tp, IncomeId: id, Body: &gen.IncomeUpdate{}})
	assert.ErrorIs(t, err, refusal)
	_, err = h.DeleteIncome(ctx, gen.DeleteIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	assert.ErrorIs(t, err, refusal)
	_, err = h.RestoreIncome(ctx, gen.RestoreIncomeRequestObject{TaxpayerId: tp, IncomeId: id})
	assert.ErrorIs(t, err, refusal)
}

func TestUpdateIncome_mergePatch(t *testing.T) {
	fake := &fakeIncomes{result: usecase.WriteResult{Income: sample(t)}}
	h := incomeshttp.NewHandler(fake)
	regime, rateValue := "601", gen.IvaRate("rate_8")
	kind, operation, residential, via := gen.CounterpartyKind("gov_federal"), gen.OperationKind("scrap"), true, false
	amount, iva := int64(2), int64(3)

	resp, err := h.UpdateIncome(authed(t), gen.UpdateIncomeRequestObject{
		TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID),
		Body: &gen.IncomeUpdate{
			RegimeCode: &regime, AmountCents: &amount, IvaRate: &rateValue,
			CounterpartyKind: &kind, OperationKind: &operation, ResidentialRent: &residential, ViaPlatform: &via,
			IvaCents:         &iva,
			InvoicedOn:       nullable.NewNullableWithValue(genDate(2026, time.September, 1)),
			CollectedOn:      nullable.NewNullNullable[openapi_types.Date](),
			PlatformActivity: nullable.NewNullableWithValue(gen.PlatformActivity("lodging")),
			CfdiUuid:         nullable.NewNullNullable[string](),
			Description:      nullable.NewNullableWithValue("Nuevo"),
		},
	})

	require.NoError(t, err)
	assert.Equal(t, callerID, fake.userID)
	assert.Equal(t, taxpayerID, fake.taxpayerID)
	assert.Equal(t, incomeID, fake.incomeID)
	p := fake.patch
	assert.Equal(t, fiscal.RegimeCode("601"), *p.Regime)
	assert.Equal(t, fiscal.ReceptorGovFederal, *p.Counterparty)
	assert.Equal(t, fiscal.OperationScrap, *p.Operation)
	assert.True(t, *p.ResidentialRent)
	assert.False(t, *p.ViaPlatform)
	assert.Equal(t, money.Cents(2), *p.Amount)
	assert.Equal(t, fiscal.IVARate8, *p.IVARate)
	assert.Equal(t, money.Cents(3), *p.IVA)
	assert.Equal(t, day(t, "2026-09-01"), *p.InvoicedOn)
	assert.True(t, p.CollectedOn.IsZero(), "null clears the collection date")
	assert.Equal(t, domain.ActivityLodging, *p.PlatformActivity)
	assert.Equal(t, "", *p.CFDIUUID, "null clears the folio")
	assert.Equal(t, "Nuevo", *p.Description)
	_, ok := resp.(gen.UpdateIncome200JSONResponse)
	require.True(t, ok, "%T", resp)
}

// Absent keys stay unchanged; JSON null decodes to a clear.
func TestUpdateIncome_absentAndNullFromJSON(t *testing.T) {
	fake := &fakeIncomes{result: usecase.WriteResult{Income: sample(t)}}
	h := incomeshttp.NewHandler(fake)
	var body gen.IncomeUpdate
	require.NoError(t, json.Unmarshal([]byte(`{"platform_activity":null,"description":null,"invoiced_on":null}`), &body))

	_, err := h.UpdateIncome(authed(t), gen.UpdateIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID), Body: &body})

	require.NoError(t, err)
	assert.Equal(t, usecase.Patch{
		PlatformActivity: new(domain.PlatformActivity("")), Description: new(""), InvoicedOn: &fiscal.Date{},
	}, fake.patch)
}

func TestGetIncome_readsTheCallersIncome(t *testing.T) {
	fake := &fakeIncomes{income: sample(t)}
	h := incomeshttp.NewHandler(fake)

	resp, err := h.GetIncome(authed(t), gen.GetIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID)})

	require.NoError(t, err)
	assert.Equal(t, []string{callerID, taxpayerID, incomeID}, []string{fake.userID, fake.taxpayerID, fake.incomeID})
	got, ok := resp.(gen.GetIncome200JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, []any{}, encode(t, got.Body)["warnings"], "reads carry no warnings")
}

func TestDeleteAndRestoreIncome(t *testing.T) {
	fake := &fakeIncomes{result: usecase.WriteResult{
		Income:   sample(t),
		Warnings: []domain.Warning{{Code: domain.WarningProvisionalParameters, Field: domain.FieldCollectedOn}},
	}}
	h := incomeshttp.NewHandler(fake)

	resp, err := h.DeleteIncome(authed(t), gen.DeleteIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID)})
	require.NoError(t, err)
	assert.IsType(t, gen.DeleteIncome204Response{}, resp)
	assert.True(t, fake.deleted)

	restored, err := h.RestoreIncome(authed(t), gen.RestoreIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID)})
	require.NoError(t, err)
	ok200, ok := restored.(gen.RestoreIncome200JSONResponse)
	require.True(t, ok, "%T", restored)
	require.Len(t, ok200.Body.Warnings, 1)
	assert.Equal(t, gen.WarningCodeProvisionalParameters, ok200.Body.Warnings[0].Code)
	assert.Equal(t, "collected_on", ok200.Body.Warnings[0].Field)
}

func TestListIncomes_mapsTheQueryAndThePage(t *testing.T) {
	fake := &fakeIncomes{page: usecase.Page{
		Items: []usecase.Income{sample(t)}, NextCursor: "next",
		Totals: usecase.PeriodTotals{Count: 3, Amount: 10, IVA: 2, ISRWithheld: 1, IVAWithheld: 4},
	}}
	h := incomeshttp.NewHandler(fake)
	regime, cursor, limit := "612", "abc", 20

	resp, err := h.ListIncomes(authed(t), gen.ListIncomesRequestObject{
		TaxpayerId: uid(taxpayerID),
		Params:     gen.ListIncomesParams{Period: "2026-09", Regime: &regime, Cursor: &cursor, Limit: &limit},
	})

	require.NoError(t, err)
	assert.Equal(t, callerID, fake.userID)
	assert.Equal(t, usecase.ListQuery{TaxpayerID: taxpayerID, Period: "2026-09", Regime: "612", Cursor: "abc", Limit: 20}, fake.list)
	page, ok := resp.(gen.ListIncomes200JSONResponse)
	require.True(t, ok, "%T", resp)
	require.Len(t, page.Body.Items, 1)
	require.NotNil(t, page.Body.NextCursor)
	assert.Equal(t, "next", *page.Body.NextCursor)
	assert.Equal(t, gen.PeriodTotals{Count: 3, AmountCents: 10, IvaCents: 2, IsrWithheldCents: 1, IvaWithheldCents: 4}, page.Body.Totals)
}

func TestListIncomes_lastPageAndDefaults(t *testing.T) {
	fake := &fakeIncomes{}
	h := incomeshttp.NewHandler(fake)

	resp, err := h.ListIncomes(authed(t), gen.ListIncomesRequestObject{TaxpayerId: uid(taxpayerID), Params: gen.ListIncomesParams{Period: "2026-09"}})

	require.NoError(t, err)
	assert.Equal(t, usecase.ListQuery{TaxpayerID: taxpayerID, Period: "2026-09"}, fake.list)
	body := encode(t, resp.(gen.ListIncomes200JSONResponse).Body)
	assert.Nil(t, body["next_cursor"])
	assert.Equal(t, []any{}, body["items"])
}

func TestGetIncomeDefaults(t *testing.T) {
	fake := &fakeIncomes{prefill: usecase.Defaults{
		IVARatesAllowed: []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate0, fiscal.IVAExempt}, SuggestedIVARate: fiscal.IVARate16,
		PlatformActivity: domain.ActivityLodging,
		LastUsed:         usecase.LastFacts{Operation: fiscal.OperationOther, ViaPlatform: true},
	}}
	h := incomeshttp.NewHandler(fake)
	on, activity := genDate(2026, time.September, 12), gen.PlatformActivity("lodging")

	resp, err := h.GetIncomeDefaults(authed(t), gen.GetIncomeDefaultsRequestObject{
		TaxpayerId: uid(taxpayerID),
		Params:     gen.GetIncomeDefaultsParams{Regime: "625", On: &on, PlatformActivity: &activity},
	})

	require.NoError(t, err)
	assert.Equal(t, callerID, fake.userID)
	assert.Equal(t, usecase.DefaultsQuery{
		TaxpayerID: taxpayerID, Regime: "625", On: "2026-09-12",
		PlatformActivity: domain.ActivityLodging,
	}, fake.defaults)
	got, ok := resp.(gen.GetIncomeDefaults200JSONResponse)
	require.True(t, ok, "%T", resp)
	assert.Equal(t, map[string]any{
		"iva_rates_allowed": []any{"rate_16", "rate_0", "exempt"}, "suggested_iva_rate": "rate_16",
		"platform_activity": "lodging",
		"last_used": map[string]any{
			"counterparty_kind": nil, "operation_kind": "other", "residential_rent": false, "via_platform": true,
		},
	}, encode(t, got.Body), "no counterparty preselected the first time (OQ-W2)")
}

func TestGetIncomeDefaults_today(t *testing.T) {
	fake := &fakeIncomes{prefill: usecase.Defaults{IVARatesAllowed: []fiscal.IVARate{fiscal.IVARate16}, SuggestedIVARate: fiscal.IVARate16}}
	h := incomeshttp.NewHandler(fake)

	resp, err := h.GetIncomeDefaults(authed(t), gen.GetIncomeDefaultsRequestObject{
		TaxpayerId: uid(taxpayerID), Params: gen.GetIncomeDefaultsParams{Regime: "612"},
	})

	require.NoError(t, err)
	assert.Empty(t, fake.defaults.On, "on defaults to today in the use case")
	assert.Empty(t, fake.defaults.PlatformActivity)
	body := encode(t, resp.(gen.GetIncomeDefaults200JSONResponse).Body)
	assert.NotContains(t, body, "isr_withholding_rate", "removed in 0.12.0 (WH-11)")
	assert.Nil(t, body["platform_activity"])
}

func TestGetIncomeDefaults_lastUsedCounterparty(t *testing.T) {
	h := incomeshttp.NewHandler(&fakeIncomes{prefill: usecase.Defaults{
		LastUsed: usecase.LastFacts{Counterparty: fiscal.ReceptorPM, Operation: fiscal.OperationRentRealEstate, ResidentialRent: true},
	}})
	resp, err := h.GetIncomeDefaults(authed(t), gen.GetIncomeDefaultsRequestObject{TaxpayerId: uid(taxpayerID)})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{
		"counterparty_kind": "pm", "operation_kind": "rent_real_estate", "residential_rent": true, "via_platform": false,
	}, encode(t, resp.(gen.GetIncomeDefaults200JSONResponse).Body)["last_used"])
}

func TestWarnings_carryTheirMessages(t *testing.T) {
	codes := []domain.WarningCode{
		domain.WarningDuplicateCFDIUUID, domain.WarningProvisionalParameters,
		domain.WarningWithholdingCPAUnverified, domain.WarningForeignCounterpartyNotSupported,
	}
	warnings := make([]domain.Warning, 0, len(codes))
	for _, c := range codes {
		warnings = append(warnings, domain.Warning{Code: c, Field: "f"})
	}
	h := incomeshttp.NewHandler(&fakeIncomes{result: usecase.WriteResult{Income: sample(t), Warnings: warnings, Created: true}})

	resp, err := h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{TaxpayerId: uid(taxpayerID), Body: &gen.IncomeCreate{}})

	require.NoError(t, err)
	got := resp.(gen.CreateIncome201JSONResponse).Body.Warnings
	require.Len(t, got, len(codes))
	seen := map[string]bool{}
	for i, w := range got {
		assert.Equal(t, gen.WarningCode(codes[i]), w.Code)
		assert.True(t, w.Code.Valid(), "a contract warning code")
		assert.NotEmpty(t, w.Message)
		assert.NotEqual(t, "Revisa este dato.", w.Message, "%s has its own text", w.Code)
		assert.False(t, seen[w.Message], "distinct texts")
		seen[w.Message] = true
	}
}

func TestWarnings_unknownCodeGetsTheGenericMessage(t *testing.T) {
	h := incomeshttp.NewHandler(&fakeIncomes{result: usecase.WriteResult{
		Income: sample(t), Warnings: []domain.Warning{{Code: "brand_new", Field: "f"}}, Created: true,
	}})

	resp, err := h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{TaxpayerId: uid(taxpayerID), Body: &gen.IncomeCreate{}})

	require.NoError(t, err)
	assert.Equal(t, "Revisa este dato.", resp.(gen.CreateIncome201JSONResponse).Body.Warnings[0].Message)
}

func TestNilBodies(t *testing.T) {
	fake := &fakeIncomes{result: usecase.WriteResult{Income: sample(t), Created: true}}
	h := incomeshttp.NewHandler(fake)

	_, err := h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{TaxpayerId: uid(taxpayerID)})
	require.NoError(t, err)
	assert.Equal(t, usecase.CreateInput{TaxpayerID: taxpayerID}, fake.create)

	_, err = h.UpdateIncome(authed(t), gen.UpdateIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID)})
	require.NoError(t, err)
	assert.Equal(t, usecase.Patch{}, fake.patch)
}

// A stored value the contract cannot carry is a server fault (500), never
// a client error.
func TestUnpresentableIncomeIsAServerFault(t *testing.T) {
	bad := sample(t)
	bad.ID = "not-a-uuid"
	for name, fake := range map[string]*fakeIncomes{
		"write": {result: usecase.WriteResult{Income: bad, Created: true}},
		"read":  {income: bad},
		"list":  {page: usecase.Page{Items: []usecase.Income{bad}}},
	} {
		t.Run(name, func(t *testing.T) {
			h := incomeshttp.NewHandler(fake)
			var err error
			switch name {
			case "write":
				_, err = h.CreateIncome(authed(t), gen.CreateIncomeRequestObject{TaxpayerId: uid(taxpayerID), Body: &gen.IncomeCreate{}})
			case "read":
				_, err = h.GetIncome(authed(t), gen.GetIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID)})
			case "list":
				_, err = h.ListIncomes(authed(t), gen.ListIncomesRequestObject{TaxpayerId: uid(taxpayerID)})
			}
			require.Error(t, err)
			_, isClient := errors.AsType[*apperr.Error](err)
			assert.False(t, isClient)
		})
	}
}

// An income the WH-08 backfill has not derived yet lacks the facts and the
// withholding the contract requires since 0.12.0 (WH-11): showing it would
// break the contract, so it is a server fault (500), never a partial body.
func TestGetIncome_underivedRecordIsAServerFault(t *testing.T) {
	for name, legacy := range map[string]func(*usecase.Income){
		"no withholding": func(i *usecase.Income) {
			i.Withholding = fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: 1, IVA: 2}}
		},
		"no counterparty": func(i *usecase.Income) { i.Counterparty = "" },
		"no operation":    func(i *usecase.Income) { i.Operation = "" },
	} {
		t.Run(name, func(t *testing.T) {
			income := sample(t)
			legacy(&income)
			h := incomeshttp.NewHandler(&fakeIncomes{income: income})

			_, err := h.GetIncome(authed(t), gen.GetIncomeRequestObject{TaxpayerId: uid(taxpayerID), IncomeId: uid(incomeID)})

			require.Error(t, err)
			_, isClient := errors.AsType[*apperr.Error](err)
			assert.False(t, isClient)
		})
	}
}

func TestPreviewIncomeWithholdings(t *testing.T) {
	record := sample(t).Withholding
	record.CPAUnverified = true
	fake := &fakeIncomes{preview: usecase.Preview{
		Withholding: record, Amount: 10_000_00, IVA: 1_600_00, Counterparty: fiscal.ReceptorGovFederal,
		Warnings: []domain.Warning{{Code: domain.WarningWithholdingCPAUnverified, Field: domain.FieldWithholding}},
	}}
	h := incomeshttp.NewHandler(fake)
	on, operation, residential, via, activity := genDate(2026, time.September, 12), gen.OperationKind("rent_movable"),
		true, false, gen.PlatformActivity("lodging")

	resp, err := h.PreviewIncomeWithholdings(authed(t), gen.PreviewIncomeWithholdingsRequestObject{
		TaxpayerId: uid(taxpayerID),
		Params: gen.PreviewIncomeWithholdingsParams{
			Regime: "612", On: &on, AmountCents: 10_000_00, IvaRate: "rate_16", IvaCents: 1_600_00,
			CounterpartyKind: "gov_federal", OperationKind: &operation, ResidentialRent: &residential,
			ViaPlatform: &via, PlatformActivity: &activity,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, usecase.PreviewQuery{
		TaxpayerID: taxpayerID, Regime: "612", On: "2026-09-12", Amount: 10_000_00, IVARate: fiscal.IVARate16,
		IVA: 1_600_00, Counterparty: fiscal.ReceptorGovFederal, Operation: fiscal.OperationRentMovable,
		ResidentialRent: true, ViaPlatform: &via, PlatformActivity: domain.ActivityLodging,
	}, fake.previewQuery)
	body := encode(t, resp.(gen.PreviewIncomeWithholdings200JSONResponse).Body)
	w := body["withholding"].(map[string]any)
	assert.Equal(t, "El gobierno federal te retiene 10 % de ISR por honorarios (LISR art. 106).",
		w["explanations"].([]any)[0].(map[string]any)["text"], "the payer named from the counterparty")
	assert.Equal(t, true, w["cpa_unverified"])
	assert.Equal(t, []any{map[string]any{
		"code": "withholding_cpa_unverified", "field": "withholding",
		"message": "Esta retención aún no la confirma un contador. Úsala como aproximado.",
	}}, body["warnings"])

	_, err = h.PreviewIncomeWithholdings(authed(t), gen.PreviewIncomeWithholdingsRequestObject{
		TaxpayerId: uid(taxpayerID), Params: gen.PreviewIncomeWithholdingsParams{Regime: "612", CounterpartyKind: "pm"},
	})
	require.NoError(t, err)
	assert.Equal(t, usecase.PreviewQuery{TaxpayerID: taxpayerID, Regime: "612", Counterparty: fiscal.ReceptorPM}, fake.previewQuery)
}

func TestPreviewIncomeWithholdings_refusals(t *testing.T) {
	refusal := apperr.New(apperr.CodeBusinessRuleViolation, "No.")
	h := incomeshttp.NewHandler(&fakeIncomes{err: refusal})
	_, err := h.PreviewIncomeWithholdings(authed(t), gen.PreviewIncomeWithholdingsRequestObject{TaxpayerId: uid(taxpayerID)})
	assert.ErrorIs(t, err, refusal)

	resp, err := h.PreviewIncomeWithholdings(t.Context(), gen.PreviewIncomeWithholdingsRequestObject{TaxpayerId: uid(taxpayerID)})
	require.NoError(t, err)
	_, ok := resp.(gen.PreviewIncomeWithholdings401JSONResponse)
	assert.True(t, ok, "%T", resp)

	underived := incomeshttp.NewHandler(&fakeIncomes{})
	_, err = underived.PreviewIncomeWithholdings(authed(t), gen.PreviewIncomeWithholdingsRequestObject{TaxpayerId: uid(taxpayerID)})
	require.Error(t, err, "a preview always derives: anything else is a server fault")
	_, isApp := errors.AsType[*apperr.Error](err)
	assert.False(t, isApp)
}
