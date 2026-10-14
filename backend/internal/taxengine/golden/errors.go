package golden

// Error is a golden harness error. Its values are constants, so callers
// match them with errors.Is and nothing can reassign them.
type Error string

func (e Error) Error() string { return string(e) }

// The harness errors.
const (
	// ErrNotHandAuthored reports a case file that does not start with
	// HandAuthoredHeader: it was written (or rewritten) by a tool.
	ErrNotHandAuthored Error = "golden: case file is not marked hand-authored"
	// ErrSyntax reports malformed YAML, an unknown field or a second
	// document.
	ErrSyntax Error = "golden: invalid case syntax"
	// ErrInexactAmount reports an amount that is not a quoted decimal with
	// exactly two places ("150.96"): unquoted numbers are YAML floats.
	ErrInexactAmount Error = "golden: amount is not an exact centavo string"
	// ErrInvalidCase reports a case whose fields break the format rules.
	ErrInvalidCase Error = "golden: invalid case"
	// ErrLayout reports a file outside <regime>-<pf|pm>/<id>.yaml, a
	// directory that disagrees with the case, or a repeated id.
	ErrLayout Error = "golden: case file layout"
	// ErrNoCases reports a golden root with no case: an empty suite gates
	// nothing.
	ErrNoCases Error = "golden: no cases"
)
