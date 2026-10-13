// Package gen holds the code generated from docs/api/openapi.yaml: the
// strict std-http server interface, models and embedded spec (api.gen.go,
// oapi-codegen) and the operationId -> x-module map that drives the feature-flag
// gate (flagmap.gen.go, cmd/flagmapgen). Never edit the *.gen.go files; run
// `make generate`. `make codegen-check` fails when they drift from the contract.
package gen

//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../../../docs/api/openapi.yaml
//go:generate go run ../../../cmd/flagmapgen -spec ../../../../docs/api/openapi.yaml -out flagmap.gen.go -pkg gen
