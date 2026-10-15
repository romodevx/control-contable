package fiscal_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func TestDerivation_Record(t *testing.T) {
	d, err := fiscal.DeriveWithholdings(fiscal.WithholdingInput{
		EmisorPersona: fiscal.PersonFisica, EmisorRegime: fiscal.RegimeBusiness, Receptor: fiscal.ReceptorGovFederal,
		Operation: fiscal.OperationProfessionalServices, Subtotal: 3_000_000, IVA: 480_000,
		IVARate: fiscal.IVARate16, On: date(t, "2026-07-10"),
	}, withholdingCatalog2026(t))
	require.NoError(t, err)

	r := d.Record()

	assert.Equal(t, fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: 300_000, IVA: 320_000},
		ISRRate:      rate(t, "0.10"), IVABasis: fiscal.IVABasisTwoThirds, Withholder: fiscal.WithholderReceptor,
		RuleIDs:      []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA7},
		TableVersion: fiscal.WithholdingTableVersion, CPAUnverified: true,
	}, r)
	assert.True(t, r.Derived())
	assert.False(t, fiscal.WithholdingRecord{}.Derived(), "a record stored before WH-06 has no derivation")
}

// A record is what DeriveWithholdings decided and nothing else: it never
// aliases the derivation's rule ids.
func TestDerivation_RecordCopiesRuleIDs(t *testing.T) {
	d := fiscal.Derivation{RuleIDs: []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA0}}

	r := d.Record()
	d.RuleIDs[0] = fiscal.RuleISR7

	assert.Equal(t, fiscal.RuleISR0, r.RuleIDs[0])
}

func TestWithholdingRecord_Net(t *testing.T) {
	r := fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: 125_000, IVA: 1_066_667}}

	net, err := r.Net(10_000_000, 1_600_000)

	require.NoError(t, err)
	assert.EqualValues(t, 10_408_333, net)

	_, err = r.Net(money.Cents(math.MaxInt64), 1)
	assert.ErrorIs(t, err, fiscal.ErrInvalidParameter)
}
