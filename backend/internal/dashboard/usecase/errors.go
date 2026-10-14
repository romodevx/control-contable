package usecase

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/dashboard/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

var (
	// ErrMissingDependency reports a Service wired without one of its ports.
	ErrMissingDependency = errors.New("dashboard: missing dependency")
	// ErrTaxpayerNotFound reports a taxpayer id that names nothing (404).
	ErrTaxpayerNotFound = errors.New("dashboard: taxpayer not found")
	// ErrFuturePeriod reports a period after the current month (400 out_of_range).
	ErrFuturePeriod = errors.New("dashboard: future period")
	// ErrParametersUnavailable reports a period before 2026-01 (422).
	ErrParametersUnavailable = errors.New("dashboard: period before the first month")
	// ErrMonthsOutOfRange reports a trend length outside 6..12 (400).
	ErrMonthsOutOfRange = errors.New("dashboard: months outside 6..12")
)

// FieldError.code values (contract getDashboard; the router answers the
// schema ones first, these are the backstop).
const (
	DetailPattern               = "pattern"
	DetailOutOfRange            = "out_of_range"
	DetailParametersUnavailable = "parameters_unavailable"
	DetailMinimum               = "minimum"
	DetailMaximum               = "maximum"
	fieldPeriod                 = "period"
	fieldMonths                 = "months"
)

// Client-safe messages (contract examples, es-MX for the fields).
const (
	msgInvalidInput          = "One or more fields are invalid." // the router's wording for every validation_failed
	msgNotFound              = "Not found."
	msgPeriodPattern         = "Usa el formato AAAA-MM."
	msgFuturePeriod          = "Aún no hay datos de un mes futuro."
	msgParametersUnavailable = "There are no tax parameters for that period."
	msgParametersField       = "Sólo hay datos desde enero de 2026."
)

func notFound(cause error) error {
	return apperr.Wrap(errors.Join(ErrTaxpayerNotFound, cause), apperr.CodeNotFound, msgNotFound)
}

func invalidField(cause error, field, detail, message string) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: detail, Message: message})
}

func invalidPeriodError(cause error) error {
	return invalidField(cause, fieldPeriod, DetailPattern, msgPeriodPattern)
}

func futurePeriodError() error {
	return invalidField(ErrFuturePeriod, fieldPeriod, DetailOutOfRange, msgFuturePeriod)
}

func parametersUnavailableError() error {
	return apperr.Wrap(ErrParametersUnavailable, apperr.CodeBusinessRuleViolation, msgParametersUnavailable).
		WithFields(apperr.FieldError{Field: fieldPeriod, Code: DetailParametersUnavailable, Message: msgParametersField})
}

func monthsError(months int) error {
	cause := fmt.Errorf("%w: %d", ErrMonthsOutOfRange, months)
	if months < domain.MinMonths {
		return invalidField(cause, fieldMonths, DetailMinimum, fmt.Sprintf("Debe ser como mínimo %d.", domain.MinMonths))
	}
	return invalidField(cause, fieldMonths, DetailMaximum, fmt.Sprintf("Debe ser como máximo %d.", domain.MaxMonths))
}
