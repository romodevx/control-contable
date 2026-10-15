package annual

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

// Use case errors. Stores return the ones that describe stored state; the
// use cases return every client-facing failure as an *apperr.Error whose
// cause wraps one of these, so callers match with errors.Is and the HTTP
// edge (ANN-15) answers the contract code as is.
var (
	// ErrTaxpayerNotFound: no such taxpayer for the caller (another user's
	// is the same answer: 404, never 403).
	ErrTaxpayerNotFound = errors.New("annual: taxpayer not found")
	// ErrRegimeNotHeld: the taxpayer never held the regime.
	ErrRegimeNotHeld = errors.New("annual: regime not held")
	// ErrDeductionNotFound: no such personal deduction for the caller.
	ErrDeductionNotFound = errors.New("annual: personal deduction not found")
	// ErrStatementNotFound: no such salary constancia for the caller.
	ErrStatementNotFound = errors.New("annual: salary statement not found")
	// ErrIDTaken: the store found the id already in use. Behind RLS that is
	// another user's row: 409 without details.
	ErrIDTaken = errors.New("annual: id taken")
	// ErrIDReused: the caller's own id with other fields, another taxpayer
	// or a deleted row (409 id_reused).
	ErrIDReused = errors.New("annual: id reused")
	// ErrInvalidField: a field rule refused the body (400).
	ErrInvalidField = errors.New("annual: invalid field")
	// ErrFutureFiscalYear: a fiscal year after the current one (400
	// out_of_range).
	ErrFutureFiscalYear = errors.New("annual: future fiscal year")
	// ErrTaxpayerArchived: a write on an archived taxpayer (422).
	ErrTaxpayerArchived = errors.New("annual: taxpayer is archived")
	// ErrPersonTypeNotAllowed: a personas físicas' input on a persona
	// moral (422 person_type_not_allowed).
	ErrPersonTypeNotAllowed = errors.New("annual: person type not allowed")
	// ErrParametersUnavailable: a fiscal year before the first with
	// parameters (422 parameters_unavailable).
	ErrParametersUnavailable = errors.New("annual: no tax parameters for the fiscal year")
	// ErrRegimeNotActive: the regime is not held in the year or month (422).
	ErrRegimeNotActive = errors.New("annual: regime not active")
	// ErrPaymentsNotApplicable: a held regime without provisional ISR (422).
	ErrPaymentsNotApplicable = errors.New("annual: regime has no provisional payments")
	// ErrLimitReached: the year's bounded list is full (422 limit_reached).
	ErrLimitReached = errors.New("annual: list bound reached")
)

// FieldError.code values of the annual inputs (MODULE-annual §5, contract
// 0.11.x).
const (
	DetailInvalid = "invalid"
	// DetailInvisibleCharacters: a label holding a control, format (zero
	// width, bidi) or line separator character (usertext, ANN-25).
	DetailInvisibleCharacters   = usertext.DetailInvisibleCharacters
	DetailRequired              = "required"
	DetailNotApplicable         = "not_applicable"
	DetailOutOfRange            = "out_of_range"
	DetailIDReused              = "id_reused"
	DetailTaxpayerArchived      = "taxpayer_archived"
	DetailPersonTypeNotAllowed  = "person_type_not_allowed"
	DetailParametersUnavailable = "parameters_unavailable"
	DetailRegimeNotActive       = "regime_not_active"
	DetailPaymentsNotApplicable = "payments_not_applicable"
	DetailLimitReached          = "limit_reached"
)

// Contract field names.
const (
	FieldID              = "id"
	FieldTaxpayerID      = "taxpayerId"
	FieldRegimeCode      = "regimeCode"
	FieldFiscalYearPath  = "fiscalYear"
	FieldFiscalYear      = "fiscal_year"
	FieldBody            = "body"
	FieldCategory        = "category"
	FieldAmount          = "amount_cents"
	FieldPaidOn          = "paid_on"
	FieldPaymentMethod   = "payment_method"
	FieldBeneficiary     = "beneficiary"
	FieldDoneeType       = "donee_type"
	FieldStudentLabel    = "student_label"
	FieldTuitionLevel    = "tuition_level"
	FieldScholarship     = "scholarship_cents"
	FieldCFDIUUID        = "cfdi_uuid"
	FieldEmployerLabel   = "employer_label"
	fieldMonths          = "months"
	fieldMonthProperty   = "month"
	fieldAmountProperty  = "amount_cents"
	fieldIndexSeparation = "."
)

// Client-safe messages: never an amount, a label or another user's data.
const (
	msgInvalidInput          = "One or more fields are invalid." // the router's wording for every validation_failed
	msgNotFound              = "Not found."
	msgConflict              = "The request conflicts with the current state."
	msgIDReused              = "This id was already used with other data."
	msgBusinessRule          = "The request breaks a business rule."
	msgArchived              = "This taxpayer is archived. Unarchive it to make changes."
	msgArchivedField         = "Desarchiva el contribuyente primero."
	msgPersonType            = "Este dato sólo aplica a personas físicas."
	msgParametersField       = "Sólo podemos estimar desde 2026."
	msgFutureYear            = "Ese ejercicio todavía no empieza."
	msgRegimeNotActive       = "No tienes este régimen en ese ejercicio."
	msgRegimeNotActiveMonth  = "No tenías este régimen en ese mes."
	msgPaymentsNotApplicable = "Este régimen no hace pagos provisionales de ISR."
	msgLimitDeductions       = "Ya registraste el máximo de 500 deducciones en ese ejercicio."
	msgLimitStatements       = "Ya registraste el máximo de 20 constancias en ese ejercicio."
	msgInvalidValue          = "Este valor no es válido."
	msgInvisibleCharacters   = "Quita los caracteres invisibles o de formato."
	msgRequired              = "Este dato es obligatorio para esa categoría."
	msgNotApplicable         = "Este dato no aplica a esa categoría."
	msgDateRange             = "La fecha no puede ser futura ni anterior a 2026."
	msgMonthNotClosed        = "Ese mes todavía no termina."
	msgMonthRepeated         = "Este mes está repetido."
)

// notFound is the 404 of an id that names nothing of the caller.
func notFound(cause error) error { return apperr.Wrap(cause, apperr.CodeNotFound, msgNotFound) }

// businessRule is a 422 business_rule_violation naming one field.
func businessRule(cause error, field, detail, fieldMessage string) error {
	return apperr.Wrap(cause, apperr.CodeBusinessRuleViolation, msgBusinessRule).
		WithFields(apperr.FieldError{Field: field, Code: detail, Message: fieldMessage})
}

func archivedError() error {
	return businessRule(ErrTaxpayerArchived, FieldTaxpayerID, DetailTaxpayerArchived, msgArchivedField)
}

func personTypeError() error {
	return businessRule(ErrPersonTypeNotAllowed, FieldTaxpayerID, DetailPersonTypeNotAllowed, msgPersonType)
}

func parametersUnavailableError(field string) error {
	return businessRule(ErrParametersUnavailable, field, DetailParametersUnavailable, msgParametersField)
}

func regimeNotActiveError(field, message string) error {
	return businessRule(ErrRegimeNotActive, field, DetailRegimeNotActive, message)
}

func paymentsNotApplicableError() error {
	return businessRule(ErrPaymentsNotApplicable, FieldRegimeCode, DetailPaymentsNotApplicable, msgPaymentsNotApplicable)
}

func limitError(message string) error {
	return businessRule(ErrLimitReached, FieldBody, DetailLimitReached, message)
}

func futureYearError(field string) error {
	return apperr.Wrap(ErrFutureFiscalYear, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: DetailOutOfRange, Message: msgFutureYear})
}

// fieldErrors is the 400 naming every refused field.
func fieldErrors(fields []apperr.FieldError) error {
	return apperr.Wrap(ErrInvalidField, apperr.CodeValidationFailed, msgInvalidInput).WithFields(fields...)
}

// present turns the error of a use case into its client form: an
// *apperr.Error passes as is, a store sentinel gets its contract code, and
// anything else is an infrastructure failure wrapped with op (500).
func present(op string, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*apperr.Error](err); ok {
		return err
	}
	switch {
	case errors.Is(err, ErrTaxpayerNotFound), errors.Is(err, ErrDeductionNotFound), errors.Is(err, ErrStatementNotFound):
		return notFound(err)
	case errors.Is(err, ErrRegimeNotHeld):
		return regimeNotActiveError(FieldRegimeCode, msgRegimeNotActive)
	case errors.Is(err, ErrIDReused):
		return apperr.Wrap(err, apperr.CodeConflict, msgConflict).
			WithFields(apperr.FieldError{Field: FieldID, Code: DetailIDReused, Message: msgIDReused})
	case errors.Is(err, ErrIDTaken):
		return apperr.Wrap(err, apperr.CodeConflict, msgConflict)
	}
	return fmt.Errorf("annual: %s: %w", op, err)
}
