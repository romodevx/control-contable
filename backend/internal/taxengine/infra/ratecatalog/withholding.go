package ratecatalog

import (
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// The withholding rate keys MODULE-withholdings §3 derives from (WH-05), as
// the params files spell them. The kernel owns the spelling of every key
// fiscal.DeriveWithholdings reads; these names re-export it. The base each
// rate applies to differs: see the accessors.
const (
	// KeyISRWithholdingRESICO is the ISR a persona moral withholds from a
	// RESICO PF, 1.25 % of the subtotal (LISR 113-J).
	KeyISRWithholdingRESICO = fiscal.KeyISRWithholdingRESICO
	// KeyIVAWithholdingFreight is the IVA a persona moral withholds on
	// ground freight of goods, 4 % of the PRICE (LIVA 1-A fr. II c; RLIVA
	// 3 fr. II, 4).
	KeyIVAWithholdingFreight = fiscal.KeyIVAWithholdingFreight
	// KeyIVAWithholdingScrap is the IVA a persona moral withholds when it
	// buys desperdicios, 100 % OF THE IVA charged (LIVA 1-A fr. II b; RLIVA 3
	// grants no reduction).
	KeyIVAWithholdingScrap fiscal.RateKey = "iva.withholding.scrap_pm"
)

// withholdingKeys are the withholding rates every served set must hold.
var withholdingKeys = []fiscal.RateKey{
	fiscal.KeyISRWithholdingFees,
	fiscal.KeyISRWithholdingRent,
	KeyISRWithholdingRESICO,
	KeyIVAWithholdingFreight,
	KeyIVAWithholdingScrap,
}

// requiredWithholdingParams are withholdingKeys as set requirements, plus
// the RMF 4.1.3 threshold IVA-7 compares against.
func requiredWithholdingParams() []taxparams.Requirement {
	reqs := make([]taxparams.Requirement, 0, len(withholdingKeys)+1)
	for _, k := range withholdingKeys {
		reqs = append(reqs, taxparams.RequireRate(taxparams.Key(k)))
	}
	return append(reqs, taxparams.RequireMoney(taxparams.Key(fiscal.KeyIVAFederalWithholdingThreshold)))
}

// The accessors below serve fiscal.DeriveWithholdings (WH-04). Each returns
// the rate in force on the record's booking date, or an error:
// ErrDateNotCovered when no set covers the date, and the wrapped taxparams
// ErrNotInForce when the key is not in force that day. They speak kernel
// types only, so a kernel consumer interface can name them; whether the
// value is provisional is Provisional(on). The 2/3, 1/2 and 1 shares of
// RLIVA 3 fr. I and LIVA 18-J stay in code (fiscal.Share*).

// ISRWithholdingFees is the ISR a persona moral withholds from a PF's
// professional fees (honorarios), a rate on the subtotal: 10 % in 2026
// (LISR 106).
func (c *Catalog[K]) ISRWithholdingFees(on fiscal.Date) (money.Rate, error) {
	return c.rate(fiscal.KeyISRWithholdingFees, on)
}

// ISRWithholdingRent is the ISR a persona moral withholds from a PF's
// real-estate rent (arrendamiento), a rate on the subtotal: 10 % in 2026
// (LISR 116).
func (c *Catalog[K]) ISRWithholdingRent(on fiscal.Date) (money.Rate, error) {
	return c.rate(fiscal.KeyISRWithholdingRent, on)
}

// ISRWithholdingRESICO is the ISR a persona moral withholds from a RESICO
// PF, a rate on the subtotal without IVA: 1.25 % in 2026 (LISR 113-J).
func (c *Catalog[K]) ISRWithholdingRESICO(on fiscal.Date) (money.Rate, error) {
	return c.rate(KeyISRWithholdingRESICO, on)
}

// IVAWithholdingFreight is the IVA a persona moral (or the Federation)
// withholds on ground freight of goods, a rate on the PRICE, not on the IVA:
// 4 % in 2026 (LIVA 1-A fr. II c; RLIVA 3 fr. II, 4).
func (c *Catalog[K]) IVAWithholdingFreight(on fiscal.Date) (money.Rate, error) {
	return c.rate(KeyIVAWithholdingFreight, on)
}

// IVAWithholdingScrap is the share OF THE IVA charged that a persona moral
// withholds when it buys desperdicios: 1 (100 %) in 2026 (LIVA 1-A fr. II b;
// RLIVA 3 grants no reduction).
func (c *Catalog[K]) IVAWithholdingScrap(on fiscal.Date) (money.Rate, error) {
	return c.rate(KeyIVAWithholdingScrap, on)
}

// IVAWithholdingFederalThreshold is the price above which the Federación
// withholds IVA from a PF on goods and services other than personal
// services and freight: $2,000.00 in 2026 (RMF 2026 rule 4.1.3). An RMF
// rule lasts one year, so in a later year it is ErrNotInForce until that
// year's RMF is transcribed.
func (c *Catalog[K]) IVAWithholdingFederalThreshold(on fiscal.Date) (money.Cents, error) {
	a, err := c.money(taxparams.Key(fiscal.KeyIVAFederalWithholdingThreshold), on)
	return a.Value, err
}

// Withholdings returns the view of c that fiscal.DeriveWithholdings reads.
func (c *Catalog[K]) Withholdings() WithholdingView[K] { return WithholdingView[K]{c: c} }

// WithholdingView is a Catalog seen through the kernel's
// fiscal.WithholdingCatalog port: its Amount takes the kernel's
// fiscal.AmountKey, whatever amount key type K the Catalog serves.
type WithholdingView[K ~string] struct {
	c *Catalog[K]
}

// Rate implements fiscal.RateCatalog.
func (v WithholdingView[K]) Rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, bool) {
	return v.c.Rate(key, on)
}

// Amount returns the money parameter key holds on day on, and false when
// none does. It implements fiscal.WithholdingCatalog.
func (v WithholdingView[K]) Amount(key fiscal.AmountKey, on fiscal.Date) (money.Cents, bool) {
	a, err := v.c.money(taxparams.Key(key), on)
	return a.Value, err == nil
}

func (c *Catalog[K]) rate(key fiscal.RateKey, on fiscal.Date) (money.Rate, error) {
	r, err := c.RateOn(key, on)
	return r.Value, err
}
