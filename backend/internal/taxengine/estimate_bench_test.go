package taxengine_test

import (
	"testing"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// BenchmarkEstimateMonth measures the pure engine side of one estimate
// (MODULE-tax-engine §2.11, OQ-T4): 3 regimes (612, 606 and the
// informational 605) with every month January..October aggregated, the
// real registry and the embedded parameters, and the previous month's IVA.
// The ports are in memory, so this is the compute budget (< 2 ms) inside
// the p95 < 300 ms of getTaxpayerEstimate, whose database side (one GROUP BY
// per table) M5-23 measures with EXPLAIN ANALYZE.
//
//	go test -run '^$' -bench EstimateMonth -benchmem ./internal/taxengine/
func BenchmarkEstimateMonth(b *testing.B) {
	c := tx612Total(b)
	c.period = "2026-10"
	c.regimes = append(c.regimes,
		txHold(b, fiscal.RegimeRent, map[domain.OptionKey]domain.OptionValue{domain.OptionRentDeductionMode: domain.TextOption("actual")}),
		txHold(b, "605", nil))
	c.regimes[2].Treatment = domain.TreatmentInformational
	c.ledger = txLedger{}
	for m := time.January; m <= time.October; m++ {
		for _, regime := range []fiscal.RegimeCode{fiscal.RegimeBusiness, fiscal.RegimeRent, "605"} {
			c.ledger.incomes = append(c.ledger.incomes, usecase.IncomeMonthSums{Regime: regime, Month: m, Sums: domain.IncomeSums{
				Cash: 30_000_00, ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 30_000_00, IVA: 4_800_00}},
				IVAWithheld: domain.Withheld{ByMoral: 1_066_67}, ISRWithheld: domain.Withheld{ByMoral: 1_000_00},
			}})
			c.ledger.expenses = append(c.ledger.expenses, usecase.ExpenseMonthSums{Regime: regime, Month: m, Sums: domain.ExpenseSums{
				DeductiblePaid: 10_000_00, CreditableIVAPaid: 1_600_00,
			}})
		}
	}
	estimator := c.estimator(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := estimator.EstimateMonth(b.Context(), txUser, txTaxpayer, c.period); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEstimateRange measures the pure engine side of the dashboard's
// trend (M6-02, MODULE-dashboard §2.4): the same three regimes, the ten
// months January..October estimated in one range (each month with its
// previous month's IVA). The database side (one aggregate per table for the
// whole window) is TestPerf_dashboardReadPath's.
//
//	go test -run '^$' -bench EstimateRange -benchmem ./internal/taxengine/
func BenchmarkEstimateRange(b *testing.B) {
	c := tx612Total(b)
	c.regimes = append(c.regimes,
		txHold(b, fiscal.RegimeRent, map[domain.OptionKey]domain.OptionValue{domain.OptionRentDeductionMode: domain.TextOption("actual")}),
		txHold(b, "605", nil))
	c.regimes[2].Treatment = domain.TreatmentInformational
	c.ledger = txLedger{}
	for m := time.January; m <= time.October; m++ {
		for _, regime := range []fiscal.RegimeCode{fiscal.RegimeBusiness, fiscal.RegimeRent, "605"} {
			c.ledger.incomes = append(c.ledger.incomes, usecase.IncomeMonthSums{Regime: regime, Month: m, Sums: domain.IncomeSums{
				Cash: 30_000_00, ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 30_000_00, IVA: 4_800_00}},
			}})
			c.ledger.expenses = append(c.ledger.expenses, usecase.ExpenseMonthSums{Regime: regime, Month: m, Sums: domain.ExpenseSums{
				DeductiblePaid: 10_000_00, CreditableIVAPaid: 1_600_00,
			}})
		}
	}
	from, err := domain.ParsePeriod("2026-01")
	if err != nil {
		b.Fatal(err)
	}
	to, err := domain.ParsePeriod("2026-10")
	if err != nil {
		b.Fatal(err)
	}
	estimator := c.estimator(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := estimator.EstimateRange(b.Context(), txUser, txTaxpayer, from, to); err != nil {
			b.Fatal(err)
		}
	}
}
