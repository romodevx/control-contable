package usecase

import (
	"time"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Warning is a non-blocking notice on a successful write (contract
// Warning): the client asks the user to double-check Field. Never stored,
// never on reads, never carries the RFC.
type Warning struct {
	Code    string
	Field   string
	Message string
}

// Warning codes (contract Warning.code).
const (
	WarningRFCCheckDigitMismatch = "rfc_check_digit_mismatch"
	WarningRegimeNeedsCPAReview  = "regime_combination_needs_cpa_review"
)

// msgRFCCheckDigitMismatch is the es-MX text of the check-digit warning.
const msgRFCCheckDigitMismatch = "Revisa tu RFC; la homoclave no coincide con el dígito verificador."

// TaxpayerRegime is a regime as the contract shows it (TaxpayerRegime).
type TaxpayerRegime struct {
	Code      string
	Name      string
	Treatment string
	// Options carries every option of the regime, defaults included.
	Options map[string]domain.OptionValue
	Window  domain.Window
	// Warnings is set only on the response of the write that added the
	// regime.
	Warnings []Warning
}

// Taxpayer is a taxpayer as the contract shows it (Taxpayer). The full RFC
// is never part of it (S-3).
type Taxpayer struct {
	ID          string
	PersonType  domain.PersonType
	DisplayName string
	RFCMasked   string
	// Regimes is ordered by start period, then code.
	Regimes []TaxpayerRegime
	// ArchivedAt is the zero time while the taxpayer is active.
	ArchivedAt time.Time
	CreatedAt  time.Time
	// Warnings is set only on the create and update responses.
	Warnings []Warning
}

// RegimeInput is one regime of a create, or the body of addTaxpayerRegime.
type RegimeInput struct {
	Code string
	// Options are the values sent; omitted options with a default are
	// filled.
	Options map[string]domain.OptionValue
	// Start is "YYYY-MM", or "" for the current month.
	Start string
}

// CreateInput is the createTaxpayer body.
type CreateInput struct {
	// ID is the optional client-generated id (idempotent retries); "" lets
	// the server generate a uuidv7.
	ID          string
	PersonType  string
	RFC         string
	DisplayName string
	Regimes     []RegimeInput
}

// CreateResult is the outcome of Create.
type CreateResult struct {
	Taxpayer Taxpayer
	// Created is false for an idempotent retry (200, nothing written).
	Created bool
}

// UpdateInput is the updateTaxpayer body; nil fields are left unchanged.
type UpdateInput struct {
	DisplayName *string
	RFC         *string
}

// EndRegimeInput is the archiveTaxpayerRegime call.
type EndRegimeInput struct {
	Code string
	// End is "YYYY-MM".
	End string
	// Confirm acknowledges that the regime has records (they are kept).
	Confirm bool
}
