package fiscal

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// RateCatalog is the port for the tax engine's dated rate parameters
// (ADR-0002 §3); M5-08 implements it over the versioned params. A value is
// resolved by the record's date, never by the fiscal year alone. A module
// that needs more (expenses' money parameters) embeds it in its own
// consumer-side port.
type RateCatalog interface {
	// Rate returns the rate key holds on day on, and false when no value
	// is in force that day (e.g. the border 8 % after its decree lapses).
	Rate(key RateKey, on Date) (money.Rate, bool)
}

// RateKey names a dated rate parameter, spelled as the engine's params
// files spell it (ADR-0002 §3). The kernel names the keys more than one
// bounded context reads; a module declares its own keys as RateKey values.
type RateKey string

// IVA rates (LIVA 1, 2-A; border decrees).
const (
	KeyIVAGeneral RateKey = "iva.rate.general_16"
	KeyIVABorder  RateKey = "iva.rate.border_8"
	KeyIVAZero    RateKey = "iva.rate.zero"
)

// ISR withholding rates a persona moral applies to a persona física, read
// by incomes (withheld from the taxpayer) and expenses (withheld by the
// taxpayer) alike (research B.2.3).
const (
	KeyISRWithholdingFees RateKey = "isr.withholding.fees_pm" // 10 %, LISR 106
	KeyISRWithholdingRent RateKey = "isr.withholding.rent_pm" // 10 %, LISR 116
)

// ApplyRate returns base × r rounded once to the centavo, half away from
// zero. Only an absurd catalog rate can overflow, a parameter fault.
func ApplyRate(base money.Cents, r money.Rate) (money.Cents, error) {
	e, err := base.MulRate(r)
	if err != nil {
		return 0, fmt.Errorf("apply rate %s: %w: %w", r, ErrInvalidParameter, err)
	}
	return e.Round(), nil
}
