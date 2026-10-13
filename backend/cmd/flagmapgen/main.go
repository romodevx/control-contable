// Command flagmapgen generates the operationId -> x-module map that drives the
// feature-flag gate (ADR-0001). It fails when any operation in the contract lacks
// an operationId or a valid x-module, so an ungated operation cannot ship.
//
// Usage (from a go:generate directive):
//
//	go run <path>/cmd/flagmapgen -spec openapi.yaml -out flagmap.gen.go -pkg gen
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const generatedFileMode = 0o644

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "flagmapgen:", err)
		os.Exit(1)
	}
}

func run() error {
	specPath := flag.String("spec", "", "path to the OpenAPI document")
	outPath := flag.String("out", "", "path of the generated Go file")
	pkg := flag.String("pkg", "", "package name of the generated Go file")
	flag.Parse()
	if *specPath == "" || *outPath == "" || *pkg == "" {
		return fmt.Errorf("-spec, -out and -pkg are required")
	}

	spec, err := os.ReadFile(*specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}
	ops, err := ParseOperations(spec)
	if err != nil {
		return fmt.Errorf("%s: %w", *specPath, err)
	}
	src, err := Render(*pkg, filepath.Base(*specPath), ops)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*outPath, src, generatedFileMode); err != nil {
		return fmt.Errorf("write %s: %w", *outPath, err)
	}
	return nil
}
