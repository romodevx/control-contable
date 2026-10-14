package domain

import (
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// Legend is the mandatory notice of the expense form and list (MODULE-expenses
// §2.7, owner rule). The app renders this server value, never its own copy.
const Legend = "Esto es un aproximado. No todos los gastos aplican para tu giro comercial."

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

// Valid reports a contract PaymentMethod.
func (m PaymentMethod) Valid() bool {
	return slices.Contains([]PaymentMethod{
		PaymentCash, PaymentTransfer, PaymentCardDebit, PaymentCardCredit, PaymentCheck, PaymentEWallet, PaymentOther,
	}, m)
}

// Deductibility is the hint shown for a category (category_regime_rules).
type Deductibility string

// The hints of MODULE-expenses §2.2.
const (
	Deductible          Deductibility = "deductible"
	PartiallyDeductible Deductibility = "partially_deductible"
	NotDeductible       Deductibility = "not_deductible"
	ISRNotApplicable    Deductibility = "isr_not_applicable"
)

// Valid reports a known Deductibility.
func (d Deductibility) Valid() bool {
	return slices.Contains([]Deductibility{Deductible, PartiallyDeductible, NotDeductible, ISRNotApplicable}, d)
}

// DeductionMode is the 606 option deduction_mode.
type DeductionMode string

// The 606 deduction modes (LISR 115).
const (
	DeductionActual  DeductionMode = "actual"
	DeductionBlind35 DeductionMode = "blind_35"
)

// Valid reports a known DeductionMode.
func (m DeductionMode) Valid() bool { return m == DeductionActual || m == DeductionBlind35 }

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

// Valid reports a known Override.
func (o Override) Valid() bool { return o >= UseDefault && o <= ClaimNotDeductible }

// ParsePaymentMethod returns the PaymentMethod spelled exactly raw, else
// fiscal.ErrInvalid.
func ParsePaymentMethod(raw string) (PaymentMethod, error) {
	return fiscal.ParseEnum[PaymentMethod](raw)
}

// ParseDeductibility returns the Deductibility spelled exactly raw, else
// fiscal.ErrInvalid.
func ParseDeductibility(raw string) (Deductibility, error) {
	return fiscal.ParseEnum[Deductibility](raw)
}

// ParseDeductionMode returns the DeductionMode spelled exactly raw, else
// fiscal.ErrInvalid.
func ParseDeductionMode(raw string) (DeductionMode, error) {
	return fiscal.ParseEnum[DeductionMode](raw)
}
