package router

import (
	"errors"
	"net/http"
	"regexp"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"

	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// pathParamPattern finds the {name} segments of an OpenAPI path template.
var pathParamPattern = regexp.MustCompile(`\{([^{}]+)\}`)

// validate checks the request against route's operation in the contract
// (parameters, content type and body; OAS 3.1 schemas use JSON Schema
// 2020-12) and answers 400 validation_failed when it does not conform.
// Security requirements are left to the Authenticator. The body is capped at
// maxBody bytes and stays readable for the handler.
func validate(route *routers.Route, maxBody int64, next http.Handler) http.Handler {
	names := pathParamNames(route.Path)
	declared := queryParamNames(route)
	options := &openapi3filter.Options{
		MultiError:          true,
		AuthenticationFunc:  openapi3filter.NoopAuthenticationFunc,
		SkipSettingDefaults: true, // handlers see what the client sent
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		}
		if unknown := undeclaredQuery(r, declared); len(unknown) > 0 {
			_ = httperr.Write(w, r, apperr.Mapper{}, invalid(errUndeclaredQuery, unknown))
			return
		}
		params := make(map[string]string, len(names))
		for _, name := range names {
			params[name] = r.PathValue(name)
		}
		err := openapi3filter.ValidateRequest(r.Context(), &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: params,
			Route:      route,
			Options:    options,
		})
		if err != nil {
			_ = httperr.Write(w, r, apperr.Mapper{}, invalid(err, fieldErrors(err)))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// errUndeclaredQuery is the cause of a 400 for a query parameter the
// operation does not declare.
var errUndeclaredQuery = errors.New("router: query parameter not declared by the operation")

// codeUndeclared names an undeclared query parameter, as JSON Schema names
// a property its object does not allow.
const codeUndeclared = "additionalProperties"

// queryParamNames are the query parameters route's operation (and its path
// item) declares.
func queryParamNames(route *routers.Route) map[string]bool {
	declared := map[string]bool{}
	var all openapi3.Parameters
	if route.PathItem != nil {
		all = append(all, route.PathItem.Parameters...)
	}
	if route.Operation != nil {
		all = append(all, route.Operation.Parameters...)
	}
	for _, p := range all {
		if p != nil && p.Value != nil && p.Value.In == openapi3.ParameterInQuery {
			declared[p.Value.Name] = true
		}
	}
	return declared
}

// undeclaredQuery names, sorted, every query parameter of r the operation
// does not declare: the contract's queries are closed (MODULE-dashboard §5,
// MODULE-tax-engine §5), so one is refused rather than silently ignored.
func undeclaredQuery(r *http.Request, declared map[string]bool) []apperr.FieldError {
	var names []string
	for name := range r.URL.Query() {
		if !declared[name] {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	out := make([]apperr.FieldError, 0, len(names))
	for _, name := range names {
		out = append(out, fieldError(name, codeUndeclared))
	}
	return out
}

func pathParamNames(path string) []string {
	var names []string
	for _, m := range pathParamPattern.FindAllStringSubmatch(path, -1) {
		names = append(names, m[1])
	}
	return names
}
