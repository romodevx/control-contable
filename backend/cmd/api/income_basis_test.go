package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	incomes "github.com/romodevx/control-contable/backend/internal/incomes/domain"
)

// regimes.income_basis is spelled three times: the catalog CHECK, the incomes
// domain (which books every income from it) and the contract's IncomeBasis
// (which the taxpayers module serves verbatim from the same column, ADR-0010).
// The domain parses the column, so the CHECK cannot drift from it unnoticed;
// this test closes the other side: a basis the domain books can never be one
// the contract (and the generated clients) reject, nor the reverse.
func TestIncomeBasisMatchesTheIncomesDomain(t *testing.T) {
	domain := []string{string(incomes.BasisCash), string(incomes.BasisAccrual)}

	contract := contractEnum(t, "IncomeBasis")

	assert.ElementsMatch(t, domain, contract)
	for _, v := range contract {
		assert.True(t, incomes.IncomeBasis(v).Valid(), v)
	}
}
