package expenseshttp_test

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
	expenseshttp "github.com/romodevx/control-contable/backend/internal/expenses/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
)

const (
	callerID   = "0199d4a2-0000-7000-8000-00000000000a"
	taxpayerID = "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80"
	expenseID  = "0199d4a3-2b3c-7e4d-9f50-6b7c8d9e0f12"
	folio      = "3A9C7E21-5B4D-4F6A-8E2C-9D1B7A3F5C60"
)

var stamp = time.Date(2026, time.September, 12, 18, 30, 0, 0, time.UTC)

// fakeExpenses records what the handler passes and answers the canned
// values or err.
type fakeExpenses struct {
	userID, taxpayerID, expenseID string
	create                        usecase.CreateInput
	createRes                     usecase.CreateResult
	update                        usecase.UpdateInput
	list                          usecase.ListInput
	page                          usecase.Page
	categories                    usecase.CategoriesInput
	categoryList                  usecase.CategoryList
	expense                       usecase.Expense
	deleted                       bool
	err                           error
}

func (f *fakeExpenses) Create(_ context.Context, in usecase.CreateInput) (usecase.CreateResult, error) {
	f.create = in
	return f.createRes, f.err
}

func (f *fakeExpenses) Update(_ context.Context, in usecase.UpdateInput) (usecase.Expense, error) {
	f.update = in
	return f.expense, f.err
}

func (f *fakeExpenses) Get(_ context.Context, userID, taxpayerID, expenseID string) (usecase.Expense, error) {
	f.userID, f.taxpayerID, f.expenseID = userID, taxpayerID, expenseID
	return f.expense, f.err
}

func (f *fakeExpenses) List(_ context.Context, in usecase.ListInput) (usecase.Page, error) {
	f.list = in
	return f.page, f.err
}

func (f *fakeExpenses) Delete(_ context.Context, userID, taxpayerID, expenseID string) error {
	f.userID, f.taxpayerID, f.expenseID, f.deleted = userID, taxpayerID, expenseID, true
	return f.err
}

func (f *fakeExpenses) Restore(_ context.Context, userID, taxpayerID, expenseID string) (usecase.Expense, error) {
	f.userID, f.taxpayerID, f.expenseID = userID, taxpayerID, expenseID
	return f.expense, f.err
}

func (f *fakeExpenses) Categories(_ context.Context, in usecase.CategoriesInput) (usecase.CategoryList, error) {
	f.categories = in
	return f.categoryList, f.err
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

func cfdi(t *testing.T, raw string) fiscal.CFDIUUID {
	t.Helper()
	u, err := fiscal.ParseCFDIUUID(raw)
	require.NoError(t, err)
	return u
}

func genDate(year int, month time.Month, d int) openapi_types.Date {
	return openapi_types.Date{Time: time.Date(year, month, d, 0, 0, 0, 0, time.UTC)}
}

// sample is a stored expense with every optional field set.
func sample(t *testing.T) usecase.Expense {
	t.Helper()
	return usecase.Expense{
		ID: expenseID, TaxpayerID: taxpayerID, Regime: "612", Category: "software",
		Amount: 50000, IVARate: fiscal.IVARate16, IVA: 8000, PaymentMethod: domain.PaymentCardDebit,
		PaidOn: day(t, "2026-09-12"), HasCFDI: true, CFDIUUID: cfdi(t, folio),
		Withheld: fiscal.Withholdings{ISR: 100, IVA: 200}, Description: "Suscripción",
		Deductible: true, Source: domain.SourceUser, Warnings: []domain.Warning{}, Legend: domain.Legend,
		CreatedAt: stamp, UpdatedAt: stamp.Add(time.Hour),
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

func TestCreateExpense_mapsTheRequestAndAnswers201(t *testing.T) {
	fake := &fakeExpenses{createRes: usecase.CreateResult{Expense: sample(t), Created: true}}
	h := expenseshttp.NewHandler(fake)
	id := uid(expenseID)
	deductible, isr, iva := false, int64(100), int64(200)
	folioLower := "3a9c7e21-5b4d-4f6a-8e2c-9d1b7a3f5c60"
	description := "  Suscripción  "

	res, err := h.CreateExpense(authed(t), gen.CreateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID),
		Body: &gen.ExpenseCreate{
			Id: &id, RegimeCode: "612", CategoryCode: "software", AmountCents: 50000, IvaRate: gen.IvaRate("rate_16"),
			IvaCents: 8000, PaymentMethod: gen.PaymentMethod("card_debit"), PaidOn: genDate(2026, time.September, 12),
			HasCfdi: true, CfdiUuid: &folioLower, Deductible: &deductible, IsrWithheldCents: &isr, IvaWithheldCents: &iva,
			Description: &description,
		},
	})

	require.NoError(t, err)
	created, ok := res.(gen.CreateExpense201JSONResponse)
	require.True(t, ok, "%T", res)
	assert.Equal(t, usecase.CreateInput{
		UserID: callerID, TaxpayerID: taxpayerID, ID: expenseID, Regime: "612", Description: description,
		Draft: domain.Draft{
			Category: "software", Amount: 50000, IVARate: fiscal.IVARate16, IVA: 8000,
			PaymentMethod: domain.PaymentCardDebit, PaidOn: day(t, "2026-09-12"), HasCFDI: true,
			CFDIUUID: cfdi(t, folio), Deductible: domain.ClaimNotDeductible,
			Withheld: fiscal.Withholdings{ISR: 100, IVA: 200},
		},
	}, fake.create)
	body := encode(t, created.Body)
	assert.Equal(t, folio, body["cfdi_uuid"])
	assert.Equal(t, "Suscripción", body["description"])
	assert.Equal(t, "user", body["deductible_source"])
	assert.Equal(t, domain.Legend, body["legend"])
	assert.Equal(t, []any{}, body["warnings"], "never null")
	assert.Equal(t, "2026-09-12", body["paid_on"])
	assert.InDelta(t, 100, body["isr_withheld_cents"], 0)
}

func TestCreateExpense_optionalFieldsAbsent(t *testing.T) {
	fake := &fakeExpenses{createRes: usecase.CreateResult{Expense: sample(t), Created: false}}
	h := expenseshttp.NewHandler(fake)

	res, err := h.CreateExpense(authed(t), gen.CreateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID),
		Body: &gen.ExpenseCreate{
			RegimeCode: "612", CategoryCode: "software", AmountCents: 50000, IvaRate: gen.IvaRate("exempt"),
			PaymentMethod: gen.PaymentMethod("cash"), PaidOn: genDate(2026, time.September, 12),
		},
	})

	require.NoError(t, err)
	_, ok := res.(gen.CreateExpense200JSONResponse)
	require.True(t, ok, "an idempotent retry is 200, got %T", res)
	assert.Empty(t, fake.create.ID, "no client id: the server chooses one")
	assert.Equal(t, domain.UseDefault, fake.create.Draft.Deductible)
	assert.True(t, fake.create.Draft.CFDIUUID.IsZero())
	assert.Equal(t, fiscal.Withholdings{}, fake.create.Draft.Withheld)
	assert.Empty(t, fake.create.Description)
}

func TestCreateExpense_malformedFolioIs400(t *testing.T) {
	fake := &fakeExpenses{}
	h := expenseshttp.NewHandler(fake)
	bad := "not-a-folio"

	_, err := h.CreateExpense(authed(t), gen.CreateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID),
		Body:       &gen.ExpenseCreate{RegimeCode: "612", CategoryCode: "x", AmountCents: 1, CfdiUuid: &bad, PaidOn: genDate(2026, 9, 1)},
	})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
	require.Len(t, appErr.Fields, 1)
	assert.Equal(t, apperr.FieldError{Field: "cfdi_uuid", Code: "invalid", Message: "This value is not valid."}, appErr.Fields[0])
	assert.Empty(t, fake.create.UserID, "the use case is not reached")
}

// Without a principal (the router authenticates first, so only a wiring
// fault gets here) every operation answers the contract's 401 and never
// reaches the use case.
func TestHandler_everyOperationRequiresTheCaller(t *testing.T) {
	fake := &fakeExpenses{err: errors.New("must not be reached")}
	h := expenseshttp.NewHandler(fake)
	ctx := t.Context()
	responses := map[string]func() (any, error){
		"categories": func() (any, error) {
			return h.ListExpenseCategories(ctx, gen.ListExpenseCategoriesRequestObject{})
		},
		"list":    func() (any, error) { return h.ListExpenses(ctx, gen.ListExpensesRequestObject{}) },
		"create":  func() (any, error) { return h.CreateExpense(ctx, gen.CreateExpenseRequestObject{}) },
		"get":     func() (any, error) { return h.GetExpense(ctx, gen.GetExpenseRequestObject{}) },
		"update":  func() (any, error) { return h.UpdateExpense(ctx, gen.UpdateExpenseRequestObject{}) },
		"delete":  func() (any, error) { return h.DeleteExpense(ctx, gen.DeleteExpenseRequestObject{}) },
		"restore": func() (any, error) { return h.RestoreExpense(ctx, gen.RestoreExpenseRequestObject{}) },
	}
	for name, call := range responses {
		t.Run(name, func(t *testing.T) {
			res, err := call()

			require.NoError(t, err)
			body := encode(t, res)
			errBody, ok := body["Body"].(map[string]any)["error"].(map[string]any)
			require.True(t, ok, "%T is a 401 response", res)
			assert.Equal(t, "unauthorized", errBody["code"])
			headers, ok := body["Headers"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, `Bearer realm="control-contable"`, headers["WWWAuthenticate"])
			assert.False(t, fake.deleted)
		})
	}
}

// Use case errors are client-ready *apperr.Error values: the handler hands
// them to the router untouched, and so any other failure (a 500).
func TestHandler_passesUseCaseErrorsThrough(t *testing.T) {
	clientErr := apperr.New(apperr.CodeEntitlementRequired, "cap").
		WithFields(apperr.FieldError{Field: "expenses_remaining", Code: "trial_cap_reached"})
	infraErr := errors.New("db down")
	for _, want := range []error{clientErr, infraErr} {
		fake := &fakeExpenses{err: want}
		h := expenseshttp.NewHandler(fake)
		ctx := authed(t)
		calls := map[string]func() error{
			"categories": func() error {
				_, err := h.ListExpenseCategories(ctx, gen.ListExpenseCategoriesRequestObject{TaxpayerId: uid(taxpayerID)})
				return err
			},
			"list": func() error {
				_, err := h.ListExpenses(ctx, gen.ListExpensesRequestObject{TaxpayerId: uid(taxpayerID)})
				return err
			},
			"create": func() error {
				_, err := h.CreateExpense(ctx, gen.CreateExpenseRequestObject{TaxpayerId: uid(taxpayerID), Body: &gen.ExpenseCreate{}})
				return err
			},
			"get": func() error {
				_, err := h.GetExpense(ctx, gen.GetExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
				return err
			},
			"update": func() error {
				_, err := h.UpdateExpense(ctx, gen.UpdateExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID), Body: &gen.ExpenseUpdate{}})
				return err
			},
			"delete": func() error {
				_, err := h.DeleteExpense(ctx, gen.DeleteExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
				return err
			},
			"restore": func() error {
				_, err := h.RestoreExpense(ctx, gen.RestoreExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
				return err
			},
		}
		for name, call := range calls {
			assert.ErrorIs(t, call(), want, name)
		}
	}
}

func TestUpdateExpense_mergePatch(t *testing.T) {
	tests := []struct {
		name string
		body gen.ExpenseUpdate
		want usecase.Patch
	}{
		{"absent fields are unchanged", gen.ExpenseUpdate{}, usecase.Patch{}},
		{"null clears the folio, the description and the override",
			gen.ExpenseUpdate{
				CfdiUuid: nullable.NewNullNullable[string](), Description: nullable.NewNullNullable[string](),
				Deductible: nullable.NewNullNullable[bool](),
			},
			usecase.Patch{CFDIUUID: &fiscal.CFDIUUID{}, Description: ptr(""), Deductible: ptr(domain.UseDefault)}},
		{"values replace",
			gen.ExpenseUpdate{
				RegimeCode: ptr("606"), CategoryCode: ptr("predial"), AmountCents: ptr(int64(700)),
				IvaRate: ptr(gen.IvaRate("rate_0")), IvaCents: ptr(int64(0)), PaymentMethod: ptr(gen.PaymentMethod("transfer")),
				PaidOn: ptr(genDate(2026, time.August, 3)), HasCfdi: ptr(true),
				CfdiUuid: nullable.NewNullableWithValue(folio), Deductible: nullable.NewNullableWithValue(true),
				IsrWithheldCents: ptr(int64(10)), IvaWithheldCents: ptr(int64(20)),
				Description: nullable.NewNullableWithValue("Predial"),
			},
			usecase.Patch{
				Regime: ptr(fiscal.RegimeCode("606")), Category: ptr(domain.CategoryCode("predial")), Amount: ptr(moneyOf(700)),
				IVARate: ptr(fiscal.IVARate0), IVA: ptr(moneyOf(0)), PaymentMethod: ptr(domain.PaymentTransfer),
				PaidOn: ptr(day(t, "2026-08-03")), HasCFDI: ptr(true), CFDIUUID: ptr(cfdi(t, folio)),
				Deductible: ptr(domain.ClaimDeductible), ISRWithheld: ptr(moneyOf(10)), IVAWithheld: ptr(moneyOf(20)),
				Description: ptr("Predial"),
			}},
		{"false is an override too", gen.ExpenseUpdate{Deductible: nullable.NewNullableWithValue(false)},
			usecase.Patch{Deductible: ptr(domain.ClaimNotDeductible)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeExpenses{expense: sample(t)}
			h := expenseshttp.NewHandler(fake)
			body := tt.body

			res, err := h.UpdateExpense(authed(t), gen.UpdateExpenseRequestObject{
				TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID), Body: &body,
			})

			require.NoError(t, err)
			_, ok := res.(gen.UpdateExpense200JSONResponse)
			require.True(t, ok, "%T", res)
			assert.Equal(t, usecase.UpdateInput{UserID: callerID, TaxpayerID: taxpayerID, ExpenseID: expenseID, Patch: tt.want}, fake.update)
		})
	}
}

func TestUpdateExpense_nullFolioDecodedFromJSONClears(t *testing.T) {
	var body gen.ExpenseUpdate
	require.NoError(t, json.Unmarshal([]byte(`{"cfdi_uuid":null,"deductible":null,"description":null}`), &body))
	fake := &fakeExpenses{expense: sample(t)}

	_, err := expenseshttp.NewHandler(fake).UpdateExpense(authed(t), gen.UpdateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID), Body: &body,
	})

	require.NoError(t, err)
	assert.Equal(t, usecase.Patch{CFDIUUID: &fiscal.CFDIUUID{}, Description: ptr(""), Deductible: ptr(domain.UseDefault)}, fake.update.Patch)
}

func TestUpdateExpense_malformedFolioIs400(t *testing.T) {
	fake := &fakeExpenses{}

	_, err := expenseshttp.NewHandler(fake).UpdateExpense(authed(t), gen.UpdateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID),
		Body: &gen.ExpenseUpdate{CfdiUuid: nullable.NewNullableWithValue("nope")},
	})

	appErr, ok := errors.AsType[*apperr.Error](err)
	require.True(t, ok, "%v", err)
	assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
	assert.Equal(t, "cfdi_uuid", appErr.Fields[0].Field)
	assert.Empty(t, fake.update.UserID)
}

func TestGetExpense_readsTheCallersExpense(t *testing.T) {
	e := sample(t)
	e.CFDIUUID, e.Description = fiscal.CFDIUUID{}, ""
	fake := &fakeExpenses{expense: e}

	res, err := expenseshttp.NewHandler(fake).GetExpense(authed(t), gen.GetExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID),
	})

	require.NoError(t, err)
	got, ok := res.(gen.GetExpense200JSONResponse)
	require.True(t, ok, "%T", res)
	assert.Equal(t, []string{callerID, taxpayerID, expenseID}, []string{fake.userID, fake.taxpayerID, fake.expenseID})
	body := encode(t, got.Body)
	assert.Contains(t, body, "cfdi_uuid")
	assert.Nil(t, body["cfdi_uuid"], "a zero folio is null")
	assert.Contains(t, body, "description")
	assert.Nil(t, body["description"], "an empty description is null")
}

func TestDeleteAndRestoreExpense(t *testing.T) {
	fake := &fakeExpenses{expense: sample(t)}
	h := expenseshttp.NewHandler(fake)

	res, err := h.DeleteExpense(authed(t), gen.DeleteExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
	require.NoError(t, err)
	assert.IsType(t, gen.DeleteExpense204Response{}, res)
	assert.True(t, fake.deleted)
	assert.Equal(t, []string{callerID, taxpayerID, expenseID}, []string{fake.userID, fake.taxpayerID, fake.expenseID})

	fake.userID = ""
	res2, err := h.RestoreExpense(authed(t), gen.RestoreExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
	require.NoError(t, err)
	_, ok := res2.(gen.RestoreExpense200JSONResponse)
	require.True(t, ok, "%T", res2)
	assert.Equal(t, callerID, fake.userID)
}

func TestListExpenses_mapsTheQueryAndThePage(t *testing.T) {
	regime, cursor, limit := "612", "opaque", 20
	fake := &fakeExpenses{page: usecase.Page{
		Items:      []usecase.Expense{sample(t)},
		NextCursor: "next",
		Totals:     usecase.Totals{Count: 1, Amount: 50000, IVA: 8000, ISRWithheld: 100, IVAWithheld: 200},
		Legend:     domain.Legend,
	}}

	res, err := expenseshttp.NewHandler(fake).ListExpenses(authed(t), gen.ListExpensesRequestObject{
		TaxpayerId: uid(taxpayerID),
		Params:     gen.ListExpensesParams{Period: "2026-09", Regime: &regime, Cursor: &cursor, Limit: &limit},
	})

	require.NoError(t, err)
	page, ok := res.(gen.ListExpenses200JSONResponse)
	require.True(t, ok, "%T", res)
	assert.Equal(t, usecase.ListInput{
		UserID: callerID, TaxpayerID: taxpayerID, Period: "2026-09", Regime: "612", Cursor: "opaque", Limit: 20,
	}, fake.list)
	require.Len(t, page.Body.Items, 1)
	require.NotNil(t, page.Body.NextCursor)
	assert.Equal(t, "next", *page.Body.NextCursor)
	assert.Equal(t, gen.PeriodTotals{Count: 1, AmountCents: 50000, IvaCents: 8000, IsrWithheldCents: 100, IvaWithheldCents: 200}, page.Body.Totals)
	assert.Equal(t, gen.Legend(domain.Legend), page.Body.Legend)
}

func TestListExpenses_lastPageAndDefaults(t *testing.T) {
	fake := &fakeExpenses{page: usecase.Page{Legend: domain.Legend}}

	res, err := expenseshttp.NewHandler(fake).ListExpenses(authed(t), gen.ListExpensesRequestObject{
		TaxpayerId: uid(taxpayerID), Params: gen.ListExpensesParams{Period: "2026-09"},
	})

	require.NoError(t, err)
	page, ok := res.(gen.ListExpenses200JSONResponse)
	require.True(t, ok, "%T", res)
	assert.Equal(t, usecase.ListInput{UserID: callerID, TaxpayerID: taxpayerID, Period: "2026-09"}, fake.list)
	assert.Nil(t, page.Body.NextCursor)
	body := encode(t, page.Body)
	assert.Equal(t, []any{}, body["items"], "an empty page is [], never null")
}

func TestListExpenseCategories(t *testing.T) {
	on := genDate(2026, time.September, 12)
	fake := &fakeExpenses{categoryList: usecase.CategoryList{
		Items: []usecase.Category{
			{Code: "software", Name: "Software", Deductibility: domain.Deductible, DefaultDeductible: true, LegalBasis: "LISR 103"},
			{Code: "restaurantes", Name: "Restaurantes", Deductibility: domain.PartiallyDeductible, Note: "Sólo una parte"},
		},
		IVARatesAllowed:  []fiscal.IVARate{fiscal.IVARate16, fiscal.IVAExempt},
		CashPaymentLimit: 200000,
		Legend:           domain.Legend,
	}}

	res, err := expenseshttp.NewHandler(fake).ListExpenseCategories(authed(t), gen.ListExpenseCategoriesRequestObject{
		TaxpayerId: uid(taxpayerID), Params: gen.ListExpenseCategoriesParams{Regime: "612", On: &on},
	})

	require.NoError(t, err)
	list, ok := res.(gen.ListExpenseCategories200JSONResponse)
	require.True(t, ok, "%T", res)
	assert.Equal(t, usecase.CategoriesInput{UserID: callerID, TaxpayerID: taxpayerID, Regime: "612", On: "2026-09-12"}, fake.categories)
	body := encode(t, list.Body)
	assert.Equal(t, []any{"rate_16", "exempt"}, body["iva_rates_allowed"])
	assert.InDelta(t, 200000, body["cash_payment_limit_cents"], 0)
	assert.Equal(t, domain.Legend, body["legend"])
	items, ok := body["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 2)
	first, second := items[0].(map[string]any), items[1].(map[string]any)
	assert.Nil(t, first["note"])
	assert.Equal(t, "LISR 103", first["legal_basis"])
	assert.Equal(t, "Sólo una parte", second["note"])
	assert.Nil(t, second["legal_basis"], "an empty basis is null")
	assert.Equal(t, "partially_deductible", second["deductibility"])
}

func TestListExpenseCategories_defaultsToToday(t *testing.T) {
	fake := &fakeExpenses{categoryList: usecase.CategoryList{Legend: domain.Legend}}

	_, err := expenseshttp.NewHandler(fake).ListExpenseCategories(authed(t), gen.ListExpenseCategoriesRequestObject{
		TaxpayerId: uid(taxpayerID), Params: gen.ListExpenseCategoriesParams{Regime: "626"},
	})

	require.NoError(t, err)
	assert.Empty(t, fake.categories.On, "the use case reads today")
}

// Every warning the module can send has its es-MX message; the codes are the
// contract's.
func TestWarnings_carryTheirMessages(t *testing.T) {
	warnings := []domain.Warning{
		{Code: domain.WarningCashOverThreshold, Field: domain.FieldPaymentMethod},
		{Code: domain.WarningNoCFDI, Field: domain.FieldHasCFDI},
		{Code: domain.WarningNotISRDeductibleInRegime, Field: domain.FieldDeductible},
		{Code: domain.WarningCategoryNotDeductible, Field: domain.FieldDeductible},
		{Code: domain.WarningDuplicateCFDIUUID, Field: domain.FieldCFDIUUID},
		{Code: domain.WarningIVANotCreditableCashPayment, Field: domain.FieldPaymentMethod},
		{Code: usecase.WarningProvisionalParameters, Field: domain.FieldPaidOn},
	}
	e := sample(t)
	e.Warnings = warnings
	fake := &fakeExpenses{createRes: usecase.CreateResult{Expense: e, Created: true}}

	res, err := expenseshttp.NewHandler(fake).CreateExpense(authed(t), gen.CreateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), Body: &gen.ExpenseCreate{PaidOn: genDate(2026, 9, 12)},
	})

	require.NoError(t, err)
	created, ok := res.(gen.CreateExpense201JSONResponse)
	require.True(t, ok, "%T", res)
	require.Len(t, created.Body.Warnings, len(warnings))
	messages := map[string]bool{}
	for i, w := range created.Body.Warnings {
		assert.Equal(t, string(warnings[i].Code), string(w.Code))
		assert.True(t, w.Code.Valid(), "%s is a contract WarningCode", w.Code)
		assert.Equal(t, warnings[i].Field, w.Field)
		assert.NotEmpty(t, w.Message, w.Code)
		assert.NotContains(t, w.Message, folio, "a message never echoes the folio")
		messages[w.Message] = true
	}
	assert.Len(t, messages, len(warnings), "one message per code")
	assert.Equal(t, "Los pagos en efectivo de más de $2,000 no son deducibles.", created.Body.Warnings[0].Message,
		"the contract's example text")
}

func TestWarnings_unknownCodeGetsTheGenericMessage(t *testing.T) {
	e := sample(t)
	e.Warnings = []domain.Warning{{Code: "something_new", Field: "amount_cents"}}
	fake := &fakeExpenses{expense: e}

	res, err := expenseshttp.NewHandler(fake).RestoreExpense(authed(t), gen.RestoreExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID),
	})

	require.NoError(t, err)
	restored, ok := res.(gen.RestoreExpense200JSONResponse)
	require.True(t, ok, "%T", res)
	require.Len(t, restored.Body.Warnings, 1)
	assert.NotEmpty(t, restored.Body.Warnings[0].Message)
}

// A date outside the calendar the kernel accepts is a 400 on paid_on, on
// create and on update.
func TestPaidOnOutsideTheCalendarIs400(t *testing.T) {
	far := genDate(10000, time.January, 1)
	fake := &fakeExpenses{}
	h := expenseshttp.NewHandler(fake)

	_, createErr := h.CreateExpense(authed(t), gen.CreateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), Body: &gen.ExpenseCreate{PaidOn: far},
	})
	_, updateErr := h.UpdateExpense(authed(t), gen.UpdateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID), Body: &gen.ExpenseUpdate{PaidOn: &far},
	})

	for _, err := range []error{createErr, updateErr} {
		appErr, ok := errors.AsType[*apperr.Error](err)
		require.True(t, ok, "%v", err)
		assert.Equal(t, apperr.CodeValidationFailed, appErr.Code)
		assert.Equal(t, "paid_on", appErr.Fields[0].Field)
	}
	assert.Empty(t, fake.create.UserID)
	assert.Empty(t, fake.update.UserID)
}

// No body reaches the use case as an empty request: it names every
// required field (the router's contract validation answers first anyway).
func TestNilBodies(t *testing.T) {
	fake := &fakeExpenses{createRes: usecase.CreateResult{Expense: sample(t), Created: true}, expense: sample(t)}
	h := expenseshttp.NewHandler(fake)

	_, err := h.CreateExpense(authed(t), gen.CreateExpenseRequestObject{TaxpayerId: uid(taxpayerID)})
	require.NoError(t, err)
	assert.Equal(t, usecase.CreateInput{UserID: callerID, TaxpayerID: taxpayerID}, fake.create)

	_, err = h.UpdateExpense(authed(t), gen.UpdateExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
	require.NoError(t, err)
	assert.Equal(t, usecase.Patch{}, fake.update.Patch)
}

// A stored expense the contract cannot carry (a corrupt id or date) is a
// server fault: a plain error the router answers as 500, never a client
// error.
func TestUnpresentableExpenseIsAServerFault(t *testing.T) {
	corrupt := map[string]func(*usecase.Expense){
		"expense id":  func(e *usecase.Expense) { e.ID = "x" },
		"taxpayer id": func(e *usecase.Expense) { e.TaxpayerID = "x" },
		"paid_on":     func(e *usecase.Expense) { e.PaidOn = fiscal.Date{} },
	}
	for name, mutate := range corrupt {
		t.Run(name, func(t *testing.T) {
			e := sample(t)
			mutate(&e)
			fake := &fakeExpenses{expense: e, createRes: usecase.CreateResult{Expense: e, Created: true}, page: usecase.Page{Items: []usecase.Expense{e}}}
			h := expenseshttp.NewHandler(fake)
			ctx := authed(t)
			errs := make([]error, 0, 5)
			_, err := h.CreateExpense(ctx, gen.CreateExpenseRequestObject{TaxpayerId: uid(taxpayerID), Body: &gen.ExpenseCreate{}})
			errs = append(errs, err)
			_, err = h.GetExpense(ctx, gen.GetExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
			errs = append(errs, err)
			_, err = h.UpdateExpense(ctx, gen.UpdateExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID), Body: &gen.ExpenseUpdate{}})
			errs = append(errs, err)
			_, err = h.RestoreExpense(ctx, gen.RestoreExpenseRequestObject{TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID)})
			errs = append(errs, err)
			_, err = h.ListExpenses(ctx, gen.ListExpensesRequestObject{TaxpayerId: uid(taxpayerID)})
			errs = append(errs, err)

			for i, err := range errs {
				require.Error(t, err, i)
				_, isClient := errors.AsType[*apperr.Error](err)
				assert.False(t, isClient, "call %d: %v", i, err)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

func moneyOf(v int64) money.Cents { return money.Cents(v) }
