package domain

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// FingerprintRegime is one regime as the estimate read it: its strategy
// input, the catalog treatment that chose the strategy, and the months of
// the fiscal year it was held up to the period.
type FingerprintRegime struct {
	Input        Input
	Treatment    Treatment
	ActiveMonths []time.Month
}

// FingerprintInput is everything an estimate is a function of.
type FingerprintInput struct {
	PersonType   fiscal.PersonType
	Period       Period
	ParameterSet string // "2026/v1"
	Provisional  bool
	Regimes      []FingerprintRegime
}

// Fingerprint is the records fingerprint (ADR-0002 §5): the lower-case hex
// SHA-256 of a canonical text of in and EngineVersion. The text is
// independent of the order of the regimes, options and tax inputs, so equal
// inputs always give an equal fingerprint (M8 marks a month "changed since
// viewed" when it moves). O(regimes × (12 + options + inputs) log).
func Fingerprint(in FingerprintInput) string {
	b := fmt.Appendf(nil, "engine %q\nparams %q %t\nperson %q\nperiod %q\n",
		EngineVersion, in.ParameterSet, in.Provisional, in.PersonType, in.Period)
	regimes := slices.Clone(in.Regimes)
	slices.SortFunc(regimes, func(a, b FingerprintRegime) int {
		return cmp.Or(cmp.Compare(a.Input.Key.Regime, b.Input.Key.Regime), cmp.Compare(a.Input.Key.PersonType, b.Input.Key.PersonType))
	})
	for _, r := range regimes {
		b = appendRegime(b, r)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// appendRegime writes one regime's canonical text. An informational regime
// computes nothing (ADR-0002 §1), so only that it is held counts: its
// options, tax inputs and record sums never change a figure and are left
// out (#98 FU1). A multi_enum is a set, written sorted.
func appendRegime(b []byte, r FingerprintRegime) []byte {
	in := r.Input
	b = fmt.Appendf(b, "regime %q %q %q active %v\n", in.Key, r.Treatment, in.Period, r.ActiveMonths)
	if r.Treatment == TreatmentInformational {
		return b
	}
	for _, key := range slices.Sorted(maps.Keys(in.Options.values)) {
		v := in.Options.values[key]
		b = fmt.Appendf(b, "option %q %d %t %q %q\n", key, v.kind, v.flag, v.text, slices.Sorted(slices.Values(v.list)))
	}
	slots := slices.SortedFunc(maps.Keys(in.Inputs.values), func(a, b inputSlot) int {
		return cmp.Or(cmp.Compare(a.key, b.key), cmp.Compare(a.month, b.month))
	})
	for _, slot := range slots {
		v := in.Inputs.values[slot]
		b = fmt.Appendf(b, "input %q %q %d %d %d %t\n", v.Key, v.Kind, v.Month, v.Money, v.Rate, v.Bool)
	}
	// The previous December's withheld IVA is January's credit. It is hashed
	// for every month: a later edit of it may mark a month as changed whose
	// figures did not move, never the reverse (M8).
	b = fmt.Appendf(b, "prior_december_withheld %d\n", in.Sums.CreditableWithheldPriorMonth(time.January))
	for m := time.January; m <= in.Sums.Through(); m++ {
		b = fmt.Appendf(b, "month %d %+v\n", m, in.Sums.Month(m))
	}
	return b
}
