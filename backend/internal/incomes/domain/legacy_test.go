package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// MODULE-withholdings §5.2: the deprecated counterparty_type of a record
// stored before WH-06 read forwards.
func TestLegacyFacts(t *testing.T) {
	tests := []struct {
		legacy      domain.Counterparty
		kind        fiscal.ReceptorKind
		viaPlatform bool
	}{
		{domain.CounterpartyMoral, fiscal.ReceptorPM, false},
		{domain.CounterpartyFisica, fiscal.ReceptorPFWithActivity, false},
		{domain.CounterpartyPublic, fiscal.ReceptorPFNoActivityOrPublic, false},
		{domain.CounterpartyPlatform, fiscal.ReceptorPM, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.legacy), func(t *testing.T) {
			kind, via, ok := domain.LegacyFacts(tt.legacy)
			assert.True(t, ok)
			assert.Equal(t, tt.kind, kind)
			assert.Equal(t, tt.viaPlatform, via)
			assert.Equal(t, tt.legacy, domain.LegacyCounterparty(kind, via), "the mirror reads it back")
		})
	}
	t.Run("unknown", func(t *testing.T) {
		_, _, ok := domain.LegacyFacts("")
		assert.False(t, ok)
	})
}
