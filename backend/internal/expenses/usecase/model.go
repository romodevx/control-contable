package usecase

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Expense is an expense as the use cases return it (contract Expense).
type Expense struct {
	ID            string
	TaxpayerID    string
	Regime        fiscal.RegimeCode
	Category      domain.CategoryCode
	Amount        money.Cents
	IVARate       fiscal.IVARate
	IVA           money.Cents
	PaymentMethod domain.PaymentMethod
	PaidOn        fiscal.Date
	HasCFDI       bool
	CFDIUUID      fiscal.CFDIUUID // zero = null
	Supplier      domain.Supplier
	Withholding   fiscal.WithholdingRecord
	// RemitDueOn is when the taxpayer must remit what it withheld (zero
	// when nothing).
	RemitDueOn  fiscal.Date
	Description string // "" = null
	Deductible  bool
	Source      domain.DeductibleSource
	// Warnings are computed on writes (create, update, restore), never
	// stored; reads carry none.
	Warnings  []domain.Warning
	Legend    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateInput is a create request after the adapter decoded it. user_id
// comes from the session and taxpayer_id from the path, never the body.
type CreateInput struct {
	UserID     string
	TaxpayerID string
	// ID is the client's idempotency key; "" lets the server choose one.
	ID          string
	Regime      fiscal.RegimeCode
	Draft       domain.Draft
	Description string // raw: trimmed and NFC-normalized here
}

// CreateResult is the stored expense and whether this call created it
// (201) or found the same request already stored (200).
type CreateResult struct {
	Expense Expense
	Created bool
}

// Patch is a merge-patch of an expense: a nil field is unchanged. A
// non-nil zero CFDIUUID or empty Description clears the field; a non-nil
// UseDefault clears the deductible override.
type Patch struct {
	Regime        *fiscal.RegimeCode
	Category      *domain.CategoryCode
	Amount        *money.Cents
	IVARate       *fiscal.IVARate
	IVA           *money.Cents
	PaymentMethod *domain.PaymentMethod
	PaidOn        *fiscal.Date
	HasCFDI       *bool
	CFDIUUID      *fiscal.CFDIUUID
	Deductible    *domain.Override
	// The supplier facts; moving SupplierKind away from pf_with_activity
	// clears the stored regime unless the patch sets one.
	SupplierKind    *domain.SupplierKind
	SupplierRegime  *domain.SupplierRegime
	OperationKind   *fiscal.OperationKind
	ResidentialRent *bool
	Description     *string
}

// UpdateInput is an edit request.
type UpdateInput struct {
	UserID     string
	TaxpayerID string
	ExpenseID  string
	Patch      Patch
}

// ListInput is a list request. Period is "YYYY-MM"; Regime "" lists every
// regime; Cursor "" is the first page; Limit 0 is the default.
type ListInput struct {
	UserID     string
	TaxpayerID string
	Period     string
	Regime     fiscal.RegimeCode
	Cursor     string
	Limit      int
}

// Page is one page of a month's expenses (contract ExpensePage).
type Page struct {
	Items []Expense
	// NextCursor is the opaque cursor of the next page; "" on the last.
	NextCursor string
	// Totals cover the whole filtered period, not the page.
	Totals Totals
	Legend string
}

// CategoriesInput asks for the categories of a regime the taxpayer holds on
// On ("YYYY-MM-DD"; "" = today in Mexico City).
type CategoriesInput struct {
	UserID     string
	TaxpayerID string
	Regime     fiscal.RegimeCode
	On         string
}

// Category is one category resolved for the taxpayer's regime and options
// (contract ExpenseCategory).
type Category struct {
	Code              domain.CategoryCode
	Name              string
	Deductibility     domain.Deductibility
	DefaultDeductible bool
	Note              string // "" = null
	LegalBasis        string
	// DefaultOperation is the E3 default of a persona moral's supplier
	// (category_regime_rules.default_operation_kind); "" for a persona
	// física.
	DefaultOperation fiscal.OperationKind
}

// CategoryList is the form's category catalog (contract ExpenseCategoryList).
type CategoryList struct {
	Items            []Category
	IVARatesAllowed  []fiscal.IVARate
	CashPaymentLimit money.Cents
	Legend           string
}

// PreviewInput is previewExpenseWithholdings: the facts of an expense the
// form is still editing. On is the payment date ("YYYY-MM-DD"; "" =
// today in Mexico City).
type PreviewInput struct {
	UserID     string
	TaxpayerID string
	Regime     fiscal.RegimeCode
	On         string
	Category   domain.CategoryCode
	Amount     money.Cents
	IVARate    fiscal.IVARate
	IVA        money.Cents
	Supplier   domain.Supplier
}

// Preview is what the matching create would derive; nothing is stored.
// Warnings are the withholding warnings only.
type Preview struct {
	Withholding fiscal.WithholdingRecord
	RemitDueOn  fiscal.Date
	Subtotal    money.Cents
	IVA         money.Cents
	Warnings    []domain.Warning
}
