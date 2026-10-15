package domain

import "github.com/romodevx/control-contable/backend/internal/kernel/fiscal"

// The expenses module's own error codes (MODULE-expenses §4). The shared
// codes (required, invalid, iva_amount_mismatch, ...) and the error model
// live in the fiscal kernel (ADR-0007).
var (
	// ErrUnknownCategory is the use case's: no category rule exists.
	ErrUnknownCategory = fiscal.ValidationError("unknown_category")
	// ErrHasCFDIMismatch: a cfdi_uuid sent with has_cfdi = false.
	ErrHasCFDIMismatch = fiscal.ValidationError("has_cfdi_mismatch")
	// ErrSupplierNotApplicable: a supplier fact from a persona física
	// taxpayer (she never withholds, WM F1, F2), or a supplier regime
	// without supplier_kind pf_with_activity.
	ErrSupplierNotApplicable = fiscal.BusinessRuleError("supplier_not_applicable")
	// ErrResidentialRentExempt: housing rent with an IVA rate other than
	// exempt (LIVA 20 fr. II); the server never changes the rate itself.
	ErrResidentialRentExempt = fiscal.BusinessRuleError("residential_rent_exempt")
)
