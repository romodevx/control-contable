package domain_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

func payee(person fiscal.PersonType, regime fiscal.RegimeCode, basis domain.IncomeBasis) domain.Payee {
	return domain.Payee{PersonType: person, Regime: regime, Basis: basis, RFCGivenToPlatform: true}
}

var (
	pf612 = payee(fiscal.PersonFisica, "612", domain.BasisCash)
	pf606 = payee(fiscal.PersonFisica, "606", domain.BasisCash)
	pf626 = payee(fiscal.PersonFisica, "626", domain.BasisCash)
	pf625 = payee(fiscal.PersonFisica, "625", domain.BasisCash)
	pm601 = payee(fiscal.PersonMoral, "601", domain.BasisAccrual)
	pm626 = payee(fiscal.PersonMoral, "626", domain.BasisCash)

	pf625goods      = domain.Payee{PersonType: fiscal.PersonFisica, Regime: "625", Basis: domain.BasisCash, RFCGivenToPlatform: true, PlatformActivities: []domain.PlatformActivity{domain.ActivityGoodsServices}}
	pf625goodsNoRFC = domain.Payee{PersonType: fiscal.PersonFisica, Regime: "625", Basis: domain.BasisCash, PlatformActivities: []domain.PlatformActivity{domain.ActivityGoodsServices}}
)

// TestPolicyFor is the applicability table of MODULE-incomes §2.6.
func TestPolicyFor(t *testing.T) {
	noRFC := pf625
	noRFC.RFCGivenToPlatform = false
	tests := []struct {
		name         string
		payee        domain.Payee
		counterparty domain.Counterparty
		activity     domain.PlatformActivity
		want         domain.WithholdingPolicy
	}{
		{"612 fees billed to a PM (LISR 106, RLIVA 3)", pf612, domain.CounterpartyMoral, "",
			policy(fiscal.KeyISRWithholdingFees, fiscal.ShareTwoThirds)},
		{"606 rent to a PM (LISR 116)", pf606, domain.CounterpartyMoral, "",
			policy(fiscal.KeyISRWithholdingRent, fiscal.ShareTwoThirds)},
		{"626 RESICO PF paid by a PM (LISR 113-J)", pf626, domain.CounterpartyMoral, "",
			policy(domain.KeyISRWithholdingRESICO, fiscal.ShareTwoThirds)},
		{"625 transport through a platform", pf625, domain.CounterpartyPlatform, domain.ActivityTransportDelivery,
			policy(domain.KeyISRPlatformTransport, fiscal.ShareHalf)},
		{"625 lodging through a platform", pf625, domain.CounterpartyPlatform, domain.ActivityLodging,
			policy(domain.KeyISRPlatformLodging, fiscal.ShareHalf)},
		{"625 goods through a platform (LIF 2026 art. 25-VI)", pf625, domain.CounterpartyPlatform, domain.ActivityGoodsServices,
			policy(domain.KeyISRPlatformGoods, fiscal.ShareHalf)},
		{"625 without RFC: 20 % and all the IVA (113-C, 18-J)", noRFC, domain.CounterpartyPlatform, domain.ActivityLodging,
			policy(domain.KeyISRPlatformNoRFC, fiscal.ShareWhole)},
		{"601 PM through a platform (B.2.3)", pm601, domain.CounterpartyPlatform, "",
			domain.WithholdingPolicy{ISR: domain.ISRAllowance{Allowed: true, Key: domain.KeyISRPlatformPersonaMoral}}},
		{"626 PM through a platform", pm626, domain.CounterpartyPlatform, "",
			domain.WithholdingPolicy{ISR: domain.ISRAllowance{Allowed: true, Key: domain.KeyISRPlatformPersonaMoral}}},
		{"PM paid by a PM: IVA only, no suggestion (freight 4 %)", pm601, domain.CounterpartyMoral, "",
			domain.WithholdingPolicy{IVA: domain.IVAAllowance{Allowed: true}}},
		{"612 paid by a PF", pf612, domain.CounterpartyFisica, "", domain.WithholdingPolicy{}},
		{"612 paid by the general public", pf612, domain.CounterpartyPublic, "", domain.WithholdingPolicy{}},
		{"601 paid by the general public", pm601, domain.CounterpartyPublic, "", domain.WithholdingPolicy{}},
		{"612 through a platform is outside the table", pf612, domain.CounterpartyPlatform, "", domain.WithholdingPolicy{}},
		{"625 paid directly by a PM is outside the table", pf625, domain.CounterpartyMoral, "", domain.WithholdingPolicy{}},
		{"625 platform without an activity", pf625, domain.CounterpartyPlatform, "", domain.WithholdingPolicy{}},
		{"605 informational regime paid by a PM", payee(fiscal.PersonFisica, "605", domain.BasisCash), domain.CounterpartyMoral, "",
			domain.WithholdingPolicy{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.PolicyFor(tt.payee, tt.counterparty, tt.activity))
		})
	}
}

func policy(isr fiscal.RateKey, iva fiscal.Share) domain.WithholdingPolicy {
	return domain.WithholdingPolicy{
		ISR: domain.ISRAllowance{Allowed: true, Key: isr},
		IVA: domain.IVAAllowance{Allowed: true, Share: iva},
	}
}

// TestSuggest computes the suggested withholdings of the research worked
// examples (A.5, B.2.3) and US-I1.
func TestSuggest(t *testing.T) {
	noRFC := pf625
	noRFC.RFCGivenToPlatform = false
	known := func(c money.Cents) fiscal.Suggestion { return fiscal.Suggestion{Cents: c, Known: true} }
	tests := []struct {
		name         string
		payee        domain.Payee
		counterparty domain.Counterparty
		activity     domain.PlatformActivity
		amount, iva  money.Cents
		wantISR      fiscal.Suggestion
		wantIVA      fiscal.Suggestion
	}{
		{"US-I1 / PF-612-IVA-01: 10,000 + 1,600 from a PM", pf612, domain.CounterpartyMoral, "",
			10_000_00, 1_600_00, known(1_000_00), known(1_066_67)},
		{"PF-612-ISR-01: 10 % of 30,000 of fees", pf612, domain.CounterpartyMoral, "",
			30_000_00, 4_800_00, known(3_000_00), known(3_200_00)},
		{"PF-626-ISR-01 / IVA-01: 1.25 % of 20,000; 2/3 of 3,200", pf626, domain.CounterpartyMoral, "",
			20_000_00, 3_200_00, known(250_00), known(2_133_33)},
		{"PF-606-ISR-02 / IVA-01: commercial rent 20,000 to a PM", pf606, domain.CounterpartyMoral, "",
			20_000_00, 3_200_00, known(2_000_00), known(2_133_33)},
		{"PF-606-IVA-01: residential rent is exempt", pf606, domain.CounterpartyMoral, "",
			20_000_00, 0, known(2_000_00), known(0)},
		{"PF-625-01: goods 30,000 + 4,800", pf625, domain.CounterpartyPlatform, domain.ActivityGoodsServices,
			30_000_00, 4_800_00, known(750_00), known(2_400_00)},
		{"PF-625-02 / US-I1: lodging 15,000", pf625, domain.CounterpartyPlatform, domain.ActivityLodging,
			15_000_00, 2_400_00, known(600_00), known(1_200_00)},
		{"625 transport 10,000 at 2.1 %", pf625, domain.CounterpartyPlatform, domain.ActivityTransportDelivery,
			10_000_00, 1_600_00, known(210_00), known(800_00)},
		{"625 without RFC: 20 % and 100 %", noRFC, domain.CounterpartyPlatform, domain.ActivityLodging,
			10_000_00, 1_600_00, known(2_000_00), known(1_600_00)},
		{"601 PM through a platform: 2.5 %", pm601, domain.CounterpartyPlatform, "",
			10_000_00, 1_600_00, known(250_00), fiscal.Suggestion{}},
		{"PM paid by a PM: IVA allowed with no suggestion", pm601, domain.CounterpartyMoral, "",
			10_000_00, 1_600_00, fiscal.Suggestion{}, fiscal.Suggestion{}},
		{"no withholding for the general public", pf612, domain.CounterpartyPublic, "",
			10_000_00, 1_600_00, fiscal.Suggestion{}, fiscal.Suggestion{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := domain.PolicyFor(tt.payee, tt.counterparty, tt.activity)
			got, err := p.Suggest(tt.amount, tt.iva, date(t, "2026-09-12"), catalog2026(t))
			require.NoError(t, err)
			assert.Equal(t, tt.wantISR, got.ISR, "isr")
			assert.Equal(t, tt.wantIVA, got.IVA, "iva")
		})
	}
}

func TestSuggestWithoutCatalogValue(t *testing.T) {
	p := domain.PolicyFor(pf612, domain.CounterpartyMoral, "")
	got, err := p.Suggest(10_000_00, 1_600_00, date(t, "2026-09-12"), fakeCatalog{})
	require.NoError(t, err)
	assert.False(t, got.ISR.Known, "a rate the catalog lacks is no suggestion")
	assert.Equal(t, fiscal.Suggestion{Cents: 1_066_67, Known: true}, got.IVA)
}

func TestSuggestParameterFault(t *testing.T) {
	absurd := fakeCatalog{fiscal.KeyISRWithholdingFees: {{from: date(t, "2022-01-01"), rate: math.MaxInt64}}}
	p := domain.PolicyFor(pf612, domain.CounterpartyMoral, "")
	_, err := p.Suggest(fiscal.MaxCents, 0, date(t, "2026-09-12"), absurd)
	require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
}

func TestSuggestShareOverflowIsAParameterFault(t *testing.T) {
	p := domain.PolicyFor(pf612, domain.CounterpartyMoral, "")
	_, err := p.Suggest(1_00, math.MaxInt64, date(t, "2026-09-12"), catalog2026(t))
	require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
}

func TestShareIsZero(t *testing.T) {
	assert.True(t, fiscal.Share{}.IsZero())
	assert.False(t, fiscal.ShareTwoThirds.IsZero())
}
