package golden

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// RawParams is a parameter file read as written, independently of the
// engine's loader (internal/taxengine/params): the boundary generator and
// the reference implementation must not share the engine's decoding.
type RawParams struct {
	FiscalYear int        `yaml:"fiscal_year"`
	Values     []RawValue `yaml:"values"`
	Tables     []RawTable `yaml:"tables"`
}

// RawValue is one value entry; Rate and Money keep their decimal text.
type RawValue struct {
	Key       string  `yaml:"key"`
	Rate      string  `yaml:"rate"`
	Money     string  `yaml:"money"`
	ValidFrom RawDate `yaml:"valid_from"`
	ValidTo   RawDate `yaml:"valid_to"`
}

// RawTable is one tariff entry; Month is 0 for a table that is not monthly.
type RawTable struct {
	Key       string   `yaml:"key"`
	Kind      string   `yaml:"kind"`
	Month     int      `yaml:"month"`
	ValidFrom RawDate  `yaml:"valid_from"`
	ValidTo   RawDate  `yaml:"valid_to"`
	Rows      []RawRow `yaml:"rows"`
}

// RawRow is one tariff row; Upper is "" on an open last row.
type RawRow struct {
	Lower    string `yaml:"lower"`
	Upper    string `yaml:"upper"`
	FixedFee string `yaml:"fixed_fee"`
	Rate     string `yaml:"rate"`
}

// RawDate is a date as written, plain (2026-01-31) or quoted; "" if absent.
type RawDate string

// UnmarshalYAML implements yaml.Unmarshaler: it keeps the scalar's text, so
// a plain date is not turned into a timestamp.
func (d *RawDate) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("line %d: date must be a scalar: %w", n.Line, ErrSyntax)
	}
	*d = RawDate(n.Value)
	return nil
}

// ReadRawParams reads a parameter file. It ignores the fields the boundary
// generator does not need (sources, basis); the engine's loader is the
// strict one.
func ReadRawParams(data []byte) (RawParams, error) {
	var rp RawParams
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(&rp); err != nil {
		return RawParams{}, fmt.Errorf("golden: raw params: %w: %w", ErrSyntax, err)
	}
	return rp, nil
}

// centavo is one centavo: the step from a row's upper bound to the next.
const centavo money.Cents = 1

// TariffPoint is one amount at a row edge of a tariff.
type TariffPoint struct {
	Table  RawTable
	Amount money.Cents
	Label  string // "row 6 lower", "row 6 upper", "row 11 upper+0.01"
}

// TariffEdges lists every row's lower bound, upper bound and upper + 0.01,
// once each, ascending (MODULE-tax-engine §3). In a contiguous table upper +
// 0.01 is the next row's lower bound and is listed as that; it survives as
// its own point only past a gap or past a bounded last row.
func TariffEdges(t RawTable) ([]TariffPoint, error) {
	type edge struct {
		raw   string
		label string
		step  money.Cents
	}
	var edges []edge
	for i, r := range t.Rows {
		edges = append(edges, edge{r.Lower, fmt.Sprintf("row %d lower", i+1), 0})
	}
	for i, r := range t.Rows {
		if r.Upper != "" {
			edges = append(edges, edge{r.Upper, fmt.Sprintf("row %d upper", i+1), 0})
		}
	}
	for i, r := range t.Rows {
		if r.Upper != "" {
			edges = append(edges, edge{r.Upper, fmt.Sprintf("row %d upper+0.01", i+1), centavo})
		}
	}
	seen := map[money.Cents]bool{}
	var points []TariffPoint
	for _, e := range edges {
		v, err := money.ParseCents(e.raw)
		if err != nil {
			return nil, fmt.Errorf("golden: %s %s: %w", t.Key, e.label, err)
		}
		v += e.step
		if seen[v] {
			continue
		}
		seen[v] = true
		points = append(points, TariffPoint{Table: t, Amount: v, Label: e.label})
	}
	slices.SortFunc(points, func(a, b TariffPoint) int { return cmp.Compare(a.Amount, b.Amount) })
	return points, nil
}

// ValidityPoint is a date at a validity edge of a value or table key.
type ValidityPoint struct {
	Key   string
	Month int  // a monthly table's month; 0 otherwise
	Table bool // a table key, not a value key
	Date  fiscal.Date
}

// ValidityEdges lists, for every entry, the day before its first day, its
// first day, its last day and the day after (an open-ended entry has no
// last day), once per key, values first, each ascending by key, month and
// date.
func ValidityEdges(rp RawParams) ([]ValidityPoint, error) {
	var points []ValidityPoint
	for _, v := range rp.Values {
		ps, err := edgesOf(v.Key, 0, false, v.ValidFrom, v.ValidTo)
		if err != nil {
			return nil, err
		}
		points = append(points, ps...)
	}
	for _, t := range rp.Tables {
		ps, err := edgesOf(t.Key, t.Month, true, t.ValidFrom, t.ValidTo)
		if err != nil {
			return nil, err
		}
		points = append(points, ps...)
	}
	slices.SortFunc(points, func(a, b ValidityPoint) int {
		return cmp.Or(compareBool(a.Table, b.Table), cmp.Compare(a.Key, b.Key), cmp.Compare(a.Month, b.Month), a.Date.Compare(b.Date))
	})
	return slices.Compact(points), nil
}

func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case b:
		return -1
	}
	return 1
}

const isoDate = "2006-01-02"

// dayShift is a date as written and the days to move it.
type dayShift struct {
	raw   RawDate
	shift int
}

func edgesOf(key string, month int, table bool, from, to RawDate) ([]ValidityPoint, error) {
	days := []dayShift{{from, -1}, {from, 0}}
	if to != "" {
		days = append(days, dayShift{to, 0}, dayShift{to, 1})
	}
	points := make([]ValidityPoint, 0, len(days))
	for _, d := range days {
		date, err := shiftDate(string(d.raw), d.shift)
		if err != nil {
			return nil, fmt.Errorf("golden: %s validity: %w", key, err)
		}
		points = append(points, ValidityPoint{Key: key, Month: month, Table: table, Date: date})
	}
	return points, nil
}

func shiftDate(raw string, days int) (fiscal.Date, error) {
	if _, err := fiscal.ParseDate(raw); err != nil {
		return fiscal.Date{}, fmt.Errorf("%q: %w", raw, err)
	}
	t, err := time.Parse(isoDate, raw)
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("%q: %w", raw, err)
	}
	return fiscal.ParseDate(t.AddDate(0, 0, days).Format(isoDate))
}
