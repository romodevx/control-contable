package ratecatalog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	expenses "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

// One adapter value satisfies both consumer-side ports (ADR-0007 §6):
// incomes reads fiscal.RateCatalog as is, expenses embeds it and adds Amount.
var (
	_ fiscal.RateCatalog   = (*catalog)(nil)
	_ expenses.RateCatalog = (*catalog)(nil)
)

func TestNewRejectsNilCatalog(t *testing.T) {
	_, err := ratecatalog.New[expenses.AmountKey](nil)
	require.ErrorIs(t, err, ratecatalog.ErrNoCatalog)
}

func TestAllowedIVARates(t *testing.T) {
	all := []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}
	noBorder := []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate0, fiscal.IVAExempt}
	tests := []struct {
		name            string
		on              string
		want            []fiscal.IVARate
		wantProvisional bool
	}{
		{name: "first day of 2026", on: "2026-01-01", want: all},
		{name: "mid 2026", on: "2026-07-15", want: all},
		{name: "border decree's last day, inclusive", on: "2026-12-31", want: all},
		// 2027 is not published: the 2026 set is the provisional fallback,
		// and the border 8 % stays lapsed (no date is hard-coded).
		{name: "border decree lapsed", on: "2027-01-01", want: noBorder, wantProvisional: true},
	}
	c := embedded(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.AllowedIVARates(date(t, tt.on))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Value)
			assert.Equal(t, tt.wantProvisional, got.Provisional)
			assert.Equal(t, taxparams.SetID{FiscalYear: 2026, Version: 1}, got.Set)
		})
	}
}

// The allowed set is exactly what the kernel's CheckIVA accepts through the
// same adapter: one rule, never a second copy.
func TestAllowedIVARatesAgreeWithCheckIVA(t *testing.T) {
	c := embedded(t)
	const amount = 1_000_00
	for _, on := range []string{"2026-12-31", "2027-01-01"} {
		day := date(t, on)
		allowed, err := c.AllowedIVARates(day)
		require.NoError(t, err)
		for _, treatment := range []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt} {
			_, err := fiscal.ExpectedIVA(amount, treatment, day, c)
			assert.Equal(t, err == nil, contains(allowed.Value, treatment), "%s on %s", treatment, on)
		}
	}
}

func contains(rates []fiscal.IVARate, r fiscal.IVARate) bool {
	for _, x := range rates {
		if x == r {
			return true
		}
	}
	return false
}

func TestCheckIVABorderRateBoundary(t *testing.T) {
	c := embedded(t)
	require.NoError(t, fiscal.CheckIVA(10_000_00, fiscal.IVARate8, 800_00, date(t, "2026-12-31"), c))
	err := fiscal.CheckIVA(10_000_00, fiscal.IVARate8, 800_00, date(t, "2027-01-01"), c)
	require.ErrorIs(t, err, fiscal.ErrRateNotValidOnDate)
	require.NoError(t, fiscal.CheckIVA(10_000_00, fiscal.IVARate16, 1_600_00, date(t, "2027-01-01"), c))
	require.NoError(t, fiscal.CheckIVA(10_000_00, fiscal.IVAExempt, 0, date(t, "2027-01-01"), c))
}

// Every rate key a consumer reads resolves on 2026 dates to the verified
// value: IVA treatments, ISR withholdings (PM 10 %, RESICO 1.25 %, the 2026
// platform rates) and the 4 % freight IVA withholding.
func TestRateByKind(t *testing.T) {
	tests := []struct {
		key  fiscal.RateKey
		want string
	}{
		{fiscal.KeyIVAGeneral, "0.16"},
		{fiscal.KeyIVABorder, "0.08"},
		{fiscal.KeyIVAZero, "0"},
		{fiscal.KeyISRWithholdingFees, "0.10"},
		{fiscal.KeyISRWithholdingRent, "0.10"},
		{incomes.KeyISRWithholdingRESICO, "0.0125"},
		{incomes.KeyISRPlatformTransport, "0.021"},
		{incomes.KeyISRPlatformLodging, "0.04"},
		{incomes.KeyISRPlatformGoods, "0.025"},
		{incomes.KeyISRPlatformNoRFC, "0.20"},
		{incomes.KeyISRPlatformPersonaMoral, "0.025"},
		{expenses.KeyIVAWithholdingFreight, "0.04"},
	}
	c := embedded(t)
	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			for _, on := range []string{"2026-01-01", "2026-12-31"} {
				got, ok := c.Rate(tt.key, date(t, on))
				require.True(t, ok, on)
				assert.Equal(t, rate(t, tt.want), got, on)

				resolved, err := c.RateOn(tt.key, date(t, on))
				require.NoError(t, err)
				assert.Equal(t, got, resolved.Value)
				assert.False(t, resolved.Provisional)
			}
		})
	}
}

// The LIF 2026 platform rates sunset with the year: 2027 needs its own set.
func TestPlatformRatesLapseWithTheYear(t *testing.T) {
	c := embedded(t)
	for _, key := range []fiscal.RateKey{incomes.KeyISRPlatformGoods, incomes.KeyISRPlatformPersonaMoral} {
		_, ok := c.Rate(key, date(t, "2027-01-01"))
		assert.False(t, ok, key)
		_, err := c.RateOn(key, date(t, "2027-01-01"))
		require.ErrorIs(t, err, taxparams.ErrNotInForce, key)
	}
}

// The suggestions the income and expense domains compute through the
// adapter: ISR at the catalog rate, IVA as the legal share (2/3).
func TestWithholdingSuggestionsThroughTheAdapter(t *testing.T) {
	c := embedded(t)
	on := date(t, "2026-06-30")
	const amount, iva = 10_000_00, 1_600_00

	fees := incomes.PolicyFor(incomes.Payee{PersonType: fiscal.PersonFisica, Regime: fiscal.RegimeBusiness}, incomes.CounterpartyMoral, "")
	got, err := fees.Suggest(amount, iva, on, c)
	require.NoError(t, err)
	assert.Equal(t, fiscal.Suggested{
		ISR: fiscal.Suggestion{Cents: 1_000_00, Known: true},
		IVA: fiscal.Suggestion{Cents: 1_066_67, Known: true}, // 2/3 of 1,600.00
	}, got)

	resico := incomes.PolicyFor(incomes.Payee{PersonType: fiscal.PersonFisica, Regime: fiscal.RegimeRESICO}, incomes.CounterpartyMoral, "")
	got, err = resico.Suggest(amount, iva, on, c)
	require.NoError(t, err)
	assert.Equal(t, fiscal.Suggestion{Cents: 125_00, Known: true}, got.ISR)

	// Expenses derive what a persona moral withholds (WH-07): a 606
	// landlord's rent, through the adapter's kernel view.
	rent, err := fiscal.DeriveWithholdings(fiscal.WithholdingInput{
		EmisorPersona: fiscal.PersonFisica, EmisorRegime: fiscal.RegimeRent, Receptor: fiscal.ReceptorPM,
		Operation: fiscal.OperationRentRealEstate, Subtotal: amount, IVA: iva, IVARate: fiscal.IVARate16, On: on,
	}, c.Withholdings())
	require.NoError(t, err)
	assert.Equal(t, fiscal.Withholdings{ISR: 1_000_00, IVA: 1_066_67}, rent.Withholdings)
}

func TestCashPaymentLimit(t *testing.T) {
	c := embedded(t)
	on := date(t, "2026-03-10")

	got, err := c.CashPaymentLimit(on)
	require.NoError(t, err)
	assert.Equal(t, cents(t, "2000.00"), got.Value)

	amount, ok := c.Amount(expenses.KeyCashPaymentLimit, on)
	require.True(t, ok)
	assert.Equal(t, got.Value, amount)
}

// January is still in force with the previous UMA: the INEGI value
// published in January takes effect on February 1 (DOF, UMA notices).
func TestUMAByDate(t *testing.T) {
	tests := []struct {
		name            string
		period          ratecatalog.UMAPeriod
		on              string
		want            string
		wantProvisional bool
	}{
		{name: "daily, first day of 2026", period: ratecatalog.UMADaily, on: "2026-01-01", want: "113.14"},
		{name: "daily, last day of January", period: ratecatalog.UMADaily, on: "2026-01-31", want: "113.14"},
		{name: "daily, February 1", period: ratecatalog.UMADaily, on: "2026-02-01", want: "117.31"},
		{name: "monthly, last day of January", period: ratecatalog.UMAMonthly, on: "2026-01-31", want: "3439.46"},
		{name: "monthly, February 1", period: ratecatalog.UMAMonthly, on: "2026-02-01", want: "3566.22"},
		{name: "annual, last day of January", period: ratecatalog.UMAAnnual, on: "2026-01-31", want: "41273.52"},
		{name: "annual, February 1", period: ratecatalog.UMAAnnual, on: "2026-02-01", want: "42794.64"},
		{name: "daily, December 31", period: ratecatalog.UMADaily, on: "2026-12-31", want: "117.31"},
		// January 2027 keeps the 2026 UMA, from the provisional fallback.
		{name: "daily, January 2027", period: ratecatalog.UMADaily, on: "2027-01-31", want: "117.31", wantProvisional: true},
	}
	c := embedded(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.UMA(tt.period, date(t, tt.on))
			require.NoError(t, err)
			assert.Equal(t, cents(t, tt.want), got.Value)
			assert.Equal(t, tt.wantProvisional, got.Provisional)
		})
	}
}

func TestUMAAfterItsValidity(t *testing.T) {
	_, err := embedded(t).UMA(ratecatalog.UMADaily, date(t, "2027-02-01"))
	require.ErrorIs(t, err, taxparams.ErrNotInForce)
}

// A date no loaded fiscal year covers, and that is not the year right after
// one (the provisional fallback), is a typed error, never a silent value.
func TestDateNotCovered(t *testing.T) {
	c := embedded(t)
	for _, on := range []string{"2025-12-31", "2022-01-01", "2028-01-01"} {
		t.Run(on, func(t *testing.T) {
			day := date(t, on)
			_, err := c.RateOn(fiscal.KeyIVAGeneral, day)
			require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)
			require.ErrorIs(t, err, taxparams.ErrNoParameters)

			_, err = c.AllowedIVARates(day)
			require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)
			_, err = c.CashPaymentLimit(day)
			require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)
			_, err = c.UMA(ratecatalog.UMADaily, day)
			require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)

			_, ok := c.Rate(fiscal.KeyIVAGeneral, day)
			assert.False(t, ok)
			_, ok = c.Amount(expenses.KeyCashPaymentLimit, day)
			assert.False(t, ok)
		})
	}
}

func TestZeroDateNotCovered(t *testing.T) {
	_, err := embedded(t).RateOn(fiscal.KeyIVAGeneral, fiscal.Date{})
	require.ErrorIs(t, err, ratecatalog.ErrDateNotCovered)
}

func TestProvisionalFallbackForANewYear(t *testing.T) {
	c := synthetic(t, syntheticSet(t, 2026, false))

	got, err := c.RateOn(fiscal.KeyIVAGeneral, date(t, "2027-03-01"))
	require.NoError(t, err)
	assert.Equal(t, rate(t, "0.16"), got.Value)
	assert.True(t, got.Provisional, "the provisional_parameters disclaimer must surface")
	assert.Equal(t, taxparams.SetID{FiscalYear: 2026, Version: 1}, got.Set)

	limit, err := c.CashPaymentLimit(date(t, "2027-03-01"))
	require.NoError(t, err)
	assert.True(t, limit.Provisional)
}

func TestPublishedYearIsPreferredOverFallback(t *testing.T) {
	c := synthetic(t, syntheticSet(t, 2026, false), syntheticSet(t, 2027, false))

	got, err := c.AllowedIVARates(date(t, "2027-01-01"))
	require.NoError(t, err)
	assert.False(t, got.Provisional)
	assert.Equal(t, taxparams.SetID{FiscalYear: 2027, Version: 1}, got.Set)
	assert.Contains(t, got.Value, fiscal.IVARate8, "the 2027 set extends the border decree")
}

func TestSetMarkedProvisional(t *testing.T) {
	c := synthetic(t, syntheticSet(t, 2026, true))
	got, err := c.RateOn(fiscal.KeyIVAGeneral, date(t, "2026-05-01"))
	require.NoError(t, err)
	assert.True(t, got.Provisional)
}

func TestKeyAbsentFromTheSet(t *testing.T) {
	c := synthetic(t, syntheticSetWithout(t, 2026, false, taxparams.Key(fiscal.KeyIVABorder)))
	on := date(t, "2026-05-01")

	_, err := c.RateOn(fiscal.KeyIVABorder, on)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
	_, ok := c.Rate(fiscal.KeyIVABorder, on)
	assert.False(t, ok)

	// The border 8 % is the one optional IVA rate: a set without it does not
	// allow it.
	allowed, err := c.AllowedIVARates(on)
	require.NoError(t, err)
	assert.Equal(t, []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate0, fiscal.IVAExempt}, allowed.Value)

	_, err = c.UMA(ratecatalog.UMAMonthly, on)
	require.ErrorIs(t, err, taxparams.ErrUnknownKey)
}

// PR #77 follow-up: a set without the general or the zero IVA rate (left out
// or misspelled) fails at boot. Otherwise the boolean port would report "not
// in force" and every 16 % record would be refused with
// rate_not_valid_on_date, silently.
func TestNewRequiresTheIVARatesInEverySet(t *testing.T) {
	tests := []struct {
		name string
		sets []*taxparams.Params
		want error
	}{
		{
			name: "general 16 % missing",
			sets: []*taxparams.Params{syntheticSetWithout(t, 2026, false, taxparams.Key(fiscal.KeyIVAGeneral))},
			want: taxparams.ErrUnknownKey,
		},
		{
			name: "zero rate missing",
			sets: []*taxparams.Params{syntheticSetWithout(t, 2026, false, taxparams.Key(fiscal.KeyIVAZero))},
			want: taxparams.ErrUnknownKey,
		},
		{
			name: "missing from a later year only",
			sets: []*taxparams.Params{
				syntheticSet(t, 2026, false),
				syntheticSetWithout(t, 2027, false, taxparams.Key(fiscal.KeyIVAGeneral)),
			},
			want: taxparams.ErrUnknownKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newSynthetic(t, tt.sets...)
			require.ErrorIs(t, err, ratecatalog.ErrMissingIVARate)
			require.ErrorIs(t, err, tt.want)
		})
	}
}

// A superseded version is never served, so only the highest version of a
// year must hold the rates: v2 fixes v1's gap.
func TestNewChecksOnlyTheServedVersion(t *testing.T) {
	fixed := synthetic(t,
		syntheticVersion(t, 2026, 1, false, taxparams.Key(fiscal.KeyIVAGeneral)),
		syntheticVersion(t, 2026, 2, false))
	got, err := fixed.RateOn(fiscal.KeyIVAGeneral, date(t, "2026-05-01"))
	require.NoError(t, err)
	assert.Equal(t, taxparams.SetID{FiscalYear: 2026, Version: 2}, got.Set)

	_, err = newSynthetic(t, syntheticVersion(t, 2026, 1, false), syntheticVersion(t, 2026, 2, false, taxparams.Key(fiscal.KeyIVAZero)))
	require.ErrorIs(t, err, ratecatalog.ErrMissingIVARate, "a broken latest version is refused")
}

// The embedded 2026 set holds every required IVA rate.
func TestNewAcceptsTheEmbeddedSets(t *testing.T) {
	sets, err := params.Embedded()
	require.NoError(t, err)
	_, err = ratecatalog.New[expenses.AmountKey](sets)
	require.NoError(t, err)
}

// Provisional reports whether a record dated on gets its values from a
// provisional set (PR #77 follow-up: the flag must reach the records); a
// date no set covers is ErrDateNotCovered, never a silent value.
func TestProvisional(t *testing.T) {
	tests := []struct {
		name    string
		sets    []*taxparams.Params
		on      string
		want    bool
		wantErr error
	}{
		{name: "published year", sets: []*taxparams.Params{syntheticSet(t, 2026, false)}, on: "2026-05-01"},
		{name: "set marked provisional", sets: []*taxparams.Params{syntheticSet(t, 2026, true)}, on: "2026-05-01", want: true},
		{name: "fallback to the previous year", sets: []*taxparams.Params{syntheticSet(t, 2026, false)}, on: "2027-01-01", want: true},
		{name: "before the first set", sets: []*taxparams.Params{syntheticSet(t, 2026, false)}, on: "2025-12-31", wantErr: ratecatalog.ErrDateNotCovered},
		{name: "zero date", sets: []*taxparams.Params{syntheticSet(t, 2026, false)}, wantErr: ratecatalog.ErrDateNotCovered},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var on fiscal.Date
			if tt.on != "" {
				on = date(t, tt.on)
			}
			got, err := synthetic(t, tt.sets...).Provisional(on)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// A rate read as money (or money as a rate) is a wiring bug: an error, and
// no value through the boolean ports.
func TestKindMismatch(t *testing.T) {
	c := synthetic(t, syntheticSet(t, 2026, false))
	on := date(t, "2026-05-01")

	_, err := c.RateOn(fiscal.RateKey(expenses.KeyCashPaymentLimit), on)
	require.ErrorIs(t, err, taxparams.ErrKindMismatch)
	_, ok := c.Amount(expenses.AmountKey(fiscal.KeyIVAGeneral), on)
	assert.False(t, ok)
}
