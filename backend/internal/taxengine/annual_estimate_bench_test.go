package taxengine_test

import (
	"testing"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// annualBenchDeductions is MODULE-annual §2.11's load: 200 deductions.
const annualBenchDeductions = 200

// BenchmarkEstimateAnnual measures the pure side of one annual estimate
// (MODULE-annual §2.11: < 5 ms): a persona física with 612, 606 and 625,
// a salary constancia, 200 personal deductions and every payment a default
// from the Estimator range (in memory here), through the use case with
// the composition root's sections and the embedded parameters. The
// database side is TestPerf_annualReadPath's (cmd/api).
//
//	go test -run '^$' -bench EstimateAnnual -benchmem ./internal/taxengine/
func BenchmarkEstimateAnnual(b *testing.B) {
	in := annualgolden.Input{
		FiscalYear: 2026, PersonType: fiscal.PersonFisica,
		Regimes: []annualgolden.RegimeInput{
			{Regime: "612", Income: 900_000_00, Deductions: 400_000_00, ISRWithheld: 30_000_00},
			{Regime: "606", Income: 240_000_00},
			{Regime: "625", Income: 120_000_00, Deductions: 20_000_00, ISRWithheld: 2_400_00},
		},
		Payments: []annualgolden.Payment{
			{Regime: "612", Amount: 50_000_00, Source: annualgolden.PaymentEstimatedByApp},
			{Regime: "606", Amount: 20_000_00, Source: annualgolden.PaymentEstimatedByApp},
			{Regime: "625", Amount: 1_000_00, Source: annualgolden.PaymentEstimatedByApp},
		},
		SalaryStatements: []annualgolden.SalaryStatement{{Accumulable: 300_000_00, ISRWithheld: 40_000_00}},
	}
	for i := range annualBenchDeductions {
		category, method := annual.CategoryMedical, annual.PaymentElectronic
		if i%10 == 0 {
			method = annual.PaymentCash
		}
		if i%3 == 0 {
			category = annual.CategoryHealthInsurance
		}
		in.PersonalDeductions = append(in.PersonalDeductions, annual.Deduction{
			Category: category, Amount: money.Cents(1_000_00 + i), PaymentMethod: method, Beneficiary: annual.BeneficiarySelf,
		})
	}
	uc := newUseCaseAnnual(b)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := uc.estimate(in); err != nil {
			b.Fatal(err)
		}
	}
}
