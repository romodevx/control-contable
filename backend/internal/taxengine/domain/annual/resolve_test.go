package annual_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

const resolveRegime = fiscal.RegimeCode("601")

// paymentSources is the PaymentSources port as three month maps.
type paymentSources struct {
	filed, taxInput map[time.Month]money.Cents
	estimated       map[time.Month]annual.AppEstimate
}

func (s paymentSources) Filed(m time.Month) (money.Cents, bool) {
	v, ok := s.filed[m]
	return v, ok
}

func (s paymentSources) TaxInput(m time.Month) (money.Cents, bool) {
	v, ok := s.taxInput[m]
	return v, ok
}

func (s paymentSources) AppEstimate(m time.Month) (annual.AppEstimate, bool) {
	v, ok := s.estimated[m]
	return v, ok
}

// TestResolvePaymentsAppliesThePrecedencePerMonth: filed > tax input > the
// app's estimate, per month (MODULE-annual §2.5); a negative app estimate is
// a zero payment.
func TestResolvePaymentsAppliesThePrecedencePerMonth(t *testing.T) {
	src := paymentSources{
		filed:    map[time.Month]money.Cents{time.January: pesos(t, "100.00"), time.March: 0},
		taxInput: map[time.Month]money.Cents{time.January: pesos(t, "999.00"), time.February: pesos(t, "200.00"), time.March: pesos(t, "999.00")},
		estimated: map[time.Month]annual.AppEstimate{
			time.January: {ISR: pesos(t, "888.00")}, time.February: {ISR: pesos(t, "888.00")},
			time.March: {ISR: pesos(t, "888.00")}, time.April: {ISR: pesos(t, "300.00")},
			time.May: {ISR: pesos(t, "-50.00")},
		},
	}

	got, err := annual.ResolvePayments(resolveRegime, fiscalYear, []time.Month{1, 2, 3, 4, 5}, src)

	require.NoError(t, err)
	assert.Equal(t, []annual.Payment{
		{Month: 1, Amount: pesos(t, "100.00"), Source: annual.PaymentFiled},
		{Month: 2, Amount: pesos(t, "200.00"), Source: annual.PaymentTaxInput},
		{Month: 3, Amount: 0, Source: annual.PaymentFiled},
		{Month: 4, Amount: pesos(t, "300.00"), Source: annual.PaymentEstimatedByApp},
		{Month: 5, Amount: 0, Source: annual.PaymentEstimatedByApp},
	}, got.Payments)
	assert.Empty(t, got.Missing)
	assert.Equal(t, 2, got.EstimatedMonths())
}

// TestResolvePaymentsNamesTheMonthsItCannotVerify: a default payment taken
// from an incomplete monthly estimate (marked FromIncompleteEstimate, as the
// PF sections read it), or a month with no source at all, is a missing
// filed_payment:<regime>:<YYYY-MM> (MODULE-annual §2.6).
func TestResolvePaymentsNamesTheMonthsItCannotVerify(t *testing.T) {
	src := paymentSources{estimated: map[time.Month]annual.AppEstimate{
		time.June: {ISR: pesos(t, "10.00"), Incomplete: true},
	}}

	got, err := annual.ResolvePayments(resolveRegime, fiscalYear, []time.Month{time.June, time.July}, src)

	require.NoError(t, err)
	assert.Equal(t, []annual.Payment{
		{Month: 6, Amount: pesos(t, "10.00"), Source: annual.PaymentEstimatedByApp, FromIncompleteEstimate: true},
	}, got.Payments)
	assert.Equal(t, []string{"filed_payment:601:2026-06", "filed_payment:601:2026-07"}, got.Missing)
}

func TestResolvePaymentsRefusesBadMonthsAndNegativeAmounts(t *testing.T) {
	tests := []struct {
		name   string
		months []time.Month
		src    paymentSources
	}{
		{name: "month zero", months: []time.Month{0}},
		{name: "month thirteen", months: []time.Month{13}},
		{name: "repeated month", months: []time.Month{time.May, time.May}},
		{name: "unordered months", months: []time.Month{time.May, time.April}},
		{name: "negative filed", months: []time.Month{time.May}, src: paymentSources{filed: map[time.Month]money.Cents{time.May: -1}}},
		{name: "negative tax input", months: []time.Month{time.May}, src: paymentSources{taxInput: map[time.Month]money.Cents{time.May: -1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := annual.ResolvePayments(resolveRegime, fiscalYear, tt.months, tt.src)

			require.ErrorIs(t, err, annual.ErrInvalidYear)
		})
	}
}

// TestPaymentTotalsSumEachSource: the credit lines group the payments by
// source, in precedence order.
func TestPaymentTotalsSumEachSource(t *testing.T) {
	r := annual.ResolvedPayments{Payments: []annual.Payment{
		{Month: 1, Amount: pesos(t, "1.10"), Source: annual.PaymentFiled},
		{Month: 2, Amount: pesos(t, "2.20"), Source: annual.PaymentEstimatedByApp},
		{Month: 3, Amount: pesos(t, "3.30"), Source: annual.PaymentFiled},
		{Amount: pesos(t, "4.40"), Source: annual.PaymentTaxInput}, // a year total
	}}

	got, err := r.Totals()

	require.NoError(t, err)
	assert.Equal(t, []annual.PaymentTotal{
		{Source: annual.PaymentFiled, Amount: pesos(t, "4.40"), Months: 2},
		{Source: annual.PaymentTaxInput, Amount: pesos(t, "4.40"), Months: 1},
		{Source: annual.PaymentEstimatedByApp, Amount: pesos(t, "2.20"), Months: 1},
	}, got)
}

func TestPaymentTotalsReportOverflow(t *testing.T) {
	r := annual.ResolvedPayments{Payments: []annual.Payment{
		{Month: 1, Amount: money.Cents(1 << 62), Source: annual.PaymentFiled},
		{Month: 2, Amount: money.Cents(1 << 62), Source: annual.PaymentFiled},
	}}

	_, err := r.Totals()

	require.ErrorIs(t, err, money.ErrOverflow)
}

func TestPaymentSourceValid(t *testing.T) {
	for _, s := range []annual.PaymentSource{annual.PaymentFiled, annual.PaymentTaxInput, annual.PaymentEstimatedByApp} {
		assert.True(t, s.Valid(), s)
	}
	assert.False(t, annual.PaymentSource("guessed").Valid())
}
