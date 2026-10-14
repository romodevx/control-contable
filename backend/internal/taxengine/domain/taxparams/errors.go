package taxparams

// Error is a parameters error. Its values are constants, so callers match
// them with errors.Is and nothing can reassign them.
type Error string

func (e Error) Error() string { return string(e) }

// Validation errors, reported by NewParams (wrapped with the offending path,
// joined so one run lists every problem).
const (
	// ErrInvalidHeader reports a fiscal year outside 2000-9999, a version
	// below 1 or a missing published_on.
	ErrInvalidHeader Error = "taxparams: invalid set header"
	// ErrInvalidSource reports no sources, or a source without id, title,
	// reference or DOF date, with a non-https URL, or with a repeated id.
	ErrInvalidSource Error = "taxparams: invalid source"
	// ErrInvalidKey reports a key that is not dotted lower-case segments of
	// [a-z0-9_], at least two of them.
	ErrInvalidKey Error = "taxparams: invalid key"
	// ErrInvalidKind reports a value or table kind outside the defined ones.
	ErrInvalidKind Error = "taxparams: invalid kind"
	// ErrKindConflict reports one key used with two kinds (rate and money,
	// value and table, marginal and flat).
	ErrKindConflict Error = "taxparams: key used with conflicting kinds"
	// ErrInvalidValidity reports a range without a start or ending before it
	// starts.
	ErrInvalidValidity Error = "taxparams: invalid validity range"
	// ErrOverlap reports two entries of one key (and month) in force on the
	// same day.
	ErrOverlap Error = "taxparams: validity ranges overlap"
	// ErrInvalidBasis reports an entry without a legal basis, or a citation
	// without article or naming an unknown source.
	ErrInvalidBasis Error = "taxparams: invalid legal basis"
	// ErrRateOutOfRange reports a rate outside [0, 1].
	ErrRateOutOfRange Error = "taxparams: rate outside [0, 1]"
	// ErrNegativeMoney reports a negative amount or fixed fee.
	ErrNegativeMoney Error = "taxparams: negative amount"
	// ErrInvalidTable reports a table without rows, a first row not starting
	// at 0.01, a row ending below its start, an unbounded row that is not the
	// last (or has an upper bound), a marginal table whose last row is
	// bounded, or a fixed fee on a flat table.
	ErrInvalidTable Error = "taxparams: invalid table"
	// ErrNotContiguous reports consecutive rows where upper + 0.01 is not the
	// next lower (a gap or an overlap).
	ErrNotContiguous Error = "taxparams: table rows are not contiguous"
	// ErrInvalidMonth reports a table month outside 1-12 (0 = not monthly).
	ErrInvalidMonth Error = "taxparams: invalid month"
	// ErrMonthlyMixed reports a key with both monthly and plain tables.
	ErrMonthlyMixed Error = "taxparams: monthly and plain tables share a key"
	// ErrMonthlyIncomplete reports a monthly table key whose validity window
	// does not hold exactly the twelve months.
	ErrMonthlyIncomplete Error = "taxparams: monthly tables are not exactly twelve"
)

// Lookup and requirement errors.
const (
	// ErrUnknownKey reports a key the set does not define.
	ErrUnknownKey Error = "taxparams: unknown key"
	// ErrKindMismatch reports a lookup or requirement of the wrong kind for
	// the key (a rate read as money, a monthly table read as plain...).
	ErrKindMismatch Error = "taxparams: kind mismatch"
	// ErrNotInForce reports a key with no entry in force on the date.
	ErrNotInForce Error = "taxparams: not in force on date"
)

// Catalog errors.
const (
	// ErrNoParameterSets reports a catalog built from no set (or a nil one).
	ErrNoParameterSets Error = "taxparams: no parameter sets"
	// ErrDuplicateVersion reports two sets with one fiscal year and version.
	ErrDuplicateVersion Error = "taxparams: duplicate fiscal year version"
	// ErrNoParameters reports a fiscal year with neither its own set nor the
	// previous year's to fall back on.
	ErrNoParameters Error = "taxparams: no parameters for fiscal year"
)
