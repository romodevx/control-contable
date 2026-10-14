package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// TestModuleErrorCodes: the incomes module's own details[].code values and
// their ErrorCode family (MODULE-incomes §4). The shared codes are the
// fiscal kernel's.
func TestModuleErrorCodes(t *testing.T) {
	tests := []struct {
		err    fiscal.Error
		class  fiscal.Class
		detail string
	}{
		{domain.ErrNotForRegime, fiscal.ClassValidation, "not_for_regime"},
		{domain.ErrActivityNotHeld, fiscal.ClassBusinessRule, "activity_not_held"},
	}
	for _, tt := range tests {
		t.Run(tt.detail, func(t *testing.T) {
			assert.Equal(t, tt.class, tt.err.Class())
			assert.Equal(t, tt.detail, tt.err.Detail())
		})
	}
}
