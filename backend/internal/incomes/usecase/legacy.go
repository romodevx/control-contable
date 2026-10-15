package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// errUnreadableCounterparty reports a legacy income whose counterparty_type
// maps to no fact: the column's CHECK rules it out, so it is a fault.
var errUnreadableCounterparty = errors.New("incomes: legacy counterparty_type maps to no fact")

// withFacts is s with the facts an income stored before WH-06 lacks while
// the WH-08 backfill has not derived it, filled with the backfill's own
// mapping (legacyFacts, MODULE-withholdings §5.2). Every read serves the
// facts the contract requires, and a patch applies to the facts the read
// served, whether or not the backfill has run: no deploy order (ADR-0008
// Amendment 2). The withholding stays as stored: a read never derives,
// since the estimate still uses the stored amounts.
func (uc *Incomes) withFacts(ctx context.Context, userID string, s StoredIncome) (StoredIncome, error) {
	if hasFacts(s) {
		return s, nil
	}
	profile, err := uc.deps.Taxpayers.Profile(ctx, userID, s.TaxpayerID)
	if err != nil {
		return StoredIncome{}, err
	}
	return factsFor(s, profile.PersonType)
}

// factsFor is withFacts for a taxpayer already read.
func factsFor(s StoredIncome, person fiscal.PersonType) (StoredIncome, error) {
	if hasFacts(s) {
		return s, nil
	}
	f, _, ok := legacyFacts(s, person)
	if !ok {
		return StoredIncome{}, fmt.Errorf("income %s: %w", s.ID, errUnreadableCounterparty)
	}
	s.Fields = f
	return s, nil
}

func hasFacts(s StoredIncome) bool {
	return s.Counterparty != "" && s.Operation != ""
}
