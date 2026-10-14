package fiscal_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

func TestErrorClassAndDetail(t *testing.T) {
	tests := []struct {
		err    fiscal.Error
		class  fiscal.Class
		detail string
	}{
		{fiscal.ErrRequired, fiscal.ClassValidation, "required"},
		{fiscal.ErrInvalid, fiscal.ClassValidation, "invalid"},
		{fiscal.ErrMustBePositive, fiscal.ClassValidation, "must_be_positive"},
		{fiscal.ErrOutOfRange, fiscal.ClassValidation, "out_of_range"},
		{fiscal.ErrRateNotValidOnDate, fiscal.ClassBusinessRule, "rate_not_valid_on_date"},
		{fiscal.ErrIVAAmountMismatch, fiscal.ClassBusinessRule, "iva_amount_mismatch"},
		{fiscal.ErrWithholdingNotApplicable, fiscal.ClassBusinessRule, "withholding_not_applicable"},
		{fiscal.ErrWithholdingExceedsBase, fiscal.ClassBusinessRule, "withholding_exceeds_base"},
		{fiscal.ErrInvalidParameter, fiscal.ClassInternal, "invalid_parameter"},
		{fiscal.ValidationError("not_for_regime"), fiscal.ClassValidation, "not_for_regime"},
		{fiscal.BusinessRuleError("activity_not_held"), fiscal.ClassBusinessRule, "activity_not_held"},
	}
	for _, tt := range tests {
		t.Run(tt.detail, func(t *testing.T) {
			assert.Equal(t, tt.class, tt.err.Class())
			assert.Equal(t, tt.detail, tt.err.Detail())
			assert.Equal(t, "fiscal: "+tt.detail, tt.err.Error())
		})
	}
}

// TestModuleErrorsAreComparable: a module declares its own codes once, as
// package variables, and callers still match them with errors.Is.
func TestModuleErrorsAreComparable(t *testing.T) {
	errNotForRegime := fiscal.ValidationError("not_for_regime")
	wrapped := fmt.Errorf("create income: %w", errNotForRegime)
	require.ErrorIs(t, wrapped, fiscal.ValidationError("not_for_regime"))
	assert.NotErrorIs(t, wrapped, fiscal.BusinessRuleError("not_for_regime"), "the class is part of the identity")
	assert.NotErrorIs(t, wrapped, fiscal.ErrInvalid)
}

func TestViolations(t *testing.T) {
	var v fiscal.Violations
	v.Add("iva_cents", fiscal.ErrIVAAmountMismatch)
	v.Add("isr_withheld_cents", fiscal.ErrWithholdingNotApplicable)
	wrapped := fmt.Errorf("create income: %w", v)

	assert.ErrorIs(t, wrapped, fiscal.ErrIVAAmountMismatch)
	assert.ErrorIs(t, wrapped, fiscal.ErrWithholdingNotApplicable)
	assert.NotErrorIs(t, wrapped, fiscal.ErrRequired)
	assert.Equal(t, fiscal.ClassBusinessRule, v.Class())
	assert.Equal(t,
		"fiscal: iva_cents: iva_amount_mismatch; isr_withheld_cents: withholding_not_applicable",
		v.Error())

	got, ok := errors.AsType[fiscal.Violations](wrapped)
	require.True(t, ok)
	assert.Len(t, got, 2)

	var violation *fiscal.Violation
	require.ErrorAs(t, wrapped, &violation)
	assert.Equal(t, "iva_cents", violation.Field)
	assert.Equal(t, "fiscal: iva_cents: iva_amount_mismatch", violation.Error())
}

func TestViolationsEmptyClass(t *testing.T) {
	assert.Equal(t, fiscal.ClassInternal, fiscal.Violations{}.Class())
}
