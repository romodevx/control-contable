package annual

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// DeductionsInput is what the calculator reads: the year's registered
// deductions and the incomes the caps depend on.
type DeductionsInput struct {
	FiscalYear int
	Deductions []Deduction
	// AccumulableIncome is the year's ingresos acumulables of the art. 152
	// chapters (the retirement cap's 10 %).
	AccumulableIncome money.Cents
	// TotalIncome is the global cap's "ingresos totales": gross income of
	// the art. 152 chapters plus exempt income; RESICO income excluded
	// (OQ-A4, CPA topic deduction_cap_resico_income, decided by the caller).
	TotalIncome money.Cents
	// PriorYearAccumulableIncome is the prior year's acumulables before
	// personal deductions (the 7 % donation cap); nil when not given.
	PriorYearAccumulableIncome *money.Cents
}

func (in DeductionsInput) validate() error {
	var problems []error
	for i, d := range in.Deductions {
		if err := d.Validate(); err != nil {
			problems = append(problems, fmt.Errorf("deduction %d: %w", i, err))
		}
	}
	if in.AccumulableIncome < 0 || in.TotalIncome < 0 {
		problems = append(problems, fmt.Errorf("incomes must not be negative: %w", ErrInvalidDeduction))
	}
	if in.PriorYearAccumulableIncome != nil && *in.PriorYearAccumulableIncome < 0 {
		problems = append(problems, fmt.Errorf("prior-year income must not be negative: %w", ErrInvalidDeduction))
	}
	return errors.Join(problems...)
}

// CategoryTotal is one category's figures: what was registered, what is
// eligible (cash and unverifiable rows left out) and what is applied. The
// global cap is shared in article order, so a category's applied amount is
// a presentation of the legal total, Deductions.CappedApplied.
type CategoryTotal struct {
	Category Category
	Claimed  money.Cents
	Eligible money.Cents
	Applied  money.Cents
}

// Deductions is the result: the totals, each category with rows (article
// order), the notes (exclusions in row order, then own caps, then the global
// cap), the CPA topics the rules raised, and whether the donations' cap
// input is missing (the section is then incomplete, MODULE-annual §2.6).
type Deductions struct {
	Claimed                money.Cents
	Applied                money.Cents
	CappedApplied          money.Cents
	UncappedApplied        money.Cents
	Categories             []CategoryTotal
	Notes                  []Note
	CPATopics              []CPATopic
	MissingPriorYearIncome bool
}

// Calculator applies LISR art. 151 and the colegiaturas decree to the
// registered deductions (MODULE-annual §2.7, "Personal deductions").
type Calculator struct{ limits LimitsReader }

// NewCalculator returns a calculator reading its limits from r.
func NewCalculator(r LimitsReader) Calculator { return Calculator{limits: r} }

// Compute applies, in the spec's fixed order: (1) eligibility, (2) each
// category's own cap, (3) the global cap on the capped group, (4) the
// uncapped categories. It is pure: no clock, no I/O beyond the limits port.
func (c Calculator) Compute(in DeductionsInput) (Deductions, error) {
	limits, err := c.limits.DeductionLimits(in.FiscalYear)
	if err != nil {
		return Deductions{}, fmt.Errorf("annual: deduction limits %d: %w", in.FiscalYear, err)
	}
	if err := limits.Validate(); err != nil {
		return Deductions{}, fmt.Errorf("annual: deduction limits %d: %w", in.FiscalYear, err)
	}
	if err := in.validate(); err != nil {
		return Deductions{}, err
	}
	r := newRun(limits)
	r.eligibility(in)
	r.ownCaps(in)
	r.globalCap(in.TotalIncome)
	return r.result()
}

// categoryState is one category through the steps.
type categoryState struct {
	CategoryTotal
	allowed money.Cents // after its own cap, before the global cap
}

// student is one tuition beneficiary: the cap is per student.
type student struct {
	tuition, scholarship, cap money.Cents
}

// run is one computation. Sums are checked; the first overflow is kept and
// returned by result, like the Ledger's sticky error.
type run struct {
	limits     Limits
	categories map[Category]*categoryState
	students   map[string]*student
	authorized money.Cents // eligible donations to authorized donees
	government money.Cents // eligible donations to government donees
	notes      []Note
	usesUMA    bool
	missing    bool
	err        error
}

func newRun(l Limits) *run {
	return &run{limits: l, categories: map[Category]*categoryState{}, students: map[string]*student{}}
}

// add sets *dst += v, keeping the first overflow.
func (r *run) add(dst *money.Cents, v money.Cents) {
	if r.err != nil {
		return
	}
	sum, err := dst.Add(v)
	if err != nil {
		r.err = fmt.Errorf("annual: deductions: %w", err)
		return
	}
	*dst = sum
}

// keep records a limit computation's error.
func (r *run) keep(v money.Cents, err error) money.Cents {
	if err != nil && r.err == nil {
		r.err = err
	}
	return v
}

func (r *run) state(c Category) *categoryState {
	s, ok := r.categories[c]
	if !ok {
		s = &categoryState{CategoryTotal: CategoryTotal{Category: c}}
		r.categories[c] = s
	}
	return s
}

// eligibility is step 1: every row is claimed; cash rows of a restricted
// category are excluded with a note; the rest are eligible.
func (r *run) eligibility(in DeductionsInput) {
	for _, d := range in.Deductions {
		s := r.state(d.Category)
		r.add(&s.Claimed, d.Amount)
		if d.Category == CategoryFuneral || d.Category == CategoryRetirementContribution || d.Beneficiary != BeneficiarySelf {
			r.usesUMA = true
		}
		if d.PaymentMethod == PaymentCash && d.Category.CashRestricted() {
			r.notes = append(r.notes, Note{
				Kind: NoteExcludedCash, Category: d.Category, Concept: tuitionStudent(d), From: d.Amount,
				Basis: cashBasis[d.Category],
			})
			continue
		}
		r.add(&s.Eligible, d.Amount)
		r.classify(d)
	}
}

func tuitionStudent(d Deduction) string {
	if d.Category != CategoryTuition {
		return ""
	}
	return strings.TrimSpace(d.Student)
}

// classify routes an eligible row to the sums its own cap needs.
func (r *run) classify(d Deduction) {
	switch d.Category {
	case CategoryDonation:
		if d.DoneeType == DoneeGovernment {
			r.add(&r.government, d.Amount)
			return
		}
		r.add(&r.authorized, d.Amount)
	case CategoryTuition:
		key := tuitionStudent(d)
		st, ok := r.students[key]
		if !ok {
			st = &student{}
			r.students[key] = st
		}
		r.add(&st.tuition, d.Amount)
		r.add(&st.scholarship, d.Scholarship)
		st.cap = max(st.cap, r.limits.TuitionCaps[d.TuitionLevel])
	}
}

// ownCaps is step 2. Categories without an own cap pass through.
func (r *run) ownCaps(in DeductionsInput) {
	for c, s := range r.categories {
		if c != CategoryFuneral && c != CategoryDonation && c != CategoryRetirementContribution && c != CategoryTuition {
			s.allowed = s.Eligible
		}
	}
	if s, ok := r.categories[CategoryFuneral]; ok {
		limit := r.keep(r.limits.umas(r.limits.FuneralCapUMAs))
		s.allowed = r.capTo(CapFuneral, CategoryFuneral, "", s.Eligible, limit, basisFuneral)
	}
	if s, ok := r.categories[CategoryDonation]; ok {
		s.allowed = r.donations(in.PriorYearAccumulableIncome)
		if r.missing {
			s.Eligible = 0
		}
	}
	if s, ok := r.categories[CategoryRetirementContribution]; ok {
		limit := min(
			r.keep(shareOf(in.AccumulableIncome, r.limits.RetirementIncomeShare)),
			r.keep(r.limits.umas(r.limits.RetirementCapUMAs)),
		)
		s.allowed = r.capTo(CapRetirement, CategoryRetirementContribution, "", s.Eligible, limit, basisRetirement)
	}
	if s, ok := r.categories[CategoryTuition]; ok {
		s.allowed = r.tuition()
	}
}

// capTo returns min(amount, limit), with a note when the cap binds.
func (r *run) capTo(k CapKind, c Category, concept string, amount, limit money.Cents, basis []domain.Citation) money.Cents {
	if amount <= limit {
		return amount
	}
	r.notes = append(r.notes, Note{Kind: NoteCapApplied, Cap: k, Category: c, Concept: concept, From: amount, To: limit, Basis: basis})
	return limit
}

// donations applies the 4 % government cap inside the 7 % cap of the
// prior year's acumulables; without that income the donations cannot be
// verified and are excluded (OQ-A5).
func (r *run) donations(prior *money.Cents) money.Cents {
	if prior == nil {
		var total money.Cents
		r.add(&total, r.authorized)
		r.add(&total, r.government)
		r.notes = append(r.notes, Note{Kind: NoteExcludedUnverified, Category: CategoryDonation, From: total, Basis: basisUnverified})
		r.missing = true
		return 0
	}
	government := r.capTo(CapGovernmentDonations, CategoryDonation, "", r.government,
		r.keep(shareOf(*prior, r.limits.GovernmentDonationShare)), basisGovernment)
	var total money.Cents
	r.add(&total, r.authorized)
	r.add(&total, government)
	return r.capTo(CapDonations, CategoryDonation, "", total, r.keep(shareOf(*prior, r.limits.DonationShare)), basisDonations)
}

// tuition caps each student's tuition net of scholarships at the higher
// amount of the levels they had in the year (decree art. 1.10: "el monto
// más alto"), students in name order.
func (r *run) tuition() money.Cents {
	var total money.Cents
	for _, name := range slices.Sorted(maps.Keys(r.students)) {
		st := r.students[name]
		net := max(0, st.tuition-st.scholarship)
		r.add(&total, r.capTo(CapTuition, CategoryTuition, name, net, st.cap, basisTuition))
	}
	return total
}

// globalCap is steps 3 and 4: the capped group is limited to min(5 UMA
// anuales, 15 % of total income) and shared in article order; the uncapped
// categories keep what their own caps allowed.
func (r *run) globalCap(totalIncome money.Cents) {
	var group money.Cents
	for _, c := range categoryOrder {
		if s, ok := r.categories[c]; ok && c.Capped() {
			r.add(&group, s.allowed)
		}
	}
	limit := min(
		r.keep(r.limits.umas(r.limits.GlobalCapUMAs)),
		r.keep(shareOf(totalIncome, r.limits.GlobalCapIncomeShare)),
	)
	remaining := r.capTo(CapGlobal, "", "", group, limit, basisGlobal)
	for _, c := range categoryOrder {
		s, ok := r.categories[c]
		switch {
		case !ok:
		case c.Capped():
			s.Applied = min(s.allowed, remaining)
			remaining -= s.Applied
		default:
			s.Applied = s.allowed
		}
	}
}

func (r *run) result() (Deductions, error) {
	out := Deductions{Notes: r.notes, MissingPriorYearIncome: r.missing}
	for _, c := range categoryOrder {
		s, ok := r.categories[c]
		if !ok {
			continue
		}
		out.Categories = append(out.Categories, s.CategoryTotal)
		r.add(&out.Claimed, s.Claimed)
		r.add(&out.Applied, s.Applied)
		if c.Capped() {
			r.add(&out.CappedApplied, s.Applied)
		} else {
			r.add(&out.UncappedApplied, s.Applied)
		}
	}
	if r.usesUMA {
		out.CPATopics = []CPATopic{TopicUMAForMinimumWage}
	}
	if r.err != nil {
		return Deductions{}, r.err
	}
	return out, nil
}
