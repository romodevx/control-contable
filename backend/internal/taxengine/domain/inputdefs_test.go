package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// defs601 are the 601 PM definitions of the catalog (contract example regime601).
var defs601 = []domain.InputDef{
	{Key: domain.InputCoeficienteUtilidad, Label: "Coeficiente de utilidad", Kind: domain.InputRate, Required: true},
	{Key: domain.InputFirstFiscalYear, Label: "Primer ejercicio fiscal", Kind: domain.InputBoolean},
	{Key: domain.InputPriorProvisionalPayments, Label: "Pagos provisionales anteriores del ejercicio", Kind: domain.InputMoneyMonthly, Required: true},
	{Key: domain.InputPTUPaid, Label: "PTU pagada en el ejercicio", Kind: domain.InputMoney},
}

func ptr[T any](v T) *T { return &v }

func submittedPTU(c int64) domain.SubmittedInput {
	return domain.SubmittedInput{Key: domain.InputPTUPaid, Cents: &c}
}

func submittedPayment(month int, c int64) domain.SubmittedInput {
	return domain.SubmittedInput{Key: domain.InputPriorProvisionalPayments, Month: &month, Cents: &c}
}

func submittedCoefficient(r string) domain.SubmittedInput {
	return domain.SubmittedInput{Key: domain.InputCoeficienteUtilidad, Rate: &r}
}

func TestCheckSubmittedInputs_accepts(t *testing.T) {
	got, violations := domain.CheckSubmittedInputs(defs601, []domain.SubmittedInput{
		submittedCoefficient("0.1500"),
		{Key: domain.InputFirstFiscalYear, Bool: ptr(false)},
		submittedPayment(1, 2_000_00),
		submittedPayment(2, 0),
		submittedPTU(0),
	})

	require.Empty(t, violations)
	assert.Equal(t, []domain.TaxInputValue{
		{Key: domain.InputCoeficienteUtilidad, Kind: domain.InputRate, Rate: 150_000},
		{Key: domain.InputFirstFiscalYear, Kind: domain.InputBoolean, Bool: false},
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.January, Money: 2_000_00},
		{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: time.February, Money: 0},
		{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: 0},
	}, got, "submission order, each with its catalog kind")

	got, violations = domain.CheckSubmittedInputs(defs601, nil)
	assert.Empty(t, violations)
	assert.Empty(t, got, "an empty set clears the inputs")
}

func TestCheckSubmittedInputs_rateBounds(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want money.Rate
		err  error
	}{
		{raw: "0", want: 0},
		{raw: "1", want: money.RateScale},
		{raw: "1.000000", want: money.RateScale},
		{raw: "0.000001", want: 1},
		{raw: "1.000001", err: domain.ErrInputOutOfRange},
		{raw: "1.5", err: domain.ErrInputOutOfRange},
		{raw: "999.999999", err: domain.ErrInputOutOfRange},
		{raw: "0.1234567", err: domain.ErrInvalidTaxInput},
		{raw: "-0.1", err: domain.ErrInputOutOfRange},
		{raw: "", err: domain.ErrInvalidTaxInput},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			got, violations := domain.CheckSubmittedInputs(defs601, []domain.SubmittedInput{submittedCoefficient(tc.raw)})
			if tc.err != nil {
				require.Len(t, violations, 1)
				assert.Equal(t, domain.InputViolation{Index: 0, Property: domain.PropertyValueRate, Err: tc.err}, violations[0])
				assert.Nil(t, got)
				return
			}
			require.Empty(t, violations)
			assert.Equal(t, tc.want, got[0].Rate)
		})
	}
}

func TestCheckSubmittedInputs_violations(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input []domain.SubmittedInput
		want  []domain.InputViolation
	}{
		{
			name:  "a key the regime does not define",
			input: []domain.SubmittedInput{submittedPTU(1), {Key: "ytd_before_app_income", Cents: ptr[int64](5)}},
			want:  []domain.InputViolation{{Index: 1, Property: domain.PropertyKey, Err: domain.ErrUnknownInput}},
		},
		{
			name:  "the value field of another kind",
			input: []domain.SubmittedInput{{Key: domain.InputPTUPaid, Rate: ptr("0.1")}},
			want: []domain.InputViolation{
				{Index: 0, Property: domain.PropertyValueRate, Err: domain.ErrInvalidTaxInput},
				{Index: 0, Property: domain.PropertyValueCents, Err: domain.ErrInvalidTaxInput},
			},
		},
		{
			name:  "both value fields",
			input: []domain.SubmittedInput{{Key: domain.InputFirstFiscalYear, Bool: ptr(true), Cents: ptr[int64](1)}},
			want:  []domain.InputViolation{{Index: 0, Property: domain.PropertyValueCents, Err: domain.ErrInvalidTaxInput}},
		},
		{
			name:  "no value field",
			input: []domain.SubmittedInput{{Key: domain.InputFirstFiscalYear}},
			want:  []domain.InputViolation{{Index: 0, Property: domain.PropertyValueBool, Err: domain.ErrInvalidTaxInput}},
		},
		{
			name:  "a month on an annual key",
			input: []domain.SubmittedInput{{Key: domain.InputPTUPaid, Cents: ptr[int64](1), Month: ptr(3)}},
			want:  []domain.InputViolation{{Index: 0, Property: domain.PropertyMonth, Err: domain.ErrInvalidTaxInput}},
		},
		{
			name:  "no month on a monthly key",
			input: []domain.SubmittedInput{{Key: domain.InputPriorProvisionalPayments, Cents: ptr[int64](1)}},
			want:  []domain.InputViolation{{Index: 0, Property: domain.PropertyMonth, Err: domain.ErrInvalidTaxInput}},
		},
		{
			name:  "a month outside 1-12",
			input: []domain.SubmittedInput{submittedPayment(13, 1), submittedPayment(0, 1)},
			want: []domain.InputViolation{
				{Index: 0, Property: domain.PropertyMonth, Err: domain.ErrInputOutOfRange},
				{Index: 1, Property: domain.PropertyMonth, Err: domain.ErrInputOutOfRange},
			},
		},
		{
			name:  "money outside [0, NonNegativeCents.maximum]",
			input: []domain.SubmittedInput{submittedPTU(-1), submittedPayment(1, 100_000_000_000_000)},
			want: []domain.InputViolation{
				{Index: 0, Property: domain.PropertyValueCents, Err: domain.ErrInputOutOfRange},
				{Index: 1, Property: domain.PropertyValueCents, Err: domain.ErrInputOutOfRange},
			},
		},
		{
			name:  "a repeated annual key",
			input: []domain.SubmittedInput{submittedPTU(1), submittedPTU(2)},
			want:  []domain.InputViolation{{Index: 1, Property: domain.PropertyKey, Err: domain.ErrDuplicateTaxInput}},
		},
		{
			name:  "a repeated key and month",
			input: []domain.SubmittedInput{submittedPayment(1, 1), submittedPayment(2, 1), submittedPayment(1, 3)},
			want:  []domain.InputViolation{{Index: 2, Property: domain.PropertyMonth, Err: domain.ErrDuplicateTaxInput}},
		},
		{
			name: "every violation at once, by index",
			input: []domain.SubmittedInput{
				{Key: "nope", Cents: ptr[int64](1)},
				submittedCoefficient("2"),
				submittedPayment(1, 1),
			},
			want: []domain.InputViolation{
				{Index: 0, Property: domain.PropertyKey, Err: domain.ErrUnknownInput},
				{Index: 1, Property: domain.PropertyValueRate, Err: domain.ErrInputOutOfRange},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, violations := domain.CheckSubmittedInputs(defs601, tc.input)
			assert.Equal(t, tc.want, violations)
			assert.Nil(t, got, "no values when any is refused")
		})
	}
}

func TestInputViolation_Error(t *testing.T) {
	v := domain.InputViolation{Index: 2, Property: domain.PropertyMonth, Err: domain.ErrInvalidTaxInput}

	assert.Equal(t, "values.2.month: taxengine: invalid tax input", v.Error())
	assert.ErrorIs(t, v, domain.ErrInvalidTaxInput)
	assert.Equal(t, "values.2.month", v.Field("values"))
}

func TestInputDefPosition(t *testing.T) {
	position := domain.InputDefPositions(defs601)

	assert.Equal(t, map[domain.InputKey]int{
		domain.InputCoeficienteUtilidad: 0, domain.InputFirstFiscalYear: 1,
		domain.InputPriorProvisionalPayments: 2, domain.InputPTUPaid: 3,
	}, position)
}
