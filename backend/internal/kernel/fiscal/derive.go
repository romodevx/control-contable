package fiscal

import (
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// WithholdingInput is what DeriveWithholdings needs to know about one
// operation (MODULE-withholdings §3.1, WM §4.1). On an income the taxpayer
// is the emisor; on an expense the supplier is, and the receptor is the
// taxpayer (always ReceptorPM). The receptor's own regime is not an input:
// every persona moral withholds alike and a persona física never does
// (OQ-W1).
type WithholdingInput struct {
	EmisorPersona PersonType
	// EmisorRegime is required for a persona física; a persona moral's
	// regime never changes the result (WM §1.2 "PM-any").
	EmisorRegime RegimeCode
	Receptor     ReceptorKind
	Operation    OperationKind
	// PlatformActivity picks a persona física's platform ISR rate; it is
	// required when she sells through a platform with her RFC given.
	PlatformActivity PlatformActivity
	// PlatformWithoutRFC is a platform sale whose seller gave the platform
	// no RFC: 20 % ISR and, for a persona física, 100 % of the IVA.
	PlatformWithoutRFC bool
	// ResidentialRent is a dwelling's rent: IVA-exempt (LIVA 20 fr. II).
	// It means nothing outside OperationRentRealEstate.
	ResidentialRent bool
	// Subtotal is the price without IVA, the base of the ISR and freight
	// withholdings.
	Subtotal money.Cents
	// IVA is the IVA the record charges, the base of the IVA fractions.
	IVA     money.Cents
	IVARate IVARate
	// On is the record's date: the rates in force that day apply.
	On Date
}

// Derivation is the result of DeriveWithholdings (MODULE-withholdings §3.3).
type Derivation struct {
	Withholdings
	// ISRRate is the ISR withholding rate applied; zero when none.
	ISRRate    money.Rate
	IVABasis   IVABasis
	Withholder Withholder
	// Net is what changes hands: Subtotal + IVA − ISR − IVA withheld.
	Net money.Cents
	// RuleIDs cites the WM §4.2 rows that decided: the ISR rule, then the
	// IVA rule.
	RuleIDs []RuleID
	// Ambiguities lists the ⚠️ WM §6 questions the derivation relied on,
	// in WM order; nil when every rule applied is confirmed.
	Ambiguities []Ambiguity
	Warnings    []WarningCode
}

// CPAUnverified reports a derivation that relied on a ⚠️ rule.
func (d Derivation) CPAUnverified() bool { return len(d.Ambiguities) > 0 }

// isrRule is the ISR row that matched; a zero key withholds nothing.
type isrRule struct {
	id          RuleID
	key         RateKey
	withholder  Withholder
	ambiguities []Ambiguity
}

// ivaRule is the IVA row that matched. aboveFederalThreshold marks IVA-7
// on an operation other than professional services: it applies only when
// the subtotal exceeds the RMF 4.1.3 threshold, else IVA-8 does.
type ivaRule struct {
	id                    RuleID
	basis                 IVABasis
	withholder            Withholder
	ambiguities           []Ambiguity
	aboveFederalThreshold bool
}

// ivaShares are the bases that are a legal fraction of the IVA charged.
var ivaShares = map[IVABasis]Share{
	IVABasisTwoThirds: ShareTwoThirds,
	IVABasisHalfOfIVA: ShareHalf,
	IVABasisAllOfIVA:  ShareWhole,
}

// platformRateKeys maps a persona física's platform activity to its ISR rate.
var platformRateKeys = map[PlatformActivity]RateKey{
	PlatformTransportDelivery: KeyISRPlatformTransport,
	PlatformLodging:           KeyISRPlatformLodging,
	PlatformGoodsServices:     KeyISRPlatformGoods,
}

// DeriveWithholdings applies the withholding decision table
// (MODULE-withholdings §3.2, WM §4.2): ISR and IVA are decided
// independently, first match wins within each, and the rates are the ones
// catalog holds on in.On.
//
// Rounding (ADR-0002 §4, ⚠️ WM A5): each amount is rounded once to the
// centavo, half away from zero: ISR = round(Subtotal × rate); IVA =
// round(IVA × Num/Den) exactly (2/3 never goes through a decimal rate), or
// round(Subtotal × 4 %) for freight, never above the IVA charged.
//
// Two deliberate readings of the table: a platform sale is decided before
// the receptor rows (the platform withholds whoever the buyer is, LISR
// 113-A), and a 0 %, exempt or residential-rent operation has no IVA to
// withhold before any IVA row (RuleIVANoTax).
//
// Invalid input is Violations (ClassValidation); a missing catalog or date,
// or an overflowing rate, is ErrInvalidParameter; a rate or threshold not in
// force on in.On is ErrRateNotValidOnDate.
func DeriveWithholdings(in WithholdingInput, catalog WithholdingCatalog) (Derivation, error) {
	if catalog == nil || in.On.IsZero() {
		return Derivation{}, fmt.Errorf("derive withholdings: no catalog or date: %w", ErrInvalidParameter)
	}
	if vs := in.violations(); len(vs) > 0 {
		return Derivation{}, vs
	}
	isr := in.isrRule()
	isrRate, isrCents, err := in.applyISR(isr, catalog)
	if err != nil {
		return Derivation{}, err
	}
	iva, err := in.ivaRuleAt(catalog)
	if err != nil {
		return Derivation{}, err
	}
	ivaCents, err := in.applyIVA(iva, catalog)
	if err != nil {
		return Derivation{}, err
	}
	net, err := in.net(isrCents, ivaCents)
	if err != nil {
		return Derivation{}, err
	}
	ambiguities := ordered(isr.ambiguities, iva.ambiguities)
	return Derivation{
		Withholdings: Withholdings{ISR: isrCents, IVA: ivaCents},
		ISRRate:      isrRate,
		IVABasis:     iva.basis,
		Withholder:   combine(isr.withholder, iva.withholder),
		Net:          net,
		RuleIDs:      []RuleID{isr.id, iva.id},
		Ambiguities:  ambiguities,
		Warnings:     in.warnings(ambiguities),
	}, nil
}

func (in WithholdingInput) violations() Violations {
	var vs Violations
	CheckEnum(&vs, "emisor_person_type", in.EmisorPersona)
	if in.EmisorPersona == PersonFisica && in.EmisorRegime == "" {
		vs.Add("emisor_regime", ErrRequired)
	}
	CheckEnum(&vs, "receptor_kind", in.Receptor)
	CheckEnum(&vs, "operation_kind", in.Operation)
	if in.fisicaOnPlatform() && !in.PlatformWithoutRFC {
		CheckEnum(&vs, "platform_activity", in.PlatformActivity)
	}
	checkCents(&vs, "subtotal_cents", in.Subtotal)
	checkCents(&vs, "iva_cents", in.IVA)
	CheckEnum(&vs, "iva_rate", in.IVARate)
	return vs
}

func checkCents(vs *Violations, field string, c money.Cents) {
	if c < 0 || c > MaxCents {
		vs.Add(field, ErrOutOfRange)
	}
}

func (in WithholdingInput) fisicaOnPlatform() bool {
	return in.Operation == OperationPlatform && in.EmisorPersona == PersonFisica
}

// chargesIVA reports an operation with IVA to withhold.
func (in WithholdingInput) chargesIVA() bool {
	residential := in.ResidentialRent && in.Operation == OperationRentRealEstate
	return in.IVARate.Taxable() && !residential
}

// isrRule is the first matching ISR row of WM §4.2.
func (in WithholdingInput) isrRule() isrRule {
	switch {
	case in.Operation == OperationPlatform:
		return in.platformISR()
	case in.Receptor.isPF():
		return isrRule{id: RuleISR0, withholder: WithholderNone}
	case in.Receptor == ReceptorForeign:
		return isrRule{id: RuleISR1, withholder: WithholderNone}
	case in.EmisorPersona == PersonMoral:
		return isrRule{id: RuleISR2, withholder: WithholderNone}
	case in.EmisorRegime == RegimeRESICO:
		return in.receptorISR(RuleISR4, KeyISRWithholdingRESICO)
	case in.EmisorRegime == RegimeBusiness && in.Operation == OperationProfessionalServices:
		return in.receptorISR(RuleISR5, KeyISRWithholdingFees)
	case in.Operation == OperationRentRealEstate && (in.EmisorRegime == RegimeRent || in.EmisorRegime == RegimeBusiness):
		return in.receptorISR(RuleISR6, KeyISRWithholdingRent)
	default:
		return isrRule{id: RuleISR7, withholder: WithholderNone, ambiguities: in.isr7Ambiguities()}
	}
}

// platformISR is ISR-2 (persona moral) or ISR-3 (persona física): the
// platform withholds (LISR 113-A, 113-C; LIF 2026 art. 25 fr. VI).
func (in WithholdingInput) platformISR() isrRule {
	if in.EmisorPersona == PersonMoral {
		key := KeyISRPlatformPersonaMoral
		if in.PlatformWithoutRFC {
			key = KeyISRPlatformPersonaMoralNoRFC
		}
		return isrRule{id: RuleISR2, key: key, withholder: WithholderPlatform}
	}
	key := platformRateKeys[in.PlatformActivity]
	if in.PlatformWithoutRFC {
		key = KeyISRPlatformNoRFC
	}
	return isrRule{id: RuleISR3, key: key, withholder: WithholderPlatform}
}

// receptorISR is a persona física billing a persona moral or a government
// (the only receptors left when it is called): the receptor withholds.
func (in WithholdingInput) receptorISR(id RuleID, key RateKey) isrRule {
	var ambiguities []Ambiguity
	if in.Receptor.isGov() {
		ambiguities = append(ambiguities, AmbiguityGovISR)
	}
	if id == RuleISR4 && in.Operation == OperationRentRealEstate {
		ambiguities = append(ambiguities, AmbiguityRESICORent)
	}
	return isrRule{id: id, key: key, withholder: WithholderReceptor, ambiguities: ambiguities}
}

// isr7Ambiguities flags the zero ISR rows the CPA has not confirmed for a
// persona física billing a persona moral or a government.
func (in WithholdingInput) isr7Ambiguities() []Ambiguity {
	var ambiguities []Ambiguity
	if in.EmisorRegime == RegimePlatforms {
		ambiguities = append(ambiguities, AmbiguityPlatformDirect)
	}
	switch in.Operation {
	case OperationCommission:
		ambiguities = append(ambiguities, AmbiguityCommissionISR)
	case OperationRentMovable:
		ambiguities = append(ambiguities, AmbiguityMovableRentISR)
	}
	return ambiguities
}

// ivaRule is the first matching IVA row of WM §4.2.
func (in WithholdingInput) ivaRule() ivaRule {
	none := func(id RuleID) ivaRule { return ivaRule{id: id, basis: IVABasisNone, withholder: WithholderNone} }
	switch {
	case !in.chargesIVA():
		return none(RuleIVANoTax)
	case in.fisicaOnPlatform():
		return in.platformIVA()
	case in.Receptor == ReceptorPFNoActivityOrPublic || in.Receptor == ReceptorForeign:
		return none(RuleIVA0)
	case in.Receptor == ReceptorPFWithActivity:
		return none(RuleIVA1)
	case in.Receptor == ReceptorGovStateMunicipal:
		return none(RuleIVA2)
	case in.Operation == OperationGroundFreight:
		return ivaRule{id: RuleIVA4, basis: IVABasisFourPercentOfPrice, withholder: WithholderReceptor}
	case in.Operation == OperationScrap && in.Receptor == ReceptorPM:
		return ivaRule{id: RuleIVA5, basis: IVABasisAllOfIVA, withholder: WithholderReceptor}
	case in.EmisorPersona == PersonMoral:
		return none(RuleIVA8)
	case in.Receptor == ReceptorPM && in.Operation.personal():
		return in.twoThirds(RuleIVA6)
	case in.Receptor == ReceptorGovFederal && in.Operation != OperationOther:
		rule := in.twoThirds(RuleIVA7, AmbiguityFederalIVAShare)
		rule.aboveFederalThreshold = in.Operation != OperationProfessionalServices
		return rule
	default:
		return none(RuleIVA8)
	}
}

// platformIVA is IVA-3: 50 % of the IVA, 100 % without RFC (LIVA 18-J
// fr. II a).
func (in WithholdingInput) platformIVA() ivaRule {
	basis := IVABasisHalfOfIVA
	if in.PlatformWithoutRFC {
		basis = IVABasisAllOfIVA
	}
	return ivaRule{id: RuleIVA3, basis: basis, withholder: WithholderPlatform}
}

// twoThirds is a 2/3 row; on the border 8 % it is 2/3 of the IVA charged
// (⚠️ WM A8).
func (in WithholdingInput) twoThirds(id RuleID, ambiguities ...Ambiguity) ivaRule {
	if in.IVARate == IVARate8 {
		ambiguities = append(ambiguities, AmbiguityBorderShare)
	}
	return ivaRule{id: id, basis: IVABasisTwoThirds, withholder: WithholderReceptor, ambiguities: ambiguities}
}

// ivaRuleAt resolves IVA-7's threshold on in.On: at or below it, IVA-8.
func (in WithholdingInput) ivaRuleAt(catalog WithholdingCatalog) (ivaRule, error) {
	rule := in.ivaRule()
	if !rule.aboveFederalThreshold {
		return rule, nil
	}
	threshold, ok := catalog.Amount(KeyIVAFederalWithholdingThreshold, in.On)
	if !ok {
		return ivaRule{}, fmt.Errorf("%s on %s: %w", KeyIVAFederalWithholdingThreshold, in.On, ErrRateNotValidOnDate)
	}
	if in.Subtotal <= threshold {
		return ivaRule{id: RuleIVA8, basis: IVABasisNone, withholder: WithholderNone}, nil
	}
	return rule, nil
}

func (in WithholdingInput) applyISR(rule isrRule, catalog RateCatalog) (money.Rate, money.Cents, error) {
	if rule.key == "" {
		return 0, 0, nil
	}
	r, err := rateOn(rule.key, in.On, catalog)
	if err != nil {
		return 0, 0, err
	}
	c, err := ApplyRate(in.Subtotal, r)
	return r, c, err
}

func (in WithholdingInput) applyIVA(rule ivaRule, catalog RateCatalog) (money.Cents, error) {
	if share, ok := ivaShares[rule.basis]; ok {
		s, err := share.Suggest(in.IVA)
		return s.Cents, err
	}
	if rule.basis != IVABasisFourPercentOfPrice {
		return 0, nil
	}
	r, err := rateOn(KeyIVAWithholdingFreight, in.On, catalog)
	if err != nil {
		return 0, err
	}
	c, err := ApplyRate(in.Subtotal, r)
	return min(c, in.IVA), err
}

func rateOn(key RateKey, on Date, catalog RateCatalog) (money.Rate, error) {
	r, ok := catalog.Rate(key, on)
	if !ok {
		return 0, fmt.Errorf("%s on %s: %w", key, on, ErrRateNotValidOnDate)
	}
	return r, nil
}

// net is Subtotal + IVA − ISR − IVA withheld; inputs within MaxCents never
// overflow.
func (in WithholdingInput) net(isr, iva money.Cents) (money.Cents, error) {
	gross, err := in.Subtotal.Add(in.IVA)
	if err == nil {
		gross, err = gross.Sub(isr)
	}
	if err == nil {
		gross, err = gross.Sub(iva)
	}
	if err != nil {
		return 0, fmt.Errorf("withholding net: %w: %w", ErrInvalidParameter, err)
	}
	return gross, nil
}

func (in WithholdingInput) warnings(ambiguities []Ambiguity) []WarningCode {
	var warnings []WarningCode
	if len(ambiguities) > 0 {
		warnings = append(warnings, WarningCPAUnverified)
	}
	if in.Receptor == ReceptorForeign {
		warnings = append(warnings, WarningForeignCounterparty)
	}
	return warnings
}

// combine is the record's withholder: the platform over the receptor over
// no one.
func combine(isr, iva Withholder) Withholder {
	for _, w := range []Withholder{WithholderPlatform, WithholderReceptor} {
		if isr == w || iva == w {
			return w
		}
	}
	return WithholderNone
}

// ordered merges ambiguity lists, once each, in WM §6 order; nil when empty.
func ordered(lists ...[]Ambiguity) []Ambiguity {
	all := slices.Concat(lists...)
	var out []Ambiguity
	for _, a := range ambiguityOrder {
		if slices.Contains(all, a) {
			out = append(out, a)
		}
	}
	return out
}
