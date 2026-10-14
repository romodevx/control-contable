package domain_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

var fingerprintFormat = regexp.MustCompile(`^[0-9a-f]{64}$`)

// fpRegime is a 612 regime for March with one month of income and the
// options and inputs given.
func fpRegime(t *testing.T, cash money.Cents, options map[domain.OptionKey]domain.OptionValue, values ...domain.TaxInputValue) domain.FingerprintRegime {
	t.Helper()
	ys, err := domain.NewYearSums(time.March, domain.MonthEntry{
		Month: time.February, Sums: domain.MonthSums{Income: domain.IncomeSums{Cash: cash}},
	})
	require.NoError(t, err)
	return domain.FingerprintRegime{
		Input: domain.Input{
			Key:    domain.StrategyKey{Regime: fiscal.RegimeBusiness, PersonType: fiscal.PersonFisica},
			Period: period(t, "2026-03"), Options: domain.NewOptions(options), Sums: ys, Inputs: inputs(t, values...),
		},
		Treatment:    domain.TreatmentFull,
		ActiveMonths: []time.Month{time.January, time.February, time.March},
	}
}

func fpInput(t *testing.T, regimes ...domain.FingerprintRegime) domain.FingerprintInput {
	return domain.FingerprintInput{
		PersonType: fiscal.PersonFisica, Period: period(t, "2026-03"), ParameterSet: "2026/v1", Regimes: regimes,
	}
}

func prior(month time.Month, c money.Cents) domain.TaxInputValue {
	return domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: month, Money: c}
}

// ADR-0002 §5: the fingerprint is SHA-256 hex over the canonical
// aggregated input, the parameter version and the engine version; equal
// inputs give equal fingerprints, whatever the order of the regimes, the
// options or the tax inputs.
func TestFingerprintIsDeterministic(t *testing.T) {
	fullyTaxable := map[domain.OptionKey]domain.OptionValue{
		domain.OptionIVAFullyTaxable: domain.BoolOption(true), "activities": domain.ListOption([]string{"a", "b"}),
	}
	a := fpRegime(t, 100_00, fullyTaxable, prior(time.January, 1_00), prior(time.February, 2_00))
	b := fpRegime(t, 100_00, fullyTaxable, prior(time.February, 2_00), prior(time.January, 1_00))
	rent := fpRegime(t, 5_00, map[domain.OptionKey]domain.OptionValue{"deduction_mode": domain.TextOption("blind_35")})
	rent.Input.Key.Regime = fiscal.RegimeRent

	first := domain.Fingerprint(fpInput(t, a, rent))
	again := domain.Fingerprint(fpInput(t, rent, b))

	assert.Regexp(t, fingerprintFormat, first)
	assert.Equal(t, first, again)
}

// Every part of the input moves the fingerprint (M8 marks a month
// "changed since viewed" by it).
func TestFingerprintChangesWithEveryInput(t *testing.T) {
	base := func() domain.FingerprintInput {
		return fpInput(t, fpRegime(t, 100_00, map[domain.OptionKey]domain.OptionValue{
			domain.OptionIVAFullyTaxable: domain.BoolOption(true),
		}, prior(time.January, 1_00)))
	}
	want := domain.Fingerprint(base())
	tests := []struct {
		name   string
		mutate func(*domain.FingerprintInput)
	}{
		{"a record sum", func(in *domain.FingerprintInput) {
			in.Regimes[0] = fpRegime(t, 100_01, map[domain.OptionKey]domain.OptionValue{
				domain.OptionIVAFullyTaxable: domain.BoolOption(true),
			}, prior(time.January, 1_00))
		}},
		{"a tax input", func(in *domain.FingerprintInput) {
			in.Regimes[0] = fpRegime(t, 100_00, map[domain.OptionKey]domain.OptionValue{
				domain.OptionIVAFullyTaxable: domain.BoolOption(true),
			}, prior(time.January, 1_01))
		}},
		{"an option", func(in *domain.FingerprintInput) {
			in.Regimes[0] = fpRegime(t, 100_00, map[domain.OptionKey]domain.OptionValue{
				domain.OptionIVAFullyTaxable: domain.BoolOption(false),
			}, prior(time.January, 1_00))
		}},
		{"the active months", func(in *domain.FingerprintInput) { in.Regimes[0].ActiveMonths = []time.Month{time.March} }},
		{"the treatment", func(in *domain.FingerprintInput) { in.Regimes[0].Treatment = domain.TreatmentInformational }},
		{"the regime", func(in *domain.FingerprintInput) { in.Regimes[0].Input.Key.Regime = fiscal.RegimeRESICO }},
		{"the person type", func(in *domain.FingerprintInput) { in.PersonType = fiscal.PersonMoral }},
		{"the period", func(in *domain.FingerprintInput) { in.Period = period(t, "2026-04") }},
		{"the parameter set", func(in *domain.FingerprintInput) { in.ParameterSet = "2026/v2" }},
		{"provisional parameters", func(in *domain.FingerprintInput) { in.Provisional = true }},
		{"one more regime", func(in *domain.FingerprintInput) {
			extra := in.Regimes[0]
			extra.Input.Key.Regime = fiscal.RegimeRent
			in.Regimes = append(in.Regimes, extra)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base()
			tt.mutate(&in)

			assert.NotEqual(t, want, domain.Fingerprint(in))
		})
	}
}
