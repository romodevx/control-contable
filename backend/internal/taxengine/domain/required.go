package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// MissingInputs checks the taxpayer's tax inputs against what the period
// requires (Strategy.RequiredInputs) and returns what is absent: a scalar key
// with no value of its kind, or the months of a money_monthly key with no
// value. An entered zero is present. Keys keep the requirement order, so the
// first one is the CTA's focus (MODULE-tax-engine §2.8); months ascend; a key
// required twice is reported once. nil means nothing is missing.
//
// A malformed requirement (blank key, unknown kind, months on a scalar key,
// a month outside 1-12, one key under two kinds) is a strategy bug and fails
// with ErrInvalidTaxInput or ErrInvalidMonth.
func MissingInputs(required []RequiredInput, in TaxInputs) ([]MissingInput, error) {
	var missing []MissingInput
	position := make(map[InputKey]int, len(required))
	kinds := make(map[InputKey]InputKind, len(required))
	for _, req := range required {
		if err := checkRequirement(req, kinds); err != nil {
			return nil, err
		}
		kinds[req.Key] = req.Kind
		months, absent := in.absent(req)
		if !absent {
			continue
		}
		if i, seen := position[req.Key]; seen {
			missing[i].Months = sortedMonths(append(missing[i].Months, months...))
			continue
		}
		position[req.Key] = len(missing)
		missing = append(missing, MissingInput{Key: req.Key, Months: months})
	}
	return missing, nil
}

func checkRequirement(req RequiredInput, kinds map[InputKey]InputKind) error {
	if kind, seen := kinds[req.Key]; seen && kind != req.Kind {
		return fmt.Errorf("required %s as %s and %s: %w", req.Key, kind, req.Kind, ErrInvalidTaxInput)
	}
	switch {
	case strings.TrimSpace(string(req.Key)) == "" || !req.Kind.Valid():
		return fmt.Errorf("required key %q kind %q: %w", req.Key, req.Kind, ErrInvalidTaxInput)
	case req.Kind != InputMoneyMonthly && len(req.Months) > 0:
		return fmt.Errorf("required %s: months on a %s input: %w", req.Key, req.Kind, ErrInvalidMonth)
	case slices.ContainsFunc(req.Months, func(m time.Month) bool { return !validMonth(m) }):
		return fmt.Errorf("required %s: months %v: %w", req.Key, req.Months, ErrInvalidMonth)
	}
	return nil
}

// absent reports whether req is unmet and, for money_monthly, which months.
func (in TaxInputs) absent(req RequiredInput) (months []time.Month, absent bool) {
	if req.Kind != InputMoneyMonthly {
		_, present := in.lookup(req.Key, req.Kind, 0)
		return nil, !present
	}
	for _, m := range req.Months {
		if _, present := in.lookup(req.Key, req.Kind, m); !present {
			months = append(months, m)
		}
	}
	months = sortedMonths(months)
	return months, len(months) > 0
}

// sortedMonths sorts months ascending and drops repeats, in place.
func sortedMonths(months []time.Month) []time.Month {
	slices.Sort(months)
	return slices.Compact(months)
}
