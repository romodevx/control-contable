package pm

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// The legal basis of each RESICO PM step (research §B.6: art. 9 on cash
// flow, LISR 206 to 215).
var (
	resicoBasis = basis{
		income: []domain.Citation{
			{Law: "LISR", Article: "207", Note: "ingresos efectivamente percibidos"},
			{Law: "LISR", Article: "9", Note: "fracción I"},
		},
		deductions: []domain.Citation{
			{Law: "LISR", Article: "208", Note: "deducciones autorizadas efectivamente erogadas"},
			{Law: "LISR", Article: "9", Note: "fracción I"},
		},
		depreciation:    []domain.Citation{{Law: "LISR", Article: "209", Note: "deducción de inversiones"}},
		profitBeforePTU: []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción I, ingresos menos deducciones"}},
		ptuPaid:         []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción I, PTU pagada en el ejercicio"}},
		taxableProfit:   []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción I, utilidad fiscal"}},
		losses: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "fracción II, pérdidas fiscales pendientes"},
			{Law: "LISR", Article: "57", Note: "pérdidas de ejercicios anteriores"},
		},
		base: []domain.Citation{{Law: "LISR", Article: "9", Note: "fracción II, resultado fiscal"}},
		rate: []domain.Citation{
			{Law: "LISR", Article: "9", Note: "primer párrafo, tasa del 30 %"},
			{Law: "LISR", Article: "212", Note: "impuesto del ejercicio conforme al artículo 9"},
		},
		tax:      []domain.Citation{{Law: "LISR", Article: "212", Note: "impuesto del ejercicio"}},
		payments: []domain.Citation{{Law: "LISR", Article: "212", Note: "acreditamiento de los pagos provisionales"}},
		interest: []domain.Citation{{Law: "LISR", Article: "54", Note: "la retención sobre intereses se considera pago provisional"}},
		balance: []domain.Citation{
			{Law: "LISR", Article: "212", Note: "declaración del ejercicio"},
			{Law: "LISR", Article: "76", Note: "fracción V, declaración anual"},
		},
	}
	resicoBasisPTUBase = []domain.Citation{{
		Law: "LISR", Article: "212", Note: "segundo párrafo, utilidad fiscal del artículo 9 como base de la PTU",
	}}
)

// resico is régimen 626 for personas morales; crediting the art. 54
// interest withholding in RESICO PM is the monthly engine's open CPA
// question (topic resico_pm_interest_withholding), so its line is flagged.
var resico = regime{kind: annual.SectionPMRESICO, code: fiscal.RegimeRESICO, basis: resicoBasis, interestNeeds: true}

// RESICO is the pm_resico section: the art. 9 annual tax on the income
// collected and the deductions paid (LISR 206 to 215, research §B.6), with
// the PTU base as its only extra: the utilidad fiscal per 212 ¶2, PTU paid
// subtracted (⚠️ resico_pm_ptu_base, CPA item 8). There is no suggested
// coeficiente: RESICO PM provisional payments are on cash flow (art. 211).
type RESICO struct{}

// NewRESICO returns the pm_resico section.
func NewRESICO() RESICO { return RESICO{} }

// Kind is pm_resico.
func (RESICO) Kind() annual.SectionKind { return resico.kind }

// Compute traces the RESICO PM annual tax of in and its PTU base.
func (RESICO) Compute(in Input, p Params) (Section, error) {
	c, err := resico.compute(in, p)
	if err != nil {
		return Section{}, fmt.Errorf("%s %d: %w", resico.kind, in.FiscalYear, err)
	}
	ptuBase := max(c.taxableProfit, 0)
	c.l.Info(domain.Step{
		Tax: domain.TaxISR, Code: LinePTUBase, Label: domain.LabelKey(LinePTUBase),
		Source: domain.SourceComputed(), Basis: resicoBasisPTUBase, NeedsCPAReview: true,
	}, ptuBase)
	s, err := c.section(resico, []annual.CPATopic{annual.TopicRESICOPMPTUBase}, Extras{PTUBase: ptuBase.Round()})
	if err != nil {
		return Section{}, fmt.Errorf("%s %d: %w", resico.kind, in.FiscalYear, err)
	}
	return s, nil
}
