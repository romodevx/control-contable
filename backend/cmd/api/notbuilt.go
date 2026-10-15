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

func (notBuilt) RegisterDevice(context.Context, gen.RegisterDeviceRequestObject) (gen.RegisterDeviceResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UnregisterDevice(context.Context, gen.UnregisterDeviceRequestObject) (gen.UnregisterDeviceResponseObject, error) {
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

func (notBuilt) ReceiveRevenueCatWebhook(context.Context, gen.ReceiveRevenueCatWebhookRequestObject) (gen.ReceiveRevenueCatWebhookResponseObject, error) {
	return nil, errNotBuilt
}

// The module_annual operations are built by ANN-15 (ADR-0009): the annual
// handlers in internal/taxengine/adapter/http/annual delete these stubs.

func (notBuilt) GetAnnualEstimate(context.Context, gen.GetAnnualEstimateRequestObject) (gen.GetAnnualEstimateResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetAnnualBreakdown(context.Context, gen.GetAnnualBreakdownRequestObject) (gen.GetAnnualBreakdownResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetAnnualInputs(context.Context, gen.GetAnnualInputsRequestObject) (gen.GetAnnualInputsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) PutAnnualInputs(context.Context, gen.PutAnnualInputsRequestObject) (gen.PutAnnualInputsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListSalaryStatements(context.Context, gen.ListSalaryStatementsRequestObject) (gen.ListSalaryStatementsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) CreateSalaryStatement(context.Context, gen.CreateSalaryStatementRequestObject) (gen.CreateSalaryStatementResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdateSalaryStatement(context.Context, gen.UpdateSalaryStatementRequestObject) (gen.UpdateSalaryStatementResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) DeleteSalaryStatement(context.Context, gen.DeleteSalaryStatementRequestObject) (gen.DeleteSalaryStatementResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) ListPersonalDeductions(context.Context, gen.ListPersonalDeductionsRequestObject) (gen.ListPersonalDeductionsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) CreatePersonalDeduction(context.Context, gen.CreatePersonalDeductionRequestObject) (gen.CreatePersonalDeductionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) UpdatePersonalDeduction(context.Context, gen.UpdatePersonalDeductionRequestObject) (gen.UpdatePersonalDeductionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) DeletePersonalDeduction(context.Context, gen.DeletePersonalDeductionRequestObject) (gen.DeletePersonalDeductionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) RestorePersonalDeduction(context.Context, gen.RestorePersonalDeductionRequestObject) (gen.RestorePersonalDeductionResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) GetFiledPayments(context.Context, gen.GetFiledPaymentsRequestObject) (gen.GetFiledPaymentsResponseObject, error) {
	return nil, errNotBuilt
}

func (notBuilt) PutFiledPayments(context.Context, gen.PutFiledPaymentsRequestObject) (gen.PutFiledPaymentsResponseObject, error) {
	return nil, errNotBuilt
}
