package domain

import (
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
)

// Citation is one legal basis of a line (contract Citation): a law, its
// article and an optional note ("párrafo quinto", "fracciones I a V").
type Citation struct {
	Law     string
	Article string
	Note    string
}

// Valid reports a citation with a law and an article.
func (c Citation) Valid() bool { return !blank(c.Law) && !blank(c.Article) }

func blank(s string) bool { return strings.TrimSpace(s) == "" }

// sourceKind is where a line's figure comes from.
type sourceKind uint8

const (
	sourceRecords sourceKind = iota + 1
	sourceComputed
	sourceTaxInput
	sourceParams
)

// Source is where a line's figure comes from (contract BreakdownLine.source):
// the records, a tax input, the parameter set, or the computation itself.
// The zero value is no source, which a ledger step refuses.
type Source struct {
	kind sourceKind
	key  string
}

// SourceRecords is a sum of the period's or year's incomes or expenses.
func SourceRecords() Source { return Source{kind: sourceRecords} }

// SourceComputed is a figure computed from earlier lines.
func SourceComputed() Source { return Source{kind: sourceComputed} }

// SourceTaxInput is a value the user entered for key.
func SourceTaxInput(key InputKey) Source { return Source{kind: sourceTaxInput, key: string(key)} }

// SourceParams is a value or tariff of the parameter set.
func SourceParams(key taxparams.Key) Source { return Source{kind: sourceParams, key: string(key)} }

// String renders the contract form: "records", "computed",
// "tax_input:<key>" or "params:<key>"; "" for the zero Source.
func (s Source) String() string {
	switch s.kind {
	case sourceRecords:
		return "records"
	case sourceComputed:
		return "computed"
	case sourceTaxInput:
		return "tax_input:" + s.key
	case sourceParams:
		return "params:" + s.key
	}
	return ""
}

// LineCode is a stable, machine identifier of a strategy's step
// ("isr.base"), which goldens and tests assert on.
type LineCode string

// LabelKey is the key of a line's Spanish concept, rendered by the adapter
// with the line's Args ("Ingresos cobrados acumulados {from}–{to}").
type LabelKey string

// LabelRoundingAdjustment is the concept of the "Ajuste por redondeo" line.
const LabelRoundingAdjustment LabelKey = "rounding_adjustment"

// Step describes one computation step before the Ledger numbers it: what it
// is, where its figure comes from and its legal basis.
type Step struct {
	Tax            TaxKind
	Code           LineCode
	Label          LabelKey
	Args           []string
	Source         Source
	Basis          []Citation
	NeedsCPAReview bool // the step applies a rule a CPA has not validated (§2.9)
}

// Line is one breakdown line: the trace of one step. Amount is the value
// shown, rounded to the centavo; Exact is the value computed with. Rate is
// set on multiply lines and Row (1-based) on lookup lines.
type Line struct {
	Seq            int
	Tax            TaxKind
	Code           LineCode
	Label          LabelKey
	Args           []string
	Operation      Operation
	Amount         money.Cents
	Exact          money.Exact
	Rate           money.Rate
	Row            int
	Basis          []Citation
	Source         Source
	NeedsCPAReview bool
}

func (l Line) clone() Line {
	l.Args = slices.Clone(l.Args)
	l.Basis = slices.Clone(l.Basis)
	return l
}
