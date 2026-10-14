package domain

import "github.com/romodevx/control-contable/backend/internal/platform/money"

// RateCatalog is the consumer-side port for the tax engine's dated
// parameters (ADR-0002 §3); M5-08 implements it over the versioned params.
// Values are resolved by the expense's paid_on, never by fiscal year alone.
type RateCatalog interface {
	// Rate returns the rate key holds on day on, and false when no value
	// is in force that day.
	Rate(key RateKey, on Date) (money.Rate, bool)
	// Amount returns the money parameter key holds on day on, and false
	// when no value is in force that day.
	Amount(key AmountKey, on Date) (money.Cents, bool)
}

// RateKey names a dated rate parameter, spelled as the engine's params
// files spell it (ADR-0002 §3).
type RateKey string

// IVA rates (LIVA 1, 2-A; border decrees).
const (
	KeyIVAGeneral RateKey = "iva.rate.general_16"
	KeyIVABorder  RateKey = "iva.rate.border_8"
	KeyIVAZero    RateKey = "iva.rate.zero"
)

// Withholding rates a persona moral applies to its suppliers (research
// B.2.3).
const (
	KeyISRWithholdingFees    RateKey = "isr.withholding.fees_pm"    // 10 %, LISR 106
	KeyISRWithholdingRent    RateKey = "isr.withholding.rent_pm"    // 10 %, LISR 116
	KeyIVAWithholdingFreight RateKey = "iva.withholding.freight_pm" // 4 % of the price, RLIVA 3 fr. II
)

// AmountKey names a dated money parameter.
type AmountKey string

// KeyCashPaymentLimit is the cash payment above which an expense is not
// deductible: $2,000.00, LISR 27 fr. III (⚠️ CPA, MODULE-expenses OQ-E4).
const KeyCashPaymentLimit AmountKey = "deduction.cash_payment_limit"
