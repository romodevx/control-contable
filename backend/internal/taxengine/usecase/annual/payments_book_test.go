package annual_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

func months(ms ...time.Month) []time.Month { return ms }

func priorPayments(t *testing.T, values map[time.Month]money.Cents) domain.TaxInputs {
	t.Helper()
	var vs []domain.TaxInputValue
	for m, v := range values {
		vs = append(vs, domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m, Money: v})
	}
	in, err := domain.NewTaxInputs(vs...)
	require.NoError(t, err)
	return in
}

// The precedence of MODULE-annual §2.5 per month: filed, then the tax
// input, then the app's estimate (max 0); every month labelled.
func TestResolveRegimePayments_precedencePerMonth(t *testing.T) {
	src := annual.RegimeSources{
		Regime: "606",
		Active: months(time.January, time.February, time.March, time.April),
		Filed:  map[time.Month]money.Cents{time.January: 100},
		Inputs: priorPayments(t, map[time.Month]money.Cents{time.January: 999, time.February: 200}),
		App: map[time.Month]annual.MonthlyISR{
			time.January: {ISR: 1}, time.February: {ISR: 2}, time.March: {ISR: 300}, time.April: {ISR: -5},
		},
	}

	got, err := annual.ResolveRegimePayments(2026, src)

	require.NoError(t, err)
	assert.Equal(t, []adomain.Payment{
		{Month: 1, Amount: 100, Source: adomain.PaymentFiled},
		{Month: 2, Amount: 200, Source: adomain.PaymentTaxInput},
		{Month: 3, Amount: 300, Source: adomain.PaymentEstimatedByApp},
		{Month: 4, Amount: 0, Source: adomain.PaymentEstimatedByApp},
	}, got.Resolved.Payments)
	assert.Empty(t, got.Resolved.Missing)
	require.Len(t, got.Months, 12)
	jan := got.Months[0]
	assert.Equal(t, ptr(money.Cents(100)), jan.Filed)
	assert.Equal(t, ptr(money.Cents(999)), jan.TaxInput)
	assert.Equal(t, ptr(money.Cents(100)), jan.Effective)
	assert.Equal(t, adomain.PaymentFiled, jan.Source)
	may := got.Months[4]
	assert.Nil(t, may.Effective)
	assert.Equal(t, adomain.PaymentSource(""), may.Source, "a month outside the active months has no payment")
}

// A cumulative regime's monthly estimate credits only the typed prior
// payments: chained, each default is the month's ISR plus what the
// estimate credited minus what the annual credits before it, so the
// defaults never pay the same cumulative tax twice.
func TestResolveRegimePayments_chainsACumulativeRegime(t *testing.T) {
	// Cumulative tax through each month: 100, 250, 450; nothing typed, so
	// each monthly estimate is the cumulative tax.
	src := annual.RegimeSources{
		Regime: "612", Cumulative: true,
		Active: months(time.January, time.February, time.March),
		Inputs: priorPayments(t, nil),
		App: map[time.Month]annual.MonthlyISR{
			time.January: {ISR: 100}, time.February: {ISR: 250}, time.March: {ISR: 450},
		},
	}

	got, err := annual.ResolveRegimePayments(2026, src)

	require.NoError(t, err)
	assert.Equal(t, []money.Cents{100, 150, 200}, amounts(got.Resolved.Payments))
	assert.Equal(t, ptr(money.Cents(150)), got.Months[1].App.Amount())
}

// A filed month in the middle replaces the default it would have been: the
// next default credits the filed value instead.
func TestResolveRegimePayments_chainCreditsTheFiledMonth(t *testing.T) {
	// February's estimate credits January's typed 100 (a tax input), so it
	// is cumulative 250 − 100 = 150; January is filed at 120 instead.
	src := annual.RegimeSources{
		Regime: "612", Cumulative: true,
		Active: months(time.January, time.February),
		Filed:  map[time.Month]money.Cents{time.January: 120},
		Inputs: priorPayments(t, map[time.Month]money.Cents{time.January: 100}),
		App:    map[time.Month]annual.MonthlyISR{time.January: {ISR: 100}, time.February: {ISR: 150}},
	}

	got, err := annual.ResolveRegimePayments(2026, src)

	require.NoError(t, err)
	assert.Equal(t, []money.Cents{120, 130}, amounts(got.Resolved.Payments))
}

// A month whose estimate was incomplete for any reason other than the
// chained prior payments is not verifiable: it is still credited, and named
// as missing.
func TestResolveRegimePayments_incompleteEstimateIsMissing(t *testing.T) {
	src := annual.RegimeSources{
		Regime: "601", Cumulative: true,
		Active: months(time.January, time.February),
		Inputs: priorPayments(t, nil),
		App:    map[time.Month]annual.MonthlyISR{time.January: {ISR: 10}, time.February: {ISR: 30, Incomplete: true}},
	}

	got, err := annual.ResolveRegimePayments(2026, src)

	require.NoError(t, err)
	assert.Equal(t, []string{"filed_payment:601:2026-02"}, got.Resolved.Missing)
	assert.True(t, got.Months[1].App.Incomplete)
	assert.Equal(t, []money.Cents{10, 20}, amounts(got.Resolved.Payments))
}

// A month with no source at all (no estimate either) is missing and not
// credited.
func TestResolveRegimePayments_monthWithoutSource(t *testing.T) {
	src := annual.RegimeSources{Regime: "606", Active: months(time.March), Inputs: priorPayments(t, nil)}

	got, err := annual.ResolveRegimePayments(2026, src)

	require.NoError(t, err)
	assert.Empty(t, got.Resolved.Payments)
	assert.Equal(t, []string{"filed_payment:606:2026-03"}, got.Resolved.Missing)
}

func TestPaymentBook_readsByRegime(t *testing.T) {
	book := annual.PaymentBook{"612": {Resolved: adomain.ResolvedPayments{Payments: []adomain.Payment{{Month: 1, Amount: 5, Source: adomain.PaymentFiled}}}}}

	got, err := book.ProvisionalPayments("612")
	require.NoError(t, err)
	assert.Len(t, got, 1)
	none, err := book.ProvisionalPayments("606")
	require.NoError(t, err)
	assert.Empty(t, none)
	assert.Equal(t, annual.SourceCounts{Filed: 1}, book.Counts("612", "606"))
}

func amounts(ps []adomain.Payment) []money.Cents {
	out := make([]money.Cents, len(ps))
	for i, p := range ps {
		out[i] = p.Amount
	}
	return out
}
