// Command covergate fails when statement coverage in a Go cover profile is below
// a threshold, after excluding packages that are not business logic (generated
// code, commands). CI runs it through `make cover-check`.
//
// Usage:
//
//	go run ./cmd/covergate -profile coverage.out -min 80 -exclude '/adapter/http/gen/|/cmd/'
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

const defaultMinPct = 80

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "covergate:", err)
		os.Exit(1)
	}
}

func run() error {
	profilePath := flag.String("profile", "", "path to the go test -coverprofile output")
	minPct := flag.Int64("min", defaultMinPct, "minimum statement coverage, in percent")
	excludeExpr := flag.String("exclude", "", "regexp of file paths left out of the measurement")
	flag.Parse()
	if *profilePath == "" {
		return fmt.Errorf("-profile is required")
	}
	exclude, err := regexp.Compile(*excludeExpr)
	if err != nil {
		return fmt.Errorf("-exclude: %w", err)
	}

	f, err := os.Open(*profilePath)
	if err != nil {
		return fmt.Errorf("open profile: %w", err)
	}
	defer f.Close()

	coverage, err := Measure(f, exclude)
	if err != nil {
		return fmt.Errorf("%s: %w", *profilePath, err)
	}
	if !coverage.Meets(*minPct) {
		return fmt.Errorf("business-logic coverage %s is below %d%%", coverage, *minPct)
	}
	fmt.Printf("business-logic coverage %s meets the %d%% gate\n", coverage, *minPct)
	return nil
}
