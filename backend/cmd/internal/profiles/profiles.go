// Package profiles adapts the taxpayers module's profile to the incomes and
// expenses ports. The modules never import each other (archtest
// cross-module); only composition roots see them all, and two of them need
// this mapping: cmd/api and cmd/backfill-withholdings.
package profiles

import (
	"context"

	expdomain "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	expenses "github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	incdomain "github.com/romodevx/control-contable/backend/internal/incomes/domain"
	incomes "github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	tpdomain "github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	taxpayers "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// The regime options (taxpayers catalog keys) the record modules read.
const (
	optionActivities         = "activities"            // 625: the income withholding rules
	optionRFCGivenToPlatform = "rfc_given_to_platform" // 625
	optionDeductionMode      = "deduction_mode"        // 606: an expense category's hint
	optionDefinitiveOption   = "definitive_option"     // 625
)

// Incomes is p as the incomes module reads it.
func Incomes(p taxpayers.TaxpayerProfile) incomes.TaxpayerProfile {
	out := incomes.TaxpayerProfile{PersonType: fiscal.PersonType(p.PersonType), Archived: p.Archived}
	for _, r := range p.Regimes {
		held := incomes.RegimeProfile{
			Code:               fiscal.RegimeCode(r.Code),
			StartPeriod:        r.Window.Start.String(),
			PlatformActivities: activities(r.Options[optionActivities]),
			RFCGivenToPlatform: flag(r.Options[optionRFCGivenToPlatform]),
		}
		if !r.Window.IsOpen() {
			held.EndPeriod = r.Window.End.String()
		}
		out.Regimes = append(out.Regimes, held)
	}
	return out
}

func activities(v tpdomain.OptionValue) []incdomain.PlatformActivity {
	list, ok := v.List()
	if !ok {
		return nil
	}
	out := make([]incdomain.PlatformActivity, len(list))
	for i, a := range list {
		out[i] = incdomain.PlatformActivity(a)
	}
	return out
}

func flag(v tpdomain.OptionValue) bool {
	b, _ := v.Bool()
	return b
}

// Expenses is p as the expenses module reads it.
func Expenses(p taxpayers.TaxpayerProfile) expenses.TaxpayerProfile {
	out := expenses.TaxpayerProfile{PersonType: fiscal.PersonType(p.PersonType), Archived: p.Archived}
	for _, r := range p.Regimes {
		out.Regimes = append(out.Regimes, regimeHold(r))
	}
	return out
}

func regimeHold(r taxpayers.RegimeProfile) expenses.RegimeHold {
	hold := expenses.RegimeHold{Code: fiscal.RegimeCode(r.Code), StartPeriod: r.Window.Start.String()}
	if !r.Window.IsOpen() {
		hold.EndPeriod = r.Window.End.String()
	}
	if mode, ok := r.Options[optionDeductionMode].Text(); ok {
		hold.Options.DeductionMode = expdomain.DeductionMode(mode)
	}
	if definitive, ok := r.Options[optionDefinitiveOption].Bool(); ok {
		hold.Options.DefinitiveOption = definitive
	}
	return hold
}

// IncomesReader is the taxpayers profile reader as incomes.TaxpayerProfiles.
// Its client errors (404) pass as they are.
type IncomesReader struct {
	Reader taxpayers.TaxpayerProfileReader
}

var _ incomes.TaxpayerProfiles = IncomesReader{}

// Profile implements incomes.TaxpayerProfiles.
func (a IncomesReader) Profile(ctx context.Context, userID, taxpayerID string) (incomes.TaxpayerProfile, error) {
	p, err := a.Reader.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return incomes.TaxpayerProfile{}, err
	}
	return Incomes(p), nil
}

// ExpensesReader is the taxpayers profile reader as
// expenses.TaxpayerProfiles. Its client errors (404) pass as they are.
type ExpensesReader struct {
	Reader taxpayers.TaxpayerProfileReader
}

var _ expenses.TaxpayerProfiles = ExpensesReader{}

// Profile implements expenses.TaxpayerProfiles.
func (a ExpensesReader) Profile(ctx context.Context, userID, taxpayerID string) (expenses.TaxpayerProfile, error) {
	p, err := a.Reader.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return expenses.TaxpayerProfile{}, err
	}
	return Expenses(p), nil
}
