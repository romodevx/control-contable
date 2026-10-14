// Package taxengine names the regime strategies the engine ships (ADR-0002
// §1). The composition root (cmd/api) builds the immutable registry from
// Strategies, and the golden gate runs the same list, so a strategy is
// either wired and gated or neither.
package taxengine

import "github.com/romodevx/control-contable/backend/internal/taxengine/domain"

// Strategies returns one strategy per (regime, person type) with a
// computation; every other catalog regime falls back to the registry's
// InformationalStrategy (605, 607, 608, 610, 611, 614, 615, 616, 620, 623,
// 624). Adding a regime is one line here (US-X3).
func Strategies() []domain.Strategy {
	return []domain.Strategy{
		domain.NewGeneralPMStrategy(), // 601 PM (M5-15)
		domain.NewStrategy603PM(),     // 603 PM, partial: IVA only (M5-18)
		domain.NewRentPFStrategy(),    // 606 PF (M5-12)
		domain.NewStrategy612PF(),     // 612 PF (M5-11)
		domain.NewStrategy621PF(),     // 621 PF, partial: bimonthly RIF (M5-17)
		domain.NewStrategy622PM(),     // 622 PM, partial: IVA only (M5-18)
		// 625 PF (M5-14, PR #94) registers here when it lands; until then
		// 625 is answered by the informational fallback.
		domain.NewRESICOPFStrategy(), // 626 PF (M5-13)
		domain.NewRESICOPMStrategy(), // 626 PM (M5-16)
	}
}
