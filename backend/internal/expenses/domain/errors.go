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
)
