package usecase

import (
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Client-safe messages. They never echo the input, an RFC or another
// user's data.
const (
	msgInvalidInput      = "One or more fields are invalid." // the router's wording for every validation_failed
	msgNotFound          = "The taxpayer was not found."
	msgRegimeNotFound    = "The taxpayer does not hold this regime."
	msgRFCRegistered     = "This RFC is already registered in your account."
	msgIDConflict        = "The request conflicts with the current state."
	msgRegimeActive      = "This regime is already active."
	msgRegimeEnded       = "This regime has already ended."
	msgIncompatible      = "These regimes cannot be held in the same periods."
	msgArchived          = "This taxpayer is archived. Unarchive it to make changes."
	msgArchivedDetail    = "Unarchive the taxpayer first."
	msgHasRecords        = "This regime has records. Confirm to end it; the records are kept."
	msgRecordsAfter      = "This regime has records after the end period."
	msgRecordsAfterField = "Choose a period equal to or after the latest record."
	msgLastActive        = "A taxpayer needs at least one active regime. Archive the taxpayer instead."
	msgLastActiveField   = "Archive the taxpayer instead."
	msgNotActive         = "The taxpayer does not hold this regime in that period."
	msgConfirm           = "Confirm the deletion."
	msgInvalidField      = "This value is not valid."
	msgOutOfRange        = "This period is out of range."
	msgNotForPersonType  = "This regime is not available for the taxpayer's person type."
	msgDuplicate         = "This regime appears more than once."
	msgRequired          = "This field is required."
	msgUnknownOption     = "This option does not exist for the regime."
	msgInvalidOption     = "This value is not allowed for the option."
	msgRFCLength         = "A persona física has a 13-character RFC; a persona moral, 12."
	msgRFCFormat         = "The RFC has characters or a shape that is not valid."
	msgRFCDate           = "The date in the RFC is not a valid date."
	msgRFCGeneric        = "A generic RFC cannot be registered as a taxpayer."
)

// fieldErrors collects the field failures of one request, so a 400 names
// every offending field at once.
type fieldErrors struct {
	fields []apperr.FieldError
	causes []error
}

func (f *fieldErrors) add(field, code, message string, cause error) {
	f.fields = append(f.fields, apperr.FieldError{Field: field, Code: code, Message: message})
	f.causes = append(f.causes, cause)
}

// err returns the 400 validation_failed, or nil when nothing was added.
func (f *fieldErrors) err() error {
	if len(f.fields) == 0 {
		return nil
	}
	return apperr.Wrap(errors.Join(f.causes...), apperr.CodeValidationFailed, msgInvalidInput).WithFields(f.fields...)
}

// addPeriod records a period failure (detail is DetailInvalid or
// DetailOutOfRange).
func (f *fieldErrors) addPeriod(field, detail string, cause error) {
	message := msgInvalidField
	if detail == DetailOutOfRange {
		message = msgOutOfRange
	}
	f.add(field, detail, message, cause)
}

// rfcDetails names the FieldError of each RFC rule (MODULE-taxpayers §4).
var rfcDetails = []struct {
	target  error
	code    string
	message string
}{
	{domain.ErrPersonTypeMismatch, DetailRFCLengthMismatch, msgRFCLength},
	{domain.ErrRFCMalformed, DetailRFCFormat, msgRFCFormat},
	{domain.ErrRFCInvalidDate, DetailRFCDate, msgRFCDate},
	{domain.ErrRFCGeneric, DetailRFCGeneric, msgRFCGeneric},
}

// addRFC records an RFC failure on the rfc field. The domain errors never
// carry the RFC, so neither does the result.
func (f *fieldErrors) addRFC(err error) {
	for _, d := range rfcDetails {
		if errors.Is(err, d.target) {
			f.add(fieldRFC, d.code, d.message, err)
			return
		}
	}
	f.add(fieldRFC, DetailRFCFormat, msgRFCFormat, err)
}

// addOptions records each option failure of OptionSet.Validate under
// prefix ("options" or "regimes.N.options"). A failure that is not an
// option error (a broken catalog definition) is returned instead: it is a
// server fault.
func (f *fieldErrors) addOptions(prefix string, err error) error {
	for _, e := range flatten(err) {
		optionErr, ok := errors.AsType[*domain.OptionError](e)
		if !ok || errors.Is(optionErr, domain.ErrInvalidOptionDef) {
			return fmt.Errorf("taxpayers: regime options: %w", err)
		}
		code, message := optionDetail(optionErr.Err)
		f.add(prefix+"."+optionErr.Key, code, message, optionErr)
	}
	return nil
}

func optionDetail(err error) (string, string) {
	switch {
	case errors.Is(err, domain.ErrOptionUnknown):
		return DetailUnknownOption, msgUnknownOption
	case errors.Is(err, domain.ErrOptionMissing):
		return DetailRequired, msgRequired
	default: // ErrOptionType, ErrOptionValue
		return DetailInvalidOption, msgInvalidOption
	}
}

// flatten returns the errors joined in err (errors.Join), or err alone.
func flatten(err error) []error {
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}
	return []error{err}
}

// archivedError is the 422 for a write on an archived taxpayer.
func archivedError() error {
	return apperr.Wrap(ErrTaxpayerArchived, apperr.CodeBusinessRuleViolation, msgArchived).
		WithFields(apperr.FieldError{Field: fieldTaxpayerID, Code: DetailTaxpayerArchived, Message: msgArchivedDetail})
}

// incompatibleError is the 422 for two overlapping regimes the catalog does
// not allow together; message is the catalog's reason.
func incompatibleError(field, code, other, reason string) error {
	cause := fmt.Errorf("%w: %s and %s", ErrIncompatibleRegimes, code, other)
	return apperr.Wrap(cause, apperr.CodeInvalidRegimeCombination, reason).WithFields(apperr.FieldError{
		Field: field, Code: DetailIncompatible, Message: code + " is incompatible with " + other + ".",
	})
}

// businessRule is a 422 business_rule_violation naming one field.
func businessRule(cause error, message, field, detail, fieldMessage string) error {
	return apperr.Wrap(cause, apperr.CodeBusinessRuleViolation, message).
		WithFields(apperr.FieldError{Field: field, Code: detail, Message: fieldMessage})
}

// notFound is the 404 of an id that names no taxpayer of the caller.
func notFound() error { return apperr.Wrap(ErrTaxpayerNotFound, apperr.CodeNotFound, msgNotFound) }

// presentRules map the sentinels stores return (and the domain errors a
// database CHECK backstops) to client errors.
var presentRules = []struct {
	target  error
	code    apperr.Code
	message string
}{
	{ErrTaxpayerNotFound, apperr.CodeNotFound, msgNotFound},
	{ErrRegimeNotHeld, apperr.CodeNotFound, msgRegimeNotFound},
	{ErrRFCAlreadyRegistered, apperr.CodeRFCAlreadyRegistered, msgRFCRegistered},
	{ErrIDConflict, apperr.CodeConflict, msgIDConflict},
	{ErrRegimeAlreadyActive, apperr.CodeConflict, msgRegimeActive},
	{ErrRegimeAlreadyEnded, apperr.CodeConflict, msgRegimeEnded},
	{ErrRegimeHasRecords, apperr.CodeRegimeHasRecords, msgHasRecords},
	{ErrIncompatibleRegimes, apperr.CodeInvalidRegimeCombination, msgIncompatible},
	{ErrRegimeNotForPersonType, apperr.CodeValidationFailed, msgNotForPersonType},
	{domain.ErrInvalidDisplayName, apperr.CodeValidationFailed, msgInvalidInput},
	{domain.ErrInvalidPeriod, apperr.CodeValidationFailed, msgInvalidInput},
	{ErrPeriodOutOfRange, apperr.CodeValidationFailed, msgInvalidInput},
}

// present turns the error of a use case into its client form: an
// *apperr.Error passes as is, a known sentinel gets its contract code, and
// anything else is an infrastructure failure wrapped with op (500).
func present(op string, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*apperr.Error](err); ok {
		return err
	}
	for _, rule := range presentRules {
		if errors.Is(err, rule.target) {
			return apperr.Wrap(err, rule.code, rule.message)
		}
	}
	return fmt.Errorf("taxpayers: %s: %w", op, err)
}
