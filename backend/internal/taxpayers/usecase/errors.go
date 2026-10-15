package usecase

import (
	"errors"

	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

// Use case errors. Stores return the ones that describe stored state; the
// use cases return every client-facing failure as an *apperr.Error whose
// cause wraps one of these (or a domain error), so callers match with
// errors.Is and the HTTP edge answers the contract code as is.
var (
	// ErrTaxpayerNotFound: no such taxpayer for the caller (another user's
	// is the same answer, US-T5: 404, never 403).
	ErrTaxpayerNotFound = errors.New("taxpayers: taxpayer not found")
	// ErrTaxpayerArchived: a write on an archived taxpayer (422).
	ErrTaxpayerArchived = errors.New("taxpayers: taxpayer is archived")
	// ErrRFCAlreadyRegistered: the caller already holds this RFC (409).
	ErrRFCAlreadyRegistered = errors.New("taxpayers: rfc already registered")
	// ErrIDConflict: a client id taken outside the caller's scope (409,
	// generic: RLS hides the row, nothing about its owner leaks).
	ErrIDConflict = errors.New("taxpayers: id conflict")
	// ErrRegimeNotHeld: the taxpayer does not hold this regime (404).
	ErrRegimeNotHeld = errors.New("taxpayers: regime not held")
	// ErrRegimeAlreadyActive: adding a regime that is open (409).
	ErrRegimeAlreadyActive = errors.New("taxpayers: regime already active")
	// ErrRegimeAlreadyEnded: ending a regime that has ended (409).
	ErrRegimeAlreadyEnded = errors.New("taxpayers: regime already ended")
	// ErrRegimeNotForPersonType: the catalog has no such regime for the
	// taxpayer's person type (400).
	ErrRegimeNotForPersonType = errors.New("taxpayers: regime not in catalog for person type")
	// ErrNoRegimes: a create without regimes (400).
	ErrNoRegimes = errors.New("taxpayers: at least one regime is required")
	// ErrDuplicateRegime: the same code twice in one create (400).
	ErrDuplicateRegime = errors.New("taxpayers: duplicate regime")
	// ErrPeriodOutOfRange: a start or end period outside its range (400).
	ErrPeriodOutOfRange = errors.New("taxpayers: period out of range")
	// ErrIncompatibleRegimes: two overlapping regimes the catalog does not
	// allow together (422 invalid_regime_combination).
	ErrIncompatibleRegimes = errors.New("taxpayers: incompatible regimes")
	// ErrRegimeHasRecords: ending a regime with records without confirm
	// (409).
	ErrRegimeHasRecords = errors.New("taxpayers: regime has records")
	// ErrRecordsAfterEndPeriod: records after the requested end (422).
	ErrRecordsAfterEndPeriod = errors.New("taxpayers: records after end period")
	// ErrLastActiveRegime: ending the only open regime (422).
	ErrLastActiveRegime = errors.New("taxpayers: last active regime")
	// ErrRegimeNotActive: TaxpayerRegimeChecker found no window of the
	// regime holding the period (422).
	ErrRegimeNotActive = errors.New("taxpayers: regime not active in period")
	// ErrConfirmRequired: a delete without confirm=true (400).
	ErrConfirmRequired = errors.New("taxpayers: confirmation required")
	// ErrInvalidID: a client or path id that is not a UUID.
	ErrInvalidID = errors.New("taxpayers: invalid id")
)

// FieldError.code values of the taxpayers module (MODULE-taxpayers §4).
const (
	DetailRFCLengthMismatch     = "rfc_length_mismatch"
	DetailRFCFormat             = "rfc_format"
	DetailRFCDate               = "rfc_date"
	DetailRFCGeneric            = "rfc_generic"
	DetailNotForPersonType      = "not_for_person_type"
	DetailDuplicate             = "duplicate"
	DetailUnknownOption         = "unknown_option"
	DetailInvalidOption         = "invalid_option"
	DetailRequired              = "required"
	DetailOutOfRange            = "out_of_range"
	DetailInvalid               = "invalid"
	DetailInvisibleCharacters   = usertext.DetailInvisibleCharacters
	DetailConfirmRequired       = "confirm_required"
	DetailIncompatible          = "incompatible"
	DetailTaxpayerArchived      = "taxpayer_archived"
	DetailLastActiveRegime      = "last_active_regime"
	DetailRecordsAfterEndPeriod = "records_after_end_period"
	DetailRegimeNotActive       = "regime_not_active"
	fieldTaxpayerID             = "taxpayerId"
	fieldRegimeCodePath         = "regimeCode"
	fieldRFC                    = "rfc"
	fieldDisplayName            = "display_name"
	fieldPersonType             = "person_type"
	fieldConfirm                = "confirm"
	fieldEndPeriod              = "end_period"
	fieldStartPeriod            = "start_period"
	fieldRegimeCode             = "regime_code"
	fieldOptions                = "options"
	fieldID                     = "id"
	fieldPeriod                 = "period"
	fieldRegimesPrefix          = "regimes"
)
