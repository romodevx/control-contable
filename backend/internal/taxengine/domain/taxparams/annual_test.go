package taxparams_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// tuitionLevel stands in for the consumer's level type (annual.TuitionLevel,
// PR #127): any string type instantiates DeductionLimits.
type tuitionLevel string

// consumerLimits copies annual.Limits (PR #127, domain/annual/limits.go)
// field for field. The conversion in the test below compiles only while
// DeductionLimits keeps the same fields, types and order, which is what
// lets the composition root convert one into the other.
type consumerLimits struct {
	AnnualUMA               money.Cents
	GlobalCapUMAs           int64
	GlobalCapIncomeShare    money.Rate
	FuneralCapUMAs          int64
	DonationShare           money.Rate
	GovernmentDonationShare money.Rate
	RetirementIncomeShare   money.Rate
	RetirementCapUMAs       int64
	TuitionCaps             map[tuitionLevel]money.Cents
}

// annualSpec is the synthetic fixture plus every deduction-limit key.
func annualSpec(t testing.TB) taxparams.Spec {
	t.Helper()
	spec := validSpec(t)
	count := func(key taxparams.Key, n int64) taxparams.Value {
		return taxparams.Value{Key: key, Kind: taxparams.KindCount, Count: n, Validity: from2026(t), Basis: basis()}
	}
	rate := func(key taxparams.Key, r string) taxparams.Value {
		return taxparams.Value{Key: key, Kind: taxparams.KindRate, Rate: mustRate(t, r), Validity: from2026(t), Basis: basis()}
	}
	amount := func(key taxparams.Key, m string) taxparams.Value {
		return taxparams.Value{Key: key, Kind: taxparams.KindMoney, Money: mustCents(t, m), Validity: from2026(t), Basis: basis()}
	}
	spec.Values = append(spec.Values,
		taxparams.Value{Key: taxparams.KeyAnnualUMA, Kind: taxparams.KindMoney, Money: mustCents(t, "1200.00"), Basis: basis(),
			Validity: taxparams.Validity{From: mustDate(t, "2026-02-01"), To: mustDate(t, "2027-01-31")}},
		count(taxparams.KeyGlobalCapUMAs, 5),
		rate(taxparams.KeyGlobalCapIncomeShare, "0.15"),
		count(taxparams.KeyFuneralCapUMAs, 1),
		rate(taxparams.KeyDonationShare, "0.07"),
		rate(taxparams.KeyGovernmentDonationShare, "0.04"),
		rate(taxparams.KeyRetirementIncomeShare, "0.10"),
		count(taxparams.KeyRetirementCapUMAs, 5),
		amount("deduction.tuition_cap.preescolar", "10.00"),
		amount("deduction.tuition_cap.primaria", "20.00"),
		amount("deduction.tuition_cap.secundaria", "30.00"),
		amount("deduction.tuition_cap.profesional_tecnico", "40.00"),
		amount("deduction.tuition_cap.bachillerato", "50.00"),
	)
	return spec
}

func annualCatalog(t testing.TB, spec taxparams.Spec) *taxparams.Catalog {
	t.Helper()
	c, err := taxparams.NewCatalog(mustParams(t, spec))
	require.NoError(t, err)
	return c
}

func TestReadDeductionLimits(t *testing.T) {
	got, err := taxparams.ReadDeductionLimits[tuitionLevel](annualCatalog(t, annualSpec(t)), 2026)
	require.NoError(t, err)
	assert.Equal(t, consumerLimits{
		AnnualUMA:     mustCents(t, "1200.00"),
		GlobalCapUMAs: 5, GlobalCapIncomeShare: mustRate(t, "0.15"),
		FuneralCapUMAs: 1,
		DonationShare:  mustRate(t, "0.07"), GovernmentDonationShare: mustRate(t, "0.04"),
		RetirementIncomeShare: mustRate(t, "0.10"), RetirementCapUMAs: 5,
		TuitionCaps: map[tuitionLevel]money.Cents{
			"preescolar": mustCents(t, "10.00"), "primaria": mustCents(t, "20.00"), "secundaria": mustCents(t, "30.00"),
			"profesional_tecnico": mustCents(t, "40.00"), "bachillerato": mustCents(t, "50.00"),
		},
	}, consumerLimits(got))
}

func TestReadDeductionLimitsUsesTheUMAInForceAtYearEnd(t *testing.T) {
	spec := annualSpec(t)
	spec.Values = append(spec.Values, taxparams.Value{
		Key: taxparams.KeyAnnualUMA, Kind: taxparams.KindMoney, Money: mustCents(t, "1100.00"), Basis: basis(),
		Validity: taxparams.Validity{From: mustDate(t, "2026-01-01"), To: mustDate(t, "2026-01-31")},
	})
	got, err := taxparams.ReadDeductionLimits[tuitionLevel](annualCatalog(t, spec), 2026)
	require.NoError(t, err)
	assert.Equal(t, mustCents(t, "1200.00"), got.AnnualUMA, "31 December, not January's carried-over UMA")
}

func TestReadDeductionLimitsFails(t *testing.T) {
	without := func(key taxparams.Key) func(testing.TB) taxparams.Spec {
		return func(t testing.TB) taxparams.Spec {
			spec := annualSpec(t)
			spec.Values = append(spec.Values[:valueIndex(t, spec, key)], spec.Values[valueIndex(t, spec, key)+1:]...)
			return spec
		}
	}
	tests := []struct {
		name string
		spec func(testing.TB) taxparams.Spec
		year int
		want error
	}{
		{"no set within reach", annualSpec, 2030, taxparams.ErrNoParameters},
		{"fallback past year 9999 has no 31 December", func(t testing.TB) taxparams.Spec {
			spec := annualSpec(t)
			spec.FiscalYear = 9999
			return spec
		}, 10000, fiscal.ErrInvalid},
		{"fallback year whose UMA has lapsed", annualSpec, 2027, taxparams.ErrNotInForce},
		{"missing UMA multiple", without(taxparams.KeyGlobalCapUMAs), 2026, taxparams.ErrUnknownKey},
		{"missing share", without(taxparams.KeyDonationShare), 2026, taxparams.ErrUnknownKey},
		{"missing tuition cap", without("deduction.tuition_cap.bachillerato"), 2026, taxparams.ErrUnknownKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := taxparams.ReadDeductionLimits[tuitionLevel](annualCatalog(t, tt.spec(t)), tt.year)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

func TestAnnualRequirementsHoldOnTheAnnualFixture(t *testing.T) {
	p := mustParams(t, annualSpec(t))
	require.NoError(t, p.Require(taxparams.DeductionLimitRequirements()...))
	err := mustParams(t, validSpec(t)).Require(taxparams.DeductionLimitRequirements()...)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
}
