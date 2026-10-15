// Package taxengine names the regime strategies the engine ships (ADR-0002
// §1). The composition root (cmd/api) builds the immutable registry from
// Strategies, and the golden gate runs the same list, so a strategy is
// either wired and gated or neither.
package taxengine

import (
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual/pm"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

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
		domain.NewStrategy625PF(),     // 625 PF (M5-14)
		domain.NewRESICOPFStrategy(),  // 626 PF (M5-13)
		domain.NewRESICOPMStrategy(),  // 626 PM (M5-16)
	}
}

// AnnualSections returns one strategy per computed annual section kind
// (ADR-0009, MODULE-annual §2.2): the composition root builds the annual
// registry from it and the annual golden gate runs the same list. Adding a
// section is one line here.
func AnnualSections(limits annual.LimitsReader) []usecaseannual.SectionStrategy {
	return []usecaseannual.SectionStrategy{
		usecaseannual.NewPFArt152Section(annual.NewPFArt152(annual.NewCalculator(limits))), // ANN-08
		usecaseannual.NewPFRESICOCheckSection(annual.NewPFRESICOCheck()),                   // ANN-09
		usecaseannual.NewPM601Section(pm.NewGeneral()),                                     // ANN-10
		usecaseannual.NewPMRESICOSection(pm.NewRESICO()),                                   // ANN-11
	}
}
