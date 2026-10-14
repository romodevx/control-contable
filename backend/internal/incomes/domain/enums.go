package domain

import "slices"

// PersonType is the payee's kind of taxpayer (contract PersonType). It is the
// incomes module's own copy: modules never import each other.
type PersonType string

// The contract's PersonType enum.
const (
	PersonFisica PersonType = "fisica"
	PersonMoral  PersonType = "moral"
)

// RegimeCode is a SAT c_RegimenFiscal code such as "612". Which codes exist
// is catalog data the taxpayers module checks; this module names only the
// regimes whose withholding rules differ (MODULE-incomes §2.6).
type RegimeCode string

// The regimes with their own income rules.
const (
	RegimeRent      RegimeCode = "606" // Arrendamiento (PF)
	RegimeBusiness  RegimeCode = "612" // Actividades empresariales y profesionales (PF)
	RegimePlatforms RegimeCode = "625" // Plataformas tecnológicas (PF)
	RegimeRESICO    RegimeCode = "626" // RESICO: PF arts. 113-E..J, PM arts. 206-215
)

// IVARate is the IVA treatment of an income (contract IvaRate).
type IVARate string

// The contract's IvaRate enum.
const (
	IVARate16 IVARate = "rate_16"
	IVARate8  IVARate = "rate_8" // border regions, while the decree is in force
	IVARate0  IVARate = "rate_0"
	IVAExempt IVARate = "exempt"
)

// Taxable reports whether the treatment charges a non-zero IVA.
func (r IVARate) Taxable() bool { return r == IVARate16 || r == IVARate8 }

// Counterparty is who paid the income (contract CounterpartyType).
type Counterparty string

// The contract's CounterpartyType enum.
const (
	CounterpartyFisica   Counterparty = "persona_fisica"
	CounterpartyMoral    Counterparty = "persona_moral"
	CounterpartyPublic   Counterparty = "publico_general"
	CounterpartyPlatform Counterparty = "plataforma"
)

// PlatformActivity is a 625 activity (contract PlatformActivity).
type PlatformActivity string

// The 625 activities of LISR 113-A.
const (
	ActivityTransportDelivery PlatformActivity = "transport_delivery"
	ActivityLodging           PlatformActivity = "lodging"
	ActivityGoodsServices     PlatformActivity = "goods_services"
)

// IncomeBasis is the regime's ISR income basis, catalog data
// (regimes.income_basis), never a switch on the regime code.
type IncomeBasis string

// The income bases.
const (
	BasisCash    IncomeBasis = "cash"
	BasisAccrual IncomeBasis = "accrual"
)

var (
	personTypes    = []PersonType{PersonFisica, PersonMoral}
	ivaRates       = []IVARate{IVARate16, IVARate8, IVARate0, IVAExempt}
	counterparties = []Counterparty{CounterpartyFisica, CounterpartyMoral, CounterpartyPublic, CounterpartyPlatform}
	activities     = []PlatformActivity{ActivityTransportDelivery, ActivityLodging, ActivityGoodsServices}
	bases          = []IncomeBasis{BasisCash, BasisAccrual}
)

// ParsePersonType returns the PersonType spelled exactly raw, else ErrInvalid.
func ParsePersonType(raw string) (PersonType, error) { return parseEnum(raw, personTypes) }

// ParseIVARate returns the IVARate spelled exactly raw, else ErrInvalid.
func ParseIVARate(raw string) (IVARate, error) { return parseEnum(raw, ivaRates) }

// ParseCounterparty returns the Counterparty spelled exactly raw, else ErrInvalid.
func ParseCounterparty(raw string) (Counterparty, error) { return parseEnum(raw, counterparties) }

// ParsePlatformActivity returns the PlatformActivity spelled exactly raw,
// else ErrInvalid.
func ParsePlatformActivity(raw string) (PlatformActivity, error) { return parseEnum(raw, activities) }

// ParseIncomeBasis returns the IncomeBasis spelled exactly raw, else ErrInvalid.
func ParseIncomeBasis(raw string) (IncomeBasis, error) { return parseEnum(raw, bases) }

// parseEnum neither trims nor folds case: the contract enums are exact.
func parseEnum[T ~string](raw string, values []T) (T, error) {
	v := T(raw)
	if !slices.Contains(values, v) {
		return "", ErrInvalid
	}
	return v, nil
}

// checkEnum records field's violation for a zero (required) or unknown
// (invalid) enum value.
func checkEnum[T ~string](vs *Violations, field string, v T, values []T) {
	switch {
	case v == "":
		vs.add(field, ErrRequired)
	case !slices.Contains(values, v):
		vs.add(field, ErrInvalid)
	}
}
