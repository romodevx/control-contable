package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// The legend is the contract's frozen const (PRD §6), and the engine
// version is what the breakdown header shows ("Motor 1.0.0", US-X2).
func TestEngineConstants(t *testing.T) {
	assert.Equal(t, "Esto es un aproximado. No todos los gastos aplican para tu giro comercial.", domain.Legend)
	assert.Equal(t, "1.0.0", domain.EngineVersion)
}
