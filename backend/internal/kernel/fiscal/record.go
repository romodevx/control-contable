package fiscal

import (
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// WithholdingRecord is a derivation as a record stores it
// (MODULE-withholdings §3.3, ADR-0008): the amounts, how they were
// computed, the rules that decided and the table version that produced
// them. Incomes and expenses store it alike; the zero value is a record
// stored before the derivation existed (no TableVersion).
type WithholdingRecord struct {
	Withholdings
	ISRRate       money.Rate
	IVABasis      IVABasis
	Withholder    Withholder
	RuleIDs       []RuleID
	TableVersion  string
	CPAUnverified bool
}

// Record is d as a record stores it, stamped with the current table
// version.
func (d Derivation) Record() WithholdingRecord {
	return WithholdingRecord{
		Withholdings: d.Withholdings, ISRRate: d.ISRRate, IVABasis: d.IVABasis, Withholder: d.Withholder,
		RuleIDs: slices.Clone(d.RuleIDs), TableVersion: WithholdingTableVersion, CPAUnverified: d.CPAUnverified(),
	}
}

// Derived reports a record whose withholdings the server derived.
func (r WithholdingRecord) Derived() bool { return r.TableVersion != "" }

// Net is what changes hands: subtotal + IVA − ISR − IVA withheld. Only
// amounts beyond the contract's range overflow, a parameter fault.
func (r WithholdingRecord) Net(subtotal, iva money.Cents) (money.Cents, error) {
	gross, err := subtotal.Add(iva)
	if err == nil {
		gross, err = gross.Sub(r.ISR)
	}
	if err == nil {
		gross, err = gross.Sub(r.IVA)
	}
	if err != nil {
		return 0, fmt.Errorf("withholding net: %w: %w", ErrInvalidParameter, err)
	}
	return gross, nil
}
