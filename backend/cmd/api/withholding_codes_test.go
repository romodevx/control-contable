package main

import (
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// The withholdings decision table lives once, in the kernel (ADR-0008), and the
// contract publishes its vocabulary. These tests fail when the two drift, so a
// value the kernel derives can never be one the contract (and the generated
// clients) reject.

// contractEnum reads the enum of a components.schemas entry from the embedded
// contract.
func contractEnum(t *testing.T, schema string) []string {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	ref, ok := spec.Components.Schemas[schema]
	require.True(t, ok, "schema %s missing from the contract", schema)
	values := make([]string, 0, len(ref.Value.Enum))
	for _, v := range ref.Value.Enum {
		s, isString := v.(string)
		require.True(t, isString, "%s enum value %v is not a string", schema, v)
		values = append(values, s)
	}
	return values
}

func TestCounterpartyKindMatchesTheKernelReceptorKinds(t *testing.T) {
	kernel := []string{
		string(fiscal.ReceptorPM), string(fiscal.ReceptorGovFederal), string(fiscal.ReceptorGovStateMunicipal),
		string(fiscal.ReceptorPFWithActivity), string(fiscal.ReceptorPFNoActivityOrPublic), string(fiscal.ReceptorForeign),
	}
	assert.ElementsMatch(t, kernel, contractEnum(t, "CounterpartyKind"))
	for _, v := range kernel {
		assert.True(t, fiscal.ReceptorKind(v).Valid(), v)
	}
}

// The wire has no `platform` operation: a platform sale is `via_platform: true`.
func TestOperationKindMatchesTheKernelOperationsButPlatform(t *testing.T) {
	kernel := []string{
		string(fiscal.OperationProfessionalServices), string(fiscal.OperationBusinessServices),
		string(fiscal.OperationRentRealEstate), string(fiscal.OperationRentMovable),
		string(fiscal.OperationSaleOfGoods), string(fiscal.OperationGroundFreight),
		string(fiscal.OperationCommission), string(fiscal.OperationScrap), string(fiscal.OperationOther),
	}
	contract := contractEnum(t, "OperationKind")
	assert.ElementsMatch(t, kernel, contract)
	assert.NotContains(t, contract, string(fiscal.OperationPlatform))
	for _, v := range contract {
		assert.True(t, fiscal.OperationKind(v).Valid(), v)
	}
}

func TestWithholdingBasisMatchesTheKernel(t *testing.T) {
	kernel := []string{
		string(fiscal.IVABasisNone), string(fiscal.IVABasisTwoThirds), string(fiscal.IVABasisHalfOfIVA),
		string(fiscal.IVABasisAllOfIVA), string(fiscal.IVABasisFourPercentOfPrice),
	}
	assert.ElementsMatch(t, kernel, contractEnum(t, "WithholdingBasis"))
}

// Every rule id the kernel can emit, including the IVA preamble (no IVA to
// withhold: 0 %, exempt, residential rent), must match WithholdingRuleId.
func TestEveryKernelRuleIDMatchesTheContractPattern(t *testing.T) {
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	pattern := regexp.MustCompile(spec.Components.Schemas["WithholdingRuleId"].Value.Pattern)
	ids := []fiscal.RuleID{
		fiscal.RuleISR0, fiscal.RuleISR1, fiscal.RuleISR2, fiscal.RuleISR3,
		fiscal.RuleISR4, fiscal.RuleISR5, fiscal.RuleISR6, fiscal.RuleISR7,
		fiscal.RuleIVANoTax, fiscal.RuleIVA0, fiscal.RuleIVA1, fiscal.RuleIVA2, fiscal.RuleIVA3,
		fiscal.RuleIVA4, fiscal.RuleIVA5, fiscal.RuleIVA6, fiscal.RuleIVA7, fiscal.RuleIVA8,
	}
	for _, id := range ids {
		assert.Regexp(t, pattern, string(id))
	}
}

func TestKernelWithholdingWarningsAreInTheContract(t *testing.T) {
	contract := contractEnum(t, "WarningCode")
	for _, code := range []fiscal.WarningCode{fiscal.WarningCPAUnverified, fiscal.WarningForeignCounterparty} {
		assert.True(t, slices.Contains(contract, string(code)), "add %q to WarningCode", code)
	}
}
