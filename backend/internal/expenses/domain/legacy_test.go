package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// MODULE-withholdings §5.2: a persona moral's expense stored before WH-07
// has no supplier facts; a typed withholding means a persona física supplier.
func TestLegacySupplier(t *testing.T) {
	tests := []struct {
		name     string
		withheld money.Cents
		want     domain.Supplier
	}{
		{"nothing withheld", 0, domain.Supplier{Kind: domain.SupplierPM}},
		{"something withheld", 1, domain.Supplier{Kind: domain.SupplierPFWithActivity, Regime: domain.SupplierRegime612}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.LegacySupplier(tt.withheld))
		})
	}
}

func TestCreditableSplit(t *testing.T) {
	tests := []struct {
		name               string
		creditable         bool
		iva, withheld      money.Cents
		paid, withheldPart money.Cents
	}{
		{"creditable", true, 640000, 426667, 213333, 426667},
		{"nothing withheld", true, 16000, 0, 16000, 0},
		{"not creditable", false, 640000, 426667, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paid, withheld := domain.CreditableSplit(tt.creditable, tt.iva, tt.withheld)
			assert.Equal(t, tt.paid, paid)
			assert.Equal(t, tt.withheldPart, withheld)
		})
	}
}
