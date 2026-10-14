package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
)

func TestErrorClassAndDetail(t *testing.T) {
	tests := []struct {
		err    domain.Error
		class  domain.Class
		detail string
	}{
		{domain.ErrRequired, domain.ClassValidation, "required"},
		{domain.ErrInvalid, domain.ClassValidation, "invalid"},
		{domain.ErrMustBePositive, domain.ClassValidation, "must_be_positive"},
		{domain.ErrOutOfRange, domain.ClassValidation, "out_of_range"},
		{domain.ErrNotForRegime, domain.ClassValidation, "not_for_regime"},
		{domain.ErrRateNotValidOnDate, domain.ClassBusinessRule, "rate_not_valid_on_date"},
		{domain.ErrIVAAmountMismatch, domain.ClassBusinessRule, "iva_amount_mismatch"},
		{domain.ErrWithholdingNotApplicable, domain.ClassBusinessRule, "withholding_not_applicable"},
		{domain.ErrWithholdingExceedsBase, domain.ClassBusinessRule, "withholding_exceeds_base"},
		{domain.ErrActivityNotHeld, domain.ClassBusinessRule, "activity_not_held"},
		{domain.ErrInvalidParameter, domain.ClassInternal, "invalid_parameter"},
	}
	for _, tt := range tests {
		t.Run(tt.detail, func(t *testing.T) {
			assert.Equal(t, tt.class, tt.err.Class())
			assert.Equal(t, tt.detail, tt.err.Detail())
			assert.Equal(t, "incomes: "+tt.detail, tt.err.Error())
		})
	}
}

func TestViolations(t *testing.T) {
	v := domain.Violations{
		{Field: domain.FieldIVACents, Err: domain.ErrIVAAmountMismatch},
		{Field: domain.FieldISRWithheldCents, Err: domain.ErrWithholdingNotApplicable},
	}
	wrapped := fmt.Errorf("create income: %w", v)

	assert.ErrorIs(t, wrapped, domain.ErrIVAAmountMismatch)
	assert.ErrorIs(t, wrapped, domain.ErrWithholdingNotApplicable)
	assert.NotErrorIs(t, wrapped, domain.ErrRequired)
	assert.Equal(t, domain.ClassBusinessRule, v.Class())
	assert.Equal(t,
		"incomes: iva_cents: iva_amount_mismatch; isr_withheld_cents: withholding_not_applicable",
		v.Error())

	got, ok := errors.AsType[domain.Violations](wrapped)
	require.True(t, ok)
	assert.Len(t, got, 2)

	var violation *domain.Violation
	require.ErrorAs(t, wrapped, &violation)
	assert.Equal(t, domain.FieldIVACents, violation.Field)
	assert.Equal(t, "incomes: iva_cents: iva_amount_mismatch", violation.Error())
}

func TestViolationsEmptyClass(t *testing.T) {
	assert.Equal(t, domain.ClassInternal, domain.Violations{}.Class())
}
