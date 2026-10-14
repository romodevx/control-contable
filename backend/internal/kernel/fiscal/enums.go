package fiscal

import "slices"

// Enum is a contract string enum: a module's own enum type joins the
// kernel's helpers (ParseEnum, CheckEnum) by declaring Valid.
type Enum interface {
	~string
	Valid() bool
}

// ParseEnum returns the T spelled exactly raw, else ErrInvalid. It neither
// trims nor folds case: the contract enums are exact.
func ParseEnum[T Enum](raw string) (T, error) {
	v := T(raw)
	if !v.Valid() {
		return "", ErrInvalid
	}
	return v, nil
}

// CheckEnum records field's violation for a zero (ErrRequired) or unknown
// (ErrInvalid) enum value.
func CheckEnum[T Enum](vs *Violations, field string, v T) {
	switch {
	case v == "":
		vs.Add(field, ErrRequired)
	case !v.Valid():
		vs.Add(field, ErrInvalid)
	}
}

// PersonType is a taxpayer's kind (contract PersonType).
type PersonType string

// The contract's PersonType enum.
const (
	PersonFisica PersonType = "fisica"
	PersonMoral  PersonType = "moral"
)

// Valid reports a contract PersonType.
func (p PersonType) Valid() bool { return p == PersonFisica || p == PersonMoral }

// ParsePersonType returns the PersonType spelled exactly raw, else ErrInvalid.
func ParsePersonType(raw string) (PersonType, error) { return ParseEnum[PersonType](raw) }

// RegimeCode is a SAT c_RegimenFiscal code such as "612". Which codes exist
// is catalog data the taxpayers module checks; the kernel names only the
// regimes whose rules some module spells out.
type RegimeCode string

// The regimes with their own rules in a module.
const (
	RegimeRent      RegimeCode = "606" // Arrendamiento (PF)
	RegimeBusiness  RegimeCode = "612" // Actividades empresariales y profesionales (PF)
	RegimePlatforms RegimeCode = "625" // Plataformas tecnológicas (PF)
	RegimeRESICO    RegimeCode = "626" // RESICO: PF arts. 113-E..J, PM arts. 206-215
)

// IVARate is the IVA treatment of a record (contract IvaRate).
type IVARate string

// The contract's IvaRate enum.
const (
	IVARate16 IVARate = "rate_16"
	IVARate8  IVARate = "rate_8" // border regions, while the decree is in force
	IVARate0  IVARate = "rate_0"
	IVAExempt IVARate = "exempt"
)

// Valid reports a contract IvaRate.
func (r IVARate) Valid() bool {
	return slices.Contains([]IVARate{IVARate16, IVARate8, IVARate0, IVAExempt}, r)
}

// Taxable reports whether the treatment charges a non-zero IVA.
func (r IVARate) Taxable() bool { return r == IVARate16 || r == IVARate8 }

// ParseIVARate returns the IVARate spelled exactly raw, else ErrInvalid.
func ParseIVARate(raw string) (IVARate, error) { return ParseEnum[IVARate](raw) }
