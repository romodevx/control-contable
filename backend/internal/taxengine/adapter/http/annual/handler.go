// Package annualhttp is the annual estimate's HTTP adapter (ADR-0009,
// PLAN ANN-15): the 15 x-module: module_annual operations of
// docs/api/openapi.yaml, translated to the annual use cases. The router
// gates them (module_annual and its prerequisites module_tax_engine and
// module_taxpayers: OFF → 404 before anything else), authenticates them and
// validates them against the contract; the use cases answer client-ready
// *apperr.Error values, which the router writes as the single envelope.
package annualhttp

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/ratelimit"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const (
	// bearerChallenge is the contract's fixed WWW-Authenticate value.
	bearerChallenge           = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	msgAuthenticationRequired = "Authentication required."
	msgRateLimited            = "Too many requests. Try again later."
	msgInternal               = "Something went wrong."
)

// Estimates is the annual estimate use case (usecaseannual.Estimator).
type Estimates interface {
	Estimate(ctx context.Context, userID, taxpayerID string, fiscalYear int) (usecaseannual.Estimate, error)
	Breakdown(ctx context.Context, userID, taxpayerID string, fiscalYear int, section string) (usecaseannual.Breakdown, error)
	Payments(ctx context.Context, userID, taxpayerID string, regime fiscal.RegimeCode, fiscalYear int) (usecaseannual.RegimePayments, error)
}

// Facts is getAnnualInputs / putAnnualInputs (usecaseannual.Inputs).
type Facts interface {
	Get(ctx context.Context, userID, taxpayerID string, fiscalYear int) (usecaseannual.InputsView, error)
	Put(ctx context.Context, userID, taxpayerID string, fiscalYear int, f usecaseannual.Facts) (usecaseannual.InputsView, error)
}

// Salary is the salary constancias' use cases (usecaseannual.SalaryStatements).
type Salary interface {
	List(ctx context.Context, userID, taxpayerID string, fiscalYear int) (usecaseannual.SalaryList, error)
	Create(ctx context.Context, userID string, in usecaseannual.SalaryInput) (usecaseannual.SalaryResult, error)
	Update(ctx context.Context, userID, taxpayerID, id string, p usecaseannual.SalaryPatch) (usecaseannual.SalaryResult, error)
	Delete(ctx context.Context, userID, taxpayerID, id string) error
}

// Deductions is the personal deductions' use cases (usecaseannual.Deductions).
type Deductions interface {
	List(ctx context.Context, userID string, q usecaseannual.DeductionQuery) (usecaseannual.DeductionList, error)
	Create(ctx context.Context, userID string, in usecaseannual.DeductionInput) (usecaseannual.DeductionResult, error)
	Update(ctx context.Context, userID, taxpayerID, id string, p usecaseannual.DeductionPatch) (usecaseannual.DeductionResult, error)
	Delete(ctx context.Context, userID, taxpayerID, id string) error
	Restore(ctx context.Context, userID, taxpayerID, id string) (usecaseannual.DeductionResult, error)
}

// Filed is getFiledPayments / putFiledPayments (usecaseannual.FiledPayments).
type Filed interface {
	Get(ctx context.Context, userID string, req usecaseannual.FiledRequest) (usecaseannual.FiledView, error)
	Put(ctx context.Context, userID string, req usecaseannual.FiledRequest, entries []usecaseannual.FiledEntry) (usecaseannual.FiledView, error)
}

// RangeBudget is the estimate_reads bucket (router.Limiter): a read that
// ran the 12-month Estimator range spends one more (ADR-0009 "Budgets").
type RangeBudget interface {
	Take(ctx context.Context, key string) (allowed bool, retryAfter time.Duration, err error)
}

// Deps are the handler's use cases.
type Deps struct {
	Estimates  Estimates
	Facts      Facts
	Salary     Salary
	Deductions Deductions
	Filed      Filed
	// Range is spent once more by a read that ran the Estimator range;
	// nil spends nothing (tests).
	Range RangeBudget
}

// Handler implements the module_annual operations of
// gen.StrictServerInterface; cmd/api embeds it in the API server. The user
// id always comes from the authenticated principal, never the request.
type Handler struct{ deps Deps }

// NewHandler returns the handler over deps.
func NewHandler(deps Deps) *Handler { return &Handler{deps: deps} }

// GetAnnualEstimate implements gen.StrictServerInterface.
func (h *Handler) GetAnnualEstimate(ctx context.Context, req gen.GetAnnualEstimateRequestObject) (gen.GetAnnualEstimateResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetAnnualEstimate401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	est, err := h.deps.Estimates.Estimate(ctx, userID, req.TaxpayerId.String(), req.FiscalYear)
	if err != nil {
		return nil, err
	}
	wait, limited, err := h.spendRange(ctx, userID, est.RangeUsed)
	switch {
	case err != nil:
		return nil, err
	case limited:
		return gen.GetAnnualEstimate429JSONResponse{TooManyRequestsJSONResponse: tooMany(ctx, wait)}, nil
	}
	return gen.GetAnnualEstimate200JSONResponse{Body: toEstimate(req.TaxpayerId, est)}, nil
}

// GetAnnualBreakdown implements gen.StrictServerInterface.
func (h *Handler) GetAnnualBreakdown(ctx context.Context, req gen.GetAnnualBreakdownRequestObject) (gen.GetAnnualBreakdownResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetAnnualBreakdown401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	b, err := h.deps.Estimates.Breakdown(ctx, userID, req.TaxpayerId.String(), req.FiscalYear, string(req.Params.Section))
	if err != nil {
		return nil, err
	}
	wait, limited, err := h.spendRange(ctx, userID, b.RangeUsed)
	switch {
	case err != nil:
		return nil, err
	case limited:
		return gen.GetAnnualBreakdown429JSONResponse{TooManyRequestsJSONResponse: tooMany(ctx, wait)}, nil
	}
	return gen.GetAnnualBreakdown200JSONResponse{Body: toBreakdown(req.TaxpayerId, b)}, nil
}

// GetFiledPayments implements gen.StrictServerInterface: the stored
// months, then every source's column from the estimate's payments.
func (h *Handler) GetFiledPayments(ctx context.Context, req gen.GetFiledPaymentsRequestObject) (gen.GetFiledPaymentsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetFiledPayments401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	fr := usecaseannual.FiledRequest{TaxpayerID: req.TaxpayerId.String(), RegimeCode: req.RegimeCode, FiscalYear: req.FiscalYear}
	view, err := h.deps.Filed.Get(ctx, userID, fr)
	if err != nil {
		return nil, err
	}
	wait, limited, err := h.spendRange(ctx, userID, true)
	switch {
	case err != nil:
		return nil, err
	case limited:
		return gen.GetFiledPayments429JSONResponse{TooManyRequestsJSONResponse: tooMany(ctx, wait)}, nil
	}
	body, err := h.filedBody(ctx, userID, fr, view)
	if err != nil {
		return nil, err
	}
	return gen.GetFiledPayments200JSONResponse{Body: body}, nil
}

// PutFiledPayments implements gen.StrictServerInterface.
func (h *Handler) PutFiledPayments(ctx context.Context, req gen.PutFiledPaymentsRequestObject) (gen.PutFiledPaymentsResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.PutFiledPayments401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if req.Body == nil {
		return nil, bodyRequired()
	}
	entries := make([]usecaseannual.FiledEntry, 0, len(req.Body.Months))
	for _, m := range req.Body.Months {
		entries = append(entries, usecaseannual.FiledEntry{Month: m.Month, Amount: cents(m.AmountCents)})
	}
	// The range's extra read is spent before the write: a 429 never
	// follows a committed replace of the months.
	wait, limited, err := h.spendRange(ctx, userID, true)
	switch {
	case err != nil:
		return nil, err
	case limited:
		return gen.PutFiledPayments429JSONResponse{TooManyRequestsJSONResponse: tooMany(ctx, wait)}, nil
	}
	fr := usecaseannual.FiledRequest{TaxpayerID: req.TaxpayerId.String(), RegimeCode: req.RegimeCode, FiscalYear: req.FiscalYear}
	view, err := h.deps.Filed.Put(ctx, userID, fr, entries)
	if err != nil {
		return nil, err
	}
	body, err := h.filedBody(ctx, userID, fr, view)
	if err != nil {
		return nil, err
	}
	return gen.PutFiledPayments200JSONResponse{Body: body}, nil
}

// filedBody joins the stored months with the estimate's columns; reading
// them runs the range, whose extra estimate read the caller has spent.
func (h *Handler) filedBody(ctx context.Context, userID string, fr usecaseannual.FiledRequest, view usecaseannual.FiledView) (
	gen.FiledPayments, error,
) {
	payments, err := h.deps.Estimates.Payments(ctx, userID, fr.TaxpayerID, view.Regime, view.FiscalYear)
	if err != nil {
		return gen.FiledPayments{}, err
	}
	return toFiled(view, payments), nil
}

// spendRange takes the extra estimate read of a range; limited reports a
// spent budget (429 with Retry-After).
func (h *Handler) spendRange(ctx context.Context, userID string, used bool) (time.Duration, bool, error) {
	if !used || h.deps.Range == nil {
		return 0, false, nil
	}
	allowed, wait, err := h.deps.Range.Take(ctx, userID)
	if err != nil {
		return 0, false, apperr.Wrap(err, apperr.CodeInternalError, msgInternal)
	}
	return wait, !allowed, nil
}

// caller returns the authenticated user's id.
func caller(ctx context.Context) (string, bool) {
	userID, err := principal.UserID(ctx)
	return userID, err == nil
}

// unauthorized is the contract's Unauthorized response.
func unauthorized(ctx context.Context) gen.UnauthorizedJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeUnauthorized
	body.Error.Message = msgAuthenticationRequired
	body.Error.RequestId = requestid.From(ctx)
	return gen.UnauthorizedJSONResponse{Body: body, Headers: gen.UnauthorizedResponseHeaders{WWWAuthenticate: bearerChallenge}}
}

// tooMany is the contract's TooManyRequests response.
func tooMany(ctx context.Context, wait time.Duration) gen.TooManyRequestsJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeRateLimited
	body.Error.Message = msgRateLimited
	body.Error.RequestId = requestid.From(ctx)
	return gen.TooManyRequestsJSONResponse{
		Body: body, Headers: gen.TooManyRequestsResponseHeaders{RetryAfter: ratelimit.RetryAfterSeconds(wait)},
	}
}

// bodyRequired backs up the router's validation of a missing body.
func bodyRequired() error {
	return apperr.New(apperr.CodeValidationFailed, "One or more fields are invalid.").
		WithFields(apperr.FieldError{Field: "body", Code: "required", Message: "Envía los datos."})
}
