package annualhttp

import (
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// fieldPaidOn is a deduction's payment date.
const fieldPaidOn = "paid_on"

// invalidDate backs up the router's date-format validation.
func invalidDate(field string) error {
	return apperr.New(apperr.CodeValidationFailed, "One or more fields are invalid.").
		WithFields(apperr.FieldError{Field: field, Code: "invalid", Message: "La fecha no es válida."})
}

func toEstimate(taxpayerID gen.Uuid, e usecaseannual.Estimate) gen.AnnualEstimate {
	set := toSet(e.ParameterSet)
	out := gen.AnnualEstimate{
		TaxpayerId: taxpayerID, FiscalYear: e.FiscalYear, PersonType: gen.PersonType(e.PersonType),
		Legend: gen.AnnualLegend(e.Legend), Disclaimers: toDisclaimers(e.Disclaimers),
		YearStatus: gen.AnnualYearStatus(e.YearStatus), Deadline: toDate(e.Deadline),
		HeadlineSections:          make([]gen.AnnualComputedSectionKind, 0, len(e.HeadlineSections)),
		Sections:                  make([]gen.AnnualSection, 0, len(e.Sections)),
		InformativeIncomeOver500k: e.InformativeIncomeOver500k, ParameterSet: set, EngineVersion: e.EngineVersion,
	}
	if !e.DataThrough.IsZero() {
		p := e.DataThrough.String()
		out.DataThrough = &p
	}
	for _, k := range e.HeadlineSections {
		out.HeadlineSections = append(out.HeadlineSections, gen.AnnualComputedSectionKind(k))
	}
	for _, s := range e.Sections {
		out.Sections = append(out.Sections, toSection(s, e.FiscalYear, set))
	}
	return out
}

func toBreakdown(taxpayerID gen.Uuid, b usecaseannual.Breakdown) gen.AnnualBreakdown {
	return gen.AnnualBreakdown{
		TaxpayerId: taxpayerID, FiscalYear: b.FiscalYear,
		Section: toSection(b.Section, b.FiscalYear, toSet(b.ParameterSet)),
		Lines:   toLines(b.Section, b.FiscalYear), Legend: gen.AnnualLegend(b.Legend),
		Disclaimers: toDisclaimers(b.Disclaimers), EngineVersion: b.EngineVersion,
	}
}

func toSet(s taxparams.SetID) gen.ParameterSetRef {
	return gen.ParameterSetRef{FiscalYear: s.FiscalYear, Version: s.Version}
}

func toDisclaimers(ds []usecaseannual.Disclaimer) []gen.Disclaimer {
	out := make([]gen.Disclaimer, 0, len(ds))
	for _, d := range ds {
		out = append(out, gen.Disclaimer{Code: gen.DisclaimerCode(d.Code), Severity: gen.DisclaimerSeverity(d.Severity), Text: d.Text})
	}
	return out
}

func toSection(s usecaseannual.Section, fiscalYear int, set gen.ParameterSetRef) gen.AnnualSection {
	out := gen.AnnualSection{
		Kind: gen.AnnualSectionKind(s.Kind), Regimes: make([]gen.RegimeCode, 0, len(s.Regimes)),
		Status: gen.AnnualSectionStatus(s.Status), MissingInputs: make([]gen.AnnualMissingInput, 0, len(s.MissingInputs)),
		ObligationReasons: make([]string, 0, len(s.ObligationReasons)), Deadline: toDate(s.Deadline),
		Notes: make([]gen.AnnualNote, 0, len(s.Notes)), CpaReviewPending: len(s.CPATopics) > 0,
		CpaTopics: make([]gen.CpaTopic, 0, len(s.CPATopics)), ParameterSet: set,
	}
	for _, r := range s.Regimes {
		out.Regimes = append(out.Regimes, string(r))
	}
	for _, m := range s.MissingInputs {
		out.MissingInputs = append(out.MissingInputs, toMissing(m))
	}
	if s.Obligation != "" {
		o := gen.AnnualObligation(s.Obligation)
		out.Obligation = &o
	}
	for _, r := range s.ObligationReasons {
		out.ObligationReasons = append(out.ObligationReasons, reasonText(r))
	}
	for _, n := range s.Notes {
		out.Notes = append(out.Notes, gen.AnnualNote{Code: gen.AnnualNoteCode(n), Text: noteText(n, fiscalYear)})
	}
	for _, t := range s.CPATopics {
		out.CpaTopics = append(out.CpaTopics, gen.CpaTopic(t))
	}
	if s.Explanation != "" {
		text := s.Explanation
		out.Explanation = &text
	}
	out.Tax = toTax(s)
	out.PmExtras = toExtras(s.PMExtras)
	if c := s.RESICOCheck; c != nil {
		out.ResicoCheck = &gen.ResicoPfCheck{
			IncomeCents: int64(c.Income), Rate: c.Rate.String(), AnnualIsrCents: int64(c.AnnualISR),
			MonthlyPaymentsCents: int64(c.Payments), WithheldCents: int64(c.Withheld), DifferenceCents: int64(c.Difference),
		}
	}
	return out
}

func toTax(s usecaseannual.Section) *gen.AnnualTax {
	t := s.Tax
	if t == nil {
		return nil
	}
	out := &gen.AnnualTax{
		AccumulableIncomeCents: int64(t.AccumulableIncome), BaseCents: int64(t.Base), IsrCents: int64(t.ISR),
		Credits: gen.AnnualCredits{
			PaymentsCents: int64(t.Payments), WithholdingsCents: int64(t.Withholdings), SalaryIsrCents: int64(t.SalaryISR),
			TotalCents: int64(s.Credits),
			PaymentSources: gen.PaymentSourceCounts{
				FiledMonths: s.Sources.Filed, TaxInputMonths: s.Sources.TaxInput, EstimatedByAppMonths: s.Sources.EstimatedByApp,
			},
		},
		Balance: gen.AnnualBalance{Kind: gen.AnnualBalanceKind(t.Balance.Kind), AmountCents: int64(t.Balance.Amount)},
	}
	if t.PersonalDeductions != nil {
		v := int64(*t.PersonalDeductions)
		out.PersonalDeductionsCents = &v
	}
	return out
}

// cuScale is the 10^4 of the suggested coeficiente's four decimals.
const cuScale = 10_000

func toExtras(e *usecaseannual.PMExtras) *gen.PmAnnualExtras {
	if e == nil {
		return nil
	}
	out := &gen.PmAnnualExtras{PtuBaseCents: int64(e.PTUBase)}
	if e.PTUToDistribute != nil {
		v := int64(*e.PTUToDistribute)
		out.PtuToDistributeCents = &v
	}
	if e.SuggestedCU != nil {
		cu := fourDecimals(*e.SuggestedCU)
		out.SuggestedCuNextYear = &cu
	}
	return out
}

// fourDecimals renders a rate with exactly four decimals ("0.1380").
func fourDecimals(r money.Rate) string {
	n := int64(r) / (money.RateScale / cuScale)
	return formatFixed(n, cuScale)
}

func toMissing(m usecaseannual.MissingInput) gen.AnnualMissingInput {
	out := gen.AnnualMissingInput{
		Key: m.Key, Label: missingLabel(m.Key), Target: gen.AnnualInputTarget(m.Target), Months: make([]gen.Period, 0, len(m.Months)),
	}
	if m.Regime != "" {
		r := string(m.Regime)
		out.RegimeCode = &r
	}
	for _, p := range m.Months {
		out.Months = append(out.Months, p.String())
	}
	return out
}

func toLines(s usecaseannual.Section, fiscalYear int) []gen.AnnualBreakdownLine {
	out := make([]gen.AnnualBreakdownLine, 0, len(s.Lines))
	for _, l := range s.Lines {
		line := gen.AnnualBreakdownLine{
			Seq: l.Seq, Concept: concept(l.Line, fiscalYear), Operation: gen.AnnualBreakdownLineOperation(l.Operation),
			AmountCents: int64(l.Amount), LegalBasis: make([]gen.Citation, 0, len(l.Basis)), Source: lineSource(l.Line),
			NeedsCpaReview: l.NeedsCPAReview,
		}
		if l.Regime != "" {
			r := string(l.Regime)
			line.RegimeCode = &r
		}
		if l.Operation == domain.OpMultiply {
			rate := l.Rate.String()
			line.Rate = &rate
		}
		for _, c := range l.Basis {
			cite := gen.Citation{Law: c.Law, Article: c.Article}
			if c.Note != "" {
				note := c.Note
				cite.Note = &note
			}
			line.LegalBasis = append(line.LegalBasis, cite)
		}
		out = append(out, line)
	}
	return out
}

// toFiled joins the stored months with the estimate's columns: a month
// outside the window or not closed has no app estimate and no payment.
func toFiled(view usecaseannual.FiledView, p usecaseannual.RegimePayments) gen.FiledPayments {
	out := gen.FiledPayments{
		RegimeCode: string(view.Regime), FiscalYear: view.FiscalYear,
		Months: make([]gen.FiledPaymentMonth, 0, len(view.Months)), UpdatedAt: view.UpdatedAt,
	}
	for i, m := range view.Months {
		month := gen.FiledPaymentMonth{
			Month: int(m.Month), Period: m.Period.String(), InWindow: m.InWindow, FiledCents: int64Ptr(m.Filed),
			Source: gen.FiledPaymentSource(sourceNone),
		}
		if i < len(p.Months) {
			col := p.Months[i]
			month.TaxInputCents = int64Ptr(col.TaxInput)
			if m.InWindow && m.Closed {
				month.AppEstimateCents = int64Ptr(col.App.Amount())
				month.AppEstimateIncomplete = col.App.Incomplete
				month.EffectiveCents = int64Ptr(col.Effective)
				if col.Source != "" {
					month.Source = gen.FiledPaymentSource(col.Source)
				}
			}
		}
		if m.InWindow && m.Closed && m.Filed != nil { // the stored month wins (precedence 1)
			month.EffectiveCents, month.Source = month.FiledCents, gen.FiledPaymentSource(adomain.PaymentFiled)
		}
		if month.EffectiveCents != nil {
			out.TotalEffectiveCents += *month.EffectiveCents
		}
		out.Months = append(out.Months, month)
	}
	return out
}

// sourceNone is FiledPaymentSource none: no payment for the month.
const sourceNone = "none"

func toInputs(v usecaseannual.InputsView) gen.AnnualInputs {
	out := gen.AnnualInputs{
		FiscalYear: v.FiscalYear, OtherExemptIncomeCents: int64(v.Facts.OtherExemptIncome),
		PriorYearAccumulableIncomeCents: int64Ptr(v.Facts.PriorYearAccumulableIncome),
		ToldEmployerWillFile:            v.Facts.ToldEmployerWillFile, UpdatedAt: v.UpdatedAt,
		RegimeTaxInputs: make([]gen.AnnualRegimeTaxInput, 0, len(v.RegimeTaxInputs)),
	}
	for _, r := range v.RegimeTaxInputs {
		out.RegimeTaxInputs = append(out.RegimeTaxInputs, gen.AnnualRegimeTaxInput{
			RegimeCode: string(r.Regime), Key: string(r.Key), Label: r.Label, ValueCents: int64Ptr(r.Value),
		})
	}
	return out
}

func toSalaryList(l usecaseannual.SalaryList) gen.SalaryStatementList {
	out := gen.SalaryStatementList{
		FiscalYear: l.FiscalYear, Items: make([]gen.SalaryStatement, 0, len(l.Items)),
		Totals: gen.SalaryStatementTotals{
			AccumulableCents: int64(l.Totals.Accumulable), ExemptCents: int64(l.Totals.Exempt), IsrWithheldCents: int64(l.Totals.ISRWithheld),
		},
	}
	for _, s := range l.Items {
		out.Items = append(out.Items, toStatement(s, nil))
	}
	return out
}

func toStatement(s usecaseannual.StoredStatement, warnings []usecaseannual.Warning) gen.SalaryStatement {
	out := gen.SalaryStatement{
		Id: mustUUID(s.ID), TaxpayerId: mustUUID(s.TaxpayerID), FiscalYear: s.FiscalYear,
		AccumulableCents: int64(s.Accumulable), ExemptCents: int64(s.Exempt), IsrWithheldCents: int64(s.ISRWithheld),
		Simultaneous: s.Simultaneous, LeftBeforeYearEnd: s.LeftBeforeYearEnd, PayerNotWithholding: s.PayerNotWithholding,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt, Warnings: toWarnings(warnings),
	}
	if s.EmployerLabel != "" {
		label := s.EmployerLabel
		out.EmployerLabel = &label
	}
	return out
}

func toDeductionList(l usecaseannual.DeductionList) gen.PersonalDeductionList {
	out := gen.PersonalDeductionList{
		FiscalYear: l.FiscalYear, Items: make([]gen.PersonalDeduction, 0, len(l.Items)),
		Totals: make([]gen.DeductionCategoryTotal, 0, len(l.Totals)),
	}
	for _, d := range l.Items {
		out.Items = append(out.Items, toDeduction(d, nil))
	}
	for _, t := range l.Totals {
		out.Totals = append(out.Totals, gen.DeductionCategoryTotal{
			Category: gen.DeductionCategory(t.Category), RegisteredCents: int64(t.Registered),
			ExcludedCents: int64(t.Excluded), Count: t.Count,
		})
	}
	return out
}

func toDeduction(d usecaseannual.Deduction, warnings []usecaseannual.Warning) gen.PersonalDeduction {
	out := gen.PersonalDeduction{
		Id: mustUUID(d.ID), TaxpayerId: mustUUID(d.TaxpayerID), FiscalYear: d.FiscalYear,
		Category: gen.DeductionCategory(d.Category), AmountCents: int64(d.Amount), PaidOn: *toDate(d.PaidOn),
		PaymentMethod: gen.DeductionPaymentMethod(d.PaymentMethod), Beneficiary: gen.DeductionBeneficiary(d.Beneficiary),
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, Warnings: toWarnings(warnings),
	}
	if d.DoneeType != "" {
		v := gen.DoneeType(d.DoneeType)
		out.DoneeType = &v
	}
	if d.StudentLabel != "" {
		v := d.StudentLabel
		out.StudentLabel = &v
	}
	if d.TuitionLevel != "" {
		v := gen.TuitionLevel(d.TuitionLevel)
		out.TuitionLevel = &v
	}
	if d.Scholarship.Valid {
		v := int64(d.Scholarship.Cents)
		out.ScholarshipCents = &v
	}
	if d.CFDIUUID != "" {
		v := d.CFDIUUID
		out.CfdiUuid = &v
	}
	if d.Excluded != "" {
		v := gen.DeductionExclusion(d.Excluded)
		out.Excluded = &v
	}
	return out
}

func toWarnings(ws []usecaseannual.Warning) []gen.Warning {
	out := make([]gen.Warning, 0, len(ws))
	for _, w := range ws {
		out = append(out, gen.Warning{Code: gen.WarningCode(w.Code), Field: w.Field, Message: warningText(w.Code)})
	}
	return out
}

// toDate converts a fiscal date; nil for the zero date.
func toDate(d fiscal.Date) *openapi_types.Date {
	if d.IsZero() {
		return nil
	}
	t, err := time.Parse(time.DateOnly, d.String())
	if err != nil { // a valid fiscal.Date always parses
		return nil
	}
	return &openapi_types.Date{Time: t}
}

// mustUUID parses an id the store returned canonical; the zero UUID never
// reaches a response from a valid row.
func mustUUID(raw string) gen.Uuid {
	id, err := uuid.Parse(raw)
	if err != nil {
		return gen.Uuid{}
	}
	return id
}

func int64Ptr(p *money.Cents) *int64 {
	if p == nil {
		return nil
	}
	v := int64(*p)
	return &v
}
