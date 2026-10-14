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

// ParamsCatalog selects the parameter set of a fiscal year
// (*taxparams.Catalog, embedded and validated at boot).
type ParamsCatalog interface {
	Select(fiscalYear int) (taxparams.Selection, error)
}

// HolidayCalendar answers whether a date is an official non-business day
// besides weekends (P0/M9 holidays table), to move the due date (CFF 12 ⚠️).
type HolidayCalendar interface {
	IsHoliday(ctx context.Context, d fiscal.Date) (bool, error)
}
