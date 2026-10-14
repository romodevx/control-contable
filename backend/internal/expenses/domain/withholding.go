package domain

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// supplierRule is how a persona moral withholds for one category: ISR at a
// rate of the price, and IVA either as a share of the IVA or at a rate of
// the price (freight).
type supplierRule struct {
	isr      fiscal.RateKey
	ivaShare fiscal.Share
	ivaPrice fiscal.RateKey
}

// supplierRules are the categories with a suggestion (MODULE-expenses §2.6).
var supplierRules = map[CategoryCode]supplierRule{
	CategoryProfessionalFees: {isr: fiscal.KeyISRWithholdingFees, ivaShare: fiscal.ShareTwoThirds},
	CategoryRent:             {isr: fiscal.KeyISRWithholdingRent, ivaShare: fiscal.ShareTwoThirds},
	CategoryFreight:          {ivaPrice: KeyIVAWithholdingFreight},
}

// MayWithhold reports whether the payer may withhold from its suppliers:
// only a persona moral (research B.2.3, LIVA 1-A).
func MayWithhold(p Payer) bool { return p.PersonType == fiscal.PersonMoral }

// SuggestWithholdings returns what a persona moral would withhold for the
// category on day on: ISR 10 % and 2/3 of the IVA on fees and rent, 4 % of
// the price on freight. It is UX only (the stored value is the user's) and
// never warns. A rate the catalog lacks is no suggestion.
func SuggestWithholdings(category CategoryCode, p Payer, amount, iva money.Cents, on fiscal.Date, catalog fiscal.RateCatalog) (fiscal.Suggested, error) {
	rule, ok := supplierRules[category]
	if !ok || !MayWithhold(p) {
		return fiscal.Suggested{}, nil
	}
	isr, err := fiscal.SuggestAtRate(rule.isr, amount, on, catalog)
	if err != nil {
		return fiscal.Suggested{}, err
	}
	ivaWithheld, err := rule.ivaSuggestion(amount, iva, on, catalog)
	if err != nil {
		return fiscal.Suggested{}, err
	}
	return fiscal.Suggested{ISR: isr, IVA: ivaWithheld}, nil
}

func (r supplierRule) ivaSuggestion(amount, iva money.Cents, on fiscal.Date, catalog fiscal.RateCatalog) (fiscal.Suggestion, error) {
	if r.ivaPrice != "" {
		return fiscal.SuggestAtRate(r.ivaPrice, amount, on, catalog)
	}
	return r.ivaShare.Suggest(iva)
}
