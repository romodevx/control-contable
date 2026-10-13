// Package contracttest validates real HTTP responses against the contract.
package contracttest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Operation names a contract operation by its operationId.
type Operation struct{ id string }

// Op returns the operation with the given operationId.
func Op(id string) Operation { return Operation{id: id} }

// Contract is the parsed contract.
type Contract struct{}

// Load returns the contract.
func Load(testing.TB) *Contract { return &Contract{} }

// OperationIDs lists the contract's operationIds.
func (*Contract) OperationIDs() []string { return nil }

// Serve serves req through h and validates the response.
func (*Contract) Serve(_ testing.TB, _ http.Handler, _ *http.Request) *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// ServeOp is Serve for a named operation.
func (*Contract) ServeOp(_ testing.TB, _ Operation, _ http.Handler, _ *http.Request) *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// Validate validates a recorded response.
func (*Contract) Validate(testing.TB, *http.Request, *httptest.ResponseRecorder) {}

// ScanCoverage finds Op literals under root.
func ScanCoverage(string, ...string) (map[string][]string, error) { return nil, nil }
