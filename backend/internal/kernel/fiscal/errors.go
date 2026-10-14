package fiscal

import "strings"

// Class is the family of contract ErrorCode a domain error maps to. The
// module's adapter owns the actual mapping (it may import apperr; the domain
// and the kernel may not).
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

// Error is a domain error: its code is the contract details[].code, so the
// adapter needs no second table, and it carries its Class. Error is
// comparable; callers match it with errors.Is. A module declares its own
// codes once, as package variables, with ValidationError or
// BusinessRuleError, and never edits the kernel to do so.
type Error struct {
	code  string
	class Class
}

// ValidationError returns the validation_failed (400) error with code.
func ValidationError(code string) Error { return Error{code: code, class: ClassValidation} }

// BusinessRuleError returns the business_rule_violation (422) error with code.
func BusinessRuleError(code string) Error { return Error{code: code, class: ClassBusinessRule} }

// Validation errors shared by every module (validation_failed, 400).
var (
	ErrRequired       = ValidationError("required")
	ErrInvalid        = ValidationError("invalid")
	ErrMustBePositive = ValidationError("must_be_positive")
	ErrOutOfRange     = ValidationError("out_of_range")
)

// Business-rule errors shared by every module (business_rule_violation, 422).
var (
	ErrRateNotValidOnDate       = BusinessRuleError("rate_not_valid_on_date")
	ErrIVAAmountMismatch        = BusinessRuleError("iva_amount_mismatch")
	ErrWithholdingNotApplicable = BusinessRuleError("withholding_not_applicable")
	ErrWithholdingExceedsBase   = BusinessRuleError("withholding_exceeds_base")
)

// ErrInvalidParameter reports server-side data the rules cannot use: a
// catalog rate whose product overflows, an unknown basis or person type, a
// missing clock date or catalog (internal_error). It is never a Violation.
var ErrInvalidParameter = Error{code: "invalid_parameter", class: ClassInternal}

func (e Error) Error() string { return "fiscal: " + e.code }

// Detail returns the contract details[].code.
func (e Error) Detail() string { return e.code }

// Class returns the ErrorCode family of e.
func (e Error) Class() Class { return e.class }

// Violation ties a domain error to the contract field it concerns.
type Violation struct {
	Field string
	Err   Error
}

func (v *Violation) Error() string { return "fiscal: " + v.Field + ": " + v.Err.Detail() }

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
	return "fiscal: " + strings.Join(parts, "; ")
}

// Unwrap exposes each *Violation to errors.Is and errors.As.
func (vs Violations) Unwrap() []error {
	errs := make([]error, len(vs))
	for i := range vs {
		errs[i] = &vs[i]
	}
	return errs
}

// Class returns the class of the violations (a pass never mixes classes),
// or ClassInternal when there are none.
func (vs Violations) Class() Class {
	if len(vs) == 0 {
		return ClassInternal
	}
	return vs[0].Err.Class()
}

// Add appends a violation of field.
func (vs *Violations) Add(field string, err Error) {
	*vs = append(*vs, Violation{Field: field, Err: err})
}
