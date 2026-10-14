package domain

import (
	"fmt"
	"io"
	"log/slog"
)

// redacted is how personal data renders in fmt and slog.
const redacted = "REDACTED"

// redactor gives a value type REDACTED output for every fmt verb and in slog.
type redactor struct{}

// Format redacts the value for every fmt verb (%v, %+v, %#v, %s, %q).
func (redactor) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the value in slog.
func (redactor) LogValue() slog.Value { return slog.StringValue(redacted) }
