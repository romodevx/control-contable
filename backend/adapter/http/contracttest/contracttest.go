// Package contracttest is the contract-test harness (ADR-0001, P0-18): it
// serves a request through a real handler (normally the full router built by
// cmd/api) and validates the response's status, headers and body against
// docs/api/openapi.yaml with libopenapi-validator, an implementation
// independent of the kin-openapi spec the server is generated from.
//
// A module's contract test names the operation it covers:
//
//	c := contracttest.Load(t)
//	rec := c.ServeOp(t, contracttest.Op("createIncome"), handler, req)
//	// ... the test's own assertions on rec
//
// ServeOp counts toward contract coverage (TestContractCoverage scans the
// backend for contracttest.Op literals); Serve validates any response without
// claiming coverage, e.g. a disabled module's 404 envelope.
//
// Validation runs in strict mode: a property or header the contract does not
// declare is a failure, so the server cannot leak undocumented fields. The
// header check is the harness's own (see transportHeaders).
package contracttest

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/pb33f/libopenapi"
	validator "github.com/pb33f/libopenapi-validator"
	"github.com/pb33f/libopenapi-validator/config"
	"github.com/pb33f/libopenapi-validator/helpers"
	"github.com/pb33f/libopenapi-validator/paths"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// SpecPath is the contract, relative to the repository root.
const SpecPath = "docs/api/openapi.yaml"

// Operation names a contract operation by its operationId. Build it with Op
// and a string literal, so contract coverage can be tracked from the source.
type Operation struct{ id string }

// Op returns the operation with the given operationId. The argument must be a
// string literal (TestContractCoverage rejects anything else).
func Op(id string) Operation { return Operation{id: id} }

// Contract is the parsed docs/api/openapi.yaml with its validator. It is safe
// for concurrent use.
type Contract struct {
	mu        sync.Mutex // guards validator, whose caches are not documented as concurrency-safe
	validator validator.Validator
	model     *v3.Document
	options   *config.ValidationOptions
	ids       []string // sorted operationIds
}

var (
	loadOnce   sync.Once
	loaded     *Contract
	errLoading error
)

// Load returns the repository contract, parsed once per test binary. It fails
// the test now when the contract cannot be found, parsed or compiled.
func Load(t testing.TB) *Contract {
	t.Helper()
	loadOnce.Do(func() { loaded, errLoading = load() })
	if errLoading != nil {
		t.Fatalf("contracttest: %v", errLoading)
	}
	return loaded
}

// openMapPaths are the body paths of the contract's open maps, whose keys
// are data, not contract: strict mode reports every key of an
// additionalProperties schema as undeclared, so their keys are exempt. Base
// validation still checks each value against the map's schema.
//
//   - RegimeOptions (TaxpayerRegime.options): keys are the catalog's
//     regimes.option_defs keys.
var openMapPaths = []string{
	"$.body.options.*",
	"$.body.**.options.*",
}

func load() (*Contract, error) {
	path, err := findSpec()
	if err != nil {
		return nil, err
	}
	src, err := os.ReadFile(path) // #nosec G304 -- the path is the repository contract found above
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", SpecPath, err)
	}
	doc, err := libopenapi.NewDocument(src)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", SpecPath, err)
	}
	model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("build %s: %w", SpecPath, err)
	}
	opts := []config.Option{
		config.WithFormatAssertions(),
		config.WithStrictMode(), // headers: see transportHeaders
		config.WithStrictIgnorePaths(openMapPaths...),
	}
	v, errs := validator.NewValidator(doc, opts...)
	if len(errs) > 0 {
		return nil, fmt.Errorf("compile %s: %w", SpecPath, errors.Join(errs...))
	}
	return &Contract{
		validator: v,
		model:     &model.Model,
		options:   config.NewValidationOptions(opts...),
		ids:       operationIDs(&model.Model),
	}, nil
}

// findSpec walks up from the working directory (a package directory under
// `go test`) to the directory holding docs/api/openapi.yaml.
func findSpec() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("working directory: %w", err)
	}
	for {
		candidate := filepath.Join(dir, filepath.FromSlash(SpecPath))
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("%s not found above the working directory", SpecPath)
		}
		dir = parent
	}
}

func operationIDs(model *v3.Document) []string {
	var ids []string
	for _, item := range model.Paths.PathItems.FromOldest() {
		for _, op := range item.GetOperations().FromOldest() {
			ids = append(ids, op.OperationId)
		}
	}
	slices.Sort(ids)
	return ids
}

// OperationIDs returns every operationId in the contract, sorted.
func (c *Contract) OperationIDs() []string { return slices.Clone(c.ids) }

// Serve sends req through h, validates the response against the operation
// req routes to and returns the recorded response for further assertions.
// A violation fails the test (without stopping it) with a readable report.
func (c *Contract) Serve(t testing.TB, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	c.Validate(t, req, rec)
	return rec
}

// ServeOp is Serve for a contract test of op: it also fails when op is not in
// the contract or req does not route to op. Calls to ServeOp are what
// TestContractCoverage counts as covering an operation.
func (c *Contract) ServeOp(t testing.TB, op Operation, h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	if !slices.Contains(c.ids, op.id) {
		t.Errorf("contracttest: %q is not an operationId in %s", op.id, SpecPath)
	} else if r, ok := c.route(req); ok && r.operation.OperationId != op.id {
		t.Errorf("contracttest: %s %s is %s, not %s", req.Method, r.path, r.operation.OperationId, op.id)
	}
	return c.Serve(t, h, req)
}

// Validate checks an already recorded response to req against the contract.
func (c *Contract) Validate(t testing.TB, req *http.Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	r, ok := c.route(req)
	if !ok {
		t.Errorf("contracttest: %s %s matches no operation in %s", req.Method, req.URL.Path, SpecPath)
		return
	}
	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	c.mu.Lock()
	_, errs := c.validator.ValidateHttpResponse(req, resp)
	c.mu.Unlock()
	errs = append(withoutLibraryHeaderChecks(errs), undeclaredHeaders(r.operation, rec.Code, rec.Header())...)
	if len(errs) > 0 {
		t.Error(report(req, r, rec, errs))
	}
}

// route is the operation a request resolves to in the contract.
type route struct {
	path      string // the contract's path template
	operation *v3.Operation
}

func (c *Contract) route(req *http.Request) (route, bool) {
	item, errs, path := paths.FindPath(req, c.model, c.options)
	if item == nil || len(errs) > 0 {
		return route{}, false
	}
	op := helpers.ExtractOperation(req, item)
	if op == nil {
		return route{}, false
	}
	return route{path: path, operation: op}, true
}
