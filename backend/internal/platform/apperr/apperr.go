// Package apperr is the application error model shared by use cases and the
// HTTP edge. An *Error carries a contract ErrorCode (docs/api/openapi.yaml)
// and a client-safe message; the original error stays wrapped for logging.
//
// The package is delivery-agnostic (no net/http): use cases may return an
// *Error directly, while domain packages keep their own sentinel errors and
// each module's adapter translates them with a Mapper. HTTP status mapping and
// the wire envelope live in platform/httperr.
package apperr

import (
	"errors"
	"slices"
)

// Code is a value of the ErrorCode enum in the OpenAPI contract.
type Code string

// The ErrorCode enum of docs/api/openapi.yaml. TestCodes_matchContract keeps
// this list and the contract identical.
const (
	CodeValidationFailed         Code = "validation_failed"
	CodeUnauthorized             Code = "unauthorized"
	CodeEntitlementRequired      Code = "entitlement_required"
	CodeNotFound                 Code = "not_found"
	CodeConflict                 Code = "conflict"
	CodeRFCAlreadyRegistered     Code = "rfc_already_registered"
	CodeRegimeHasRecords         Code = "regime_has_records"
	CodeInvalidRegimeCombination Code = "invalid_regime_combination"
	CodeBusinessRuleViolation    Code = "business_rule_violation"
	CodeTaxpayerLimitReached     Code = "taxpayer_limit_reached"
	CodeRateLimited              Code = "rate_limited"
	CodeUpstreamUnavailable      Code = "upstream_unavailable"
	CodeInternalError            Code = "internal_error"
)

// InternalMessage is the only message a client sees for an unmapped error.
const InternalMessage = "Something went wrong."

var allCodes = []Code{
	CodeValidationFailed,
	CodeUnauthorized,
	CodeEntitlementRequired,
	CodeNotFound,
	CodeConflict,
	CodeRFCAlreadyRegistered,
	CodeRegimeHasRecords,
	CodeInvalidRegimeCombination,
	CodeBusinessRuleViolation,
	CodeTaxpayerLimitReached,
	CodeRateLimited,
	CodeUpstreamUnavailable,
	CodeInternalError,
}

// Codes returns every contract error code, in declaration order.
func Codes() []Code { return slices.Clone(allCodes) }

// Valid reports whether c belongs to the contract's ErrorCode enum.
func (c Code) Valid() bool { return slices.Contains(allCodes, c) }

// FieldError points at one offending input field (contract FieldError).
type FieldError struct {
	Field   string
	Code    string
	Message string
}

// Error is an application error with a contract code and a client-safe
// message. Message is shown to users: it must never contain PII.
type Error struct {
	Code    Code
	Message string
	Fields  []FieldError
	cause   error
}

// New returns an *Error without an underlying cause.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap returns an *Error that keeps cause for errors.Is/As and logging.
func Wrap(cause error, code Code, message string) *Error {
	return &Error{Code: code, Message: message, cause: cause}
}

// Error implements error. The text includes the cause and is meant for logs,
// not for clients: clients only see Code and Message.
func (e *Error) Error() string {
	text := string(e.Code) + ": " + e.Message
	if e.cause != nil {
		text += ": " + e.cause.Error()
	}
	return text
}

// Unwrap returns the underlying cause, if any.
func (e *Error) Unwrap() error { return e.cause }

// WithFields returns a copy of e with fields appended; e is not modified.
func (e *Error) WithFields(fields ...FieldError) *Error {
	clone := *e
	clone.Fields = append(slices.Clone(e.Fields), fields...)
	return &clone
}

// Rule maps a domain error (matched with errors.Is) to a contract code.
type Rule struct {
	Target  error
	Code    Code
	Message string
}

// Mapper translates any error into an *Error. Each module's adapter builds
// one from its domain sentinels; the zero Mapper maps only typed errors.
type Mapper struct {
	rules []Rule
}

// NewMapper returns a Mapper that tries rules in order.
func NewMapper(rules ...Rule) Mapper {
	return Mapper{rules: slices.Clone(rules)}
}

// Resolve returns the *Error for err: a typed *Error anywhere in the chain
// wins, then the first matching Rule, else internal_error with a generic
// message, so an unexpected error never leaks its text to the client. The
// result always wraps err. Resolve(nil) is nil.
func (m Mapper) Resolve(err error) *Error {
	if err == nil {
		return nil
	}
	if typed, ok := errors.AsType[*Error](err); ok {
		if error(typed) == err { //nolint:errorlint // identity, not matching: is err the typed error itself?
			return typed
		}
		resolved := Wrap(err, typed.Code, typed.Message)
		resolved.Fields = slices.Clone(typed.Fields)
		return resolved
	}
	for _, rule := range m.rules {
		if errors.Is(err, rule.Target) {
			return Wrap(err, rule.Code, rule.Message)
		}
	}
	return Wrap(err, CodeInternalError, InternalMessage)
}
