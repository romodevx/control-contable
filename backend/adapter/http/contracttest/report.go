package contracttest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"

	liberrors "github.com/pb33f/libopenapi-validator/errors"
)

// report renders a failed validation so the reader sees what the contract
// expects next to what the server sent:
//
//	GET /healthz (getHealth): response 200 does not conform to docs/api/openapi.yaml
//	  declared: 200, 429, 500, 503
//	  - Response body for 'GET /healthz' failed to validate schema
//	      $.status: value must be one of 'ok', 'degraded'
//	got:
//	  HTTP 200
//	  Content-Type: application/json
//	  {
//	    "status": "great", ...
func report(req *http.Request, r route, rec *httptest.ResponseRecorder, errs []*liberrors.ValidationError) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s (%s): response %d does not conform to %s\n",
		req.Method, r.path, r.operation.OperationId, rec.Code, SpecPath)
	fmt.Fprintf(&b, "  declared: %s\n", strings.Join(declaredStatuses(r), ", "))
	for _, e := range errs {
		writeViolation(&b, e)
	}
	b.WriteString("got:\n")
	fmt.Fprintf(&b, "  HTTP %d\n", rec.Code)
	writeHeaders(&b, rec.Header())
	writeBody(&b, rec.Body.Bytes())
	return b.String()
}

func declaredStatuses(r route) []string {
	if r.operation.Responses == nil {
		return nil
	}
	var codes []string
	if r.operation.Responses.Codes != nil {
		for code := range r.operation.Responses.Codes.KeysFromOldest() {
			codes = append(codes, code)
		}
	}
	slices.Sort(codes)
	if r.operation.Responses.Default != nil {
		codes = append(codes, "default")
	}
	return codes
}

func writeViolation(b *strings.Builder, e *liberrors.ValidationError) {
	fmt.Fprintf(b, "  - %s\n", e.Message)
	if e.Reason != "" && e.Reason != e.Message {
		fmt.Fprintf(b, "      %s\n", e.Reason)
	}
	for _, f := range e.SchemaValidationErrors {
		where := f.FieldPath
		if where == "" {
			where = f.KeywordLocation
		}
		fmt.Fprintf(b, "      %s: %s\n", where, f.Reason)
	}
}

func writeHeaders(b *strings.Builder, h http.Header) {
	names := make([]string, 0, len(h))
	for name := range h {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		fmt.Fprintf(b, "  %s: %s\n", name, strings.Join(h.Values(name), ", "))
	}
}

func writeBody(b *strings.Builder, body []byte) {
	if len(body) == 0 {
		b.WriteString("  (empty body)\n")
		return
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "  ", "  "); err == nil {
		body = pretty.Bytes()
	}
	fmt.Fprintf(b, "  %s\n", body)
}
