package taxparams

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Params is one validated, immutable parameter set: a fiscal year's version.
// Every lookup resolves by date (ADR-0002 §3), never by fiscal year alone:
// period-level values use the last day of the period, record-level checks the
// record's date. Lookups are O(log n) in the entries of one key.
type Params struct {
	fiscalYear  int
	version     int
	publishedOn Date
	provisional bool
	sources     []Source
	shapes      map[Key]shape
	values      map[Key][]Value     // per key, sorted by Validity.From
	tables      map[tableID][]Table // per key and month, sorted by Validity.From
}

// shape is what a key names: a value of one kind, or plain or monthly tables.
type shape struct {
	value   ValueKind // 0 for a table key
	monthly bool
}

func (s shape) String() string {
	switch {
	case s.value != 0:
		return s.value.String()
	case s.monthly:
		return "monthly table"
	}
	return "table"
}

// NewParams validates spec and returns its immutable Params, or every
// problem found, joined (each wraps one of the validation errors).
func NewParams(spec Spec) (*Params, error) {
	v := &validator{}
	v.header(spec)
	v.sourceList(spec.Sources)
	for i, val := range spec.Values {
		v.value(i, val)
	}
	for i, t := range spec.Tables {
		v.table(i, t)
	}
	v.kinds(spec)
	v.overlaps(spec)
	v.monthlyWindows(spec)
	if len(v.problems) > 0 {
		return nil, fmt.Errorf("taxparams: set %d/v%d: %w", spec.FiscalYear, spec.Version, errors.Join(v.problems...))
	}
	return build(spec), nil
}

// build indexes a valid spec, copying every slice so neither the caller's
// spec nor a lookup's result can change the set.
func build(spec Spec) *Params {
	p := &Params{
		fiscalYear:  spec.FiscalYear,
		version:     spec.Version,
		publishedOn: spec.PublishedOn,
		provisional: spec.Provisional,
		sources:     slices.Clone(spec.Sources),
		shapes:      map[Key]shape{},
		values:      map[Key][]Value{},
		tables:      map[tableID][]Table{},
	}
	for _, val := range spec.Values {
		p.shapes[val.Key] = shape{value: val.Kind}
		p.values[val.Key] = append(p.values[val.Key], val.clone())
	}
	for _, t := range spec.Tables {
		p.shapes[t.Key] = shape{monthly: t.Month != 0}
		id := tableID{key: t.Key, month: t.Month}
		p.tables[id] = append(p.tables[id], t.clone())
	}
	for _, entries := range p.values {
		slices.SortFunc(entries, func(a, b Value) int { return a.Validity.From.Compare(b.Validity.From) })
	}
	for _, entries := range p.tables {
		slices.SortFunc(entries, func(a, b Table) int { return a.Validity.From.Compare(b.Validity.From) })
	}
	return p
}

// FiscalYear returns the fiscal year the set was published for.
func (p *Params) FiscalYear() int { return p.fiscalYear }

// Version returns the set's version within its fiscal year.
func (p *Params) Version() int { return p.version }

// PublishedOn returns the day the set was written.
func (p *Params) PublishedOn() Date { return p.publishedOn }

// Provisional reports a set marked provisional in its file.
func (p *Params) Provisional() bool { return p.provisional }

// Sources returns a copy of the set's official sources.
func (p *Params) Sources() []Source { return slices.Clone(p.sources) }

// At returns the value of key in force on d: ErrUnknownKey, ErrKindMismatch
// for a table key, or ErrNotInForce when no entry covers d.
func (p *Params) At(key Key, d Date) (Value, error) {
	s, err := p.shapeOf(key)
	if err != nil {
		return Value{}, err
	}
	if s.value == 0 {
		return Value{}, kindMismatch(key, s, "value")
	}
	val, ok := inForce(p.values[key], func(v Value) Validity { return v.Validity }, d)
	if !ok {
		return Value{}, notInForce(key, d)
	}
	return val.clone(), nil
}

// Rate returns the rate key holds on d.
func (p *Params) Rate(key Key, d Date) (money.Rate, error) {
	val, err := p.typedValue(key, KindRate, d)
	return val.Rate, err
}

// Money returns the amount key holds on d.
func (p *Params) Money(key Key, d Date) (money.Cents, error) {
	val, err := p.typedValue(key, KindMoney, d)
	return val.Money, err
}

// Flag returns the switch key holds on d.
func (p *Params) Flag(key Key, d Date) (bool, error) {
	val, err := p.typedValue(key, KindFlag, d)
	return val.Flag, err
}

func (p *Params) typedValue(key Key, kind ValueKind, d Date) (Value, error) {
	s, err := p.shapeOf(key)
	if err != nil {
		return Value{}, err
	}
	if s.value != kind {
		return Value{}, kindMismatch(key, s, kind.String())
	}
	return p.At(key, d)
}

// Table returns the plain (not monthly) table key holds on d.
func (p *Params) Table(key Key, d Date) (Table, error) {
	return p.table(key, 0, d)
}

// MonthlyTable returns the month's table of a monthly key in force on d
// (for the art. 106 accumulated tariffs: the table of the period's month).
func (p *Params) MonthlyTable(key Key, month time.Month, d Date) (Table, error) {
	if month < time.January || month > time.December {
		return Table{}, fmt.Errorf("%s month %d: %w", key, month, ErrInvalidMonth)
	}
	return p.table(key, month, d)
}

func (p *Params) table(key Key, month time.Month, d Date) (Table, error) {
	want := shape{monthly: month != 0}
	s, err := p.shapeOf(key)
	if err != nil {
		return Table{}, err
	}
	if s != want {
		return Table{}, kindMismatch(key, s, want.String())
	}
	t, ok := inForce(p.tables[tableID{key: key, month: month}], func(t Table) Validity { return t.Validity }, d)
	if !ok {
		return Table{}, notInForce(key, d)
	}
	return t.clone(), nil
}

func (p *Params) shapeOf(key Key) (shape, error) {
	s, ok := p.shapes[key]
	if !ok {
		return shape{}, fmt.Errorf("%s: %w", key, ErrUnknownKey)
	}
	return s, nil
}

func kindMismatch(key Key, got shape, want string) error {
	return fmt.Errorf("%s is a %s, not a %s: %w", key, got, want, ErrKindMismatch)
}

func notInForce(key Key, d Date) error {
	return fmt.Errorf("%s on %s: %w", key, d, ErrNotInForce)
}

// inForce returns the entry whose validity contains d. Entries are sorted by
// start and never overlap, so the only candidate is the last one starting on
// or before d: a binary search, O(log n).
func inForce[T any](entries []T, validity func(T) Validity, d Date) (T, bool) {
	i := sort.Search(len(entries), func(i int) bool { return validity(entries[i]).From.Compare(d) > 0 })
	if i == 0 {
		var zero T
		return zero, false
	}
	e := entries[i-1]
	return e, validity(e).Contains(d)
}

// Requirement is one parameter a consumer (a regime strategy) needs, checked
// at startup with Params.Require so a missing key fails the boot, not an
// estimate.
type Requirement struct {
	key  Key
	want shape
}

// RequireRate requires key to be a rate.
func RequireRate(key Key) Requirement { return Requirement{key: key, want: shape{value: KindRate}} }

// RequireMoney requires key to be an amount.
func RequireMoney(key Key) Requirement { return Requirement{key: key, want: shape{value: KindMoney}} }

// RequireFlag requires key to be a flag.
func RequireFlag(key Key) Requirement { return Requirement{key: key, want: shape{value: KindFlag}} }

// RequireTable requires key to be a plain table.
func RequireTable(key Key) Requirement { return Requirement{key: key} }

// RequireMonthlyTables requires key to be a monthly table key; validation
// already guarantees its twelve months.
func RequireMonthlyTables(key Key) Requirement {
	return Requirement{key: key, want: shape{monthly: true}}
}

// Require checks every requirement and returns all failures joined
// (ErrUnknownKey or ErrKindMismatch, each naming its key).
func (p *Params) Require(reqs ...Requirement) error {
	var problems []error
	for _, r := range reqs {
		s, err := p.shapeOf(r.key)
		if err == nil && s != r.want {
			err = kindMismatch(r.key, s, r.want.String())
		}
		if err != nil {
			problems = append(problems, err)
		}
	}
	return errors.Join(problems...)
}
