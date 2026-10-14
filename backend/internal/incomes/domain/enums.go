package domain

import (
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// Counterparty is who paid the income (contract CounterpartyType).
type Counterparty string

// The contract's CounterpartyType enum.
const (
	CounterpartyFisica   Counterparty = "persona_fisica"
	CounterpartyMoral    Counterparty = "persona_moral"
	CounterpartyPublic   Counterparty = "publico_general"
	CounterpartyPlatform Counterparty = "plataforma"
)

// Valid reports a contract CounterpartyType.
func (c Counterparty) Valid() bool {
	return slices.Contains([]Counterparty{CounterpartyFisica, CounterpartyMoral, CounterpartyPublic, CounterpartyPlatform}, c)
}

// PlatformActivity is a 625 activity (contract PlatformActivity).
type PlatformActivity string

// The 625 activities of LISR 113-A.
const (
	ActivityTransportDelivery PlatformActivity = "transport_delivery"
	ActivityLodging           PlatformActivity = "lodging"
	ActivityGoodsServices     PlatformActivity = "goods_services"
)

// Valid reports a contract PlatformActivity.
func (a PlatformActivity) Valid() bool {
	return slices.Contains([]PlatformActivity{ActivityTransportDelivery, ActivityLodging, ActivityGoodsServices}, a)
}

// IncomeBasis is the regime's ISR income basis, catalog data
// (regimes.income_basis), never a switch on the regime code.
type IncomeBasis string

// The income bases.
const (
	BasisCash    IncomeBasis = "cash"
	BasisAccrual IncomeBasis = "accrual"
)

// Valid reports a known IncomeBasis.
func (b IncomeBasis) Valid() bool { return b == BasisCash || b == BasisAccrual }

// ParseCounterparty returns the Counterparty spelled exactly raw, else
// fiscal.ErrInvalid.
func ParseCounterparty(raw string) (Counterparty, error) { return fiscal.ParseEnum[Counterparty](raw) }

// ParsePlatformActivity returns the PlatformActivity spelled exactly raw,
// else fiscal.ErrInvalid.
func ParsePlatformActivity(raw string) (PlatformActivity, error) {
	return fiscal.ParseEnum[PlatformActivity](raw)
}

// ParseIncomeBasis returns the IncomeBasis spelled exactly raw, else
// fiscal.ErrInvalid.
func ParseIncomeBasis(raw string) (IncomeBasis, error) { return fiscal.ParseEnum[IncomeBasis](raw) }
