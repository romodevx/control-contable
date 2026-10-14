package domain

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Share is a legal fraction Num/Den of a base, for the IVA withholdings a
// Rate cannot hold exactly (2/3 = 10.6667 % of the price, ADR-0002 §4). The
// zero Share means "no suggestion".
type Share struct {
	Num, Den int64
}

// The IVA withholding shares of the law. They are fractions of the IVA, not
// yearly rates: RLIVA 3 fr. I (2/3) and LIVA 18-J fr. II a) (50 %, 100 %
// without RFC).
var (
	ShareTwoThirds = Share{Num: 2, Den: 3}
	ShareHalf      = Share{Num: 1, Den: 2}
	ShareWhole     = Share{Num: 1, Den: 1}
)

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

// Payee is the taxpayer receiving the income, as the taxpayers module
// reports it (person type, regime and its catalog basis, 625 options).
type Payee struct {
	PersonType         PersonType
	Regime             RegimeCode
	Basis              IncomeBasis
	PlatformActivities []PlatformActivity // 625 option "activities"
	RFCGivenToPlatform bool               // 625 option "rfc_given_to_platform"
}

// Withholdings are the amounts a payer withheld from the income.
type Withholdings struct {
	ISR money.Cents
	IVA money.Cents
}

// ISRAllowance says whether ISR may be withheld and which dated rate the
// suggestion uses (an empty Key: none).
type ISRAllowance struct {
	Allowed bool
	Key     RateKey
}

// IVAAllowance says whether IVA may be withheld and which share of the IVA
// the suggestion uses (a zero Share: none).
type IVAAllowance struct {
	Allowed bool
	Share   Share
}

// WithholdingPolicy is what a counterparty may withhold from a payee.
type WithholdingPolicy struct {
	ISR ISRAllowance
	IVA IVAAllowance
}

// pmFeesKeys are the ISR withholdings a persona moral makes on payments to
// a persona física, by the física's regime (research B.2.3).
var pmFeesKeys = map[RegimeCode]RateKey{
	RegimeBusiness: KeyISRWithholdingFees,
	RegimeRent:     KeyISRWithholdingRent,
	RegimeRESICO:   KeyISRWithholdingRESICO,
}

// platformKeys are the 625 ISR withholdings by activity (LISR 113-A).
var platformKeys = map[PlatformActivity]RateKey{
	ActivityTransportDelivery: KeyISRPlatformTransport,
	ActivityLodging:           KeyISRPlatformLodging,
	ActivityGoodsServices:     KeyISRPlatformGoods,
}

// PolicyFor is the applicability table of MODULE-incomes §2.6. Anything
// outside it allows no withholding: a value > 0 is
// ErrWithholdingNotApplicable.
func PolicyFor(p Payee, c Counterparty, activity PlatformActivity) WithholdingPolicy {
	if p.PersonType == PersonMoral {
		return moralPolicy(c)
	}
	switch {
	case c == CounterpartyMoral:
		return fisicaPaidByMoral(p.Regime)
	case c == CounterpartyPlatform && p.Regime == RegimePlatforms:
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
func fisicaPaidByMoral(regime RegimeCode) WithholdingPolicy {
	key, ok := pmFeesKeys[regime]
	if !ok {
		return WithholdingPolicy{}
	}
	return WithholdingPolicy{
		ISR: ISRAllowance{Allowed: true, Key: key},
		IVA: IVAAllowance{Allowed: true, Share: ShareTwoThirds},
	}
}

// platformPolicy: 625 by activity and half the IVA; without an RFC given to
// the platform, 20 % and all of the IVA.
func platformPolicy(activity PlatformActivity, rfcGiven bool) WithholdingPolicy {
	key, ok := platformKeys[activity]
	if !ok {
		return WithholdingPolicy{}
	}
	share := ShareHalf
	if !rfcGiven {
		key, share = KeyISRPlatformNoRFC, ShareWhole
	}
	return WithholdingPolicy{
		ISR: ISRAllowance{Allowed: true, Key: key},
		IVA: IVAAllowance{Allowed: true, Share: share},
	}
}

// Suggestion is a suggested withholding; Known is false when there is none.
type Suggestion struct {
	Cents money.Cents
	Known bool
}

// Suggested are the suggested withholdings of an income.
type Suggested struct {
	ISR Suggestion
	IVA Suggestion
}

// Suggest computes the withholdings the client prefills (MODULE-incomes
// §2.6): ISR at the catalog rate in force on day on (the booking date; LIF
// rates change yearly), IVA as the policy's share of iva. A rate the catalog
// lacks on that day is no suggestion, never an error.
func (p WithholdingPolicy) Suggest(amount, iva money.Cents, on Date, catalog RateCatalog) (Suggested, error) {
	isr, err := p.ISR.suggest(amount, on, catalog)
	if err != nil {
		return Suggested{}, err
	}
	ivaWithheld, err := p.IVA.suggest(iva)
	if err != nil {
		return Suggested{}, err
	}
	return Suggested{ISR: isr, IVA: ivaWithheld}, nil
}

func (a ISRAllowance) suggest(amount money.Cents, on Date, catalog RateCatalog) (Suggestion, error) {
	if a.Key == "" {
		return Suggestion{}, nil
	}
	r, ok := catalog.Rate(a.Key, on)
	if !ok {
		return Suggestion{}, nil
	}
	c, err := applyRate(amount, r)
	return Suggestion{Cents: c, Known: err == nil}, err
}

func (a IVAAllowance) suggest(iva money.Cents) (Suggestion, error) {
	if a.Share.IsZero() {
		return Suggestion{}, nil
	}
	c, err := a.Share.of(iva)
	return Suggestion{Cents: c, Known: err == nil}, err
}

// differs reports a known suggestion farther than IVATolerance from the
// withheld amount: the CFDI wins, but the user is told.
func (s Suggestion) differs(withheld money.Cents) bool {
	return s.Known && !withinTolerance(withheld, s.Cents)
}
