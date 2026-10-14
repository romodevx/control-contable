package golden

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// HandAuthoredHeader starts every case file. YAML serializers drop
// comments, so a file a tool re-wrote fails to load (no update flag,
// MODULE-tax-engine US-X3).
const HandAuthoredHeader = "# golden: hand-authored"

const (
	tagString = "!!str"
	tagBool   = "!!bool"
)

var (
	caseID      = regexp.MustCompile(`^(PF|PM|TX)(-[0-9A-Z]+)+$`)
	caseSource  = regexp.MustCompile(`^(research:(PF|PM)(-[0-9A-Z]+)+|cpa:[a-z0-9][a-z0-9._-]*)$`)
	regimeCode  = regexp.MustCompile(`^[0-9]{3}$`)
	centsLexeme = regexp.MustCompile(`^-?[0-9]+\.[0-9]{2}$`)
)

// caseDoc is one case file as written.
type caseDoc struct {
	ID              string      `yaml:"id"`
	Source          string      `yaml:"source"`
	SkipReason      *string     `yaml:"skip_reason"`
	PendingStrategy bool        `yaml:"pending_strategy"`
	Input           inputDoc    `yaml:"input"`
	Expected        expectedDoc `yaml:"expected"`
}

type inputDoc struct {
	Regime       string               `yaml:"regime"`
	PersonType   string               `yaml:"person_type"`
	Period       string               `yaml:"period"`
	ActiveMonths []int                `yaml:"active_months"`
	Options      map[string]optionDoc `yaml:"options"`
	Months       []monthDoc           `yaml:"months"`
	TaxInputs    []taxInputDoc        `yaml:"tax_inputs"`
}

type monthDoc struct {
	Month    int         `yaml:"month"`
	Income   incomeDoc   `yaml:"income"`
	Expenses expensesDoc `yaml:"expenses"`
}

type incomeDoc struct {
	Cash         centsText   `yaml:"cash"`
	Accrual      centsText   `yaml:"accrual"`
	IVACollected centsText   `yaml:"iva_collected"`
	IVAWithheld  withheldDoc `yaml:"iva_withheld"`
	ISRWithheld  withheldDoc `yaml:"isr_withheld"`
}

type withheldDoc struct {
	ByMoral    centsText `yaml:"by_moral"`
	ByPlatform centsText `yaml:"by_platform"`
}

type expensesDoc struct {
	DeductiblePaid     centsText `yaml:"deductible_paid"`
	CreditableIVAPaid  centsText `yaml:"creditable_iva_paid"`
	PredialPaid        centsText `yaml:"predial_paid"`
	ISRWithheldToRemit centsText `yaml:"isr_withheld_to_remit"`
	IVAWithheldToRemit centsText `yaml:"iva_withheld_to_remit"`
}

type taxInputDoc struct {
	Key   string     `yaml:"key"`
	Kind  string     `yaml:"kind"`
	Month int        `yaml:"month"`
	Money *centsText `yaml:"money"`
	Rate  *rateText  `yaml:"rate"`
	Bool  *bool      `yaml:"bool"`
}

type expectedDoc struct {
	Status          string       `yaml:"status"`
	Taxes           []string     `yaml:"taxes"`
	ISR             *centsText   `yaml:"isr"`
	IVA             *ivaDoc      `yaml:"iva"`
	WithheldToRemit *centsText   `yaml:"withheld_to_remit"`
	Lines           []lineDoc    `yaml:"lines"`
	MissingInputs   []missingDoc `yaml:"missing_inputs"`
	CPATopics       []string     `yaml:"cpa_topics"`
	Disclaimers     []string     `yaml:"disclaimers"`
}

type ivaDoc struct {
	ChargedCollected     centsText `yaml:"charged_collected"`
	CreditablePaid       centsText `yaml:"creditable_paid"`
	WithheldFromTaxpayer centsText `yaml:"withheld_from_taxpayer"`
	DefinitiveExcluded   centsText `yaml:"definitive_excluded"`
	// Net is the regime's IVA as the source states it ("IVA to pay
	// 2,133.33"); it must equal charged − excluded − creditable − withheld
	// (MODULE-tax-engine §2.6), which catches an authoring slip.
	Net centsText `yaml:"net"`
}

type lineDoc struct {
	Tax    string    `yaml:"tax"`
	Code   string    `yaml:"code"`
	Op     string    `yaml:"op"`
	Amount centsText `yaml:"amount"`
}

type missingDoc struct {
	Key    string `yaml:"key"`
	Months []int  `yaml:"months"`
}

// centsText is an amount in pesos with exactly two decimals, quoted. The
// zero value (an absent field) is zero.
type centsText struct{ cents money.Cents }

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *centsText) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode || n.ShortTag() != tagString || !centsLexeme.MatchString(n.Value) {
		return fmt.Errorf("line %d: %s %q: %w", n.Line, n.ShortTag(), n.Value, ErrInexactAmount)
	}
	v, err := money.ParseCents(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w: %w", n.Line, ErrInexactAmount, err)
	}
	c.cents = v
	return nil
}

// rateText is a quoted decimal fraction ("0.15").
type rateText struct{ rate money.Rate }

// UnmarshalYAML implements yaml.Unmarshaler.
func (r *rateText) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode || n.ShortTag() != tagString {
		return fmt.Errorf("line %d: %s %q: %w", n.Line, n.ShortTag(), n.Value, ErrInexactAmount)
	}
	v, err := money.ParseRate(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w: %w", n.Line, ErrInexactAmount, err)
	}
	r.rate = v
	return nil
}

// optionDoc is a regime option: a boolean, a string, or a list of strings.
type optionDoc struct{ value domain.OptionValue }

// UnmarshalYAML implements yaml.Unmarshaler.
func (o *optionDoc) UnmarshalYAML(n *yaml.Node) error {
	switch {
	case n.Kind == yaml.ScalarNode && n.ShortTag() == tagBool:
		var b bool
		if err := n.Decode(&b); err != nil {
			return fmt.Errorf("line %d: %w: %w", n.Line, ErrInvalidCase, err)
		}
		o.value = domain.BoolOption(b)
		return nil
	case n.Kind == yaml.ScalarNode && n.ShortTag() == tagString:
		o.value = domain.TextOption(n.Value)
		return nil
	case n.Kind == yaml.SequenceNode:
		items := make([]string, 0, len(n.Content))
		for _, item := range n.Content {
			if item.Kind != yaml.ScalarNode || item.ShortTag() != tagString {
				return fmt.Errorf("line %d: option list item: %w", item.Line, ErrInvalidCase)
			}
			items = append(items, item.Value)
		}
		o.value = domain.ListOption(items)
		return nil
	}
	return fmt.Errorf("line %d: option must be a boolean, a string or a list of strings: %w", n.Line, ErrInvalidCase)
}

// Decode parses one hand-authored case file; file is its path under the
// golden root, used in errors and kept in Case.File. Every rule broken is
// reported in one error.
func Decode(data []byte, file string) (Case, error) {
	if !bytes.HasPrefix(data, []byte(HandAuthoredHeader)) {
		return Case{}, fmt.Errorf("%s: first line must start with %q: %w", file, HandAuthoredHeader, ErrNotHandAuthored)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var doc caseDoc
	if err := dec.Decode(&doc); err != nil {
		if errors.Is(err, ErrInexactAmount) || errors.Is(err, ErrInvalidCase) {
			return Case{}, fmt.Errorf("%s: %w", file, err)
		}
		return Case{}, fmt.Errorf("%s: %w: %w", file, ErrSyntax, err)
	}
	var extra yaml.Node
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return Case{}, fmt.Errorf("%s: more than one document: %w", file, ErrSyntax)
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
	cv.problems = append(cv.problems, fmt.Errorf(format+": %w", append(args, ErrInvalidCase)...))
}

func (cv *converter) check(what string, err error) {
	if err != nil {
		cv.problems = append(cv.problems, fmt.Errorf("%s: %w: %w", what, ErrInvalidCase, err))
	}
}

func (cv *converter) convert(doc caseDoc) (Case, error) {
	c := Case{ID: doc.ID, Source: doc.Source, PendingStrategy: doc.PendingStrategy}
	if !caseID.MatchString(doc.ID) {
		cv.fail("id %q: want PF-/PM-/TX- and upper-case segments", doc.ID)
	}
	if !caseSource.MatchString(doc.Source) {
		cv.fail("source %q: want research:<worked example id> or cpa:<ref>", doc.Source)
	}
	if doc.SkipReason != nil {
		c.SkipReason = strings.TrimSpace(*doc.SkipReason)
		if c.SkipReason == "" {
			cv.fail("skip_reason is blank: name the unverified rule")
		}
	}
	c.Input, c.ActiveMonths = cv.input(doc.Input)
	c.Want = cv.expected(doc.Expected)
	if len(cv.problems) > 0 {
		return Case{}, errors.Join(cv.problems...)
	}
	return c, nil
}

func (cv *converter) input(doc inputDoc) (domain.Input, []time.Month) {
	in := domain.Input{Key: domain.StrategyKey{
		Regime: fiscal.RegimeCode(doc.Regime), PersonType: fiscal.PersonType(doc.PersonType),
	}}
	if !regimeCode.MatchString(doc.Regime) {
		cv.fail("regime %q: want a three-digit SAT code", doc.Regime)
	}
	if !in.Key.PersonType.Valid() {
		cv.fail("person_type %q: want fisica or moral", doc.PersonType)
	}
	period, err := domain.ParsePeriod(doc.Period)
	if err != nil {
		cv.check("period", err)
		return in, nil
	}
	in.Period = period
	in.Options = cv.options(doc.Options)
	in.Sums = cv.sums(period.Month(), doc.Months)
	in.Inputs = cv.taxInputs(doc.TaxInputs)
	return in, cv.activeMonths(period.Month(), doc.ActiveMonths)
}

func (cv *converter) options(docs map[string]optionDoc) domain.Options {
	values := make(map[domain.OptionKey]domain.OptionValue, len(docs))
	for k, v := range docs {
		values[domain.OptionKey(k)] = v.value
	}
	return domain.NewOptions(values)
}

func (cv *converter) sums(through time.Month, docs []monthDoc) domain.YearSums {
	entries := make([]domain.MonthEntry, 0, len(docs))
	for _, m := range docs {
		entries = append(entries, domain.MonthEntry{Month: time.Month(m.Month), Sums: domain.MonthSums{
			Income: domain.IncomeSums{
				Cash: m.Income.Cash.cents, Accrual: m.Income.Accrual.cents, IVACollected: m.Income.IVACollected.cents,
				IVAWithheld: withheld(m.Income.IVAWithheld), ISRWithheld: withheld(m.Income.ISRWithheld),
			},
			Expenses: domain.ExpenseSums{
				DeductiblePaid: m.Expenses.DeductiblePaid.cents, CreditableIVAPaid: m.Expenses.CreditableIVAPaid.cents,
				PredialPaid:        m.Expenses.PredialPaid.cents,
				ISRWithheldToRemit: m.Expenses.ISRWithheldToRemit.cents,
				IVAWithheldToRemit: m.Expenses.IVAWithheldToRemit.cents,
			},
		}})
	}
	sums, err := domain.NewYearSums(through, entries...)
	cv.check("months", err)
	return sums
}

func withheld(d withheldDoc) domain.Withheld {
	return domain.Withheld{ByMoral: d.ByMoral.cents, ByPlatform: d.ByPlatform.cents}
}

func (cv *converter) taxInputs(docs []taxInputDoc) domain.TaxInputs {
	values := make([]domain.TaxInputValue, 0, len(docs))
	for _, d := range docs {
		v := domain.TaxInputValue{Key: domain.InputKey(d.Key), Kind: domain.InputKind(d.Kind), Month: time.Month(d.Month)}
		given := 0
		if d.Money != nil {
			v.Money, given = d.Money.cents, given+1
		}
		if d.Rate != nil {
			v.Rate, given = d.Rate.rate, given+1
		}
		if d.Bool != nil {
			v.Bool, given = *d.Bool, given+1
		}
		if given != 1 || !kindMatches(v.Kind, d) {
			cv.fail("tax input %s: give exactly the one value its kind %q takes", d.Key, d.Kind)
			continue
		}
		values = append(values, v)
	}
	inputs, err := domain.NewTaxInputs(values...)
	cv.check("tax_inputs", err)
	return inputs
}

func kindMatches(k domain.InputKind, d taxInputDoc) bool {
	switch k {
	case domain.InputMoney, domain.InputMoneyMonthly:
		return d.Money != nil
	case domain.InputRate:
		return d.Rate != nil
	case domain.InputBoolean:
		return d.Bool != nil
	}
	return false
}

func (cv *converter) activeMonths(through time.Month, months []int) []time.Month {
	if months == nil {
		out := make([]time.Month, 0, int(through))
		for m := time.January; m <= through; m++ {
			out = append(out, m)
		}
		return out
	}
	out := make([]time.Month, 0, len(months))
	for _, m := range months {
		if m < int(time.January) || m > int(through) {
			cv.fail("active_months: month %d outside January..%s", m, through)
		}
		out = append(out, time.Month(m))
	}
	return out
}

func (cv *converter) expected(doc expectedDoc) Expected {
	e := Expected{Status: domain.Status(doc.Status)}
	if !e.Status.Valid() {
		cv.fail("expected.status %q", doc.Status)
	}
	e.Taxes = cv.taxes(doc.Taxes)
	e.ISR = cv.isr(e, doc.ISR)
	e.IVA = cv.iva(e, doc.IVA)
	if doc.WithheldToRemit != nil {
		if !e.asserts(domain.TaxWithholding) {
			cv.fail("expected.withheld_to_remit given but taxes has no withholding")
		}
		e.WithheldToRemit = doc.WithheldToRemit.cents
	}
	e.Lines = cv.lines(e, doc.Lines)
	e.MissingInputs = cv.missing(doc.MissingInputs)
	e.CPATopics = validList(cv, "cpa_topics", doc.CPATopics, func(s string) domain.CPATopic { return domain.CPATopic(s) })
	e.Disclaimers = validList(cv, "disclaimers", doc.Disclaimers, func(s string) domain.DisclaimerCode { return domain.DisclaimerCode(s) })
	return e
}

func (cv *converter) taxes(raw []string) []domain.TaxKind {
	if len(raw) == 0 {
		cv.fail("expected.taxes is empty: name the taxes the case asserts")
	}
	taxes := make([]domain.TaxKind, 0, len(raw))
	for _, s := range raw {
		k := domain.TaxKind(s)
		if !k.Valid() || slices.Contains(taxes, k) {
			cv.fail("expected.taxes: %q unknown or repeated", s)
			continue
		}
		taxes = append(taxes, k)
	}
	return taxes
}

func (cv *converter) isr(e Expected, doc *centsText) *money.Cents {
	if doc == nil {
		return nil
	}
	if !e.asserts(domain.TaxISR) {
		cv.fail("expected.isr given but taxes has no isr")
	}
	v := doc.cents
	return &v
}

func (cv *converter) iva(e Expected, doc *ivaDoc) *domain.IVAComponents {
	if doc == nil {
		return nil
	}
	if !e.asserts(domain.TaxIVA) {
		cv.fail("expected.iva given but taxes has no iva")
	}
	c := domain.IVAComponents{
		ChargedCollected: doc.ChargedCollected.cents, CreditablePaid: doc.CreditablePaid.cents,
		WithheldFromTaxpayer: doc.WithheldFromTaxpayer.cents, DefinitiveExcluded: doc.DefinitiveExcluded.cents,
	}
	if net := ivaNet(c); net != doc.Net.cents {
		cv.fail("expected.iva.net %s, but its components net to %s", doc.Net.cents, net)
	}
	return &c
}

// ivaNet is MODULE-tax-engine §2.6 for one regime: charged − excluded −
// creditable − withheld. Case amounts are far below int64 limits.
func ivaNet(c domain.IVAComponents) money.Cents {
	return c.ChargedCollected - c.DefinitiveExcluded - c.CreditablePaid - c.WithheldFromTaxpayer
}

func (cv *converter) lines(e Expected, docs []lineDoc) []LineWant {
	if docs == nil {
		cv.fail("expected.lines is missing: an authored case lists every line ([] when there is none)")
		return nil
	}
	lines := make([]LineWant, 0, len(docs))
	for i, d := range docs {
		l := LineWant{Tax: domain.TaxKind(d.Tax), Code: domain.LineCode(d.Code), Op: domain.Operation(d.Op), Amount: d.Amount.cents}
		if !e.asserts(l.Tax) || strings.TrimSpace(d.Code) == "" || !l.Op.Valid() {
			cv.fail("expected.lines[%d]: tax %q must be asserted, code set and op a contract operation (got %q)", i, d.Tax, d.Op)
		}
		lines = append(lines, l)
	}
	return lines
}

func (cv *converter) missing(docs []missingDoc) []domain.MissingInput {
	out := make([]domain.MissingInput, 0, len(docs))
	for _, d := range docs {
		mi := domain.MissingInput{Key: domain.InputKey(d.Key)}
		for _, m := range d.Months {
			if m < int(time.January) || m > int(time.December) {
				cv.fail("expected.missing_inputs %s: month %d", d.Key, m)
			}
			mi.Months = append(mi.Months, time.Month(m))
		}
		if strings.TrimSpace(d.Key) == "" {
			cv.fail("expected.missing_inputs: key is blank")
		}
		out = append(out, mi)
	}
	return out
}

// validList converts a list of contract enum values, absent meaning empty.
func validList[T interface{ Valid() bool }](cv *converter, field string, raw []string, conv func(string) T) []T {
	out := make([]T, 0, len(raw))
	for _, s := range raw {
		v := conv(s)
		if !v.Valid() {
			cv.fail("expected.%s: %q is not a contract value", field, s)
			continue
		}
		out = append(out, v)
	}
	return out
}
