package taxparams

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// Validity is the inclusive range of days an entry is in force, as calendar
// days of the fiscal shared kernel (ADR-0007, ADR-0002 §4). A zero To means
// open-ended: in force until a later entry of the key replaces it (tariffs
// carry on until the next publication; a decree with a sunset, like the 8%
// border IVA, sets To).
type Validity struct {
	From fiscal.Date
	To   fiscal.Date
}

// IsOpenEnded reports whether v has no end.
func (v Validity) IsOpenEnded() bool { return v.To.IsZero() }

// Contains reports whether d is in [From, To].
func (v Validity) Contains(d fiscal.Date) bool {
	return !d.Before(v.From) && (v.IsOpenEnded() || !d.After(v.To))
}

// String renders v as "[from, to]" or "[from, open)".
func (v Validity) String() string {
	if v.IsOpenEnded() {
		return fmt.Sprintf("[%s, open)", v.From)
	}
	return fmt.Sprintf("[%s, %s]", v.From, v.To)
}

// endsBefore reports whether v ends strictly before d, so a range starting
// on d does not overlap it.
func (v Validity) endsBefore(d fiscal.Date) bool {
	return !v.IsOpenEnded() && v.To.Before(d)
}
