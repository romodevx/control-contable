// Package logging builds the process logger: log/slog JSON lines with PII
// redaction and the request id of the context on every line (ADR-0001).
//
// Redaction fails closed, in two layers:
//   - by key: an attribute (or group) whose key names sensitive data (rfc,
//     email, password, token, secret, authorization, cookie, api key, curp,
//     pepper, private key) is replaced by Redacted, whatever its value;
//   - by value: RFCs, CURPs, email addresses, bearer credentials, JWTs,
//     refresh tokens, URL credentials and credential key/value pairs
//     ("password=…", "Authorization: Basic …", dumped headers, structs) are
//     masked inside the message and every string, error, fmt.Stringer or
//     other non-scalar value.
//
// Values implementing slog.LogValuer (config.Secret) are resolved first.
//
// Value scrubbing is pattern-based and cannot catch everything, so the rule
// for callers stays: log identifiers (user id, taxpayer id, request id), never
// payloads (request or response bodies, rows, whole structs).
package logging

import (
	"io"
	"log/slog"
)

// RequestIDKey is the top-level key of the request id on every line.
const RequestIDKey = "request_id"

// New returns a JSON logger writing to w at level, with redaction and the
// request id of the context.
func New(w io.Writer, level slog.Leveler) *slog.Logger {
	return slog.New(NewHandler(w, level))
}

// NewHandler returns the redacting JSON handler wrapped so that each record
// logged with a context carrying a request id gets RequestIDKey.
func NewHandler(w io.Writer, level slog.Leveler) slog.Handler {
	base := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level, ReplaceAttr: redactAttr})
	return &contextHandler{base: base, built: base}
}
