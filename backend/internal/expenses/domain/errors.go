package domain

import "strings"

// Class is the family of contract ErrorCode a domain error maps to. The
// module's adapter owns the actual mapping (it may import apperr; the domain
// may not).
type Class int

// The classes, in the order the use case reports them.
const (
	// ClassInternal maps to internal_error: a server fault such as an
	// inconsistent parameter. It is also the class of an empty Violations.
	ClassInternal Class = iota
	// ClassValidation maps to validation_failed (400).
	ClassValidation
	// ClassBusinessRule maps to business_rule_violation (422).
	ClassBusinessRule
)

// Error is a domain error of the expenses module. Its value is the contract
// details[].code (MODULE-expenses §4, MODULE-incomes §4), so the adapter
// needs no second table. Values are constants: callers match them with
// errors.Is.
type Error string

// Validation errors (validation_failed, 400).
const (
	ErrRequired       Error = "required"
	ErrInvalid        Error = "invalid"
	ErrMustBePositive Error = "must_be_positive"
	ErrOutOfRange     Error = "out_of_range"
	// ErrUnknownCategory is the use case's: no category rule exists.
	ErrUnknownCategory Error = "unknown_category"
	ErrHasCFDIMismatch Error = "has_cfdi_mismatch"
)

// Business-rule errors (business_rule_violation, 422).
const (
	ErrRateNotValidOnDate       Error = "rate_not_valid_on_date"
	ErrIVAAmountMismatch        Error = "iva_amount_mismatch"
	ErrWithholdingNotApplicable Error = "withholding_not_applicable"
	ErrWithholdingExceedsBase   Error = "withholding_exceeds_base"
)

// ErrInvalidParameter reports server-side data the rules cannot use: a
// catalog rate whose product overflows, a missing cash-payment limit, an
// inconsistent category rule or person type, a missing clock date or catalog
// (internal_error).
const ErrInvalidParameter Error = "invalid_parameter"

func (e Error) Error() string { return "expenses: " + string(e) }

// Detail returns the contract details[].code.
func (e Error) Detail() string { return string(e) }

// Class returns the ErrorCode family of e.
func (e Error) Class() Class {
	switch e {
	case ErrRequired, ErrInvalid, ErrMustBePositive, ErrOutOfRange, ErrUnknownCategory, ErrHasCFDIMismatch:
		return ClassValidation
	case ErrRateNotValidOnDate, ErrIVAAmountMismatch, ErrWithholdingNotApplicable, ErrWithholdingExceedsBase:
		return ClassBusinessRule
	default:
		return ClassInternal
	}
}

// Violation ties a domain error to the contract field it concerns.
type Violation struct {
	Field string
	Err   Error
}

func (v *Violation) Error() string { return "expenses: " + v.Field + ": " + v.Err.Detail() }

// Unwrap returns the domain error.
func (v *Violation) Unwrap() error { return v.Err }

// Violations is every client mistake found in one class, in field order.
// errors.Is matches any of its errors; errors.As finds the first *Violation.
type Violations []Violation

func (vs Violations) Error() string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = v.Field + ": " + v.Err.Detail()
	}
	return "expenses: " + strings.Join(parts, "; ")
}

// Unwrap exposes each *Violation to errors.Is and errors.As.
func (vs Violations) Unwrap() []error {
	errs := make([]error, len(vs))
	for i := range vs {
		errs[i] = &vs[i]
	}
	return errs
}

// Class returns the class of the violations (Assess never mixes classes),
// or ClassInternal when there are none.
func (vs Violations) Class() Class {
	if len(vs) == 0 {
		return ClassInternal
	}
	return vs[0].Err.Class()
}

// add appends a violation of field.
func (vs *Violations) add(field string, err Error) {
	*vs = append(*vs, Violation{Field: field, Err: err})
}
