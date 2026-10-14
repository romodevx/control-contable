package taxengine_test

import (
	"math/big"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

// The reference implementation: slow and obviously correct, written from
// the law's wording, sharing nothing with the engine but the parameter
// file's text. Every expected value of a generated case comes from here,
// never from the engine (PLAN M5-07).
//
//   - Lookup: a linear scan for the row whose [lower, upper] holds the
//     amount (upper absent: "en adelante").
//   - Marginal tariff (LISR 96, 106, 116; RIF): fixed fee + (amount − lower)
//     × rate. Flat tariff (LISR 113-E): amount × rate.
//   - Arithmetic in math/big rationals, rounded once to the centavo, half
//     away from zero (ADR-0002 §4).

const flatKind = "flat"

// refDecimal parses a decimal string from the parameter file exactly.
func refDecimal(s string) *big.Rat {
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		panic("reference: not a decimal: " + s)
	}
	return r
}

// refPesos is an amount in centavos as pesos.
func refPesos(c money.Cents) *big.Rat { return big.NewRat(int64(c), 100) }

// refRate is a fixed-point rate as a fraction.
func refRate(r money.Rate) *big.Rat { return big.NewRat(int64(r), money.RateScale) }

// refRow returns the row holding amount, or false above (or between) rows.
func refRow(t golden.RawTable, amount *big.Rat) (golden.RawRow, bool) {
	for _, r := range t.Rows {
		above := amount.Cmp(refDecimal(r.Lower)) >= 0
		below := r.Upper == "" || amount.Cmp(refDecimal(r.Upper)) <= 0
		if above && below {
			return r, true
		}
	}
	return golden.RawRow{}, false
}

// refTax is the tariff's tax on amount, rounded to the centavo, or false
// when no row holds the amount.
func refTax(t golden.RawTable, amount money.Cents) (money.Cents, bool) {
	x := refPesos(amount)
	r, ok := refRow(t, x)
	if !ok {
		return 0, false
	}
	if t.Kind == flatKind {
		return refRound(new(big.Rat).Mul(x, refDecimal(r.Rate))), true
	}
	excess := new(big.Rat).Sub(x, refDecimal(r.Lower))
	marginal := new(big.Rat).Mul(excess, refDecimal(r.Rate))
	return refRound(new(big.Rat).Add(refDecimal(r.FixedFee), marginal)), true
}

// refRound rounds pesos to the centavo, half away from zero.
func refRound(pesos *big.Rat) money.Cents {
	scaled := new(big.Rat).Mul(pesos, big.NewRat(100, 1))
	num, den := new(big.Int).Abs(scaled.Num()), scaled.Denom()
	q, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if new(big.Int).Mul(rem, big.NewInt(2)).Cmp(den) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if scaled.Sign() < 0 {
		q.Neg(q)
	}
	return money.Cents(q.Int64())
}

// refInForce returns the index of the entry whose [from, to] holds date
// (ISO dates compare as text; an empty to is open-ended), or -1.
func refInForce(froms, tos []golden.RawDate, date string) int {
	for i := range froms {
		if string(froms[i]) <= date && (tos[i] == "" || date <= string(tos[i])) {
			return i
		}
	}
	return -1
}
