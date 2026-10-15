package estimates

import (
	"context"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// errUnknownTreatment reports a catalog treatment outside the contract enum.
var errUnknownTreatment = errors.New("taxengine: unknown regime treatment")

// Profiles adapts the taxpayers module's profile reader to the
// estimate's TaxpayerProfileReader port. Modules never import each other
// (ADR-0001), so the composition root translates the types. Its client
// errors (404) pass as they are.
type Profiles struct {
	reader tpusecase.TaxpayerProfileReader
}

var _ usecase.TaxpayerProfileReader = Profiles{}

// NewProfiles adapts reader.
func NewProfiles(reader tpusecase.TaxpayerProfileReader) Profiles { return Profiles{reader: reader} }

// Profile implements usecase.TaxpayerProfileReader.
func (a Profiles) Profile(ctx context.Context, userID, taxpayerID string) (usecase.TaxpayerProfile, error) {
	p, err := a.reader.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return usecase.TaxpayerProfile{}, err
	}
	out := usecase.TaxpayerProfile{PersonType: fiscal.PersonType(p.PersonType), Archived: p.Archived}
	for _, r := range p.Regimes {
		hold, err := estimateRegimeHold(r)
		if err != nil {
			return usecase.TaxpayerProfile{}, fmt.Errorf("regime %s: %w", r.Code, err)
		}
		out.Regimes = append(out.Regimes, hold)
	}
	return out, nil
}

// estimateRegimeHold is one regime of the profile as the estimate holds it.
func estimateRegimeHold(r tpusecase.RegimeProfile) (usecase.RegimeHold, error) {
	treatment := domain.Treatment(r.Treatment)
	if !treatment.Valid() {
		return usecase.RegimeHold{}, fmt.Errorf("%q: %w", r.Treatment, errUnknownTreatment)
	}
	start, err := domain.ParsePeriod(r.Window.Start.String())
	if err != nil {
		return usecase.RegimeHold{}, err
	}
	hold := usecase.RegimeHold{
		Code: fiscal.RegimeCode(r.Code), Treatment: treatment, Options: estimateOptions(r.Options), Start: start,
	}
	if !r.Window.IsOpen() {
		if hold.End, err = domain.ParsePeriod(r.Window.End.String()); err != nil {
			return usecase.RegimeHold{}, err
		}
	}
	return hold, nil
}

// estimateOptions copies each option value by its shape; a value holding
// nothing is left out (absent: the strategy's default).
func estimateOptions(options map[string]tpdomain.OptionValue) domain.Options {
	values := make(map[domain.OptionKey]domain.OptionValue, len(options))
	for key, v := range options {
		if value, ok := estimateOption(v); ok {
			values[domain.OptionKey(key)] = value
		}
	}
	return domain.NewOptions(values)
}

// estimateOption converts one value; false when it holds nothing.
func estimateOption(v tpdomain.OptionValue) (domain.OptionValue, bool) {
	if b, ok := v.Bool(); ok {
		return domain.BoolOption(b), true
	}
	if s, ok := v.Text(); ok {
		return domain.TextOption(s), true
	}
	if items, ok := v.List(); ok {
		return domain.ListOption(items), true
	}
	return domain.OptionValue{}, false
}
