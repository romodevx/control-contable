package taxparams

import (
	"fmt"
	"sort"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Bracket is the row of a tariff an amount falls in.
type Bracket struct {
	Kind   TableKind
	Number int // 1-based row number, as the published table counts ("renglón 6")
	Row    Row
}

// Bracket returns the row amount falls in: the last row whose lower bound is
// at or below amount, and the first row for anything in (0, 0.01). So every
// positive amount maps to exactly one row, including an amount with a
// fraction of a centavo between one row's upper bound and the next row's
// lower bound. The search is sort.Search over rows sorted by lower bound
// (NewParams guarantees the order), O(log k).
//
// It fails with ErrAmountNotPositive for amount <= 0, ErrAboveTable above
// the ceiling of a bounded (flat) table, and ErrInvalidTable for a table
// without rows.
func (t Table) Bracket(amount money.Exact) (Bracket, error) {
	if len(t.Rows) == 0 {
		return Bracket{}, fmt.Errorf("%s: no rows: %w", t.Key, ErrInvalidTable)
	}
	if amount <= 0 {
		return Bracket{}, fmt.Errorf("%s: amount %d: %w", t.Key, amount, ErrAmountNotPositive)
	}
	// Compare in centavos, rounded down: lower <= amount exactly when
	// lower <= floor(amount / 0.01), and nothing is converted up to Exact, so
	// no bound can overflow.
	whole := money.Cents(amount / money.ExactPerCent)
	above := sort.Search(len(t.Rows), func(i int) bool { return t.Rows[i].Lower > whole })
	i := max(above-1, 0)
	r := t.Rows[i]
	if !r.Unbounded && i == len(t.Rows)-1 && exceeds(amount, r.Upper) {
		return Bracket{}, fmt.Errorf("%s: amount %d above %s: %w", t.Key, amount, r.Upper, ErrAboveTable)
	}
	return Bracket{Kind: t.Kind, Number: i + 1, Row: r}, nil
}

// exceeds reports amount > limit without converting limit to Exact.
func exceeds(amount money.Exact, limit money.Cents) bool {
	whole, fraction := money.Cents(amount/money.ExactPerCent), amount%money.ExactPerCent
	return whole > limit || (whole == limit && fraction > 0)
}

// Tax returns the tax table t computes on amount: the bracket's fixed fee
// plus its rate over the excess of the lower bound for a marginal tariff
// (LISR 96, 106), or the bracket's rate over the whole amount for a flat one
// (RESICO PF, LISR 113-E). Unrounded: the caller rounds once, at the end.
//
// The published fixed fee is applied literally. Real tariffs are continuous
// only up to a few centavos: SAT does not chain the published fees to the
// centavo, so crossing some row edges the tax goes down by up to n centavos
// on the n-th accumulated table of LISR 106 (Anexo 8 2026, November edge
// 5->6: 0.10). That is DOF data; the engine neither corrects nor smooths it
// (params/tariff_continuity_2026_test.go pins and bounds it).
func (t Table) Tax(amount money.Exact) (money.Exact, error) {
	b, err := t.Bracket(amount)
	if err != nil {
		return 0, err
	}
	return b.Tax(amount)
}

// Excess returns the part of amount the bracket's rate applies to: amount
// minus the lower bound for a marginal row (zero below it, for an amount
// under 0.01 in the first row), the whole amount for a flat one.
func (b Bracket) Excess(amount money.Exact) money.Exact {
	if b.Kind == TableFlat {
		return amount
	}
	// Bracket only picks a row whose lower bound is at most amount (or the
	// first row), so the conversion cannot overflow once amount is above it.
	if !exceeds(amount, b.Row.Lower) {
		return 0
	}
	return amount - money.Exact(b.Row.Lower)*money.ExactPerCent
}

// Tax returns the fixed fee plus the rate over the excess (see Table.Tax).
// It fails only with money.ErrOverflow.
func (b Bracket) Tax(amount money.Exact) (money.Exact, error) {
	fee, err := b.Row.FixedFee.Exact()
	if err != nil {
		return 0, fmt.Errorf("row %d fixed fee: %w", b.Number, err)
	}
	marginal, err := b.Excess(amount).MulRate(b.Row.Rate)
	if err != nil {
		return 0, fmt.Errorf("row %d excess: %w", b.Number, err)
	}
	return fee.Add(marginal)
}
