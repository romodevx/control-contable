package pm

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// rateKey is the persona moral ISR rate (LISR 9): 30 %.
const rateKey taxparams.Key = "isr.rate.persona_moral"

// The line codes of the art. 9 computation, in order. Depreciation, PTU
// paid, losses and the interest withholding have a line only when not zero;
// isr.no_profit replaces the losses after a negative utilidad fiscal; one
// payments line per source that has payments.
const (
	LineIncome              domain.LineCode = "isr.income"
	LineDeductions          domain.LineCode = "isr.deductions"
	LineDepreciation        domain.LineCode = "isr.depreciation"
	LineProfitBeforePTU     domain.LineCode = "isr.profit_before_ptu"
	LinePTUPaid             domain.LineCode = "isr.ptu_paid"
	LineTaxableProfit       domain.LineCode = "isr.taxable_profit"
	LineNoProfit            domain.LineCode = "isr.no_profit"
	LinePriorYearLosses     domain.LineCode = "isr.prior_year_losses"
	LineBase                domain.LineCode = "isr.base"
	LineRate                domain.LineCode = "isr.rate"
	LineAnnualTax           domain.LineCode = "isr.annual_tax"
	LinePaymentsFiled       domain.LineCode = "isr.payments_filed"
	LinePaymentsTaxInput    domain.LineCode = "isr.payments_tax_input"
	LinePaymentsEstimated   domain.LineCode = "isr.payments_estimated"
	LineInterestWithholding domain.LineCode = "isr.interest_withholding"
	LineBalance             domain.LineCode = "isr.balance"
	LineRefundOnlyPaid      domain.LineCode = "isr.refund_only_paid"
)

// inputFiledPayments names the annual_filed_payments the user confirmed;
// a tax_input payment comes from the M5 prior_provisional_payments
// (MODULE-annual §2.5).
const inputFiledPayments domain.InputKey = "filed_payments"

// paymentLines are the code and source of each payment source's credit.
var paymentLines = map[annual.PaymentSource]struct {
	code   domain.LineCode
	source domain.Source
}{
	annual.PaymentFiled:          {LinePaymentsFiled, domain.SourceTaxInput(inputFiledPayments)},
	annual.PaymentTaxInput:       {LinePaymentsTaxInput, domain.SourceTaxInput(domain.InputPriorProvisionalPayments)},
	annual.PaymentEstimatedByApp: {LinePaymentsEstimated, domain.SourceComputed()},
}

// basis is the legal basis of each art. 9 step of one regime.
type basis struct {
	income, deductions, depreciation, profitBeforePTU, ptuPaid, taxableProfit,
	losses, base, rate, tax, payments, interest, balance []domain.Citation
}

// regime is what sets one PM section's art. 9 computation apart.
type regime struct {
	kind          annual.SectionKind
	code          fiscal.RegimeCode
	basis         basis
	interestNeeds bool // the interest withholding credit is a CPA question
}

// basisRefund: only what was effectively paid or withheld is refunded.
var basisRefund = []domain.Citation{{Law: "LISR", Article: "152", Note: "devolución de lo efectivamente pagado o retenido"}}

// computation is the art. 9 trace of one regime's year and the exact
// subtotals the extras read.
type computation struct {
	l               *domain.Ledger
	profitBeforePTU money.Exact // renta gravable before PTU paid and losses
	taxableProfit   money.Exact // utilidad fiscal
	tax             Tax
	payments        annual.ResolvedPayments
}

// compute traces the art. 9 annual tax of in, up to the balance.
func (r regime) compute(in Input, p Params) (*computation, error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	yearEnd, err := fiscal.ParseDate(fmt.Sprintf("%04d-12-31", in.FiscalYear))
	if err != nil {
		return nil, fmt.Errorf("fiscal year %d: %w", in.FiscalYear, err)
	}
	rate, err := p.Rate(rateKey, yearEnd)
	if err != nil {
		return nil, err
	}
	totals, err := in.Payments.Totals()
	if err != nil {
		return nil, err
	}
	c := &computation{l: domain.NewLedger(), payments: in.Payments}
	var x exact
	c.profit(r, in, &x)
	base := c.base(r, in, &x)
	c.l.Scale(isrStep(LineRate, domain.SourceParams(rateKey), r.basis.rate), rate)
	isr := c.l.Result(isrStep(LineAnnualTax, domain.SourceComputed(), r.basis.tax))
	paid := c.credit(r, in, totals, &x)
	balance := c.l.Result(isrStep(LineBalance, domain.SourceComputed(), r.basis.balance)).Round()
	if balance < 0 && in.Payments.EstimatedMonths() > 0 {
		c.l.Info(isrStep(LineRefundOnlyPaid, domain.SourceComputed(), basisRefund), 0)
	}
	if x.err != nil {
		return nil, x.err
	}
	c.tax = Tax{
		AccumulableIncome: in.Income, Base: base.Round(), ISR: isr.Round(),
		Payments: paid, Withholdings: in.InterestWithholding, Balance: balanceOf(balance),
	}
	return c, nil
}

// profit traces income − deductions − depreciation = renta gravable, − PTU
// paid = utilidad fiscal.
func (c *computation) profit(r regime, in Input, x *exact) {
	c.l.Input(isrStep(LineIncome, domain.SourceRecords(), r.basis.income), x.of(in.Income))
	c.l.Subtract(isrStep(LineDeductions, domain.SourceRecords(), r.basis.deductions), x.of(in.Deductions))
	if in.Depreciation != 0 {
		c.l.Subtract(isrStep(LineDepreciation, domain.SourceTaxInput(InputAnnualDepreciation), r.basis.depreciation), x.of(in.Depreciation))
	}
	c.profitBeforePTU = c.l.Result(isrStep(LineProfitBeforePTU, domain.SourceComputed(), r.basis.profitBeforePTU))
	if in.PTUPaid != 0 {
		c.l.Subtract(isrStep(LinePTUPaid, domain.SourceTaxInput(domain.InputPTUPaid), r.basis.ptuPaid), x.of(in.PTUPaid))
	}
	c.taxableProfit = c.l.Result(isrStep(LineTaxableProfit, domain.SourceComputed(), r.basis.taxableProfit))
}

// base traces the resultado fiscal: the utilidad fiscal less the pending
// losses applied up to it (the rest stays pending, LISR 57), or zero after
// a loss year, whose losses are not applied.
func (c *computation) base(r regime, in Input, x *exact) money.Exact {
	switch {
	case c.taxableProfit < 0:
		c.l.Input(isrStep(LineNoProfit, domain.SourceComputed(), r.basis.base), 0)
	case in.PriorYearLosses != 0:
		c.l.Subtract(isrStep(LinePriorYearLosses, domain.SourceTaxInput(domain.InputPriorYearLosses), r.basis.losses),
			min(x.of(in.PriorYearLosses), c.taxableProfit))
	}
	return c.l.Result(isrStep(LineBase, domain.SourceComputed(), r.basis.base))
}

// credit subtracts the provisional payments, one line per source with the
// number of months as its argument, and the interest withholding; it
// returns the payments credited.
func (c *computation) credit(r regime, in Input, totals []annual.PaymentTotal, x *exact) money.Cents {
	var paid money.Cents
	for _, t := range totals {
		line := paymentLines[t.Source]
		step := isrStep(line.code, line.source, r.basis.payments)
		step.Args = []string{strconv.Itoa(t.Months)}
		c.l.Subtract(step, x.of(t.Amount))
		paid += t.Amount // Totals summed every payment without overflow
	}
	if in.InterestWithholding != 0 {
		step := isrStep(LineInterestWithholding, domain.SourceTaxInput(domain.InputInterestWithholding), r.basis.interest)
		step.NeedsCPAReview = r.interestNeeds
		c.l.Subtract(step, x.of(in.InterestWithholding))
	}
	return paid
}

// section closes the computation as a section of r: status, topics,
// disclaimers and lines.
func (c *computation) section(r regime, topics []annual.CPATopic, extras Extras) (Section, error) {
	lines, err := c.l.Lines()
	if err != nil {
		return Section{}, err
	}
	s := Section{
		Kind: r.kind, Regime: r.code, Status: annual.StatusComplete, Obligation: annual.ObligationObligated,
		MissingInputs: slices.Clone(c.payments.Missing), CPATopics: slices.Clone(topics),
		Disclaimers: []annual.DisclaimerCode{annual.DisclaimerInflationAdjustmentExcluded},
		Tax:         c.tax, Extras: extras, Lines: lines,
	}
	if len(s.MissingInputs) > 0 {
		s.Status = annual.StatusIncomplete
	}
	if c.payments.EstimatedMonths() > 0 {
		s.Disclaimers = append(s.Disclaimers, annual.DisclaimerPaymentsEstimated)
	}
	return s, nil
}

func (in Input) validate() error {
	figures := []money.Cents{in.Income, in.Deductions, in.Depreciation, in.PTUPaid, in.PriorYearLosses, in.InterestWithholding}
	if slices.ContainsFunc(figures, func(v money.Cents) bool { return v < 0 }) {
		return fmt.Errorf("figures must not be negative: %w", ErrInvalidInput)
	}
	if slices.ContainsFunc(in.Payments.Payments, func(p annual.Payment) bool { return p.Amount < 0 || !p.Source.Valid() }) {
		return fmt.Errorf("payments must be non-negative with a known source: %w", ErrInvalidInput)
	}
	if in.FiscalYear <= 0 {
		return fmt.Errorf("fiscal year %d: %w", in.FiscalYear, ErrInvalidInput)
	}
	return nil
}

// exact converts centavos with a sticky error, so the steps read as one
// sequence; the ledger reports its own errors through Lines.
type exact struct{ err error }

func (x *exact) of(v money.Cents) money.Exact {
	if x.err != nil {
		return 0
	}
	e, err := v.Exact()
	x.err = err
	return e
}

func isrStep(code domain.LineCode, source domain.Source, basis []domain.Citation) domain.Step {
	return domain.Step{Tax: domain.TaxISR, Code: code, Label: domain.LabelKey(code), Source: source, Basis: basis}
}
