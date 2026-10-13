package router

import (
	"net/http"
	"regexp"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"

	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/httperr"
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
	options := &openapi3filter.Options{
		MultiError:          true,
		AuthenticationFunc:  openapi3filter.NoopAuthenticationFunc,
		SkipSettingDefaults: true, // handlers see what the client sent
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBody)
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

func pathParamNames(path string) []string {
	var names []string
	for _, m := range pathParamPattern.FindAllStringSubmatch(path, -1) {
		names = append(names, m[1])
	}
	return names
}
