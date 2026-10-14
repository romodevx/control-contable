package domain

// Error is a tax engine domain error. Its values are constants, so callers
// match them with errors.Is and nothing can reassign them.
type Error string

func (e Error) Error() string { return string(e) }

// The domain errors.
const (
	// ErrInvalidPeriod reports a period that is not "YYYY-MM" with a month
	// 01-12 and a year 0001-9999.
	ErrInvalidPeriod Error = "taxengine: invalid period"
	// ErrInvalidMonth reports a month outside 1-12, after the period, or a
	// month range that ends before it starts.
	ErrInvalidMonth Error = "taxengine: invalid month"
	// ErrInvalidTaxInput reports a tax input with no key, an unknown kind,
	// negative money or a rate outside [0, 1].
	ErrInvalidTaxInput Error = "taxengine: invalid tax input"
	// ErrDuplicateTaxInput reports one key (and month) given twice, or one
	// key given with two kinds.
	ErrDuplicateTaxInput Error = "taxengine: duplicate tax input"
	// ErrUnknownInput reports a submitted tax input whose key the regime's
	// definitions (regimes.required_inputs) do not name (400 unknown_input).
	ErrUnknownInput Error = "taxengine: tax input not defined for the regime"
	// ErrInputOutOfRange reports a submitted tax input outside its bounds: a
	// rate outside [0, 1], money outside [0, fiscal.MaxCents] or a month
	// outside 1-12 (400 out_of_range).
	ErrInputOutOfRange Error = "taxengine: tax input out of range"
	// ErrInvalidStep reports a ledger step without code, label or source,
	// or with an unknown tax.
	ErrInvalidStep Error = "taxengine: invalid ledger step"
	// ErrMissingCitation reports a ledger step that needs a legal basis
	// (every step except info and rounding adjustments) and has none, or a
	// citation without law or article.
	ErrMissingCitation Error = "taxengine: ledger step without legal basis"
	// ErrNoSegment reports an arithmetic step before any input or lookup has
	// started a running value.
	ErrNoSegment Error = "taxengine: ledger step before an input"
	// ErrInvalidStrategy reports a nil strategy, or one whose key has a
	// blank regime code or an unknown person type, at registry build.
	ErrInvalidStrategy Error = "taxengine: invalid strategy"
	// ErrDuplicateStrategy reports two strategies registered under one key.
	ErrDuplicateStrategy Error = "taxengine: duplicate strategy"
	// ErrStrategyMismatch reports an estimate for a key other than the
	// input's: a strategy wired to the wrong regime.
	ErrStrategyMismatch Error = "taxengine: strategy key mismatch"
)
