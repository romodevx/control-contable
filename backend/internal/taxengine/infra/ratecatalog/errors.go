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
)
