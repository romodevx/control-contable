package fiscal_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// fakeWithholdingCatalog is the test double of fiscal.WithholdingCatalog:
// the dated rates of fakeCatalog plus dated money parameters (WH-05 adds the
// real keys to params/data).
type fakeWithholdingCatalog struct {
	fakeCatalog
	amounts map[fiscal.AmountKey][]datedAmount
}

type datedAmount struct {
	from, to fiscal.Date
	cents    money.Cents
}

func (c fakeWithholdingCatalog) Amount(key fiscal.AmountKey, on fiscal.Date) (money.Cents, bool) {
	for _, v := range c.amounts[key] {
		if on.Compare(v.from) >= 0 && (v.to.IsZero() || on.Compare(v.to) <= 0) {
			return v.cents, true
		}
	}
	return 0, false
}

// withholdingCatalog2026 holds the 2026 withholding parameters of
// params/data/2026/v1.yaml; the LIF 2026 platform rates lapse on
// 2026-12-31, as they do there.
func withholdingCatalog2026(t *testing.T) fakeWithholdingCatalog {
	t.Helper()
	open := func(r string) []datedRate {
		return []datedRate{{from: date(t, "2026-01-01"), rate: rate(t, r)}}
	}
	lif2026 := func(r string) []datedRate {
		return []datedRate{{from: date(t, "2026-01-01"), to: date(t, "2026-12-31"), rate: rate(t, r)}}
	}
	return fakeWithholdingCatalog{
		fakeCatalog: fakeCatalog{
			fiscal.KeyISRWithholdingFees:           open("0.10"),
			fiscal.KeyISRWithholdingRent:           open("0.10"),
			fiscal.KeyISRWithholdingRESICO:         open("0.0125"),
			fiscal.KeyIVAWithholdingFreight:        open("0.04"),
			fiscal.KeyISRPlatformTransport:         open("0.021"),
			fiscal.KeyISRPlatformLodging:           open("0.04"),
			fiscal.KeyISRPlatformGoods:             lif2026("0.025"),
			fiscal.KeyISRPlatformNoRFC:             open("0.20"),
			fiscal.KeyISRPlatformPersonaMoral:      lif2026("0.025"),
			fiscal.KeyISRPlatformPersonaMoralNoRFC: lif2026("0.20"),
		},
		amounts: map[fiscal.AmountKey][]datedAmount{
			fiscal.KeyIVAFederalWithholdingThreshold: {{from: date(t, "2026-01-01"), to: date(t, "2026-12-31"), cents: 2_000_00}},
		},
	}
}

// in is the defaults every case overrides: a 612 PF billing a company
// $10,000.00 + IVA 16 % in July 2026.
func in(t *testing.T, edit func(*fiscal.WithholdingInput)) fiscal.WithholdingInput {
	t.Helper()
	v := fiscal.WithholdingInput{
		EmisorPersona: fiscal.PersonFisica,
		EmisorRegime:  fiscal.RegimeBusiness,
		Receptor:      fiscal.ReceptorPM,
		Operation:     fiscal.OperationProfessionalServices,
		Subtotal:      10_000_00,
		IVA:           1_600_00,
		IVARate:       fiscal.IVARate16,
		On:            date(t, "2026-07-10"),
	}
	edit(&v)
	return v
}

// pm turns the default emisor into a persona moral (601).
func pm(v *fiscal.WithholdingInput) {
	v.EmisorPersona = fiscal.PersonMoral
	v.EmisorRegime = "601"
}

func ids(isr, iva fiscal.RuleID) []fiscal.RuleID { return []fiscal.RuleID{isr, iva} }

// want is the part of a Derivation a decision-table case pins.
type want struct {
	isr, iva    money.Cents
	isrRate     string
	basis       fiscal.IVABasis
	withholder  fiscal.Withholder
	rules       []fiscal.RuleID
	ambiguities []fiscal.Ambiguity
	warnings    []fiscal.WarningCode
}

func assertDerivation(t *testing.T, w want, got fiscal.Derivation) {
	t.Helper()
	assert.Equal(t, w.isr, got.ISR, "ISR withheld")
	assert.Equal(t, w.iva, got.IVA, "IVA withheld")
	wantRate := money.Rate(0)
	if w.isrRate != "" {
		wantRate = rate(t, w.isrRate)
	}
	assert.Equal(t, wantRate, got.ISRRate, "ISR rate")
	assert.Equal(t, w.basis, got.IVABasis, "IVA basis")
	assert.Equal(t, w.withholder, got.Withholder, "withholder")
	assert.Equal(t, w.rules, got.RuleIDs, "rule ids")
	assert.Equal(t, w.ambiguities, got.Ambiguities, "ambiguities")
	assert.Equal(t, w.warnings, got.Warnings, "warnings")
	assert.Equal(t, len(w.ambiguities) > 0, got.CPAUnverified(), "cpa unverified")
}

// TestDeriveWithholdingsGoldens: WM §5 WH-01..WH-12 to the centavo, the
// owner's example first (MODULE-withholdings US-W1).
func TestDeriveWithholdingsGoldens(t *testing.T) {
	tests := []struct {
		name  string
		input func(*fiscal.WithholdingInput)
		want  want
		net   money.Cents
	}{
		{
			name: "WH-01 owner: RESICO PF services to a PM",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime = fiscal.RegimeRESICO
				v.Subtotal, v.IVA = 100_000_00, 16_000_00
			},
			want: want{
				isr: 1_250_00, iva: 10_666_67, isrRate: "0.0125", basis: fiscal.IVABasisTwoThirds,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR4, fiscal.RuleIVA6),
			},
			net: 104_083_33,
		},
		{
			name:  "WH-02 612 honorarios to a PM",
			input: func(v *fiscal.WithholdingInput) { v.Subtotal, v.IVA = 50_000_00, 8_000_00 },
			want: want{
				isr: 5_000_00, iva: 5_333_33, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR5, fiscal.RuleIVA6),
			},
			net: 47_666_67,
		},
		{
			name: "WH-03 606 commercial rent to a PM",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRent, fiscal.OperationRentRealEstate
				v.Subtotal, v.IVA = 30_000_00, 4_800_00
			},
			want: want{
				isr: 3_000_00, iva: 3_200_00, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR6, fiscal.RuleIVA6),
			},
			net: 28_600_00,
		},
		{
			name: "WH-03b 606 residential rent to a PM: exempt",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRent, fiscal.OperationRentRealEstate
				v.ResidentialRent, v.IVARate = true, fiscal.IVAExempt
				v.Subtotal, v.IVA = 30_000_00, 0
			},
			want: want{
				isr: 3_000_00, isrRate: "0.10", basis: fiscal.IVABasisNone,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR6, fiscal.RuleIVANoTax),
			},
			net: 27_000_00,
		},
		{
			name: "WH-04 612 lawyer to a 612 doctor (PF to PF)",
			input: func(v *fiscal.WithholdingInput) {
				v.Receptor = fiscal.ReceptorPFWithActivity
				v.Subtotal, v.IVA = 20_000_00, 3_200_00
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR0, fiscal.RuleIVA1)},
			net:  23_200_00,
		},
		{
			name: "WH-05 612-E sells goods to a PM",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation = fiscal.OperationSaleOfGoods
				v.Subtotal, v.IVA = 50_000_00, 8_000_00
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA8)},
			net:  58_000_00,
		},
		{
			name: "WH-05b RESICO PF sells goods to a PM (F5)",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRESICO, fiscal.OperationSaleOfGoods
				v.Subtotal, v.IVA = 50_000_00, 8_000_00
			},
			want: want{
				isr: 625_00, isrRate: "0.0125", basis: fiscal.IVABasisNone,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR4, fiscal.RuleIVA8),
			},
			net: 57_375_00,
		},
		{
			name: "WH-06 PM carrier to a PM: freight 4 % of the price",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Operation = fiscal.OperationGroundFreight
			},
			want: want{
				iva: 400_00, basis: fiscal.IVABasisFourPercentOfPrice,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR2, fiscal.RuleIVA4),
			},
			net: 11_200_00,
		},
		{
			name:  "WH-06b PF 612-E carrier to a PM",
			input: func(v *fiscal.WithholdingInput) { v.Operation = fiscal.OperationGroundFreight },
			want: want{
				iva: 400_00, basis: fiscal.IVABasisFourPercentOfPrice,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA4),
			},
			net: 11_200_00,
		},
		{
			name: "WH-07 PM to PM specialized services: the repealed 6 % is gone",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Operation = fiscal.OperationBusinessServices
				v.Subtotal, v.IVA = 100_000_00, 16_000_00
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR2, fiscal.RuleIVA8)},
			net:  116_000_00,
		},
		{
			name: "WH-08 PF sells scrap to a PM: 100 % of the IVA",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation = fiscal.OperationScrap
				v.Subtotal, v.IVA = 20_000_00, 3_200_00
			},
			want: want{
				iva: 3_200_00, basis: fiscal.IVABasisAllOfIVA,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA5),
			},
			net: 20_000_00,
		},
		{
			name: "WH-09 625 PF sells goods via a platform, RFC given",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
				v.PlatformActivity = fiscal.PlatformGoodsServices
			},
			want: want{
				isr: 250_00, iva: 800_00, isrRate: "0.025", basis: fiscal.IVABasisHalfOfIVA,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR3, fiscal.RuleIVA3),
			},
			net: 10_550_00,
		},
		{
			name: "WH-10 612 honorarios to público en general",
			input: func(v *fiscal.WithholdingInput) {
				v.Receptor = fiscal.ReceptorPFNoActivityOrPublic
				v.Subtotal, v.IVA = 5_000_00, 800_00
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR0, fiscal.RuleIVA0)},
			net:  5_800_00,
		},
		{
			name:  "WH-11 PM 601 pays a PF 612 consultant (expense side)",
			input: func(v *fiscal.WithholdingInput) { v.Subtotal, v.IVA = 40_000_00, 6_400_00 },
			want: want{
				isr: 4_000_00, iva: 4_266_67, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR5, fiscal.RuleIVA6),
			},
			net: 38_133_33,
		},
		{
			name: "WH-12 612 honorarios to the Federación: MVP defaults, flagged",
			input: func(v *fiscal.WithholdingInput) {
				v.Receptor = fiscal.ReceptorGovFederal
				v.Subtotal, v.IVA = 30_000_00, 4_800_00
			},
			want: want{
				isr: 3_000_00, iva: 3_200_00, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds,
				withholder: fiscal.WithholderReceptor, rules: ids(fiscal.RuleISR5, fiscal.RuleIVA7),
				ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityGovISR, fiscal.AmbiguityFederalIVAShare},
				warnings:    []fiscal.WarningCode{fiscal.WarningCPAUnverified},
			},
			net: 28_600_00,
		},
	}
	catalog := withholdingCatalog2026(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fiscal.DeriveWithholdings(in(t, tt.input), catalog)
			require.NoError(t, err)
			assertDerivation(t, tt.want, got)
			assert.Equal(t, tt.net, got.Net, "net")
		})
	}
}

// TestDeriveWithholdingsDecisionTable: one case per row of
// MODULE-withholdings §3.2 not already pinned by a golden, every receptor
// kind and every ⚠️ ambiguity included. Subtotal 10,000.00, IVA 1,600.00.
func TestDeriveWithholdingsDecisionTable(t *testing.T) {
	cpa := []fiscal.WarningCode{fiscal.WarningCPAUnverified}
	twoThirds := money.Cents(1_066_67)
	tests := []struct {
		name  string
		input func(*fiscal.WithholdingInput)
		want  want
	}{
		{
			name:  "foreign receptor: nothing withheld, not supported",
			input: func(v *fiscal.WithholdingInput) { v.Receptor = fiscal.ReceptorForeign },
			want: want{
				basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone,
				rules: ids(fiscal.RuleISR1, fiscal.RuleIVA0), warnings: []fiscal.WarningCode{fiscal.WarningForeignCounterparty},
			},
		},
		{
			name:  "foreign receptor, PM emisor: not supported either",
			input: func(v *fiscal.WithholdingInput) { pm(v); v.Receptor = fiscal.ReceptorForeign },
			want: want{
				basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone,
				rules: ids(fiscal.RuleISR1, fiscal.RuleIVA0), warnings: []fiscal.WarningCode{fiscal.WarningForeignCounterparty},
			},
		},
		{
			name: "PM sells through a platform: 2.5 % ISR, no IVA",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Operation = fiscal.OperationPlatform
			},
			want: want{
				isr: 250_00, isrRate: "0.025", basis: fiscal.IVABasisNone,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR2, fiscal.RuleIVA8),
			},
		},
		{
			name: "PM through a platform without RFC: 20 %",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Operation, v.PlatformWithoutRFC = fiscal.OperationPlatform, true
			},
			want: want{
				isr: 2_000_00, isrRate: "0.20", basis: fiscal.IVABasisNone,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR2, fiscal.RuleIVA8),
			},
		},
		{
			name: "625 transport via platform: 2.1 %",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
				v.PlatformActivity = fiscal.PlatformTransportDelivery
			},
			want: want{
				isr: 210_00, iva: 800_00, isrRate: "0.021", basis: fiscal.IVABasisHalfOfIVA,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR3, fiscal.RuleIVA3),
			},
		},
		{
			name: "625 lodging via platform: 4 %",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
				v.PlatformActivity = fiscal.PlatformLodging
			},
			want: want{
				isr: 400_00, iva: 800_00, isrRate: "0.04", basis: fiscal.IVABasisHalfOfIVA,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR3, fiscal.RuleIVA3),
			},
		},
		{
			name: "625 via platform without RFC: 20 % ISR, 100 % IVA",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
				v.PlatformWithoutRFC = true
			},
			want: want{
				isr: 2_000_00, iva: 1_600_00, isrRate: "0.20", basis: fiscal.IVABasisAllOfIVA,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR3, fiscal.RuleIVA3),
			},
		},
		{
			name: "625 via platform to the public: the platform still withholds",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
				v.PlatformActivity = fiscal.PlatformGoodsServices
				v.Receptor = fiscal.ReceptorPFNoActivityOrPublic
			},
			want: want{
				isr: 250_00, iva: 800_00, isrRate: "0.025", basis: fiscal.IVABasisHalfOfIVA,
				withholder: fiscal.WithholderPlatform, rules: ids(fiscal.RuleISR3, fiscal.RuleIVA3),
			},
		},
		{
			name:  "625 billing a PM directly: no ISR, A6",
			input: func(v *fiscal.WithholdingInput) { v.EmisorRegime = fiscal.RegimePlatforms },
			want: want{
				iva: twoThirds, basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA6), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityPlatformDirect}, warnings: cpa,
			},
		},
		{
			name: "625 selling goods to a PM directly: nothing, A6",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationSaleOfGoods
			},
			want: want{
				basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA8), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityPlatformDirect}, warnings: cpa,
			},
		},
		{
			name:  "PM emisor, scrap to a PM: 100 % of the IVA",
			input: func(v *fiscal.WithholdingInput) { pm(v); v.Operation = fiscal.OperationScrap },
			want: want{
				iva: 1_600_00, basis: fiscal.IVABasisAllOfIVA, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR2, fiscal.RuleIVA5),
			},
		},
		{
			name: "PM emisor, freight to the Federación: 4 %",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Operation, v.Receptor = fiscal.OperationGroundFreight, fiscal.ReceptorGovFederal
			},
			want: want{
				iva: 400_00, basis: fiscal.IVABasisFourPercentOfPrice, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR2, fiscal.RuleIVA4),
			},
		},
		{
			name: "PM emisor, scrap to the Federación: nothing",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Operation, v.Receptor = fiscal.OperationScrap, fiscal.ReceptorGovFederal
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR2, fiscal.RuleIVA8)},
		},
		{
			name: "PM emisor, honorarios to a state: nothing",
			input: func(v *fiscal.WithholdingInput) {
				pm(v)
				v.Receptor = fiscal.ReceptorGovStateMunicipal
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR2, fiscal.RuleIVA2)},
		},
		{
			name: "RESICO PF commercial rent to a PM: 1.25 %, A3",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRESICO, fiscal.OperationRentRealEstate
			},
			want: want{
				isr: 125_00, iva: twoThirds, isrRate: "0.0125", basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR4, fiscal.RuleIVA6), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityRESICORent}, warnings: cpa,
			},
		},
		{
			name: "RESICO PF to a state: 1.25 % ISR (A1), no IVA withheld",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Receptor = fiscal.RegimeRESICO, fiscal.ReceptorGovStateMunicipal
			},
			want: want{
				isr: 125_00, isrRate: "0.0125", basis: fiscal.IVABasisNone, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR4, fiscal.RuleIVA2), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityGovISR}, warnings: cpa,
			},
		},
		{
			name: "RESICO PF other operation: ISR only",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRESICO, fiscal.OperationOther
			},
			want: want{
				isr: 125_00, isrRate: "0.0125", basis: fiscal.IVABasisNone, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR4, fiscal.RuleIVA8),
			},
		},
		{
			name: "RESICO PF movable rent: 1.25 % and 2/3",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRESICO, fiscal.OperationRentMovable
			},
			want: want{
				isr: 125_00, iva: twoThirds, isrRate: "0.0125", basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR4, fiscal.RuleIVA6),
			},
		},
		{
			name:  "612 commercial rent to a PM: 10 %",
			input: func(v *fiscal.WithholdingInput) { v.Operation = fiscal.OperationRentRealEstate },
			want: want{
				isr: 1_000_00, iva: twoThirds, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR6, fiscal.RuleIVA6),
			},
		},
		{
			name: "606 rent to a state: 10 % ISR (A1), no IVA withheld",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation = fiscal.RegimeRent, fiscal.OperationRentRealEstate
				v.Receptor = fiscal.ReceptorGovStateMunicipal
			},
			want: want{
				isr: 1_000_00, isrRate: "0.10", basis: fiscal.IVABasisNone, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR6, fiscal.RuleIVA2), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityGovISR}, warnings: cpa,
			},
		},
		{
			name:  "612 movable rent to a PM: no ISR (A10), 2/3",
			input: func(v *fiscal.WithholdingInput) { v.Operation = fiscal.OperationRentMovable },
			want: want{
				iva: twoThirds, basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA6), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityMovableRentISR}, warnings: cpa,
			},
		},
		{
			name:  "PF commission to a PM: no ISR (A7), 2/3",
			input: func(v *fiscal.WithholdingInput) { v.Operation = fiscal.OperationCommission },
			want: want{
				iva: twoThirds, basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA6), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityCommissionISR}, warnings: cpa,
			},
		},
		{
			name:  "PF business services to a PM: nothing (F4)",
			input: func(v *fiscal.WithholdingInput) { v.Operation = fiscal.OperationBusinessServices },
			want:  want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA8)},
		},
		{
			name:  "PF other operation to a PM: nothing",
			input: func(v *fiscal.WithholdingInput) { v.Operation = fiscal.OperationOther },
			want:  want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA8)},
		},
		{
			name: "PF freight to the Federación: 4 %",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.Receptor = fiscal.OperationGroundFreight, fiscal.ReceptorGovFederal
			},
			want: want{
				iva: 400_00, basis: fiscal.IVABasisFourPercentOfPrice, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA4),
			},
		},
		{
			name: "PF freight to a state: nothing",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.Receptor = fiscal.OperationGroundFreight, fiscal.ReceptorGovStateMunicipal
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA2)},
		},
		{
			name: "PF goods to the Federación above $2,000: 2/3 (A2)",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.Receptor = fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal
			},
			want: want{
				iva: twoThirds, basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA7), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityFederalIVAShare}, warnings: cpa,
			},
		},
		{
			name: "PF goods to the Federación of exactly $2,000: nothing",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.Receptor = fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal
				v.Subtotal, v.IVA = 2_000_00, 320_00
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA8)},
		},
		{
			name: "PF other operation to the Federación: never IVA",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.Receptor = fiscal.OperationOther, fiscal.ReceptorGovFederal
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR7, fiscal.RuleIVA8)},
		},
		{
			name: "PF scrap to a PF: nothing",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.Receptor = fiscal.OperationScrap, fiscal.ReceptorPFWithActivity
			},
			want: want{basis: fiscal.IVABasisNone, withholder: fiscal.WithholderNone, rules: ids(fiscal.RuleISR0, fiscal.RuleIVA1)},
		},
		{
			name: "border 8 %: 2/3 of the IVA charged (A8)",
			input: func(v *fiscal.WithholdingInput) {
				v.IVARate, v.IVA = fiscal.IVARate8, 800_00
			},
			want: want{
				isr: 1_000_00, iva: 533_33, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR5, fiscal.RuleIVA6), ambiguities: []fiscal.Ambiguity{fiscal.AmbiguityBorderShare}, warnings: cpa,
			},
		},
		{
			name: "border 8 % freight stays 4 % of the price",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.IVARate, v.IVA = fiscal.OperationGroundFreight, fiscal.IVARate8, 800_00
			},
			want: want{
				iva: 400_00, basis: fiscal.IVABasisFourPercentOfPrice, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA4),
			},
		},
		{
			name: "IVA 0 %: no IVA to withhold",
			input: func(v *fiscal.WithholdingInput) {
				v.IVARate, v.IVA = fiscal.IVARate0, 0
			},
			want: want{
				isr: 1_000_00, isrRate: "0.10", basis: fiscal.IVABasisNone, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR5, fiscal.RuleIVANoTax),
			},
		},
		{
			name: "residential rent at a taxable rate is still no IVA (LIVA 20 fr. II)",
			input: func(v *fiscal.WithholdingInput) {
				v.EmisorRegime, v.Operation, v.ResidentialRent = fiscal.RegimeRent, fiscal.OperationRentRealEstate, true
			},
			want: want{
				isr: 1_000_00, isrRate: "0.10", basis: fiscal.IVABasisNone, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR6, fiscal.RuleIVANoTax),
			},
		},
		{
			name: "the residential flag means nothing outside real-estate rent",
			input: func(v *fiscal.WithholdingInput) {
				v.ResidentialRent = true
			},
			want: want{
				isr: 1_000_00, iva: twoThirds, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR5, fiscal.RuleIVA6),
			},
		},
		{
			name: "freight never withholds more IVA than was charged",
			input: func(v *fiscal.WithholdingInput) {
				v.Operation, v.IVA = fiscal.OperationGroundFreight, 300_00
			},
			want: want{
				iva: 300_00, basis: fiscal.IVABasisFourPercentOfPrice, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR7, fiscal.RuleIVA4),
			},
		},
		{
			name: "ambiguities are listed once, in ID order",
			input: func(v *fiscal.WithholdingInput) {
				v.Receptor, v.IVARate, v.IVA = fiscal.ReceptorGovFederal, fiscal.IVARate8, 800_00
			},
			want: want{
				isr: 1_000_00, iva: 533_33, isrRate: "0.10", basis: fiscal.IVABasisTwoThirds, withholder: fiscal.WithholderReceptor,
				rules: ids(fiscal.RuleISR5, fiscal.RuleIVA7),
				ambiguities: []fiscal.Ambiguity{
					fiscal.AmbiguityGovISR, fiscal.AmbiguityFederalIVAShare, fiscal.AmbiguityBorderShare,
				},
				warnings: cpa,
			},
		},
	}
	catalog := withholdingCatalog2026(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fiscal.DeriveWithholdings(in(t, tt.input), catalog)
			require.NoError(t, err)
			assertDerivation(t, tt.want, got)
		})
	}
}

func TestDeriveWithholdingsRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input func(*fiscal.WithholdingInput)
		field string
		err   fiscal.Error
	}{
		{"no emisor person type", func(v *fiscal.WithholdingInput) { v.EmisorPersona = "" }, "emisor_person_type", fiscal.ErrRequired},
		{"unknown emisor person type", func(v *fiscal.WithholdingInput) { v.EmisorPersona = "x" }, "emisor_person_type", fiscal.ErrInvalid},
		{"PF without regime", func(v *fiscal.WithholdingInput) { v.EmisorRegime = "" }, "emisor_regime", fiscal.ErrRequired},
		{"no receptor kind", func(v *fiscal.WithholdingInput) { v.Receptor = "" }, "receptor_kind", fiscal.ErrRequired},
		{"unknown receptor kind", func(v *fiscal.WithholdingInput) { v.Receptor = "x" }, "receptor_kind", fiscal.ErrInvalid},
		{"no operation kind", func(v *fiscal.WithholdingInput) { v.Operation = "" }, "operation_kind", fiscal.ErrRequired},
		{"unknown operation kind", func(v *fiscal.WithholdingInput) { v.Operation = "subcontracting" }, "operation_kind", fiscal.ErrInvalid},
		{"PF platform sale without activity", func(v *fiscal.WithholdingInput) {
			v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
		}, "platform_activity", fiscal.ErrRequired},
		{"unknown platform activity", func(v *fiscal.WithholdingInput) {
			v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
			v.PlatformActivity = "x"
		}, "platform_activity", fiscal.ErrInvalid},
		{"negative subtotal", func(v *fiscal.WithholdingInput) { v.Subtotal = -1 }, "subtotal_cents", fiscal.ErrOutOfRange},
		{"subtotal above the contract maximum", func(v *fiscal.WithholdingInput) { v.Subtotal = fiscal.MaxCents + 1 }, "subtotal_cents", fiscal.ErrOutOfRange},
		{"negative IVA", func(v *fiscal.WithholdingInput) { v.IVA = -1 }, "iva_cents", fiscal.ErrOutOfRange},
		{"no IVA rate", func(v *fiscal.WithholdingInput) { v.IVARate = "" }, "iva_rate", fiscal.ErrRequired},
	}
	catalog := withholdingCatalog2026(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fiscal.DeriveWithholdings(in(t, tt.input), catalog)
			var vs fiscal.Violations
			require.ErrorAs(t, err, &vs)
			require.ErrorIs(t, err, tt.err)
			assert.Equal(t, tt.field, vs[0].Field)
		})
	}
}

func TestDeriveWithholdingsPMNeedsNoRegime(t *testing.T) {
	_, err := fiscal.DeriveWithholdings(in(t, func(v *fiscal.WithholdingInput) {
		v.EmisorPersona, v.EmisorRegime = fiscal.PersonMoral, ""
	}), withholdingCatalog2026(t))
	require.NoError(t, err)
}

func TestDeriveWithholdingsParameterFaults(t *testing.T) {
	t.Run("no catalog", func(t *testing.T) {
		_, err := fiscal.DeriveWithholdings(in(t, func(*fiscal.WithholdingInput) {}), nil)
		require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
	})
	t.Run("no date", func(t *testing.T) {
		_, err := fiscal.DeriveWithholdings(in(t, func(v *fiscal.WithholdingInput) { v.On = fiscal.Date{} }), withholdingCatalog2026(t))
		require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
	})
	t.Run("absurd rate overflows", func(t *testing.T) {
		catalog := withholdingCatalog2026(t)
		catalog.fakeCatalog[fiscal.KeyISRWithholdingFees] = []datedRate{{from: date(t, "2026-01-01"), rate: money.Rate(1 << 62)}}
		_, err := fiscal.DeriveWithholdings(in(t, func(v *fiscal.WithholdingInput) { v.Subtotal = fiscal.MaxCents }), catalog)
		require.ErrorIs(t, err, fiscal.ErrInvalidParameter)
	})
}

// TestDeriveWithholdingsNeedsRatesInForce: a rate or threshold the catalog
// lacks on the record's date is not guessed (the LIF 2026 platform rate
// lapses on 2026-12-31).
func TestDeriveWithholdingsNeedsRatesInForce(t *testing.T) {
	tests := []struct {
		name  string
		input func(*fiscal.WithholdingInput)
	}{
		{"platform rate after the LIF lapses", func(v *fiscal.WithholdingInput) {
			v.EmisorRegime, v.Operation = fiscal.RegimePlatforms, fiscal.OperationPlatform
			v.PlatformActivity = fiscal.PlatformGoodsServices
			v.On = date(t, "2027-01-04")
		}},
		{"ISR rate before any value", func(v *fiscal.WithholdingInput) { v.On = date(t, "2025-12-31") }},
		{"freight rate before any value", func(v *fiscal.WithholdingInput) {
			v.Operation, v.On = fiscal.OperationGroundFreight, date(t, "2025-12-31")
		}},
		{"federal threshold before any value", func(v *fiscal.WithholdingInput) {
			v.Operation, v.Receptor = fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal
			v.On = date(t, "2025-12-31")
		}},
		{"federal threshold after the RMF 2026 lapses", func(v *fiscal.WithholdingInput) {
			v.Operation, v.Receptor = fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal
			v.On = date(t, "2027-01-04")
		}},
	}
	catalog := withholdingCatalog2026(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fiscal.DeriveWithholdings(in(t, tt.input), catalog)
			require.ErrorIs(t, err, fiscal.ErrRateNotValidOnDate)
		})
	}
}

// TestDeriveWithholdingsReadsTheFederalThreshold: IVA-7's $2,000 is the
// catalog's dated value (RMF 4.1.3), never a constant in code.
func TestDeriveWithholdingsReadsTheFederalThreshold(t *testing.T) {
	goodsToFederation := func(subtotal money.Cents) fiscal.WithholdingInput {
		return in(t, func(v *fiscal.WithholdingInput) {
			v.Operation, v.Receptor = fiscal.OperationSaleOfGoods, fiscal.ReceptorGovFederal
			v.Subtotal, v.IVA = subtotal, subtotal*16/100
		})
	}
	catalog := withholdingCatalog2026(t)
	catalog.amounts = map[fiscal.AmountKey][]datedAmount{
		fiscal.KeyIVAFederalWithholdingThreshold: {{from: date(t, "2026-01-01"), cents: 3_000_00}},
	}
	tests := []struct {
		subtotal money.Cents
		want     fiscal.RuleID
	}{
		{2_500_00, fiscal.RuleIVA8}, // above $2,000, not above this catalog's $3,000
		{3_000_00, fiscal.RuleIVA8},
		{3_000_01, fiscal.RuleIVA7},
	}
	for _, tt := range tests {
		got, err := fiscal.DeriveWithholdings(goodsToFederation(tt.subtotal), catalog)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got.RuleIDs[1], "subtotal %d", tt.subtotal)
	}
}

// TestDeriveWithholdingsInvariants: MODULE-withholdings §3.2 properties over
// random valid inputs (fixed seed, deterministic).
func TestDeriveWithholdingsInvariants(t *testing.T) {
	const samples = 20_000
	catalog := withholdingCatalog2026(t)
	on := date(t, "2026-07-10")
	rng := rand.New(rand.NewPCG(2026, 7)) //nolint:gosec // deterministic test data, not security
	for range samples {
		v := randomInput(rng, on)
		got, err := fiscal.DeriveWithholdings(v, catalog)
		require.NoError(t, err, "%+v", v)

		again, err := fiscal.DeriveWithholdings(v, catalog)
		require.NoError(t, err)
		require.Equal(t, got, again, "deterministic: %+v", v)

		require.GreaterOrEqual(t, got.ISR, money.Cents(0), "%+v", v)
		require.LessOrEqual(t, got.ISR, v.Subtotal, "ISR ≤ subtotal: %+v", v)
		require.GreaterOrEqual(t, got.IVA, money.Cents(0), "%+v", v)
		require.LessOrEqual(t, got.IVA, v.IVA, "IVA withheld ≤ IVA: %+v", v)
		require.Equal(t, v.Subtotal+v.IVA-got.ISR-got.IVA, got.Net, "net: %+v", v)
		require.GreaterOrEqual(t, got.Net, money.Cents(0), "%+v", v)
		require.Len(t, got.RuleIDs, 2, "%+v", v)
		require.Equal(t, len(got.Ambiguities) > 0, got.CPAUnverified())

		assertNoOneWithholdsWhenItShouldNot(t, v, got)
	}
}

// assertNoOneWithholdsWhenItShouldNot: a PF payer never withholds (F1, F2),
// PM → PM withholds only on freight and scrap, and nothing that is not
// taxed carries IVA withheld.
func assertNoOneWithholdsWhenItShouldNot(t *testing.T, v fiscal.WithholdingInput, got fiscal.Derivation) {
	t.Helper()
	viaPlatform := v.Operation == fiscal.OperationPlatform
	pfPayer := v.Receptor == fiscal.ReceptorPFWithActivity || v.Receptor == fiscal.ReceptorPFNoActivityOrPublic
	if pfPayer && !viaPlatform {
		require.Equal(t, fiscal.Withholdings{}, got.Withholdings, "PF payer: %+v", v)
	}
	require.NotEqual(t, fiscal.WithholderReceptor, pfPayerWithholder(pfPayer, got), "PF payer: %+v", v)
	pmToPM := v.EmisorPersona == fiscal.PersonMoral && v.Receptor == fiscal.ReceptorPM
	freightOrScrap := v.Operation == fiscal.OperationGroundFreight || v.Operation == fiscal.OperationScrap
	if pmToPM && !freightOrScrap && !viaPlatform {
		require.Equal(t, fiscal.Withholdings{}, got.Withholdings, "PM → PM: %+v", v)
	}
	if !v.IVARate.Taxable() {
		require.Zero(t, got.IVA, "untaxed: %+v", v)
	}
}

func pfPayerWithholder(pfPayer bool, got fiscal.Derivation) fiscal.Withholder {
	if !pfPayer {
		return fiscal.WithholderNone
	}
	return got.Withholder
}

func randomInput(rng *rand.Rand, on fiscal.Date) fiscal.WithholdingInput {
	persons := []fiscal.PersonType{fiscal.PersonFisica, fiscal.PersonMoral}
	regimes := []fiscal.RegimeCode{fiscal.RegimeRent, fiscal.RegimeBusiness, "621", fiscal.RegimePlatforms, fiscal.RegimeRESICO}
	receptors := []fiscal.ReceptorKind{
		fiscal.ReceptorPM, fiscal.ReceptorGovFederal, fiscal.ReceptorGovStateMunicipal,
		fiscal.ReceptorPFWithActivity, fiscal.ReceptorPFNoActivityOrPublic, fiscal.ReceptorForeign,
	}
	operations := []fiscal.OperationKind{
		fiscal.OperationProfessionalServices, fiscal.OperationBusinessServices, fiscal.OperationRentRealEstate,
		fiscal.OperationRentMovable, fiscal.OperationSaleOfGoods, fiscal.OperationGroundFreight,
		fiscal.OperationCommission, fiscal.OperationScrap, fiscal.OperationPlatform, fiscal.OperationOther,
	}
	activities := []fiscal.PlatformActivity{
		fiscal.PlatformTransportDelivery, fiscal.PlatformLodging, fiscal.PlatformGoodsServices,
	}
	treatments := []fiscal.IVARate{fiscal.IVARate16, fiscal.IVARate8, fiscal.IVARate0, fiscal.IVAExempt}
	pick := func(n int) int { return rng.IntN(n) }

	treatment := treatments[pick(len(treatments))]
	subtotal := money.Cents(rng.Int64N(int64(fiscal.MaxCents) / 4))
	var iva money.Cents
	if treatment.Taxable() {
		// The recorded IVA may drift from subtotal × rate (IVATolerance).
		iva = money.Cents(rng.Int64N(int64(subtotal)/5 + 1))
	}
	return fiscal.WithholdingInput{
		EmisorPersona:      persons[pick(len(persons))],
		EmisorRegime:       regimes[pick(len(regimes))],
		Receptor:           receptors[pick(len(receptors))],
		Operation:          operations[pick(len(operations))],
		PlatformActivity:   activities[pick(len(activities))],
		PlatformWithoutRFC: pick(2) == 0,
		ResidentialRent:    pick(2) == 0,
		Subtotal:           subtotal,
		IVA:                iva,
		IVARate:            treatment,
		On:                 on,
	}
}

func TestWithholdingTableVersion(t *testing.T) {
	assert.Equal(t, "2026.1", fiscal.WithholdingTableVersion)
}
