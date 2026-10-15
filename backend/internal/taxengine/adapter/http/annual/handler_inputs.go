package annualhttp

import (
	"context"

	"github.com/oapi-codegen/nullable"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// GetAnnualInputs implements gen.StrictServerInterface.
func (h *Handler) GetAnnualInputs(ctx context.Context, req gen.GetAnnualInputsRequestObject) (gen.GetAnnualInputsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetAnnualInputs401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	view, err := h.deps.Facts.Get(ctx, userID, req.TaxpayerId.String(), req.FiscalYear)
	if err != nil {
		return nil, err
	}
	return gen.GetAnnualInputs200JSONResponse{Body: toInputs(view)}, nil
}

// PutAnnualInputs implements gen.StrictServerInterface.
func (h *Handler) PutAnnualInputs(ctx context.Context, req gen.PutAnnualInputsRequestObject) (gen.PutAnnualInputsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.PutAnnualInputs401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, bodyRequired()
	}
	facts := usecaseannual.Facts{
		OtherExemptIncome: cents(req.Body.OtherExemptIncomeCents), ToldEmployerWillFile: req.Body.ToldEmployerWillFile,
	}
	if p := req.Body.PriorYearAccumulableIncomeCents; p != nil {
		v := cents(*p)
		facts.PriorYearAccumulableIncome = &v
	}
	view, err := h.deps.Facts.Put(ctx, userID, req.TaxpayerId.String(), req.FiscalYear, facts)
	if err != nil {
		return nil, err
	}
	return gen.PutAnnualInputs200JSONResponse{Body: toInputs(view)}, nil
}

// ListSalaryStatements implements gen.StrictServerInterface.
func (h *Handler) ListSalaryStatements(ctx context.Context, req gen.ListSalaryStatementsRequestObject) (gen.ListSalaryStatementsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListSalaryStatements401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	list, err := h.deps.Salary.List(ctx, userID, req.TaxpayerId.String(), req.Params.FiscalYear)
	if err != nil {
		return nil, err
	}
	return gen.ListSalaryStatements200JSONResponse{Body: toSalaryList(list)}, nil
}

// CreateSalaryStatement implements gen.StrictServerInterface: 201 for a
// new constancia, 200 for an idempotent replay.
func (h *Handler) CreateSalaryStatement(ctx context.Context, req gen.CreateSalaryStatementRequestObject) (gen.CreateSalaryStatementResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.CreateSalaryStatement401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, bodyRequired()
	}
	b := req.Body
	in := usecaseannual.SalaryInput{
		TaxpayerID: req.TaxpayerId.String(), FiscalYear: b.FiscalYear, EmployerLabel: b.EmployerLabel,
		Accumulable: cents(b.AccumulableCents), Exempt: cents(b.ExemptCents), ISRWithheld: cents(b.IsrWithheldCents),
		Simultaneous: deref(b.Simultaneous), LeftBeforeYearEnd: deref(b.LeftBeforeYearEnd), PayerNotWithholding: deref(b.PayerNotWithholding),
	}
	if b.Id != nil {
		in.ID = b.Id.String()
	}
	res, err := h.deps.Salary.Create(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	body := toStatement(res.Statement, res.Warnings)
	if res.Created {
		return gen.CreateSalaryStatement201JSONResponse{Body: body}, nil
	}
	return gen.CreateSalaryStatement200JSONResponse{Body: body}, nil
}

// UpdateSalaryStatement implements gen.StrictServerInterface (merge-patch).
func (h *Handler) UpdateSalaryStatement(ctx context.Context, req gen.UpdateSalaryStatementRequestObject) (gen.UpdateSalaryStatementResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UpdateSalaryStatement401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, bodyRequired()
	}
	b := req.Body
	patch := usecaseannual.SalaryPatch{
		FiscalYear: b.FiscalYear, EmployerLabel: optional(b.EmployerLabel, func(s string) string { return s }),
		Accumulable: centsPtr(b.AccumulableCents), Exempt: centsPtr(b.ExemptCents), ISRWithheld: centsPtr(b.IsrWithheldCents),
		Simultaneous: b.Simultaneous, LeftBeforeYearEnd: b.LeftBeforeYearEnd, PayerNotWithholding: b.PayerNotWithholding,
	}
	res, err := h.deps.Salary.Update(ctx, userID, req.TaxpayerId.String(), req.StatementId.String(), patch)
	if err != nil {
		return nil, err
	}
	return gen.UpdateSalaryStatement200JSONResponse{Body: toStatement(res.Statement, res.Warnings)}, nil
}

// DeleteSalaryStatement implements gen.StrictServerInterface.
func (h *Handler) DeleteSalaryStatement(ctx context.Context, req gen.DeleteSalaryStatementRequestObject) (gen.DeleteSalaryStatementResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.DeleteSalaryStatement401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err := h.deps.Salary.Delete(ctx, userID, req.TaxpayerId.String(), req.StatementId.String()); err != nil {
		return nil, err
	}
	return gen.DeleteSalaryStatement204Response{}, nil
}

// ListPersonalDeductions implements gen.StrictServerInterface.
func (h *Handler) ListPersonalDeductions(ctx context.Context, req gen.ListPersonalDeductionsRequestObject) (gen.ListPersonalDeductionsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListPersonalDeductions401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	q := usecaseannual.DeductionQuery{TaxpayerID: req.TaxpayerId.String(), FiscalYear: req.Params.FiscalYear}
	if req.Params.Category != nil {
		q.Category = adomain.Category(*req.Params.Category)
	}
	list, err := h.deps.Deductions.List(ctx, userID, q)
	if err != nil {
		return nil, err
	}
	return gen.ListPersonalDeductions200JSONResponse{Body: toDeductionList(list)}, nil
}

// CreatePersonalDeduction implements gen.StrictServerInterface: 201 for a
// new deduction, 200 for an idempotent replay.
func (h *Handler) CreatePersonalDeduction(ctx context.Context, req gen.CreatePersonalDeductionRequestObject) (gen.CreatePersonalDeductionResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.CreatePersonalDeduction401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, bodyRequired()
	}
	b := req.Body
	paidOn, err := fiscal.ParseDate(b.PaidOn.String())
	if err != nil {
		return nil, invalidDate(fieldPaidOn)
	}
	in := usecaseannual.DeductionInput{
		TaxpayerID: req.TaxpayerId.String(), Category: adomain.Category(b.Category), Amount: cents(b.AmountCents), PaidOn: paidOn,
		PaymentMethod: adomain.PaymentMethod(b.PaymentMethod), Beneficiary: adomain.Beneficiary(b.Beneficiary),
		DoneeType:    convertPtr(b.DoneeType, func(v gen.DoneeType) adomain.DoneeType { return adomain.DoneeType(v) }),
		StudentLabel: b.StudentLabel, Scholarship: centsPtr(b.ScholarshipCents), CFDIUUID: b.CfdiUuid,
		TuitionLevel: convertPtr(b.TuitionLevel, func(v gen.TuitionLevel) adomain.TuitionLevel { return adomain.TuitionLevel(v) }),
	}
	if b.Id != nil {
		in.ID = b.Id.String()
	}
	res, err := h.deps.Deductions.Create(ctx, userID, in)
	if err != nil {
		return nil, err
	}
	body := toDeduction(res.Deduction, res.Warnings)
	if res.Created {
		return gen.CreatePersonalDeduction201JSONResponse{Body: body}, nil
	}
	return gen.CreatePersonalDeduction200JSONResponse{Body: body}, nil
}

// UpdatePersonalDeduction implements gen.StrictServerInterface (merge-patch).
func (h *Handler) UpdatePersonalDeduction(ctx context.Context, req gen.UpdatePersonalDeductionRequestObject) (gen.UpdatePersonalDeductionResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UpdatePersonalDeduction401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, bodyRequired()
	}
	patch, err := deductionPatch(*req.Body)
	if err != nil {
		return nil, err
	}
	res, err := h.deps.Deductions.Update(ctx, userID, req.TaxpayerId.String(), req.DeductionId.String(), patch)
	if err != nil {
		return nil, err
	}
	return gen.UpdatePersonalDeduction200JSONResponse{Body: toDeduction(res.Deduction, res.Warnings)}, nil
}

// DeletePersonalDeduction implements gen.StrictServerInterface (soft delete).
func (h *Handler) DeletePersonalDeduction(ctx context.Context, req gen.DeletePersonalDeductionRequestObject) (gen.DeletePersonalDeductionResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.DeletePersonalDeduction401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err := h.deps.Deductions.Delete(ctx, userID, req.TaxpayerId.String(), req.DeductionId.String()); err != nil {
		return nil, err
	}
	return gen.DeletePersonalDeduction204Response{}, nil
}

// RestorePersonalDeduction implements gen.StrictServerInterface (Undo).
func (h *Handler) RestorePersonalDeduction(ctx context.Context, req gen.RestorePersonalDeductionRequestObject) (gen.RestorePersonalDeductionResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.RestorePersonalDeduction401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	res, err := h.deps.Deductions.Restore(ctx, userID, req.TaxpayerId.String(), req.DeductionId.String())
	if err != nil {
		return nil, err
	}
	return gen.RestorePersonalDeduction200JSONResponse{Body: toDeduction(res.Deduction, res.Warnings)}, nil
}

// deductionPatch maps the merge-patch body: absent unchanged, null cleared.
func deductionPatch(b gen.PersonalDeductionUpdate) (usecaseannual.DeductionPatch, error) {
	p := usecaseannual.DeductionPatch{
		Category:      convertPtr(b.Category, func(v gen.DeductionCategory) adomain.Category { return adomain.Category(v) }),
		Amount:        centsPtr(b.AmountCents),
		PaymentMethod: convertPtr(b.PaymentMethod, func(v gen.DeductionPaymentMethod) adomain.PaymentMethod { return adomain.PaymentMethod(v) }),
		Beneficiary:   convertPtr(b.Beneficiary, func(v gen.DeductionBeneficiary) adomain.Beneficiary { return adomain.Beneficiary(v) }),
		DoneeType:     optional(b.DoneeType, func(v gen.DoneeType) adomain.DoneeType { return adomain.DoneeType(v) }),
		StudentLabel:  optional(b.StudentLabel, func(s string) string { return s }),
		TuitionLevel:  optional(b.TuitionLevel, func(v gen.TuitionLevel) adomain.TuitionLevel { return adomain.TuitionLevel(v) }),
		Scholarship:   optional(b.ScholarshipCents, func(v int64) money.Cents { return money.Cents(v) }),
		CFDIUUID:      optional(b.CfdiUuid, func(s string) string { return s }),
	}
	if b.PaidOn != nil {
		d, err := fiscal.ParseDate(b.PaidOn.String())
		if err != nil {
			return usecaseannual.DeductionPatch{}, invalidDate(fieldPaidOn)
		}
		p.PaidOn = &d
	}
	return p, nil
}

// optional maps a merge-patch nullable to the use case's Optional.
func optional[T, U any](n nullable.Nullable[T], conv func(T) U) usecaseannual.Optional[U] {
	switch {
	case !n.IsSpecified():
		return usecaseannual.Optional[U]{}
	case n.IsNull():
		return usecaseannual.Null[U]()
	}
	return usecaseannual.Value(conv(n.MustGet()))
}

func convertPtr[T, U any](p *T, conv func(T) U) *U {
	if p == nil {
		return nil
	}
	v := conv(*p)
	return &v
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func cents(v int64) money.Cents { return money.Cents(v) }

func centsPtr(p *int64) *money.Cents { return convertPtr(p, cents) }
