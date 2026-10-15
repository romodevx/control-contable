package domain

import "github.com/romodevx/control-contable/backend/internal/kernel/fiscal"

// The incomes module's own error codes (MODULE-incomes §4). The shared codes
// (required, invalid, iva_amount_mismatch, ...) and the error model live in
// the fiscal kernel (ADR-0007).
var (
	// ErrNotForRegime: platform_activity sent for a regime other than 625, or
	// via_platform for one other than 625 or a persona moral's.
	ErrNotForRegime = fiscal.ValidationError("not_for_regime")
	// ErrActivityNotHeld: a 625 activity the taxpayer does not hold.
	ErrActivityNotHeld = fiscal.BusinessRuleError("activity_not_held")
	// ErrResidentialRentExempt: housing rent with an IVA rate other than
	// exempt (LIVA 20 fr. II); the server refuses instead of changing the
	// rate (ADR-0008 §5).
	ErrResidentialRentExempt = fiscal.BusinessRuleError("residential_rent_exempt")
)
