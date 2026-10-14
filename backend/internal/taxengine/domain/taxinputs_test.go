package domain_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func monthly(m time.Month, c money.Cents) domain.TaxInputValue {
	return domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m, Money: c}
}

func TestTaxInputsLookups(t *testing.T) {
	in, err := domain.NewTaxInputs(
		domain.TaxInputValue{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: 120_000_00},
		domain.TaxInputValue{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: 150_000},
		domain.TaxInputValue{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: true},
		monthly(time.January, 2_000_00),
		monthly(time.March, 1_000_00),
	)
	require.NoError(t, err)

	ptu, ok := in.Money(domain.InputPTUPaid)
	assert.True(t, ok)
	assert.Equal(t, money.Cents(120_000_00), ptu)
	_, ok = in.Money(domain.InputPriorYearLosses)
	assert.False(t, ok, "absent key")
	_, ok = in.Money(domain.InputCoeficienteUtilidad)
	assert.False(t, ok, "a rate is not money")

	cu, ok := in.Rate(domain.InputCoeficienteUtilidad)
	assert.True(t, ok)
	assert.Equal(t, money.Rate(150_000), cu)
	_, ok = in.Rate(domain.InputPTUPaid)
	assert.False(t, ok)

	first, ok := in.Bool(domain.InputFirstFiscalYear)
	assert.True(t, ok)
	assert.True(t, first)
	_, ok = in.Bool(domain.InputPTUPaid)
	assert.False(t, ok)

	jan, ok := in.MonthlyMoney(domain.InputPriorProvisionalPayments, time.January)
	assert.True(t, ok)
	assert.Equal(t, money.Cents(2_000_00), jan)
	_, ok = in.MonthlyMoney(domain.InputPriorProvisionalPayments, time.February)
	assert.False(t, ok, "missing month")
	_, ok = in.MonthlyMoney(domain.InputPTUPaid, 0)
	assert.False(t, ok, "a plain amount is not monthly")
}

func TestTaxInputsSumBefore(t *testing.T) {
	in, err := domain.NewTaxInputs(monthly(time.January, 2_000_00), monthly(time.March, 1_000_00), monthly(time.April, 7))
	require.NoError(t, err)
	tests := []struct {
		month time.Month
		want  money.Cents
	}{
		{time.January, 0},
		{time.February, 2_000_00},
		{time.March, 2_000_00}, // February missing: summed as zero
		{time.April, 3_000_00},
		{time.December, 3_000_07},
	}
	for _, tt := range tests {
		got, err := in.SumBefore(domain.InputPriorProvisionalPayments, tt.month)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got, tt.month.String())
	}
	_, err = in.SumBefore(domain.InputPriorProvisionalPayments, 0)
	require.ErrorIs(t, err, domain.ErrInvalidMonth)
	_, err = in.SumBefore(domain.InputPriorProvisionalPayments, 13)
	require.ErrorIs(t, err, domain.ErrInvalidMonth)

	huge, err := domain.NewTaxInputs(monthly(time.January, math.MaxInt64), monthly(time.February, 1))
	require.NoError(t, err)
	_, err = huge.SumBefore(domain.InputPriorProvisionalPayments, time.March)
	require.ErrorIs(t, err, money.ErrOverflow)
}

func TestNewTaxInputsRejects(t *testing.T) {
	tests := []struct {
		name    string
		values  []domain.TaxInputValue
		wantErr error
	}{
		{name: "unknown kind", values: []domain.TaxInputValue{{Key: domain.InputPTUPaid, Kind: "text"}}, wantErr: domain.ErrInvalidTaxInput},
		{name: "blank key", values: []domain.TaxInputValue{{Kind: domain.InputMoney}}, wantErr: domain.ErrInvalidTaxInput},
		{name: "negative money", values: []domain.TaxInputValue{{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: -1}}, wantErr: domain.ErrInvalidTaxInput},
		{name: "negative monthly money", values: []domain.TaxInputValue{monthly(time.May, -1)}, wantErr: domain.ErrInvalidTaxInput},
		{name: "rate below zero", values: []domain.TaxInputValue{{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: -1}}, wantErr: domain.ErrInvalidTaxInput},
		{name: "rate above one", values: []domain.TaxInputValue{{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: money.RateScale + 1}}, wantErr: domain.ErrInvalidTaxInput},
		{name: "monthly without month", values: []domain.TaxInputValue{monthly(0, 1)}, wantErr: domain.ErrInvalidMonth},
		{name: "monthly month 13", values: []domain.TaxInputValue{monthly(13, 1)}, wantErr: domain.ErrInvalidMonth},
		{name: "month on a plain key", values: []domain.TaxInputValue{{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Month: time.May}}, wantErr: domain.ErrInvalidMonth},
		{name: "repeated key", values: []domain.TaxInputValue{
			{Key: domain.InputPTUPaid, Kind: domain.InputMoney}, {Key: domain.InputPTUPaid, Kind: domain.InputMoney},
		}, wantErr: domain.ErrDuplicateTaxInput},
		{name: "repeated month", values: []domain.TaxInputValue{monthly(time.May, 1), monthly(time.May, 2)}, wantErr: domain.ErrDuplicateTaxInput},
		{name: "one key, two kinds", values: []domain.TaxInputValue{
			monthly(time.May, 1), {Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoney},
		}, wantErr: domain.ErrDuplicateTaxInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewTaxInputs(tt.values...)
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestZeroTaxInputsHoldNothing(t *testing.T) {
	var in domain.TaxInputs
	_, ok := in.Money(domain.InputPTUPaid)
	assert.False(t, ok)
	sum, err := in.SumBefore(domain.InputPriorProvisionalPayments, time.December)
	require.NoError(t, err)
	assert.Zero(t, sum)
}

func TestInputKindValid(t *testing.T) {
	for _, k := range []domain.InputKind{domain.InputMoney, domain.InputMoneyMonthly, domain.InputRate, domain.InputBoolean} {
		assert.True(t, k.Valid(), k)
	}
	assert.False(t, domain.InputKind("").Valid())
}
