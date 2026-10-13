// Package money holds the fixed-point types for amounts and rates (ADR-0001 O7,
// ADR-0002 §4): Cents (int64 centavos) for storage and the API, Rate (10^-6)
// for tax rates and coeficientes, and Exact (10^-6 peso) for the tax engine's
// intermediate amounts.
//
// Products go through a 128-bit intermediate (math/bits), so they are exact
// for every int64 operand; a result outside int64 is ErrOverflow, never a
// silent wrap. Rounding is half away from zero. There are no floats anywhere,
// and the package imports only the standard library, so any domain may use it.
package money
