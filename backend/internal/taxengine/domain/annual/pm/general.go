package pm

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// generalRegime is régimen 601, General de Ley Personas Morales.
const generalRegime fiscal.RegimeCode = "601"

// ProvisionalPTURate is the workers' share of the renta gravable: 10 %.
// ⚠️ ptu_rate: read from a secondary source (research §B.5 step 8); the
// Comisión Nacional resolution in the DOF is not in the parameter set yet
// (ANN-05 finding 4, CPA-REVIEW). When a ptu.rate parameter read off the
// DOF lands, the strategy reads it instead and this constant goes.
const ProvisionalPTURate = money.Rate(money.RateScale / 10)

// cuScale is 10^4: the suggested coeficiente has four decimals (⚠️
// cu_rounding, research §B.5 step 9).
const cuScale = 10_000

// The line codes of the 601 extras, after the balance.
const (
	LinePTUBase         domain.LineCode = "ptu.base"
	LinePTURate         domain.LineCode = "ptu.rate"
	LinePTUToDistribute domain.LineCode = "ptu.to_distribute"
)

// The legal basis of each 601 step (research §B.5).
var (
	generalBasis = basis{
		income: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "fracción I, ingresos acumulables del ejercicio"},
			{Law: "LISR", Article: "17", Note: "cuándo se obtienen los ingresos acumulables"},
		},
		deductions: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "fracción I, deducciones autorizadas"},
			{Law: "LISR", Article: "25", Note: "deducciones autorizadas"},
		},
		depreciation: []domain.Citation{
			{Law: "LISR", Article: "31", Note: "deducción de inversiones por porcentajes anuales"},
			{Law: "LISR", Article: "34", Note: "porcentajes máximos de deducción"},
		},
		profitBeforePTU: []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción I, ingresos acumulables menos deducciones"}},
		ptuPaid:         []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción I, PTU pagada en el ejercicio"}},
		taxableProfit:   []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción I, utilidad fiscal"}},
		losses: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "fracción II, pérdidas fiscales pendientes"},
			{Law: "LISR", Article: "57", Note: "pérdidas de ejercicios anteriores"},
		},
		base: []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción II, resultado fiscal"}},
		rate: []domain.Citation{{Law: "LISR", Article: "9", Note: "primer párrafo, tasa del 30 %"}},
		tax:  []domain.Citation{{Law: "LISR", Article: "9", Note: "impuesto del ejercicio"}},
		payments: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "tercer párrafo, impuesto del ejercicio menos pagos provisionales"},
			{Law: "LISR", Article: "14", Note: "pagos provisionales a cuenta del impuesto del ejercicio"},
		},
		interest: []domain.Citation{{Law: "LISR", Article: "54", Note: "la retención sobre intereses se considera pago provisional"}},
		balance: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "tercer párrafo, declaración del ejercicio"},
			{Law: "LISR", Article: "76", Note: "fracción V, declaración anual"},
		},
	}
	generalBasisPTUBase = []domain.Citation{{
		Law: "LISR", Article: "9", Note: "renta gravable para la PTU, sin disminuir PTU pagada ni pérdidas",
	}}
	generalBasisPTURate = []domain.Citation{{
		Law: "LFT", Article: "117", Note: "porcentaje que determine la Comisión Nacional para la PTU",
	}}
)

var general = regime{kind: annual.SectionPM601, code: generalRegime, basis: generalBasis}

// General is the pm_601 section: régimen 601's annual return (LISR 9,
// research §B.5), with the PTU to distribute (⚠️ ptu_rate) and the
// suggested coeficiente de utilidad for the next year (⚠️ cu_rounding) as
// informative extras. The 28 fr. XXX adjustment of the PTU base and the
// annual inflation adjustment are not modelled (disclaimer
// inflation_adjustment_excluded).
type General struct{}

// NewGeneral returns the pm_601 section.
func NewGeneral() General { return General{} }

// Kind is pm_601.
func (General) Kind() annual.SectionKind { return general.kind }

// Compute traces the 601 annual tax of in and its extras.
func (General) Compute(in Input, p Params) (Section, error) {
	c, err := general.compute(in, p)
	if err != nil {
		return Section{}, fmt.Errorf("%s %d: %w", general.kind, in.FiscalYear, err)
	}
	extras, err := c.generalExtras(in)
	if err != nil {
		return Section{}, fmt.Errorf("%s %d: %w", general.kind, in.FiscalYear, err)
	}
	s, err := c.section(general, []annual.CPATopic{annual.TopicCURounding, annual.TopicPTURate}, extras)
	if err != nil {
		return Section{}, fmt.Errorf("%s %d: %w", general.kind, in.FiscalYear, err)
	}
	return s, nil
}

// generalExtras traces the PTU to distribute, ProvisionalPTURate × the
// renta gravable before PTU paid and losses (zero after a loss), and
// computes the suggested coeficiente: utilidad fiscal ÷ ingresos nominales.
func (c *computation) generalExtras(in Input) (Extras, error) {
	ptuBase := max(c.profitBeforePTU, 0)
	c.l.Input(domain.Step{
		Tax: domain.TaxISR, Code: LinePTUBase, Label: domain.LabelKey(LinePTUBase),
		Source: domain.SourceComputed(), Basis: generalBasisPTUBase,
	}, ptuBase)
	c.l.Scale(domain.Step{
		Tax: domain.TaxISR, Code: LinePTURate, Label: domain.LabelKey(LinePTURate),
		Source: domain.SourceComputed(), Basis: generalBasisPTURate, NeedsCPAReview: true,
	}, ProvisionalPTURate)
	toDistribute := c.l.Result(domain.Step{
		Tax: domain.TaxISR, Code: LinePTUToDistribute, Label: domain.LabelKey(LinePTUToDistribute),
		Source: domain.SourceComputed(), Basis: generalBasisPTURate, NeedsCPAReview: true,
	})
	cu, err := suggestedCU(c.taxableProfit.Round(), in.Income)
	if err != nil {
		return Extras{}, err
	}
	return Extras{PTUBase: ptuBase.Round(), PTUToDistribute: toDistribute.Round(), SuggestedCU: cu}, nil
}

// suggestedCU is profit ÷ income rounded half up to four decimals (LISR 14
// fr. I; ingresos nominales are the acumulables without the inflation
// adjustment, which is not modelled), or 0 when there is no profit or no
// income.
func suggestedCU(profit, income money.Cents) (money.Rate, error) {
	if profit <= 0 || income <= 0 {
		return 0, nil
	}
	// MulRatio is used as an exact 128-bit profit × 10^4 ÷ income with half
	// away from zero rounding; the Exact type only carries the integer.
	q, err := money.Exact(profit).MulRatio(cuScale, int64(income))
	if err != nil {
		return 0, fmt.Errorf("suggested coeficiente: %w", err)
	}
	return money.Rate(int64(q) * (money.RateScale / cuScale)), nil
}
