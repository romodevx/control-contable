package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	modeHeaderPrefix = "mode: "
	// A profile line is "<file>:<startLine>.<startCol>,<endLine>.<endCol> <numStmts> <count>".
	profileFields = 3
	percent       = 100
	// Percentages are reported to one decimal place, truncated (never rounded up).
	tenthsOfPercent = 1000
)

// Coverage is the statement coverage of the packages a profile reports.
type Coverage struct {
	Covered int64
	Total   int64
}

// Meets reports whether coverage is at least minPct percent, using integer
// arithmetic so 79.95% never passes an 80% gate.
func (c Coverage) Meets(minPct int64) bool {
	return c.Covered*percent >= minPct*c.Total
}

func (c Coverage) String() string {
	tenths := c.Covered * tenthsOfPercent / c.Total
	return fmt.Sprintf("%d.%d%% (%d/%d statements)", tenths/10, tenths%10, c.Covered, c.Total)
}

type block struct {
	statements int64
	covered    bool
}

// Measure reads a `go test -coverprofile` profile and returns the statement
// coverage of every block whose file path does not match exclude. A block that
// appears more than once (several test binaries) counts once, as covered if
// any run executed it.
func Measure(r io.Reader, exclude *regexp.Regexp) (Coverage, error) {
	blocks, err := parseBlocks(r, exclude)
	if err != nil {
		return Coverage{}, err
	}
	if len(blocks) == 0 {
		return Coverage{}, errors.New("profile has no statements left after exclusions")
	}
	var c Coverage
	for _, b := range blocks {
		c.Total += b.statements
		if b.covered {
			c.Covered += b.statements
		}
	}
	return c, nil
}

func parseBlocks(r io.Reader, exclude *regexp.Regexp) (map[string]block, error) {
	sc := bufio.NewScanner(r)
	if !sc.Scan() || !strings.HasPrefix(sc.Text(), modeHeaderPrefix) {
		return nil, errors.New("profile does not start with a mode line")
	}
	blocks := make(map[string]block)
	for lineNo := 2; sc.Scan(); lineNo++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		key, b, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		if exclude.MatchString(key) {
			continue
		}
		prev := blocks[key]
		blocks[key] = block{statements: b.statements, covered: prev.covered || b.covered}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read profile: %w", err)
	}
	return blocks, nil
}

func parseLine(line string) (string, block, error) {
	fields := strings.Fields(line)
	if len(fields) != profileFields {
		return "", block{}, fmt.Errorf("want %d fields, got %d", profileFields, len(fields))
	}
	statements, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return "", block{}, fmt.Errorf("statements: %w", err)
	}
	count, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return "", block{}, fmt.Errorf("count: %w", err)
	}
	return fields[0], block{statements: statements, covered: count > 0}, nil
}
