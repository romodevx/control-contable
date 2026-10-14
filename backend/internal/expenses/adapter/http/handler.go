// Package expenseshttp is the expenses module's HTTP adapter: it translates
// between the contract (docs/api/openapi.yaml) and the expense use cases.
package expenseshttp

import (
	"context"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

const (
	// bearerChallenge is the contract's fixed WWW-Authenticate value
	// (components.headers.WWWAuthenticate).
	bearerChallenge = `Bearer realm="control-contable"` //nolint:gosec // G101: an auth challenge, not a credential
	// msgAuthenticationRequired is the contract's Unauthorized message.
	msgAuthenticationRequired = "Authentication required."
)

// ExpenseManager is the expense use cases (usecase.Expenses, M4-05/M4-06).
type ExpenseManager interface {
	Create(ctx context.Context, in usecase.CreateInput) (usecase.CreateResult, error)
	Update(ctx context.Context, in usecase.UpdateInput) (usecase.Expense, error)
	Get(ctx context.Context, userID, taxpayerID, expenseID string) (usecase.Expense, error)
	List(ctx context.Context, in usecase.ListInput) (usecase.Page, error)
	Delete(ctx context.Context, userID, taxpayerID, expenseID string) error
	Restore(ctx context.Context, userID, taxpayerID, expenseID string) (usecase.Expense, error)
	Categories(ctx context.Context, in usecase.CategoriesInput) (usecase.CategoryList, error)
}

// Handler implements the module_expenses operations of
// gen.StrictServerInterface (M4-07). cmd/api embeds it in the API server.
//
// Every operation is secured: the user id comes from the principal the
// router's authenticator verified and the taxpayer id from the path, never
// from the body. Use case errors are client-ready *apperr.Error values
// (402 naming the expenses cap, 409 id_reused, 422 regime_not_active, 404
// across users, 400 with field details) and are returned as is; the router
// writes the envelope, and any other failure becomes the generic 500.
type Handler struct {
	expenses ExpenseManager
}

// NewHandler returns the handler over the use cases.
func NewHandler(expenses ExpenseManager) *Handler {
	return &Handler{expenses: expenses}
}

// ListExpenseCategories implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/expense-categories).
func (h *Handler) ListExpenseCategories(ctx context.Context, req gen.ListExpenseCategoriesRequestObject) (gen.ListExpenseCategoriesResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListExpenseCategories401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	in := usecase.CategoriesInput{UserID: userID, TaxpayerID: req.TaxpayerId.String(), Regime: regimeOf(req.Params.Regime)}
	if req.Params.On != nil {
		in.On = req.Params.On.String()
	}
	list, err := h.expenses.Categories(ctx, in)
	if err != nil {
		return nil, err
	}
	return gen.ListExpenseCategories200JSONResponse{
		Body: toCategoryList(list),
	}, nil
}

// ListExpenses implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/expenses).
func (h *Handler) ListExpenses(ctx context.Context, req gen.ListExpensesRequestObject) (gen.ListExpensesResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.ListExpenses401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	page, err := h.expenses.List(ctx, toListInput(userID, req))
	if err != nil {
		return nil, err
	}
	body, err := toPage(page)
	if err != nil {
		return nil, err
	}
	return gen.ListExpenses200JSONResponse{Body: body}, nil
}

// CreateExpense implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/expenses): 201 when created, 200 for an
// idempotent retry of a client id the caller already owns.
func (h *Handler) CreateExpense(ctx context.Context, req gen.CreateExpenseRequestObject) (gen.CreateExpenseResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.CreateExpense401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	in, err := toCreateInput(userID, req.TaxpayerId.String(), req.Body)
	if err != nil {
		return nil, err
	}
	result, err := h.expenses.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	body, err := toExpense(result.Expense)
	if err != nil {
		return nil, err
	}
	if !result.Created {
		return gen.CreateExpense200JSONResponse{Body: body}, nil
	}
	return gen.CreateExpense201JSONResponse{Body: body}, nil
}

// GetExpense implements gen.StrictServerInterface
// (GET /v1/taxpayers/{taxpayerId}/expenses/{expenseId}).
func (h *Handler) GetExpense(ctx context.Context, req gen.GetExpenseRequestObject) (gen.GetExpenseResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.GetExpense401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	e, err := h.expenses.Get(ctx, userID, req.TaxpayerId.String(), req.ExpenseId.String())
	if err != nil {
		return nil, err
	}
	body, err := toExpense(e)
	if err != nil {
		return nil, err
	}
	return gen.GetExpense200JSONResponse{Body: body}, nil
}

// UpdateExpense implements gen.StrictServerInterface
// (PATCH /v1/taxpayers/{taxpayerId}/expenses/{expenseId}): merge-patch,
// where null clears cfdi_uuid, description or the deductible override.
func (h *Handler) UpdateExpense(ctx context.Context, req gen.UpdateExpenseRequestObject) (gen.UpdateExpenseResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.UpdateExpense401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	patch, err := toPatch(req.Body)
	if err != nil {
		return nil, err
	}
	e, err := h.expenses.Update(ctx, usecase.UpdateInput{
		UserID: userID, TaxpayerID: req.TaxpayerId.String(), ExpenseID: req.ExpenseId.String(), Patch: patch,
	})
	if err != nil {
		return nil, err
	}
	body, err := toExpense(e)
	if err != nil {
		return nil, err
	}
	return gen.UpdateExpense200JSONResponse{Body: body}, nil
}

// DeleteExpense implements gen.StrictServerInterface
// (DELETE /v1/taxpayers/{taxpayerId}/expenses/{expenseId}): a soft delete,
// 204 also when already deleted.
func (h *Handler) DeleteExpense(ctx context.Context, req gen.DeleteExpenseRequestObject) (gen.DeleteExpenseResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.DeleteExpense401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	if err := h.expenses.Delete(ctx, userID, req.TaxpayerId.String(), req.ExpenseId.String()); err != nil {
		return nil, err
	}
	return gen.DeleteExpense204Response{}, nil
}

// RestoreExpense implements gen.StrictServerInterface
// (POST /v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore): Undo,
// never spending a trial slot.
func (h *Handler) RestoreExpense(ctx context.Context, req gen.RestoreExpenseRequestObject) (gen.RestoreExpenseResponseObject, error) {
	userID, ok := caller(ctx)
	if !ok {
		return gen.RestoreExpense401JSONResponse{UnauthorizedJSONResponse: unauthorized(ctx)}, nil
	}
	e, err := h.expenses.Restore(ctx, userID, req.TaxpayerId.String(), req.ExpenseId.String())
	if err != nil {
		return nil, err
	}
	body, err := toExpense(e)
	if err != nil {
		return nil, err
	}
	return gen.RestoreExpense200JSONResponse{Body: body}, nil
}

// caller returns the authenticated user's id.
func caller(ctx context.Context) (string, bool) {
	userID, err := principal.UserID(ctx)
	return userID, err == nil
}

// unauthorized is the contract's Unauthorized response: the generic 401
// with the bearer challenge.
func unauthorized(ctx context.Context) gen.UnauthorizedJSONResponse {
	var body gen.ErrorResponse
	body.Error.Code = gen.ErrorCodeUnauthorized
	body.Error.Message = msgAuthenticationRequired
	body.Error.RequestId = requestid.From(ctx)
	return gen.UnauthorizedJSONResponse{Body: body, Headers: gen.UnauthorizedResponseHeaders{WWWAuthenticate: bearerChallenge}}
}
