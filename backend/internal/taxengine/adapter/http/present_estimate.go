package taxenginehttp

import (
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// The response bodies of getTaxpayerEstimate and getEstimateBreakdown,
// field for field as the contract (TaxpayerEstimate, Breakdown): every
// required field is sent, null where the contract says so, and lists are
// never null. The generated types omit nil fields, which the contract does
// not allow here.

type parameterSetJSON struct {
	FiscalYear int `json:"fiscal_year"`
	Version    int `json:"version"`
}

type taxResultJSON struct {
	AmountCents         int64  `json:"amount_cents"`
	PayableRoundedCents int64  `json:"payable_rounded_cents"`
	Outcome             string `json:"outcome"`
}

type ivaComponentsJSON struct {
	ChargedCollectedCents     int64 `json:"charged_collected_cents"`
	CreditablePaidCents       int64 `json:"creditable_paid_cents"`
	WithheldFromTaxpayerCents int64 `json:"withheld_from_taxpayer_cents"`
	DefinitiveExcludedCents   int64 `json:"definitive_excluded_cents"`
}

type missingInputJSON struct {
	Key    string   `json:"key"`
	Label  string   `json:"label"`
	Months []string `json:"months,omitempty"` // only for money_monthly keys (minItems 1)
}

type regimeEstimateJSON struct {
	RegimeCode           string             `json:"regime_code"`
	RegimeName           string             `json:"regime_name"`
	Treatment            string             `json:"treatment"`
	Status               string             `json:"status"`
	MissingInputs        []missingInputJSON `json:"missing_inputs"`
	Explanation          *string            `json:"explanation"`
	ISR                  *taxResultJSON     `json:"isr"`
	IVAComponents        *ivaComponentsJSON `json:"iva_components"`
	WithheldToRemitCents int64              `json:"withheld_to_remit_cents"`
	CPAReviewPending     bool               `json:"cpa_review_pending"`
	CPATopics            []string           `json:"cpa_topics"`
	ParameterSet         parameterSetJSON   `json:"parameter_set"`
}

type disclaimerJSON struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Text     string `json:"text"`
}

type totalsJSON struct {
	IVA                      taxResultJSON `json:"iva"`
	ISR                      taxResultJSON `json:"isr"`
	WithholdingsToRemitCents int64         `json:"withholdings_to_remit_cents"`
	TotalPayableRoundedCents int64         `json:"total_payable_rounded_cents"`
}

// EstimateBody is contract TaxpayerEstimate as a response sends it: the
// body of getTaxpayerEstimate and the dashboard's estimate
// (PresentEstimate).
type EstimateBody struct {
	TaxpayerID         string               `json:"taxpayer_id"`
	Period             string               `json:"period"`
	DueDate            string               `json:"due_date"`
	Legend             string               `json:"legend"`
	Disclaimers        []disclaimerJSON     `json:"disclaimers"`
	Regimes            []regimeEstimateJSON `json:"regimes"`
	Totals             totalsJSON           `json:"totals"`
	ParameterSet       parameterSetJSON     `json:"parameter_set"`
	RecordsFingerprint string               `json:"records_fingerprint"`
	FiguresFingerprint string               `json:"figures_fingerprint"`
	EngineVersion      string               `json:"engine_version"`
}

type citationJSON struct {
	Law     string  `json:"law"`
	Article string  `json:"article"`
	Note    *string `json:"note"`
}

type breakdownLineJSON struct {
	Seq            int            `json:"seq"`
	Tax            string         `json:"tax"`
	Concept        string         `json:"concept"`
	Operation      string         `json:"operation"`
	AmountCents    int64          `json:"amount_cents"`
	Rate           *string        `json:"rate"`
	LegalBasis     []citationJSON `json:"legal_basis"`
	Source         string         `json:"source"`
	NeedsCPAReview bool           `json:"needs_cpa_review"`
}

type breakdownJSON struct {
	TaxpayerID    string              `json:"taxpayer_id"`
	Period        string              `json:"period"`
	Regime        regimeEstimateJSON  `json:"regime"`
	Lines         []breakdownLineJSON `json:"lines"`
	Legend        string              `json:"legend"`
	Disclaimers   []disclaimerJSON    `json:"disclaimers"`
	EngineVersion string              `json:"engine_version"`
}

// PresentEstimate renders e as the contract's TaxpayerEstimate: what
// getTaxpayerEstimate answers, and what the dashboard (M6) embeds as its
// estimate, so the two show the same totals to the centavo.
func PresentEstimate(e usecase.TaxpayerEstimate) EstimateBody { return toEstimate(e) }

func toEstimate(e usecase.TaxpayerEstimate) EstimateBody {
	set := toParameterSet(e.ParameterSet)
	regimes := make([]regimeEstimateJSON, 0, len(e.Regimes))
	for _, r := range e.Regimes {
		regimes = append(regimes, toRegime(r, e.Period, set))
	}
	return EstimateBody{
		TaxpayerID:  e.TaxpayerID,
		Period:      e.Period.String(),
		DueDate:     e.DueDate.String(),
		Legend:      e.Legend,
		Disclaimers: toDisclaimers(e.Disclaimers, disclaimerFacts{personType: e.PersonType, periodYear: e.Period.Year(), paramsYear: e.ParameterSet.FiscalYear}),
		Regimes:     regimes,
		Totals: totalsJSON{
			IVA: toTaxResult(e.Totals.IVA), ISR: toTaxResult(e.Totals.ISR),
			WithholdingsToRemitCents: int64(e.Totals.WithholdingsToRemit), TotalPayableRoundedCents: int64(e.Totals.TotalPayable),
		},
		ParameterSet:       set,
		RecordsFingerprint: e.RecordsFingerprint,
		FiguresFingerprint: e.FiguresFingerprint,
		EngineVersion:      e.EngineVersion,
	}
}

func toBreakdown(b usecase.RegimeBreakdown) breakdownJSON {
	lines := make([]breakdownLineJSON, 0, len(b.Lines))
	for _, l := range b.Lines {
		lines = append(lines, toLine(b.Regime.Key.Regime, b.Period, l))
	}
	return breakdownJSON{
		TaxpayerID:    b.TaxpayerID,
		Period:        b.Period.String(),
		Regime:        toRegime(b.Regime, b.Period, toParameterSet(b.ParameterSet)),
		Lines:         lines,
		Legend:        b.Legend,
		Disclaimers:   toDisclaimers(b.Disclaimers, disclaimerFacts{personType: b.PersonType, periodYear: b.Period.Year(), paramsYear: b.ParameterSet.FiscalYear}),
		EngineVersion: b.EngineVersion,
	}
}

func toParameterSet(s taxparams.SetID) parameterSetJSON {
	return parameterSetJSON{FiscalYear: s.FiscalYear, Version: s.Version}
}

func toRegime(r usecase.RegimeResult, p domain.Period, set parameterSetJSON) regimeEstimateJSON {
	out := regimeEstimateJSON{
		RegimeCode:           string(r.Key.Regime),
		RegimeName:           regimeName(r),
		Treatment:            string(r.Treatment),
		Status:               string(r.Status),
		MissingInputs:        toMissing(r, p),
		Explanation:          explanation(r.Key, r.Explanation, r.Text.Explanation),
		WithheldToRemitCents: int64(r.WithheldToRemit),
		CPAReviewPending:     r.CPAReviewPending(),
		CPATopics:            make([]string, 0, len(r.CPATopics)),
		ParameterSet:         set,
	}
	for _, t := range r.CPATopics {
		out.CPATopics = append(out.CPATopics, string(t))
	}
	if r.ISR != nil {
		isr := toTaxResult(*r.ISR)
		out.ISR = &isr
	}
	if r.IVA != nil {
		out.IVAComponents = &ivaComponentsJSON{
			ChargedCollectedCents:     int64(r.IVA.ChargedCollected),
			CreditablePaidCents:       int64(r.IVA.CreditablePaid),
			WithheldFromTaxpayerCents: int64(r.IVA.WithheldFromTaxpayer),
			DefinitiveExcludedCents:   int64(r.IVA.DefinitiveExcluded),
		}
	}
	return out
}

// regimeName is the catalog's name; a regime the catalog read missed shows
// its code rather than an empty name.
func regimeName(r usecase.RegimeResult) string {
	if r.Text.Name != "" {
		return r.Text.Name
	}
	return "Régimen " + string(r.Key.Regime)
}

// toMissing labels each missing input from the catalog and names its
// months in the period's year.
func toMissing(r usecase.RegimeResult, p domain.Period) []missingInputJSON {
	out := make([]missingInputJSON, 0, len(r.MissingInputs))
	for _, m := range r.MissingInputs {
		label, ok := r.Text.InputLabels[m.Key]
		if !ok {
			label = string(m.Key)
		}
		item := missingInputJSON{Key: string(m.Key), Label: label}
		for _, month := range m.Months {
			item.Months = append(item.Months, periodOf(p.Year(), month))
		}
		out = append(out, item)
	}
	return out
}

func periodOf(year int, m time.Month) string { return fmt.Sprintf("%04d-%02d", year, int(m)) }

func toTaxResult(r domain.TaxResult) taxResultJSON {
	return taxResultJSON{AmountCents: int64(r.Amount), PayableRoundedCents: int64(r.Payable()), Outcome: string(r.Outcome())}
}

func toDisclaimers(ds []domain.Disclaimer, f disclaimerFacts) []disclaimerJSON {
	out := make([]disclaimerJSON, 0, len(ds))
	for _, d := range ds {
		out = append(out, disclaimerJSON{Code: string(d.Code), Severity: string(d.Severity), Text: disclaimerText(d, f)})
	}
	return out
}

// toLine renders one line; rate only on multiply lines (contract
// BreakdownLine.rate).
func toLine(regime fiscal.RegimeCode, p domain.Period, l domain.Line) breakdownLineJSON {
	out := breakdownLineJSON{
		Seq:            l.Seq,
		Tax:            string(l.Tax),
		Concept:        Concept(regime, p, l),
		Operation:      string(l.Operation),
		AmountCents:    int64(l.Amount),
		LegalBasis:     make([]citationJSON, 0, len(l.Basis)),
		Source:         l.Source.String(),
		NeedsCPAReview: l.NeedsCPAReview,
	}
	if l.Operation == domain.OpMultiply {
		rate := l.Rate.String()
		out.Rate = &rate
	}
	for _, c := range l.Basis {
		cite := citationJSON{Law: c.Law, Article: c.Article}
		if c.Note != "" {
			note := c.Note
			cite.Note = &note
		}
		out.LegalBasis = append(out.LegalBasis, cite)
	}
	return out
}
