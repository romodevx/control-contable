package annualgolden

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/golden"
)

var (
	caseID     = regexp.MustCompile(`^AN-[0-9]{2}(-[0-9A-Z]+)*$`)
	caseSource = regexp.MustCompile(`^(research:AN-[0-9]{2}|(cpa|spec):[a-z0-9][a-z0-9._-]*)$`)
	regimeCode = regexp.MustCompile(`^[0-9]{3}$`)
)

// lastMonth is December: a payment month is 1..12, or 0 for a year total.
const lastMonth = 12

// amount is a quoted centavo string (golden.DecodeCents).
type amount struct{ cents money.Cents }

// UnmarshalYAML implements yaml.Unmarshaler.
func (a *amount) UnmarshalYAML(n *yaml.Node) error {
	v, err := golden.DecodeCents(n)
	a.cents = v
	return err
}

func (a *amount) ptr() *money.Cents {
	if a == nil {
		return nil
	}
	v := a.cents
	return &v
}

// rate is a quoted decimal fraction (golden.DecodeRate).
type rate struct{ rate money.Rate }

// UnmarshalYAML implements yaml.Unmarshaler.
func (r *rate) UnmarshalYAML(n *yaml.Node) error {
	v, err := golden.DecodeRate(n)
	r.rate = v
	return err
}

type caseDoc struct {
	ID              string      `yaml:"id"`
	Source          string      `yaml:"source"`
	SkipReason      *string     `yaml:"skip_reason"`
	PendingStrategy bool        `yaml:"pending_strategy"`
	Input           inputDoc    `yaml:"input"`
	Expected        expectedDoc `yaml:"expected"`
}

type inputDoc struct {
	FiscalYear         int            `yaml:"fiscal_year"`
	PersonType         string         `yaml:"person_type"`
	Regimes            []regimeDoc    `yaml:"regimes"`
	Payments           []paymentDoc   `yaml:"payments"`
	SalaryStatements   []salaryDoc    `yaml:"salary_statements"`
	PersonalDeductions []deductionDoc `yaml:"personal_deductions"`
	AnnualInputs       annualDoc      `yaml:"annual_inputs"`
}

type regimeDoc struct {
	Regime      string            `yaml:"regime"`
	Income      amount            `yaml:"income"`
	Deductions  amount            `yaml:"deductions"`
	Predial     amount            `yaml:"predial"`
	ISRWithheld amount            `yaml:"isr_withheld"`
	Options     map[string]string `yaml:"options"`
	TaxInputs   []taxInputDoc     `yaml:"tax_inputs"`
}

type taxInputDoc struct {
	Key   string `yaml:"key"`
	Money amount `yaml:"money"`
}

type paymentDoc struct {
	Regime string `yaml:"regime"`
	Month  int    `yaml:"month"`
	Amount amount `yaml:"amount"`
	Source string `yaml:"source"`
}

type salaryDoc struct {
	Employer            string `yaml:"employer"`
	Accumulable         amount `yaml:"accumulable"`
	Exempt              amount `yaml:"exempt"`
	ISRWithheld         amount `yaml:"isr_withheld"`
	Simultaneous        bool   `yaml:"simultaneous"`
	LeftBeforeYearEnd   bool   `yaml:"left_before_year_end"`
	PayerNotWithholding bool   `yaml:"payer_not_withholding"`
}

type deductionDoc struct {
	Category      string `yaml:"category"`
	Amount        amount `yaml:"amount"`
	PaymentMethod string `yaml:"payment_method"`
	Beneficiary   string `yaml:"beneficiary"`
	DoneeType     string `yaml:"donee_type"`
	Student       string `yaml:"student"`
	TuitionLevel  string `yaml:"tuition_level"`
	Scholarship   amount `yaml:"scholarship"`
}

type annualDoc struct {
	PriorYearAccumulableIncome *amount `yaml:"prior_year_accumulable_income"`
	OtherExemptIncome          amount  `yaml:"other_exempt_income"`
	ToldEmployerWillFile       bool    `yaml:"told_employer_will_file"`
}

type expectedDoc struct {
	InformativeIncomeOver500k *bool        `yaml:"informative_income_over_500k"`
	Disclaimers               []string     `yaml:"disclaimers"`
	Sections                  []sectionDoc `yaml:"sections"`
}

type sectionDoc struct {
	Kind          string       `yaml:"kind"`
	Regimes       []string     `yaml:"regimes"`
	Status        string       `yaml:"status"`
	Obligation    string       `yaml:"obligation"`
	MissingInputs []string     `yaml:"missing_inputs"`
	CPATopics     []string     `yaml:"cpa_topics"`
	Tax           *taxDoc      `yaml:"tax"`
	Difference    *amount      `yaml:"difference"`
	PMExtras      *pmExtrasDoc `yaml:"pm_extras"`
	Lines         []lineDoc    `yaml:"lines"`
}

type taxDoc struct {
	AccumulableIncome  *amount     `yaml:"accumulable_income"`
	PersonalDeductions *amount     `yaml:"personal_deductions"`
	Base               *amount     `yaml:"base"`
	ISR                *amount     `yaml:"isr"`
	Credits            creditsDoc  `yaml:"credits"`
	Balance            *balanceDoc `yaml:"balance"`
}

type creditsDoc struct {
	Payments     *amount `yaml:"payments"`
	Withholdings *amount `yaml:"withholdings"`
	SalaryISR    *amount `yaml:"salary_isr"`
}

type balanceDoc struct {
	Kind   string `yaml:"kind"`
	Amount amount `yaml:"amount"`
}

type pmExtrasDoc struct {
	PTUBase         *amount `yaml:"ptu_base"`
	PTUToDistribute *amount `yaml:"ptu_to_distribute"`
	SuggestedCU     *rate   `yaml:"suggested_cu_next_year"`
}

type lineDoc struct {
	Code   string `yaml:"code"`
	Op     string `yaml:"op"`
	Amount amount `yaml:"amount"`
}

// Decode parses one hand-authored annual case; file is its path under the
// annual golden root. Every rule broken is reported in one error, with the
// monthly harness's error values (golden.ErrSyntax, ErrInexactAmount, ...).
func Decode(data []byte, file string) (Case, error) {
	if !bytes.HasPrefix(data, []byte(golden.HandAuthoredHeader)) {
		return Case{}, fmt.Errorf("%s: first line must start with %q: %w", file, golden.HandAuthoredHeader, golden.ErrNotHandAuthored)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var doc caseDoc
	if err := dec.Decode(&doc); err != nil {
		if errors.Is(err, golden.ErrInexactAmount) {
			return Case{}, fmt.Errorf("%s: %w", file, err)
		}
		return Case{}, fmt.Errorf("%s: %w: %w", file, golden.ErrSyntax, err)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return Case{}, fmt.Errorf("%s: more than one document: %w", file, golden.ErrSyntax)
	}
	c, err := (&converter{}).convert(doc)
	if err != nil {
		return Case{}, fmt.Errorf("%s: %w", file, err)
	}
	c.File = file
	return c, nil
}

// converter turns a caseDoc into a Case, collecting every problem.
type converter struct{ problems []error }

func (cv *converter) fail(format string, args ...any) {
	cv.problems = append(cv.problems, fmt.Errorf(format+": %w", append(args, golden.ErrInvalidCase)...))
}

func (cv *converter) convert(doc caseDoc) (Case, error) {
	c := Case{ID: doc.ID, Source: doc.Source, PendingStrategy: doc.PendingStrategy}
	if !caseID.MatchString(doc.ID) {
		cv.fail("id %q: want AN-<nn> and optional upper-case segments", doc.ID)
	}
	if !caseSource.MatchString(doc.Source) {
		cv.fail("source %q: want research:AN-<nn>, cpa:<ref> or spec:<ref>", doc.Source)
	}
	if doc.SkipReason != nil {
		c.SkipReason = strings.TrimSpace(*doc.SkipReason)
		if c.SkipReason == "" {
			cv.fail("skip_reason is blank: name the unverified rule")
		}
	}
	c.Input = cv.input(doc.Input)
	c.Want = cv.expected(doc.Expected)
	if len(cv.problems) > 0 {
		return Case{}, errors.Join(cv.problems...)
	}
	return c, nil
}

func (cv *converter) input(doc inputDoc) Input {
	in := Input{FiscalYear: doc.FiscalYear, PersonType: fiscal.PersonType(doc.PersonType)}
	if doc.FiscalYear < 1 {
		cv.fail("fiscal_year %d", doc.FiscalYear)
	}
	if !in.PersonType.Valid() {
		cv.fail("person_type %q: want fisica or moral", doc.PersonType)
	}
	for _, r := range doc.Regimes {
		in.Regimes = append(in.Regimes, cv.regime(r))
	}
	for _, p := range doc.Payments {
		in.Payments = append(in.Payments, cv.payment(p, in.Regimes))
	}
	for _, s := range doc.SalaryStatements {
		in.SalaryStatements = append(in.SalaryStatements, cv.salary(s))
	}
	for i, d := range doc.PersonalDeductions {
		in.PersonalDeductions = append(in.PersonalDeductions, cv.deduction(i, d))
	}
	in.AnnualInputs = AnnualInputs{
		PriorYearAccumulableIncome: doc.AnnualInputs.PriorYearAccumulableIncome.ptr(),
		OtherExemptIncome:          cv.nonNegative("annual_inputs.other_exempt_income", doc.AnnualInputs.OtherExemptIncome),
		ToldEmployerWillFile:       doc.AnnualInputs.ToldEmployerWillFile,
	}
	if p := in.AnnualInputs.PriorYearAccumulableIncome; p != nil && *p < 0 {
		cv.fail("annual_inputs.prior_year_accumulable_income is negative")
	}
	return in
}

func (cv *converter) nonNegative(field string, a amount) money.Cents {
	if a.cents < 0 {
		cv.fail("%s is negative", field)
	}
	return a.cents
}

func (cv *converter) regime(d regimeDoc) RegimeInput {
	if !regimeCode.MatchString(d.Regime) {
		cv.fail("regime %q: want a three-digit SAT code", d.Regime)
	}
	r := RegimeInput{
		Regime:      fiscal.RegimeCode(d.Regime),
		Income:      cv.nonNegative("regime "+d.Regime+" income", d.Income),
		Deductions:  cv.nonNegative("regime "+d.Regime+" deductions", d.Deductions),
		Predial:     cv.nonNegative("regime "+d.Regime+" predial", d.Predial),
		ISRWithheld: cv.nonNegative("regime "+d.Regime+" isr_withheld", d.ISRWithheld),
		Options:     d.Options,
	}
	for _, ti := range d.TaxInputs {
		if strings.TrimSpace(ti.Key) == "" {
			cv.fail("regime %s: a tax input has no key", d.Regime)
		}
		r.TaxInputs = append(r.TaxInputs, TaxInput{Key: ti.Key, Money: cv.nonNegative("tax input "+ti.Key, ti.Money)})
	}
	return r
}

func (cv *converter) payment(d paymentDoc, regimes []RegimeInput) Payment {
	p := Payment{Regime: fiscal.RegimeCode(d.Regime), Month: d.Month, Amount: d.Amount.cents, Source: PaymentSource(d.Source)}
	if !slices.ContainsFunc(regimes, func(r RegimeInput) bool { return r.Regime == p.Regime }) {
		cv.fail("payment of regime %q, which the case does not hold", d.Regime)
	}
	if d.Month < 0 || d.Month > lastMonth {
		cv.fail("payment month %d: want 1..12, or none for the year's total", d.Month)
	}
	if !p.Source.Valid() {
		cv.fail("payment source %q: want filed, tax_input or estimated_by_app", d.Source)
	}
	cv.nonNegative("payment amount", d.Amount)
	return p
}

func (cv *converter) salary(d salaryDoc) SalaryStatement {
	return SalaryStatement{
		Employer:            d.Employer,
		Accumulable:         cv.nonNegative("salary accumulable", d.Accumulable),
		Exempt:              cv.nonNegative("salary exempt", d.Exempt),
		ISRWithheld:         cv.nonNegative("salary isr_withheld", d.ISRWithheld),
		Simultaneous:        d.Simultaneous,
		LeftBeforeYearEnd:   d.LeftBeforeYearEnd,
		PayerNotWithholding: d.PayerNotWithholding,
	}
}

func (cv *converter) deduction(i int, d deductionDoc) annual.Deduction {
	ded := annual.Deduction{
		Category: annual.Category(d.Category), Amount: d.Amount.cents,
		PaymentMethod: annual.PaymentMethod(d.PaymentMethod), Beneficiary: annual.Beneficiary(d.Beneficiary),
		DoneeType: annual.DoneeType(d.DoneeType), Student: d.Student, TuitionLevel: annual.TuitionLevel(d.TuitionLevel),
		Scholarship: d.Scholarship.cents,
	}
	if err := ded.Validate(); err != nil {
		cv.problems = append(cv.problems, fmt.Errorf("personal_deductions[%d]: %w: %w", i, golden.ErrInvalidCase, err))
	}
	return ded
}

func (cv *converter) expected(doc expectedDoc) Expected {
	e := Expected{InformativeIncomeOver500k: doc.InformativeIncomeOver500k}
	e.Disclaimers = validList(cv, "disclaimers", doc.Disclaimers, func(s string) annual.DisclaimerCode { return annual.DisclaimerCode(s) })
	if len(doc.Sections) == 0 {
		cv.fail("expected.sections is empty: an estimate always has a section")
	}
	var kinds []annual.SectionKind
	for _, s := range doc.Sections {
		w := cv.section(s)
		if slices.Contains(kinds, w.Kind) && w.Kind != annual.SectionExplanation {
			cv.fail("expected.sections: %s repeated", w.Kind)
		}
		kinds = append(kinds, w.Kind)
		e.Sections = append(e.Sections, w)
	}
	return e
}

func (cv *converter) section(d sectionDoc) SectionWant {
	w := SectionWant{
		Kind: annual.SectionKind(d.Kind), Status: annual.Status(d.Status), Obligation: annual.Obligation(d.Obligation),
		MissingInputs: d.MissingInputs, Difference: d.Difference.ptr(),
	}
	if !w.Kind.Valid() {
		cv.fail("section kind %q", d.Kind)
	}
	if !w.Status.Valid() {
		cv.fail("section %s status %q", d.Kind, d.Status)
	}
	if d.Obligation != "" && !w.Obligation.Valid() {
		cv.fail("section %s obligation %q", d.Kind, d.Obligation)
	}
	if d.Regimes != nil {
		w.Regimes = make([]fiscal.RegimeCode, 0, len(d.Regimes))
		for _, r := range d.Regimes {
			if !regimeCode.MatchString(r) {
				cv.fail("section %s regime %q", d.Kind, r)
			}
			w.Regimes = append(w.Regimes, fiscal.RegimeCode(r))
		}
	}
	w.CPATopics = validList(cv, "cpa_topics", d.CPATopics, func(s string) annual.CPATopic { return annual.CPATopic(s) })
	w.Tax = cv.tax(d.Kind, d.Tax)
	if d.PMExtras != nil {
		w.PMExtras = &PMExtras{PTUBase: d.PMExtras.PTUBase.ptr(), PTUToDistribute: d.PMExtras.PTUToDistribute.ptr()}
		if d.PMExtras.SuggestedCU != nil {
			cu := d.PMExtras.SuggestedCU.rate
			w.PMExtras.SuggestedCU = &cu
		}
	}
	w.Lines = cv.lines(d.Kind, d.Lines)
	return w
}

func (cv *converter) tax(kind string, d *taxDoc) *TaxWant {
	if d == nil {
		return nil
	}
	t := &TaxWant{
		AccumulableIncome: d.AccumulableIncome.ptr(), PersonalDeductions: d.PersonalDeductions.ptr(),
		Base: d.Base.ptr(), ISR: d.ISR.ptr(),
		Payments: d.Credits.Payments.ptr(), Withholdings: d.Credits.Withholdings.ptr(), SalaryISR: d.Credits.SalaryISR.ptr(),
	}
	if d.Balance != nil {
		t.Balance = &Balance{Kind: annual.BalanceKind(d.Balance.Kind), Amount: d.Balance.Amount.cents}
		if !t.Balance.Kind.Valid() || t.Balance.Amount < 0 {
			cv.fail("section %s balance: kind %q must be a_cargo, a_favor or cero and the amount not negative", kind, d.Balance.Kind)
		}
	}
	return t
}

func (cv *converter) lines(kind string, docs []lineDoc) []Line {
	if docs == nil {
		return nil
	}
	lines := make([]Line, 0, len(docs))
	for i, d := range docs {
		l := Line{Code: domain.LineCode(d.Code), Op: domain.Operation(d.Op), Amount: d.Amount.cents}
		if strings.TrimSpace(d.Code) == "" || !l.Op.Valid() {
			cv.fail("section %s lines[%d]: code must be set and op a contract operation (got %q)", kind, i, d.Op)
		}
		lines = append(lines, l)
	}
	return lines
}

// validList converts a list of enum values; nil stays nil (not asserted).
func validList[T interface{ Valid() bool }](cv *converter, field string, raw []string, conv func(string) T) []T {
	if raw == nil {
		return nil
	}
	out := make([]T, 0, len(raw))
	for _, s := range raw {
		v := conv(s)
		if !v.Valid() {
			cv.fail("expected %s: %q is not a spec value", field, s)
			continue
		}
		out = append(out, v)
	}
	return out
}
