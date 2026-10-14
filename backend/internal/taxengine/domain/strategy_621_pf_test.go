package domain_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The 621 PF tests name their helpers with a rif prefix: every regime
// strategy test shares package domain_test (PLAN PG-M5-S runs them in
// parallel), so a generic name would collide.

var rifKey = domain.StrategyKey{Regime: "621", PersonType: fiscal.PersonFisica}

// rifParams is a synthetic parameter set holding only the bimonthly RIF
// tariff (Anexo 8 B.VII). Row 2 reproduces the shape of the real row 6 of
// PF-621-01 (lower 35,067.29, fee 3,713.68, 21.36 %); the real 2026 table
// is exercised by the golden gate, never transcribed here.
func rifParams(t testing.TB) *taxparams.Params {
	t.Helper()
	d := func(s string) fiscal.Date {
		date, err := fiscal.ParseDate(s)
		require.NoError(t, err)
		return date
	}
	p, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear: 2026, Version: 1, PublishedOn: d("2026-10-13"),
		Sources: []taxparams.Source{{
			ID: "dof-synthetic", Title: "Synthetic Anexo 8", URL: "https://example.test/anexo8",
			DOFDate: d("2025-12-31"), Reference: "Anexo 8, B.VII",
		}},
		Tables: []taxparams.Table{{
			Key: "isr.rif.bimonthly", Kind: taxparams.TableMarginal,
			Validity: taxparams.Validity{From: d("2026-01-01"), To: d("2026-12-31")},
			Rows: []taxparams.Row{
				{Lower: 1, Upper: 35_067_28, Rate: rate(t, "0.10")},
				{Lower: 35_067_29, Unbounded: true, FixedFee: 3_713_68, Rate: rate(t, "0.2136")},
			},
			Basis: []taxparams.Citation{{Source: "dof-synthetic", Article: "Anexo 8, B.VII"}},
		}},
	})
	require.NoError(t, err)
	return p
}

// rifMonth is one month of RIF records: cash income, deductions paid and
// the IVA charged at 16 % and paid (creditable).
type rifMonth struct {
	month                  time.Month
	cash, deductible       string
	ivaCharged, creditable string
}

func rifInput(t *testing.T, raw, year string, months ...rifMonth) domain.Input {
	t.Helper()
	p := period(t, raw)
	entries := make([]domain.MonthEntry, 0, len(months))
	for _, m := range months {
		s := domain.MonthSums{
			Income:   domain.IncomeSums{Cash: rifCents(t, m.cash)},
			Expenses: domain.ExpenseSums{DeductiblePaid: rifCents(t, m.deductible), CreditableIVAPaid: rifCents(t, m.creditable)},
		}
		s.Income.ByRate.Rate16.IVA = rifCents(t, m.ivaCharged)
		entries = append(entries, domain.MonthEntry{Month: m.month, Sums: s})
	}
	ys, err := domain.NewYearSums(p.Month(), entries...)
	require.NoError(t, err)
	opts := map[domain.OptionKey]domain.OptionValue{}
	if year != "" {
		opts[domain.OptionRIFYear] = domain.TextOption(year)
	}
	return domain.Input{Key: rifKey, Period: p, Sums: ys, Options: domain.NewOptions(opts)}
}

func rifCents(t testing.TB, s string) money.Cents {
	t.Helper()
	if s == "" {
		return 0
	}
	return cents(t, s)
}

// pf62101 is the research worked example PF-621-01: bimester Sep-Oct 2026,
// year 6, income 100,000 and deductions 60,000.
func pf62101(t *testing.T) domain.Input {
	return rifInput(t, "2026-10", "6",
		rifMonth{month: time.September, cash: "50000.00", deductible: "30000.00"},
		rifMonth{month: time.October, cash: "50000.00", deductible: "30000.00"},
	)
}

func rifShown(lines []domain.Line, tax domain.TaxKind) []string {
	var out []string
	for _, l := range lines {
		if l.Tax == tax {
			out = append(out, fmt.Sprintf("%s %s %s", l.Code, l.Operation, l.Amount))
		}
	}
	return out
}

func TestStrategy621PFIdentity(t *testing.T) {
	s := domain.NewStrategy621PF()

	assert.Equal(t, rifKey, s.Key())
	assert.Equal(t, domain.TreatmentPartial, s.Treatment())
	assert.Empty(t, s.RequiredInputs(domain.Options{}, period(t, "2026-10"), domain.Facts{}),
		"621 has no tax inputs: the year in RIF is a regime option")
}

// PF-621-01 rounds ONCE, at the end (ADR-0002 §4): 4,767.306856 × 50 % =
// 2,383.653428 -> 2,383.65. The research halves the rounded 4,767.31 and
// states 2,383.66; the golden keeps that figure, skipped, until the CPA
// says which the law means (CPA-REVIEW, RIF rounding item).
func TestStrategy621PFComputesPF62101WithOneFinalRounding(t *testing.T) {
	est, err := domain.NewStrategy621PF().Estimate(pf62101(t), rifParams(t))
	require.NoError(t, err)

	require.NotNil(t, est.ISR)
	assert.Equal(t, cents(t, "2383.65"), est.ISR.Amount)
	assert.Equal(t, money.Exact(2_383_653_428), est.ISR.Exact)
	assert.Equal(t, []string{
		"isr.income_bimester input 100000.00",
		"isr.deductions_bimester subtract 60000.00",
		"isr.base result 40000.00",
		"isr.tariff_row lookup 3713.68",
		"isr.tariff_excess multiply 1053.63",
		"isr.caused result 4767.31",
		"isr.rif_reduction multiply 2383.65",
		"isr.payable result 2383.65",
	}, rifShown(est.Lines, domain.TaxISR))
	assert.Equal(t, domain.TreatmentPartial, est.Coverage)
	assert.Equal(t, domain.StatusPartial, est.Status())
	assert.Equal(t, domain.LabelKey("regime_explanation:621/fisica"), est.Explanation)
	assert.Contains(t, est.Disclaimers, domain.DisclaimerDeductibilityClaimed)
}

func TestStrategy621PFTracesTheBimesterAndTheReduction(t *testing.T) {
	est, err := domain.NewStrategy621PF().Estimate(pf62101(t), rifParams(t))
	require.NoError(t, err)

	income := lineByCode(t, est.Lines, "isr.income_bimester")
	assert.Equal(t, []string{"2026-09", "2026-10"}, income.Args)
	assert.Equal(t, "records", income.Source.String())
	assert.Equal(t, 2, lineByCode(t, est.Lines, "isr.tariff_row").Row)
	assert.Equal(t, "params:isr.rif.bimonthly", lineByCode(t, est.Lines, "isr.tariff_row").Source.String())

	reduction := lineByCode(t, est.Lines, "isr.rif_reduction")
	assert.Equal(t, rate(t, "0.50"), reduction.Rate, "the rate is what remains after the 50 % reduction")
	assert.Equal(t, []string{"6", "0.5"}, reduction.Args, "year in RIF and the reduction")
	assert.True(t, reduction.NeedsCPAReview, "the reduction table is not re-verified (CPA-REVIEW)")
	assert.Contains(t, articles(reduction.Basis), "111")
	for _, l := range est.Lines {
		if l.Code != "isr.rif_reduction" && l.Tax == domain.TaxISR {
			assert.False(t, l.NeedsCPAReview, "%s applies a verified rule", l.Code)
		}
	}
	assertCited(t, est.Lines)
}

// Years 1-10 reduce the ISR by 100 % down to 10 % (art. 111, texto vigente
// hasta 2021 ⚠️): the remaining fraction is 0 % in year 1 and 90 % in year 10.
func TestStrategy621PFReducesByTheYearInRIF(t *testing.T) {
	tests := []struct {
		year string
		want string
	}{
		{"1", "0.00"}, {"2", "476.73"}, {"3", "953.46"}, {"4", "1430.19"}, {"5", "1906.92"},
		{"6", "2383.65"}, {"7", "2860.38"}, {"8", "3337.11"}, {"9", "3813.85"}, {"10", "4290.58"},
	}
	for _, tt := range tests {
		t.Run("year "+tt.year, func(t *testing.T) {
			in := pf62101(t)
			in.Options = domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.TextOption(tt.year)})

			est, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))
			require.NoError(t, err)

			assert.Equal(t, cents(t, tt.want), est.ISR.Amount)
		})
	}
}

func TestStrategy621PFRejectsAYearOutsideRIF(t *testing.T) {
	tests := map[string]domain.Options{
		"absent":    {},
		"zero":      domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.TextOption("0")}),
		"eleven":    domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.TextOption("11")}),
		"not a num": domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.TextOption("six")}),
		"a boolean": domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRIFYear: domain.BoolOption(true)}),
	}
	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			for _, raw := range []string{"2026-09", "2026-10"} {
				in := rifInput(t, raw, "")
				in.Options = opts

				_, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))

				require.ErrorIs(t, err, domain.ErrInvalidRIFYear, raw)
			}
		})
	}
}

// The first month of a bimester has no figure: the RIF pays by bimester
// (MODULE-tax-engine §2.4), and the line names the month to look at.
func TestStrategy621PFHasNoFigureInTheBimestersFirstMonth(t *testing.T) {
	in := rifInput(t, "2026-09", "6",
		rifMonth{month: time.September, cash: "50000.00", deductible: "30000.00", ivaCharged: "8000.00"})

	est, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))
	require.NoError(t, err)

	assert.Nil(t, est.ISR)
	assert.Nil(t, est.IVA)
	assert.Zero(t, est.WithheldToRemit)
	assert.Equal(t, domain.StatusPartial, est.Status())
	assert.Equal(t, domain.LabelKey("regime_explanation:621/fisica"), est.Explanation)
	assert.Empty(t, est.CPATopics, "no figure, no unverified rule applied")
	assert.Empty(t, est.Disclaimers)
	require.Len(t, est.Lines, 1)
	pending := est.Lines[0]
	assert.Equal(t, domain.LineCode("isr.rif_bimester_pending"), pending.Code)
	assert.Equal(t, domain.OpInfo, pending.Operation)
	assert.Equal(t, []string{"2026-10"}, pending.Args, "consulta octubre")
}

// The bimester is months n-1 and n only: never January..n (art. 111 pays
// bimonthly definitive payments, not cumulative ones).
func TestStrategy621PFSumsOnlyTheBimester(t *testing.T) {
	in := rifInput(t, "2026-10", "6",
		rifMonth{month: time.August, cash: "999999.00", deductible: "1.00"},
		rifMonth{month: time.September, cash: "50000.00", deductible: "30000.00"},
		rifMonth{month: time.October, cash: "50000.00", deductible: "30000.00"},
	)

	est, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))
	require.NoError(t, err)

	assert.Equal(t, cents(t, "100000.00"), lineByCode(t, est.Lines, "isr.income_bimester").Amount)
	assert.Equal(t, cents(t, "2383.65"), est.ISR.Amount)
}

func TestStrategy621PFBimesterOfEveryEvenMonth(t *testing.T) {
	for m := time.February; m <= time.December; m += 2 {
		raw := fmt.Sprintf("2026-%02d", int(m))
		t.Run(raw, func(t *testing.T) {
			est, err := domain.NewStrategy621PF().Estimate(rifInput(t, raw, "6"), rifParams(t))
			require.NoError(t, err)

			got := lineByCode(t, est.Lines, "isr.income_bimester").Args
			assert.Equal(t, []string{fmt.Sprintf("2026-%02d", int(m-1)), raw}, got)
		})
	}
}

func TestStrategy621PFBaseAtOrBelowZeroCausesNoISR(t *testing.T) {
	in := rifInput(t, "2026-10", "6",
		rifMonth{month: time.September, cash: "10000.00", deductible: "30000.00"},
		rifMonth{month: time.October, cash: "10000.00"})

	est, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))
	require.NoError(t, err)

	assert.Equal(t, money.Cents(0), est.ISR.Amount)
	assert.Equal(t, []string{
		"isr.income_bimester input 20000.00",
		"isr.deductions_bimester subtract 30000.00",
		"isr.base result -10000.00",
		"isr.caused input 0.00",
		"isr.rif_reduction multiply 0.00",
		"isr.payable result 0.00",
	}, rifShown(est.Lines, domain.TaxISR))
}

func TestStrategy621PFClaimsDeductibilityOnlyWithDeductions(t *testing.T) {
	in := rifInput(t, "2026-10", "6",
		rifMonth{month: time.September, cash: "50000.00"}, rifMonth{month: time.October, cash: "50000.00"})

	est, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))
	require.NoError(t, err)

	assert.NotContains(t, est.Disclaimers, domain.DisclaimerDeductibilityClaimed)
}

// The IVA is bimonthly like the ISR (OQ-T7 DECIDED, ⚠️ rif_bimonthly_iva):
// both months' components, one info line flagging the open CPA item.
func TestStrategy621PFComputesTheIVAByBimester(t *testing.T) {
	in := rifInput(t, "2026-10", "6",
		rifMonth{month: time.August, ivaCharged: "99999.00"},
		rifMonth{month: time.September, cash: "50000.00", ivaCharged: "8000.00", creditable: "1000.00"},
		rifMonth{month: time.October, cash: "50000.00", ivaCharged: "8000.00", creditable: "600.00"},
	)

	est, err := domain.NewStrategy621PF().Estimate(in, rifParams(t))
	require.NoError(t, err)

	require.NotNil(t, est.IVA)
	assert.Equal(t, domain.IVAComponents{ChargedCollected: cents(t, "16000.00"), CreditablePaid: cents(t, "1600.00")}, *est.IVA)
	assert.Equal(t, []string{
		"iva.rif_bimester info 0.00",
		"iva.charged.rate_16 input 16000.00",
		"iva.charged result 16000.00",
		"iva.creditable subtract 1600.00",
		"iva.withheld.moral subtract 0.00",
		"iva.regime_net result 14400.00",
	}, rifShown(est.Lines, domain.TaxIVA))
	bimester := lineByCode(t, est.Lines, "iva.rif_bimester")
	assert.True(t, bimester.NeedsCPAReview)
	assert.Equal(t, []string{"2026-09", "2026-10"}, bimester.Args)
	assert.Contains(t, articles(bimester.Basis), "5-E")
	assert.ElementsMatch(t, []domain.CPATopic{domain.TopicRIFReduction, domain.TopicRIFBimonthlyIVA}, est.CPATopics)
	assert.True(t, est.CPAReviewPending(), "the stronger warning: every 621 figure is unverified")
}

// Through the registry, a 621 figure is partial (never complete) and
// carries the CPA topics the use case turns into the strong warning.
func TestStrategy621PFThroughTheRegistry(t *testing.T) {
	r, err := domain.NewRegistry(domain.NewStrategy621PF())
	require.NoError(t, err)

	est, err := r.Estimate(pf62101(t), []time.Month{time.September, time.October}, rifParams(t))
	require.NoError(t, err)

	assert.Equal(t, domain.StatusPartial, est.Status())
	assert.Empty(t, est.MissingInputs)
	assert.True(t, est.CPAReviewPending())
}

func TestStrategy621PFFailsWithoutTheBimonthlyTable(t *testing.T) {
	_, err := domain.NewStrategy621PF().Estimate(pf62101(t), gpmParams(t, "0.30"))

	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
}
