package taxengine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The targeted half of the engine property suite (M5-24): every published
// tariff edge, a persona moral's withholdings to remit, the regression of
// the bug the suite found, and the generator's own coverage.

// --- (2b) ISR at every published row edge ---

// edgeDrop is an ISR that went down crossing a row edge: the documented
// exceptions, logged and pinned, never hidden.
type edgeDrop struct {
	key    domain.StrategyKey
	period string
	edge   int // row edge -> edge+1, 1-based
	cents  money.Cents
}

// edgeRegime sweeps one progressive tariff through a single-regime
// taxpayer whose ISR base is the period's cash income alone.
type edgeRegime struct {
	key     domain.StrategyKey
	table   taxparams.Key
	monthly bool // one art. 106 accumulated table per month
	months  []time.Month
	options map[domain.OptionKey]domain.OptionValue
}

var edgeRegimes = []edgeRegime{
	{
		key: domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}, table: "isr.art106.accumulated",
		monthly: true, months: allMonths(),
	},
	{
		key: domain.StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica}, table: "isr.art116.monthly",
		months:  []time.Month{time.March},
		options: map[domain.OptionKey]domain.OptionValue{domain.OptionRentDeductionMode: domain.TextOption("actual")},
	},
	{ // rif_year 10: the smallest reduction, so the tariff shows through.
		key: domain.StrategyKey{Regime: "621", PersonType: fiscal.PersonFisica}, table: "isr.rif.bimonthly",
		months:  []time.Month{time.February, time.August, time.December},
		options: map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.TextOption("10")},
	},
}

func allMonths() []time.Month {
	months := make([]time.Month, 0, int(time.December))
	for m := time.January; m <= time.December; m++ {
		months = append(months, m)
	}
	return months
}

// novemberEdge5 is the drop the params continuity test documents: the 612
// November table goes from 20,425.34 at row 5's upper bound to 20,425.24
// at row 6's lower bound (DOF data, applied as published).
var novemberEdge5 = edgeDrop{
	key: domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica}, period: "2026-11", edge: 5, cents: 10,
}

// TestEstimatePropertyISRAtTariffEdges: at every row edge of the 2026
// progressive tariffs, the ISR one centavo past the upper bound is not
// below the ISR at it, except by the documented drift of the published
// fixed fees (isrDropBound). Each drop is logged, and the November one is
// pinned, so smoothing the published table would fail here too.
func TestEstimatePropertyISRAtTariffEdges(t *testing.T) {
	var drops []edgeDrop
	for _, er := range edgeRegimes {
		for _, m := range er.months {
			for i, row := range er.rows(t, m) {
				at, past := er.isrOn(t, m, row.Upper), er.isrOn(t, m, row.Upper+1)
				drop := at - past
				if drop <= 0 {
					continue
				}
				drops = append(drops, edgeDrop{key: er.key, period: periodString(m), edge: i + 1, cents: drop})
				t.Logf("%s %s edge %d->%d: ISR %s -> %s (published fee drift)", er.key, periodString(m), i+1, i+2, at, past)
				assert.LessOrEqual(t, drop, isrDropBound(er.key, m), "%s %s edge %d->%d", er.key, periodString(m), i+1, i+2)
			}
		}
	}
	assert.Contains(t, drops, novemberEdge5, "the documented November drop must show, not be smoothed away")
	t.Logf("%d edges with a within-drift drop", len(drops))
}

// rows are the bounded rows of the regime's table for month m.
func (er edgeRegime) rows(t *testing.T, m time.Month) []taxparams.Row {
	t.Helper()
	_, catalog := propEngine()
	sel, err := catalog.Select(propYear)
	require.NoError(t, err)
	on := mustPeriod(propYear, m).End()
	var table taxparams.Table
	if er.monthly {
		table, err = sel.Params.MonthlyTable(er.table, m, on)
	} else {
		table, err = sel.Params.Table(er.table, on)
	}
	require.NoError(t, err)
	require.Equal(t, taxparams.TableMarginal, table.Kind)
	return table.Rows[:len(table.Rows)-1]
}

// isrOn estimates the regime alone with cash income in month m only: no
// deductions, no withholdings, zero prior payments (complete).
func (er edgeRegime) isrOn(t *testing.T, m time.Month, income money.Cents) money.Cents {
	t.Helper()
	c := propCase{
		person: er.key.PersonType, period: periodString(m),
		holds: []usecase.RegimeHold{{
			Code: er.key.Regime, Treatment: domain.TreatmentFull, Options: domain.NewOptions(er.options),
			Start: mustPeriod(propYear-1, time.July),
		}},
		incomes: []usecase.IncomeMonthSums{{Regime: er.key.Regime, Month: m, Sums: domain.IncomeSums{Cash: income, Accrual: income}}},
		inputs:  map[fiscal.RegimeCode][]domain.TaxInputValue{},
	}
	for prior := time.January; prior < m; prior++ {
		c.inputs[er.key.Regime] = append(c.inputs[er.key.Regime], domain.TaxInputValue{
			Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: prior,
		})
	}
	r, ok := regimeOf(c.estimate(t), er.key)
	require.True(t, ok)
	require.NotNil(t, r.ISR, "%s %s: no ISR", er.key, c.period)
	return r.ISR.Amount
}

// --- (4b) a persona moral's withholdings to remit are all of them ---

// knownToRemitGaps are regimes that drop part of what the PM withheld from
// its suppliers: real bugs, each pinned by a regression test that fails as
// soon as the gap closes, so the entry goes with the fix.
var knownToRemitGaps = map[domain.StrategyKey]string{
	{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}: "626 PM drops the ISR it withheld from its suppliers",
}

// TestEstimatePropertyPMWithholdingsToRemitAreComplete: a persona moral
// remits every ISR and IVA it withheld from its suppliers in the month
// (§1, TX-601-REMIT-01), as a figure of its own.
func TestEstimatePropertyPMWithholdingsToRemitAreComplete(t *testing.T) {
	var checked int
	forEachCase(t, func(t *testing.T, c propCase) {
		if c.person != fiscal.PersonMoral {
			return
		}
		month := c.month(t)
		for _, r := range c.estimate(t).Regimes {
			_, known := knownToRemitGaps[r.Key]
			if r.Status == domain.StatusInformational || known {
				continue
			}
			checked++
			isr, iva := c.toRemitOf(r.Key.Regime, month)
			assert.Equal(t, isr+iva, r.WithheldToRemit, "case %d %s %s: withholdings to remit", c.index, r.Key, c.period)
		}
	})
	require.Positive(t, checked)
}

func (c propCase) toRemitOf(code fiscal.RegimeCode, m time.Month) (isr, iva money.Cents) {
	for _, row := range c.expenses {
		if row.Regime == code && row.Month == m {
			return row.Sums.ISRWithheldToRemit, row.Sums.IVAWithheldToRemit
		}
	}
	return 0, 0
}

// TestRegression626PMRemitsTheISRItWithheld is the minimal case behind the
// knownToRemitGaps entry, found by M5-24: a RESICO PM that withheld
// 5,000.00 ISR (10 % of 50,000.00 in fees, LISR 106 ¶5) and 5,333.33 IVA
// from a supplier in March must remit 10,333.33, as 601, 603 and 622 do
// (TX-601-REMIT-01). It skips, loudly, while the bug stands and fails once
// it is fixed, so the pending entry cannot outlive the fix.
func TestRegression626PMRemitsTheISRItWithheld(t *testing.T) {
	key := domain.StrategyKey{Regime: fiscal.RegimeRESICO, PersonType: fiscal.PersonMoral}
	c := propCase{
		person: fiscal.PersonMoral, period: "2026-03",
		holds: []usecase.RegimeHold{{Code: key.Regime, Treatment: domain.TreatmentFull, Start: mustPeriod(propYear, time.January)}},
		expenses: []usecase.ExpenseMonthSums{{Regime: key.Regime, Month: time.March, Sums: domain.ExpenseSums{
			DeductiblePaid: 50_000 * propPeso, ISRWithheldToRemit: 5_000 * propPeso, IVAWithheldToRemit: 5_333_33,
		}}},
		inputs: map[fiscal.RegimeCode][]domain.TaxInputValue{},
	}
	r, ok := regimeOf(c.estimate(t), key)
	require.True(t, ok)

	const want = money.Cents(10_333_33)
	reason, pending := knownToRemitGaps[key]
	switch {
	case r.WithheldToRemit == want && pending:
		t.Fatalf("fixed: remove %s from knownToRemitGaps", key)
	case r.WithheldToRemit == want:
		return
	case pending:
		t.Skipf("KNOWN BUG (M5-24): %s: remits %s, want %s", reason, r.WithheldToRemit, want)
	}
	t.Fatalf("%s remits %s, want %s", key, r.WithheldToRemit, want)
}

// --- the generator is not vacuous ---

// TestEstimatePropertyGeneratorCoversEveryStrategy: the cases give every
// registered strategy a figure and reach every status, so a passing
// property says something about each of them.
func TestEstimatePropertyGeneratorCoversEveryStrategy(t *testing.T) {
	seen := map[domain.StrategyKey]bool{}
	statuses := map[domain.Status]bool{}
	forEachCase(t, func(t *testing.T, c propCase) {
		for _, r := range c.estimate(t).Regimes {
			statuses[r.Status] = true
			if r.ISR != nil || r.IVA != nil {
				seen[r.Key] = true
			}
		}
	})
	for _, s := range taxengine.Strategies() {
		assert.True(t, seen[s.Key()], "no generated case gives %s a figure", s.Key())
	}
	for _, s := range []domain.Status{domain.StatusComplete, domain.StatusIncomplete, domain.StatusPartial, domain.StatusInformational} {
		assert.True(t, statuses[s], "no generated case is %s", s)
	}
}
