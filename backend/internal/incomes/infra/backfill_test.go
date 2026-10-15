package infra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The backfill writes a subset of what a write may: the facts and the
// derivation, never an amount, a date or a description.
func TestDerivationColumns_areWritableFactsAndDerivation(t *testing.T) {
	assert.Subset(t, writableColumns, derivationColumns)
	for _, never := range []string{"amount_cents", "iva_cents", "iva_rate", "invoiced_on", "collected_on", "description", "cfdi_uuid"} {
		assert.NotContains(t, derivationColumns, never)
	}
}
