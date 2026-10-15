package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// TestModuleErrorCodes: the expenses module's own details[].code values and
// their ErrorCode family (MODULE-expenses §4). The shared codes are the
// fiscal kernel's.
func TestModuleErrorCodes(t *testing.T) {
	tests := []struct {
		err    fiscal.Error
		class  fiscal.Class
		detail string
	}{
		{domain.ErrUnknownCategory, fiscal.ClassValidation, "unknown_category"},
		{domain.ErrHasCFDIMismatch, fiscal.ClassValidation, "has_cfdi_mismatch"},
		{domain.ErrSupplierNotApplicable, fiscal.ClassBusinessRule, "supplier_not_applicable"},
		{domain.ErrResidentialRentExempt, fiscal.ClassBusinessRule, "residential_rent_exempt"},
	}
	for _, tt := range tests {
		t.Run(tt.detail, func(t *testing.T) {
			assert.Equal(t, tt.class, tt.err.Class())
			assert.Equal(t, tt.detail, tt.err.Detail())
		})
	}
}
