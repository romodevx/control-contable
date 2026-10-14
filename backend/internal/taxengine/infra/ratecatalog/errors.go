package ratecatalog

// Error is a rate catalog error. Its values are constants, matched with
// errors.Is. Lookup failures inside a set wrap the taxparams errors instead
// (ErrNotInForce, ErrUnknownKey, ErrKindMismatch).
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrNoCatalog reports a Catalog built without parameter sets.
	ErrNoCatalog Error = "ratecatalog: no parameter catalog"
	// ErrDateNotCovered reports a date whose fiscal year has neither its own
	// set nor the previous year's to fall back on (it also wraps
	// taxparams.ErrNoParameters). The contract maps it to
	// parameters_unavailable.
	ErrDateNotCovered Error = "ratecatalog: no parameter set covers the date"
	// ErrMissingIVARate reports a parameter set without the general 16 % or
	// the zero IVA rate (it also wraps the taxparams lookup error). New
	// refuses it: the boolean port would otherwise turn the gap into "not in
	// force" and refuse every such record with rate_not_valid_on_date.
	ErrMissingIVARate Error = "ratecatalog: a parameter set lacks a required IVA rate"
)
