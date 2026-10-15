package domain

import (
	"errors"
	"fmt"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Contract field names, as violations and warnings report them.
const (
	FieldAmountCents      = "amount_cents"
	FieldIVARate          = "iva_rate"
	FieldIVACents         = "iva_cents"
	FieldCounterpartyKind = "counterparty_kind"
	FieldOperationKind    = "operation_kind"
	FieldResidentialRent  = "residential_rent"
	FieldViaPlatform      = "via_platform"
	FieldWithholding      = "withholding"
	FieldInvoicedOn       = "invoiced_on"
	FieldCollectedOn      = "collected_on"
	FieldPlatformActivity = "platform_activity"
	FieldCFDIUUID         = "cfdi_uuid"
)

// WarningCode is a non-blocking notice (contract Warning.code, MODULE-incomes
// §2.8, MODULE-withholdings §6). Codes are append-only.
type WarningCode string

// The income warnings. DuplicateCFDIUUID needs the stored incomes and
// ProvisionalParameters the parameter set a date reads, so the use case
// adds them. withholding_differs_from_rate is retired (ADR-0008): no typed
// amount is left to compare.
const (
	WarningDuplicateCFDIUUID WarningCode = "duplicate_cfdi_uuid"
	// WarningProvisionalParameters: the rates checked on a date of the
	// income come from a provisional parameter set (the year is not
	// published yet, or the set is marked provisional).
	WarningProvisionalParameters WarningCode = "provisional_parameters"
	// WarningWithholdingCPAUnverified: the derivation relied on a rule a
	// CPA has not confirmed (⚠️ WM §6).
	WarningWithholdingCPAUnverified WarningCode = "withholding_cpa_unverified"
	// WarningForeignCounterpartyNotSupported: billed abroad; taxes the app
	// does not compute may apply.
	WarningForeignCounterpartyNotSupported WarningCode = "foreign_counterparty_not_supported"
)

// kernelWarnings place the derivation's warnings on their contract field.
var kernelWarnings = map[fiscal.WarningCode]Warning{
	fiscal.WarningCPAUnverified:       {Code: WarningWithholdingCPAUnverified, Field: FieldWithholding},
	fiscal.WarningForeignCounterparty: {Code: WarningForeignCounterpartyNotSupported, Field: FieldCounterpartyKind},
}

// Warning is a notice on one field. It never echoes an amount, a description
// or a CFDI UUID.
type Warning struct {
	Code  WarningCode
	Field string
}

// Draft is an income as the user entered it, after the adapter decoded the
// request and the use case resolved the defaults of the hidden questions
// (MODULE-withholdings §2.1). Zero values are absent fields: a zero Date is
// null, an empty PlatformActivity lets a single 625 activity be the
// default.
type Draft struct {
	Amount           money.Cents // subtotal before IVA, MXN
	IVARate          fiscal.IVARate
	IVA              money.Cents
	Counterparty     fiscal.ReceptorKind  // Q1 "¿A quién le facturaste?"
	Operation        fiscal.OperationKind // Q2, never OperationPlatform (that is ViaPlatform)
	ResidentialRent  bool                 // Q3
	ViaPlatform      bool                 // Q4
	Dates            Dates
	PlatformActivity PlatformActivity
}

// Assessment is a valid income's derived facts: what the use case stores
// besides the draft (booked_on, the resolved 625 activity, the derived
// withholding) and what the response shows (pending_collection, warnings).
type Assessment struct {
	BookedOn          fiscal.Date
	PendingCollection bool
	PlatformActivity  PlatformActivity
	Withholding       fiscal.WithholdingRecord
	Warnings          []Warning
}

// Assess applies MODULE-incomes §2 and MODULE-withholdings §3 to a draft for
// payee, with today the current Mexico City day (Clock port) and catalog the
// dated parameters.
//
// Validation (validation_failed) runs first and reports every field; only a
// draft that passes it meets the business rules (business_rule_violation),
// which are also all reported. Both come back as fiscal.Violations. Then the
// withholdings are derived (fiscal.DeriveWithholdings): there is no typed
// amount (ADR-0008). A parameter fault is fiscal.ErrInvalidParameter. The
// regime window (CheckRegime), the idempotency key and the trial cap belong
// to the use case.
func Assess(d Draft, p Payee, today fiscal.Date, catalog fiscal.WithholdingCatalog) (Assessment, error) {
	if err := checkContext(p, today, catalog); err != nil {
		return Assessment{}, err
	}
	activity, vs := validate(d, p, today)
	if len(vs) > 0 {
		return Assessment{}, vs
	}
	return applyRules(d, p, activity, catalog)
}

// Derive applies the withholding decision table to a valid draft for payee
// on the income's IVA date (collected_on, else invoiced_on: the payer
// withholds when it pays, MODULE-withholdings §4.1). activity is the
// resolved 625 activity.
func Derive(d Draft, p Payee, activity PlatformActivity, catalog fiscal.WithholdingCatalog) (fiscal.Derivation, error) {
	operation := d.Operation
	if d.ViaPlatform {
		operation = fiscal.OperationPlatform
	}
	derivation, err := fiscal.DeriveWithholdings(fiscal.WithholdingInput{
		EmisorPersona: p.PersonType, EmisorRegime: p.Regime, Receptor: d.Counterparty, Operation: operation,
		PlatformActivity:   fiscal.PlatformActivity(activity),
		PlatformWithoutRFC: d.ViaPlatform && p.Regime == fiscal.RegimePlatforms && !p.RFCGivenToPlatform,
		ResidentialRent:    d.ResidentialRent, Subtotal: d.Amount, IVA: d.IVA, IVARate: d.IVARate,
		On: d.Dates.RateDate(),
	}, catalog)
	if vs, ok := errors.AsType[fiscal.Violations](err); ok {
		// The draft was validated: a kernel violation is the server's fault.
		return fiscal.Derivation{}, fmt.Errorf("derive withholdings: %s: %w", vs.Error(), fiscal.ErrInvalidParameter)
	}
	if errors.Is(err, fiscal.ErrRateNotValidOnDate) {
		// The boot requires every withholding parameter (ratecatalog): a gap
		// on a covered date is a parameter fault, not the client's.
		return fiscal.Derivation{}, fmt.Errorf("derive withholdings: %s: %w", err.Error(), fiscal.ErrInvalidParameter)
	}
	return derivation, err
}

// DefaultOperation is Q2's first-time default by regime
// (MODULE-withholdings §2.1): any persona moral sells business services.
func DefaultOperation(person fiscal.PersonType, regime fiscal.RegimeCode) fiscal.OperationKind {
	if person == fiscal.PersonMoral {
		return fiscal.OperationBusinessServices
	}
	if op, ok := regimeOperations[regime]; ok {
		return op
	}
	return fiscal.OperationOther
}

// regimeOperations are the persona física regimes with a Q2 default.
var regimeOperations = map[fiscal.RegimeCode]fiscal.OperationKind{
	fiscal.RegimeBusiness: fiscal.OperationProfessionalServices,
	fiscal.RegimeRent:     fiscal.OperationRentRealEstate,
	fiscal.RegimeRESICO:   fiscal.OperationProfessionalServices,
	regimeRIF:             fiscal.OperationSaleOfGoods,
}

// regimeRIF is the Régimen de Incorporación Fiscal (621).
const regimeRIF fiscal.RegimeCode = "621"

// DefaultViaPlatform is Q4's default: yes in 625, no elsewhere.
func DefaultViaPlatform(regime fiscal.RegimeCode) bool { return regime == fiscal.RegimePlatforms }

// checkContext rejects server-side inputs the rules cannot use.
func checkContext(p Payee, today fiscal.Date, catalog fiscal.RateCatalog) error {
	switch {
	case !p.Basis.Valid():
		return fmt.Errorf("income basis %q: %w", p.Basis, fiscal.ErrInvalidParameter)
	case !p.PersonType.Valid():
		return fmt.Errorf("person type %q: %w", p.PersonType, fiscal.ErrInvalidParameter)
	case today.IsZero():
		return fmt.Errorf("no current date: %w", fiscal.ErrInvalidParameter)
	case catalog == nil:
		return fmt.Errorf("no rate catalog: %w", fiscal.ErrInvalidParameter)
	}
	return nil
}

// validate is the schema-level pass. It returns the effective 625 activity.
func validate(d Draft, p Payee, today fiscal.Date) (PlatformActivity, fiscal.Violations) {
	var vs fiscal.Violations
	checkAmount(&vs, d.Amount)
	fiscal.CheckEnum(&vs, FieldIVARate, d.IVARate)
	checkBounds(&vs, FieldIVACents, d.IVA)
	fiscal.CheckEnum(&vs, FieldCounterpartyKind, d.Counterparty)
	checkOperation(&vs, d.Operation)
	if d.ViaPlatform && p.Regime != fiscal.RegimePlatforms && p.PersonType != fiscal.PersonMoral {
		vs.Add(FieldViaPlatform, ErrNotForRegime)
	}
	checkDate(&vs, FieldInvoicedOn, d.Dates.InvoicedOn, p.Basis, today)
	checkDate(&vs, FieldCollectedOn, d.Dates.CollectedOn, p.Basis, today)
	activity := resolveActivity(&vs, d.PlatformActivity, p)
	return activity, vs
}

// checkOperation accepts the contract's OperationKind values: a platform
// sale is via_platform, never an operation kind on the wire.
func checkOperation(vs *fiscal.Violations, op fiscal.OperationKind) {
	if op == fiscal.OperationPlatform {
		vs.Add(FieldOperationKind, fiscal.ErrInvalid)
		return
	}
	fiscal.CheckEnum(vs, FieldOperationKind, op)
}

func checkAmount(vs *fiscal.Violations, amount money.Cents) {
	switch {
	case amount < 1:
		vs.Add(FieldAmountCents, fiscal.ErrMustBePositive)
	case amount > fiscal.MaxCents:
		vs.Add(FieldAmountCents, fiscal.ErrOutOfRange)
	}
}

func checkBounds(vs *fiscal.Violations, field string, c money.Cents) {
	if c < 0 || c > fiscal.MaxCents {
		vs.Add(field, fiscal.ErrOutOfRange)
	}
}

// checkDate requires the basis's date and keeps every date within
// [fiscal.MinRecordDate, today]: future dates are refused.
func checkDate(vs *fiscal.Violations, field string, on fiscal.Date, basis IncomeBasis, today fiscal.Date) {
	switch {
	case on.IsZero() && field == basis.RequiredDateField():
		vs.Add(field, fiscal.ErrRequired)
	case on.IsZero():
	case on.Before(fiscal.MinRecordDate) || on.After(today):
		vs.Add(field, fiscal.ErrOutOfRange)
	}
}

// resolveActivity: platform_activity belongs to 625 only, defaults to the
// taxpayer's activity when they hold exactly one, else is required.
func resolveActivity(vs *fiscal.Violations, activity PlatformActivity, p Payee) PlatformActivity {
	if p.Regime != fiscal.RegimePlatforms {
		if activity != "" {
			vs.Add(FieldPlatformActivity, ErrNotForRegime)
		}
		return ""
	}
	if activity == "" && len(p.PlatformActivities) == 1 {
		return p.PlatformActivities[0]
	}
	fiscal.CheckEnum(vs, FieldPlatformActivity, activity)
	return activity
}

// applyRules is the business pass over a validated draft, then the
// derivation.
func applyRules(d Draft, p Payee, activity PlatformActivity, catalog fiscal.WithholdingCatalog) (Assessment, error) {
	var vs fiscal.Violations
	if err := checkIVAInto(&vs, d, catalog); err != nil {
		return Assessment{}, err
	}
	if d.ResidentialRent && d.IVARate != fiscal.IVAExempt {
		vs.Add(FieldIVARate, ErrResidentialRentExempt)
	}
	if activity != "" && !slices.Contains(p.PlatformActivities, activity) {
		vs.Add(FieldPlatformActivity, ErrActivityNotHeld)
	}
	if len(vs) > 0 {
		return Assessment{}, vs
	}
	derivation, err := Derive(d, p, activity, catalog)
	if err != nil {
		return Assessment{}, err
	}
	return Assessment{
		BookedOn:          p.Basis.BookingDate(d.Dates),
		PendingCollection: p.Basis.PendingCollection(d.Dates),
		PlatformActivity:  activity,
		Withholding:       derivation.Record(),
		Warnings:          WarningsOf(derivation),
	}, nil
}

// WarningsOf places the derivation's warnings on their fields.
func WarningsOf(d fiscal.Derivation) []Warning {
	var out []Warning
	for _, code := range d.Warnings {
		if w, ok := kernelWarnings[code]; ok {
			out = append(out, w)
		}
	}
	return out
}

// checkIVAInto records the IVA rule's violation on its field; a parameter
// fault is returned instead.
func checkIVAInto(vs *fiscal.Violations, d Draft, catalog fiscal.RateCatalog) error {
	err := fiscal.CheckIVA(d.Amount, d.IVARate, d.IVA, d.Dates.RateDate(), catalog)
	switch {
	case errors.Is(err, fiscal.ErrRateNotValidOnDate):
		vs.Add(FieldIVARate, fiscal.ErrRateNotValidOnDate)
	case errors.Is(err, fiscal.ErrIVAAmountMismatch):
		vs.Add(FieldIVACents, fiscal.ErrIVAAmountMismatch)
	case err != nil:
		return err
	}
	return nil
}
