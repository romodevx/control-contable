package router

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// Field names and codes of the contract's FieldError. Messages are generic
// on purpose: validator text quotes the rejected value, which may be PII.
const (
	fieldBody    = "body"
	fieldRequest = "request"

	codeRequired = "required"
	codeInvalid  = "invalid"
	codeTooLarge = "too_large"

	messageRequired = "This field is required."
	messageInvalid  = "This value is not valid."
)

var (
	// leafPattern reads a santhosh-tekuri/jsonschema leaf error, which
	// kin-openapi uses for OAS 3.1: "at '<JSON pointer>': <message>".
	leafPattern = regexp.MustCompile(`at '([^']*)': (.*)$`)
	// quotedNames reads the property names of "missing properties 'a', 'b'".
	quotedNames = regexp.MustCompile(`'([^']+)'`)
	// keywordPrefix reads "minLength: got 0, want 1" style messages.
	keywordPrefix = regexp.MustCompile(`^([A-Za-z]+): `)
)

// fieldErrors turns a kin-openapi request validation error into the
// contract's FieldError list, in the order the validator reported them.
func fieldErrors(err error) []apperr.FieldError {
	// A direct assertion, not errors.As: a schema error nests its own
	// MultiError deeper in the chain, which must not be taken for this level.
	if multi, ok := err.(openapi3.MultiError); ok { //nolint:errorlint // see above

		var out []apperr.FieldError
		for _, e := range multi {
			out = append(out, fieldErrors(e)...)
		}
		return out
	}
	var reqErr *openapi3filter.RequestError
	if !errors.As(err, &reqErr) {
		return []apperr.FieldError{fieldError(fieldRequest, codeInvalid)}
	}
	if reqErr.Parameter != nil {
		return parameterErrors(reqErr.Parameter.Name, reqErr.Err)
	}
	return bodyErrors(reqErr.Err)
}

func parameterErrors(name string, err error) []apperr.FieldError {
	if errors.Is(err, openapi3filter.ErrInvalidRequired) {
		return []apperr.FieldError{fieldError(name, codeRequired)}
	}
	leaves := schemaLeaves(err)
	if len(leaves) == 0 {
		return []apperr.FieldError{fieldError(name, codeInvalid)}
	}
	out := make([]apperr.FieldError, 0, len(leaves))
	for _, leaf := range leaves {
		_, code := classify(leaf)
		out = append(out, fieldError(name, code))
	}
	return out
}

func bodyErrors(err error) []apperr.FieldError {
	var tooLarge *http.MaxBytesError
	switch {
	case errors.Is(err, openapi3filter.ErrInvalidRequired):
		return []apperr.FieldError{fieldError(fieldBody, codeRequired)}
	case errors.As(err, &tooLarge):
		return []apperr.FieldError{fieldError(fieldBody, codeTooLarge)}
	}
	leaves := schemaLeaves(err)
	if len(leaves) == 0 {
		return []apperr.FieldError{fieldError(fieldBody, codeInvalid)}
	}
	var out []apperr.FieldError
	for _, leaf := range leaves {
		fields, code := classify(leaf)
		for _, f := range fields {
			out = append(out, fieldError(f, code))
		}
	}
	return out
}

// schemaLeaves returns the innermost schema errors under err.
func schemaLeaves(err error) []*openapi3.SchemaError {
	// The legacy validator hands back a bare MultiError of siblings; errors.As
	// would stop at the first one, so each sibling is walked on its own.
	if multi, ok := err.(openapi3.MultiError); ok { //nolint:errorlint // see above
		var leaves []*openapi3.SchemaError
		for _, e := range multi {
			leaves = append(leaves, schemaLeaves(e)...)
		}
		return leaves
	}
	var schemaErr *openapi3.SchemaError
	if !errors.As(err, &schemaErr) {
		return nil
	}
	var nested openapi3.MultiError
	if schemaErr.Origin == nil || !errors.As(schemaErr.Origin, &nested) {
		return []*openapi3.SchemaError{schemaErr}
	}
	var leaves []*openapi3.SchemaError
	for _, e := range nested {
		leaves = append(leaves, schemaLeaves(e)...)
	}
	return leaves
}

// classify reads a leaf into the offending field paths (dotted, relative to
// the body; "body" for the root) and a contract code named after the JSON
// Schema keyword that failed.
//
// kin-openapi validates with two engines: the JSON Schema 2020-12 one, whose
// leaves only carry "at '<pointer>': <message>" text, and its legacy
// validator, which it falls back to when the schema reaches a $ref. Legacy
// leaves carry the pointer and the keyword as structured fields instead.
func classify(leaf *openapi3.SchemaError) (fields []string, code string) {
	m := leafPattern.FindStringSubmatch(leaf.Reason)
	if m == nil {
		return classifyLegacy(leaf)
	}
	pointer, message := m[1], m[2]
	switch {
	case strings.HasPrefix(message, "missing propert"):
		return childFields(pointer, message), codeRequired
	case strings.HasPrefix(message, "additional propert"):
		return childFields(pointer, message), "additionalProperties"
	case strings.HasPrefix(message, "got "):
		return []string{dotted(pointer)}, "type"
	case strings.Contains(message, "does not match pattern"):
		return []string{dotted(pointer)}, "pattern"
	case strings.HasPrefix(message, "value must be one of"):
		return []string{dotted(pointer)}, "enum"
	case strings.HasPrefix(message, "value must be"):
		return []string{dotted(pointer)}, "const"
	}
	if k := keywordPrefix.FindStringSubmatch(message); k != nil {
		return []string{dotted(pointer)}, k[1]
	}
	return []string{dotted(pointer)}, codeInvalid
}

// legacyCodes maps a legacy SchemaField to the contract code: the keyword
// the JSON Schema 2020-12 engine reports for the same failure.
var legacyCodes = map[string]string{
	"required":         codeRequired,
	"properties":       "additionalProperties",
	"type":             "type",
	"nullable":         "type",
	"enum":             "enum",
	"const":            "const",
	"pattern":          "pattern",
	"format":           "format",
	"minLength":        "minLength",
	"maxLength":        "maxLength",
	"minItems":         "minItems",
	"maxItems":         "maxItems",
	"uniqueItems":      "uniqueItems",
	"minimum":          "minimum",
	"maximum":          "maximum",
	"exclusiveMinimum": "exclusiveMinimum",
	"exclusiveMaximum": "exclusiveMaximum",
	"multipleOf":       "multipleOf",
	"minProperties":    "minProperties",
	"maxProperties":    "maxProperties",
}

// unsupportedProperty reads the name out of the legacy
// `property "<name>" is unsupported` reason, whose pointer names the parent.
var unsupportedProperty = regexp.MustCompile(`^property "([^"]+)" is unsupported$`)

// classifyLegacy reads a leaf of kin-openapi's legacy validator.
func classifyLegacy(leaf *openapi3.SchemaError) (fields []string, code string) {
	code, known := legacyCodes[leaf.SchemaField]
	if !known {
		code = codeInvalid
	}
	pointer := ""
	if path := leaf.JSONPointer(); len(path) > 0 {
		pointer = "/" + strings.Join(path, "/")
	}
	if m := unsupportedProperty.FindStringSubmatch(leaf.Reason); m != nil {
		pointer += "/" + m[1]
	}
	return []string{dotted(pointer)}, code
}

// childFields names each property quoted in message under pointer.
func childFields(pointer, message string) []string {
	var fields []string
	for _, m := range quotedNames.FindAllStringSubmatch(message, -1) {
		fields = append(fields, dotted(pointer+"/"+m[1]))
	}
	if len(fields) == 0 {
		return []string{dotted(pointer)}
	}
	return fields
}

// dotted renders a JSON pointer as a dotted path ("/a/0/b" → "a.0.b").
func dotted(pointer string) string {
	path := strings.ReplaceAll(strings.TrimPrefix(pointer, "/"), "/", ".")
	if path == "" {
		return fieldBody
	}
	return path
}

// bindingFieldErrors names the parameter behind an error of the generated
// wrapper (path, query or header binding), the backstop for formats that
// kin-openapi does not assert under JSON Schema 2020-12 (uuid, date, email).
func bindingFieldErrors(err error) []apperr.FieldError {
	var (
		required      *gen.RequiredParamError
		requiredHdr   *gen.RequiredHeaderError
		invalidFormat *gen.InvalidParamFormatError
		unmarshaling  *gen.UnmarshalingParamError
		tooMany       *gen.TooManyValuesForParamError
		cookie        *gen.UnescapedCookieParamError
	)
	switch {
	case errors.As(err, &required):
		return []apperr.FieldError{fieldError(required.ParamName, codeRequired)}
	case errors.As(err, &requiredHdr):
		return []apperr.FieldError{fieldError(requiredHdr.ParamName, codeRequired)}
	case errors.As(err, &invalidFormat):
		return []apperr.FieldError{fieldError(invalidFormat.ParamName, "format")}
	case errors.As(err, &unmarshaling):
		return []apperr.FieldError{fieldError(unmarshaling.ParamName, "format")}
	case errors.As(err, &tooMany):
		return []apperr.FieldError{fieldError(tooMany.ParamName, codeInvalid)}
	case errors.As(err, &cookie):
		return []apperr.FieldError{fieldError(cookie.ParamName, codeInvalid)}
	}
	// The strict handler's body decoding (formats such as email or date).
	return []apperr.FieldError{fieldError(fieldBody, "format")}
}

func fieldError(field, code string) apperr.FieldError {
	message := messageInvalid
	if code == codeRequired {
		message = messageRequired
	}
	return apperr.FieldError{Field: field, Code: code, Message: message}
}
