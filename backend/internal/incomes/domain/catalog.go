package domain

import "github.com/romodevx/control-contable/backend/internal/kernel/fiscal"

// The ISR withholding rates only incomes reads (research A.2.5, A.2.6,
// B.2.3); the IVA rates and the PM fees/rent rates are the kernel's. Every
// value comes from fiscal.RateCatalog (ADR-0002 §3), resolved by the
// record's date.
const (
	KeyISRWithholdingRESICO    fiscal.RateKey = "isr.withholding.resico_pf" // 1.25 %, LISR 113-J
	KeyISRPlatformTransport    fiscal.RateKey = "isr.platform.transport_delivery"
	KeyISRPlatformLodging      fiscal.RateKey = "isr.platform.lodging"
	KeyISRPlatformGoods        fiscal.RateKey = "isr.platform.goods_services" // 2.5 % in 2026, LIF art. 25-VI
	KeyISRPlatformNoRFC        fiscal.RateKey = "isr.platform.no_rfc"         // 20 %, LISR 113-C fr. IV
	KeyISRPlatformPersonaMoral fiscal.RateKey = "isr.platform.persona_moral"  // 2.5 % in 2026
)
