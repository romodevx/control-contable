// Package domain holds the taxpayers module's value objects and domain
// errors:
//
//   - PersonType: "fisica" or "moral", the contract's PersonType enum.
//   - RFC: a normalized Registro Federal de Contribuyentes with its shape,
//     date and SAT check digit validated; the person type is inferred from
//     its length (13 física, 12 moral) and may be checked against a declared
//     type (NewRFCFor). Generic RFCs are rejected.
//   - OptionSet: a regime's option definitions (the catalog's option_defs)
//     validating a taxpayer's per-regime options, independent of how either
//     is stored.
//
// Errors are constants of type Error matched with errors.Is; option errors
// arrive wrapped in an *OptionError that names the offending key. None of
// them carries an RFC. The module's adapter maps each one to a contract
// ErrorCode (taxpayershttp.ErrorMapper).
//
// The package imports the standard library, golang.org/x/text/unicode/norm
// (ADR-0004) and internal/platform/money only. An RFC is personal data for
// a persona física: it renders as REDACTED in fmt and slog.
package domain
