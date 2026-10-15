package usecase

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// Use case errors. Stores return the ones that describe stored state; the
// use cases return every client-facing failure as an *apperr.Error whose
// cause wraps one of these (or a fiscal domain error), so callers match
// with errors.Is and the HTTP edge (M4-07) answers the contract code as is.
var (
	// ErrExpenseNotFound: no such expense for the caller and taxpayer (404;
	// another user's is the same answer, never 403).
	ErrExpenseNotFound = errors.New("expenses: expense not found")
	// ErrIDConflict: an id taken outside the caller's scope (409 without a
	// detail: RLS hides the row, nothing about its owner leaks).
	ErrIDConflict = errors.New("expenses: id conflict")
	// ErrIDReused: the caller reused an id with other content, for another
	// taxpayer, or of a deleted expense (409 id_reused).
	ErrIDReused = errors.New("expenses: id reused")
	// ErrCategoryNotFound: the catalog has no such category (400
	// unknown_category).
	ErrCategoryNotFound = errors.New("expenses: category not found")
	// ErrParametersUnavailable: no tax parameter set covers the date (400
	// out_of_range on the date field).
	ErrParametersUnavailable = errors.New("expenses: no tax parameters for the date")
	// ErrTaxpayerArchived: a write on an archived taxpayer (422).
	ErrTaxpayerArchived = errors.New("expenses: taxpayer is archived")
	// ErrRegimeNotActive: the taxpayer does not hold the regime on the
	// categories' date (422).
	ErrRegimeNotActive = errors.New("expenses: regime not active on date")
)

// FieldError.code values the use cases add to the fiscal kernel's.
const (
	DetailIDReused         = "id_reused"
	DetailTaxpayerArchived = "taxpayer_archived"
	DetailRegimeNotActive  = "regime_not_active"
	DetailTooLong          = "too_long"
)

// Contract field names the use cases report (the domain names the rest).
const (
	fieldID          = "id"
	fieldTaxpayerID  = "taxpayerId"
	fieldRegimeCode  = "regime_code"
	fieldDescription = "description"
	fieldPeriod      = "period"
	fieldCursor      = "cursor"
	fieldLimit       = "limit"
	fieldRegime      = "regime" // the categories and list query parameter
	fieldOn          = "on"
)

// Client-safe messages: they never echo an amount, a description or a folio.
const (
	msgInvalidInput  = "One or more fields are invalid." // the router's wording for every validation_failed
	msgBusinessRule  = "The expense breaks a business rule."
	msgNotFound      = "The expense was not found."
	msgTaxpayerGone  = "The taxpayer was not found."
	msgIDConflict    = "The request conflicts with the current state."
	msgIDReused      = "This id was already used for another expense."
	msgArchived      = "This taxpayer is archived. Unarchive it to make changes."
	msgArchivedField = "Unarchive the taxpayer first."
	msgNotActive     = "The taxpayer does not hold this regime on that date."
	msgFieldInvalid  = "This value is not valid."
)

// detailMessages are the field messages of each details[].code.
var detailMessages = map[string]string{
	fiscal.ErrRequired.Detail():              "This field is required.",
	fiscal.ErrInvalid.Detail():               msgFieldInvalid,
	fiscal.ErrMustBePositive.Detail():        "This amount must be greater than zero.",
	fiscal.ErrOutOfRange.Detail():            "This value is out of range.",
	fiscal.ErrRateNotValidOnDate.Detail():    "This IVA rate is not valid on the payment date.",
	fiscal.ErrIVAAmountMismatch.Detail():     "The IVA does not match the rate.",
	domain.ErrSupplierNotApplicable.Detail(): "Only a persona moral withholds from its suppliers.",
	domain.ErrResidentialRentExempt.Detail(): "Housing rent is exempt from IVA.",
	"unknown_category":                       "This category does not exist.",
	"has_cfdi_mismatch":                      "A folio fiscal needs has_cfdi to be true.",
	DetailTooLong:                            "This text is too long.",
}

func detailMessage(code string) string {
	if m, ok := detailMessages[code]; ok {
		return m
	}
	return msgFieldInvalid
}

// violationsError turns the domain's violations into the client error of
// their class: validation_failed (400) or business_rule_violation (422).
func violationsError(vs fiscal.Violations) error {
	code, message := apperr.CodeValidationFailed, msgInvalidInput
	switch vs.Class() {
	case fiscal.ClassValidation:
	case fiscal.ClassBusinessRule:
		code, message = apperr.CodeBusinessRuleViolation, msgBusinessRule
	default:
		return fmt.Errorf("expenses: unclassified violations: %w", vs)
	}
	fields := make([]apperr.FieldError, 0, len(vs))
	for _, v := range vs {
		fields = append(fields, apperr.FieldError{Field: v.Field, Code: v.Err.Detail(), Message: detailMessage(v.Err.Detail())})
	}
	return apperr.Wrap(vs, code, message).WithFields(fields...)
}

// invalidField is a 400 naming one field with one fiscal error.
func invalidField(field string, err fiscal.Error) error {
	return violationsError(fiscal.Violations{{Field: field, Err: err}})
}

// notFound is the 404 of an expense id that names no expense of the caller
// under the path's taxpayer.
func notFound() error { return apperr.Wrap(ErrExpenseNotFound, apperr.CodeNotFound, msgNotFound) }

// taxpayerNotFound is the 404 of a malformed taxpayer id.
func taxpayerNotFound() error { return apperr.New(apperr.CodeNotFound, msgTaxpayerGone) }

// archivedError is the 422 for a write on an archived taxpayer.
func archivedError() error {
	return apperr.Wrap(ErrTaxpayerArchived, apperr.CodeBusinessRuleViolation, msgArchived).
		WithFields(apperr.FieldError{Field: fieldTaxpayerID, Code: DetailTaxpayerArchived, Message: msgArchivedField})
}

// idReusedError is the 409 of a caller's id reused with other content.
func idReusedError() error {
	return apperr.Wrap(ErrIDReused, apperr.CodeConflict, msgIDConflict).
		WithFields(apperr.FieldError{Field: fieldID, Code: DetailIDReused, Message: msgIDReused})
}

// regimeNotActiveError is the 422 of a categories request for a regime the
// taxpayer does not hold on the date.
func regimeNotActiveError() error {
	return apperr.Wrap(ErrRegimeNotActive, apperr.CodeBusinessRuleViolation, msgNotActive).
		WithFields(apperr.FieldError{Field: fieldRegime, Code: DetailRegimeNotActive, Message: msgNotActive})
}

// present turns a use case failure into its client form: an *apperr.Error
// passes as is, a known sentinel gets its contract code, and anything else
// is an infrastructure failure wrapped with op (500).
func present(op string, err error) error {
	if err == nil {
		return nil
	}
	err = unjoin(err)
	if _, ok := errors.AsType[*apperr.Error](err); ok {
		return err
	}
	switch {
	case errors.Is(err, ErrExpenseNotFound):
		return notFound()
	case errors.Is(err, ErrIDConflict):
		return apperr.Wrap(err, apperr.CodeConflict, msgIDConflict)
	}
	return fmt.Errorf("expenses: %s: %w", op, err)
}

// unjoin returns the work's own error when the unit of work joined it with
// nothing else (tx.Manager joins the rollback's error, nil when the
// rollback succeeded), so the edge gets the *apperr.Error itself. A real
// rollback failure stays joined: the client error is still found with
// errors.As, and the failure reaches the logs.
func unjoin(err error) error {
	if _, isViolations := err.(fiscal.Violations); isViolations { //nolint:errorlint // only the outer error is unjoined
		return err
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		if errs := joined.Unwrap(); len(errs) == 1 {
			return errs[0]
		}
	}
	return err
}
