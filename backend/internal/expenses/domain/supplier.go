package domain

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// SupplierKind is "¿Quién te vendió o te dio el servicio?"
// (MODULE-withholdings §2.2 E1, contract SupplierKind).
type SupplierKind string

// The contract's SupplierKind enum.
const (
	SupplierPM             SupplierKind = "pm"
	SupplierPFWithActivity SupplierKind = "pf_with_activity"
	SupplierForeign        SupplierKind = "foreign"
)

// Valid reports a contract SupplierKind.
func (k SupplierKind) Valid() bool {
	return slices.Contains([]SupplierKind{SupplierPM, SupplierPFWithActivity, SupplierForeign}, k)
}

// SupplierRegime is a persona física supplier's regime as its CFDI prints it
// (E2, contract SupplierRegime); SupplierRegimeUnknown is "No sé".
type SupplierRegime string

// The contract's SupplierRegime enum.
const (
	SupplierRegime612     SupplierRegime = "612"
	SupplierRegime606     SupplierRegime = "606"
	SupplierRegime626     SupplierRegime = "626"
	SupplierRegime625     SupplierRegime = "625"
	SupplierRegime621     SupplierRegime = "621"
	SupplierRegimeUnknown SupplierRegime = "unknown"
)

// Valid reports a contract SupplierRegime.
func (r SupplierRegime) Valid() bool {
	return slices.Contains([]SupplierRegime{
		SupplierRegime612, SupplierRegime606, SupplierRegime626, SupplierRegime625, SupplierRegime621, SupplierRegimeUnknown,
	}, r)
}

// assumedRegime is what "No sé" derives as: the most common persona física
// regime billing a company (MODULE-withholdings §2.2).
const assumedRegime = fiscal.RegimeBusiness

// ParseSupplierKind returns the SupplierKind spelled exactly raw, else
// fiscal.ErrInvalid.
func ParseSupplierKind(raw string) (SupplierKind, error) { return fiscal.ParseEnum[SupplierKind](raw) }

// ParseSupplierRegime returns the SupplierRegime spelled exactly raw, else
// fiscal.ErrInvalid.
func ParseSupplierRegime(raw string) (SupplierRegime, error) {
	return fiscal.ParseEnum[SupplierRegime](raw)
}

// Supplier is what a persona moral says about whoever billed the expense
// (E1–E4). A zero field is absent: the operation then defaults to the
// category's and an omitted regime of a persona física supplier is "No sé".
// A persona física taxpayer sends none of it.
type Supplier struct {
	Kind            SupplierKind
	Regime          SupplierRegime
	Operation       fiscal.OperationKind
	ResidentialRent bool
}

// fieldsSet lists the contract fields s carries, in contract order.
func (s Supplier) fieldsSet() []string {
	var fields []string
	if s.Kind != "" {
		fields = append(fields, FieldSupplierKind)
	}
	if s.Regime != "" {
		fields = append(fields, FieldSupplierRegime)
	}
	if s.Operation != "" {
		fields = append(fields, FieldOperationKind)
	}
	if s.ResidentialRent {
		fields = append(fields, FieldResidentialRent)
	}
	return fields
}

// checkSupplier is the schema-level pass over the supplier facts: known
// values, and a persona moral must say who the supplier is.
func checkSupplier(vs *fiscal.Violations, s Supplier, p Payer) {
	if s.Kind == "" && p.PersonType == fiscal.PersonMoral {
		vs.Add(FieldSupplierKind, fiscal.ErrRequired)
	}
	if s.Kind != "" && !s.Kind.Valid() {
		vs.Add(FieldSupplierKind, fiscal.ErrInvalid)
	}
	if s.Regime != "" && !s.Regime.Valid() {
		vs.Add(FieldSupplierRegime, fiscal.ErrInvalid)
	}
	if s.Operation != "" && (!s.Operation.Valid() || s.Operation == fiscal.OperationPlatform) {
		vs.Add(FieldOperationKind, fiscal.ErrInvalid)
	}
}

// supplierRules is the business pass over the supplier facts.
func supplierRules(vs *fiscal.Violations, d Draft, p Payer) {
	if p.PersonType != fiscal.PersonMoral {
		for _, field := range d.Supplier.fieldsSet() {
			vs.Add(field, ErrSupplierNotApplicable)
		}
		return
	}
	if d.Supplier.Regime != "" && d.Supplier.Kind != SupplierPFWithActivity {
		vs.Add(FieldSupplierRegime, ErrSupplierNotApplicable)
	}
	if d.Supplier.ResidentialRent && d.IVARate != fiscal.IVAExempt {
		vs.Add(FieldIVARate, ErrResidentialRentExempt)
	}
}

// resolve fills the defaults a persona moral's omitted facts take: the
// category's operation and "No sé" for a persona física supplier.
func (s Supplier) resolve(p Payer, rule CategoryRule) Supplier {
	if p.PersonType != fiscal.PersonMoral {
		return Supplier{}
	}
	if s.Operation == "" {
		s.Operation = rule.DefaultOperation
	}
	if s.Operation == "" {
		s.Operation = fiscal.OperationBusinessServices
	}
	if s.Kind == SupplierPFWithActivity && s.Regime == "" {
		s.Regime = SupplierRegimeUnknown
	}
	return s
}

// withholdingInput is the kernel's view of the expense: the supplier is the
// emisor, the taxpayer (a persona moral) the receptor. A persona física
// taxpayer never withholds, which the table expresses as a persona física
// receptor; a foreign supplier is out of scope (WM A13) and reads as a
// foreign counterparty, which derives zero.
func withholdingInput(d Draft, p Payer, s Supplier) fiscal.WithholdingInput {
	in := fiscal.WithholdingInput{
		EmisorPersona: fiscal.PersonMoral, Receptor: fiscal.ReceptorPM, Operation: s.Operation,
		ResidentialRent: s.ResidentialRent, Subtotal: d.Amount, IVA: d.IVA, IVARate: d.IVARate, On: d.PaidOn,
	}
	switch {
	case p.PersonType != fiscal.PersonMoral:
		in.Receptor, in.Operation = fiscal.ReceptorPFWithActivity, fiscal.OperationOther
	case s.Kind == SupplierForeign:
		in.Receptor = fiscal.ReceptorForeign
	case s.Kind == SupplierPFWithActivity:
		in.EmisorPersona, in.EmisorRegime = fiscal.PersonFisica, fiscal.RegimeCode(s.Regime)
		if s.Regime == SupplierRegimeUnknown {
			in.EmisorRegime = assumedRegime
		}
	}
	return in
}

// derive runs the kernel's decision table. Its own validation failures
// cannot follow this package's (the facts were checked), so any is a server
// fault, never a client error.
func derive(d Draft, p Payer, s Supplier, catalog fiscal.WithholdingCatalog) (fiscal.Derivation, error) {
	derivation, err := fiscal.DeriveWithholdings(withholdingInput(d, p, s), catalog)
	if vs, ok := errors.AsType[fiscal.Violations](err); ok {
		return fiscal.Derivation{}, fmt.Errorf("derive withholdings: %s: %w", vs.Error(), fiscal.ErrInvalidParameter)
	}
	return derivation, err
}

// remitDay is the day of the month after the payment by which a persona
// moral remits what it withheld (RMF 4.1.11, LISR 96, LIVA 1-A).
const remitDay = 17

// RemitDueOn is the 17th of the month after paidOn.
func RemitDueOn(paidOn fiscal.Date) fiscal.Date {
	paid, err := time.Parse(time.DateOnly, paidOn.String())
	if err != nil {
		return fiscal.Date{}
	}
	due := time.Date(paid.Year(), paid.Month()+1, remitDay, 0, 0, 0, 0, time.UTC)
	out, err := fiscal.ParseDate(due.Format(time.DateOnly))
	if err != nil {
		return fiscal.Date{}
	}
	return out
}

// RemittanceDue is the due date of a withholding the taxpayer must remit,
// or the zero Date when there is none.
func RemittanceDue(r fiscal.WithholdingRecord, paidOn fiscal.Date) fiscal.Date {
	if r.Withholder != fiscal.WithholderReceptor || r.ISR+r.IVA == 0 {
		return fiscal.Date{}
	}
	return RemitDueOn(paidOn)
}

// withholdingWarnings are the warnings of the derivation and the facts.
func withholdingWarnings(d fiscal.Derivation, s Supplier, due fiscal.Date) []Warning {
	var warnings []Warning
	if s.Kind == SupplierPFWithActivity && s.Regime == SupplierRegimeUnknown {
		warnings = append(warnings, Warning{Code: WarningSupplierRegimeUnknown, Field: FieldSupplierRegime})
	}
	for _, code := range d.Warnings {
		switch code {
		case fiscal.WarningForeignCounterparty:
			warnings = append(warnings, Warning{Code: WarningForeignSupplier, Field: FieldSupplierKind})
		case fiscal.WarningCPAUnverified:
			warnings = append(warnings, Warning{Code: WarningCPAUnverified, Field: FieldWithholding})
		}
	}
	if !due.IsZero() {
		warnings = append(warnings, Warning{Code: WarningDeductionRequiresRemittance, Field: FieldWithholding, DueOn: due})
	}
	return warnings
}

// CreditableSplit splits a creditable expense's IVA into the share paid to
// the supplier (credited in the payment month) and the share withheld
// (credited the month after its remittance, LIVA 5 fr. IV). OQ-W5: the
// all-or-nothing creditability of the expense applies to both parts.
func CreditableSplit(creditable bool, iva, withheld money.Cents) (paid, withheldPart money.Cents) {
	if !creditable {
		return 0, 0
	}
	return iva - withheld, withheld
}

// LegacySupplier is the supplier a persona moral's expense stored before
// WH-07 is assumed to have (MODULE-withholdings §5.2): a typed withholding
// means a persona física with activity in 612, none a persona moral. The
// operation is left to the category's default.
func LegacySupplier(withheld money.Cents) Supplier {
	if withheld > 0 {
		return Supplier{Kind: SupplierPFWithActivity, Regime: SupplierRegime612}
	}
	return Supplier{Kind: SupplierPM}
}
