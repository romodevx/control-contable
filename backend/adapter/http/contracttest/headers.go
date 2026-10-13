package contracttest

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	liberrors "github.com/pb33f/libopenapi-validator/errors"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// transportHeaders are the only response headers a contract test accepts
// without the response declaring them. Each is outside the API's semantics:
//   - Content-Type: checked against the response's declared media types.
//   - Content-Length, Date: HTTP framing, set by net/http.
//   - X-Request-Id: set on every response by the request-id middleware; the
//     error envelope already carries it as request_id, which the contract
//     declares.
//   - X-Content-Type-Options: nosniff, set by httperr on every error envelope;
//     a browser-hardening directive no API client reads.
//
// Every other header (Cache-Control, ETag, Vary, Set-Cookie, ...) changes what
// a client does, so it must be declared by the response that sends it.
// libopenapi-validator's strict mode is not relied on for headers: it skips
// responses that declare none and ignores a long default list (Set-Cookie,
// Cache-Control, ETag, Vary, ...).
var transportHeaders = []string{
	"content-type", "content-length", "date", "x-request-id", "x-content-type-options",
}

// undeclaredHeaders reports each response header that neither the matched
// response (by status, else default) declares nor transportHeaders lists.
func undeclaredHeaders(op *v3.Operation, status int, got http.Header) []*liberrors.ValidationError {
	declared := declaredHeaders(op, status)
	var names []string
	for name := range got {
		lower := strings.ToLower(name)
		if !slices.Contains(transportHeaders, lower) && !slices.Contains(declared, lower) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	errs := make([]*liberrors.ValidationError, 0, len(names))
	for _, name := range names {
		errs = append(errs, &liberrors.ValidationError{
			ValidationType:    liberrors.StrictValidationType,
			ValidationSubType: liberrors.StrictSubTypeHeader,
			Message:           fmt.Sprintf("response header '%s' is not declared by response %d", name, status),
			Reason:            fmt.Sprintf("declared: [%s]", strings.Join(declared, ", ")),
		})
	}
	return errs
}

// declaredHeaders are the lower-cased header names the response for status
// declares.
func declaredHeaders(op *v3.Operation, status int) []string {
	if op.Responses == nil {
		return nil
	}
	resp := op.Responses.FindResponseByCode(status)
	if resp == nil {
		resp = op.Responses.Default
	}
	if resp == nil || resp.Headers == nil {
		return nil
	}
	var names []string
	for name := range resp.Headers.KeysFromOldest() {
		names = append(names, strings.ToLower(name))
	}
	return names
}

// withoutLibraryHeaderChecks drops libopenapi-validator's own undeclared
// header findings; undeclaredHeaders is the single, stricter source.
func withoutLibraryHeaderChecks(errs []*liberrors.ValidationError) []*liberrors.ValidationError {
	return slices.DeleteFunc(errs, func(e *liberrors.ValidationError) bool {
		return e.ValidationType == liberrors.StrictValidationType && e.ValidationSubType == liberrors.StrictSubTypeHeader
	})
}
