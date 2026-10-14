package usecase

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Use case errors. Stores return the ones that describe stored state; the
// use cases return every client-facing failure as an *apperr.Error whose
// cause wraps one of these (or a domain error), so callers match with
// errors.Is and the HTTP edge answers the contract code as is.
var (
	// ErrTaxpayerNotFound: no such taxpayer for the caller (another user's
	// is the same answer: 404, never 403).
	ErrTaxpayerNotFound = errors.New("taxengine: taxpayer not found")
	// ErrRegimeNotHeld: the taxpayer never held the regime.
	ErrRegimeNotHeld = errors.New("taxengine: regime not held")
	// ErrRegimeNotActive: the taxpayer does not hold the regime in any
	// month of the fiscal year (422 regime_not_active).
	ErrRegimeNotActive = errors.New("taxengine: regime not active in the fiscal year")
	// ErrTaxpayerArchived: a write on an archived taxpayer (422).
	ErrTaxpayerArchived = errors.New("taxengine: taxpayer is archived")
	// ErrParametersUnavailable: a fiscal year before the first with tax
	// parameters (422 parameters_unavailable, OQ-T9).
	ErrParametersUnavailable = errors.New("taxengine: no tax parameters for the fiscal year")
	// ErrFutureFiscalYear: a fiscal year after the current one (400
	// out_of_range).
	ErrFutureFiscalYear = errors.New("taxengine: future fiscal year")
	// ErrFuturePeriod: a period after the current month (400 out_of_range).
	ErrFuturePeriod = errors.New("taxengine: future period")
)

// FieldError.code values of the tax engine (MODULE-tax-engine §5).
const (
	DetailInvalid               = "invalid"
	DetailOutOfRange            = "out_of_range"
	DetailUnknownInput          = "unknown_input"
	DetailTaxpayerArchived      = "taxpayer_archived"
	DetailRegimeNotActive       = "regime_not_active"
	DetailParametersUnavailable = "parameters_unavailable"
	fieldTaxpayerID             = "taxpayerId"
	fieldRegimeCode             = "regimeCode"
	fieldFiscalYear             = "fiscal_year"
	fieldPeriod                 = "period"
	fieldValues                 = "values"
)

// Client-safe messages: es-MX for the fields (contract examples), never an
// amount, a rate or another user's data.
const (
	msgInvalidInput          = "One or more fields are invalid." // the router's wording for every validation_failed
	msgNotFound              = "Not found."
	msgFutureYear            = "Ese ejercicio todavía no empieza."
	msgFuturePeriod          = "Ese mes todavía no empieza."
	msgInvalidPeriod         = "Usa un mes con el formato AAAA-MM."
	msgArchived              = "This taxpayer is archived. Unarchive it to make changes."
	msgArchivedField         = "Desarchiva el contribuyente primero."
	msgRegimeNotActive       = "The taxpayer does not hold this regime in that fiscal year."
	msgRegimeNotActiveField  = "No tienes este régimen en ese ejercicio."
	msgParametersUnavailable = "There are no tax parameters for that fiscal year."
	msgParametersField       = "Sólo podemos estimar desde 2026."
	msgUnknownInput          = "Este dato no aplica a este régimen."
	msgInvalidValue          = "Este valor no corresponde al tipo de dato."
	msgDuplicateValue        = "Este dato está repetido."
	msgOutOfRange            = "Este valor está fuera del rango permitido."
)

// notFound is the 404 of an id that names no taxpayer of the caller.
func notFound(cause error) error { return apperr.Wrap(cause, apperr.CodeNotFound, msgNotFound) }

// businessRule is a 422 business_rule_violation naming one field.
func businessRule(cause error, message, field, detail, fieldMessage string) error {
	return apperr.Wrap(cause, apperr.CodeBusinessRuleViolation, message).
		WithFields(apperr.FieldError{Field: field, Code: detail, Message: fieldMessage})
}

func archivedError() error {
	return businessRule(ErrTaxpayerArchived, msgArchived, fieldTaxpayerID, DetailTaxpayerArchived, msgArchivedField)
}

func regimeNotActiveError() error {
	return businessRule(ErrRegimeNotActive, msgRegimeNotActive, fieldRegimeCode, DetailRegimeNotActive, msgRegimeNotActiveField)
}

// parametersUnavailableError is the 422 of a fiscal year or period before
// the first published parameters, naming field.
func parametersUnavailableError(field string) error {
	return businessRule(ErrParametersUnavailable, msgParametersUnavailable, field, DetailParametersUnavailable, msgParametersField)
}

// invalidPeriodError is the 400 of a period that is not "YYYY-MM".
func invalidPeriodError(cause error) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: fieldPeriod, Code: DetailInvalid, Message: msgInvalidPeriod})
}

// futurePeriodError is the 400 of a month after the current one (§2.10).
func futurePeriodError() error {
	return apperr.Wrap(ErrFuturePeriod, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: fieldPeriod, Code: DetailOutOfRange, Message: msgFuturePeriod})
}

func futureYearError() error {
	return apperr.Wrap(ErrFutureFiscalYear, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: fieldFiscalYear, Code: DetailOutOfRange, Message: msgFutureYear})
}

// violationDetails names the FieldError of each value rule (§5).
var violationDetails = []struct {
	target  error
	code    string
	message string
}{
	{domain.ErrUnknownInput, DetailUnknownInput, msgUnknownInput},
	{domain.ErrInputOutOfRange, DetailOutOfRange, msgOutOfRange},
	{domain.ErrDuplicateTaxInput, DetailInvalid, msgDuplicateValue},
	{domain.ErrInvalidTaxInput, DetailInvalid, msgInvalidValue},
}

// valuesError is the 400 naming every refused value as values.<i>.<property>.
func valuesError(violations []domain.InputViolation) error {
	fields := make([]apperr.FieldError, 0, len(violations))
	causes := make([]error, 0, len(violations))
	for _, v := range violations {
		code, message := DetailInvalid, msgInvalidValue
		for _, d := range violationDetails {
			if errors.Is(v.Err, d.target) {
				code, message = d.code, d.message
				break
			}
		}
		fields = append(fields, apperr.FieldError{Field: v.Field(fieldValues), Code: code, Message: message})
		causes = append(causes, v)
	}
	return apperr.Wrap(errors.Join(causes...), apperr.CodeValidationFailed, msgInvalidInput).WithFields(fields...)
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
	case errors.Is(err, ErrTaxpayerNotFound):
		return notFound(err)
	case errors.Is(err, ErrRegimeNotHeld):
		return regimeNotActiveError()
	}
	return fmt.Errorf("taxengine: %s: %w", op, err)
}
