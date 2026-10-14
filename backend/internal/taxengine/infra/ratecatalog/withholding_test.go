package ratecatalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expenses "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
)

// withholdingRates is the consumer-side shape fiscal.DeriveWithholdings
// (WH-04) is expected to read: every accessor speaks kernel types only, so a
// kernel interface can name it without importing the tax engine.
type withholdingRates interface {
	ISRWithholdingFees(on fiscal.Date) (money.Rate, error)
	ISRWithholdingRent(on fiscal.Date) (money.Rate, error)
	ISRWithholdingRESICO(on fiscal.Date) (money.Rate, error)
	IVAWithholdingFreight(on fiscal.Date) (money.Rate, error)
	IVAWithholdingScrap(on fiscal.Date) (money.Rate, error)
}

var _ withholdingRates = (*catalog)(nil)

type withholdingAccessor struct {
	name string
	key  fiscal.RateKey
	get  func(*catalog, fiscal.Date) (money.Rate, error)
	want string // the 2026 value, docs/reviews/WH-05-params-verification.md
}

func withholdingAccessors() []withholdingAccessor {
	return []withholdingAccessor{
		{"ISR fees (LISR 106)", fiscal.KeyISRWithholdingFees, (*catalog).ISRWithholdingFees, "0.10"},
		{"ISR rent (LISR 116)", fiscal.KeyISRWithholdingRent, (*catalog).ISRWithholdingRent, "0.10"},
		{"ISR RESICO PF (LISR 113-J)", ratecatalog.KeyISRWithholdingRESICO, (*catalog).ISRWithholdingRESICO, "0.0125"},
		{"IVA freight (RLIVA 3 II)", ratecatalog.KeyIVAWithholdingFreight, (*catalog).IVAWithholdingFreight, "0.04"},
		{"IVA scrap (LIVA 1-A II b)", ratecatalog.KeyIVAWithholdingScrap, (*catalog).IVAWithholdingScrap, "1"},
	}
}

// The adapter's keys are the ones the modules already read: one spelling.
func TestWithholdingKeysMatchTheModules(t *testing.T) {
	assert.Equal(t, incomes.KeyISRWithholdingRESICO, ratecatalog.KeyISRWithholdingRESICO)
	assert.Equal(t, expenses.KeyIVAWithholdingFreight, ratecatalog.KeyIVAWithholdingFreight)
}

func TestWithholdingRatesOnTheYearBoundaries(t *testing.T) {
	c := embedded(t)
	for _, a := range withholdingAccessors() {
		t.Run(a.name, func(t *testing.T) {
			for _, on := range []string{"2026-01-01", "2026-12-31"} {
				got, err := a.get(c, date(t, on))
				require.NoError(t, err, on)
				assert.Equal(t, rate(t, a.want), got, on)
			}
			// The boolean port reads the same value.
			got, ok := c.Rate(a.key, date(t, "2026-06-30"))
			require.True(t, ok)
			assert.Equal(t, rate(t, a.want), got)
		})
	}
}

// No set covers 2025 (no 2024 to fall back on): an error, never a rate.
func TestWithholdingRateBeforeAnySet(t *testing.T) {
	c := embedded(t)
	for _, a := range withholdingAccessors() {
		t.Run(a.name, func(t *testing.T) {
			_, err := a.get(c, date(t, "2025-12-31"))
			require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)
		})
	}
}

// 2027 is not published: the 2026 rate is served, and Provisional tells the
// record to carry the disclaimer.
func TestWithholdingRateInAnUnpublishedYear(t *testing.T) {
	c := embedded(t)
	on := date(t, "2027-01-01")
	for _, a := range withholdingAccessors() {
		got, err := a.get(c, on)
		require.NoError(t, err, a.name)
		assert.Equal(t, rate(t, a.want), got, a.name)
	}
	provisional, err := c.Provisional(on)
	require.NoError(t, err)
	assert.True(t, provisional)
}

// A rate that starts mid-year is an error the day before it starts.
func TestWithholdingRateNotYetInForce(t *testing.T) {
	for _, a := range withholdingAccessors() {
		t.Run(a.name, func(t *testing.T) {
			c := synthetic(t, syntheticSetStartingOn(t, 2026, a.key, date(t, "2026-07-01")))

			_, err := a.get(c, date(t, "2026-06-30"))
			require.ErrorIs(t, err, taxparams.ErrNotInForce)

			got, err := a.get(c, date(t, "2026-07-01"))
			require.NoError(t, err)
			assert.Equal(t, rate(t, syntheticWithholdingRates[a.key]), got)
		})
	}
}

// A set without a withholding rate fails at boot: otherwise every income or
// expense needing it would fail at derivation time, one by one.
func TestNewRequiresTheWithholdingRatesInEverySet(t *testing.T) {
	for _, a := range withholdingAccessors() {
		t.Run(a.name, func(t *testing.T) {
			_, err := newSynthetic(t, syntheticSetWithout(t, 2026, false, taxparams.Key(a.key)))
			require.ErrorIs(t, err, ratecatalog.ErrMissingWithholdingRate)
			require.ErrorIs(t, err, taxparams.ErrUnknownKey)
		})
	}
	_, err := newSynthetic(t, syntheticSet(t, 2026, false),
		syntheticSetWithout(t, 2027, false, taxparams.Key(ratecatalog.KeyIVAWithholdingScrap)))
	require.ErrorIs(t, err, ratecatalog.ErrMissingWithholdingRate, "missing from a later year only")
}
