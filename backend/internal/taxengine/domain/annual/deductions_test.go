package annual_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

const fiscalYear = 2026

// fakeLimits is the LimitsReader port; ANN-04's parameter file replaces it.
type fakeLimits struct {
	limits annual.Limits
	err    error
	year   int
}

func (f *fakeLimits) DeductionLimits(year int) (annual.Limits, error) {
	f.year = year
	return f.limits, f.err
}

func pesos(t testing.TB, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

func rate(t testing.TB, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

// limits2026 are the research values (docs/research/annual-return-2026.md
// §B.2): UMA anual 42,794.64; the ANN-04 parameter file will carry them.
func limits2026(t testing.TB) annual.Limits {
	t.Helper()
	return annual.Limits{
		AnnualUMA:               pesos(t, "42794.64"),
		GlobalCapUMAs:           5,
		GlobalCapIncomeShare:    rate(t, "0.15"),
		FuneralCapUMAs:          1,
		DonationShare:           rate(t, "0.07"),
		GovernmentDonationShare: rate(t, "0.04"),
		RetirementIncomeShare:   rate(t, "0.10"),
		RetirementCapUMAs:       5,
		TuitionCaps: map[annual.TuitionLevel]money.Cents{
			annual.LevelPreescolar:         pesos(t, "14200.00"),
			annual.LevelPrimaria:           pesos(t, "12900.00"),
			annual.LevelSecundaria:         pesos(t, "19900.00"),
			annual.LevelProfesionalTecnico: pesos(t, "17100.00"),
			annual.LevelBachillerato:       pesos(t, "24500.00"),
		},
	}
}

func calculator(t testing.TB) annual.Calculator {
	t.Helper()
	return annual.NewCalculator(&fakeLimits{limits: limits2026(t)})
}

func electronic(c annual.Category, amount money.Cents) annual.Deduction {
	return annual.Deduction{Category: c, Amount: amount, PaymentMethod: annual.PaymentElectronic, Beneficiary: annual.BeneficiarySelf}
}

func cash(c annual.Category, amount money.Cents) annual.Deduction {
	d := electronic(c, amount)
	d.PaymentMethod = annual.PaymentCash
	return d
}

func donation(donee annual.DoneeType, amount money.Cents) annual.Deduction {
	d := electronic(annual.CategoryDonation, amount)
	d.DoneeType = donee
	return d
}

func tuition(student string, level annual.TuitionLevel, amount, scholarship money.Cents) annual.Deduction {
	d := electronic(annual.CategoryTuition, amount)
	d.Student, d.TuitionLevel, d.Scholarship = student, level, scholarship
	return d
}

func ptr(c money.Cents) *money.Cents { return &c }

func capNote(cap annual.CapKind, c annual.Category, concept string, from, to money.Cents) annual.Note {
	return annual.Note{Kind: annual.NoteCapApplied, Cap: cap, Category: c, Concept: concept, From: from, To: to}
}

// stripBasis compares notes without their citations, which
// TestNotesCiteTheirLegalBasis checks.
func stripBasis(notes []annual.Note) []annual.Note {
	out := make([]annual.Note, len(notes))
	for i, n := range notes {
		n.Basis = nil
		out[i] = n
	}
	return out
}

// TestComputeDeductionsReproducesTheResearchExamples pins the deduction
// part of AN-01, AN-02 and AN-08 (docs/research/annual-return-2026.md
// Part D).
func TestComputeDeductionsReproducesTheResearchExamples(t *testing.T) {
	tests := []struct {
		name        string
		in          annual.DeductionsInput
		wantApplied string
		wantCapped  string
		wantNotes   []annual.Note
	}{
		{
			name: "AN-01 medical under the global cap",
			in: annual.DeductionsInput{
				Deductions:        []annual.Deduction{electronic(annual.CategoryMedical, pesos(t, "40000.00"))},
				AccumulableIncome: pesos(t, "500000.00"), TotalIncome: pesos(t, "900000.00"),
			},
			wantApplied: "40000.00", wantCapped: "40000.00",
		},
		{
			name: "AN-02 six kinds: capped group 105,000, retirement 50,000, primaria 12,900",
			in: annual.DeductionsInput{
				Deductions: []annual.Deduction{
					electronic(annual.CategoryMedical, pesos(t, "30000.00")),
					electronic(annual.CategoryHealthInsurance, pesos(t, "25000.00")),
					electronic(annual.CategoryMortgageRealInterest, pesos(t, "40000.00")),
					donation(annual.DoneeAuthorized, pesos(t, "10000.00")),
					electronic(annual.CategoryRetirementContribution, pesos(t, "50000.00")),
					tuition("Hijo 1", annual.LevelPrimaria, pesos(t, "30000.00"), 0),
				},
				AccumulableIncome: pesos(t, "760000.00"), TotalIncome: pesos(t, "1020000.00"),
				PriorYearAccumulableIncome: ptr(pesos(t, "700000.00")),
			},
			wantApplied: "167900.00", wantCapped: "105000.00",
			wantNotes: []annual.Note{
				capNote(annual.CapTuition, annual.CategoryTuition, "Hijo 1", pesos(t, "30000.00"), pesos(t, "12900.00")),
			},
		},
		{
			name: "AN-08 the 15 % global cap binds: 80,000 claimed, 45,000 applied",
			in: annual.DeductionsInput{
				Deductions: []annual.Deduction{
					electronic(annual.CategoryMedical, pesos(t, "60000.00")),
					electronic(annual.CategoryHealthInsurance, pesos(t, "20000.00")),
				},
				AccumulableIncome: pesos(t, "300000.00"), TotalIncome: pesos(t, "300000.00"),
			},
			wantApplied: "45000.00", wantCapped: "45000.00",
			wantNotes: []annual.Note{
				capNote(annual.CapGlobal, "", "", pesos(t, "80000.00"), pesos(t, "45000.00")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.in.FiscalYear = fiscalYear

			got, err := calculator(t).Compute(tt.in)

			require.NoError(t, err)
			assert.Equal(t, tt.wantApplied, got.Applied.String())
			assert.Equal(t, tt.wantCapped, got.CappedApplied.String())
			assert.Equal(t, got.Applied, got.CappedApplied+got.UncappedApplied)
			assert.Equal(t, tt.wantNotes, nilIfEmpty(stripBasis(got.Notes)))
			assert.False(t, got.MissingPriorYearIncome)
		})
	}
}

func TestComputeDeductionsReadsTheLimitsOfTheFiscalYear(t *testing.T) {
	reader := &fakeLimits{limits: limits2026(t)}

	_, err := annual.NewCalculator(reader).Compute(annual.DeductionsInput{FiscalYear: fiscalYear})

	require.NoError(t, err)
	assert.Equal(t, fiscalYear, reader.year)
}

// TestCashIsExcludedOnlyWhereTheLawRequiresElectronicPayment covers the
// four cash-restricted categories (LISR 151 fr. I and VII, decree art.
// 1.8) and shows a cash payment stays deductible elsewhere.
func TestCashIsExcludedOnlyWhereTheLawRequiresElectronicPayment(t *testing.T) {
	restricted := []annual.Category{
		annual.CategoryMedical, annual.CategoryMedicalDisability, annual.CategorySchoolTransport, annual.CategoryTuition,
	}
	all := []annual.Category{
		annual.CategoryMedical, annual.CategoryMedicalDisability, annual.CategoryFuneral, annual.CategoryDonation,
		annual.CategoryMortgageRealInterest, annual.CategoryRetirementContribution, annual.CategoryHealthInsurance,
		annual.CategorySchoolTransport, annual.CategoryLocalPayrollTax, annual.CategoryTuition,
	}
	amount := pesos(t, "1000.00")
	for _, c := range all {
		t.Run(string(c), func(t *testing.T) {
			d := cash(c, amount)
			switch c {
			case annual.CategoryDonation:
				d.DoneeType = annual.DoneeAuthorized
			case annual.CategoryTuition:
				d.Student, d.TuitionLevel = "Hija", annual.LevelSecundaria
			}
			in := annual.DeductionsInput{
				FiscalYear: fiscalYear, Deductions: []annual.Deduction{d},
				AccumulableIncome: pesos(t, "500000.00"), TotalIncome: pesos(t, "500000.00"),
				PriorYearAccumulableIncome: ptr(pesos(t, "500000.00")),
			}

			got, err := calculator(t).Compute(in)

			require.NoError(t, err)
			excluded := isRestricted(restricted, c)
			assert.Equal(t, excluded, c.CashRestricted())
			if excluded {
				assert.Zero(t, got.Applied)
				want := annual.Note{Kind: annual.NoteExcludedCash, Category: c, From: amount}
				if c == annual.CategoryTuition {
					want.Concept = "Hija"
				}
				assert.Equal(t, []annual.Note{want}, stripBasis(got.Notes))
				return
			}
			assert.Equal(t, amount, got.Applied)
			assert.Empty(t, got.Notes)
		})
	}
}

func isRestricted(restricted []annual.Category, c annual.Category) bool {
	for _, r := range restricted {
		if r == c {
			return true
		}
	}
	return false
}

// TestOwnCapsAtTheirBoundary checks every own cap one centavo below, at
// and above it (MODULE-annual §3 generated cases).
func TestOwnCapsAtTheirBoundary(t *testing.T) {
	type boundary struct {
		name     string
		build    func(money.Cents) annual.Deduction
		acc      string // ingresos acumulables
		prior    string // prior-year acumulables
		cap      string
		capKind  annual.CapKind
		category annual.Category
		concept  string
		uncapped bool
	}
	cases := []boundary{
		{name: "funeral 1 UMA", build: func(a money.Cents) annual.Deduction { return electronic(annual.CategoryFuneral, a) },
			acc: "1000000.00", cap: "42794.64", capKind: annual.CapFuneral, category: annual.CategoryFuneral},
		{name: "donations 7 % of prior-year income", build: func(a money.Cents) annual.Deduction { return donation(annual.DoneeAuthorized, a) },
			acc: "1000000.00", prior: "1000000.00", cap: "70000.00", capKind: annual.CapDonations, category: annual.CategoryDonation},
		{name: "government donations 4 %", build: func(a money.Cents) annual.Deduction { return donation(annual.DoneeGovernment, a) },
			acc: "1000000.00", prior: "1000000.00", cap: "40000.00", capKind: annual.CapGovernmentDonations, category: annual.CategoryDonation},
		{name: "retirement 10 % of income", build: func(a money.Cents) annual.Deduction {
			return electronic(annual.CategoryRetirementContribution, a)
		}, acc: "1000000.00", cap: "100000.00", capKind: annual.CapRetirement, category: annual.CategoryRetirementContribution, uncapped: true},
		{name: "retirement 5 UMA", build: func(a money.Cents) annual.Deduction {
			return electronic(annual.CategoryRetirementContribution, a)
		}, acc: "3000000.00", cap: "213973.20", capKind: annual.CapRetirement, category: annual.CategoryRetirementContribution, uncapped: true},
	}
	for _, level := range []struct {
		level annual.TuitionLevel
		cap   string
	}{
		{annual.LevelPreescolar, "14200.00"}, {annual.LevelPrimaria, "12900.00"}, {annual.LevelSecundaria, "19900.00"},
		{annual.LevelProfesionalTecnico, "17100.00"}, {annual.LevelBachillerato, "24500.00"},
	} {
		cases = append(cases, boundary{
			name:  "tuition " + string(level.level),
			build: func(a money.Cents) annual.Deduction { return tuition("Alumno", level.level, a, 0) },
			acc:   "1000000.00", cap: level.cap, capKind: annual.CapTuition, category: annual.CategoryTuition,
			concept: "Alumno", uncapped: true,
		})
	}
	for _, tt := range cases {
		limit := pesos(t, tt.cap)
		for _, delta := range []money.Cents{-1, 0, 1} {
			t.Run(tt.name+" "+(limit+delta).String(), func(t *testing.T) {
				in := annual.DeductionsInput{
					FiscalYear: fiscalYear, Deductions: []annual.Deduction{tt.build(limit + delta)},
					AccumulableIncome: pesos(t, tt.acc), TotalIncome: pesos(t, "100000000.00"),
				}
				if tt.prior != "" {
					in.PriorYearAccumulableIncome = ptr(pesos(t, tt.prior))
				}

				got, err := calculator(t).Compute(in)

				require.NoError(t, err)
				want := min(limit+delta, limit)
				assert.Equal(t, want, got.Applied)
				if tt.uncapped {
					assert.Equal(t, want, got.UncappedApplied)
				} else {
					assert.Equal(t, want, got.CappedApplied)
				}
				var wantNotes []annual.Note
				if delta > 0 {
					wantNotes = []annual.Note{capNote(tt.capKind, tt.category, tt.concept, limit+delta, limit)}
				}
				assert.Equal(t, wantNotes, nilIfEmpty(stripBasis(got.Notes)))
			})
		}
	}
}

func nilIfEmpty(notes []annual.Note) []annual.Note {
	if len(notes) == 0 {
		return nil
	}
	return notes
}

// TestGlobalCapAtItsBoundary checks min(5 UMA, 15 % of total income) one
// centavo around each limit.
func TestGlobalCapAtItsBoundary(t *testing.T) {
	tests := []struct {
		name        string
		totalIncome string
		cap         string
	}{
		{"15 % of total income binds", "1000000.00", "150000.00"},
		{"5 UMA anuales binds", "2000000.00", "213973.20"},
		{"15 % rounds half away from zero to the centavo", "100.03", "15.00"},
	}
	for _, tt := range tests {
		limit := pesos(t, tt.cap)
		for _, delta := range []money.Cents{-1, 0, 1} {
			t.Run(tt.name+" "+(limit+delta).String(), func(t *testing.T) {
				in := annual.DeductionsInput{
					FiscalYear:        fiscalYear,
					Deductions:        []annual.Deduction{electronic(annual.CategoryMedical, limit+delta)},
					AccumulableIncome: pesos(t, tt.totalIncome), TotalIncome: pesos(t, tt.totalIncome),
				}

				got, err := calculator(t).Compute(in)

				require.NoError(t, err)
				assert.Equal(t, min(limit+delta, limit), got.CappedApplied)
				var wantNotes []annual.Note
				if delta > 0 {
					wantNotes = []annual.Note{capNote(annual.CapGlobal, "", "", limit+delta, limit)}
				}
				assert.Equal(t, wantNotes, nilIfEmpty(stripBasis(got.Notes)))
			})
		}
	}
}

func TestUncappedCategoriesStayOutsideTheGlobalCap(t *testing.T) {
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			electronic(annual.CategoryMedical, pesos(t, "20000.00")),
			electronic(annual.CategoryMedicalDisability, pesos(t, "90000.00")),
			electronic(annual.CategoryRetirementContribution, pesos(t, "10000.00")),
			tuition("Ana", annual.LevelBachillerato, pesos(t, "24500.00"), 0),
		},
		AccumulableIncome: pesos(t, "100000.00"), TotalIncome: pesos(t, "100000.00"),
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	assert.Equal(t, "15000.00", got.CappedApplied.String())
	assert.Equal(t, "124500.00", got.UncappedApplied.String())
	assert.Equal(t, "139500.00", got.Applied.String())
	assert.Equal(t, "144500.00", got.Claimed.String())
}

func TestTuitionCapIsPerStudentWithTheHigherAmountOfTwoLevels(t *testing.T) {
	tests := []struct {
		name string
		rows []annual.Deduction
		want string
	}{
		{
			name: "two students, one cap each",
			rows: []annual.Deduction{
				tuition("Ana", annual.LevelPrimaria, pesos(t, "20000.00"), 0),
				tuition("Luis", annual.LevelPrimaria, pesos(t, "20000.00"), 0),
			},
			want: "25800.00",
		},
		{
			name: "preescolar then primaria: the higher amount (preescolar 14,200), decree art. 1.10",
			rows: []annual.Deduction{
				tuition("Ana", annual.LevelPreescolar, pesos(t, "8000.00"), 0),
				tuition("Ana", annual.LevelPrimaria, pesos(t, "8000.00"), 0),
			},
			want: "14200.00",
		},
		{
			name: "scholarships reduce the tuition before the cap",
			rows: []annual.Deduction{tuition("Ana", annual.LevelSecundaria, pesos(t, "25000.00"), pesos(t, "10000.00"))},
			want: "15000.00",
		},
		{
			name: "a scholarship above the tuition leaves zero",
			rows: []annual.Deduction{tuition("Ana", annual.LevelSecundaria, pesos(t, "5000.00"), pesos(t, "6000.00"))},
			want: "0.00",
		},
		{
			name: "a cash row of the student is excluded with its scholarship",
			rows: []annual.Deduction{
				tuition("Ana", annual.LevelSecundaria, pesos(t, "5000.00"), 0),
				func() annual.Deduction {
					d := tuition("Ana", annual.LevelSecundaria, pesos(t, "5000.00"), pesos(t, "1000.00"))
					d.PaymentMethod = annual.PaymentCash
					return d
				}(),
			},
			want: "5000.00",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := annual.DeductionsInput{
				FiscalYear: fiscalYear, Deductions: tt.rows,
				AccumulableIncome: pesos(t, "500000.00"), TotalIncome: pesos(t, "500000.00"),
			}

			got, err := calculator(t).Compute(in)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Applied.String())
		})
	}
}

func TestGovernmentDonationsCountInsideTheSevenPercent(t *testing.T) {
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			donation(annual.DoneeGovernment, pesos(t, "50000.00")),
			donation(annual.DoneeAuthorized, pesos(t, "40000.00")),
		},
		AccumulableIncome: pesos(t, "1000000.00"), TotalIncome: pesos(t, "10000000.00"),
		PriorYearAccumulableIncome: ptr(pesos(t, "1000000.00")),
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	assert.Equal(t, "70000.00", got.Applied.String())
	assert.Equal(t, []annual.Note{
		capNote(annual.CapGovernmentDonations, annual.CategoryDonation, "", pesos(t, "50000.00"), pesos(t, "40000.00")),
		capNote(annual.CapDonations, annual.CategoryDonation, "", pesos(t, "80000.00"), pesos(t, "70000.00")),
	}, stripBasis(got.Notes))
}

// TestDonationsWithoutPriorYearIncomeAreExcluded is OQ-A5: the 7 % cap
// cannot be verified, so donations are left out and the input is missing.
func TestDonationsWithoutPriorYearIncomeAreExcluded(t *testing.T) {
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			donation(annual.DoneeAuthorized, pesos(t, "3000.00")),
			donation(annual.DoneeGovernment, pesos(t, "2000.00")),
			electronic(annual.CategoryMedical, pesos(t, "1000.00")),
		},
		AccumulableIncome: pesos(t, "500000.00"), TotalIncome: pesos(t, "500000.00"),
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	assert.True(t, got.MissingPriorYearIncome)
	assert.Equal(t, "1000.00", got.Applied.String())
	assert.Equal(t, []annual.Note{{
		Kind: annual.NoteExcludedUnverified, Category: annual.CategoryDonation, From: pesos(t, "5000.00"),
	}}, stripBasis(got.Notes))
}

func TestNoDonationsNeedNoPriorYearIncome(t *testing.T) {
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear, Deductions: []annual.Deduction{electronic(annual.CategoryMedical, 100)},
		AccumulableIncome: 100000, TotalIncome: 100000,
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	assert.False(t, got.MissingPriorYearIncome)
}

// TestGlobalCapIsSharedInArticleOrder fills the capped categories in the
// order of LISR art. 151, so each category's applied amount is
// deterministic; the legal figure is the total.
func TestGlobalCapIsSharedInArticleOrder(t *testing.T) {
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			electronic(annual.CategoryLocalPayrollTax, pesos(t, "5000.00")),
			electronic(annual.CategoryHealthInsurance, pesos(t, "20000.00")),
			electronic(annual.CategoryMedical, pesos(t, "30000.00")),
		},
		AccumulableIncome: pesos(t, "300000.00"), TotalIncome: pesos(t, "300000.00"),
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	assert.Equal(t, []annual.CategoryTotal{
		{Category: annual.CategoryMedical, Claimed: pesos(t, "30000.00"), Eligible: pesos(t, "30000.00"), Applied: pesos(t, "30000.00")},
		{Category: annual.CategoryHealthInsurance, Claimed: pesos(t, "20000.00"), Eligible: pesos(t, "20000.00"), Applied: pesos(t, "15000.00")},
		{Category: annual.CategoryLocalPayrollTax, Claimed: pesos(t, "5000.00"), Eligible: pesos(t, "5000.00"), Applied: 0},
	}, got.Categories)
}

func TestUMARulesRaiseTheirCPATopic(t *testing.T) {
	dependant := electronic(annual.CategoryMedical, 100)
	dependant.Beneficiary = annual.BeneficiaryDescendant
	tests := []struct {
		name string
		row  annual.Deduction
		want []annual.CPATopic
	}{
		{"funeral", electronic(annual.CategoryFuneral, 100), []annual.CPATopic{annual.TopicUMAForMinimumWage}},
		{"retirement", electronic(annual.CategoryRetirementContribution, 100), []annual.CPATopic{annual.TopicUMAForMinimumWage}},
		{"a dependant's expense", dependant, []annual.CPATopic{annual.TopicUMAForMinimumWage}},
		{"own medical expense", electronic(annual.CategoryMedical, 100), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := annual.DeductionsInput{
				FiscalYear: fiscalYear, Deductions: []annual.Deduction{tt.row}, AccumulableIncome: 100000, TotalIncome: 100000,
			}

			got, err := calculator(t).Compute(in)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got.CPATopics)
		})
	}
}

func TestNotesCiteTheirLegalBasisAndBecomeInfoLines(t *testing.T) {
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			cash(annual.CategoryMedical, pesos(t, "10.00")),
			electronic(annual.CategoryFuneral, pesos(t, "50000.00")),
			donation(annual.DoneeGovernment, pesos(t, "50000.00")),
			electronic(annual.CategoryRetirementContribution, pesos(t, "50000.00")),
			tuition("Ana", annual.LevelPrimaria, pesos(t, "50000.00"), 0),
			electronic(annual.CategoryHealthInsurance, pesos(t, "90000.00")),
		},
		AccumulableIncome: pesos(t, "100000.00"), TotalIncome: pesos(t, "100000.00"),
		PriorYearAccumulableIncome: ptr(pesos(t, "100000.00")),
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	require.Len(t, got.Notes, 6)
	ledger := domain.NewLedger()
	for _, n := range got.Notes {
		require.NotEmpty(t, n.Basis, "%s %s", n.Kind, n.Cap)
		for _, c := range n.Basis {
			assert.True(t, c.Valid())
		}
		exact, err := n.To.Exact()
		require.NoError(t, err)
		ledger.Info(n.Step(), exact)
	}
	lines, err := ledger.Lines()
	require.NoError(t, err)
	require.Len(t, lines, len(got.Notes))
	assert.Equal(t, domain.LineCode("deductions.excluded_cash"), lines[0].Code)
	assert.Equal(t, []string{"medical", "", "10.00", "0.00"}, lines[0].Args)
	assert.Equal(t, domain.LineCode("deductions.cap_applied"), lines[len(lines)-1].Code)
	assert.Equal(t, domain.LabelKey("annual_deduction_cap_applied"), lines[len(lines)-1].Label)
	assert.Equal(t, domain.OpInfo, lines[len(lines)-1].Operation)
}

// TestTuitionNotesCiteTheDecreeArticleOfTheirRule pins the colegiaturas
// citations to the decree DOF 26-12-2013 as published: art. 1.9 states the
// payment means (cheque, transfer, card) and art. 1.10 the cap table per
// level (ANN-05 finding 2).
func TestTuitionNotesCiteTheDecreeArticleOfTheirRule(t *testing.T) {
	paidInCash := tuition("Ana", annual.LevelPrimaria, pesos(t, "100.00"), 0)
	paidInCash.PaymentMethod = annual.PaymentCash
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			paidInCash,
			tuition("Luis", annual.LevelPrimaria, pesos(t, "50000.00"), 0),
		},
		AccumulableIncome: pesos(t, "100000.00"), TotalIncome: pesos(t, "100000.00"),
	}

	got, err := calculator(t).Compute(in)

	require.NoError(t, err)
	cited := map[annual.NoteKind][]string{}
	for _, n := range got.Notes {
		for _, c := range n.Basis {
			cited[n.Kind] = append(cited[n.Kind], c.Law+" art. "+c.Article)
		}
	}
	assert.Equal(t, map[annual.NoteKind][]string{
		annual.NoteExcludedCash: {"Decreto DOF 26-12-2013 art. 1.9"},
		annual.NoteCapApplied:   {"Decreto DOF 26-12-2013 art. 1.10"},
	}, cited)
}

func TestComputeDeductionsRejectsInvalidInput(t *testing.T) {
	valid := func() annual.DeductionsInput {
		return annual.DeductionsInput{
			FiscalYear: fiscalYear, Deductions: []annual.Deduction{electronic(annual.CategoryMedical, 100)},
			AccumulableIncome: 100, TotalIncome: 100,
		}
	}
	withRow := func(mut func(*annual.Deduction)) annual.DeductionsInput {
		in := valid()
		mut(&in.Deductions[0])
		return in
	}
	tests := []struct {
		name string
		in   annual.DeductionsInput
	}{
		{"zero amount", withRow(func(d *annual.Deduction) { d.Amount = 0 })},
		{"negative amount", withRow(func(d *annual.Deduction) { d.Amount = -1 })},
		{"unknown category", withRow(func(d *annual.Deduction) { d.Category = "gym" })},
		{"unknown payment method", withRow(func(d *annual.Deduction) { d.PaymentMethod = "crypto" })},
		{"unknown beneficiary", withRow(func(d *annual.Deduction) { d.Beneficiary = "friend" })},
		{"donee on a non-donation", withRow(func(d *annual.Deduction) { d.DoneeType = annual.DoneeAuthorized })},
		{"donation without donee", withRow(func(d *annual.Deduction) { d.Category = annual.CategoryDonation })},
		{"tuition fields on a non-tuition", withRow(func(d *annual.Deduction) { d.Student = "Ana" })},
		{"scholarship on a non-tuition", withRow(func(d *annual.Deduction) { d.Scholarship = 1 })},
		{"tuition without student", withRow(func(d *annual.Deduction) {
			d.Category, d.TuitionLevel = annual.CategoryTuition, annual.LevelPrimaria
		})},
		{"tuition without level", withRow(func(d *annual.Deduction) { d.Category, d.Student = annual.CategoryTuition, "Ana" })},
		{"negative scholarship", withRow(func(d *annual.Deduction) {
			d.Category, d.Student, d.TuitionLevel, d.Scholarship = annual.CategoryTuition, "Ana", annual.LevelPrimaria, -1
		})},
		{"negative accumulable income", func() annual.DeductionsInput { in := valid(); in.AccumulableIncome = -1; return in }()},
		{"negative total income", func() annual.DeductionsInput { in := valid(); in.TotalIncome = -1; return in }()},
		{"negative prior-year income", func() annual.DeductionsInput {
			in := valid()
			in.PriorYearAccumulableIncome = ptr(-1)
			return in
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calculator(t).Compute(tt.in)

			assert.ErrorIs(t, err, annual.ErrInvalidDeduction)
		})
	}
}

func TestComputeDeductionsRejectsInvalidLimits(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*annual.Limits)
	}{
		{"no UMA", func(l *annual.Limits) { l.AnnualUMA = 0 }},
		{"no global UMA multiple", func(l *annual.Limits) { l.GlobalCapUMAs = 0 }},
		{"income share above 1", func(l *annual.Limits) { l.GlobalCapIncomeShare = money.RateScale + 1 }},
		{"negative donation share", func(l *annual.Limits) { l.DonationShare = -1 }},
		{"government share above the donation share", func(l *annual.Limits) { l.GovernmentDonationShare = l.DonationShare + 1 }},
		{"no funeral multiple", func(l *annual.Limits) { l.FuneralCapUMAs = 0 }},
		{"no retirement multiple", func(l *annual.Limits) { l.RetirementCapUMAs = 0 }},
		{"zero retirement share", func(l *annual.Limits) { l.RetirementIncomeShare = 0 }},
		{"a tuition level without a cap", func(l *annual.Limits) { delete(l.TuitionCaps, annual.LevelBachillerato) }},
		{"a non-positive tuition cap", func(l *annual.Limits) { l.TuitionCaps[annual.LevelPrimaria] = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := limits2026(t)
			tt.mut(&limits)

			_, err := annual.NewCalculator(&fakeLimits{limits: limits}).Compute(annual.DeductionsInput{FiscalYear: fiscalYear})

			assert.ErrorIs(t, err, annual.ErrInvalidLimits)
		})
	}
}

func TestComputeDeductionsWrapsTheLimitsReaderError(t *testing.T) {
	boom := errors.New("no parameter set for 2031")

	_, err := annual.NewCalculator(&fakeLimits{err: boom}).Compute(annual.DeductionsInput{FiscalYear: 2031})

	assert.ErrorIs(t, err, boom)
}

func TestComputeDeductionsReportsOverflow(t *testing.T) {
	const huge = money.Cents(1 << 62)
	in := annual.DeductionsInput{
		FiscalYear: fiscalYear,
		Deductions: []annual.Deduction{
			electronic(annual.CategoryMortgageRealInterest, huge),
			electronic(annual.CategoryMortgageRealInterest, huge),
		},
		AccumulableIncome: 100, TotalIncome: 100,
	}

	_, err := calculator(t).Compute(in)

	assert.ErrorIs(t, err, money.ErrOverflow)
}
