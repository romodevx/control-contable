package taxparams

import (
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Key names one parameter, as dotted lower-case segments
// ("iva.rate.border_8", "isr.art106.accumulated").
type Key string

// unknownKind is the String of a kind outside the defined ones.
const unknownKind = "unknown"

// ValueKind is what a Value holds. The zero value is invalid on purpose, so
// a forgotten kind fails validation instead of reading as a rate.
type ValueKind int

// The value kinds.
const (
	// KindRate is a fraction in [0, 1] (money.Rate, 10^-6 exact).
	KindRate ValueKind = iota + 1
	// KindMoney is a non-negative amount in centavos (money.Cents).
	KindMoney
	// KindFlag is a boolean switch (e.g. rounding.payment_whole_pesos).
	KindFlag
	// KindCount is a positive whole number the law states as one, e.g. the
	// "cinco veces el valor anual de la UMA" of LISR 151: never a rate or an
	// amount computed from it.
	KindCount
	// KindDate is a calendar day, e.g. the deadline of an annual return.
	KindDate
)

func (k ValueKind) String() string {
	switch k {
	case KindRate:
		return "rate"
	case KindMoney:
		return "money"
	case KindFlag:
		return "flag"
	case KindCount:
		return "count"
	case KindDate:
		return "date"
	}
	return unknownKind
}

// TableKind is how a table's rows are applied. The zero value is invalid.
type TableKind int

// The table kinds.
const (
	// TableMarginal is a tariff: fixed fee + rate over the excess of the
	// row's lower bound (LISR arts. 96 and 106).
	TableMarginal TableKind = iota + 1
	// TableFlat holds non-marginal brackets: the row's rate applies to the
	// whole amount (RESICO PF, LISR art. 113-E). Its rows have no fixed fee.
	TableFlat
)

func (k TableKind) String() string {
	switch k {
	case TableMarginal:
		return "marginal"
	case TableFlat:
		return "flat"
	}
	return unknownKind
}

// Source is one official publication a set transcribes values from.
type Source struct {
	ID        string // referenced by Citation.Source
	Title     string
	URL       string      // https only
	DOFDate   fiscal.Date // publication date in the Diario Oficial de la Federación
	Reference string      // article, rule or annex section, e.g. "Anexo 8, B.VI"
}

// Citation is the legal basis of one value or table.
type Citation struct {
	Source  string // a Source.ID of the same set
	Article string
}

// Value is one dated scalar parameter. Exactly the field its Kind names is
// meaningful.
type Value struct {
	Key      Key
	Kind     ValueKind
	Rate     money.Rate
	Money    money.Cents
	Flag     bool
	Count    int64
	Date     fiscal.Date
	Validity Validity
	Basis    []Citation
}

// Row is one tariff row: [Lower, Upper] in centavos, or [Lower, ∞) when
// Unbounded ("en adelante").
type Row struct {
	Lower     money.Cents
	Upper     money.Cents
	Unbounded bool
	FixedFee  money.Cents
	Rate      money.Rate
}

// Table is one dated tariff, stored verbatim as published. Month is 1-12 for
// a monthly table (the twelve art. 106 accumulated tables), 0 otherwise.
type Table struct {
	Key      Key
	Kind     TableKind
	Month    time.Month
	Validity Validity
	Rows     []Row
	Basis    []Citation
}

// Spec is the raw content of one parameter set, as decoded from its file.
// NewParams validates it into a Params.
type Spec struct {
	FiscalYear  int
	Version     int
	PublishedOn fiscal.Date
	// Provisional marks a set published before its official sources (a draft
	// of next year's values). Selecting it adds the provisional_parameters
	// disclaimer, as the previous-year fallback does.
	Provisional bool
	Sources     []Source
	Values      []Value
	Tables      []Table
}

func (v Value) clone() Value {
	v.Basis = slices.Clone(v.Basis)
	return v
}

func (t Table) clone() Table {
	t.Rows = slices.Clone(t.Rows)
	t.Basis = slices.Clone(t.Basis)
	return t
}
