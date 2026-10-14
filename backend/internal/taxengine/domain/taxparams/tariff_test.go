package taxparams_test

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Synthetic tariffs only: the real 2026 tables arrive with M5-04 and are
// checked row by row by the golden suite (M5-07), never here.

// marginalTable is a three-row marginal tariff shaped like art. 96: each
// fixed fee is the previous fee plus (upper - lower) × rate, rounded.
func marginalTable(t testing.TB) taxparams.Table {
	return taxparams.Table{
		Key: keyMonthly, Kind: taxparams.TableMarginal,
		Rows: []taxparams.Row{
			{Lower: mustCents(t, "0.01"), Upper: mustCents(t, "746.04"), Rate: mustRate(t, "0.0192")},
			{Lower: mustCents(t, "746.05"), Upper: mustCents(t, "6332.05"), FixedFee: mustCents(t, "14.32"), Rate: mustRate(t, "0.064")},
			{Lower: mustCents(t, "6332.06"), Unbounded: true, FixedFee: mustCents(t, "371.82"), Rate: mustRate(t, "0.1088")},
		},
	}
}

// flatTable holds non-marginal brackets ending at a ceiling, like RESICO PF.
func flatTable(t testing.TB) taxparams.Table {
	return taxparams.Table{
		Key: keyBrackets, Kind: taxparams.TableFlat,
		Rows: []taxparams.Row{
			{Lower: mustCents(t, "0.01"), Upper: mustCents(t, "25000.00"), Rate: mustRate(t, "0.01")},
			{Lower: mustCents(t, "25000.01"), Upper: mustCents(t, "50000.00"), Rate: mustRate(t, "0.011")},
		},
	}
}

func mustExact(t testing.TB, pesos string) money.Exact {
	t.Helper()
	e, err := mustCents(t, pesos).Exact()
	require.NoError(t, err)
	return e
}

// subCent is an Exact amount strictly between two centavos.
const subCent = money.Exact(money.ExactPerCent / 2)

func TestBracketPicksTheRowOfAnAmount(t *testing.T) {
	marginal, flat := marginalTable(t), flatTable(t)
	tests := []struct {
		name   string
		table  taxparams.Table
		amount money.Exact
		number int
	}{
		{name: "one millionth of a peso", table: marginal, amount: 1, number: 1},
		{name: "first lower bound", table: marginal, amount: mustExact(t, "0.01"), number: 1},
		{name: "first upper bound", table: marginal, amount: mustExact(t, "746.04"), number: 1},
		{name: "between upper and the next lower", table: marginal, amount: mustExact(t, "746.04") + subCent, number: 1},
		{name: "second lower bound", table: marginal, amount: mustExact(t, "746.05"), number: 2},
		{name: "second upper bound", table: marginal, amount: mustExact(t, "6332.05"), number: 2},
		{name: "unbounded row", table: marginal, amount: mustExact(t, "6332.06"), number: 3},
		{name: "largest amount", table: marginal, amount: math.MaxInt64, number: 3},
		{name: "flat first bracket", table: flat, amount: mustExact(t, "25000.00"), number: 1},
		{name: "flat ceiling", table: flat, amount: mustExact(t, "50000.00"), number: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.table.Bracket(tt.amount)
			require.NoError(t, err)
			assert.Equal(t, tt.number, b.Number)
			assert.Equal(t, tt.table.Rows[tt.number-1], b.Row)
			assert.Equal(t, tt.table.Kind, b.Kind)
		})
	}
}

func TestBracketRejectsAmountsOutsideTheTable(t *testing.T) {
	tests := []struct {
		name    string
		table   taxparams.Table
		amount  money.Exact
		wantErr error
	}{
		{name: "zero", table: marginalTable(t), amount: 0, wantErr: taxparams.ErrAmountNotPositive},
		{name: "negative", table: marginalTable(t), amount: -1, wantErr: taxparams.ErrAmountNotPositive},
		{name: "above a bounded table", table: flatTable(t), amount: mustExact(t, "50000.00") + 1, wantErr: taxparams.ErrAboveTable},
		{name: "no rows", table: taxparams.Table{Key: keyMonthly}, amount: 1, wantErr: taxparams.ErrInvalidTable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.table.Bracket(tt.amount)
			require.ErrorIs(t, err, tt.wantErr)
			_, err = tt.table.Tax(tt.amount)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTaxAppliesTheBracket(t *testing.T) {
	tests := []struct {
		name   string
		table  taxparams.Table
		amount money.Exact
		excess money.Exact
		want   money.Exact
	}{
		// 14.32 + (1000.00 - 746.05) × 0.064 = 14.32 + 16.2528.
		{name: "marginal", table: marginalTable(t), amount: mustExact(t, "1000.00"), excess: mustExact(t, "253.95"), want: 30_572_800},
		{name: "marginal at a lower bound", table: marginalTable(t), amount: mustExact(t, "746.05"), excess: 0, want: mustExact(t, "14.32")},
		{name: "below the first lower bound", table: marginalTable(t), amount: subCent, excess: 0, want: 0},
		// Non-marginal: the bracket's rate applies to the whole amount.
		{name: "flat", table: flatTable(t), amount: mustExact(t, "25000.01"), excess: mustExact(t, "25000.01"), want: 275_000_110},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.table.Bracket(tt.amount)
			require.NoError(t, err)
			assert.Equal(t, tt.excess, b.Excess(tt.amount))
			got, err := tt.table.Tax(tt.amount)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTaxReportsOverflow(t *testing.T) {
	huge := money.Cents(math.MaxInt64)
	tests := []struct {
		name  string
		table taxparams.Table
	}{
		{name: "fixed fee beyond Exact", table: taxparams.Table{Kind: taxparams.TableMarginal, Rows: []taxparams.Row{
			{Lower: 1, Unbounded: true, FixedFee: huge},
		}}},
		{name: "fee plus excess", table: taxparams.Table{Kind: taxparams.TableMarginal, Rows: []taxparams.Row{
			{Lower: 1, Unbounded: true, FixedFee: huge / money.ExactPerCent, Rate: money.RateScale},
		}}},
		{name: "rate beyond one", table: taxparams.Table{Kind: taxparams.TableMarginal, Rows: []taxparams.Row{
			{Lower: 1, Unbounded: true, Rate: 2 * money.RateScale},
		}}},
		{name: "flat rate beyond one", table: taxparams.Table{Kind: taxparams.TableFlat, Rows: []taxparams.Row{
			{Lower: 1, Unbounded: true, Rate: 2 * money.RateScale},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.table.Tax(math.MaxInt64)
			require.ErrorIs(t, err, money.ErrOverflow)
		})
	}
}

// The property tests run on synthetic tables drawn from a fixed seed, so a
// failure always reproduces (ADR-0002 §5).
const (
	propertyTables  = 200
	amountsPerTable = 200
	maxRows         = 12
	maxRowWidth     = 10_000_000 // centavos: 100,000 pesos
	rateStep        = 100        // 0.0001
	maxRateSteps    = 3_500      // 0.35
)

func newRand() *rand.Rand {
	return rand.New(rand.NewPCG(2026, 5)) //nolint:gosec // a fixed, reproducible seed is the point; nothing here is secret
}

// randomTable draws a valid tariff: rows contiguous from 0.01, rates
// non-decreasing. A marginal table ends unbounded and its fees follow the
// published rule fee[i+1] = round(fee[i] + (upper[i] - lower[i]) × rate[i]);
// a flat one may end at a ceiling.
func randomTable(t testing.TB, rng *rand.Rand, kind taxparams.TableKind) taxparams.Table {
	t.Helper()
	n := 1 + rng.IntN(maxRows)
	rows := make([]taxparams.Row, n)
	lower, fee, rate := money.Cents(1), money.Cents(0), money.Rate(0)
	for i := range rows {
		rate += money.Rate(rng.IntN(maxRateSteps/n+1) * rateStep)
		rows[i] = taxparams.Row{Lower: lower, Upper: lower + money.Cents(rng.IntN(maxRowWidth)), Rate: rate}
		if kind == taxparams.TableMarginal {
			rows[i].FixedFee = fee
			fee = nextFee(t, rows[i])
		}
		lower = rows[i].Upper + 1
	}
	if kind == taxparams.TableMarginal || rng.IntN(2) == 0 {
		rows[n-1].Unbounded, rows[n-1].Upper = true, 0
	}
	return taxparams.Table{Key: keyMonthly, Kind: kind, Rows: rows}
}

func nextFee(t testing.TB, r taxparams.Row) money.Cents {
	t.Helper()
	fee, err := r.FixedFee.Exact()
	require.NoError(t, err)
	excess, err := (r.Upper - r.Lower).MulRate(r.Rate)
	require.NoError(t, err)
	total, err := fee.Add(excess)
	require.NoError(t, err)
	return total.Round()
}

// randomAmount mixes the amounts that break lookups: sub-centavo, exact row
// bounds, the gap between a row's upper and the next lower, and anything.
func randomAmount(t testing.TB, rng *rand.Rand, table taxparams.Table) money.Exact {
	t.Helper()
	r := table.Rows[rng.IntN(len(table.Rows))]
	upper := r.Upper
	if r.Unbounded {
		upper = r.Lower + maxRowWidth
	}
	var cents money.Cents
	switch rng.IntN(4) {
	case 0:
		return money.Exact(1 + rng.IntN(money.ExactPerCent-1)) // (0, 0.01)
	case 1:
		cents = r.Lower
	case 2:
		e, err := upper.Exact()
		require.NoError(t, err)
		return e + money.Exact(rng.IntN(money.ExactPerCent)) // [upper, upper + 0.01)
	default:
		cents = r.Lower + money.Cents(rng.Int64N(int64(upper-r.Lower)+1))
	}
	e, err := cents.Exact()
	require.NoError(t, err)
	return e
}

// onCentGrid rounds e to the centavo, the grid tariffs are published on.
func onCentGrid(t testing.TB, e money.Exact) money.Exact {
	t.Helper()
	grid, err := e.Round().Exact()
	require.NoError(t, err)
	return grid
}

// covers is the lookup's specification, checked row by row: row i holds
// the amounts from its lower bound (anything above 0 for the first row) up
// to the next row's lower bound (or the ceiling, or ∞ for the last row).
func covers(table taxparams.Table, i int, amount money.Exact) bool {
	lastRow := len(table.Rows) - 1
	r := table.Rows[i]
	from := i == 0 || money.Exact(r.Lower)*money.ExactPerCent <= amount
	switch {
	case i < lastRow:
		return from && amount < money.Exact(table.Rows[i+1].Lower)*money.ExactPerCent
	case r.Unbounded:
		return from
	}
	return from && amount <= money.Exact(r.Upper)*money.ExactPerCent
}

func TestBracketIsTotalOverPositiveAmounts(t *testing.T) {
	rng := newRand()
	for range propertyTables {
		for _, kind := range []taxparams.TableKind{taxparams.TableMarginal, taxparams.TableFlat} {
			table := randomTable(t, rng, kind)
			for range amountsPerTable {
				amount := randomAmount(t, rng, table)
				matching := slices.Collect(func(yield func(int) bool) {
					for i := range table.Rows {
						if covers(table, i, amount) && !yield(i) {
							return
						}
					}
				})
				b, err := table.Bracket(amount)
				if len(matching) == 0 {
					require.ErrorIs(t, err, taxparams.ErrAboveTable, "amount %d over %v", amount, table.Rows)
					continue
				}
				require.Len(t, matching, 1, "amount %d is in exactly one row", amount)
				require.NoError(t, err)
				require.Equal(t, matching[0]+1, b.Number, "amount %d over %v", amount, table.Rows)
			}
		}
	}
}

// TestTaxIsMonotonic: more income never means less tax. On the centavo grid
// the tariffs are published on, the rounded tax never decreases, even across
// a row edge (the fee rule makes the edge exact to the centavo); inside one
// row the exact tax never decreases either.
func TestTaxIsMonotonic(t *testing.T) {
	rng := newRand()
	for range propertyTables {
		for _, kind := range []taxparams.TableKind{taxparams.TableMarginal, taxparams.TableFlat} {
			table := randomTable(t, rng, kind)
			amounts := make([]money.Exact, 0, amountsPerTable+2*len(table.Rows))
			for range amountsPerTable {
				amounts = append(amounts, onCentGrid(t, randomAmount(t, rng, table)))
			}
			for _, r := range table.Rows {
				amounts = append(amounts, money.Exact(r.Lower)*money.ExactPerCent, money.Exact(r.Upper)*money.ExactPerCent)
			}
			slices.Sort(amounts)
			var previous money.Cents
			for _, amount := range amounts {
				tax, err := table.Tax(amount)
				if amount <= 0 || err != nil {
					continue // the empty upper of an unbounded row, or above a ceiling
				}
				require.GreaterOrEqual(t, tax.Round(), previous, "tax at %d over %v", amount, table.Rows)
				previous = tax.Round()
			}
		}
	}
}

func TestTaxIsMonotonicInsideARow(t *testing.T) {
	rng := newRand()
	for range propertyTables {
		for _, kind := range []taxparams.TableKind{taxparams.TableMarginal, taxparams.TableFlat} {
			table := randomTable(t, rng, kind)
			for range amountsPerTable {
				a, b := randomAmount(t, rng, table), randomAmount(t, rng, table)
				if a > b {
					a, b = b, a
				}
				ba, errA := table.Bracket(a)
				bb, errB := table.Bracket(b)
				if errA != nil || errB != nil || ba.Number != bb.Number {
					continue
				}
				taxA, err := table.Tax(a)
				require.NoError(t, err)
				taxB, err := table.Tax(b)
				require.NoError(t, err)
				require.LessOrEqual(t, taxA, taxB, "row %d: tax(%d) > tax(%d)", ba.Number, a, b)
			}
		}
	}
}

// TestMarginalTaxIsContinuousAtRowEdges: crossing from a row's upper bound to
// the next row's lower bound moves the tax by at most half a centavo (the
// rounding of the published fixed fee), never by a jump.
func TestMarginalTaxIsContinuousAtRowEdges(t *testing.T) {
	rng := newRand()
	for range propertyTables {
		table := randomTable(t, rng, taxparams.TableMarginal)
		for i := range len(table.Rows) - 1 {
			atUpper, err := table.Tax(money.Exact(table.Rows[i].Upper) * money.ExactPerCent)
			require.NoError(t, err)
			atNextLower, err := table.Tax(money.Exact(table.Rows[i+1].Lower) * money.ExactPerCent)
			require.NoError(t, err)
			gap := atNextLower - atUpper
			require.LessOrEqual(t, max(gap, -gap), subCent, "edge %d of %v", i+1, table.Rows)
		}
	}
}
