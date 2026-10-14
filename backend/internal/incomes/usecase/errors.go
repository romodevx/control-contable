package usecase

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// Use case errors. Stores return ErrIncomeNotFound and ErrIDTaken; the use
// cases return every client failure as an *apperr.Error whose cause wraps
// one of these (or a fiscal error), so callers match with errors.Is.
var (
	// ErrIncomeNotFound: no such income for the caller and taxpayer, or it
	// is soft-deleted where only a live one counts (404).
	ErrIncomeNotFound = errors.New("incomes: income not found")
	// ErrIDTaken: the store found the id already in use. Behind RLS that
	// is another user's income: 409 without details.
	ErrIDTaken = errors.New("incomes: id taken")
	// ErrIDReused: the caller's own id with other fields, another
	// taxpayer or a deleted income (409 id_reused).
	ErrIDReused = errors.New("incomes: id reused")
	// ErrTaxpayerArchived: a write on an archived taxpayer (422).
	ErrTaxpayerArchived = errors.New("incomes: taxpayer is archived")
	// ErrRegimeNotHeld: the taxpayer never held the income's regime (422
	// regime_not_active, as CheckRegime answers a window miss).
	ErrRegimeNotHeld = errors.New("incomes: regime not held")
	// ErrInvalidID: a client id that is not a UUID (400).
	ErrInvalidID = errors.New("incomes: invalid id")
	// ErrInvalidPeriod: a list period that is not YYYY-MM (400).
	ErrInvalidPeriod = errors.New("incomes: invalid period")
	// ErrInvalidLimit: a page size outside [1, MaxLimit] (400).
	ErrInvalidLimit = errors.New("incomes: invalid limit")
	// ErrInvalidCursor: a cursor this server did not issue (400).
	ErrInvalidCursor = errors.New("incomes: invalid cursor")
)

// ErrTooLong reports a description over maxDescription characters after
// normalization (validation_failed).
var ErrTooLong = fiscal.ValidationError("too_long")

// Contract field names the use cases report beyond the domain's.
const (
	fieldID          = "id"
	fieldTaxpayerID  = "taxpayerId"
	fieldRegimeCode  = "regime_code"
	fieldDescription = "description"
	fieldPeriod      = "period"
	fieldLimit       = "limit"
	fieldCursor      = "cursor"
)

// Detail codes the use cases report beyond the domain's.
const (
	detailIDReused         = "id_reused"
	detailTaxpayerArchived = "taxpayer_archived"
	detailRegimeNotActive  = "regime_not_active"
)

// Client-safe messages. They never echo an amount, a description or a UUID
// (S-5).
const (
	msgInvalidInput  = "One or more fields are invalid."
	msgBusinessRule  = "The income breaks a business rule."
	msgNotFound      = "The income was not found."
	msgTaxpayerNF    = "The taxpayer was not found."
	msgConflict      = "The request conflicts with the current state."
	msgIDReused      = "This id was already used with other data."
	msgArchived      = "This taxpayer is archived. Unarchive it to make changes."
	msgArchivedField = "Unarchive the taxpayer first."
	msgNotActive     = "The taxpayer does not hold this regime on the income's dates."
	msgInvalidField  = "This value is not valid."
	msgRequired      = "This field is required."
	msgLimit         = "Choose a page size from 1 to 100."
)

// detailMessages are the client messages of the detail codes.
var detailMessages = map[string]string{
	fiscal.ErrRequired.Detail():                 msgRequired,
	fiscal.ErrInvalid.Detail():                  msgInvalidField,
	fiscal.ErrMustBePositive.Detail():           "The amount must be greater than zero.",
	fiscal.ErrOutOfRange.Detail():               "The date cannot be in the future nor before 2022.",
	ErrTooLong.Detail():                         "This text is too long.",
	domain.ErrNotForRegime.Detail():             "This field only applies to another regime.",
	domain.ErrActivityNotHeld.Detail():          "The taxpayer does not hold this activity.",
	domain.ErrParametersUnavailable.Detail():    "There are no tax parameters for this date yet.",
	fiscal.ErrRateNotValidOnDate.Detail():       "This IVA rate is not valid on the income's date.",
	fiscal.ErrIVAAmountMismatch.Detail():        "The IVA does not match the rate.",
	fiscal.ErrWithholdingNotApplicable.Detail(): "This payer does not withhold this tax.",
	fiscal.ErrWithholdingExceedsBase.Detail():   "The withholding exceeds its base.",
}

// classCodes map a fiscal error class to its contract code and message.
var classCodes = map[fiscal.Class]struct {
	code    apperr.Code
	message string
}{
	fiscal.ClassValidation:   {apperr.CodeValidationFailed, msgInvalidInput},
	fiscal.ClassBusinessRule: {apperr.CodeBusinessRuleViolation, msgBusinessRule},
}

// present turns a use case error into its client form: an *apperr.Error
// passes as is, fiscal violations and known sentinels get their contract
// code, anything else is an infrastructure failure wrapped with op (500).
func present(op string, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*apperr.Error](err); ok {
		return err
	}
	if vs, ok := errors.AsType[fiscal.Violations](err); ok {
		if class, known := classCodes[vs.Class()]; known {
			return apperr.Wrap(err, class.code, class.message).WithFields(fieldsOf(vs)...)
		}
	}
	switch {
	case errors.Is(err, ErrIncomeNotFound):
		return apperr.Wrap(err, apperr.CodeNotFound, msgNotFound)
	case errors.Is(err, ErrIDReused):
		return apperr.Wrap(err, apperr.CodeConflict, msgConflict).
			WithFields(apperr.FieldError{Field: fieldID, Code: detailIDReused, Message: msgIDReused})
	case errors.Is(err, ErrIDTaken):
		return apperr.Wrap(err, apperr.CodeConflict, msgConflict)
	}
	return fmt.Errorf("incomes: %s: %w", op, err)
}

func fieldsOf(vs fiscal.Violations) []apperr.FieldError {
	fields := make([]apperr.FieldError, len(vs))
	for i, v := range vs {
		message, ok := detailMessages[v.Err.Detail()]
		if !ok {
			message = msgInvalidField
		}
		fields[i] = apperr.FieldError{Field: v.Field, Code: v.Err.Detail(), Message: message}
	}
	return fields
}

// fieldError is a 400 validation_failed naming one field.
func fieldError(cause error, field string, detail fiscal.Error, message string) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: detail.Detail(), Message: message})
}

// archivedError is the 422 for a write on an archived taxpayer.
func archivedError() error {
	return apperr.Wrap(ErrTaxpayerArchived, apperr.CodeBusinessRuleViolation, msgArchived).
		WithFields(apperr.FieldError{Field: fieldTaxpayerID, Code: detailTaxpayerArchived, Message: msgArchivedField})
}

// regimeNotHeldError is the 422 for a regime the taxpayer never held.
func regimeNotHeldError() error {
	return apperr.Wrap(ErrRegimeNotHeld, apperr.CodeBusinessRuleViolation, msgNotActive).
		WithFields(apperr.FieldError{Field: fieldRegimeCode, Code: detailRegimeNotActive, Message: msgNotActive})
}

// taxpayerNotFound is the 404 of a malformed taxpayer path id (QA O3).
func taxpayerNotFound() error { return apperr.New(apperr.CodeNotFound, msgTaxpayerNF) }

// incomeNotFound is the 404 of an income the caller cannot see live.
func incomeNotFound() error { return apperr.Wrap(ErrIncomeNotFound, apperr.CodeNotFound, msgNotFound) }
