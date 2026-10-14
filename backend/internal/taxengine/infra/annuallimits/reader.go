// Package annuallimits adapts the published parameter sets to the
// personal-deductions calculator's port, annual.LimitsReader (ANN-04 feeds
// ANN-07).
package annuallimits

import (
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Reader implements annual.LimitsReader over a parameter catalog.
type Reader struct{ sets *taxparams.Catalog }

// New returns a Reader over sets.
func New(sets *taxparams.Catalog) *Reader { return &Reader{sets: sets} }

// DeductionLimits returns the limits of fiscalYear, each resolved on 31
// December of that year (taxparams.ReadDeductionLimits). The conversion
// compiles only while taxparams.DeductionLimits keeps the fields, types and
// order of annual.Limits.
func (r *Reader) DeductionLimits(fiscalYear int) (annual.Limits, error) {
	limits, err := taxparams.ReadDeductionLimits[annual.TuitionLevel](r.sets, fiscalYear)
	if err != nil {
		return annual.Limits{}, err
	}
	return annual.Limits(limits), nil
}
