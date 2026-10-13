package domain

import (
	"fmt"
	"io"
	"log/slog"
)

// redacted is how every secret or PII value renders in fmt and slog.
const redacted = "REDACTED"

// redactor gives a value type REDACTED output for every fmt verb and in slog.
// Embedding it is the only way the value types in this package print.
type redactor struct{}

// Format redacts the value for every fmt verb (%v, %+v, %#v, %s, %q, %d).
func (redactor) Format(f fmt.State, _ rune) { _, _ = io.WriteString(f, redacted) }

// LogValue redacts the value in slog.
func (redactor) LogValue() slog.Value { return slog.StringValue(redacted) }
