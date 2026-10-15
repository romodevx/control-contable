package explain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/explain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

const fiscalYear = 2026

var errParams = errors.New("params unavailable")

// fakeParams is the Params port: the 2026 annual dates and threshold.
type fakeParams struct {
	dates map[taxparams.Key]string
	money map[taxparams.Key]money.Cents
	err   error
	asked []string
}

func (f *fakeParams) Date(key taxparams.Key, on fiscal.Date) (fiscal.Date, error) {
	f.asked = append(f.asked, string(key)+"@"+on.String())
	if f.err != nil {
		return fiscal.Date{}, f.err
	}
	raw, ok := f.dates[key]
	if !ok {
		return fiscal.Date{}, fmt.Errorf("no date %s", key)
	}
	return fiscal.ParseDate(raw)
}

func (f *fakeParams) Money(key taxparams.Key, on fiscal.Date) (money.Cents, error) {
	f.asked = append(f.asked, string(key)+"@"+on.String())
	if f.err != nil {
		return 0, f.err
	}
	v, ok := f.money[key]
	if !ok {
		return 0, fmt.Errorf("no money %s", key)
	}
	return v, nil
}

func params2026() *fakeParams {
	return &fakeParams{
		dates: map[taxparams.Key]string{
			taxparams.KeyDeadlinePF: "2027-04-30", taxparams.KeyDeadlinePM: "2027-03-31",
			taxparams.KeyDeadlineNonprofitInformative: "2027-02-15",
		},
		money: map[taxparams.Key]money.Cents{taxparams.KeyInformativeIncomeThreshold: 500_000_00},
	}
}

func date(t testing.TB, s string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(s)
	require.NoError(t, err)
	return d
}

// TestSectionsExplainOnlyTheExplanationRegimes is the MODULE-annual §1
// table: no annual for 625 definitive, 615, 616, 610; explanation and
// deadline for the rest; nothing for the regimes another section covers.
func TestSectionsExplainOnlyTheExplanationRegimes(t *testing.T) {
	tests := []struct {
		name       string
		person     fiscal.PersonType
		held       explain.Held
		code       explain.Code
		obligation annual.Obligation
		deadline   explain.DeadlineKind
	}{
		{"625 definitive", fiscal.PersonFisica, explain.Held{Regime: "625", DefinitiveOption: true}, explain.CodePlatformDefinitive, annual.ObligationNotObligated, explain.DeadlineNone},
		{"615 prizes", fiscal.PersonFisica, explain.Held{Regime: "615"}, explain.CodePrizesDefinitive, annual.ObligationNotObligated, explain.DeadlineNone},
		{"616 no obligations", fiscal.PersonFisica, explain.Held{Regime: "616"}, explain.CodeNoObligations, annual.ObligationNotObligated, explain.DeadlineNone},
		{"610 PF foreign", fiscal.PersonFisica, explain.Held{Regime: "610"}, explain.CodeForeignResident, annual.ObligationNotObligated, explain.DeadlineNone},
		{"610 PM foreign", fiscal.PersonMoral, explain.Held{Regime: "610"}, explain.CodeForeignResident, annual.ObligationNotObligated, explain.DeadlineNone},
		{"603 nonprofit", fiscal.PersonMoral, explain.Held{Regime: "603"}, explain.CodeNonprofitInformative, annual.ObligationObligated, explain.DeadlineNonprofitInformative},
		{"607 sale of goods", fiscal.PersonFisica, explain.Held{Regime: "607"}, explain.CodeCPAComputes, annual.ObligationObligated, explain.DeadlinePF},
		{"608 other income", fiscal.PersonFisica, explain.Held{Regime: "608"}, explain.CodeCPAComputes, annual.ObligationObligated, explain.DeadlinePF},
		{"611 dividends", fiscal.PersonFisica, explain.Held{Regime: "611"}, explain.CodeCPAConfirms, "", explain.DeadlinePF},
		{"614 interest", fiscal.PersonFisica, explain.Held{Regime: "614"}, explain.CodeCPAConfirms, "", explain.DeadlinePF},
		{"621 RIF", fiscal.PersonFisica, explain.Held{Regime: "621"}, explain.CodeCPAConfirms, "", explain.DeadlinePF},
		{"620 cooperatives", fiscal.PersonMoral, explain.Held{Regime: "620"}, explain.CodeCPAComputes, annual.ObligationObligated, explain.DeadlineNone},
		{"622 AGAPES", fiscal.PersonMoral, explain.Held{Regime: "622"}, explain.CodeCPAComputes, annual.ObligationObligated, explain.DeadlinePM},
		{"623 groups", fiscal.PersonMoral, explain.Held{Regime: "623"}, explain.CodeCPAComputes, annual.ObligationObligated, explain.DeadlinePM},
		{"624 coordinados", fiscal.PersonMoral, explain.Held{Regime: "624"}, explain.CodeCPAComputes, annual.ObligationObligated, explain.DeadlinePM},
		{"unknown regime", fiscal.PersonMoral, explain.Held{Regime: "699"}, explain.CodeAskCPA, "", explain.DeadlineNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := explain.Sections(tt.person, []explain.Held{tt.held})

			require.Len(t, got, 1)
			s := got[0]
			assert.Equal(t, annual.SectionExplanation, s.Kind)
			assert.Equal(t, tt.held.Regime, s.Regime)
			assert.Equal(t, annual.StatusInformational, s.Status)
			assert.Equal(t, tt.code, s.Code)
			assert.Equal(t, tt.obligation, s.Obligation)
			assert.Equal(t, tt.deadline, s.Deadline)
			if tt.code != explain.CodeNoObligations && tt.code != explain.CodeAskCPA {
				assert.NotEmpty(t, s.Basis, "a legal basis for %s", tt.held.Regime)
			}
			for _, c := range s.Basis {
				assert.True(t, c.Valid(), "%+v", c)
			}
		})
	}
}

func TestSectionsSkipTheComputedRegimes(t *testing.T) {
	pf := []explain.Held{{Regime: "605"}, {Regime: "606"}, {Regime: "612"}, {Regime: "625"}, {Regime: "626"}}
	pm := []explain.Held{{Regime: "601"}, {Regime: "626"}}

	assert.Empty(t, explain.Sections(fiscal.PersonFisica, pf))
	assert.Empty(t, explain.Sections(fiscal.PersonMoral, pm))
}

// TestSectionsAreInRegimeOrder: deterministic, one per regime held.
func TestSectionsAreInRegimeOrder(t *testing.T) {
	got := explain.Sections(fiscal.PersonFisica, []explain.Held{{Regime: "621"}, {Regime: "612"}, {Regime: "607"}, {Regime: "607"}})

	require.Len(t, got, 2)
	assert.Equal(t, fiscal.RegimeCode("607"), got[0].Regime)
	assert.Equal(t, fiscal.RegimeCode("621"), got[1].Regime)
}

// TestDeadlinesComeFromParams: PF 30-abr-2027, PM 31-mar-2027, 603
// informative 15-feb-2027 (MODULE-annual §2.9), read in force on the
// fiscal year's last day; the business-day move is the calendar's.
func TestDeadlinesComeFromParams(t *testing.T) {
	tests := []struct {
		kind explain.DeadlineKind
		want string
	}{
		{explain.DeadlinePF, "2027-04-30"},
		{explain.DeadlinePM, "2027-03-31"},
		{explain.DeadlineNonprofitInformative, "2027-02-15"},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			p := params2026()

			got, err := explain.Deadline(p, fiscalYear, tt.kind)

			require.NoError(t, err)
			assert.Equal(t, date(t, tt.want), got)
			assert.Len(t, p.asked, 1)
			assert.True(t, strings.HasSuffix(p.asked[0], "@2026-12-31"), p.asked)
		})
	}
}

func TestDeadlineErrors(t *testing.T) {
	_, err := explain.Deadline(params2026(), fiscalYear, explain.DeadlineNone)
	require.ErrorIs(t, err, explain.ErrNoDeadline)

	_, err = explain.Deadline(params2026(), fiscalYear, "weekly")
	require.ErrorIs(t, err, explain.ErrNoDeadline)

	_, err = explain.Deadline(&fakeParams{err: errParams}, fiscalYear, explain.DeadlinePM)
	require.ErrorIs(t, err, errParams)

	_, err = explain.Deadline(params2026(), 0, explain.DeadlinePM)
	require.Error(t, err)
}

func TestSectionDeadline(t *testing.T) {
	assert.Equal(t, explain.DeadlinePF, explain.SectionDeadline(annual.SectionPFArt152))
	assert.Equal(t, explain.DeadlinePF, explain.SectionDeadline(annual.SectionPFRESICOCheck))
	assert.Equal(t, explain.DeadlinePM, explain.SectionDeadline(annual.SectionPM601))
	assert.Equal(t, explain.DeadlinePM, explain.SectionDeadline(annual.SectionPMRESICO))
	assert.Equal(t, explain.DeadlineNone, explain.SectionDeadline(annual.SectionExplanation), "an explanation carries its own")
}

// TestInformativeIncomeAboveTheThreshold: total income including exempt
// and definitively taxed income strictly above $500,000 (LISR 150 ¶3).
func TestInformativeIncomeAboveTheThreshold(t *testing.T) {
	for total, want := range map[money.Cents]bool{500_000_00: false, 500_000_01: true, 0: false} {
		got, err := explain.InformativeIncome(params2026(), fiscalYear, total)

		require.NoError(t, err)
		assert.Equal(t, want, got, total.String())
	}
	_, err := explain.InformativeIncome(&fakeParams{err: errParams}, fiscalYear, 1)
	require.ErrorIs(t, err, errParams)
	_, err = explain.InformativeIncome(params2026(), fiscalYear, -1)
	require.ErrorIs(t, err, explain.ErrNegativeIncome)
}

// TestDisclaimersOrderAndTopics: annual_estimate first, then strong,
// warning, info, then code (MODULE-annual §2.8); no_annual_iva always;
// cpa_review_pending names every topic once.
func TestDisclaimersOrderAndTopics(t *testing.T) {
	got, err := explain.Disclaimers(explain.DisclaimerFacts{
		Raised: []annual.DisclaimerCode{
			annual.DisclaimerInflationAdjustmentExcluded, annual.DisclaimerPaymentsEstimated,
			annual.DisclaimerYearInProgress, annual.DisclaimerInflationAdjustmentExcluded,
		},
		Topics:     []annual.CPATopic{annual.TopicPTURate, annual.TopicCURounding, annual.TopicPTURate},
		RMFPending: true,
	})

	require.NoError(t, err)
	assert.Equal(t, []explain.Disclaimer{
		{Code: annual.DisclaimerAnnualEstimate, Severity: annual.SeverityInfo},
		{Code: annual.DisclaimerCPAReviewPending, Severity: annual.SeverityStrong, Topics: []annual.CPATopic{annual.TopicCURounding, annual.TopicPTURate}},
		{Code: annual.DisclaimerPaymentsEstimated, Severity: annual.SeverityWarning},
		{Code: annual.DisclaimerYearInProgress, Severity: annual.SeverityWarning},
		{Code: annual.DisclaimerInflationAdjustmentExcluded, Severity: annual.SeverityInfo},
		{Code: annual.DisclaimerNoAnnualIVA, Severity: annual.SeverityInfo},
		{Code: annual.DisclaimerRMF2027Pending, Severity: annual.SeverityInfo},
	}, got)
}

func TestDisclaimersMinimal(t *testing.T) {
	got, err := explain.Disclaimers(explain.DisclaimerFacts{})

	require.NoError(t, err)
	assert.Equal(t, []explain.Disclaimer{
		{Code: annual.DisclaimerAnnualEstimate, Severity: annual.SeverityInfo},
		{Code: annual.DisclaimerNoAnnualIVA, Severity: annual.SeverityInfo},
	}, got)
}

func TestDisclaimersRefuseUnknownCodesAndTopics(t *testing.T) {
	_, err := explain.Disclaimers(explain.DisclaimerFacts{Raised: []annual.DisclaimerCode{"made_up"}})
	require.ErrorIs(t, err, explain.ErrUnknownCode)

	_, err = explain.Disclaimers(explain.DisclaimerFacts{Topics: []annual.CPATopic{"made_up"}})
	require.ErrorIs(t, err, explain.ErrUnknownCode)

	_, err = explain.Disclaimers(explain.DisclaimerFacts{Raised: []annual.DisclaimerCode{annual.DisclaimerCPAReviewPending}})
	require.ErrorIs(t, err, explain.ErrUnknownCode, "cpa_review_pending is raised by its topics")
}
