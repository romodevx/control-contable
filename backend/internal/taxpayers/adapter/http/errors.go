// Package taxpayershttp is the taxpayers module's HTTP adapter: it
// translates between the contract (docs/api/openapi.yaml) and the taxpayers
// use cases.
package taxpayershttp

import (
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// Client-safe messages. They never echo the input, an option key or an RFC.
const (
	msgInvalidPersonType  = "Choose persona física or persona moral."
	msgRFCMalformed       = "Enter an RFC of 13 characters (persona física) or 12 (persona moral)."
	msgRFCInvalidDate     = "The date in the RFC is not a valid date."
	msgRFCCheckDigit      = "The RFC is not valid. Check its last character."
	msgRFCGeneric         = "A generic RFC cannot be registered as a taxpayer."
	msgPersonTypeMismatch = "The RFC does not match the person type."
	msgInvalidOptions     = "One or more regime options are invalid."
)

// ErrorMapper maps taxpayers domain errors to contract error codes.
// ErrInvalidOptionDef has no rule: a broken catalog is a server fault and
// resolves to internal_error, as does any infrastructure failure.
var ErrorMapper = apperr.NewMapper(
	apperr.Rule{Target: domain.ErrInvalidPersonType, Code: apperr.CodeValidationFailed, Message: msgInvalidPersonType},
	apperr.Rule{Target: domain.ErrRFCMalformed, Code: apperr.CodeValidationFailed, Message: msgRFCMalformed},
	apperr.Rule{Target: domain.ErrRFCInvalidDate, Code: apperr.CodeValidationFailed, Message: msgRFCInvalidDate},
	apperr.Rule{Target: domain.ErrRFCCheckDigit, Code: apperr.CodeValidationFailed, Message: msgRFCCheckDigit},
	apperr.Rule{Target: domain.ErrRFCGeneric, Code: apperr.CodeValidationFailed, Message: msgRFCGeneric},
	apperr.Rule{Target: domain.ErrPersonTypeMismatch, Code: apperr.CodeValidationFailed, Message: msgPersonTypeMismatch},
	apperr.Rule{Target: domain.ErrOptionUnknown, Code: apperr.CodeValidationFailed, Message: msgInvalidOptions},
	apperr.Rule{Target: domain.ErrOptionMissing, Code: apperr.CodeValidationFailed, Message: msgInvalidOptions},
	apperr.Rule{Target: domain.ErrOptionType, Code: apperr.CodeValidationFailed, Message: msgInvalidOptions},
	apperr.Rule{Target: domain.ErrOptionValue, Code: apperr.CodeValidationFailed, Message: msgInvalidOptions},
)
