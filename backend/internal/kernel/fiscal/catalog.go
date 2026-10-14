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

// The withholding rates DeriveWithholdings reads besides the fees and rent
// ones (MODULE-withholdings §3.2; params/data, WH-05).
const (
	KeyISRWithholdingRESICO         RateKey = "isr.withholding.resico_pf"         // 1.25 %, LISR 113-J
	KeyIVAWithholdingFreight        RateKey = "iva.withholding.freight_pm"        // 4 % of the price, RLIVA 3 fr. II
	KeyISRPlatformTransport         RateKey = "isr.platform.transport_delivery"   // 2.1 %, LISR 113-A fr. I
	KeyISRPlatformLodging           RateKey = "isr.platform.lodging"              // 4 %, LISR 113-A fr. II
	KeyISRPlatformGoods             RateKey = "isr.platform.goods_services"       // 2.5 % in 2026, LIF art. 25 fr. VI
	KeyISRPlatformNoRFC             RateKey = "isr.platform.no_rfc"               // 20 %, LISR 113-C fr. IV
	KeyISRPlatformPersonaMoral      RateKey = "isr.platform.persona_moral"        // 2.5 % in 2026, LIF art. 25 fr. VI
	KeyISRPlatformPersonaMoralNoRFC RateKey = "isr.platform.persona_moral_no_rfc" // 20 % in 2026, LIF art. 25 fr. VI
)

// WithholdingCatalog is DeriveWithholdings' consumer-side port: the dated
// rates plus the one dated money threshold it reads. WH-05 provides the
// values over the versioned params (ADR-0002 §3).
type WithholdingCatalog interface {
	RateCatalog
	// Amount returns the money parameter key holds on day on, and false
	// when no value is in force that day.
	Amount(key AmountKey, on Date) (money.Cents, bool)
}

// AmountKey names a dated money parameter, spelled as the params files
// spell it.
type AmountKey string

// KeyIVAFederalWithholdingThreshold is the payment above which the
// Federación withholds IVA from a persona física on an operation other than
// professional services: $2,000.00, RMF 4.1.3 (⚠️ WM A2).
const KeyIVAFederalWithholdingThreshold AmountKey = "iva.withholding.federal_threshold"

// ApplyRate returns base × r rounded once to the centavo, half away from
// zero. Only an absurd catalog rate can overflow, a parameter fault.
func ApplyRate(base money.Cents, r money.Rate) (money.Cents, error) {
	e, err := base.MulRate(r)
	if err != nil {
		return 0, fmt.Errorf("apply rate %s: %w: %w", r, ErrInvalidParameter, err)
	}
	return e.Round(), nil
}
