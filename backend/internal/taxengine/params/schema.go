package params

import (
	"errors"
	"fmt"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// YAML tags the decoder accepts for decimals and dates.
const (
	tagString    = "!!str"
	tagTimestamp = "!!timestamp"
)

// Table kinds as written in the files.
var tableKinds = map[string]taxparams.TableKind{
	"marginal": taxparams.TableMarginal,
	"flat":     taxparams.TableFlat,
}

// fileDoc is one parameter file. Field names are the schema; KnownFields
// rejects anything else.
type fileDoc struct {
	FiscalYear  int         `yaml:"fiscal_year"`
	Version     int         `yaml:"version"`
	PublishedOn dateText    `yaml:"published_on"`
	Provisional bool        `yaml:"provisional"`
	Sources     []sourceDoc `yaml:"sources"`
	Values      []valueDoc  `yaml:"values"`
	Tables      []tableDoc  `yaml:"tables"`
}

type sourceDoc struct {
	ID        string   `yaml:"id"`
	Title     string   `yaml:"title"`
	URL       string   `yaml:"url"`
	DOFDate   dateText `yaml:"dof_date"`
	Reference string   `yaml:"reference"`
}

type citationDoc struct {
	Source  string `yaml:"source"`
	Article string `yaml:"article"`
}

type valueDoc struct {
	Key       string        `yaml:"key"`
	Rate      *decimalText  `yaml:"rate"`
	Money     *decimalText  `yaml:"money"`
	Flag      *bool         `yaml:"flag"`
	ValidFrom dateText      `yaml:"valid_from"`
	ValidTo   dateText      `yaml:"valid_to"`
	Basis     []citationDoc `yaml:"basis"`
}

type tableDoc struct {
	Key       string        `yaml:"key"`
	Kind      string        `yaml:"kind"`
	Month     int           `yaml:"month"`
	ValidFrom dateText      `yaml:"valid_from"`
	ValidTo   dateText      `yaml:"valid_to"`
	Basis     []citationDoc `yaml:"basis"`
	Rows      []rowDoc      `yaml:"rows"`
}

type rowDoc struct {
	Lower    decimalText  `yaml:"lower"`
	Upper    *decimalText `yaml:"upper"`     // absent on the last row: "en adelante"
	FixedFee *decimalText `yaml:"fixed_fee"` // absent = 0.00 (flat tables have none)
	Rate     decimalText  `yaml:"rate"`
}

// decimalText is a decimal kept as its exact text. Only a quoted string is
// accepted: an unquoted 0.08 is a YAML float, and money never touches one.
type decimalText string

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *decimalText) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode || n.ShortTag() != tagString {
		return fmt.Errorf("line %d: %s %q: %w", n.Line, n.ShortTag(), n.Value, ErrInexactNumber)
	}
	*d = decimalText(n.Value)
	return nil
}

// dateText is a calendar date, written plain (2026-01-31) or quoted.
type dateText struct{ taxparams.Date }

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *dateText) UnmarshalYAML(n *yaml.Node) error {
	tag := n.ShortTag()
	if n.Kind != yaml.ScalarNode || (tag != tagString && tag != tagTimestamp) {
		return fmt.Errorf("line %d: %s %q: %w", n.Line, tag, n.Value, taxparams.ErrInvalidDate)
	}
	parsed, err := taxparams.ParseDate(n.Value)
	if err != nil {
		return fmt.Errorf("line %d: %w", n.Line, err)
	}
	d.Date = parsed
	return nil
}

// converter turns a decoded file into a taxparams.Spec, collecting every
// decimal or kind problem so one run reports them all.
type converter struct {
	problems []error
}

func (c *converter) report(err error) { c.problems = append(c.problems, err) }

func (c *converter) cents(path string, text decimalText) money.Cents {
	v, err := money.ParseCents(string(text))
	if err != nil {
		c.report(fmt.Errorf("%s: %w", path, err))
	}
	return v
}

func (c *converter) rate(path string, text decimalText) money.Rate {
	v, err := money.ParseRate(string(text))
	if err != nil {
		c.report(fmt.Errorf("%s: %w", path, err))
	}
	return v
}

func (c *converter) spec(doc fileDoc) (taxparams.Spec, error) {
	spec := taxparams.Spec{
		FiscalYear:  doc.FiscalYear,
		Version:     doc.Version,
		PublishedOn: doc.PublishedOn.Date,
		Provisional: doc.Provisional,
	}
	for _, s := range doc.Sources {
		spec.Sources = append(spec.Sources, taxparams.Source{
			ID: s.ID, Title: s.Title, URL: s.URL, DOFDate: s.DOFDate.Date, Reference: s.Reference,
		})
	}
	for i, v := range doc.Values {
		spec.Values = append(spec.Values, c.value(fmt.Sprintf("values[%d] %s", i, v.Key), v))
	}
	for i, t := range doc.Tables {
		spec.Tables = append(spec.Tables, c.table(fmt.Sprintf("tables[%d] %s", i, t.Key), t))
	}
	return spec, errors.Join(c.problems...)
}

func (c *converter) value(path string, doc valueDoc) taxparams.Value {
	v := taxparams.Value{
		Key:      taxparams.Key(doc.Key),
		Validity: taxparams.Validity{From: doc.ValidFrom.Date, To: doc.ValidTo.Date},
		Basis:    citations(doc.Basis),
	}
	kinds := 0
	if doc.Rate != nil {
		kinds++
		v.Kind, v.Rate = taxparams.KindRate, c.rate(path, *doc.Rate)
	}
	if doc.Money != nil {
		kinds++
		v.Kind, v.Money = taxparams.KindMoney, c.cents(path, *doc.Money)
	}
	if doc.Flag != nil {
		kinds++
		v.Kind, v.Flag = taxparams.KindFlag, *doc.Flag
	}
	if kinds != 1 {
		c.report(fmt.Errorf("%s: %d kinds: %w", path, kinds, ErrValueKind))
	}
	return v
}

func (c *converter) table(path string, doc tableDoc) taxparams.Table {
	kind, ok := tableKinds[doc.Kind]
	if !ok {
		c.report(fmt.Errorf("%s: kind %q: %w", path, doc.Kind, ErrTableKind))
	}
	t := taxparams.Table{
		Key:      taxparams.Key(doc.Key),
		Kind:     kind,
		Month:    time.Month(doc.Month),
		Validity: taxparams.Validity{From: doc.ValidFrom.Date, To: doc.ValidTo.Date},
		Basis:    citations(doc.Basis),
	}
	for i, r := range doc.Rows {
		rowPath := fmt.Sprintf("%s rows[%d]", path, i)
		row := taxparams.Row{
			Lower:     c.cents(rowPath+" lower", r.Lower),
			Unbounded: r.Upper == nil,
			Rate:      c.rate(rowPath+" rate", r.Rate),
		}
		if r.Upper != nil {
			row.Upper = c.cents(rowPath+" upper", *r.Upper)
		}
		if r.FixedFee != nil {
			row.FixedFee = c.cents(rowPath+" fixed_fee", *r.FixedFee)
		}
		t.Rows = append(t.Rows, row)
	}
	return t
}

func citations(docs []citationDoc) []taxparams.Citation {
	out := make([]taxparams.Citation, 0, len(docs))
	for _, d := range docs {
		out = append(out, taxparams.Citation{Source: d.Source, Article: d.Article})
	}
	return out
}
