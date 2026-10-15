package usecase

import (
	"errors"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

var (
	// ErrMissingDependency reports a Service wired without one of its ports.
	ErrMissingDependency = errors.New("history: missing dependency")
	// ErrTaxpayerNotFound reports a taxpayer that is unknown or another
	// user's (404); ViewStore.SaveView returns it for the foreign key.
	ErrTaxpayerNotFound = errors.New("history: taxpayer not found")
	// ErrFuture reports a fiscal year or period after the current one (400
	// out_of_range).
	ErrFuture = errors.New("history: future year or period")
	// ErrParametersUnavailable reports a year or period before 2026 (422).
	ErrParametersUnavailable = errors.New("history: before the first fiscal year")
	// ErrEstimateCoverage reports a range estimate that does not answer one
	// month per month asked for (a wiring bug, 500).
	ErrEstimateCoverage = errors.New("history: the range estimate does not cover the months")
)

// FieldError.code values beyond the JSON Schema keywords (contract
// getHistory, markHistoryViewed; the router answers the schema ones first,
// these are the backstop).
const (
	DetailPattern               = "pattern"
	DetailOutOfRange            = "out_of_range"
	DetailParametersUnavailable = "parameters_unavailable"
	fieldFiscalYear             = "fiscal_year"
	fieldPeriod                 = "period"
)

// Client-safe messages (contract examples, es-MX for the fields). Values
// the client sent are never echoed.
const (
	msgInvalidInput      = "One or more fields are invalid." // the router's wording for every validation_failed
	msgNotFound          = "Not found."
	msgInvalidFormat     = "Formato no válido."
	msgPeriodPattern     = "Usa el formato AAAA-MM."
	msgFutureYear        = "Aún no hay datos de un año futuro."
	msgFuturePeriod      = "Aún no hay datos de un mes futuro."
	msgNoYearData        = "There is no data before 2026."
	msgNoYearDataField   = "Sólo hay datos desde 2026."
	msgNoPeriodData      = "There is no data before January 2026."
	msgNoPeriodDataField = "Sólo hay datos desde enero de 2026."
)

// notFound is the hidden 404: the same envelope as the module OFF (§8).
// The cause stays for the logs.
func notFound(cause error) error {
	return apperr.Wrap(errors.Join(ErrTaxpayerNotFound, cause), apperr.CodeNotFound, msgNotFound)
}

func invalidField(cause error, field, detail, message string) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: detail, Message: message})
}

func unavailable(field, message, fieldMessage string) error {
	return apperr.Wrap(ErrParametersUnavailable, apperr.CodeBusinessRuleViolation, message).
		WithFields(apperr.FieldError{Field: field, Code: DetailParametersUnavailable, Message: fieldMessage})
}
