package domain

import (
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Ledger is the computation trace of one regime estimate: strategies compute
// only by appending steps to it, so the breakdown is the math (ADR-0002 §1).
//
// It keeps a running value. Input, Lookup and Scale start a new running
// value (an anchor); Add, Subtract and AddProduct change it; Result closes a
// subtotal and the running value carries on from it. ResultRounded and
// ScaleRounded are the one exception to "the running value stays exact":
// they carry on from the figure as shown, for a figure later steps are
// defined against as the user sees it (the annual ISR the credits are
// subtracted from, MODULE-annual §2.7). Both round once, with Exact.Round. Each line shows its
// value rounded to the centavo while the running value stays exact, so at a
// Result the shown lines since the last anchor may miss the rounded result
// by a centavo or two: the Ledger then appends a rounding_adjustment line
// first ("Ajuste por redondeo"), and the shown lines always reconcile
// (MODULE-tax-engine §2.7). Every step except Info and the adjustment needs
// a legal basis.
//
// Errors are sticky, like bufio.Writer's: the first one turns every later
// step into a no-op, and Lines reports it. The zero Ledger is ready to use.
type Ledger struct {
	lines   []Line
	running money.Exact
	shown   money.Cents // the shown lines since the last anchor, summed
	open    bool        // an anchor has started a running value
	err     error
}

// NewLedger returns an empty Ledger.
func NewLedger() *Ledger { return &Ledger{} }

// Running returns the exact running value.
func (l *Ledger) Running() money.Exact { return l.running }

// Lines returns a copy of the lines in order, or the first error.
func (l *Ledger) Lines() ([]Line, error) {
	if l.err != nil {
		return nil, l.err
	}
	lines := make([]Line, len(l.lines))
	for i, line := range l.lines {
		lines[i] = line.clone()
	}
	return lines, nil
}

// Input starts a running value from amount: a sum of records or a tax input.
func (l *Ledger) Input(s Step, amount money.Exact) {
	if l.ready(s, true, false) {
		l.anchor(s, OpInput, amount, 0, 0)
	}
}

// Add adds amount to the running value.
func (l *Ledger) Add(s Step, amount money.Exact) {
	if l.ready(s, true, true) {
		running, err := l.running.Add(amount)
		l.change(s, OpAdd, running, err, amount, 0, amount.Round())
	}
}

// Subtract subtracts amount from the running value. Its line shows amount
// itself; the operation carries the sign.
func (l *Ledger) Subtract(s Step, amount money.Exact) {
	if l.ready(s, true, true) {
		running, err := l.running.Sub(amount)
		l.change(s, OpSubtract, running, err, amount, 0, -amount.Round())
	}
}

// AddProduct adds base × r to the running value: a multiply line showing
// the product and the rate, such as a tariff's excess over the lower bound
// times its marginal rate.
func (l *Ledger) AddProduct(s Step, base money.Exact, r money.Rate) {
	if !l.ready(s, true, true) {
		return
	}
	product, err := base.MulRate(r)
	if l.fail(s, err) {
		return
	}
	running, err := l.running.Add(product)
	l.change(s, OpMultiply, running, err, product, r, product.Round())
}

// Scale multiplies the running value by r and starts a new running value
// from the product: a multiply line such as "utilidad × 30 %".
func (l *Ledger) Scale(s Step, r money.Rate) {
	if !l.ready(s, true, true) {
		return
	}
	product, err := l.running.MulRate(r)
	if !l.fail(s, err) {
		l.anchor(s, OpMultiply, product, r, 0)
	}
}

// ScaleRounded is Scale for a product later steps take as shown, such as
// the RESICO annual ISR the payments are subtracted from (MODULE-annual
// §2.7): the line shows the exact product, the running value starts from
// it rounded to the centavo.
func (l *Ledger) ScaleRounded(s Step, r money.Rate) {
	l.Scale(s, r)
	l.carryShown(s)
}

// Lookup finds base's row in table and starts a new running value from the
// row's fixed fee: a lookup line naming the row. It returns the bracket, so
// the strategy can AddProduct its excess times its rate; the zero Bracket
// after an error.
func (l *Ledger) Lookup(s Step, table taxparams.Table, base money.Exact) taxparams.Bracket {
	if !l.ready(s, true, false) {
		return taxparams.Bracket{}
	}
	b, err := table.Bracket(base)
	if l.fail(s, err) {
		return taxparams.Bracket{}
	}
	fee, err := b.Row.FixedFee.Exact()
	if l.fail(s, err) {
		return taxparams.Bracket{}
	}
	l.anchor(s, OpLookup, fee, 0, b.Number)
	return b
}

// Result closes a subtotal or the tax result: a result line showing the
// running value rounded once, preceded by a rounding_adjustment line when
// the lines shown since the last anchor do not add up to it. It returns the
// exact running value, which carries on; zero after an error.
func (l *Ledger) Result(s Step) money.Exact {
	if !l.ready(s, true, true) {
		return 0
	}
	rounded := l.running.Round()
	if gap := rounded - l.shown; gap != 0 {
		// gap is a centavo per rounded line at most: no overflow.
		l.append(Line{
			Tax: s.Tax, Code: LineCode(LabelRoundingAdjustment), Label: LabelRoundingAdjustment,
			Operation: OpRoundingAdjustment, Amount: gap, Exact: money.Exact(gap) * money.ExactPerCent,
			Source: SourceComputed(),
		})
	}
	l.append(l.line(s, OpResult, l.running, 0, 0))
	l.shown = rounded
	return l.running
}

// ResultRounded closes a figure that later steps take as shown, such as
// the annual ISR the credits are subtracted from (MODULE-annual §2.7:
// balance = ISR − credits): like Result, but the running value carries on
// from the rounded result, which it returns. Without it a half centavo of
// the exact value could round a later result past the figure (an ISR of
// 100.005 shown 100.01, less 100.01 paid, would be −0.01).
func (l *Ledger) ResultRounded(s Step) money.Exact {
	l.Result(s)
	l.carryShown(s)
	if l.err != nil {
		return 0
	}
	return l.running
}

// carryShown moves the running value to the figure shown for it (the
// last line's rounded value), so later steps start from what the user
// sees. A centavo rounded up past Exact's range is ErrOverflow.
func (l *Ledger) carryShown(s Step) {
	if l.err != nil {
		return
	}
	shown, err := l.shown.Exact()
	if !l.fail(s, err) {
		l.running = shown
	}
}

// Info records a fact with no arithmetic ("Retención definitiva"); it needs
// no legal basis and leaves the running value alone.
func (l *Ledger) Info(s Step, amount money.Exact) {
	if l.ready(s, false, false) {
		l.append(l.line(s, OpInfo, amount, 0, 0))
	}
}

// ready validates s and the ledger's state, recording the first error. It
// reports whether the step may run.
func (l *Ledger) ready(s Step, needsBasis, needsRunning bool) bool {
	if l.err != nil {
		return false
	}
	var err error
	switch {
	case !s.Tax.Valid() || s.Code == "" || s.Label == "" || s.Source.String() == "":
		err = ErrInvalidStep
	case needsBasis && (len(s.Basis) == 0 || slices.ContainsFunc(s.Basis, func(c Citation) bool { return !c.Valid() })):
		err = ErrMissingCitation
	case needsRunning && !l.open:
		err = ErrNoSegment
	}
	return !l.fail(s, err)
}

// fail records err (if any) against step s and reports whether it did.
func (l *Ledger) fail(s Step, err error) bool {
	if err == nil {
		return false
	}
	l.err = fmt.Errorf("ledger step %d %s: %w", len(l.lines)+1, s.Code, err)
	return true
}

// anchor starts a running value at v.
func (l *Ledger) anchor(s Step, op Operation, v money.Exact, r money.Rate, row int) {
	l.running, l.shown, l.open = v, v.Round(), true
	l.append(l.line(s, op, v, r, row))
}

// change moves the running value to running (unless computing it failed
// with err): the line shows v, and shown is what it adds to the shown sum.
func (l *Ledger) change(s Step, op Operation, running money.Exact, err error, v money.Exact, r money.Rate, shown money.Cents) {
	if l.fail(s, err) {
		return
	}
	l.running, l.shown = running, l.shown+shown
	l.append(l.line(s, op, v, r, 0))
}

func (l *Ledger) line(s Step, op Operation, v money.Exact, r money.Rate, row int) Line {
	return Line{
		Tax: s.Tax, Code: s.Code, Label: s.Label, Args: slices.Clone(s.Args), Operation: op,
		Amount: v.Round(), Exact: v, Rate: r, Row: row, Basis: slices.Clone(s.Basis),
		Source: s.Source, NeedsCPAReview: s.NeedsCPAReview,
	}
}

func (l *Ledger) append(line Line) {
	line.Seq = len(l.lines) + 1
	l.lines = append(l.lines, line)
}
