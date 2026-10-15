package infra

import (
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// PR #136 note 1: the column grants do not stop a generic map that echoes
// request fields, so the write maps are built from one allowlist.
func TestColumnsOf_writesExactlyTheAllowlist(t *testing.T) {
	for name, w := range map[string]fiscal.WithholdingRecord{
		"derived":   {RuleIDs: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6}, TableVersion: "2026.1"},
		"underived": {},
	} {
		got := slices.Sorted(maps.Keys(columnsOf(usecase.Fields{}, w)))
		assert.Equal(t, slices.Sorted(slices.Values(writableColumns)), got, name)
	}
}

func TestColumnsOf_derivedWithholding(t *testing.T) {
	r, err := money.ParseRate("0.0125")
	require.NoError(t, err)
	c := columnsOf(usecase.Fields{Counterparty: fiscal.ReceptorPM, Operation: fiscal.OperationProfessionalServices},
		fiscal.WithholdingRecord{
			Withholdings: fiscal.Withholdings{ISR: 125_000, IVA: 1_066_667}, ISRRate: r,
			IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
			RuleIDs: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6}, TableVersion: "2026.1",
		})

	assert.Equal(t, string(domain.CounterpartyMoral), c["counterparty_type"], "the deprecated mirror")
	assert.EqualValues(t, 125_000, c["isr_withheld_cents"])
	assert.EqualValues(t, 1_066_667, c["iva_withheld_cents"])
	assert.Equal(t, gorm.Expr("string_to_array(?, ?)", "ISR-4,IVA-6", ","), c["withholding_rule_ids"])
	assert.Equal(t, gorm.Expr("CAST(? AS numeric)", "0.0125"), c["withholding_isr_rate"])
	assert.Equal(t, "two_thirds_of_iva", c["withholding_iva_basis"])
	assert.Equal(t, "receptor", c["withholding_withholder"])
}

func TestWithholdingColumns_record(t *testing.T) {
	version, rate, basis, withholder := "2026.1", "0.012500", "two_thirds_of_iva", "receptor"
	got, err := withholdingColumns{
		RuleIDs: "ISR-4,IVA-6", TableVersion: &version, ISRRate: &rate, IVABasis: &basis, Withholder: &withholder,
	}.record(125_000, 1_066_667)
	require.NoError(t, err)
	r, err := money.ParseRate("0.0125")
	require.NoError(t, err)
	assert.Equal(t, fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: 125_000, IVA: 1_066_667}, ISRRate: r,
		IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
		RuleIDs: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6}, TableVersion: "2026.1",
	}, got)

	legacy, err := withholdingColumns{}.record(7, 8)
	require.NoError(t, err)
	assert.Equal(t, fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: 7, IVA: 8}}, legacy, "a pre-WH-06 row")

	half, err := withholdingColumns{RuleIDs: "ISR-4,IVA-6", TableVersion: &version}.record(1, 2)
	require.NoError(t, err)
	assert.False(t, half.Derived(), "a derivation without its detail is served as underived")

	bad := "1e9"
	_, err = withholdingColumns{TableVersion: &version, ISRRate: &bad}.record(0, 0)
	assert.Error(t, err)
}
