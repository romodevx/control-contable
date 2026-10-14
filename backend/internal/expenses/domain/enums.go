package domain

import "slices"

// Legend is the mandatory notice of the expense form and list (MODULE-expenses
// §2.7, owner rule). The app renders this server value, never its own copy.
const Legend = "Esto es un aproximado. No todos los gastos aplican para tu giro comercial."

// PersonType is the paying taxpayer's kind (contract PersonType), the
// expenses module's own copy: modules never import each other.
type PersonType string

// The contract's PersonType enum.
const (
	PersonFisica PersonType = "fisica"
	PersonMoral  PersonType = "moral"
)

// RegimeCode is a SAT c_RegimenFiscal code such as "612".
type RegimeCode string

// The regimes whose options change a category's hint (§2.3).
const (
	RegimeRent      RegimeCode = "606"
	RegimePlatforms RegimeCode = "625"
	RegimeRESICO    RegimeCode = "626"
)

// IVARate is the IVA treatment of an expense (contract IvaRate).
type IVARate string

// The contract's IvaRate enum.
const (
	IVARate16 IVARate = "rate_16"
	IVARate8  IVARate = "rate_8"
	IVARate0  IVARate = "rate_0"
	IVAExempt IVARate = "exempt"
)

// Taxable reports whether the treatment charges a non-zero IVA.
func (r IVARate) Taxable() bool { return r == IVARate16 || r == IVARate8 }

// PaymentMethod is how the expense was paid (contract PaymentMethod).
type PaymentMethod string

// The contract's PaymentMethod enum.
const (
	PaymentCash       PaymentMethod = "cash"
	PaymentTransfer   PaymentMethod = "transfer"
	PaymentCardDebit  PaymentMethod = "card_debit"
	PaymentCardCredit PaymentMethod = "card_credit"
	PaymentCheck      PaymentMethod = "check"
	PaymentEWallet    PaymentMethod = "e_wallet"
	PaymentOther      PaymentMethod = "other"
)

// Deductibility is the hint shown for a category (category_regime_rules).
type Deductibility string

// The hints of MODULE-expenses §2.2.
const (
	Deductible          Deductibility = "deductible"
	PartiallyDeductible Deductibility = "partially_deductible"
	NotDeductible       Deductibility = "not_deductible"
	ISRNotApplicable    Deductibility = "isr_not_applicable"
)

// DeductionMode is the 606 option deduction_mode.
type DeductionMode string

// The 606 deduction modes (LISR 115).
const (
	DeductionActual  DeductionMode = "actual"
	DeductionBlind35 DeductionMode = "blind_35"
)

// DeductibleSource says where the effective deductible came from.
type DeductibleSource string

// The contract's deductible_source enum.
const (
	SourceCategoryDefault DeductibleSource = "category_default"
	SourceUser            DeductibleSource = "user"
)

// Override is the user's deductible switch: a boolean or "use the default"
// (contract deductible: boolean | null).
type Override int8

// The override states. The zero value uses the default.
const (
	UseDefault Override = iota
	ClaimDeductible
	ClaimNotDeductible
)

var (
	personTypes    = []PersonType{PersonFisica, PersonMoral}
	ivaRates       = []IVARate{IVARate16, IVARate8, IVARate0, IVAExempt}
	paymentMethods = []PaymentMethod{
		PaymentCash, PaymentTransfer, PaymentCardDebit, PaymentCardCredit, PaymentCheck, PaymentEWallet, PaymentOther,
	}
	deductibilities = []Deductibility{Deductible, PartiallyDeductible, NotDeductible, ISRNotApplicable}
	deductionModes  = []DeductionMode{DeductionActual, DeductionBlind35}
	overrides       = []Override{UseDefault, ClaimDeductible, ClaimNotDeductible}
)

// ParsePersonType returns the PersonType spelled exactly raw, else ErrInvalid.
func ParsePersonType(raw string) (PersonType, error) { return parseEnum(raw, personTypes) }

// ParseIVARate returns the IVARate spelled exactly raw, else ErrInvalid.
func ParseIVARate(raw string) (IVARate, error) { return parseEnum(raw, ivaRates) }

// ParsePaymentMethod returns the PaymentMethod spelled exactly raw, else
// ErrInvalid.
func ParsePaymentMethod(raw string) (PaymentMethod, error) { return parseEnum(raw, paymentMethods) }

// ParseDeductibility returns the Deductibility spelled exactly raw, else
// ErrInvalid.
func ParseDeductibility(raw string) (Deductibility, error) { return parseEnum(raw, deductibilities) }

// ParseDeductionMode returns the DeductionMode spelled exactly raw, else
// ErrInvalid.
func ParseDeductionMode(raw string) (DeductionMode, error) { return parseEnum(raw, deductionModes) }

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
