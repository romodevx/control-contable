package domain

import "github.com/romodevx/control-contable/backend/internal/kernel/fiscal"

// The incomes module's own error codes (MODULE-incomes §4). The shared codes
// (required, invalid, iva_amount_mismatch, ...) and the error model live in
// the fiscal kernel (ADR-0007).
var (
	// ErrNotForRegime: platform_activity sent for a regime other than 625.
	ErrNotForRegime = fiscal.ValidationError("not_for_regime")
	// ErrActivityNotHeld: a 625 activity the taxpayer does not hold.
	ErrActivityNotHeld = fiscal.BusinessRuleError("activity_not_held")
)
