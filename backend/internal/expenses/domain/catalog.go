package domain

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// RateCatalog is the expenses module's consumer-side port for the tax
// engine's dated parameters (ADR-0002 §3): the kernel's rates plus the money
// parameters only expenses reads. M5-08 implements it. Values are resolved
// by the expense's paid_on, never by fiscal year alone.
type RateCatalog interface {
	fiscal.RateCatalog
	// Amount returns the money parameter key holds on day on, and false
	// when no value is in force that day.
	Amount(key AmountKey, on fiscal.Date) (money.Cents, bool)
}

// KeyIVAWithholdingFreight is the IVA a persona moral withholds on freight:
// 4 % of the price, RLIVA 3 fr. II (research B.2.3). The fees and rent rates
// are the kernel's.
const KeyIVAWithholdingFreight fiscal.RateKey = "iva.withholding.freight_pm"

// AmountKey names a dated money parameter.
type AmountKey string

// KeyCashPaymentLimit is the cash payment above which an expense is not
// deductible: $2,000.00, LISR 27 fr. III (⚠️ CPA, MODULE-expenses OQ-E4).
const KeyCashPaymentLimit AmountKey = "deduction.cash_payment_limit"
