package domain

import "github.com/romodevx/control-contable/backend/internal/platform/money"

// RateCatalog is the consumer-side port for the tax engine's dated
// parameters (ADR-0002 §3); M5-08 implements it over the versioned params.
// A value is resolved by the record's date, never by the fiscal year alone.
type RateCatalog interface {
	// Rate returns the rate key holds on day on, and false when no value
	// is in force that day (e.g. the border 8 % after its decree lapses).
	Rate(key RateKey, on Date) (money.Rate, bool)
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

// ISR withholding rates a payer applies (research A.2.3-A.2.6, B.2.3).
const (
	KeyISRWithholdingFees      RateKey = "isr.withholding.fees_pm"   // 10 %, LISR 106
	KeyISRWithholdingRent      RateKey = "isr.withholding.rent_pm"   // 10 %, LISR 116
	KeyISRWithholdingRESICO    RateKey = "isr.withholding.resico_pf" // 1.25 %, LISR 113-J
	KeyISRPlatformTransport    RateKey = "isr.platform.transport_delivery"
	KeyISRPlatformLodging      RateKey = "isr.platform.lodging"
	KeyISRPlatformGoods        RateKey = "isr.platform.goods_services" // 2.5 % in 2026, LIF art. 25-VI
	KeyISRPlatformNoRFC        RateKey = "isr.platform.no_rfc"         // 20 %, LISR 113-C fr. IV
	KeyISRPlatformPersonaMoral RateKey = "isr.platform.persona_moral"  // 2.5 % in 2026
)
