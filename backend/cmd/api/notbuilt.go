package main

import (
	"context"
	"reflect"
	"unicode"
	"unicode/utf8"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// errNotBuilt is what a not-built operation answers if it is ever reached:
// the hidden-module 404 envelope (P0-29 FU2), never a panic or a 500.
var errNotBuilt = apperr.New(apperr.CodeNotFound, "Not found.")

// notBuilt is the explicit stub for every contract operation whose module is
// not built yet. Each method answers errNotBuilt, and router.Options.NotBuilt
// (fed by notBuiltOperations) answers the 404 before the stub is reached, so
// a flag turned ON early shows nothing new. It is embedded in api at the same
// depth as the module handlers: a module that implements one of these
// operations is a compile error ("ambiguous selector") until its stub line is
// deleted here, so a stub can never silently shadow, or be shadowed by, a
// built handler.
type notBuilt struct{}

// notBuiltOperations returns the operationIds notBuilt stands in for: its
// method names with the first letter lowered (oapi-codegen upper-cases the
// operationId's first letter and keeps the rest).
func notBuiltOperations() []string {
	typ := reflect.TypeFor[notBuilt]()
	ids := make([]string, 0, typ.NumMethod())
	for i := range typ.NumMethod() {
		ids = append(ids, lowerFirst(typ.Method(i).Name))
	}
	return ids
}

func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

func (notBuilt) Login(context.Context, gen.LoginRequestObject) (gen.LoginResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) Logout(context.Context, gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) RefreshSession(context.Context, gen.RefreshSessionRequestObject) (gen.RefreshSessionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) RegisterDevice(context.Context, gen.RegisterDeviceRequestObject) (gen.RegisterDeviceResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UnregisterDevice(context.Context, gen.UnregisterDeviceRequestObject) (gen.UnregisterDeviceResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListExpenseCategories(context.Context, gen.ListExpenseCategoriesRequestObject) (gen.ListExpenseCategoriesResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListLegalDocuments(context.Context, gen.ListLegalDocumentsRequestObject) (gen.ListLegalDocumentsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) DeleteMe(context.Context, gen.DeleteMeRequestObject) (gen.DeleteMeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetMe(context.Context, gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdateMe(context.Context, gen.UpdateMeRequestObject) (gen.UpdateMeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListRegimes(context.Context, gen.ListRegimesRequestObject) (gen.ListRegimesResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetReminderSettings(context.Context, gen.GetReminderSettingsRequestObject) (gen.GetReminderSettingsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) PutReminderSettings(context.Context, gen.PutReminderSettingsRequestObject) (gen.PutReminderSettingsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetSubscription(context.Context, gen.GetSubscriptionRequestObject) (gen.GetSubscriptionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) SyncSubscription(context.Context, gen.SyncSubscriptionRequestObject) (gen.SyncSubscriptionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetTaxCalendar(context.Context, gen.GetTaxCalendarRequestObject) (gen.GetTaxCalendarResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListTaxpayers(context.Context, gen.ListTaxpayersRequestObject) (gen.ListTaxpayersResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) CreateTaxpayer(context.Context, gen.CreateTaxpayerRequestObject) (gen.CreateTaxpayerResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) DeleteTaxpayer(context.Context, gen.DeleteTaxpayerRequestObject) (gen.DeleteTaxpayerResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetTaxpayer(context.Context, gen.GetTaxpayerRequestObject) (gen.GetTaxpayerResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdateTaxpayer(context.Context, gen.UpdateTaxpayerRequestObject) (gen.UpdateTaxpayerResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ArchiveTaxpayer(context.Context, gen.ArchiveTaxpayerRequestObject) (gen.ArchiveTaxpayerResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetDashboard(context.Context, gen.GetDashboardRequestObject) (gen.GetDashboardResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetTaxpayerEstimate(context.Context, gen.GetTaxpayerEstimateRequestObject) (gen.GetTaxpayerEstimateResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetEstimateBreakdown(context.Context, gen.GetEstimateBreakdownRequestObject) (gen.GetEstimateBreakdownResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListExpenses(context.Context, gen.ListExpensesRequestObject) (gen.ListExpensesResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) CreateExpense(context.Context, gen.CreateExpenseRequestObject) (gen.CreateExpenseResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) DeleteExpense(context.Context, gen.DeleteExpenseRequestObject) (gen.DeleteExpenseResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetExpense(context.Context, gen.GetExpenseRequestObject) (gen.GetExpenseResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdateExpense(context.Context, gen.UpdateExpenseRequestObject) (gen.UpdateExpenseResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) RestoreExpense(context.Context, gen.RestoreExpenseRequestObject) (gen.RestoreExpenseResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetHistory(context.Context, gen.GetHistoryRequestObject) (gen.GetHistoryResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) MarkHistoryViewed(context.Context, gen.MarkHistoryViewedRequestObject) (gen.MarkHistoryViewedResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListIncomes(context.Context, gen.ListIncomesRequestObject) (gen.ListIncomesResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) CreateIncome(context.Context, gen.CreateIncomeRequestObject) (gen.CreateIncomeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) DeleteIncome(context.Context, gen.DeleteIncomeRequestObject) (gen.DeleteIncomeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetIncome(context.Context, gen.GetIncomeRequestObject) (gen.GetIncomeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdateIncome(context.Context, gen.UpdateIncomeRequestObject) (gen.UpdateIncomeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) RestoreIncome(context.Context, gen.RestoreIncomeRequestObject) (gen.RestoreIncomeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) AddTaxpayerRegime(context.Context, gen.AddTaxpayerRegimeRequestObject) (gen.AddTaxpayerRegimeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ArchiveTaxpayerRegime(context.Context, gen.ArchiveTaxpayerRegimeRequestObject) (gen.ArchiveTaxpayerRegimeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdateTaxpayerRegime(context.Context, gen.UpdateTaxpayerRegimeRequestObject) (gen.UpdateTaxpayerRegimeResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetTaxInputs(context.Context, gen.GetTaxInputsRequestObject) (gen.GetTaxInputsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) PutTaxInputs(context.Context, gen.PutTaxInputsRequestObject) (gen.PutTaxInputsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ReceiveRevenueCatWebhook(context.Context, gen.ReceiveRevenueCatWebhookRequestObject) (gen.ReceiveRevenueCatWebhookResponseObject, error) {
	return nil, errNotBuilt
}
