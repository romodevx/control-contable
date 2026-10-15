package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// TaxpayerProfileReader is what the estimate reads about a taxpayer from
// the taxpayers module (M2, its TaxpayerProfileReader): cmd/api adapts it.
// It runs nested in the estimate's unit of work, so the composition root
// must give both the same tx.Manager. Errors are *apperr.Error (404
// not_found for an unknown or another user's taxpayer).
type TaxpayerProfileReader interface {
	Profile(ctx context.Context, userID, taxpayerID string) (TaxpayerProfile, error)
}

// TaxpayerProfile is a taxpayer as the estimate sees it. An archived
// taxpayer is still estimated (MODULE-taxpayers §2.7).
type TaxpayerProfile struct {
	PersonType fiscal.PersonType
	Archived   bool
	Regimes    []RegimeHold
}

// RegimeHold is one regime the taxpayer holds or held: the catalog's
// treatment, its options (defaults included) and its period window.
type RegimeHold struct {
	Code      fiscal.RegimeCode
	Treatment domain.Treatment
	Options   domain.Options
	Start     domain.Period
	// End is the zero Period while the regime is open.
	End domain.Period
}

// RegimeTextReader reads the catalog texts of the regimes a taxpayer holds
// or held (the global regimes catalog, its row for the taxpayer's person
// type), in the estimate's unit of work. A regime the catalog lacks is
// absent from the map.
type RegimeTextReader interface {
	RegimeTexts(ctx context.Context, taxpayerID string) (map[fiscal.RegimeCode]RegimeText, error)
}

// RegimeText is what the catalog says about one regime, in es-MX: what a
// response shows next to the engine's figures.
type RegimeText struct {
	// Name is the regime's name as the SAT catalog spells it.
	Name string
	// Explanation is regimes.explanation, the text of the label key
	// domain.RegimeExplanation; "" when the catalog has none.
	Explanation string
	// InputLabels are the TaxInputDef labels of the regime's tax inputs.
	InputLabels map[domain.InputKey]string
}

// ParamsCatalog selects the parameter set of a fiscal year
// (*taxparams.Catalog, embedded and validated at boot).
type ParamsCatalog interface {
	Select(fiscalYear int) (taxparams.Selection, error)
}

// HolidayCalendar reads the holidays that move a due date (the M9-03
// holidays table) as the kernel's HolidaySet, inside the caller's unit of
// work. The due-date rule itself is fiscal.MonthlyDueDate and
// fiscal.NextBusinessDay (ADR-0007 Amendment 1).
type HolidayCalendar interface {
	Holidays(ctx context.Context) (fiscal.HolidaySet, error)
}
