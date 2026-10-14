package taxinputs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

func ptr[T any](v T) *T { return &v }

func TestDecodeDefs(t *testing.T) {
	defs, err := decodeDefs(`[{"key":"coeficiente_utilidad","label":"Coeficiente de utilidad","kind":"rate","required":true},
		{"key":"first_fiscal_year","label":"Primer ejercicio fiscal","kind":"boolean","required":false}]`)

	require.NoError(t, err)
	assert.Equal(t, []domain.InputDef{
		{Key: domain.InputCoeficienteUtilidad, Label: "Coeficiente de utilidad", Kind: domain.InputRate, Required: true},
		{Key: domain.InputFirstFiscalYear, Label: "Primer ejercicio fiscal", Kind: domain.InputBoolean},
	}, defs)

	_, err = decodeDefs(`[{"key":"x","label":"x","kind":"percent","required":true}]`)
	assert.ErrorIs(t, err, domain.ErrInvalidTaxInput, "an unknown kind is a broken catalog")
	_, err = decodeDefs(`{}`)
	assert.Error(t, err)
}

func TestRegimeRowToHeld(t *testing.T) {
	held, err := regimeRow{StartPeriod: "2023-07", EndPeriod: ptr("2025-02"), RequiredInputs: "[]"}.toHeld()
	require.NoError(t, err)
	assert.Equal(t, "2023-07", held.Start.String())
	assert.Equal(t, "2025-02", held.End.String())
	assert.Empty(t, held.Definitions)

	for _, bad := range []regimeRow{
		{StartPeriod: "2023-13", RequiredInputs: "[]"},
		{StartPeriod: "2023-01", EndPeriod: ptr("x"), RequiredInputs: "[]"},
		{StartPeriod: "2023-01", RequiredInputs: "null-ish"},
	} {
		_, err := bad.toHeld()
		assert.Error(t, err)
	}
}

func TestToNewRow_setsOnlyTheKindsColumns(t *testing.T) {
	key := usecase.InputsKey{TaxpayerID: "t", Regime: "601", FiscalYear: 2026}
	base := newValueRow{UserID: "u", TaxpayerID: "t", RegimeCode: "601", FiscalYear: 2026}
	for _, tc := range []struct {
		in   usecase.StoredInput
		want func(newValueRow) newValueRow
	}{
		{usecase.StoredInput{Key: "ptu_paid", Kind: domain.InputMoney, Cents: 5},
			func(r newValueRow) newValueRow { r.InputKey, r.ValueCents = "ptu_paid", ptr[int64](5); return r }},
		{usecase.StoredInput{Key: "prior_provisional_payments", Kind: domain.InputMoneyMonthly, Month: time.March, Cents: 7},
			func(r newValueRow) newValueRow {
				r.InputKey, r.Month, r.ValueCents = "prior_provisional_payments", ptr[int16](3), ptr[int64](7)
				return r
			}},
		{usecase.StoredInput{Key: "coeficiente_utilidad", Kind: domain.InputRate, Rate: "0.1500"},
			func(r newValueRow) newValueRow {
				r.InputKey, r.ValueRate = "coeficiente_utilidad", ptr("0.1500")
				return r
			}},
		{usecase.StoredInput{Key: "first_fiscal_year", Kind: domain.InputBoolean, Bool: true},
			func(r newValueRow) newValueRow { r.InputKey, r.ValueBool = "first_fiscal_year", ptr(true); return r }},
	} {
		assert.Equal(t, tc.want(base), toNewRow("u", key, tc.in), tc.in.Kind)
	}
}

func TestValueRowToDomain(t *testing.T) {
	v, err := valueRow{InputKey: "coeficiente_utilidad", Kind: "rate", ValueRate: ptr("0.1500")}.toDomain()
	require.NoError(t, err)
	assert.Equal(t, domain.TaxInputValue{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: money.Rate(150_000)}, v)

	v, err = valueRow{InputKey: "prior_provisional_payments", Kind: "money_monthly", Month: ptr[int16](2), ValueCents: ptr[int64](9)}.toDomain()
	require.NoError(t, err)
	assert.Equal(t, domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.February, Money: 9}, v)

	_, err = valueRow{InputKey: "coeficiente_utilidad", Kind: "rate", ValueRate: ptr("x")}.toDomain()
	assert.Error(t, err)
}
