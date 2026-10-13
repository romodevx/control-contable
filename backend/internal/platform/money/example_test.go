package money_test

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Example_pf612ISR reproduces research worked example PF-612-ISR-01: the
// marginal tariff is computed in Exact and rounded once, at the end.
func Example_pf612ISR() {
	fixedFee, _ := money.ParseCents("5570.52")
	base, _ := money.ParseCents("60000.00")
	lowerLimit, _ := money.ParseCents("52600.93")
	rate, _ := money.ParseRate("0.2136")

	excess, _ := base.Sub(lowerLimit)
	marginal, _ := excess.MulRate(rate)
	fee, _ := fixedFee.Exact()
	isr, _ := fee.Add(marginal)

	fmt.Println(isr.Round())
	// Output: 7150.96
}

// Example_ivaWithheld reproduces PF-612-IVA-01: two thirds of the IVA is an
// exact ratio, not a truncated rate.
func Example_ivaWithheld() {
	ivaOnFees, _ := money.ParseCents("1600.00")
	exact, _ := ivaOnFees.Exact()
	withheld, _ := exact.MulRatio(2, 3)

	fmt.Println(withheld.Round())
	// Output: 1066.67
}
