package expenseshttp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	expenseshttp "github.com/romodevx/control-contable/backend/internal/expenses/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// usW6 is US-W6 as the use cases return it: a 601 pays a 612 consultant
// 40,000.00 + IVA 6,400.00 on 2026-06-15.
func usW6(t *testing.T) usecase.Expense {
	t.Helper()
	e := sample(t)
	e.Regime, e.Category, e.Amount, e.IVA, e.PaidOn = "601", "servicios_profesionales", 40_000_00, 6_400_00, day(t, "2026-06-15")
	rate, err := money.ParseRate("0.10")
	require.NoError(t, err)
	e.Supplier = domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612,
		Operation: fiscal.OperationProfessionalServices}
	e.Withholding = fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: 4_000_00, IVA: 4_266_67}, ISRRate: rate,
		IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
		RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6}, TableVersion: fiscal.WithholdingTableVersion,
	}
	e.RemitDueOn = day(t, "2026-07-17")
	return e
}

// The contract's moralFeesWithholding example (previewExpenseWithholdings)
// is the oracle of the served object.
func TestGetExpense_servesTheStoredWithholding(t *testing.T) {
	fake := &fakeExpenses{expense: usW6(t)}

	res, err := expenseshttp.NewHandler(fake).GetExpense(authed(t), gen.GetExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID),
	})

	require.NoError(t, err)
	got, ok := res.(gen.GetExpense200JSONResponse)
	require.True(t, ok, "%T", res)
	body := encode(t, got.Body)
	assert.Equal(t, map[string]any{
		"isr_cents": 400000.0, "iva_cents": 426667.0, "isr_rate": "0.1", "iva_basis": "two_thirds_of_iva",
		"withholder": "you", "net_cents": 3813333.0, "remit_due_on": "2026-07-17",
		"rule_ids": []any{"ISR-5", "IVA-6"},
		"explanations": []any{
			map[string]any{"rule_id": "ISR-5", "text": "Debes retener 10 % de ISR por honorarios de persona física (LISR art. 106)."},
			map[string]any{"rule_id": "IVA-6", "text": "Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3)."},
		},
		"table_version": "2026.1", "cpa_unverified": false,
	}, body["withholding"])
	assert.Equal(t, "pf_with_activity", body["supplier_kind"])
	assert.Equal(t, "612", body["supplier_regime"])
	assert.Equal(t, "professional_services", body["operation_kind"])
	assert.Equal(t, false, body["residential_rent"])
	assert.InDelta(t, 400000, body["isr_withheld_cents"], 0, "the read-only mirror")
}

// A persona física's expense: no supplier facts, a zero withholding with
// withholder none and no remittance date.
func TestGetExpense_personaFisicaWithholdsNothing(t *testing.T) {
	e := sample(t)
	e.Withholding.Withholdings = fiscal.Withholdings{}
	fake := &fakeExpenses{expense: e}

	res, err := expenseshttp.NewHandler(fake).GetExpense(authed(t), gen.GetExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID),
	})

	require.NoError(t, err)
	body := encode(t, res.(gen.GetExpense200JSONResponse).Body)
	withholding, ok := body["withholding"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "none", withholding["withholder"])
	assert.Nil(t, withholding["remit_due_on"])
	assert.InDelta(t, 58000, withholding["net_cents"], 0)
	assert.NotContains(t, body, "supplier_kind")
}

// A record stored before WH-07 has no derivation: withholding is absent
// until the WH-08 backfill.
func TestGetExpense_legacyRecordHasNoWithholding(t *testing.T) {
	e := sample(t)
	e.Withholding = fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: 100}}
	fake := &fakeExpenses{expense: e}

	res, err := expenseshttp.NewHandler(fake).GetExpense(authed(t), gen.GetExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID),
	})

	require.NoError(t, err)
	body := encode(t, res.(gen.GetExpense200JSONResponse).Body)
	assert.NotContains(t, body, "withholding")
	assert.InDelta(t, 100, body["isr_withheld_cents"], 0)
}

// US-W8: a typed amount, whatever its value, is the 400 a schema without
// the field answers; the use case is never reached.
func TestTypedWithholdingsAre400(t *testing.T) {
	zero, some := int64(0), int64(1_000_00)
	fake := &fakeExpenses{}
	h := expenseshttp.NewHandler(fake)

	_, createErr := h.CreateExpense(authed(t), gen.CreateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), Body: &gen.ExpenseCreate{IsrWithheldCents: &zero, IvaWithheldCents: &some},
	})
	_, updateErr := h.UpdateExpense(authed(t), gen.UpdateExpenseRequestObject{
		TaxpayerId: uid(taxpayerID), ExpenseId: uid(expenseID), Body: &gen.ExpenseUpdate{IvaWithheldCents: &zero},
	})

	for name, err := range map[string]error{"create": createErr, "update": updateErr} {
		appErr, ok := err.(*apperr.Error) //nolint:errorlint // the adapter returns the *apperr.Error itself
		require.True(t, ok, "%s: %T", name, err)
		assert.Equal(t, apperr.CodeValidationFailed, appErr.Code, name)
		for _, f := range appErr.Fields {
			assert.Equal(t, "additionalProperties", f.Code, name)
		}
	}
	createFields := createErr.(*apperr.Error).Fields //nolint:errorlint // as above
	assert.Equal(t, []string{"isr_withheld_cents", "iva_withheld_cents"}, []string{createFields[0].Field, createFields[1].Field})
	assert.Equal(t, usecase.CreateInput{}, fake.create)
	assert.Equal(t, usecase.UpdateInput{}, fake.update)
}

func TestPreviewExpenseWithholdings_mapsTheQueryAndTheResult(t *testing.T) {
	stored := usW6(t)
	fake := &fakeExpenses{previewRes: usecase.Preview{
		Withholding: stored.Withholding, RemitDueOn: stored.RemitDueOn, Subtotal: 40_000_00, IVA: 6_400_00,
		Warnings: []domain.Warning{{Code: domain.WarningDeductionRequiresRemittance, Field: domain.FieldWithholding,
			DueOn: day(t, "2026-07-17")}},
	}}
	on := genDate(2026, time.June, 15)
	kind, operation, residential := gen.SupplierKind("pf_with_activity"), gen.OperationKind("professional_services"), false

	res, err := expenseshttp.NewHandler(fake).PreviewExpenseWithholdings(authed(t), gen.PreviewExpenseWithholdingsRequestObject{
		TaxpayerId: uid(taxpayerID),
		Params: gen.PreviewExpenseWithholdingsParams{
			Regime: "601", On: &on, AmountCents: 40_000_00, IvaRate: "rate_16", IvaCents: 6_400_00,
			CategoryCode: "servicios_profesionales", SupplierKind: &kind, OperationKind: &operation, ResidentialRent: &residential,
		},
	})

	require.NoError(t, err)
	assert.Equal(t, usecase.PreviewInput{
		UserID: callerID, TaxpayerID: taxpayerID, Regime: "601", On: "2026-06-15", Category: "servicios_profesionales",
		Amount: 40_000_00, IVARate: fiscal.IVARate16, IVA: 6_400_00,
		Supplier: domain.Supplier{Kind: domain.SupplierPFWithActivity, Operation: fiscal.OperationProfessionalServices},
	}, fake.preview)
	got, ok := res.(gen.PreviewExpenseWithholdings200JSONResponse)
	require.True(t, ok, "%T", res)
	assert.Equal(t, int64(38_133_33), got.Body.Withholding.NetCents)
	assert.Equal(t, gen.WithholderYou, got.Body.Withholding.Withholder)
	require.Len(t, got.Body.Warnings, 1)
	assert.Equal(t, "Este gasto solo es deducible si enteras las retenciones a más tardar el 17 de julio.",
		got.Body.Warnings[0].Message, "the contract's example text")
}

// A preview the use case could not derive is a server fault, never an empty
// card.
func TestPreviewExpenseWithholdings_underivedIsAServerFault(t *testing.T) {
	fake := &fakeExpenses{}

	_, err := expenseshttp.NewHandler(fake).PreviewExpenseWithholdings(authed(t), gen.PreviewExpenseWithholdingsRequestObject{
		TaxpayerId: uid(taxpayerID),
	})

	require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
}

func TestListExpenseCategories_noDefaultOperationOnTheWireYet(t *testing.T) {
	fake := &fakeExpenses{categoryList: usecase.CategoryList{Items: []usecase.Category{
		{Code: "fletes", Name: "Fletes", Deductibility: domain.Deductible, DefaultOperation: fiscal.OperationGroundFreight},
	}}}

	res, err := expenseshttp.NewHandler(fake).ListExpenseCategories(authed(t), gen.ListExpenseCategoriesRequestObject{
		TaxpayerId: uid(taxpayerID),
	})

	require.NoError(t, err)
	require.Len(t, res.(gen.ListExpenseCategories200JSONResponse).Body.Items, 1)
}
