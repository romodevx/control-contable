package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

// notUserText are the free strings of request bodies that are not text a
// user types to be shown back, keyed "operationId:path", with the reason
// the invisible-character rule (ADR-0004 Amendment 1) does not apply.
var notUserText = map[string]string{ //nolint:gosec // G101: contract field names, not credentials
	"createExpense:category_code":                         "catalog code, checked against the catalog",
	"updateExpense:category_code":                         "catalog code, checked against the catalog",
	"createTaxpayer:rfc":                                  "own grammar, checked by the RFC rules",
	"updateTaxpayer:rfc":                                  "own grammar, checked by the RFC rules",
	"login:password":                                      "never displayed",
	"verifyEmail:password":                                "never displayed",
	"deleteMe:password":                                   "never displayed",
	"logout:refresh_token":                                "opaque token, compared by hash",
	"refreshSession:refresh_token":                        "opaque token, compared by hash",
	"register:accepted_documents[].version":               "checked against the published versions",
	"receiveRevenueCatWebhook:api_version":                "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.app_user_id":          "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.original_app_user_id": "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.entitlement_ids[]":    "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.id":                   "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.product_id":           "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.store":                "server-to-server, never displayed",
	"receiveRevenueCatWebhook:event.type":                 "server-to-server, never displayed",
}

// ADR-0004 Amendment 1: every free string a request body accepts (no enum,
// format, pattern or const) is user text that documents the
// invisible_characters rule, or is listed in notUserText with its reason.
// The list comes from the contract itself, so a new text field fails here
// until it takes a side.
func TestUserTextFields_documentTheInvisibleRule(t *testing.T) {
	fields := requestFreeStrings(t)
	require.NotEmpty(t, fields)
	for key, schema := range fields {
		if _, exempt := notUserText[key]; exempt {
			continue
		}
		assert.Contains(t, schema.Description, usertext.DetailInvisibleCharacters,
			"%s is user text: document the rule or list it in notUserText", key)
	}
	for key := range notUserText {
		assert.Contains(t, fields, key, "notUserText lists %s, which no request body has", key)
	}
}

// requestFreeStrings maps "operationId:path" to each free string schema of
// the contract's JSON request bodies.
func requestFreeStrings(t *testing.T) map[string]*openapi3.Schema {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	out := make(map[string]*openapi3.Schema)
	for _, path := range sortedKeys(spec.Paths.Map()) {
		for _, op := range spec.Paths.Value(path).Operations() {
			if op.RequestBody == nil || op.RequestBody.Value == nil {
				continue
			}
			for _, media := range op.RequestBody.Value.Content {
				collectFreeStrings(op.OperationID+":", "", media.Schema, out)
			}
		}
	}
	return out
}

func collectFreeStrings(prefix, path string, ref *openapi3.SchemaRef, out map[string]*openapi3.Schema) {
	if ref == nil || ref.Value == nil {
		return
	}
	s := ref.Value
	for _, group := range [][]*openapi3.SchemaRef{s.AllOf, s.OneOf, s.AnyOf} {
		for _, sub := range group {
			collectFreeStrings(prefix, path, sub, out)
		}
	}
	if isFreeString(s) {
		out[prefix+path] = s
	}
	for name, prop := range s.Properties {
		collectFreeStrings(prefix, joinPath(path, name), prop, out)
	}
	if s.Items != nil {
		collectFreeStrings(prefix, path+"[]", s.Items, out)
	}
}

func isFreeString(s *openapi3.Schema) bool {
	return s.Type != nil && s.Type.Includes(openapi3.TypeString) &&
		len(s.Enum) == 0 && s.Format == "" && s.Pattern == "" && s.Const == nil
}

func joinPath(path, name string) string {
	if path == "" {
		return name
	}
	return strings.Join([]string{path, name}, ".")
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
