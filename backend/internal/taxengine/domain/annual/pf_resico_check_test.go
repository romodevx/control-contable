package annual_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

func resicoCheckYear(income, withheld money.Cents) annual.PFYear {
	return annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{
		{Regime: fiscal.RegimeRESICO, Income: income, ISRWithheld: withheld},
	}}
}

func resicoCheckCompute(t *testing.T, in annual.PFYear, pay annual.PaymentsReader) annual.Section {
	t.Helper()
	got, err := annual.NewPFRESICOCheck().Compute(in, pay, art152Params(t))
	require.NoError(t, err)
	return got
}

// TestPFRESICOCheckReproducesAN03 is research AN-03 (a) and (b): the 113-F
// annual rate on the whole income, minus the monthly ISR, is an informative
// difference, never a balance (RMF 2026 rule 3.13.7).
func TestPFRESICOCheckReproducesAN03(t *testing.T) {
	tests := []struct {
		name     string
		income   string
		paid     string
		wantDiff string
		wantRow  string
	}{
		{"a: 480,000 at 1.10 %", "480000.00", "6600.00", "-1320.00", "0.011"},
		{"b: 300,001 at 1.10 %", "300001.00", "3025.01", "275.00", "0.011"},
		{"300,000 is the 1.00 % row", "300000.00", "3000.00", "0.00", "0.01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pay := art152Payments{fiscal.RegimeRESICO: {art152Filed(pesos(t, tt.paid))}}

			s := resicoCheckCompute(t, resicoCheckYear(pesos(t, tt.income), 0), pay)

			assert.Equal(t, annual.SectionPFRESICOCheck, s.Kind)
			assert.Equal(t, []fiscal.RegimeCode{fiscal.RegimeRESICO}, s.Regimes)
			assert.Equal(t, annual.StatusInformational, s.Status)
			assert.Equal(t, annual.ObligationRelieved, s.Obligation)
			assert.Equal(t, []annual.CPATopic{annual.TopicRESICOPFReliefRenewal}, s.CPATopics)
			assert.Nil(t, s.Tax, "never a payable or refundable balance")
			require.NotNil(t, s.RESICOCheck)
			assert.Equal(t, pesos(t, tt.wantDiff), s.RESICOCheck.Difference)
			assert.Contains(t, s.Disclaimers, annual.DisclaimerRESICOPFDefinitive)
			assert.Equal(t, rate(t, tt.wantRow), art152Line(t, s.Lines, annual.LineRESICOCheckRate).Rate)
			for _, l := range s.Lines {
				if l.Operation != domain.OpInfo && l.Operation != domain.OpRoundingAdjustment {
					assert.NotEmpty(t, l.Basis, "%s cites its legal basis", l.Code)
				}
			}
		})
	}
}

func TestPFRESICOCheckSubtractsTheMoralWithholding(t *testing.T) {
	pay := art152Payments{fiscal.RegimeRESICO: {
		{Month: 1, Amount: pesos(t, "1000.00"), Source: annual.PaymentFiled},
		{Month: 2, Amount: pesos(t, "500.00"), Source: annual.PaymentEstimatedByApp},
	}}

	s := resicoCheckCompute(t, resicoCheckYear(pesos(t, "200000.00"), pesos(t, "300.00")), pay)

	require.NotNil(t, s.RESICOCheck)
	assert.Equal(t, annual.RESICOCheck{
		Income: pesos(t, "200000.00"), Rate: rate(t, "0.01"), AnnualISR: pesos(t, "2000.00"),
		Payments: pesos(t, "1500.00"), Withheld: pesos(t, "300.00"), Difference: pesos(t, "200.00"),
	}, *s.RESICOCheck, "contract ResicoPfCheck: 2,000 - 1,500 - 300 (LISR 113-J)")
	assert.Equal(t, []domain.LineCode{
		annual.LineRESICOCheckIncome, annual.LineRESICOCheckRate,
		annual.LineRESICOCheckPayments, annual.LineRESICOCheckPayments, annual.LineRESICOCheckWithheld,
		annual.LineRESICOCheckDifference, annual.LineRESICOCheckDefinitive,
	}, art152Codes(withoutRounding(s.Lines)))
	assert.Contains(t, s.Disclaimers, annual.DisclaimerPaymentsEstimated)
	assert.Equal(t, annual.StatusInformational, s.Status, "informational wins over an estimated payment")
}

func TestPFRESICOCheckWithoutIncome(t *testing.T) {
	pay := art152Payments{fiscal.RegimeRESICO: {art152Filed(pesos(t, "100.00"))}}

	s := resicoCheckCompute(t, resicoCheckYear(0, 0), pay)

	require.NotNil(t, s.RESICOCheck)
	assert.Equal(t, pesos(t, "-100.00"), s.RESICOCheck.Difference)
	assert.Equal(t, money.Cents(0), s.RESICOCheck.AnnualISR)
	assert.Equal(t, rate(t, "0.01"), s.RESICOCheck.Rate, "no income is the first 113-F row")
	assert.NotContains(t, art152Codes(s.Lines), annual.LineRESICOCheckRate)
}

// TestPFRESICOCheckAboveTheTable: above 3,500,000 the taxpayer leaves
// RESICO (113-E, 113-I); there is no 113-F rate, so no difference.
func TestPFRESICOCheckAboveTheTable(t *testing.T) {
	s := resicoCheckCompute(t, resicoCheckYear(pesos(t, "3500000.01"), 0), art152Payments{})

	assert.Nil(t, s.RESICOCheck)
	assert.Equal(t, annual.StatusInformational, s.Status)
	assert.Equal(t, domain.OpInfo, art152Line(t, s.Lines, annual.LineRESICOCheckAboveTable).Operation)
}

func TestPFRESICOCheckApplies(t *testing.T) {
	c := annual.NewPFRESICOCheck()

	assert.True(t, c.Applies(resicoCheckYear(0, 0)))
	assert.False(t, c.Applies(annual.PFYear{FiscalYear: 2026, Regimes: []annual.RegimeYear{{Regime: fiscal.RegimeBusiness}}}))
}

func TestPFRESICOCheckErrors(t *testing.T) {
	boom := errors.New("payments unavailable")
	t.Run("port", func(t *testing.T) {
		_, err := annual.NewPFRESICOCheck().Compute(resicoCheckYear(1, 0), art152FailingPayments{err: boom}, art152Params(t))
		require.ErrorIs(t, err, boom)
	})
	t.Run("negative income", func(t *testing.T) {
		_, err := annual.NewPFRESICOCheck().Compute(resicoCheckYear(-1, 0), art152Payments{}, art152Params(t))
		require.ErrorIs(t, err, annual.ErrInvalidYear)
	})
	t.Run("not held", func(t *testing.T) {
		_, err := annual.NewPFRESICOCheck().Compute(annual.PFYear{FiscalYear: 2026}, art152Payments{}, art152Params(t))
		require.ErrorIs(t, err, annual.ErrInvalidYear)
	})
	t.Run("no annual 113-F table", func(t *testing.T) {
		_, err := annual.NewPFRESICOCheck().Compute(resicoCheckYear(1, 0), art152Payments{}, rentOnlyParams(t))
		require.Error(t, err)
	})
}
