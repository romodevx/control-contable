package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Share is a legal fraction Num/Den of a base, for the IVA withholding a Rate
// cannot hold exactly (2/3, ADR-0002 §4). The zero Share means "none".
type Share struct {
	Num, Den int64
}

// ShareTwoThirds is the IVA a persona moral withholds on PF fees and rent
// (RLIVA 3 fr. I): a fraction of the IVA, not a yearly rate.
var ShareTwoThirds = Share{Num: 2, Den: 3}

// IsZero reports the zero Share.
func (s Share) IsZero() bool { return s.Den == 0 }

// of returns base × Num/Den rounded once to the centavo, half away from
// zero. Only a base beyond money.Exact's range fails, a parameter fault.
func (s Share) of(base money.Cents) (money.Cents, error) {
	e, err := base.Exact()
	if err == nil {
		e, err = e.MulRatio(s.Num, s.Den)
	}
	if err != nil {
		return 0, fmt.Errorf("apply share %d/%d: %w: %w", s.Num, s.Den, ErrInvalidParameter, err)
	}
	return e.Round(), nil
}

// Withholdings are the amounts the payer withheld from its supplier.
type Withholdings struct {
	ISR money.Cents
	IVA money.Cents
}

// Suggestion is a suggested withholding; Known is false when there is none.
type Suggestion struct {
	Cents money.Cents
	Known bool
}

// Suggested are the suggested withholdings of an expense.
type Suggested struct {
	ISR Suggestion
	IVA Suggestion
}

// supplierRule is how a persona moral withholds for one category: ISR at a
// rate of the price, and IVA either as a share of the IVA or at a rate of
// the price (freight).
type supplierRule struct {
	isr      RateKey
	ivaShare Share
	ivaPrice RateKey
}

// supplierRules are the categories with a suggestion (MODULE-expenses §2.6).
var supplierRules = map[CategoryCode]supplierRule{
	CategoryProfessionalFees: {isr: KeyISRWithholdingFees, ivaShare: ShareTwoThirds},
	CategoryRent:             {isr: KeyISRWithholdingRent, ivaShare: ShareTwoThirds},
	CategoryFreight:          {ivaPrice: KeyIVAWithholdingFreight},
}

// MayWithhold reports whether the payer may withhold from its suppliers:
// only a persona moral (research B.2.3, LIVA 1-A).
func MayWithhold(p Payer) bool { return p.PersonType == PersonMoral }

// SuggestWithholdings returns what a persona moral would withhold for the
// category on day on: ISR 10 % and 2/3 of the IVA on fees and rent, 4 % of
// the price on freight. It is UX only (the stored value is the user's) and
// never warns. A rate the catalog lacks is no suggestion.
func SuggestWithholdings(category CategoryCode, p Payer, amount, iva money.Cents, on Date, catalog RateCatalog) (Suggested, error) {
	rule, ok := supplierRules[category]
	if !ok || !MayWithhold(p) {
		return Suggested{}, nil
	}
	isr, err := rateSuggestion(rule.isr, amount, on, catalog)
	if err != nil {
		return Suggested{}, err
	}
	ivaWithheld, err := rule.ivaSuggestion(amount, iva, on, catalog)
	if err != nil {
		return Suggested{}, err
	}
	return Suggested{ISR: isr, IVA: ivaWithheld}, nil
}

func (r supplierRule) ivaSuggestion(amount, iva money.Cents, on Date, catalog RateCatalog) (Suggestion, error) {
	if r.ivaPrice != "" {
		return rateSuggestion(r.ivaPrice, amount, on, catalog)
	}
	c, err := r.ivaShare.of(iva)
	return Suggestion{Cents: c, Known: err == nil}, err
}

func rateSuggestion(key RateKey, base money.Cents, on Date, catalog RateCatalog) (Suggestion, error) {
	if key == "" {
		return Suggestion{}, nil
	}
	r, ok := catalog.Rate(key, on)
	if !ok {
		return Suggestion{}, nil
	}
	c, err := applyRate(base, r)
	return Suggestion{Cents: c, Known: err == nil}, err
}
