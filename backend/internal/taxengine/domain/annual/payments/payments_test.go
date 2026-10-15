package payments_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/payments"
)

const (
	fiscalYear = 2026
	regime601  = fiscal.RegimeCode("601")
)

// sources is the Sources port as three month maps.
type sources struct {
	filed, taxInput map[time.Month]money.Cents
	estimated       map[time.Month]payments.AppEstimate
}

func (s sources) Filed(m time.Month) (money.Cents, bool) {
	v, ok := s.filed[m]
	return v, ok
}

func (s sources) TaxInput(m time.Month) (money.Cents, bool) {
	v, ok := s.taxInput[m]
	return v, ok
}

func (s sources) AppEstimate(m time.Month) (payments.AppEstimate, bool) {
	v, ok := s.estimated[m]
	return v, ok
}

func cents(t testing.TB, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(s)
	require.NoError(t, err)
	return c
}

// TestResolveAppliesThePrecedencePerMonth: filed > tax input > the app's
// estimate, per month (MODULE-annual §2.5); a negative app estimate is a
// zero payment.
func TestResolveAppliesThePrecedencePerMonth(t *testing.T) {
	src := sources{
		filed:    map[time.Month]money.Cents{time.January: cents(t, "100.00"), time.March: 0},
		taxInput: map[time.Month]money.Cents{time.January: cents(t, "999.00"), time.February: cents(t, "200.00"), time.March: cents(t, "999.00")},
		estimated: map[time.Month]payments.AppEstimate{
			time.January: {ISR: cents(t, "888.00")}, time.February: {ISR: cents(t, "888.00")},
			time.March: {ISR: cents(t, "888.00")}, time.April: {ISR: cents(t, "300.00")},
			time.May: {ISR: cents(t, "-50.00")},
		},
	}

	got, err := payments.Resolve(regime601, fiscalYear, []time.Month{1, 2, 3, 4, 5}, src)

	require.NoError(t, err)
	assert.Equal(t, []payments.Payment{
		{Month: time.January, Amount: cents(t, "100.00"), Source: payments.SourceFiled},
		{Month: time.February, Amount: cents(t, "200.00"), Source: payments.SourceTaxInput},
		{Month: time.March, Amount: 0, Source: payments.SourceFiled},
		{Month: time.April, Amount: cents(t, "300.00"), Source: payments.SourceEstimatedByApp},
		{Month: time.May, Amount: 0, Source: payments.SourceEstimatedByApp},
	}, got.Payments)
	assert.Empty(t, got.Missing)
	assert.Equal(t, 2, got.EstimatedMonths())
}

// TestResolveNamesTheMonthsItCannotVerify: a default payment taken from an
// incomplete monthly estimate, or a month with no source at all, is a
// missing filed_payment:<regime>:<YYYY-MM> (MODULE-annual §2.6).
func TestResolveNamesTheMonthsItCannotVerify(t *testing.T) {
	src := sources{estimated: map[time.Month]payments.AppEstimate{
		time.June: {ISR: cents(t, "10.00"), Incomplete: true},
	}}

	got, err := payments.Resolve(regime601, fiscalYear, []time.Month{time.June, time.July}, src)

	require.NoError(t, err)
	assert.Equal(t, []payments.Payment{
		{Month: time.June, Amount: cents(t, "10.00"), Source: payments.SourceEstimatedByApp},
	}, got.Payments)
	assert.Equal(t, []string{"filed_payment:601:2026-06", "filed_payment:601:2026-07"}, got.Missing)
}

func TestResolveRefusesBadMonthsAndNegativeAmounts(t *testing.T) {
	tests := []struct {
		name   string
		months []time.Month
		src    sources
	}{
		{name: "month zero", months: []time.Month{0}},
		{name: "month thirteen", months: []time.Month{13}},
		{name: "repeated month", months: []time.Month{time.May, time.May}},
		{name: "unordered months", months: []time.Month{time.May, time.April}},
		{name: "negative filed", months: []time.Month{time.May}, src: sources{filed: map[time.Month]money.Cents{time.May: -1}}},
		{name: "negative tax input", months: []time.Month{time.May}, src: sources{taxInput: map[time.Month]money.Cents{time.May: -1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := payments.Resolve(regime601, fiscalYear, tt.months, tt.src)

			require.ErrorIs(t, err, payments.ErrInvalidPayments)
		})
	}
}

// TestTotalsSumEachSource: the credit lines group the payments by source.
func TestTotalsSumEachSource(t *testing.T) {
	r := payments.Resolved{Payments: []payments.Payment{
		{Month: time.January, Amount: cents(t, "1.10"), Source: payments.SourceFiled},
		{Month: time.February, Amount: cents(t, "2.20"), Source: payments.SourceEstimatedByApp},
		{Month: time.March, Amount: cents(t, "3.30"), Source: payments.SourceFiled},
		{Amount: cents(t, "4.40"), Source: payments.SourceTaxInput}, // a year total
	}}

	got, err := r.Totals()

	require.NoError(t, err)
	assert.Equal(t, []payments.Total{
		{Source: payments.SourceFiled, Amount: cents(t, "4.40"), Months: 2},
		{Source: payments.SourceTaxInput, Amount: cents(t, "4.40"), Months: 1},
		{Source: payments.SourceEstimatedByApp, Amount: cents(t, "2.20"), Months: 1},
	}, got)
}

func TestTotalsReportOverflow(t *testing.T) {
	r := payments.Resolved{Payments: []payments.Payment{
		{Month: time.January, Amount: money.Cents(1 << 62), Source: payments.SourceFiled},
		{Month: time.February, Amount: money.Cents(1 << 62), Source: payments.SourceFiled},
	}}

	_, err := r.Totals()

	require.ErrorIs(t, err, money.ErrOverflow)
}

func TestSourceValid(t *testing.T) {
	for _, s := range []payments.Source{payments.SourceFiled, payments.SourceTaxInput, payments.SourceEstimatedByApp} {
		assert.True(t, s.Valid(), s)
	}
	assert.False(t, payments.Source("guessed").Valid())
}
