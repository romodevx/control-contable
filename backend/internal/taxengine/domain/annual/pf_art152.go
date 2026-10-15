package annual

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The parameters pf_art152 reads (params/data/<year>), each on 31 December
// of the fiscal year.
const (
	paramArt152Tariff      taxparams.Key = "isr.art152.annual"
	paramRentBlind         taxparams.Key = "isr.rent.blind_deduction"
	paramSalaryThreshold                 = taxparams.KeySalaryAnnualThreshold
	paramInformativeIncome               = taxparams.KeyInformativeIncomeThreshold
)

// The line codes of pf_art152, in the order they can appear: the business
// chapter (612 and 625 non-definitive), the rent chapter (606), the sum of
// the chapters, the personal deductions (after their notes), the tariff,
// the credits and the balance.
const (
	LineArt152BusinessIncome     domain.LineCode = "annual.business.income"
	LineArt152BusinessDeductions domain.LineCode = "annual.business.deductions"
	LineArt152Depreciation       domain.LineCode = "annual.business.depreciation"
	LineArt152PTUPaid            domain.LineCode = "annual.business.ptu_paid"
	LineArt152PriorYearLosses    domain.LineCode = "annual.business.prior_year_losses"
	LineArt152BusinessResult     domain.LineCode = "annual.business.result"
	LineArt152BusinessLoss       domain.LineCode = "annual.business.loss"
	LineArt152RentIncome         domain.LineCode = "annual.rent.income"
	LineArt152RentBlind          domain.LineCode = "annual.rent.blind_deduction"
	LineArt152RentPredial        domain.LineCode = "annual.rent.predial"
	LineArt152RentDeductions     domain.LineCode = "annual.rent.deductions"
	LineArt152RentResult         domain.LineCode = "annual.rent.result"
	LineArt152RentLoss           domain.LineCode = "annual.rent.loss"
	LineArt152Salary             domain.LineCode = "annual.accumulable.salary"
	LineArt152BusinessTaxable    domain.LineCode = "annual.accumulable.business"
	LineArt152RentTaxable        domain.LineCode = "annual.accumulable.rent"
	LineArt152AccumulableIncome  domain.LineCode = "annual.accumulable_income"
	LineArt152PersonalDeductions domain.LineCode = "annual.personal_deductions"
	LineArt152Base               domain.LineCode = "annual.base"
	LineArt152NoBase             domain.LineCode = "annual.no_base"
	LineArt152TariffRow          domain.LineCode = "annual.tariff_row"
	LineArt152TariffExcess       domain.LineCode = "annual.tariff_excess"
	LineArt152ISR                domain.LineCode = "annual.isr"
	LineArt152Payments           domain.LineCode = "annual.credit.payments"
	LineArt152Withheld           domain.LineCode = "annual.credit.withheld"
	LineArt152SalaryISR          domain.LineCode = "annual.credit.salary_isr"
	LineArt152Balance            domain.LineCode = "annual.balance"
)

// The missing-input names of pf_art152 (MODULE-annual §2.6); a default
// payment from an incomplete estimate is filed_payment:<regime>:<YYYY-MM>.
const (
	MissingSalaryStatement            = "salary_statement"
	MissingPriorYearAccumulableIncome = "prior_year_accumulable_income"
)

// The legal basis of each pf_art152 step (docs/research/annual-return-2026.md §B.1).
var (
	basisBusinessIncome = []domain.Citation{
		{Law: lawLISR, Article: "109", Note: "primer párrafo: ingresos efectivamente percibidos en el ejercicio"},
	}
	basisPlatformIncome = []domain.Citation{
		{Law: lawLISR, Article: "113-A", Note: "tercer párrafo: ingresos por plataformas sin la opción definitiva"},
		{Law: lawLISR, Article: "109", Note: "se calculan como actividad empresarial (sujeto a validación)"},
	}
	basisBusinessDeductions = []domain.Citation{{Law: lawLISR, Article: "109", Note: "primer párrafo: deducciones autorizadas efectivamente erogadas"}}
	basisDepreciation       = []domain.Citation{{Law: lawLISR, Article: "104", Note: "deducción de inversiones del ejercicio"}}
	basisPTUPaid            = []domain.Citation{{Law: lawLISR, Article: "109", Note: "fracción I: PTU pagada en el ejercicio"}}
	basisPriorLosses        = []domain.Citation{{Law: lawLISR, Article: "109", Note: "fracción I: pérdidas fiscales de ejercicios anteriores pendientes"}}
	basisBusinessResult     = []domain.Citation{{Law: lawLISR, Article: "109", Note: "utilidad gravable; la pérdida se amortiza en ejercicios siguientes"}}
	basisRentIncome         = []domain.Citation{{Law: lawLISR, Article: "114", Note: "ingresos por arrendamiento cobrados en el ejercicio"}}
	basisRentBlind          = []domain.Citation{{Law: lawLISR, Article: "115", Note: "segundo párrafo: deducción opcional del 35 % más el impuesto predial"}}
	basisRentActual         = []domain.Citation{{Law: lawLISR, Article: "115", Note: "fracciones I a V: deducciones autorizadas"}}
	basisRentResult         = []domain.Citation{{Law: lawLISR, Article: "152", Note: "primer párrafo: ingresos del Capítulo III después de deducciones"}}
	basisSalary             = []domain.Citation{
		{Law: lawLISR, Article: "94", Note: "ingresos por salarios acumulables según la constancia"},
		{Law: lawLISR, Article: "152", Note: "primer párrafo: suma de los ingresos de los Capítulos I, II Sección I y III"},
	}
	basisSum                = []domain.Citation{{Law: lawLISR, Article: "152", Note: "primer párrafo: suma de los ingresos de los Capítulos I, II Sección I y III"}}
	basisPersonalDeductions = []domain.Citation{
		{Law: lawLISR, Article: "151", Note: "deducciones personales con sus límites"},
		{Law: "Decreto DOF 26-12-2013", Article: "1.8", Note: "colegiaturas, fuera del límite global"},
	}
	basisBase    = []domain.Citation{{Law: lawLISR, Article: "152", Note: "primer párrafo: base del impuesto anual"}}
	basisNoBase  = []domain.Citation{{Law: lawLISR, Article: "152", Note: "sin base gravable no hay impuesto"}}
	basisTariff  = []domain.Citation{{Law: lawLISR, Article: "152", Note: "tarifa anual"}, {Law: "RMF 2026", Article: "Anexo 8", Note: "apartado C.II"}}
	basisISR     = []domain.Citation{{Law: lawLISR, Article: "152", Note: "impuesto del ejercicio"}}
	basisPayment = []domain.Citation{
		{Law: lawLISR, Article: "152", Note: "fracción I: se acreditan los pagos provisionales efectivamente enterados"},
	}
	basisWithheld = []domain.Citation{
		{Law: lawLISR, Article: "106", Note: "último párrafo: retención del 10 % por personas morales"},
		{Law: lawLISR, Article: "116", Note: "retención del 10 % sobre arrendamiento"},
		{Law: lawLISR, Article: "113-A", Note: "tercer párrafo: retenciones de plataformas como pagos provisionales"},
	}
	basisSalaryISR = []domain.Citation{{Law: lawLISR, Article: "96", Note: "primer párrafo: retenciones de salarios a cuenta del impuesto anual"}}
	basisBalance   = []domain.Citation{{Law: lawLISR, Article: "152", Note: "saldo a cargo o a favor; solo se devuelve lo efectivamente pagado o retenido"}}
)

// PFArt152 is the pf_art152 section (MODULE-annual §2.7, research §B.1):
// one art. 152 computation for every chapter a persona física holds —
// salary (605, from the constancias), business (612, plus 625 without the
// definitive option, CPA topic platform_nondefinitive_annual) and rent
// (606). Each chapter is floored at zero on its own (a 612 loss stays in
// 612, LISR 109); their sum less the personal deductions is the base of
// the annual tariff; the provisional payments (by source), the ISR
// withheld from the taxpayer and the salary ISR are credited.
type PFArt152 struct{ deductions Calculator }

// NewPFArt152 returns the strategy with its personal-deductions calculator.
func NewPFArt152(c Calculator) PFArt152 { return PFArt152{deductions: c} }

// Kind is pf_art152.
func (PFArt152) Kind() SectionKind { return SectionPFArt152 }

// PFArt152Result is the section and the estimate-level LISR 150 ¶3 fact it
// decides: total income, exempt and definitively taxed included, above the
// informative threshold.
type PFArt152Result struct {
	Section                   Section
	InformativeIncomeOver500k bool
}

// art152Chapters are what the chapters of one year hold.
type art152Chapters struct {
	salaryHeld bool
	business   []RegimeYear // 612, then 625 non-definitive
	rent       *RegimeYear
}

func chaptersOf(in PFYear) art152Chapters {
	c := art152Chapters{salaryHeld: in.holds(regimeSalary) || len(in.SalaryStatements) > 0}
	for _, code := range []fiscal.RegimeCode{fiscal.RegimeBusiness, fiscal.RegimePlatforms} {
		if r, ok := in.regime(code); ok && !r.platformDefinitive() {
			c.business = append(c.business, r)
		}
	}
	if r, ok := in.regime(fiscal.RegimeRent); ok {
		c.rent = &r
	}
	return c
}

// regimes are the codes the section covers, ascending.
func (c art152Chapters) regimes() []fiscal.RegimeCode {
	var out []fiscal.RegimeCode
	if c.salaryHeld {
		out = append(out, regimeSalary)
	}
	for _, r := range c.business {
		out = append(out, r.Regime)
	}
	if c.rent != nil {
		out = append(out, c.rent.Regime)
	}
	slices.Sort(out)
	return out
}

// withPayments are the regimes whose provisional payments are credited.
func (c art152Chapters) withPayments() []RegimeYear {
	out := slices.Clone(c.business)
	if c.rent != nil {
		out = append(out, *c.rent)
	}
	slices.SortFunc(out, func(a, b RegimeYear) int { return cmp.Compare(a.Regime, b.Regime) })
	return out
}

// Applies reports a year holding any art. 152 chapter: 605 (or a
// constancia, OQ-A2), 606, 612 or 625 without the definitive option.
func (PFArt152) Applies(in PFYear) bool { return len(chaptersOf(in).regimes()) > 0 }

// Compute traces the section. It is pure: the params, the year and the
// payments port only.
func (s PFArt152) Compute(in PFYear, payments PaymentsReader, p *taxparams.Params) (PFArt152Result, error) {
	out, err := s.compute(in, payments, p)
	if err != nil {
		return PFArt152Result{}, fmt.Errorf("%s %d: %w", SectionPFArt152, in.FiscalYear, err)
	}
	return out, nil
}

// art152Run is one computation's state.
type art152Run struct {
	*annualTrace
	in       PFYear
	chapters art152Chapters
	day      fiscal.Date
	params   *taxparams.Params
	missing  []string
	topics   []CPATopic
	notes    []DisclaimerCode
	tax      Tax
}

func (s PFArt152) compute(in PFYear, payments PaymentsReader, p *taxparams.Params) (PFArt152Result, error) {
	if err := in.validate(); err != nil {
		return PFArt152Result{}, err
	}
	day, err := fiscal.ParseDate(fmt.Sprintf("%04d-12-31", in.FiscalYear))
	if err != nil {
		return PFArt152Result{}, fmt.Errorf("annual: fiscal year: %w", err)
	}
	r := &art152Run{annualTrace: newAnnualTrace(), in: in, chapters: chaptersOf(in), day: day, params: p}
	accumulable, err := r.accumulable()
	if err != nil {
		return PFArt152Result{}, err
	}
	if err := r.personalDeductions(s.deductions, accumulable); err != nil {
		return PFArt152Result{}, err
	}
	if err := r.annualTax(); err != nil {
		return PFArt152Result{}, err
	}
	if err := r.credits(payments); err != nil {
		return PFArt152Result{}, err
	}
	return r.result()
}

// accumulable traces each chapter and their sum (§B.1 steps 1–6) and
// returns the ingresos acumulables.
func (r *art152Run) accumulable() (money.Cents, error) {
	var parts []art152Part
	if r.chapters.salaryHeld {
		parts = append(parts, r.salary())
	}
	if len(r.chapters.business) > 0 {
		parts = append(parts, r.business())
	}
	if r.chapters.rent != nil {
		rent, err := r.rentChapter(*r.chapters.rent)
		if err != nil {
			return 0, err
		}
		parts = append(parts, rent)
	}
	if len(parts) == 0 {
		parts = append(parts, art152Part{step: annualStep(LineArt152Salary, "annual_salary", domain.SourceRecords(), basisSalary)})
	}
	for i, part := range parts {
		if i == 0 {
			r.l.Input(part.step, part.amount)
			continue
		}
		r.l.Add(part.step, part.amount)
	}
	total := r.l.Result(annualStep(LineArt152AccumulableIncome, "annual_accumulable_income", domain.SourceComputed(), basisSum))
	r.tax.AccumulableIncome = total.Round()
	return r.tax.AccumulableIncome, r.err
}

// art152Part is one chapter's contribution to the sum.
type art152Part struct {
	step   domain.Step
	amount money.Exact
}

// salary is Σ accumulable of the constancias (§B.1 step 1); a 605 without
// any constancia counts 0 and the section is incomplete (§2.6).
func (r *art152Run) salary() art152Part {
	if len(r.in.SalaryStatements) == 0 {
		r.missing = append(r.missing, MissingSalaryStatement)
	}
	var total money.Cents
	for _, st := range r.in.SalaryStatements {
		total = r.sum(total, st.Accumulable)
	}
	args := strconv.Itoa(len(r.in.SalaryStatements))
	return art152Part{step: annualStep(LineArt152Salary, "annual_salary", domain.SourceRecords(), basisSalary, args), amount: r.exact(total)}
}

// business traces 612 and 625 non-definitive as one Cap. II result (§B.1
// steps 2–3) and contributes max(0, result): a loss stays in the chapter.
func (r *art152Run) business() art152Part {
	for i, y := range r.chapters.business {
		platform := y.Regime == fiscal.RegimePlatforms
		income := r.chapterStep(LineArt152BusinessIncome, "annual_business_income", domain.SourceRecords(), basisBusinessIncome, y.Regime, platform)
		if platform {
			income.Basis = slices.Clone(basisPlatformIncome)
			r.topics = append(r.topics, TopicPlatformNonDefinitiveAnnual)
		}
		if i == 0 {
			r.l.Input(income, r.exact(y.Income))
		} else {
			r.l.Add(income, r.exact(y.Income))
		}
		r.l.Subtract(r.chapterStep(LineArt152BusinessDeductions, "annual_business_deductions", domain.SourceRecords(), basisBusinessDeductions, y.Regime, platform), r.exact(y.Deductions))
		r.subtractInput(y, LineArt152Depreciation, "annual_depreciation", domain.InputKey("annual_depreciation"), basisDepreciation, y.Depreciation)
		r.subtractInput(y, LineArt152PTUPaid, "annual_ptu_paid", domain.InputPTUPaid, basisPTUPaid, y.PTUPaid)
		r.subtractInput(y, LineArt152PriorYearLosses, "annual_prior_year_losses", domain.InputPriorYearLosses, basisPriorLosses, y.PriorYearLosses)
	}
	result := r.l.Result(annualStep(LineArt152BusinessResult, "annual_business_result", domain.SourceComputed(), basisBusinessResult))
	return r.floored(result, LineArt152BusinessLoss, "annual_business_loss",
		annualStep(LineArt152BusinessTaxable, "annual_business_taxable", domain.SourceComputed(), basisSum))
}

// chapterStep is a step of a regime's figure; a 625 step needs CPA review.
func (r *art152Run) chapterStep(code domain.LineCode, label domain.LabelKey, src domain.Source, basis []domain.Citation, regime fiscal.RegimeCode, platform bool) domain.Step {
	s := annualStep(code, label, src, basis, string(regime))
	s.NeedsCPAReview = platform
	return s
}

// subtractInput subtracts a typed tax input when the user entered one.
func (r *art152Run) subtractInput(y RegimeYear, code domain.LineCode, label domain.LabelKey, key domain.InputKey, basis []domain.Citation, v money.Cents) {
	if v == 0 {
		return
	}
	s := r.chapterStep(code, label, domain.SourceTaxInput(key), basis, y.Regime, y.Regime == fiscal.RegimePlatforms)
	r.l.Subtract(s, r.exact(v))
}

// floored returns the chapter's contribution max(0, result); a negative
// result is shown as the chapter's loss, carried forward in its chapter.
func (r *art152Run) floored(result money.Exact, lossCode domain.LineCode, lossLabel domain.LabelKey, taxable domain.Step) art152Part {
	if result >= 0 {
		return art152Part{step: taxable, amount: result}
	}
	r.l.Info(annualStep(lossCode, lossLabel, domain.SourceComputed(), basisBusinessResult), -result)
	return art152Part{step: taxable, amount: 0}
}

// rentChapter traces 606 (§B.1 step 4): rent collected less the blind
// deduction and the predial, or the actual deductions, per the regime's
// deduction_mode; it contributes max(0, result).
func (r *art152Run) rentChapter(y RegimeYear) (art152Part, error) {
	blind, err := rentBlind(y.Options)
	if err != nil {
		return art152Part{}, err
	}
	regime := string(y.Regime)
	income := r.exact(y.Income)
	r.l.Input(annualStep(LineArt152RentIncome, "annual_rent_income", domain.SourceRecords(), basisRentIncome, regime), income)
	if blind {
		rate, err := r.params.Rate(paramRentBlind, r.day)
		if err != nil {
			return art152Part{}, err
		}
		deduction, err := income.MulRate(rate)
		if err != nil {
			return art152Part{}, fmt.Errorf("annual: 606 blind deduction: %w", err)
		}
		r.l.Subtract(annualStep(LineArt152RentBlind, "annual_rent_blind_deduction", domain.SourceParams(paramRentBlind), basisRentBlind, regime, rate.String()), deduction)
		r.l.Subtract(annualStep(LineArt152RentPredial, "annual_rent_predial", domain.SourceRecords(), basisRentBlind, regime), r.exact(y.Predial))
	} else {
		r.l.Subtract(annualStep(LineArt152RentDeductions, "annual_rent_deductions", domain.SourceRecords(), basisRentActual, regime), r.exact(y.Deductions))
	}
	result := r.l.Result(annualStep(LineArt152RentResult, "annual_rent_result", domain.SourceComputed(), basisRentResult, regime))
	return r.floored(result, LineArt152RentLoss, "annual_rent_loss",
		annualStep(LineArt152RentTaxable, "annual_rent_taxable", domain.SourceComputed(), basisSum)), nil
}

// rentBlind reads the 606 deduction_mode as the monthly strategy does:
// absent is the catalog default, the blind deduction.
func rentBlind(o domain.Options) (bool, error) {
	switch mode, set := o.Text(domain.OptionRentDeductionMode); {
	case !set || mode == "blind_35":
		return true, nil
	case mode == "actual":
		return false, nil
	default:
		return false, fmt.Errorf("%q: %w", mode, domain.ErrRentDeductionMode)
	}
}

// personalDeductions runs the ANN-07 calculator (§B.2): its notes become
// info lines, then the applied total is subtracted and the base closed.
func (r *art152Run) personalDeductions(c Calculator, accumulable money.Cents) error {
	got, err := c.Compute(DeductionsInput{
		FiscalYear: r.in.FiscalYear, Deductions: r.in.Deductions,
		AccumulableIncome: accumulable, TotalIncome: r.capIncome(),
		PriorYearAccumulableIncome: r.in.Inputs.PriorYearAccumulableIncome,
	})
	if err != nil {
		return err
	}
	for _, n := range got.Notes {
		r.l.Info(n.Step(), r.exact(n.To))
	}
	if got.Applied > 0 {
		r.l.Subtract(annualStep(LineArt152PersonalDeductions, "annual_personal_deductions", domain.SourceComputed(), basisPersonalDeductions), r.exact(got.Applied))
		r.notes = append(r.notes, DisclaimerDeductionsAsEntered)
	}
	if got.MissingPriorYearIncome {
		r.missing = append(r.missing, MissingPriorYearAccumulableIncome)
	}
	r.topics = append(r.topics, got.CPATopics...)
	if r.in.holds(fiscal.RegimeRESICO) && slices.ContainsFunc(r.in.Deductions, func(d Deduction) bool { return d.Category.Capped() }) {
		r.topics = append(r.topics, TopicDeductionCapRESICOIncome)
	}
	r.tax.PersonalDeductions = got.Applied
	r.tax.Base = r.l.Result(annualStep(LineArt152Base, "annual_base", domain.SourceComputed(), basisBase)).Round()
	return r.err
}

// capIncome is the global cap's total income (§2.7): gross income of the
// art. 152 chapters, salary accumulable and exempt, and the other exempt
// income; RESICO income excluded (OQ-A4).
func (r *art152Run) capIncome() money.Cents {
	total := r.salaryGross()
	for _, y := range r.chapters.business {
		total = r.sum(total, y.Income)
	}
	if r.chapters.rent != nil {
		total = r.sum(total, r.chapters.rent.Income)
	}
	return r.sum(total, r.in.Inputs.OtherExemptIncome)
}

// salaryGross is Σ accumulable + exempt of the constancias.
func (r *art152Run) salaryGross() money.Cents {
	var total money.Cents
	for _, st := range r.in.SalaryStatements {
		total = r.sum(total, st.Accumulable, st.Exempt)
	}
	return total
}

// annualTax applies the annual tariff to a positive base (§B.1 step 9).
func (r *art152Run) annualTax() error {
	base := r.l.Running()
	if base <= 0 {
		r.l.Input(annualStep(LineArt152NoBase, "annual_no_base", domain.SourceComputed(), basisNoBase), 0)
	} else {
		table, err := r.params.Table(paramArt152Tariff, r.day)
		if err != nil {
			return err
		}
		b := r.l.Lookup(annualStep(LineArt152TariffRow, "annual_tariff_row", domain.SourceParams(paramArt152Tariff), basisTariff), table, base)
		r.l.AddProduct(annualStep(LineArt152TariffExcess, "annual_tariff_excess", domain.SourceParams(paramArt152Tariff), basisTariff), b.Excess(base), b.Row.Rate)
	}
	r.tax.ISR = r.l.Result(annualStep(LineArt152ISR, "annual_isr", domain.SourceComputed(), basisISR)).Round()
	return r.err
}

// credits subtracts, in order, the provisional payments of each regime by
// source, the ISR withheld from the taxpayer per regime and the salary ISR
// (§2.5, §B.1 step 10), then closes the balance (step 11).
func (r *art152Run) credits(reader PaymentsReader) error {
	regimes := r.chapters.withPayments()
	for _, y := range regimes {
		got, err := readPayments(reader, r.in.FiscalYear, y.Regime)
		if err != nil {
			return err
		}
		r.tax.Payments = r.sum(r.tax.Payments, r.subtractPayments(got.groups, LineArt152Payments, "annual_credit_payments", basisPayment))
		r.missing = append(r.missing, got.missing...)
		if got.estimated {
			r.notes = append(r.notes, DisclaimerPaymentsEstimated)
		}
	}
	for _, y := range regimes {
		if y.ISRWithheld > 0 {
			r.l.Subtract(r.chapterStep(LineArt152Withheld, "annual_credit_withheld", domain.SourceRecords(), basisWithheld, y.Regime, false), r.exact(y.ISRWithheld))
			r.tax.Withholdings = r.sum(r.tax.Withholdings, y.ISRWithheld)
		}
	}
	for _, st := range r.in.SalaryStatements {
		r.tax.SalaryISR = r.sum(r.tax.SalaryISR, st.ISRWithheld)
	}
	if r.tax.SalaryISR > 0 {
		r.l.Subtract(annualStep(LineArt152SalaryISR, "annual_credit_salary_isr", domain.SourceRecords(), basisSalaryISR), r.exact(r.tax.SalaryISR))
	}
	r.tax.Balance = balanceOf(r.l.Result(annualStep(LineArt152Balance, "annual_balance", domain.SourceComputed(), basisBalance)))
	return r.err
}

func (r *art152Run) result() (PFArt152Result, error) {
	obligation, err := r.obligation()
	if err != nil {
		return PFArt152Result{}, err
	}
	over, err := r.informativeIncomeOver()
	if err != nil {
		return PFArt152Result{}, err
	}
	lines, err := r.lines()
	if err != nil {
		return PFArt152Result{}, err
	}
	s := Section{
		Kind: SectionPFArt152, Regimes: r.chapters.regimes(), Status: StatusComplete, Obligation: obligation,
		MissingInputs: r.missing, CPATopics: sortedTopics(r.topics), Tax: &r.tax, Lines: lines,
	}
	if len(s.MissingInputs) > 0 {
		s.Status = StatusIncomplete
		r.notes = append(r.notes, DisclaimerIncompleteInputs)
	}
	if len(s.CPATopics) > 0 {
		r.notes = append(r.notes, DisclaimerCPAReviewPending)
	}
	if over {
		r.notes = append(r.notes, DisclaimerInformativeIncome)
	}
	s.Disclaimers = slices.Compact(slices.Sorted(slices.Values(r.notes)))
	return PFArt152Result{Section: s, InformativeIncomeOver500k: over}, nil
}

// obligation is the MODULE-annual §2.4 table (LISR 98 fr. III, 150): any
// business or rent income obliges; salary obliges only with one of the
// fr. III cases, and the $400,000 test alone raises its CPA topic.
func (r *art152Run) obligation() (Obligation, error) {
	if slices.ContainsFunc(r.chapters.withPayments(), func(y RegimeYear) bool { return y.Income > 0 }) {
		return ObligationObligated, nil
	}
	if !r.chapters.salaryHeld {
		return ObligationNotObligated, nil
	}
	simultaneous := 0
	for _, st := range r.in.SalaryStatements {
		if st.LeftBeforeYearEnd || st.PayerNotWithholding {
			return ObligationObligated, nil
		}
		if st.Simultaneous {
			simultaneous++
		}
	}
	if simultaneous >= 2 || r.in.Inputs.ToldEmployerWillFile {
		return ObligationObligated, nil
	}
	threshold, err := r.params.Money(paramSalaryThreshold, r.day)
	if err != nil {
		return "", err
	}
	if r.salaryGross() > threshold {
		r.topics = append(r.topics, TopicSalaryObligationBasis)
		return ObligationObligated, nil
	}
	return ObligationOptional, nil
}

// informativeIncomeOver is LISR 150 ¶3: every income of the year — exempt,
// definitively taxed and RESICO included — above the threshold.
func (r *art152Run) informativeIncomeOver() (bool, error) {
	threshold, err := r.params.Money(paramInformativeIncome, r.day)
	if err != nil {
		return false, err
	}
	total := r.sum(r.salaryGross(), r.in.Inputs.OtherExemptIncome)
	for _, y := range r.in.Regimes {
		total = r.sum(total, y.Income)
	}
	return total > threshold, r.err
}
