package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Names in this file carry a platform prefix: every regime strategy lives
// in package domain (PLAN PG-M5-S), so a generic name would collide.

// The 625 regime options (MODULE-taxpayers §2.3; catalog regimes.option_defs).
const (
	platformOptionActivities OptionKey = "activities"            // multi_enum, ≥ 1
	platformOptionDefinitive OptionKey = "definitive_option"     // default false
	platformOptionRFCGiven   OptionKey = "rfc_given_to_platform" // default true
)

// The parameter keys of the platform ISR rates and the definitive cap.
const (
	platformRateNoRFC     taxparams.Key = "isr.platform.no_rfc"
	platformDefinitiveCap taxparams.Key = "regime.platform.definitive_option_cap"
)

// The 625 line codes. The ISR ones are appended in this order; the IVA
// lines are ComputeIVA's, followed by the creditable IVA not applied.
const (
	platformLineIncome          LineCode = "isr.platform_income"
	platformLineRate            LineCode = "isr.platform_rate"
	platformLineWithheld        LineCode = "isr.withheld_by_platform"
	platformLineNoFavor         LineCode = "isr.definitive_no_favor"
	platformLinePayable         LineCode = "isr.payable"
	platformLineProvisional     LineCode = "isr.withheld_by_platform_provisional"
	platformLineDefinitiveInfo  LineCode = "isr.withheld_by_platform_definitive"
	platformLineCreditableNotIn LineCode = "iva.creditable_not_applied"
)

// The explanations of a partial 625 estimate (§2.2): what is not estimated.
const (
	// platformExplainProvisional: no definitive option, or prior-year
	// income over the cap. "Calcula tu pago en 612 o RESICO y acredita
	// estas retenciones" (MODULE-tax-engine §2.4).
	platformExplainProvisional LabelKey = "platform_provisional_withholding"
	// platformExplainMixed: several activities with the RFC given; the
	// month's sums do not split the income by activity, so no single rate
	// applies to it.
	platformExplainMixed LabelKey = "platform_mixed_activities"
)

// platformRate is one ISR withholding rate and its legal basis.
type platformRate struct {
	key   taxparams.Key
	basis []Citation
}

// platformActivityRates are the 2026 rates by activity (research A.2.6),
// keyed by the activities option value (= incomes platform_activity).
var platformActivityRates = map[string]platformRate{
	"transport_delivery": {
		key:   "isr.platform.transport_delivery",
		basis: []Citation{{Law: "LISR", Article: "113-A", Note: "fracción I, tasa del 2.1 %"}},
	},
	"lodging": {
		key:   "isr.platform.lodging",
		basis: []Citation{{Law: "LISR", Article: "113-A", Note: "fracción II, tasa del 4 %"}},
	},
	"goods_services": {
		key:   "isr.platform.goods_services",
		basis: []Citation{{Law: "LIF 2026", Article: "25", Note: "fracción VI, tasa del 2.5 %"}},
	},
}

// platformNoRFCRate applies to every activity when the platform does not
// have the taxpayer's RFC.
var platformNoRFCRate = platformRate{
	key:   platformRateNoRFC,
	basis: []Citation{{Law: "LISR", Article: "113-C", Note: "fracción IV, tasa del 20 % sin RFC"}},
}

// The legal basis of the other 625 steps (MODULE-tax-engine §2.5).
var (
	platformBasisIncome = []Citation{
		{Law: "LISR", Article: "113-A", Note: "ingresos cobrados sin IVA"},
		{Law: "LISR", Article: "113-A", Note: "último párrafo, ingresos cobrados directamente"},
	}
	platformBasisWithheld    = []Citation{{Law: "LISR", Article: "113-A", Note: "retención por la plataforma"}}
	platformBasisDefinitive  = []Citation{{Law: "LISR", Article: "113-B", Note: "la retención se considera definitiva"}}
	platformBasisProvisional = []Citation{
		{Law: "LISR", Article: "113-A", Note: "retención como pago provisional"},
	}
	platformBasisPayable = []Citation{
		{Law: "LISR", Article: "113-A", Note: "último párrafo, pago de los ingresos cobrados directamente"},
		{Law: "LISR", Article: "113-B", Note: "opción de retención definitiva"},
	}
	platformBasisNotCredited = []Citation{
		{Law: "LIVA", Article: "18-L"}, {Law: "LIVA", Article: "18-M", Note: "opción de retención definitiva"},
	}
)

// Strategy625PF computes regime 625 for a persona física, Plataformas
// tecnológicas (LISR 113-A to 113-D; LIF 2026 art. 25 fr. VI; LIVA 18-J,
// 18-L, 18-M; research A.2.6, MODULE-tax-engine §2.4).
//
// With the definitive option and prior-year income ≤ the cap (LISR 113-B):
//
//	ISR = month's cash income × the activity rate (20 % without RFC)
//	      − ISR withheld by the platform
//
// The withholding is definitive: income collected directly through no
// platform is self-paid at the same rate (113-A last ¶), and a withholding
// above the rate leaves no balance in favor. The IVA left after the
// platform's 50 % withholding is settled by it (18-M ⚠️, the M5-09 hook),
// and no creditable IVA reduces it (PR #79 follow-up); the IVA charged on
// direct collections stays under the general rules (conservative: the
// 18-M 8 % on direct collections needs a platform/direct split the sums
// do not carry).
//
// Without the option or over the cap, the estimate is partial: the
// withholdings are provisional credits of a 612 or RESICO computation, so
// no ISR is estimated, and the IVA follows the general rules. Several
// activities with the RFC given are partial too: no single rate applies.
type Strategy625PF struct{}

// NewStrategy625PF returns the 625 PF strategy.
func NewStrategy625PF() Strategy625PF { return Strategy625PF{} }

// Key is 625 for a persona física.
func (Strategy625PF) Key() StrategyKey {
	return StrategyKey{Regime: fiscal.RegimePlatforms, PersonType: fiscal.PersonFisica}
}

// Treatment is full: the catalog's level. One estimate may be partial
// (RegimeEstimate.Coverage).
func (Strategy625PF) Treatment() Treatment { return TreatmentFull }

// RequiredInputs is the prior-year income, only with the definitive option
// (MODULE-tax-engine §2.8): it decides the eligibility.
func (Strategy625PF) RequiredInputs(o Options, _ Period, _ Facts) []RequiredInput {
	if opted, _ := o.Bool(platformOptionDefinitive); !opted {
		return nil
	}
	return []RequiredInput{{Key: InputPriorYearIncome, Kind: InputMoney}}
}

// Estimate computes the month's ISR (when it can) and IVA components
// through one Ledger.
func (s Strategy625PF) Estimate(in Input, p *taxparams.Params) (RegimeEstimate, error) {
	definitive, err := platformDefinitive(in, p)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("625: %w", err)
	}
	month := in.Sums.Month(in.Period.Month())
	l := NewLedger()
	est := RegimeEstimate{Key: s.Key(), Coverage: TreatmentFull}
	if err := platformISR(l, &est, month, in, p, definitive); err != nil {
		return RegimeEstimate{}, fmt.Errorf("625 ISR: %w", err)
	}
	iva, err := platformIVA(l, month, in.Options, definitive)
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("625: %w", err)
	}
	lines, err := l.Lines()
	if err != nil {
		return RegimeEstimate{}, fmt.Errorf("625: %w", err)
	}
	est.IVA = &iva.Components
	est.WithheldToRemit = iva.WithheldToRemit
	est.Lines = lines
	est.CPATopics = iva.CPATopics
	est.Disclaimers = iva.Disclaimers
	return est, nil
}

// platformDefinitive reports the definitive option in effect: opted for,
// and prior-year income within the cap (LISR 113-B fr. I). A missing
// prior-year income counts as zero, eligible (§2.8).
func platformDefinitive(in Input, p *taxparams.Params) (bool, error) {
	if opted, _ := in.Options.Bool(platformOptionDefinitive); !opted {
		return false, nil
	}
	limit, err := p.Money(platformDefinitiveCap, in.Period.End())
	if err != nil {
		return false, err
	}
	prior, _ := in.Inputs.Money(InputPriorYearIncome)
	return prior <= limit, nil
}

// platformRFCGiven reads rfc_given_to_platform, true when unset.
func platformRFCGiven(o Options) bool {
	given, set := o.Bool(platformOptionRFCGiven)
	return !set || given
}

// platformRateFor resolves the one ISR rate of the month's income: 20 %
// without RFC, else the rate of the single activity; false when there are
// several, none or an unknown one.
func platformRateFor(o Options) (platformRate, bool) {
	if !platformRFCGiven(o) {
		return platformNoRFCRate, true
	}
	activities, _ := o.List(platformOptionActivities)
	if len(activities) != 1 {
		return platformRate{}, false
	}
	r, ok := platformActivityRates[activities[0]]
	return r, ok
}

func platformStep(tax TaxKind, code LineCode, source Source, basis []Citation) Step {
	return Step{Tax: tax, Code: code, Label: LabelKey(code), Source: source, Basis: basis}
}

// platformISR traces the ISR into l and fills est's ISR, or makes est
// partial with the withholding shown as an info line.
func platformISR(l *Ledger, est *RegimeEstimate, month MonthSums, in Input, p *taxparams.Params, definitive bool) error {
	withheld, err := month.Income.ISRWithheld.ByPlatform.Exact()
	if err != nil {
		return err
	}
	if !definitive {
		l.Info(platformStep(TaxISR, platformLineProvisional, SourceRecords(), platformBasisProvisional), withheld)
		est.Coverage, est.Explanation = TreatmentPartial, platformExplainProvisional
		return nil
	}
	r, ok := platformRateFor(in.Options)
	if !ok {
		l.Info(platformStep(TaxISR, platformLineDefinitiveInfo, SourceRecords(), platformBasisDefinitive), withheld)
		est.Coverage, est.Explanation = TreatmentPartial, platformExplainMixed
		return nil
	}
	rate, err := p.Rate(r.key, in.Period.End())
	if err != nil {
		return err
	}
	income, err := month.Income.Cash.Exact()
	if err != nil {
		return err
	}
	l.Input(platformStep(TaxISR, platformLineIncome, SourceRecords(), platformBasisIncome), income)
	l.Scale(platformStep(TaxISR, platformLineRate, SourceParams(r.key), r.basis), rate)
	l.Subtract(platformStep(TaxISR, platformLineWithheld, SourceRecords(), platformBasisWithheld), withheld)
	if over := l.Running(); over < 0 {
		l.Add(platformStep(TaxISR, platformLineNoFavor, SourceComputed(), platformBasisDefinitive), -over)
	}
	payable := l.Result(platformStep(TaxISR, platformLinePayable, SourceComputed(), platformBasisPayable))
	isr := NewTaxResult(payable)
	est.ISR = &isr
	// A ledger error is sticky: ComputeIVA and Lines report it.
	return nil
}

// platformIVA computes the month's IVA. Under the definitive option the
// platform's withholding settles the IVA left on its sales (18-M ⚠️), and
// the creditable IVA is not taken: it is shown as an info line instead.
func platformIVA(l *Ledger, month MonthSums, o Options, definitive bool) (IVAResult, error) {
	in := IVAInput{Sums: month, Options: o}
	if !definitive {
		return ComputeIVA(l, in)
	}
	notCredited := month.Expenses.CreditableIVAPaid
	in.Sums.Expenses.CreditableIVAPaid = 0
	in.Definitive = platformDefinitiveIVA(platformRFCGiven(o))
	result, err := ComputeIVA(l, in)
	if err != nil || notCredited == 0 {
		return result, err
	}
	amount, err := notCredited.Exact()
	if err != nil {
		return IVAResult{}, err
	}
	s := platformStep(TaxIVA, platformLineCreditableNotIn, SourceRecords(), platformBasisNotCredited)
	s.NeedsCPAReview = true
	l.Info(s, amount) // a ledger error is sticky: Lines reports it
	return result, nil
}

// platformDefinitiveIVA is the 625 DefinitiveIVARule (LIVA 18-M ⚠️): with
// the RFC the platform withholds half the IVA of its sales (LIVA 18-J fr.
// II a)), so the other half, equal to the withholding, is settled by it.
// It never exceeds what is left of the IVA charged after the withholding,
// so IVA on direct collections stays payable. Without the RFC the
// platform withholds all of it and nothing is left to exclude.
func platformDefinitiveIVA(rfcGiven bool) DefinitiveIVARule {
	return func(income IncomeSums) (money.Cents, error) {
		if !rfcGiven {
			return 0, nil
		}
		charged, err := income.ByRate.Rate16.IVA.Add(income.ByRate.Rate8.IVA)
		if err != nil {
			return 0, err
		}
		withheld := income.IVAWithheld.ByPlatform
		rest, err := charged.Sub(withheld)
		if err != nil {
			return 0, err
		}
		return max(min(withheld, rest), 0), nil
	}
}
