package taxengine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// --- (7) the IVA withheld from suppliers is credited the month after ---

// withheldBump is the extra withheld creditable IVA a property adds.
const withheldBump = money.Cents(1_234_56)

// TestEstimatePropertyWithheldIVAIsCreditedTheNextMonth (WH-09, LIVA 5
// fr. IV): for a persona moral's regime held in n − 1 and n, more IVA
// withheld from suppliers in n − 1 lowers n's consolidated IVA by exactly
// that amount and raises the regime's creditable IVA by it; more withheld in
// n leaves n's IVA alone (it waits for n + 1).
func TestEstimatePropertyWithheldIVAIsCreditedTheNextMonth(t *testing.T) {
	checked := map[fiscal.RegimeCode]int{}
	forEachCase(t, func(t *testing.T, c propCase) {
		n := c.month(t)
		regime, ok := c.withholdingPMRegime(n)
		if !ok {
			return
		}
		checked[regime]++
		got := c.estimate(t)

		prior := c.withWithheldCredit(regime, n-1, withheldBump).estimate(t)
		assert.Equal(t, got.IVA.Result.Amount-withheldBump, prior.IVA.Result.Amount,
			"case %d %s %s: the withheld IVA of the month before is credited in full", c.index, regime, c.period)
		before, _ := regimeOf(got, domain.StrategyKey{Regime: regime, PersonType: fiscal.PersonMoral})
		after, _ := regimeOf(prior, domain.StrategyKey{Regime: regime, PersonType: fiscal.PersonMoral})
		assert.Equal(t, before.IVA.CreditablePaid+withheldBump, after.IVA.CreditablePaid, "case %d %s", c.index, regime)
		assert.Contains(t, after.CPATopics, domain.TopicWithheldIVACreditTiming, "case %d %s", c.index, regime)

		same := c.withWithheldCredit(regime, n, withheldBump).estimate(t)
		assert.Equal(t, got.IVA.Result, same.IVA.Result, "case %d %s: this month's withheld IVA waits for the next", c.index, regime)
	})
	for _, code := range []fiscal.RegimeCode{"601", "603", "622", fiscal.RegimeRESICO} {
		assert.Positive(t, checked[code], "no case checks %s/moral", code)
	}
}

// withholdingPMRegime is a persona moral's regime with a strategy, held in
// n − 1 and n of the fiscal year (January's previous December is another
// fiscal year: the use case tests cover it).
func (c propCase) withholdingPMRegime(n time.Month) (fiscal.RegimeCode, bool) {
	if c.person != fiscal.PersonMoral || n == time.January {
		return "", false
	}
	for _, h := range c.holds {
		if h.Treatment != domain.TreatmentInformational && holdCovers(h, n-1) && holdCovers(h, n) {
			return h.Code, true
		}
	}
	return "", false
}

// withWithheldCredit adds amount to the regime's withheld creditable IVA of
// month m, adding the row when the month has none.
func (c propCase) withWithheldCredit(regime fiscal.RegimeCode, m time.Month, amount money.Cents) propCase {
	out := c.clone()
	for i, row := range out.expenses {
		if row.Regime == regime && row.Month == m {
			out.expenses[i].Sums.CreditableIVAWithheld += amount
			return out
		}
	}
	out.expenses = append(out.expenses, usecase.ExpenseMonthSums{
		Regime: regime, Month: m, Sums: domain.ExpenseSums{CreditableIVAWithheld: amount},
	})
	return out
}

// TestEstimatePropertyGeneratorSplitsTheCreditableIVA: the generated
// persona moral expenses carry both creditable parts within the IVA, as
// the expenses table's CHECKs require (creditable paid + withheld <= IVA,
// withheld part <= IVA withheld), so the properties see the split.
func TestEstimatePropertyGeneratorSplitsTheCreditableIVA(t *testing.T) {
	split := 0
	forEachCase(t, func(t *testing.T, c propCase) {
		for _, row := range c.expenses {
			e := row.Sums
			assert.LessOrEqual(t, e.CreditableIVAPaid+e.CreditableIVAWithheld, e.IVAPaid.Rate16.IVA, "case %d", c.index)
			assert.LessOrEqual(t, e.CreditableIVAWithheld, e.IVAWithheldToRemit, "case %d", c.index)
			if e.CreditableIVAWithheld > 0 {
				split++
			}
		}
	})
	assert.Positive(t, split, "no generated expense has a withheld creditable part")
}
