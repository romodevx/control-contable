package annual_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/explain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func TestNewEstimator_refusesAMissingPort(t *testing.T) {
	_, err := annual.NewEstimator(annual.EstimateDeps{})

	require.ErrorIs(t, err, annual.ErrMissingDependency)
}

// Order of refusals: the taxpayer id (404), a future year (400), a year
// before the parameters (422), then the taxpayer itself (404).
func TestEstimate_refusals(t *testing.T) {
	w := newAnnWorld()
	w.now = time.Date(2026, time.October, 15, 12, 0, 0, 0, clock.Location())
	uc := w.estimator(t)

	_, err := uc.Estimate(t.Context(), userA, "not-a-uuid", 2026)
	requireCode(t, err, apperr.CodeNotFound)
	_, err = uc.Estimate(t.Context(), userA, fisicaA, 2027)
	assert.Equal(t, []string{"fiscalYear/out_of_range"}, requireCode(t, err, apperr.CodeValidationFailed))
	_, err = uc.Estimate(t.Context(), userA, fisicaA, 2025)
	assert.Equal(t, []string{"fiscalYear/parameters_unavailable"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
	_, err = uc.Estimate(t.Context(), userA, fisicaB, 2026)
	requireCode(t, err, apperr.CodeNotFound)
	assert.Empty(t, w.monthly.calls)
}

// US-A1 / AN-01: a closed year whose payments were all filed. No range
// call, one read-only unit of work, the spec's figures, deadline and
// legend.
func TestEstimate_AN01_closedYearFiled(t *testing.T) {
	w := newAnnWorld()
	w.an01()
	w.fileYear("612", 55_773_86)

	got, err := w.estimator(t).Estimate(t.Context(), userA, fisicaA, 2026)

	require.NoError(t, err)
	assert.Equal(t, []string{"ro"}, w.tx.modes)
	assert.Empty(t, w.monthly.calls, "every month is filed: no default needed")
	assert.False(t, got.RangeUsed)
	assert.Equal(t, fisicaA, got.TaxpayerID)
	assert.Equal(t, annual.YearClosed, got.YearStatus)
	assert.Equal(t, mustPeriod("2026-12"), got.DataThrough)
	assert.Equal(t, explain.Legend, got.Legend)
	assert.Equal(t, mustDate("2027-04-30"), got.Deadline)
	assert.Equal(t, []adomain.SectionKind{adomain.SectionPFArt152}, got.HeadlineSections)
	assert.True(t, got.InformativeIncomeOver500k)
	assert.Equal(t, domain.EngineVersion, got.EngineVersion)
	assert.Equal(t, 2026, got.ParameterSet.FiscalYear)
	require.Len(t, got.Sections, 1)
	s := got.Sections[0]
	assert.Equal(t, adomain.StatusComplete, s.Status)
	assert.Equal(t, adomain.ObligationObligated, s.Obligation)
	assert.Equal(t, []annual.ReasonCode{annual.ReasonBusinessIncome}, s.ObligationReasons)
	assert.Equal(t, mustDate("2027-04-30"), s.Deadline)
	require.NotNil(t, s.Tax)
	assert.Equal(t, money.Cents(76_365_86), s.Tax.ISR)
	assert.Equal(t, adomain.Balance{Kind: adomain.BalanceAFavor, Amount: 9_408_00}, s.Tax.Balance)
	assert.Equal(t, annual.SourceCounts{Filed: 12}, s.Sources)
	assert.Equal(t, money.Cents(85_773_86), s.Credits)
	assert.Empty(t, s.Notes, "a refund under 10,000 from filed payments needs no note")
	codes := disclaimerCodes(got.Disclaimers)
	assert.Equal(t, adomain.DisclaimerAnnualEstimate, codes[0])
	assert.ElementsMatch(t, []adomain.DisclaimerCode{
		adomain.DisclaimerAnnualEstimate, adomain.DisclaimerDeductionsAsEntered, adomain.DisclaimerInformativeIncome,
		adomain.DisclaimerNoAnnualIVA, adomain.DisclaimerRMF2027Pending,
	}, codes)
	for _, d := range got.Disclaimers {
		assert.NotEmpty(t, d.Text, d.Code)
	}
}

// US-A9: with nothing filed every month's default is the app's estimate,
// from one Estimator range call inside the unit of work. A cumulative
// regime's defaults are chained: their sum is the year's cumulative ISR,
// not the sum of the cumulative estimates.
func TestEstimate_defaultPaymentsComeFromTheRange(t *testing.T) {
	w := newAnnWorld()
	w.an01()
	cumulative := map[time.Month]money.Cents{}
	for m := time.January; m <= time.December; m++ {
		cumulative[m] = money.Cents(int64(m) * 5_000_00) // YTD tax, nothing typed
	}
	w.monthly.isr["612"] = cumulative

	got, err := w.estimator(t).Estimate(t.Context(), userA, fisicaA, 2026)

	require.NoError(t, err)
	assert.Equal(t, []string{userA + "|" + userA + "|2026-01..2026-12"}, w.monthly.calls)
	assert.True(t, got.RangeUsed)
	s := got.Sections[0]
	assert.Equal(t, adomain.StatusComplete, s.Status, "defaults are not missing inputs")
	assert.Equal(t, annual.SourceCounts{EstimatedByApp: 12}, s.Sources)
	assert.Equal(t, money.Cents(12*5_000_00), s.Tax.Payments)
	assert.Contains(t, disclaimerCodes(got.Disclaimers), adomain.DisclaimerPaymentsEstimated)
	assert.Equal(t, adomain.BalanceAFavor, s.Tax.Balance.Kind)
	assert.Contains(t, s.Notes, annual.NoteRefundOnlyPaid)
	assert.Contains(t, s.Notes, annual.NoteRefundNeedsEFirma)
	for _, l := range s.Lines {
		if l.Code == adomain.LineArt152Payments {
			assert.Equal(t, fiscal.RegimeCode("612"), l.Regime)
		}
	}
}

// A default from an incomplete monthly estimate makes the section
// incomplete and names the months to confirm (§2.6).
func TestEstimate_incompleteMonthlyEstimate(t *testing.T) {
	w := newAnnWorld()
	w.an01()
	w.fileYear("612", 55_773_86)
	delete(w.filed.byRegime["612"], time.March)
	w.monthly.isr["612"] = map[time.Month]money.Cents{time.March: 1_000_00}
	w.monthly.incomplete = map[fiscal.RegimeCode][]time.Month{"612": {time.March}}

	got, err := w.estimator(t).Estimate(t.Context(), userA, fisicaA, 2026)

	require.NoError(t, err)
	s := got.Sections[0]
	assert.Equal(t, adomain.StatusIncomplete, s.Status)
	assert.Equal(t, []annual.MissingInput{{
		Key: "filed_payment", Target: annual.TargetFiledPayments, Regime: "612", Months: []domain.Period{mustPeriod("2026-03")},
	}}, s.MissingInputs)
	assert.Contains(t, disclaimerCodes(got.Disclaimers), adomain.DisclaimerIncompleteInputs)
}

// US-A12: the current year counts the records through the last closed
// month, with the year_in_progress warning, and never extrapolates.
func TestEstimate_yearInProgress(t *testing.T) {
	w := newAnnWorld()
	w.now = time.Date(2026, time.October, 15, 12, 0, 0, 0, clock.Location())
	w.income("612", 2026, time.September, 100_000_00, 0)
	w.income("612", 2026, time.October, 999_000_00, 0) // the open month never counts
	w.monthly.isr["612"] = map[time.Month]money.Cents{}

	got, err := w.estimator(t).Estimate(t.Context(), userA, fisicaA, 2026)

	require.NoError(t, err)
	assert.Equal(t, annual.YearInProgress, got.YearStatus)
	assert.Equal(t, mustPeriod("2026-09"), got.DataThrough)
	assert.Equal(t, time.September, w.ledger.queries[0].Through)
	assert.Equal(t, []string{userA + "|" + userA + "|2026-01..2026-09"}, w.monthly.calls)
	assert.Equal(t, money.Cents(100_000_00), got.Sections[0].Tax.AccumulableIncome)
	assert.Contains(t, disclaimerCodes(got.Disclaimers), adomain.DisclaimerYearInProgress)
	assert.Equal(t, annual.SourceCounts{EstimatedByApp: 9}, got.Sections[0].Sources)
}

// In January of the current year no month is closed: nothing is read
// from the ledger nor estimated, and data_through is null.
func TestEstimate_januaryOfTheCurrentYear(t *testing.T) {
	w := newAnnWorld()
	w.now = time.Date(2026, time.January, 20, 12, 0, 0, 0, clock.Location())

	got, err := w.estimator(t).Estimate(t.Context(), userA, fisicaA, 2026)

	require.NoError(t, err)
	assert.True(t, got.DataThrough.IsZero())
	assert.Empty(t, w.ledger.queries)
	assert.Empty(t, w.monthly.calls)
	assert.Equal(t, adomain.ObligationNotObligated, got.Sections[0].Obligation)
}

// US-A7 + US-A11: a PM's 601 section and its 603 explanation, each with its
// own deadline; the estimate's deadline is the earliest, moved to the next
// business day.
func TestEstimate_personaMoral(t *testing.T) {
	w := newAnnWorld()
	w.income("601", 2026, time.May, 10_000_000_00, 0)
	w.expense("601", 2026, time.May, 8_500_000_00)
	w.taxInputs.values["601"] = []domain.TaxInputValue{
		{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: 120_000_00},
		{Key: domain.InputPriorYearLosses, Kind: domain.InputMoney, Money: 200_000_00},
		{Key: domain.InputInterestWithholding, Kind: domain.InputMoney, Money: 2_000_00},
	}
	w.fileYear("601", 330_000_00)
	w.holidays = annHolidays{days: []string{"2027-02-15"}}

	got, err := w.estimator(t).Estimate(t.Context(), userA, moralA, 2026)

	require.NoError(t, err)
	require.Len(t, got.Sections, 2)
	pm, expl := got.Sections[0], got.Sections[1]
	assert.Equal(t, adomain.SectionPM601, pm.Kind)
	assert.Equal(t, adomain.Balance{Kind: adomain.BalanceACargo, Amount: 22_000_00}, pm.Tax.Balance)
	assert.Equal(t, mustDate("2027-03-31"), pm.Deadline)
	assert.Equal(t, []annual.ReasonCode{annual.ReasonPersonaMoral}, pm.ObligationReasons)
	assert.Contains(t, pm.Notes, annual.NoteCUSuggestion)
	assert.Equal(t, adomain.SectionExplanation, expl.Kind)
	assert.Equal(t, []fiscal.RegimeCode{"603"}, expl.Regimes)
	assert.Equal(t, adomain.StatusInformational, expl.Status)
	assert.Equal(t, mustDate("2027-02-16"), expl.Deadline, "15 February is a holiday here")
	assert.Contains(t, expl.Explanation, "16 de febrero de 2027")
	assert.Nil(t, expl.Tax)
	assert.Equal(t, mustDate("2027-02-16"), got.Deadline)
	assert.Equal(t, []adomain.SectionKind{adomain.SectionPM601}, got.HeadlineSections)
	assert.Contains(t, disclaimerCodes(got.Disclaimers), adomain.DisclaimerCPAReviewPending)
}

// The breakdown is the estimate seen from one section: its lines carry
// their regime, the disclaimers bear on it; a kind the estimate lacks is
// 422 section_not_present.
func TestBreakdown(t *testing.T) {
	w := newAnnWorld()
	w.an01()
	w.fileYear("612", 55_773_86)
	uc := w.estimator(t)

	got, err := uc.Breakdown(t.Context(), userA, fisicaA, 2026, "pf_art152")

	require.NoError(t, err)
	assert.Equal(t, adomain.SectionPFArt152, got.Section.Kind)
	require.NotEmpty(t, got.Section.Lines)
	assert.Equal(t, fiscal.RegimeCode("612"), got.Section.Lines[0].Regime)
	assert.Equal(t, explain.Legend, got.Legend)
	assert.Contains(t, disclaimerCodes(got.Disclaimers), adomain.DisclaimerDeductionsAsEntered)

	_, err = uc.Breakdown(t.Context(), userA, fisicaA, 2026, "pm_601")
	assert.Equal(t, []string{"section/section_not_present"}, requireCode(t, err, apperr.CodeBusinessRuleViolation))
}

// Same inputs, same estimate and fingerprint; any figure that moves
// moves the fingerprint (MODULE-annual §2.11).
func TestEstimate_isDeterministic(t *testing.T) {
	w := newAnnWorld()
	w.an01()
	w.fileYear("612", 55_773_86)
	uc := w.estimator(t)

	a, err := uc.Estimate(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	b, err := uc.Estimate(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	assert.Equal(t, a, b)
	assert.Len(t, a.Fingerprint, 64)

	w.inputs.in.Deductions[0].Amount = 41_000_00
	c, err := uc.Estimate(t.Context(), userA, fisicaA, 2026)
	require.NoError(t, err)
	assert.NotEqual(t, a.Fingerprint, c.Fingerprint)
}

// The filed-payments screen's columns: each month's filed value, tax
// input, the app's (chained) estimate and the one credited.
func TestPayments_columns(t *testing.T) {
	w := newAnnWorld()
	w.now = time.Date(2026, time.April, 15, 12, 0, 0, 0, clock.Location())
	w.filed.byRegime["612"] = map[time.Month]money.Cents{time.January: 500_00}
	w.monthly.isr["612"] = map[time.Month]money.Cents{time.January: 400_00, time.February: 900_00, time.March: 1_500_00}

	got, err := w.estimator(t).Payments(t.Context(), userA, fisicaA, "612", 2026)

	require.NoError(t, err)
	assert.Equal(t, []string{userA + "|" + userA + "|2026-01..2026-03"}, w.monthly.calls)
	require.Len(t, got.Months, 12)
	assert.Equal(t, ptr(money.Cents(400_00)), got.Months[0].App.Amount())
	assert.Equal(t, adomain.PaymentFiled, got.Months[0].Source)
	assert.Equal(t, ptr(money.Cents(400_00)), got.Months[1].App.Amount(), "900 − the 500 filed in January")
	assert.Nil(t, got.Months[3].Effective, "April is not closed")
}

func disclaimerCodes(ds []annual.Disclaimer) []adomain.DisclaimerCode {
	out := make([]adomain.DisclaimerCode, len(ds))
	for i, d := range ds {
		out[i] = d.Code
	}
	return out
}
