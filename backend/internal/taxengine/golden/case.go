package golden

import (
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Case is one golden: a strategy input and what the strategy must return.
type Case struct {
	ID     string
	File   string // path under the golden root; "" for a generated case
	Source string // "research:<worked example id>", "cpa:<ref>" or "generated:<edge>"
	// SkipReason, when set, names the unverified (⚠️) rule the expected
	// values depend on: the case is reported as skipped and never run.
	SkipReason string
	// PendingStrategy marks an authored case whose strategy is not
	// registered yet. The runner fails a stale marker, so it cannot hide a
	// case once its strategy lands.
	PendingStrategy bool
	// Generated cases come from the boundary generator: they are pending
	// while their strategy is missing and run as soon as it is registered.
	Generated bool
	Input     domain.Input
	// ActiveMonths are the months of the fiscal year the regime was held,
	// up to the period (RequiredInputs facts); January..period by default.
	ActiveMonths []time.Month
	Want         Expected
}

// Expected is what a case asserts. Taxes scopes the per-tax figures and
// lines; a nil list is not asserted (generated cases only: an authored
// case asserts every list, absent meaning empty).
type Expected struct {
	Status domain.Status
	Taxes  []domain.TaxKind
	// ISR is asserted when Taxes has isr; nil expects no ISR figure.
	ISR *money.Cents
	// IVA is asserted when Taxes has iva; nil expects no IVA figure.
	IVA *domain.IVAComponents
	// WithheldToRemit is asserted when Taxes has withholding.
	WithheldToRemit money.Cents
	Lines           []LineWant
	MissingInputs   []domain.MissingInput
	CPATopics       []domain.CPATopic
	Disclaimers     []domain.DisclaimerCode
}

// LineWant is one expected breakdown line of an asserted tax, in order.
type LineWant struct {
	Tax    domain.TaxKind
	Code   domain.LineCode
	Amount money.Cents
}

// asserts reports whether the case asserts tax k.
func (e Expected) asserts(k domain.TaxKind) bool { return slices.Contains(e.Taxes, k) }
