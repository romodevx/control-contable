package domain

// Error is a domain error of the taxpayers module. Its values are constants,
// so callers match them with errors.Is and nothing can reassign them.
type Error string

func (e Error) Error() string { return string(e) }

// Domain errors. None of them ever carries the offending RFC. The comment on
// each names the contract ErrorCode the adapter maps it to.
const (
	// ErrInvalidPersonType reports a person type other than "fisica" or
	// "moral" (validation_failed).
	ErrInvalidPersonType Error = "taxpayers: invalid person type"
	// ErrRFCMalformed reports an RFC that is not 3 (moral) or 4 (física)
	// letters from A-Z, Ñ and &, six digits and a 3-character A-Z/0-9
	// homoclave (validation_failed).
	ErrRFCMalformed Error = "taxpayers: malformed rfc"
	// ErrRFCInvalidDate reports an RFC whose YYMMDD part is not a calendar
	// date (validation_failed).
	ErrRFCInvalidDate Error = "taxpayers: rfc date is not a calendar date"
	// ErrRFCGeneric reports one of the SAT's generic RFCs, which identify no
	// taxpayer (validation_failed).
	ErrRFCGeneric Error = "taxpayers: generic rfc is not a taxpayer"
	// ErrPersonTypeMismatch reports an RFC whose length implies a person type
	// other than the declared one (validation_failed).
	ErrPersonTypeMismatch Error = "taxpayers: rfc does not match the person type"
	// ErrOptionUnknown reports an option key the regime does not define
	// (validation_failed).
	ErrOptionUnknown Error = "taxpayers: unknown option"
	// ErrOptionMissing reports a required option with neither a value nor a
	// default (validation_failed).
	ErrOptionMissing Error = "taxpayers: required option missing"
	// ErrOptionType reports an option value of the wrong JSON type for its
	// definition (validation_failed).
	ErrOptionType Error = "taxpayers: option value has the wrong type"
	// ErrOptionValue reports an option value of the right type that the
	// definition does not allow (validation_failed).
	ErrOptionValue Error = "taxpayers: option value not allowed"
	// ErrInvalidOptionDef reports an inconsistent catalog definition: a
	// server fault, never the client's (internal_error).
	ErrInvalidOptionDef Error = "taxpayers: invalid option definition"
	// ErrInvalidPeriod reports a period that is not "YYYY-MM"
	// (validation_failed).
	ErrInvalidPeriod Error = "taxpayers: invalid period"
	// ErrInvalidDisplayName reports a display name that is empty after
	// trimming, longer than 80 characters or not printable text
	// (validation_failed).
	ErrInvalidDisplayName Error = "taxpayers: invalid display name"
)

// OptionError ties an option error to the key it concerns, so the adapter
// can name the field. Key is a catalog key or a client-sent key, never an
// RFC.
type OptionError struct {
	Key string
	Err error
}

func (e *OptionError) Error() string { return "taxpayers: option " + e.Key + ": " + e.Err.Error() }

// Unwrap returns the underlying domain error.
func (e *OptionError) Unwrap() error { return e.Err }
