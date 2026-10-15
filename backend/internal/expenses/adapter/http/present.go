package expenseshttp

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Client-safe messages of a field the adapter cannot decode; they never
// echo the value (the router's and the use cases' wording).
const (
	msgInvalidInput = "One or more fields are invalid."
	msgFieldInvalid = "This value is not valid."
)

// dateLayout is a contract date (BusinessDate) and fiscal.Date's String.
const dateLayout = "2006-01-02"

// invalidField is the 400 naming one field the adapter could not decode
// (contract validation runs first, so this is a backstop).
func invalidField(field string, cause error) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: fiscal.ErrInvalid.Detail(), Message: msgFieldInvalid})
}

func regimeOf(code string) fiscal.RegimeCode { return fiscal.RegimeCode(code) }

// toCreateInput decodes a create body. A nil body is an empty draft: the
// use case reports every required field.
func toCreateInput(userID, taxpayerID string, body *gen.ExpenseCreate) (usecase.CreateInput, error) {
	in := usecase.CreateInput{UserID: userID, TaxpayerID: taxpayerID}
	if body == nil {
		return in, nil
	}
	paidOn, err := toDate(body.PaidOn)
	if err != nil {
		return usecase.CreateInput{}, err
	}
	var folio fiscal.CFDIUUID
	if body.CfdiUuid != nil {
		if folio, err = toFolio(*body.CfdiUuid); err != nil {
			return usecase.CreateInput{}, err
		}
	}
	if body.Id != nil {
		in.ID = body.Id.String()
	}
	if body.Description != nil {
		in.Description = *body.Description
	}
	in.Regime = regimeOf(body.RegimeCode)
	in.Draft = domain.Draft{
		Category: domain.CategoryCode(body.CategoryCode), Amount: money.Cents(body.AmountCents),
		IVARate: fiscal.IVARate(body.IvaRate), IVA: money.Cents(body.IvaCents),
		PaymentMethod: domain.PaymentMethod(body.PaymentMethod), PaidOn: paidOn, HasCFDI: body.HasCfdi,
		CFDIUUID: folio, Deductible: toOverride(body.Deductible),
		Supplier: supplierOf(body.SupplierKind, body.SupplierRegime, body.OperationKind, body.ResidentialRent),
	}
	return in, nil
}

// toPatch decodes a merge-patch: an absent field stays nil; null clears
// the folio and the description (to their zero values) and the override
// (UseDefault).
func toPatch(body *gen.ExpenseUpdate) (usecase.Patch, error) {
	var p usecase.Patch
	if body == nil {
		return p, nil
	}
	if body.PaidOn != nil {
		d, err := toDate(*body.PaidOn)
		if err != nil {
			return usecase.Patch{}, err
		}
		p.PaidOn = &d
	}
	folio, err := patchFolio(body.CfdiUuid)
	if err != nil {
		return usecase.Patch{}, err
	}
	p.CFDIUUID = folio
	p.Regime = convert(body.RegimeCode, regimeOf)
	p.Category = convert(body.CategoryCode, func(s string) domain.CategoryCode { return domain.CategoryCode(s) })
	p.Amount = convert(body.AmountCents, toCents)
	p.IVARate = convert(body.IvaRate, func(r gen.IvaRate) fiscal.IVARate { return fiscal.IVARate(r) })
	p.IVA = convert(body.IvaCents, toCents)
	p.PaymentMethod = convert(body.PaymentMethod, func(m gen.PaymentMethod) domain.PaymentMethod { return domain.PaymentMethod(m) })
	p.HasCFDI = body.HasCfdi
	p.SupplierKind = convert(body.SupplierKind, func(k gen.SupplierKind) domain.SupplierKind { return domain.SupplierKind(k) })
	p.SupplierRegime = convert(body.SupplierRegime, func(r gen.SupplierRegime) domain.SupplierRegime { return domain.SupplierRegime(r) })
	p.OperationKind = convert(body.OperationKind, func(o gen.OperationKind) fiscal.OperationKind { return fiscal.OperationKind(o) })
	p.ResidentialRent = body.ResidentialRent
	p.Deductible = patchOverride(body.Deductible)
	p.Description = patchDescription(body.Description)
	return p, nil
}

// patchFolio is nil when absent, a zero folio for null (clear), else the
// parsed folio.
func patchFolio(field nullable.Nullable[string]) (*fiscal.CFDIUUID, error) {
	if !field.IsSpecified() {
		return nil, nil
	}
	if field.IsNull() {
		return &fiscal.CFDIUUID{}, nil
	}
	folio, err := toFolio(field.MustGet())
	if err != nil {
		return nil, err
	}
	return &folio, nil
}

// patchOverride is nil when absent, UseDefault for null, else the claim.
func patchOverride(field nullable.Nullable[bool]) *domain.Override {
	if !field.IsSpecified() {
		return nil
	}
	o := domain.UseDefault
	if !field.IsNull() {
		o = toOverride(new(field.MustGet()))
	}
	return &o
}

// patchDescription is nil when absent and "" (clear) for null.
func patchDescription(field nullable.Nullable[string]) *string {
	if !field.IsSpecified() {
		return nil
	}
	text := ""
	if !field.IsNull() {
		text = field.MustGet()
	}
	return &text
}

// toOverride maps the contract's deductible (boolean | null).
func toOverride(deductible *bool) domain.Override {
	switch {
	case deductible == nil:
		return domain.UseDefault
	case *deductible:
		return domain.ClaimDeductible
	default:
		return domain.ClaimNotDeductible
	}
}

// toDate maps a contract date; the zero value is an absent date (the use
// case reports it as required).
func toDate(d openapi_types.Date) (fiscal.Date, error) {
	if d.IsZero() {
		return fiscal.Date{}, nil
	}
	out, err := fiscal.ParseDate(d.String())
	if err != nil {
		return fiscal.Date{}, invalidField(domain.FieldPaidOn, err)
	}
	return out, nil
}

func toFolio(raw string) (fiscal.CFDIUUID, error) {
	folio, err := fiscal.ParseCFDIUUID(raw)
	if err != nil {
		return fiscal.CFDIUUID{}, invalidField(domain.FieldCFDIUUID, err)
	}
	return folio, nil
}

func toCents(v int64) money.Cents { return money.Cents(v) }

// convert maps an optional value, keeping nil.
func convert[From, To any](v *From, f func(From) To) *To {
	if v == nil {
		return nil
	}
	out := f(*v)
	return &out
}

func toListInput(userID string, req gen.ListExpensesRequestObject) usecase.ListInput {
	in := usecase.ListInput{UserID: userID, TaxpayerID: req.TaxpayerId.String(), Period: req.Params.Period}
	if req.Params.Regime != nil {
		in.Regime = regimeOf(*req.Params.Regime)
	}
	if req.Params.Cursor != nil {
		in.Cursor = *req.Params.Cursor
	}
	if req.Params.Limit != nil {
		in.Limit = *req.Params.Limit
	}
	return in
}

// toExpense maps a stored expense. Its ids and date come from the
// database's own types, so a value that fails to parse is a server fault
// (500), never a client error.
func toExpense(e usecase.Expense) (gen.Expense, error) {
	id, err := uuid.Parse(e.ID)
	if err != nil {
		return gen.Expense{}, fmt.Errorf("expenses http: expense id: %w", err)
	}
	taxpayerID, err := uuid.Parse(e.TaxpayerID)
	if err != nil {
		return gen.Expense{}, fmt.Errorf("expenses http: taxpayer id of %s: %w", e.ID, err)
	}
	paidOn, err := time.Parse(dateLayout, e.PaidOn.String())
	if err != nil {
		return gen.Expense{}, fmt.Errorf("expenses http: paid_on of %s: %w", e.ID, err)
	}
	withholding, err := storedWithholding(e.Withholding, e.Amount, e.IVA, e.RemitDueOn)
	if err != nil {
		return gen.Expense{}, fmt.Errorf("expenses http: withholding of %s: %w", e.ID, err)
	}
	kind, regime, operation := supplierFields(e.Supplier)
	return gen.Expense{
		Id: id, TaxpayerId: taxpayerID, RegimeCode: string(e.Regime), CategoryCode: string(e.Category),
		AmountCents: int64(e.Amount), IvaRate: gen.IvaRate(e.IVARate), IvaCents: int64(e.IVA),
		PaymentMethod: gen.PaymentMethod(e.PaymentMethod), PaidOn: openapi_types.Date{Time: paidOn}, HasCfdi: e.HasCFDI,
		CfdiUuid:         nullableText(e.CFDIUUID.String()),
		IsrWithheldCents: int64(e.Withholding.ISR), IvaWithheldCents: int64(e.Withholding.IVA),
		SupplierKind: kind, SupplierRegime: regime, OperationKind: operation,
		ResidentialRent: e.Supplier.ResidentialRent, Withholding: withholding,
		Description: nullableText(e.Description), Deductible: e.Deductible,
		DeductibleSource: gen.DeductibleSource(e.Source), Warnings: toWarnings(e.Warnings),
		Legend: gen.Legend(e.Legend), CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}, nil
}

// nullableText is nil (JSON null) for "".
func nullableText(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toPage(p usecase.Page) (gen.ExpensePage, error) {
	items := make([]gen.Expense, 0, len(p.Items))
	for _, e := range p.Items {
		item, err := toExpense(e)
		if err != nil {
			return gen.ExpensePage{}, err
		}
		items = append(items, item)
	}
	return gen.ExpensePage{
		Items: items, NextCursor: nullableText(p.NextCursor), Legend: gen.Legend(p.Legend),
		Totals: gen.PeriodTotals{
			Count: p.Totals.Count, AmountCents: int64(p.Totals.Amount), IvaCents: int64(p.Totals.IVA),
			IsrWithheldCents: int64(p.Totals.ISRWithheld), IvaWithheldCents: int64(p.Totals.IVAWithheld),
		},
	}, nil
}

func toCategoryList(l usecase.CategoryList) gen.ExpenseCategoryList {
	items := make([]gen.ExpenseCategory, 0, len(l.Items))
	for _, c := range l.Items {
		items = append(items, gen.ExpenseCategory{
			Code: string(c.Code), Name: c.Name, Deductibility: gen.Deductibility(c.Deductibility),
			DefaultDeductible: c.DefaultDeductible, Note: nullableText(c.Note), LegalBasis: nullableText(c.LegalBasis),
			DefaultOperationKind: defaultOperationKind(c.DefaultOperation),
		})
	}
	rates := make([]gen.IvaRate, 0, len(l.IVARatesAllowed))
	for _, r := range l.IVARatesAllowed {
		rates = append(rates, gen.IvaRate(r))
	}
	return gen.ExpenseCategoryList{
		Items: items, IvaRatesAllowed: rates, CashPaymentLimitCents: int64(l.CashPaymentLimit), Legend: gen.Legend(l.Legend),
	}
}

// defaultOperationKind is the category's E3 default, absent for a persona
// física's category (contract ExpenseCategory.default_operation_kind).
func defaultOperationKind(o fiscal.OperationKind) *gen.OperationKind {
	if o == "" {
		return nil
	}
	out := gen.OperationKind(o)
	return &out
}
