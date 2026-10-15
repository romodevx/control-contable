package annual

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/explain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// YearStatus is closed or in_progress (contract AnnualYearStatus).
type YearStatus string

// The year statuses.
const (
	YearClosed     YearStatus = "closed"
	YearInProgress YearStatus = "in_progress"
)

// InputTarget is the screen that completes a missing input (contract
// AnnualInputTarget).
type InputTarget string

// The input targets.
const (
	TargetSalaryStatements   InputTarget = "salary_statements"
	TargetPersonalDeductions InputTarget = "personal_deductions"
	TargetAnnualInputs       InputTarget = "annual_inputs"
	TargetFiledPayments      InputTarget = "filed_payments"
	TargetTaxInputs          InputTarget = "tax_inputs"
)

// MissingInput is a required input a section does not have; Regime and
// Months are set for filed payments only.
type MissingInput struct {
	Key    string
	Target InputTarget
	Regime fiscal.RegimeCode
	Months []domain.Period
}

// missingFiledPayment is the key of a month to confirm.
const missingFiledPayment = "filed_payment"

// NoteCode is an informative note under a section's figure (contract
// AnnualNoteCode).
type NoteCode string

// The note codes, in the order a section lists them.
const (
	NoteRefundOnlyPaid    NoteCode = "refund_only_paid"
	NoteRefundNeedsEFirma NoteCode = "refund_needs_efirma"
	NoteLossCarryforward  NoteCode = "loss_carryforward"
	NoteCUSuggestion      NoteCode = "cu_suggestion"
)

// efirmaRefund is the saldo a favor from which a refund request needs the
// e.firma (research §A.2): 10,000.00.
const efirmaRefund money.Cents = 1_000_000

// ReasonCode is one plain-words reason behind a section's obligation; the
// adapter renders it (contract obligation_reasons).
type ReasonCode string

// The obligation reasons (MODULE-annual §2.4).
const (
	ReasonBusinessIncome        ReasonCode = "business_income"
	ReasonPlatformIncome        ReasonCode = "platform_income"
	ReasonRentIncome            ReasonCode = "rent_income"
	ReasonSimultaneousEmployers ReasonCode = "simultaneous_employers"
	ReasonLeftBeforeYearEnd     ReasonCode = "left_before_year_end"
	ReasonPayerNotWithholding   ReasonCode = "payer_not_withholding"
	ReasonToldEmployerWillFile  ReasonCode = "told_employer_will_file"
	ReasonSalaryOverThreshold   ReasonCode = "salary_over_threshold"
	ReasonSalaryOnly            ReasonCode = "salary_only"
	ReasonNoIncome              ReasonCode = "no_income"
	ReasonRESICORelief          ReasonCode = "resico_relief"
	ReasonPersonaMoral          ReasonCode = "persona_moral"
)

// Disclaimer is an annual disclaimer with its es-MX text.
type Disclaimer struct {
	Code     adomain.DisclaimerCode
	Severity adomain.Severity
	Text     string
}

// Line is one breakdown line and the regime it belongs to ("" for a
// section-wide line).
type Line struct {
	domain.Line
	Regime fiscal.RegimeCode
}

// Section is one section of the estimate (contract AnnualSection).
type Section struct {
	Kind              adomain.SectionKind
	Regimes           []fiscal.RegimeCode
	Status            adomain.Status
	MissingInputs     []MissingInput
	Obligation        adomain.Obligation // "" is null
	ObligationReasons []ReasonCode
	Deadline          fiscal.Date // zero is null
	Tax               *Tax
	// Credits and Sources go with Tax: their total and how many
	// regime-months came from each payment source.
	Credits     money.Cents
	Sources     SourceCounts
	PMExtras    *PMExtras
	RESICOCheck *adomain.RESICOCheck
	Explanation string // "" is null
	Notes       []NoteCode
	CPATopics   []adomain.CPATopic
	Lines       []Line
	// raised are the disclaimer codes the section raised.
	raised []adomain.DisclaimerCode
}

// Estimate is the annual estimate (contract AnnualEstimate).
type Estimate struct {
	TaxpayerID                string
	FiscalYear                int
	PersonType                fiscal.PersonType
	Legend                    string
	Disclaimers               []Disclaimer
	YearStatus                YearStatus
	DataThrough               domain.Period // zero is null
	Deadline                  fiscal.Date   // zero is null
	HeadlineSections          []adomain.SectionKind
	Sections                  []Section
	InformativeIncomeOver500k bool
	ParameterSet              taxparams.SetID
	EngineVersion             string
	// Fingerprint hashes everything above: equal inputs, parameters and
	// engine version give an equal fingerprint (MODULE-annual §2.11).
	Fingerprint string
	// RangeUsed reports the Estimator range ran (one more estimate read).
	RangeUsed bool
}

// Breakdown is one computed section's trace (contract AnnualBreakdown).
type Breakdown struct {
	TaxpayerID    string
	FiscalYear    int
	Section       Section
	Legend        string
	Disclaimers   []Disclaimer
	ParameterSet  taxparams.SetID
	EngineVersion string
	RangeUsed     bool
}

// ErrSectionNotPresent reports that the breakdown asked for a section the estimate
// does not have (422 section_not_present).
var ErrSectionNotPresent = errors.New("annual: section not present")

const (
	fieldSection         = "section"
	detailSectionMissing = "section_not_present"
	msgSectionMissing    = "Tu estimación anual no tiene esta sección."
)

func sectionNotPresentError() error {
	return businessRule(ErrSectionNotPresent, fieldSection, detailSectionMissing, msgSectionMissing)
}

// rmfPendingYear is the last fiscal year whose annual facilities wait for
// the next year's RMF (rmf_2027_pending): the RMF 2027 is not published.
const rmfPendingYear = 2026

// build derives the estimate from what was read: no I/O.
func (uc *Estimator) build(id string, w yearWindow, sel taxparams.Selection, r readings) (Estimate, error) {
	year, err := yearOf(w, r)
	if err != nil {
		return Estimate{}, err
	}
	computed, err := uc.deps.Sections.Compute(year, sel.Params)
	if err != nil {
		return Estimate{}, err
	}
	est := Estimate{
		TaxpayerID: id, FiscalYear: w.fiscalYear, PersonType: r.profile.PersonType, Legend: explain.Legend,
		YearStatus: YearClosed, ParameterSet: taxparams.SetID{FiscalYear: sel.Params.FiscalYear(), Version: sel.Params.Version()},
		EngineVersion: domain.EngineVersion, RangeUsed: r.rangeUsed,
	}
	if w.inProgress {
		est.YearStatus = YearInProgress
	}
	if w.through > 0 {
		if est.DataThrough, err = domain.NewPeriod(w.fiscalYear, w.through); err != nil {
			return Estimate{}, err
		}
	}
	for _, c := range computed {
		s, err := sectionOf(c, year, r)
		if err != nil {
			return Estimate{}, err
		}
		est.Sections = append(est.Sections, s)
		est.InformativeIncomeOver500k = est.InformativeIncomeOver500k || c.InformativeIncomeOver500k
	}
	est.Sections = append(est.Sections, explanationSections(r)...)
	if err := est.estimateFacts(year, sel.Params, computed); err != nil {
		return Estimate{}, err
	}
	if est.Disclaimers, err = estimateDisclaimers(est, w); err != nil {
		return Estimate{}, err
	}
	est.Fingerprint, err = fingerprint(est)
	return est, err
}

// estimateFacts sets the headline, the earliest deadline and, for a
// persona física without pf_art152, the informative-income fact.
func (est *Estimate) estimateFacts(y Year, p *taxparams.Params, computed []ComputedSection) error {
	for _, s := range est.Sections {
		if s.Tax != nil && s.Kind != adomain.SectionPFRESICOCheck {
			est.HeadlineSections = append(est.HeadlineSections, s.Kind)
		}
		if !s.Deadline.IsZero() && (est.Deadline.IsZero() || s.Deadline.String() < est.Deadline.String()) {
			est.Deadline = s.Deadline
		}
	}
	est.HeadlineSections = slices.Compact(est.HeadlineSections)
	hasArt152 := slices.ContainsFunc(computed, func(c ComputedSection) bool { return c.Kind == adomain.SectionPFArt152 })
	if est.PersonType != fiscal.PersonFisica || hasArt152 {
		return nil
	}
	total := y.Inputs.OtherExemptIncome
	for _, r := range y.Regimes {
		var err error
		if total, err = total.Add(r.Income); err != nil {
			return err
		}
	}
	over, err := explain.InformativeIncome(p, y.FiscalYear, total)
	est.InformativeIncomeOver500k = over
	return err
}

// yearOf assembles the strategies' year: each held regime's months in its
// window through data_through.
func yearOf(w yearWindow, r readings) (Year, error) {
	y := Year{
		FiscalYear: w.fiscalYear, PersonType: r.profile.PersonType,
		InterestWithholding: map[fiscal.RegimeCode]money.Cents{},
		SalaryStatements:    r.year.SalaryStatements, Deductions: r.year.Deductions, Inputs: r.year.Facts,
		Payments: r.book,
	}
	for _, h := range r.held {
		ry, err := regimeYearOf(h, r.profile.PersonType, w, r)
		if err != nil {
			return Year{}, fmt.Errorf("regime %s: %w", h.Code, err)
		}
		y.Regimes = append(y.Regimes, ry)
		if v, ok := r.inputs[h.Code].Money(domain.InputInterestWithholding); ok {
			y.InterestWithholding[h.Code] = v
		}
	}
	return y, nil
}

// regimeYearOf sums one regime's active months: income on its basis
// (accrual for 601, collected otherwise) plus the income before the app,
// deductions paid plus those before the app, predial, the ISR withheld
// from the taxpayer, and its annual tax inputs.
func regimeYearOf(h usecase.RegimeHold, pt fiscal.PersonType, w yearWindow, r readings) (adomain.RegimeYear, error) {
	active := activeMonths(h, w)
	in := r.inputs[h.Code]
	before := func(key domain.InputKey) money.Cents { v, _ := in.Money(key); return v }
	accrual := pt == fiscal.PersonMoral && h.Code == "601"
	y := adomain.RegimeYear{
		Regime: h.Code, Options: h.Options, Income: before(domain.InputYTDBeforeAppIncome),
		Deductions: before(domain.InputYTDBeforeAppDeductions), Depreciation: before(adomain.InputAnnualDepreciation),
		PTUPaid: before(domain.InputPTUPaid), PriorYearLosses: before(domain.InputPriorYearLosses),
	}
	var err error
	for _, row := range r.incomes {
		if row.Regime != h.Code || !slices.Contains(active, row.Month) {
			continue
		}
		income := row.Sums.Cash
		if accrual {
			income = row.Sums.Accrual
		}
		if y.Income, err = sumMoney(y.Income, income); err != nil {
			return adomain.RegimeYear{}, err
		}
		if y.ISRWithheld, err = sumMoney(y.ISRWithheld, row.Sums.ISRWithheld.ByMoral, row.Sums.ISRWithheld.ByPlatform); err != nil {
			return adomain.RegimeYear{}, err
		}
	}
	for _, row := range r.expenses {
		if row.Regime != h.Code || !slices.Contains(active, row.Month) {
			continue
		}
		if y.Deductions, err = sumMoney(y.Deductions, row.Sums.DeductiblePaid); err != nil {
			return adomain.RegimeYear{}, err
		}
		if y.Predial, err = sumMoney(y.Predial, row.Sums.PredialPaid); err != nil {
			return adomain.RegimeYear{}, err
		}
	}
	return y, nil
}

// sectionOf completes a computed section: deadline, reasons, notes,
// structured missing inputs, credits, text and line regimes.
func sectionOf(c ComputedSection, y Year, r readings) (Section, error) {
	s := Section{
		Kind: c.Kind, Regimes: c.Regimes, Status: c.Status, MissingInputs: missingInputs(c.MissingInputs),
		Obligation: c.Obligation, ObligationReasons: obligationReasons(c, y), Tax: c.Tax, PMExtras: c.PMExtras,
		RESICOCheck: c.RESICOCheck, CPATopics: c.CPATopics, raised: c.Disclaimers,
		Deadline: r.deadlines[explain.SectionDeadline(c.Kind)], Sources: y.Payments.Counts(c.Regimes...),
	}
	if c.Tax != nil {
		var err error
		if s.Credits, err = sumMoney(c.Tax.Payments, c.Tax.Withholdings, c.Tax.SalaryISR); err != nil {
			return Section{}, err
		}
	}
	if c.Kind == adomain.SectionPFRESICOCheck {
		s.Explanation = explain.DisclaimerText(explain.Disclaimer{Code: adomain.DisclaimerRESICOPFDefinitive}, explain.TextFacts{FiscalYear: y.FiscalYear})
	}
	s.Notes = notesOf(c, s.Sources)
	regimes := LineRegimes(c.Kind, c.Regimes, c.Lines)
	s.Lines = make([]Line, len(c.Lines))
	for i, l := range c.Lines {
		s.Lines[i] = Line{Line: l, Regime: regimes[i]}
	}
	return s, nil
}

// explanationSections are one section per explanation-only regime held.
func explanationSections(r readings) []Section {
	held := make([]explain.Held, 0, len(r.held))
	for _, h := range r.held {
		held = append(held, explain.Held{Regime: h.Code, DefinitiveOption: definitiveOption(h)})
	}
	var out []Section
	for _, e := range explain.Sections(r.profile.PersonType, held) {
		deadline := r.deadlines[e.Deadline]
		out = append(out, Section{
			Kind: e.Kind, Regimes: []fiscal.RegimeCode{e.Regime}, Status: e.Status, Obligation: e.Obligation,
			Deadline: deadline, Explanation: explain.ExplanationText(e.Code, deadline),
		})
	}
	return out
}

// missingInputs structures the domain's names (contract
// AnnualMissingInput): the filed-payment months of one regime become one
// entry with its months.
func missingInputs(names []string) []MissingInput {
	var out []MissingInput
	for _, name := range names {
		parts := strings.Split(name, ":")
		if len(parts) != 3 || parts[0] != missingFiledPayment {
			out = append(out, MissingInput{Key: name, Target: targetOf(name)})
			continue
		}
		period, err := domain.ParsePeriod(parts[2])
		if err != nil {
			continue // the domain names valid months only
		}
		regime := fiscal.RegimeCode(parts[1])
		i := slices.IndexFunc(out, func(m MissingInput) bool { return m.Key == missingFiledPayment && m.Regime == regime })
		if i < 0 {
			out = append(out, MissingInput{Key: missingFiledPayment, Target: TargetFiledPayments, Regime: regime})
			i = len(out) - 1
		}
		out[i].Months = append(out[i].Months, period)
	}
	return out
}

func targetOf(name string) InputTarget {
	switch name {
	case adomain.MissingSalaryStatement:
		return TargetSalaryStatements
	case adomain.MissingPriorYearAccumulableIncome:
		return TargetAnnualInputs
	}
	return TargetTaxInputs
}

// notesOf are the section's informative notes, in code order.
func notesOf(c ComputedSection, sources SourceCounts) []NoteCode {
	var out []NoteCode
	if c.Tax != nil && c.Tax.Balance.Kind == adomain.BalanceAFavor {
		if sources.EstimatedByApp > 0 {
			out = append(out, NoteRefundOnlyPaid)
		}
		if c.Tax.Balance.Amount >= efirmaRefund {
			out = append(out, NoteRefundNeedsEFirma)
		}
	}
	lossLines := []domain.LineCode{adomain.LineArt152BusinessLoss, adomain.LineArt152RentLoss, pm.LineNoProfit}
	if slices.ContainsFunc(c.Lines, func(l domain.Line) bool { return slices.Contains(lossLines, l.Code) }) {
		out = append(out, NoteLossCarryforward)
	}
	if c.Kind == adomain.SectionPM601 {
		out = append(out, NoteCUSuggestion)
	}
	return out
}

// obligationReasons explain the section's obligation (MODULE-annual §2.4).
func obligationReasons(c ComputedSection, y Year) []ReasonCode {
	switch {
	case c.Kind == adomain.SectionPM601 || c.Kind == adomain.SectionPMRESICO:
		return []ReasonCode{ReasonPersonaMoral}
	case c.Obligation == adomain.ObligationRelieved:
		return []ReasonCode{ReasonRESICORelief}
	case c.Obligation == adomain.ObligationOptional:
		return []ReasonCode{ReasonSalaryOnly}
	case c.Obligation == adomain.ObligationNotObligated:
		return []ReasonCode{ReasonNoIncome}
	case c.Obligation != adomain.ObligationObligated:
		return nil
	}
	return art152Reasons(c, y)
}

// incomeReasons are the chapters that oblige when they had income.
var incomeReasons = map[fiscal.RegimeCode]ReasonCode{
	fiscal.RegimeBusiness: ReasonBusinessIncome, fiscal.RegimePlatforms: ReasonPlatformIncome, fiscal.RegimeRent: ReasonRentIncome,
}

func art152Reasons(c ComputedSection, y Year) []ReasonCode {
	var out []ReasonCode
	for _, code := range c.Regimes {
		r, ok := y.regime(code)
		if reason, obliges := incomeReasons[code]; ok && obliges && r.Income > 0 {
			out = append(out, reason)
		}
	}
	simultaneous := 0
	flags := map[ReasonCode]bool{}
	for _, st := range y.SalaryStatements {
		flags[ReasonLeftBeforeYearEnd] = flags[ReasonLeftBeforeYearEnd] || st.LeftBeforeYearEnd
		flags[ReasonPayerNotWithholding] = flags[ReasonPayerNotWithholding] || st.PayerNotWithholding
		if st.Simultaneous {
			simultaneous++
		}
	}
	flags[ReasonSimultaneousEmployers] = simultaneous >= 2
	flags[ReasonToldEmployerWillFile] = y.Inputs.ToldEmployerWillFile && len(y.SalaryStatements) > 0
	flags[ReasonSalaryOverThreshold] = slices.Contains(c.CPATopics, adomain.TopicSalaryObligationBasis)
	for _, reason := range []ReasonCode{
		ReasonSimultaneousEmployers, ReasonLeftBeforeYearEnd, ReasonPayerNotWithholding, ReasonToldEmployerWillFile, ReasonSalaryOverThreshold,
	} {
		if flags[reason] {
			out = append(out, reason)
		}
	}
	return out
}

// estimateDisclaimers are every section's disclaimers, the year's and the
// RMF's, ordered and rendered (§2.8).
func estimateDisclaimers(est Estimate, w yearWindow) ([]Disclaimer, error) {
	facts := explain.DisclaimerFacts{RMFPending: est.FiscalYear <= rmfPendingYear}
	estimated := 0
	for _, s := range est.Sections {
		facts.Raised = append(facts.Raised, s.raised...)
		facts.Topics = append(facts.Topics, s.CPATopics...)
		estimated += s.Sources.EstimatedByApp
	}
	if w.inProgress {
		facts.Raised = append(facts.Raised, adomain.DisclaimerYearInProgress)
	}
	if est.InformativeIncomeOver500k {
		facts.Raised = append(facts.Raised, adomain.DisclaimerInformativeIncome)
	}
	return render(facts, explain.TextFacts{FiscalYear: est.FiscalYear, DataThrough: w.through, EstimatedMonths: estimated})
}

// sectionDisclaimers are the estimate's disclaimers that bear on s: the
// always-present ones, the year's and the section's own.
func sectionDisclaimers(est Estimate, s Section) ([]Disclaimer, error) {
	facts := explain.DisclaimerFacts{RMFPending: est.FiscalYear <= rmfPendingYear, Raised: s.raised, Topics: s.CPATopics}
	if est.YearStatus == YearInProgress {
		facts.Raised = append(slices.Clone(facts.Raised), adomain.DisclaimerYearInProgress)
	}
	through := time.Month(0)
	if !est.DataThrough.IsZero() {
		through = est.DataThrough.Month()
	}
	return render(facts, explain.TextFacts{FiscalYear: est.FiscalYear, DataThrough: through, EstimatedMonths: s.Sources.EstimatedByApp})
}

func render(facts explain.DisclaimerFacts, text explain.TextFacts) ([]Disclaimer, error) {
	ds, err := explain.Disclaimers(facts)
	if err != nil {
		return nil, err
	}
	out := make([]Disclaimer, len(ds))
	for i, d := range ds {
		out[i] = Disclaimer{Code: d.Code, Severity: d.Severity, Text: explain.DisclaimerText(d, text)}
	}
	return out, nil
}

// fingerprintLine is what the fingerprint reads of a line.
type fingerprintLine struct {
	Code      domain.LineCode
	Operation domain.Operation
	Amount    money.Cents
	Rate      money.Rate
	Args      []string
	Source    string
	Regime    fiscal.RegimeCode
}

// fingerprint is the SHA-256 of the estimate's canonical JSON (struct
// fields in declaration order, no map), with each line's source rendered.
func fingerprint(est Estimate) (string, error) {
	type section struct {
		Section
		Deadline string
		Lines    []fingerprintLine
	}
	view := struct {
		Estimate
		DataThrough, Deadline string
		Sections              []section
	}{Estimate: est, DataThrough: est.DataThrough.String(), Deadline: est.Deadline.String()}
	for _, s := range est.Sections {
		fs := section{Section: s, Deadline: s.Deadline.String()}
		for _, l := range s.Lines {
			fs.Lines = append(fs.Lines, fingerprintLine{
				Code: l.Code, Operation: l.Operation, Amount: l.Amount, Rate: l.Rate, Args: l.Args, Source: l.Source.String(), Regime: l.Regime,
			})
		}
		view.Sections = append(view.Sections, fs)
	}
	view.Fingerprint, view.RangeUsed = "", false // not figures
	data, err := json.Marshal(view)
	if err != nil {
		return "", fmt.Errorf("annual: fingerprint: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
