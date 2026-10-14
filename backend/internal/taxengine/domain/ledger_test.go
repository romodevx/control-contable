package domain_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

func lisr(article string) []domain.Citation {
	return []domain.Citation{{Law: "LISR", Article: article}}
}

func step(code domain.LineCode, source domain.Source) domain.Step {
	return domain.Step{
		Tax: domain.TaxISR, Code: code, Label: domain.LabelKey(code), Source: source, Basis: lisr("106"),
	}
}

func pesos(t testing.TB, s string) money.Exact {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	e, err := c.Exact()
	require.NoError(t, err)
	return e
}

func rate(t testing.TB, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

// accumulatedMarch is a synthetic one-row-above-base tariff whose row 6
// reproduces the shape of PF-612-ISR-01 (fee 5,570.52, rate 21.36 %).
func accumulatedMarch(t testing.TB) taxparams.Table {
	return taxparams.Table{
		Key: "isr.art106.accumulated", Kind: taxparams.TableMarginal, Month: 3,
		Rows: []taxparams.Row{
			{Lower: 1, Upper: 52_600_93, Rate: rate(t, "0.1")},
			{Lower: 52_600_94, Unbounded: true, FixedFee: 5_570_52, Rate: rate(t, "0.2136")},
		},
	}
}

// TestLedgerTracesA612Computation follows the contract's 612 breakdown
// example: each figure is a line, and the lines are the computation.
func TestLedgerTracesA612Computation(t *testing.T) {
	l := domain.NewLedger()
	l.Input(step("isr.income_ytd", domain.SourceRecords()), pesos(t, "90000.00"))
	l.Subtract(step("isr.deductions_ytd", domain.SourceRecords()), pesos(t, "30000.00"))
	base := l.Result(step("isr.base", domain.SourceComputed()))
	assert.Equal(t, pesos(t, "60000.00"), base)

	b := l.Lookup(step("isr.tariff_fee", domain.SourceParams("isr.art106.accumulated")), accumulatedMarch(t), base)
	assert.Equal(t, 2, b.Number)
	l.AddProduct(step("isr.tariff_excess", domain.SourceParams("isr.art106.accumulated")), b.Excess(base), b.Row.Rate)
	caused := l.Result(step("isr.caused", domain.SourceComputed()))
	assert.Equal(t, money.Exact(7_150_959_216), caused)

	l.Subtract(step("isr.prior_payments", domain.SourceTaxInput(domain.InputPriorProvisionalPayments)), pesos(t, "4000.00"))
	l.Subtract(step("isr.withheld", domain.SourceRecords()), pesos(t, "3000.00"))
	l.Info(domain.Step{Tax: domain.TaxISR, Code: "isr.note", Label: "isr.note", Source: domain.SourceComputed()}, 0)
	payable := l.Result(step("isr.payable", domain.SourceComputed()))
	assert.Equal(t, money.Exact(150_959_216), payable)
	assert.Equal(t, payable, l.Running())

	lines, err := l.Lines()
	require.NoError(t, err)
	type shown struct {
		op     domain.Operation
		amount money.Cents
	}
	var got []shown
	for i, line := range lines {
		assert.Equal(t, i+1, line.Seq)
		got = append(got, shown{line.Operation, line.Amount})
	}
	assert.Equal(t, []shown{
		{domain.OpInput, 90_000_00}, {domain.OpSubtract, 30_000_00}, {domain.OpResult, 60_000_00},
		{domain.OpLookup, 5_570_52}, {domain.OpMultiply, 1_580_44}, {domain.OpResult, 7_150_96},
		{domain.OpSubtract, 4_000_00}, {domain.OpSubtract, 3_000_00}, {domain.OpInfo, 0}, {domain.OpResult, 150_96},
	}, got)
	assert.Equal(t, 2, lines[3].Row, "the lookup names its row")
	assert.Equal(t, rate(t, "0.2136"), lines[4].Rate, "the multiply shows its rate")
	assert.Equal(t, "tax_input:prior_provisional_payments", lines[6].Source.String())
	assert.Equal(t, money.Exact(1_580_439_216), lines[4].Exact)
}

func TestLedgerAddsARoundingAdjustment(t *testing.T) {
	tests := []struct {
		name       string
		first, add string // pesos with a sub-centavo part, in millionths
		adjustment money.Cents
		result     money.Cents
	}{
		// The spec's illustration: 100.004 + 50.004 = 150.008 → 150.01.
		{name: "up", first: "100.004", add: "50.004", adjustment: 1, result: 150_01},
		{name: "down", first: "100.006", add: "50.006", adjustment: -1, result: 150_01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := domain.NewLedger()
			l.Input(step("a", domain.SourceRecords()), exactPesos(t, tt.first))
			l.Add(step("b", domain.SourceRecords()), exactPesos(t, tt.add))
			l.Result(step("c", domain.SourceComputed()))
			lines, err := l.Lines()
			require.NoError(t, err)
			require.Len(t, lines, 4)
			adj := lines[2]
			assert.Equal(t, domain.OpRoundingAdjustment, adj.Operation)
			assert.Equal(t, tt.adjustment, adj.Amount)
			assert.Equal(t, domain.LabelRoundingAdjustment, adj.Label)
			assert.Equal(t, domain.SourceComputed(), adj.Source)
			assert.Empty(t, adj.Basis)
			assert.Equal(t, domain.TaxISR, adj.Tax)
			assert.Equal(t, tt.result, lines[3].Amount)
			assert.Equal(t, 4, lines[3].Seq)
		})
	}
}

// exactPesos parses up to six decimals of a peso into Exact.
func exactPesos(t testing.TB, s string) money.Exact {
	t.Helper()
	r, err := money.ParseRate(s) // same grammar and scale: 10^-6
	require.NoError(t, err)
	return money.Exact(r)
}

func TestLedgerScaleStartsFromTheRunningValue(t *testing.T) {
	l := domain.NewLedger()
	l.Input(step("income", domain.SourceRecords()), pesos(t, "1000.00"))
	l.Result(step("profit", domain.SourceComputed()))
	l.Scale(step("rate_30", domain.SourceParams("isr.rate.pm")), rate(t, "0.30"))
	l.Subtract(step("prior", domain.SourceTaxInput(domain.InputPriorProvisionalPayments)), pesos(t, "50.00"))
	got := l.Result(step("payable", domain.SourceComputed()))
	assert.Equal(t, pesos(t, "250.00"), got)
	lines, err := l.Lines()
	require.NoError(t, err)
	assert.Equal(t, money.Cents(300_00), lines[2].Amount)
	assert.Equal(t, domain.OpMultiply, lines[2].Operation)
	assert.Equal(t, rate(t, "0.3"), lines[2].Rate)
}

func TestLedgerLinesAreACopy(t *testing.T) {
	l := domain.NewLedger()
	l.Input(step("a", domain.SourceRecords()), 1)
	lines, err := l.Lines()
	require.NoError(t, err)
	lines[0].Basis[0].Law = "mutated"
	lines[0].Amount = 99
	again, err := l.Lines()
	require.NoError(t, err)
	assert.Equal(t, "LISR", again[0].Basis[0].Law)
	assert.Equal(t, money.Cents(0), again[0].Amount)
}

func TestLedgerErrorsAreSticky(t *testing.T) {
	noBasis := domain.Step{Tax: domain.TaxISR, Code: "x", Label: "x", Source: domain.SourceRecords()}
	badCitation := noBasis
	badCitation.Basis = []domain.Citation{{Law: "LISR"}}
	ok := step("ok", domain.SourceRecords())
	tests := []struct {
		name    string
		run     func(l *domain.Ledger)
		wantErr error
	}{
		{name: "add before an input", run: func(l *domain.Ledger) { l.Add(ok, 1) }, wantErr: domain.ErrNoSegment},
		{name: "subtract before an input", run: func(l *domain.Ledger) { l.Subtract(ok, 1) }, wantErr: domain.ErrNoSegment},
		{name: "product before an input", run: func(l *domain.Ledger) { l.AddProduct(ok, 1, 1) }, wantErr: domain.ErrNoSegment},
		{name: "scale before an input", run: func(l *domain.Ledger) { l.Scale(ok, 1) }, wantErr: domain.ErrNoSegment},
		{name: "result before an input", run: func(l *domain.Ledger) { l.Result(ok) }, wantErr: domain.ErrNoSegment},
		{name: "no citation", run: func(l *domain.Ledger) { l.Input(noBasis, 1) }, wantErr: domain.ErrMissingCitation},
		{name: "citation without article", run: func(l *domain.Ledger) { l.Input(badCitation, 1) }, wantErr: domain.ErrMissingCitation},
		{name: "unknown tax", run: func(l *domain.Ledger) {
			s := ok
			s.Tax = "ieps"
			l.Input(s, 1)
		}, wantErr: domain.ErrInvalidStep},
		{name: "no code", run: func(l *domain.Ledger) {
			s := ok
			s.Code = ""
			l.Info(s, 1)
		}, wantErr: domain.ErrInvalidStep},
		{name: "no label", run: func(l *domain.Ledger) {
			s := ok
			s.Label = ""
			l.Input(s, 1)
		}, wantErr: domain.ErrInvalidStep},
		{name: "no source", run: func(l *domain.Ledger) {
			s := ok
			s.Source = domain.Source{}
			l.Input(s, 1)
		}, wantErr: domain.ErrInvalidStep},
		{name: "lookup of a zero base", run: func(l *domain.Ledger) {
			l.Lookup(ok, accumulatedMarch(t), 0)
		}, wantErr: taxparams.ErrAmountNotPositive},
		{name: "lookup fee overflow", run: func(l *domain.Ledger) {
			table := taxparams.Table{Kind: taxparams.TableMarginal, Rows: []taxparams.Row{{Lower: 1, Unbounded: true, FixedFee: math.MaxInt64}}}
			l.Lookup(ok, table, 1)
		}, wantErr: money.ErrOverflow},
		{name: "add overflow", run: func(l *domain.Ledger) {
			l.Input(ok, math.MaxInt64)
			l.Add(ok, 1)
		}, wantErr: money.ErrOverflow},
		{name: "subtract overflow", run: func(l *domain.Ledger) {
			l.Input(ok, math.MinInt64)
			l.Subtract(ok, 1)
		}, wantErr: money.ErrOverflow},
		{name: "product overflow", run: func(l *domain.Ledger) {
			l.Input(ok, 0)
			l.AddProduct(ok, math.MaxInt64, 2*money.RateScale)
		}, wantErr: money.ErrOverflow},
		{name: "product sum overflow", run: func(l *domain.Ledger) {
			l.Input(ok, math.MaxInt64)
			l.AddProduct(ok, math.MaxInt64, money.RateScale)
		}, wantErr: money.ErrOverflow},
		{name: "scale overflow", run: func(l *domain.Ledger) {
			l.Input(ok, math.MaxInt64)
			l.Scale(ok, 2*money.RateScale)
		}, wantErr: money.ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := domain.NewLedger()
			tt.run(l)
			_, err := l.Lines()
			require.ErrorIs(t, err, tt.wantErr)

			// Every later step is a no-op that keeps the first error.
			l.Input(ok, 5)
			l.Add(ok, 1)
			l.Subtract(ok, 1)
			l.AddProduct(ok, 1, 1)
			l.Scale(ok, 1)
			l.Info(ok, 1)
			assert.Equal(t, taxparams.Bracket{}, l.Lookup(ok, accumulatedMarch(t), 1))
			assert.Zero(t, l.Result(ok))
			lines, err := l.Lines()
			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, lines)
		})
	}
}

// TestLedgerLinesReconcile is the breakdown invariant (MODULE-tax-engine
// §2.7): between two anchors, the displayed lines plus the rounding
// adjustment add up to the displayed result, and that result is the exact
// running value rounded once. Random traces, fixed seed.
func TestLedgerLinesReconcile(t *testing.T) {
	const (
		traces      = 500
		stepsMax    = 30
		amountRange = 1_000_000_000_000 // one million pesos, in Exact units
	)
	rng := rand.New(rand.NewPCG(2026, 2)) //nolint:gosec // a fixed, reproducible seed is the point; nothing here is secret
	amount := func() money.Exact { return money.Exact(rng.Int64N(amountRange)) }
	ok := step("s", domain.SourceComputed())
	for range traces {
		l := domain.NewLedger()
		l.Input(ok, amount())
		scaled := map[int]bool{} // seq of Scale lines: anchors that show as multiply
		for range 1 + rng.IntN(stepsMax) {
			switch rng.IntN(5) {
			case 0:
				l.Add(ok, amount())
			case 1:
				l.Subtract(ok, amount())
			case 2:
				l.AddProduct(ok, amount(), money.Rate(rng.Int64N(money.RateScale)))
			case 3:
				l.Scale(ok, money.Rate(rng.Int64N(money.RateScale)))
				lines, err := l.Lines()
				require.NoError(t, err)
				scaled[len(lines)] = true
			default:
				l.Result(ok)
			}
		}
		exact := l.Result(ok)
		lines, err := l.Lines()
		require.NoError(t, err)
		assert.Equal(t, exact.Round(), lines[len(lines)-1].Amount)

		var shown money.Cents
		for _, line := range lines {
			switch {
			case line.Operation == domain.OpInput, line.Operation == domain.OpLookup, scaled[line.Seq]:
				shown = line.Amount
			case line.Operation == domain.OpSubtract:
				shown -= line.Amount
			case line.Operation == domain.OpResult:
				require.Equal(t, shown, line.Amount, "seq %d: displayed lines + adjustment = result", line.Seq)
				require.Equal(t, line.Exact.Round(), line.Amount, "seq %d: rounded once", line.Seq)
			default: // add, multiply (a product added), rounding_adjustment
				shown += line.Amount
			}
		}
	}
}
