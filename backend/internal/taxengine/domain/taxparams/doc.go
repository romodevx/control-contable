// Package taxparams is the tax engine's typed, in-memory model of the
// versioned fiscal parameters (ADR-0002 §3): dated rates, amounts and flags,
// and tariff tables, each resolved by date with Params.At and its typed
// siblings.
//
// A Params is built only through NewParams, which validates the whole set
// and reports every problem at once (fail-fast at startup, never during an
// estimate): validity ranges per key never overlap, rates are exact decimals
// in [0, 1], amounts are non-negative centavos, tariff rows are contiguous
// from 0.01 with no gaps or overlaps, and a monthly table key carries exactly
// the twelve months. A Catalog holds every published set and picks the one an
// estimate uses (Catalog.Select), falling back to the previous fiscal year,
// flagged provisional, while the new year is not published yet.
//
// The package is pure: no I/O, no clock (dates are passed in, as the fiscal
// shared kernel's fiscal.Date, ADR-0007), no floats. The
// YAML files and their decoding live in internal/taxengine/params; the schema
// is documented in docs/architecture/tax-params-schema.md.
package taxparams
