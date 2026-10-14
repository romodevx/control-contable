// Package params loads the tax engine's versioned fiscal parameters
// (ADR-0002 §3) from YAML files laid out as <fiscal_year>/v<N>.yaml, embedded
// in the binary from data/ with go:embed.
//
// Load decodes every file strictly (unknown fields, unquoted numbers, a second
// document and malformed dates are errors), checks that each file's path
// agrees with its fiscal_year and version, validates it into a
// taxparams.Params and returns the taxparams.Catalog. Every bad file is
// reported, with its path, in one error: a bad file fails the boot (and CI),
// never an estimate. The schema is docs/architecture/tax-params-schema.md.
package params
