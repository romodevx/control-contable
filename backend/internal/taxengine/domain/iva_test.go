package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func cents(t testing.TB, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

// twoThirds is the IVA a persona moral withholds (RLIVA 3 fr. I) as the
// record stores it: 2/3 of the IVA, through MulRatio, rounded once. The
// engine never computes it; the fixtures do, the way a CFDI would.
func twoThirds(t testing.TB, iva money.Cents) money.Cents {
	t.Helper()
	s, err := fiscal.ShareTwoThirds.Suggest(iva)
	require.NoError(t, err)
	return s.Cents
}

func codes(lines []domain.Line) []domain.LineCode {
	out := make([]domain.LineCode, len(lines))
	for i, l := range lines {
		out[i] = l.Code
	}
	return out
}

func lineByCode(t testing.TB, lines []domain.Line, code domain.LineCode) domain.Line {
	t.Helper()
	for _, l := range lines {
		if l.Code == code {
			return l
		}
	}
	require.Failf(t, "line not found", "%s", code)
	return domain.Line{}
}

// platformDefinitive is a stand-in for the 625 strategy's rule (M5-14): the
// platform's 50 % withholding settles the IVA of the platform sales
// (LIVA 18-M ⚠️), so what the withholding did not cover is excluded.
func platformDefinitive(in domain.IncomeSums) (money.Cents, error) {
	return in.ByRate.Rate16.IVA.Sub(in.IVAWithheld.ByPlatform)
}

// TestComputeIVAWorkedExamples are the research's worked IVA examples
// (docs/research/tax-regimes-mx-2026.md §A.4, §B.4) as one regime's
// components; the consolidation across regimes is M5-22's.
func TestComputeIVAWorkedExamples(t *testing.T) {
	tests := []struct {
		name       string
		sums       func(t testing.TB) domain.MonthSums
		options    domain.Options
		definitive domain.DefinitiveIVARule
		want       domain.IVAComponents
		wantNet    string
		wantRemit  string
		wantTopics []domain.CPATopic
	}{
		{
			// Collected 30,000 + IVA 4,800 (10,000 of it billed to a PM);
			// IVA paid on expenses 1,600; withheld 2/3 × 1,600.
			name: "PF-612-IVA-01",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{
					Income: domain.IncomeSums{
						ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "30000.00"), IVA: cents(t, "4800.00")}},
						IVAWithheld: domain.Withheld{ByMoral: twoThirds(t, cents(t, "1600.00"))},
					},
					Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "1600.00")},
				}
			},
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "4800.00"), CreditablePaid: cents(t, "1600.00"),
				WithheldFromTaxpayer: cents(t, "1066.67"),
			},
			wantNet: "2133.33", wantRemit: "0.00",
		},
		{
			name: "PF-626-IVA-01",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{
					Income: domain.IncomeSums{
						ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "60000.00"), IVA: cents(t, "9600.00")}},
						IVAWithheld: domain.Withheld{ByMoral: twoThirds(t, cents(t, "3200.00"))},
					},
					Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "2000.00")},
				}
			},
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "9600.00"), CreditablePaid: cents(t, "2000.00"),
				WithheldFromTaxpayer: cents(t, "2133.33"),
			},
			wantNet: "5466.67", wantRemit: "0.00",
		},
		{
			name: "PF-606-IVA-01 commercial rent to a PM",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{Income: domain.IncomeSums{
					ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "20000.00"), IVA: cents(t, "3200.00")}},
					IVAWithheld: domain.Withheld{ByMoral: twoThirds(t, cents(t, "3200.00"))},
				}}
			},
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "3200.00"), WithheldFromTaxpayer: cents(t, "2133.33"),
			},
			wantNet: "1066.67", wantRemit: "0.00",
		},
		{
			name: "PF-606-IVA-01 residential rent is exempt",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{Income: domain.IncomeSums{
					ByRate: domain.IVAByRate{ExemptBase: cents(t, "20000.00")},
				}}
			},
			want:    domain.IVAComponents{},
			wantNet: "0.00", wantRemit: "0.00",
		},
		{
			// IVA withheld from a PF lawyer (fees 50,000 → 8,000 × 2/3) is
			// remitted apart and not also credited (OQ-T6 b).
			name: "PM-601-IVA-01",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{
					Income: domain.IncomeSums{
						ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "2000000.00"), IVA: cents(t, "320000.00")}},
					},
					Expenses: domain.ExpenseSums{
						CreditableIVAPaid:  cents(t, "200000.00"),
						IVAWithheldToRemit: twoThirds(t, cents(t, "8000.00")),
					},
				}
			},
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "320000.00"), CreditablePaid: cents(t, "200000.00"),
			},
			wantNet: "120000.00", wantRemit: "5333.33",
		},
		{
			// TX-NET-01's 606 side: a contribution in favor, netted by M5-22.
			name: "TX-NET-01 606 side",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{
					Income: domain.IncomeSums{
						ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "20000.00"), IVA: cents(t, "3200.00")}},
						IVAWithheld: domain.Withheld{ByMoral: twoThirds(t, cents(t, "3200.00"))},
					},
					Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "3000.00")},
				}
			},
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "3200.00"), CreditablePaid: cents(t, "3000.00"),
				WithheldFromTaxpayer: cents(t, "2133.33"),
			},
			wantNet: "-1933.33", wantRemit: "0.00",
		},
		{
			// Sales 30,000 through the app; IVA 4,800, 50 % withheld by the
			// platform and definitive under 18-M ⚠️: additional IVA 0.
			name: "PF-625-01 definitive platform IVA",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{Income: domain.IncomeSums{
					ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "30000.00"), IVA: cents(t, "4800.00")}},
					IVAWithheld: domain.Withheld{ByPlatform: cents(t, "2400.00")},
				}}
			},
			definitive: platformDefinitive,
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "4800.00"), WithheldFromTaxpayer: cents(t, "2400.00"),
				DefinitiveExcluded: cents(t, "2400.00"),
			},
			wantNet: "0.00", wantRemit: "0.00",
			wantTopics: []domain.CPATopic{domain.TopicPlatformDefinitiveIVA},
		},
		{
			// The border rate is the record's: the incomes module checked it
			// on the record's IVA date, so the engine sums it as entered.
			name: "border 8 % plus 16 %, 0 % and exempt collections",
			sums: func(t testing.TB) domain.MonthSums {
				return domain.MonthSums{
					Income: domain.IncomeSums{ByRate: domain.IVAByRate{
						Rate16:    domain.RateSums{Base: cents(t, "10000.00"), IVA: cents(t, "1600.00")},
						Rate8:     domain.RateSums{Base: cents(t, "10000.00"), IVA: cents(t, "800.00")},
						Rate0Base: cents(t, "5000.00"),
					}},
					Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "400.00")},
				}
			},
			want: domain.IVAComponents{
				ChargedCollected: cents(t, "2400.00"), CreditablePaid: cents(t, "400.00"),
			},
			wantNet: "2000.00", wantRemit: "0.00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := domain.NewLedger()
			got, err := domain.ComputeIVA(l, domain.IVAInput{
				Sums: tt.sums(t), Options: tt.options, Definitive: tt.definitive,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Components)
			assert.Equal(t, cents(t, tt.wantNet), got.Net)
			assert.Equal(t, cents(t, tt.wantRemit), got.WithheldToRemit)
			assert.Equal(t, tt.wantTopics, got.CPATopics)

			lines, err := l.Lines()
			require.NoError(t, err)
			net := lineByCode(t, lines, domain.LineIVANet)
			assert.Equal(t, domain.OpResult, net.Operation)
			assert.Equal(t, cents(t, tt.wantNet), net.Amount)
			assertCited(t, lines)
		})
	}
}

// assertCited checks §2.5: every line except info and rounding adjustments
// cites at least one article, and IVA lines cite the LIVA (or the border
// decree).
func assertCited(t *testing.T, lines []domain.Line) {
	t.Helper()
	for _, l := range lines {
		if l.Operation == domain.OpInfo || l.Operation == domain.OpRoundingAdjustment {
			continue
		}
		require.NotEmpty(t, l.Basis, "%s has no citation", l.Code)
	}
}

func TestComputeIVALinesTracePF612(t *testing.T) {
	l := domain.NewLedger()
	_, err := domain.ComputeIVA(l, domain.IVAInput{Sums: domain.MonthSums{
		Income: domain.IncomeSums{
			ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: cents(t, "30000.00"), IVA: cents(t, "4800.00")}},
			IVAWithheld: domain.Withheld{ByMoral: cents(t, "1066.67")},
		},
		Expenses: domain.ExpenseSums{CreditableIVAPaid: cents(t, "1600.00")},
	}})
	require.NoError(t, err)
	lines, err := l.Lines()
	require.NoError(t, err)

	assert.Equal(t, []domain.LineCode{
		domain.LineIVAChargedRate16, domain.LineIVACharged,
		domain.LineIVACreditable, domain.LineIVAWithheldByMoral, domain.LineIVANet,
	}, codes(lines), "zero optional lines are left out")
	assert.Equal(t, []domain.Operation{
		domain.OpInput, domain.OpResult, domain.OpSubtract, domain.OpSubtract, domain.OpResult,
	}, []domain.Operation{lines[0].Operation, lines[1].Operation, lines[2].Operation, lines[3].Operation, lines[4].Operation})
	for _, line := range lines {
		assert.Equal(t, domain.TaxIVA, line.Tax, "%s", line.Code)
		assert.False(t, line.NeedsCPAReview, "%s", line.Code)
	}
	assert.Equal(t, []domain.Citation{
		{Law: "LIVA", Article: "1", Note: "tasa general del 16 %"},
		{Law: "LIVA", Article: "1-B", Note: "IVA efectivamente cobrado"},
	}, lines[0].Basis)
	assert.Equal(t, "records", lines[0].Source.String())
	assert.Equal(t, cents(t, "1066.67"), lines[3].Amount)
	assert.Equal(t, "computed", lines[4].Source.String())
}

func TestComputeIVAOptionalLines(t *testing.T) {
	l := domain.NewLedger()
	_, err := domain.ComputeIVA(l, domain.IVAInput{
		Sums: domain.MonthSums{
			Income: domain.IncomeSums{
				ByRate: domain.IVAByRate{
					Rate8: domain.RateSums{Base: 100_00, IVA: 8_00}, Rate0Base: 50_00, ExemptBase: 70_00,
				},
				IVAWithheld: domain.Withheld{ByPlatform: 4_00},
			},
			Expenses: domain.ExpenseSums{IVAWithheldToRemit: 3_00},
		},
		Definitive: func(domain.IncomeSums) (money.Cents, error) { return 4_00, nil },
	})
	require.NoError(t, err)
	lines, err := l.Lines()
	require.NoError(t, err)

	assert.Equal(t, []domain.LineCode{
		domain.LineIVAChargedRate16, domain.LineIVAChargedRate8, domain.LineIVABaseRate0,
		domain.LineIVABaseExempt, domain.LineIVACharged, domain.LineIVADefinitiveExcluded,
		domain.LineIVACreditable, domain.LineIVAWithheldByMoral, domain.LineIVAWithheldByPlatform,
		domain.LineIVANet, domain.LineIVAWithheldToRemit,
	}, codes(lines))

	border := lineByCode(t, lines, domain.LineIVAChargedRate8)
	assert.Equal(t, domain.OpAdd, border.Operation)
	assert.Equal(t, money.Cents(8_00), border.Amount)
	assert.Contains(t, border.Basis[0].Law, "región fronteriza")

	for code, want := range map[domain.LineCode]money.Cents{domain.LineIVABaseRate0: 50_00, domain.LineIVABaseExempt: 70_00} {
		info := lineByCode(t, lines, code)
		assert.Equal(t, domain.OpInfo, info.Operation, "%s", code)
		assert.Equal(t, want, info.Amount, "%s", code)
		assert.NotEmpty(t, info.Basis, "%s cites the article anyway", code)
	}

	excluded := lineByCode(t, lines, domain.LineIVADefinitiveExcluded)
	assert.True(t, excluded.NeedsCPAReview)
	assert.Equal(t, []domain.Citation{
		{Law: "LIVA", Article: "18-L"}, {Law: "LIVA", Article: "18-M", Note: "retención definitiva"},
	}, excluded.Basis)

	remit := lineByCode(t, lines, domain.LineIVAWithheldToRemit)
	assert.Equal(t, domain.TaxWithholding, remit.Tax)
	assert.Equal(t, domain.OpInput, remit.Operation)
	assert.Equal(t, money.Cents(3_00), remit.Amount)
	assert.Equal(t, "LIVA", remit.Basis[0].Law)
	assert.Equal(t, "1-A", remit.Basis[0].Article)
}

func TestComputeIVAMixedActivity(t *testing.T) {
	notFully := domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		domain.OptionIVAFullyTaxable: domain.BoolOption(false),
	})
	fully := domain.NewOptions(map[domain.OptionKey]domain.OptionValue{
		domain.OptionIVAFullyTaxable: domain.BoolOption(true),
	})
	taxed := domain.IncomeSums{ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 100_00, IVA: 16_00}}}
	withExempt := taxed
	withExempt.ByRate.ExemptBase = 100_00
	creditable := domain.ExpenseSums{CreditableIVAPaid: 5_00}

	tests := []struct {
		name    string
		options domain.Options
		sums    domain.MonthSums
		want    []domain.DisclaimerCode
	}{
		{name: "option false with creditable IVA", options: notFully, sums: domain.MonthSums{Income: taxed, Expenses: creditable},
			want: []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity}},
		{name: "option false without creditable IVA", options: notFully, sums: domain.MonthSums{Income: taxed}},
		{name: "option unset defaults to fully taxable", sums: domain.MonthSums{Income: taxed, Expenses: creditable}},
		{name: "option true", options: fully, sums: domain.MonthSums{Income: taxed, Expenses: creditable}},
		{name: "exempt collections with creditable IVA", options: fully, sums: domain.MonthSums{Income: withExempt, Expenses: creditable},
			want: []domain.DisclaimerCode{domain.DisclaimerIVAMixedActivity}},
		{name: "exempt collections without creditable IVA", sums: domain.MonthSums{Income: withExempt}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ComputeIVA(domain.NewLedger(), domain.IVAInput{Sums: tt.sums, Options: tt.options})
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Disclaimers)
			assert.Equal(t, tt.sums.Expenses.CreditableIVAPaid, got.Components.CreditablePaid,
				"credited in full, never prorated (LIVA 5 fr. V deferred)")
		})
	}
}

func TestComputeIVADefinitiveHook(t *testing.T) {
	sums := domain.MonthSums{Income: domain.IncomeSums{
		ByRate: domain.IVAByRate{Rate16: domain.RateSums{Base: 100_00, IVA: 16_00}},
	}}
	hookErr := errors.New("625 rule failed")
	tests := []struct {
		name    string
		rule    domain.DefinitiveIVARule
		wantErr error
	}{
		{name: "negative", rule: func(domain.IncomeSums) (money.Cents, error) { return -1, nil }, wantErr: domain.ErrInvalidIVAExclusion},
		{name: "above charged", rule: func(domain.IncomeSums) (money.Cents, error) { return 16_01, nil }, wantErr: domain.ErrInvalidIVAExclusion},
		{name: "rule error", rule: func(domain.IncomeSums) (money.Cents, error) { return 0, hookErr }, wantErr: hookErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.ComputeIVA(domain.NewLedger(), domain.IVAInput{Sums: sums, Definitive: tt.rule})
			require.ErrorIs(t, err, tt.wantErr)
		})
	}

	t.Run("zero exclusion adds no line and no topic", func(t *testing.T) {
		l := domain.NewLedger()
		got, err := domain.ComputeIVA(l, domain.IVAInput{
			Sums: sums, Definitive: func(domain.IncomeSums) (money.Cents, error) { return 0, nil },
		})
		require.NoError(t, err)
		assert.Empty(t, got.CPATopics)
		lines, err := l.Lines()
		require.NoError(t, err)
		assert.NotContains(t, codes(lines), domain.LineIVADefinitiveExcluded)
	})
}

func TestComputeIVAFailures(t *testing.T) {
	for name, income := range map[string]domain.IncomeSums{
		"charged IVA beyond Exact": {ByRate: domain.IVAByRate{Rate16: domain.RateSums{IVA: math.MaxInt64}}},
		"charged IVA sum":          {ByRate: domain.IVAByRate{Rate16: domain.RateSums{IVA: math.MaxInt64}, Rate8: domain.RateSums{IVA: 1}}},
		"withheld IVA sum":         {IVAWithheld: domain.Withheld{ByMoral: math.MaxInt64, ByPlatform: 1}},
	} {
		t.Run("overflowing "+name, func(t *testing.T) {
			_, err := domain.ComputeIVA(domain.NewLedger(), domain.IVAInput{Sums: domain.MonthSums{Income: income}})
			require.ErrorIs(t, err, money.ErrOverflow)
		})
	}
	t.Run("a ledger that already failed", func(t *testing.T) {
		l := domain.NewLedger()
		l.Add(step("isr.orphan", domain.SourceComputed()), 1)
		_, err := domain.ComputeIVA(l, domain.IVAInput{})
		require.ErrorIs(t, err, domain.ErrNoSegment)
	})
}

// TestComputeIVAIsDeterministic runs the same input twice: same lines,
// same result (no map-order dependence).
func TestComputeIVAIsDeterministic(t *testing.T) {
	in := domain.IVAInput{Sums: domain.MonthSums{
		Income: domain.IncomeSums{
			ByRate:      domain.IVAByRate{Rate16: domain.RateSums{Base: 1, IVA: 2}, Rate8: domain.RateSums{IVA: 3}, Rate0Base: 4, ExemptBase: 5},
			IVAWithheld: domain.Withheld{ByMoral: 1, ByPlatform: 1},
		},
		Expenses: domain.ExpenseSums{CreditableIVAPaid: 1, IVAWithheldToRemit: 1},
	}}
	run := func() ([]domain.Line, domain.IVAResult) {
		l := domain.NewLedger()
		r, err := domain.ComputeIVA(l, in)
		require.NoError(t, err)
		lines, err := l.Lines()
		require.NoError(t, err)
		return lines, r
	}
	l1, r1 := run()
	l2, r2 := run()
	assert.Equal(t, l1, l2)
	assert.Equal(t, r1, r2)
}
