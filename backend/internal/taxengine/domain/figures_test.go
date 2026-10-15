package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func figureResult(c money.Cents) *domain.TaxResult {
	r := domain.NewTaxResult(money.Exact(c) * money.ExactPerCent)
	return &r
}

// figuresBase is a 612 + 606 month: one regime with ISR and IVA, a CPA
// topic and a line; one partial regime with ISR only.
func figuresBase(t *testing.T) domain.FiguresInput {
	t.Helper()
	due, err := fiscal.ParseDate("2026-04-17")
	require.NoError(t, err)
	business := domain.RegimeEstimate{
		Key:      domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica},
		Coverage: domain.TreatmentFull,
		ISR:      figureResult(1_234_56),
		IVA:      &domain.IVAComponents{ChargedCollected: 1_600_00, CreditablePaid: 400_00},
		Lines:    []domain.Line{{Seq: 1}},
		CPATopics: []domain.CPATopic{
			"iva_netting",
		},
		Explanation: "explanation:612",
	}
	rent := domain.RegimeEstimate{
		Key:      domain.StrategyKey{Regime: fiscal.RegimeRent, PersonType: fiscal.PersonFisica},
		Coverage: domain.TreatmentPartial,
		ISR:      figureResult(-50_00),
	}
	return domain.FiguresInput{
		Totals: domain.Totals{
			IVA: *figureResult(1_200_00), ISR: *figureResult(1_234_56),
			WithholdingsToRemit: 0, TotalPayable: 2_434_56,
		},
		DueDate: due,
		Regimes: []domain.RegimeEstimate{business, rent},
	}
}

// The scheme is pinned (MODULE-history §2.4): changing the canonical text
// marks every viewed month, so it needs an ADR-0002 amendment and a
// re-baseline of estimate_views. Update this value only with both. It is
// the SHA-256 of this text, checked by hand with `shasum -a 256`:
//
//	figures/1
//	due "2026-04-17"
//	totals isr 123456 iva 120000 remit 0 total 243456
//	regime "606/fisica" status "partial" cpa false remit 0 isr -5000 iva -
//	regime "612/fisica" status "complete" cpa true remit 0 isr 123456 iva 160000 40000 0 0
func TestFiguresFingerprintGolden(t *testing.T) {
	assert.Equal(t, "215e419b172756ab0d8433ab363a87c39903c69eda211e767b7177c8c5f1f5cf", domain.FiguresFingerprint(figuresBase(t)))
}

// Every figure the user sees moves the hash.
func TestFiguresFingerprintChangesWithEveryShownFigure(t *testing.T) {
	want := domain.FiguresFingerprint(figuresBase(t))
	tests := []struct {
		name   string
		mutate func(*domain.FiguresInput)
	}{
		{"total ISR", func(in *domain.FiguresInput) { in.Totals.ISR = *figureResult(1_234_57) }},
		{"total IVA", func(in *domain.FiguresInput) { in.Totals.IVA = *figureResult(-1_200_00) }},
		{"withholdings to remit", func(in *domain.FiguresInput) { in.Totals.WithholdingsToRemit = 1 }},
		{"total payable", func(in *domain.FiguresInput) { in.Totals.TotalPayable = 2_434_57 }},
		{"due date", func(in *domain.FiguresInput) {
			d, err := fiscal.ParseDate("2026-04-20")
			require.NoError(t, err)
			in.DueDate = d
		}},
		{"a regime's ISR", func(in *domain.FiguresInput) { in.Regimes[1].ISR = figureResult(-50_01) }},
		{"a regime's ISR becomes absent", func(in *domain.FiguresInput) { in.Regimes[1].ISR = nil }},
		{"a regime's IVA", func(in *domain.FiguresInput) { in.Regimes[0].IVA.CreditablePaid = 400_01 }},
		{"a regime's IVA withheld", func(in *domain.FiguresInput) { in.Regimes[0].IVA.WithheldFromTaxpayer = 1 }},
		{"a regime's IVA excluded", func(in *domain.FiguresInput) { in.Regimes[0].IVA.DefinitiveExcluded = 1 }},
		{"a regime's IVA becomes absent", func(in *domain.FiguresInput) { in.Regimes[0].IVA = nil }},
		{"a regime's withholdings", func(in *domain.FiguresInput) { in.Regimes[0].WithheldToRemit = 1 }},
		{"a regime's status", func(in *domain.FiguresInput) {
			in.Regimes[0].MissingInputs = []domain.MissingInput{{Key: domain.InputPriorProvisionalPayments}}
		}},
		{"a regime's CPA review", func(in *domain.FiguresInput) { in.Regimes[0].CPATopics = nil }},
		{"one regime fewer", func(in *domain.FiguresInput) { in.Regimes = in.Regimes[:1] }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := figuresBase(t)
			tt.mutate(&in)

			assert.NotEqual(t, want, domain.FiguresFingerprint(in))
		})
	}
}

// What the user does not see never moves the hash (OQ-H4): texts,
// explanations, breakdown lines, sub-centavo precision, the regimes'
// order, which CPA topic is pending, and which input is missing.
func TestFiguresFingerprintIgnoresWhatIsNotShown(t *testing.T) {
	want := domain.FiguresFingerprint(figuresBase(t))
	tests := []struct {
		name   string
		mutate func(*domain.FiguresInput)
	}{
		{"breakdown lines", func(in *domain.FiguresInput) {
			in.Regimes[0].Lines = []domain.Line{{Seq: 2}, {Seq: 1}}
		}},
		{"explanation", func(in *domain.FiguresInput) { in.Regimes[0].Explanation = "other" }},
		{"regime disclaimers", func(in *domain.FiguresInput) {
			in.Regimes[0].Disclaimers = []domain.DisclaimerCode{"deductibility_claimed"}
		}},
		{"sub-centavo precision", func(in *domain.FiguresInput) {
			r := domain.NewTaxResult(money.Exact(1_234_56)*money.ExactPerCent + 1)
			in.Regimes[0].ISR = &r
		}},
		{"regime order", func(in *domain.FiguresInput) {
			in.Regimes[0], in.Regimes[1] = in.Regimes[1], in.Regimes[0]
		}},
		{"which CPA topic", func(in *domain.FiguresInput) { in.Regimes[0].CPATopics = []domain.CPATopic{"other_topic"} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := figuresBase(t)
			tt.mutate(&in)

			assert.Equal(t, want, domain.FiguresFingerprint(in))
		})
	}
}

func TestFiguresFingerprintFormat(t *testing.T) {
	assert.Regexp(t, fingerprintFormat, domain.FiguresFingerprint(domain.FiguresInput{}))
}
