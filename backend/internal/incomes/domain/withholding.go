package domain

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Payee is the taxpayer receiving the income, as the taxpayers module
// reports it (person type, regime and its catalog basis, 625 options).
type Payee struct {
	PersonType         fiscal.PersonType
	Regime             fiscal.RegimeCode
	Basis              IncomeBasis
	PlatformActivities []PlatformActivity // 625 option "activities"
	RFCGivenToPlatform bool               // 625 option "rfc_given_to_platform"
}

// ISRAllowance says whether ISR may be withheld and which dated rate the
// suggestion uses (an empty Key: none).
type ISRAllowance struct {
	Allowed bool
	Key     fiscal.RateKey
}

// IVAAllowance says whether IVA may be withheld and which share of the IVA
// the suggestion uses (a zero Share: none).
type IVAAllowance struct {
	Allowed bool
	Share   fiscal.Share
}

// WithholdingPolicy is what a counterparty may withhold from a payee.
type WithholdingPolicy struct {
	ISR ISRAllowance
	IVA IVAAllowance
}

// pmFeesKeys are the ISR withholdings a persona moral makes on payments to
// a persona física, by the física's regime (research B.2.3).
var pmFeesKeys = map[fiscal.RegimeCode]fiscal.RateKey{
	fiscal.RegimeBusiness: fiscal.KeyISRWithholdingFees,
	fiscal.RegimeRent:     fiscal.KeyISRWithholdingRent,
	fiscal.RegimeRESICO:   KeyISRWithholdingRESICO,
}

// platformKeys are the 625 ISR withholdings by activity (LISR 113-A).
var platformKeys = map[PlatformActivity]fiscal.RateKey{
	ActivityTransportDelivery: KeyISRPlatformTransport,
	ActivityLodging:           KeyISRPlatformLodging,
	ActivityGoodsServices:     KeyISRPlatformGoods,
}

// PolicyFor is the applicability table of MODULE-incomes §2.6. Anything
// outside it allows no withholding: a value > 0 is
// fiscal.ErrWithholdingNotApplicable.
func PolicyFor(p Payee, c Counterparty, activity PlatformActivity) WithholdingPolicy {
	if p.PersonType == fiscal.PersonMoral {
		return moralPolicy(c)
	}
	switch {
	case c == CounterpartyMoral:
		return fisicaPaidByMoral(p.Regime)
	case c == CounterpartyPlatform && p.Regime == fiscal.RegimePlatforms:
		return platformPolicy(activity, p.RFCGivenToPlatform)
	}
	return WithholdingPolicy{}
}

// moralPolicy: a platform withholds ISR from a PM (2.5 % in 2026, no IVA);
// a PM may withhold IVA from a PM (4 % freight), with no suggestion.
func moralPolicy(c Counterparty) WithholdingPolicy {
	switch c {
	case CounterpartyPlatform:
		return WithholdingPolicy{ISR: ISRAllowance{Allowed: true, Key: KeyISRPlatformPersonaMoral}}
	case CounterpartyMoral:
		return WithholdingPolicy{IVA: IVAAllowance{Allowed: true}}
	}
	return WithholdingPolicy{}
}

// fisicaPaidByMoral: 612 fees and 606 rent at 10 %, 626 at 1.25 %, plus
// 2/3 of the IVA.
func fisicaPaidByMoral(regime fiscal.RegimeCode) WithholdingPolicy {
	key, ok := pmFeesKeys[regime]
	if !ok {
		return WithholdingPolicy{}
	}
	return WithholdingPolicy{
		ISR: ISRAllowance{Allowed: true, Key: key},
		IVA: IVAAllowance{Allowed: true, Share: fiscal.ShareTwoThirds},
	}
}

// platformPolicy: 625 by activity and half the IVA; without an RFC given to
// the platform, 20 % and all of the IVA.
func platformPolicy(activity PlatformActivity, rfcGiven bool) WithholdingPolicy {
	key, ok := platformKeys[activity]
	if !ok {
		return WithholdingPolicy{}
	}
	share := fiscal.ShareHalf
	if !rfcGiven {
		key, share = KeyISRPlatformNoRFC, fiscal.ShareWhole
	}
	return WithholdingPolicy{
		ISR: ISRAllowance{Allowed: true, Key: key},
		IVA: IVAAllowance{Allowed: true, Share: share},
	}
}

// Suggest computes the withholdings the client prefills (MODULE-incomes
// §2.6): ISR at the catalog rate in force on day on (the booking date; LIF
// rates change yearly), IVA as the policy's share of iva. A rate the catalog
// lacks on that day is no suggestion, never an error.
func (p WithholdingPolicy) Suggest(amount, iva money.Cents, on fiscal.Date, catalog fiscal.RateCatalog) (fiscal.Suggested, error) {
	isr, err := fiscal.SuggestAtRate(p.ISR.Key, amount, on, catalog)
	if err != nil {
		return fiscal.Suggested{}, err
	}
	ivaWithheld, err := p.IVA.Share.Suggest(iva)
	if err != nil {
		return fiscal.Suggested{}, err
	}
	return fiscal.Suggested{ISR: isr, IVA: ivaWithheld}, nil
}
