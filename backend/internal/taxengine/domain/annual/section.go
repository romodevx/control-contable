package annual

import (
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Section is one computed section of the annual estimate (MODULE-annual
// §2.1): the regimes it covers, its status and obligation, the figures and
// the breakdown lines that trace them. The use case (ANN-14) adds the
// deadline, the explanation texts and the estimate-level disclaimers.
type Section struct {
	Kind          SectionKind
	Regimes       []fiscal.RegimeCode // ascending code order
	Status        Status
	Obligation    Obligation
	MissingInputs []string
	CPATopics     []CPATopic
	// Disclaimers are the codes this section raises; the estimate orders
	// and deduplicates them (§2.8).
	Disclaimers []DisclaimerCode
	// Tax is the balance of a section that has one; nil for an
	// informational section.
	Tax *Tax
	// RESICOCheck is the pf_resico_check reconciliation; nil otherwise,
	// and nil when no 113-F rate applies (income above the table).
	RESICOCheck *RESICOCheck
	Lines       []domain.Line
}

// RESICOCheck is the LISR 113-F informative reconciliation (contract
// ResicoPfCheck): Income x Rate = AnnualISR, less the monthly Payments and
// the 1.25 % Withheld, is the Difference. Each figure is rounded once; the
// Difference is never payable nor refundable (RMF 2026 rule 3.13.7).
type RESICOCheck struct {
	Income     money.Cents
	Rate       money.Rate
	AnnualISR  money.Cents
	Payments   money.Cents
	Withheld   money.Cents
	Difference money.Cents
}

// Tax is a section's figures, each rounded once to the centavo.
type Tax struct {
	AccumulableIncome  money.Cents
	PersonalDeductions money.Cents
	Base               money.Cents
	ISR                money.Cents
	Payments           money.Cents
	Withholdings       money.Cents
	SalaryISR          money.Cents
	Balance            Balance
}

// Balance is a section's balance: its kind is the sign and Amount is never
// negative.
type Balance struct {
	Kind   BalanceKind
	Amount money.Cents
}

// balanceOf rounds an exact tax-minus-credits value once: positive is a
// cargo, negative a favor, a value that rounds to zero is cero.
func balanceOf(v money.Exact) Balance {
	switch rounded := v.Round(); {
	case rounded > 0:
		return Balance{Kind: BalanceACargo, Amount: rounded}
	case rounded < 0:
		return Balance{Kind: BalanceAFavor, Amount: -rounded}
	}
	return Balance{Kind: BalanceCero}
}

// annualTrace drives one section's Ledger: centavos convert to Exact with a
// sticky error beside the Ledger's own, as the monthly strategies do.
type annualTrace struct {
	l   *domain.Ledger
	err error
}

func newAnnualTrace() *annualTrace { return &annualTrace{l: domain.NewLedger()} }

func (t *annualTrace) exact(c money.Cents) money.Exact {
	if t.err != nil {
		return 0
	}
	e, err := c.Exact()
	if err != nil {
		t.err = fmt.Errorf("annual: %w", err)
	}
	return e
}

// sum adds centavos, keeping the first overflow.
func (t *annualTrace) sum(values ...money.Cents) money.Cents {
	var total money.Cents
	for _, v := range values {
		next, err := total.Add(v)
		if err != nil && t.err == nil {
			t.err = fmt.Errorf("annual: %w", err)
		}
		total = next
	}
	return total
}

// lines closes the trace: its lines, or the first error.
func (t *annualTrace) lines() ([]domain.Line, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.l.Lines()
}

// annualStep is an ISR step of an annual section.
func annualStep(code domain.LineCode, label domain.LabelKey, src domain.Source, basis []domain.Citation, args ...string) domain.Step {
	return domain.Step{
		Tax: domain.TaxISR, Code: code, Label: label, Args: args, Source: src, Basis: slices.Clone(basis),
	}
}

// sortedTopics returns the topics in code order without repeats, or nil.
func sortedTopics(topics []CPATopic) []CPATopic {
	if len(topics) == 0 {
		return nil
	}
	return slices.Compact(slices.Sorted(slices.Values(topics)))
}
