package fiscal

import "slices"

// WithholdingTableVersion identifies the decision table DeriveWithholdings
// implements; it bumps on any rule or rate change so stored derivations can
// be re-derived (MODULE-withholdings §5.3).
const WithholdingTableVersion = "2026.1"

// ReceptorKind is who receives the CFDI and pays (WM §1.1); on an income it
// is the contract counterparty_kind, on an expense always ReceptorPM.
type ReceptorKind string

// The receptor kinds of WM §4.1.
const (
	ReceptorPM                   ReceptorKind = "pm"
	ReceptorGovFederal           ReceptorKind = "gov_federal"
	ReceptorGovStateMunicipal    ReceptorKind = "gov_state_municipal"
	ReceptorPFWithActivity       ReceptorKind = "pf_with_activity"
	ReceptorPFNoActivityOrPublic ReceptorKind = "pf_no_activity_or_public"
	ReceptorForeign              ReceptorKind = "foreign"
)

// Valid reports a known ReceptorKind.
func (k ReceptorKind) Valid() bool {
	return slices.Contains([]ReceptorKind{
		ReceptorPM, ReceptorGovFederal, ReceptorGovStateMunicipal,
		ReceptorPFWithActivity, ReceptorPFNoActivityOrPublic, ReceptorForeign,
	}, k)
}

// isPF reports a persona física payer, with or without activity: it never
// withholds (WM F1, F2).
func (k ReceptorKind) isPF() bool {
	return k == ReceptorPFWithActivity || k == ReceptorPFNoActivityOrPublic
}

// isGov reports a government payer (WM A1 for its ISR withholding).
func (k ReceptorKind) isGov() bool {
	return k == ReceptorGovFederal || k == ReceptorGovStateMunicipal
}

// OperationKind is what the CFDI sells (WM §4.1). There is no
// subcontracting kind: its 6 % IVA withholding was repealed (WM F6).
type OperationKind string

// The operation kinds of WM §4.1.
const (
	OperationProfessionalServices OperationKind = "professional_services"
	OperationBusinessServices     OperationKind = "business_services"
	OperationRentRealEstate       OperationKind = "rent_real_estate"
	OperationRentMovable          OperationKind = "rent_movable"
	OperationSaleOfGoods          OperationKind = "sale_of_goods"
	OperationGroundFreight        OperationKind = "ground_freight"
	OperationCommission           OperationKind = "commission"
	OperationScrap                OperationKind = "scrap"
	OperationPlatform             OperationKind = "platform"
	OperationOther                OperationKind = "other"
)

// Valid reports a known OperationKind.
func (o OperationKind) Valid() bool {
	return slices.Contains([]OperationKind{
		OperationProfessionalServices, OperationBusinessServices, OperationRentRealEstate,
		OperationRentMovable, OperationSaleOfGoods, OperationGroundFreight,
		OperationCommission, OperationScrap, OperationPlatform, OperationOther,
	}, o)
}

// personal reports the operations whose IVA a persona moral withholds 2/3 of
// when a persona física provides them (LIVA 1-A fr. II a, d; RLIVA 3 fr. I).
func (o OperationKind) personal() bool {
	return slices.Contains([]OperationKind{
		OperationProfessionalServices, OperationRentRealEstate, OperationRentMovable, OperationCommission,
	}, o)
}

// PlatformActivity is what a persona física sells through a digital
// platform; it picks the platform's ISR rate (LISR 113-A).
type PlatformActivity string

// The platform activities of LISR 113-A fr. I–III.
const (
	PlatformTransportDelivery PlatformActivity = "transport_delivery"
	PlatformLodging           PlatformActivity = "lodging"
	PlatformGoodsServices     PlatformActivity = "goods_services"
)

// Valid reports a known PlatformActivity.
func (a PlatformActivity) Valid() bool {
	return slices.Contains([]PlatformActivity{PlatformTransportDelivery, PlatformLodging, PlatformGoodsServices}, a)
}

// IVABasis is how the IVA withheld was computed (MODULE-withholdings §3.3).
type IVABasis string

// The IVA withholding bases.
const (
	IVABasisNone               IVABasis = "none"
	IVABasisTwoThirds          IVABasis = "two_thirds_of_iva"
	IVABasisHalfOfIVA          IVABasis = "half_of_iva"
	IVABasisAllOfIVA           IVABasis = "all_of_iva"
	IVABasisFourPercentOfPrice IVABasis = "four_percent_of_price"
)

// Withholder is who withholds; the module maps it to its side's wording
// ("counterparty" on an income, "you" on an expense).
type Withholder string

// The withholders of WM §4.2 outputs.
const (
	WithholderNone     Withholder = "none"
	WithholderReceptor Withholder = "receptor"
	WithholderPlatform Withholder = "platform"
)

// RuleID cites the WM §4.2 row a withholding came from.
type RuleID string

// The ISR rules of WM §4.2, in evaluation order.
const (
	RuleISR0 RuleID = "ISR-0" // receptor is a persona física
	RuleISR1 RuleID = "ISR-1" // receptor is foreign
	RuleISR2 RuleID = "ISR-2" // emisor is a persona moral (platform: 2.5 %)
	RuleISR3 RuleID = "ISR-3" // persona física via a platform
	RuleISR4 RuleID = "ISR-4" // RESICO PF: 1.25 %
	RuleISR5 RuleID = "ISR-5" // 612 professional services: 10 %
	RuleISR6 RuleID = "ISR-6" // 606/612 real-estate rent: 10 %
	RuleISR7 RuleID = "ISR-7" // otherwise: 0
)

// The IVA rules of WM §4.2. RuleIVANoTax is the table's preamble: a 0 %,
// exempt or residential-rent operation charges no IVA to withhold.
const (
	RuleIVANoTax RuleID = "IVA-NONE"
	RuleIVA0     RuleID = "IVA-0" // receptor is público en general or foreign
	RuleIVA1     RuleID = "IVA-1" // receptor is a PF with activity
	RuleIVA2     RuleID = "IVA-2" // receptor is a state or municipality
	RuleIVA3     RuleID = "IVA-3" // persona física via a platform
	RuleIVA4     RuleID = "IVA-4" // ground freight: 4 % of the price
	RuleIVA5     RuleID = "IVA-5" // scrap: 100 %
	RuleIVA6     RuleID = "IVA-6" // PF personal services or rent to a PM: 2/3
	RuleIVA7     RuleID = "IVA-7" // PF to the Federación: 2/3 (⚠️ A2)
	RuleIVA8     RuleID = "IVA-8" // otherwise: 0
)

// Ambiguity cites a WM §6 question the CPA has not answered; the rule that
// raised it applies the MVP default.
type Ambiguity string

// The ambiguities DeriveWithholdings can raise, in WM §6 order.
const (
	AmbiguityGovISR          Ambiguity = "A1"  // government withholding ISR
	AmbiguityFederalIVAShare Ambiguity = "A2"  // the Federación's IVA fraction
	AmbiguityRESICORent      Ambiguity = "A3"  // RESICO PF rent: 1.25 % not 10 %
	AmbiguityPlatformDirect  Ambiguity = "A6"  // 625 PF billing a PM directly
	AmbiguityCommissionISR   Ambiguity = "A7"  // PF commission agent's ISR
	AmbiguityBorderShare     Ambiguity = "A8"  // 2/3 of the border 8 % IVA
	AmbiguityMovableRentISR  Ambiguity = "A10" // PF movable rent's ISR
)

// ambiguityOrder is the WM §6 order a Derivation lists them in.
var ambiguityOrder = []Ambiguity{
	AmbiguityGovISR, AmbiguityFederalIVAShare, AmbiguityRESICORent, AmbiguityPlatformDirect,
	AmbiguityCommissionISR, AmbiguityBorderShare, AmbiguityMovableRentISR,
}

// WarningCode is a non-blocking warning a derivation adds to its record
// (MODULE-withholdings §6). Module-only warnings (supplier_regime_unknown,
// deduction_requires_remittance) are the module's to add.
type WarningCode string

// The warnings DeriveWithholdings raises.
const (
	WarningCPAUnverified       WarningCode = "withholding_cpa_unverified"
	WarningForeignCounterparty WarningCode = "foreign_counterparty_not_supported"
)
