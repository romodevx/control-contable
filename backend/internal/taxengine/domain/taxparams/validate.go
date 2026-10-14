package taxparams

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

const (
	// minFiscalYear and maxFiscalYear bound a set's fiscal year: the app
	// starts in 2026, and anything before 2000 is a typo.
	minFiscalYear = 2000
	maxFiscalYear = 9999
	minVersion    = 1
	// minKeySegments: a key is at least "<tax>.<name>".
	minKeySegments = 2
	httpsScheme    = "https"
	// oneCentavo is where every tariff starts (0.01) and the step between a
	// row's upper bound and the next row's lower bound.
	oneCentavo money.Cents = 1
	// allMonths is the bitmask with bits 1-12 set: a complete monthly window.
	allMonths uint16 = 1<<13 - 2
)

// validator collects every problem in a Spec, so one failed boot lists them
// all instead of one per fix-and-retry.
type validator struct {
	problems []error
	sources  map[string]bool
}

func (v *validator) report(format string, args ...any) {
	v.problems = append(v.problems, fmt.Errorf(format, args...))
}

func (v *validator) header(s Spec) {
	if s.FiscalYear < minFiscalYear || s.FiscalYear > maxFiscalYear {
		v.report("fiscal_year %d outside %d-%d: %w", s.FiscalYear, minFiscalYear, maxFiscalYear, ErrInvalidHeader)
	}
	if s.Version < minVersion {
		v.report("version %d below %d: %w", s.Version, minVersion, ErrInvalidHeader)
	}
	if s.PublishedOn.IsZero() {
		v.report("published_on missing: %w", ErrInvalidHeader)
	}
}

func (v *validator) sourceList(sources []Source) {
	if len(sources) == 0 {
		v.report("sources: none: %w", ErrInvalidSource)
	}
	v.sources = make(map[string]bool, len(sources))
	for i, src := range sources {
		path := fmt.Sprintf("sources[%d] %s", i, src.ID)
		if blank(src.ID) || blank(src.Title) || blank(src.Reference) || src.DOFDate.IsZero() {
			v.report("%s: id, title, reference and dof_date are required: %w", path, ErrInvalidSource)
		}
		if !isHTTPSURL(src.URL) {
			v.report("%s: url %q is not an absolute https URL: %w", path, src.URL, ErrInvalidSource)
		}
		if v.sources[src.ID] {
			v.report("%s: repeated id: %w", path, ErrInvalidSource)
		}
		v.sources[src.ID] = true
	}
}

func isHTTPSURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == httpsScheme && u.Host != ""
}

func blank(s string) bool { return strings.TrimSpace(s) == "" }

func (v *validator) entry(path string, key Key, validity Validity, basis []Citation) {
	if !validKey(key) {
		v.report("%s: %w", path, ErrInvalidKey)
	}
	if validity.From.IsZero() || (!validity.IsOpenEnded() && validity.To.Compare(validity.From) < 0) {
		v.report("%s: validity %s: %w", path, validity, ErrInvalidValidity)
	}
	if len(basis) == 0 {
		v.report("%s: no basis: %w", path, ErrInvalidBasis)
	}
	for i, c := range basis {
		if blank(c.Article) || !v.sources[c.Source] {
			v.report("%s basis[%d]: needs an article and a known source, got %q: %w", path, i, c.Source, ErrInvalidBasis)
		}
	}
}

// validKey reports dotted segments of [a-z0-9_], at least minKeySegments.
func validKey(k Key) bool {
	segments := strings.Split(string(k), ".")
	if len(segments) < minKeySegments {
		return false
	}
	for _, s := range segments {
		if s == "" || strings.IndexFunc(s, func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_'
		}) >= 0 {
			return false
		}
	}
	return true
}

func (v *validator) rate(path string, r money.Rate) {
	if r < 0 || r > money.RateScale {
		v.report("%s: rate %s: %w", path, r, ErrRateOutOfRange)
	}
}

func (v *validator) amount(path string, c money.Cents) {
	if c < 0 {
		v.report("%s: amount %s: %w", path, c, ErrNegativeMoney)
	}
}

func (v *validator) value(i int, val Value) {
	path := fmt.Sprintf("values[%d] %s", i, val.Key)
	v.entry(path, val.Key, val.Validity, val.Basis)
	switch val.Kind {
	case KindRate:
		v.rate(path, val.Rate)
	case KindMoney:
		v.amount(path, val.Money)
	case KindFlag:
	default:
		v.report("%s: value kind %d: %w", path, val.Kind, ErrInvalidKind)
	}
}

func (v *validator) table(i int, t Table) {
	path := fmt.Sprintf("tables[%d] %s", i, t.Key)
	v.entry(path, t.Key, t.Validity, t.Basis)
	if t.Kind != TableMarginal && t.Kind != TableFlat {
		v.report("%s: table kind %d: %w", path, t.Kind, ErrInvalidKind)
	}
	if t.Month < 0 || t.Month > time.December {
		v.report("%s: month %d: %w", path, t.Month, ErrInvalidMonth)
	}
	v.rows(path, t)
}

// rows checks a tariff is total from 0.01: contiguous rows (upper + 0.01 ==
// next lower), each well-formed, and only the last may be unbounded.
func (v *validator) rows(path string, t Table) {
	if len(t.Rows) == 0 {
		v.report("%s: no rows: %w", path, ErrInvalidTable)
		return
	}
	if t.Rows[0].Lower != oneCentavo {
		v.report("%s rows[0]: lower %s, want 0.01: %w", path, t.Rows[0].Lower, ErrInvalidTable)
	}
	last := len(t.Rows) - 1
	for i, r := range t.Rows {
		rowPath := fmt.Sprintf("%s rows[%d]", path, i)
		v.rate(rowPath, r.Rate)
		v.amount(rowPath, r.FixedFee)
		if t.Kind == TableFlat && r.FixedFee != 0 {
			v.report("%s: fixed fee on a flat table: %w", rowPath, ErrInvalidTable)
		}
		if r.Unbounded {
			if i != last || r.Upper != 0 {
				v.report("%s: only the last row may be unbounded, and it has no upper: %w", rowPath, ErrInvalidTable)
			}
			continue
		}
		if r.Upper < r.Lower {
			v.report("%s: upper %s below lower %s: %w", rowPath, r.Upper, r.Lower, ErrInvalidTable)
		}
		if i < last && !adjacent(r.Upper, t.Rows[i+1].Lower) {
			v.report("%s rows[%d]: lower %s, want %s + 0.01: %w", path, i+1, t.Rows[i+1].Lower, r.Upper, ErrNotContiguous)
		}
	}
}

// adjacent reports next == upper + 0.01 without overflowing.
func adjacent(upper, next money.Cents) bool {
	return next > upper && next-upper == oneCentavo
}

// kinds checks that each key keeps one kind across its entries: one value
// kind, or one table kind and either monthly or plain tables.
func (v *validator) kinds(s Spec) {
	valueKinds := map[Key]ValueKind{}
	for i, val := range s.Values {
		if first, seen := valueKinds[val.Key]; seen && first != val.Kind {
			v.report("values[%d] %s: %s after %s: %w", i, val.Key, val.Kind, first, ErrKindConflict)
			continue
		}
		valueKinds[val.Key] = val.Kind
	}
	tableKinds := map[Key]TableKind{}
	monthly := map[Key]bool{}
	for i, t := range s.Tables {
		path := fmt.Sprintf("tables[%d] %s", i, t.Key)
		if _, isValue := valueKinds[t.Key]; isValue {
			v.report("%s: key already names a value: %w", path, ErrKindConflict)
		}
		if first, seen := tableKinds[t.Key]; seen && first != t.Kind {
			v.report("%s: %s after %s: %w", path, t.Kind, first, ErrKindConflict)
		}
		if first, seen := monthly[t.Key]; seen && first != (t.Month != 0) {
			v.report("%s: %w", path, ErrMonthlyMixed)
		}
		tableKinds[t.Key] = t.Kind
		monthly[t.Key] = t.Month != 0
	}
}

// overlaps reports, per key (and month for tables), two entries in force on
// one day. Sorting by start makes it one pass: O(n log n).
func (v *validator) overlaps(s Spec) {
	valueGroups := groupBy(s.Values, func(val Value) tableID { return tableID{key: val.Key} })
	for _, group := range valueGroups {
		checkOverlap(v, "values", group, func(val Value) Validity { return val.Validity })
	}
	tableGroups := groupBy(s.Tables, func(t Table) tableID { return tableID{key: t.Key, month: t.Month} })
	for _, group := range tableGroups {
		checkOverlap(v, "tables", group, func(t Table) Validity { return t.Validity })
	}
}

// indexed is an entry with its position in the spec, for error paths.
type indexed[T any] struct {
	pos   int
	entry T
	id    tableID
}

// groupBy groups entries by id, in order of first appearance, so error
// messages are deterministic.
func groupBy[T any](entries []T, id func(T) tableID) [][]indexed[T] {
	at := map[tableID]int{}
	var groups [][]indexed[T]
	for i, e := range entries {
		k := id(e)
		g, seen := at[k]
		if !seen {
			g = len(groups)
			at[k] = g
			groups = append(groups, nil)
		}
		groups[g] = append(groups[g], indexed[T]{pos: i, entry: e, id: k})
	}
	return groups
}

func checkOverlap[T any](v *validator, field string, group []indexed[T], validity func(T) Validity) {
	sorted := slices.SortedStableFunc(slices.Values(group), func(a, b indexed[T]) int {
		return validity(a.entry).From.Compare(validity(b.entry).From)
	})
	for i := 1; i < len(sorted); i++ {
		prev, next := validity(sorted[i-1].entry), validity(sorted[i].entry)
		if !prev.endsBefore(next.From) {
			v.report("%s[%d] %s: %s overlaps %s[%d] %s: %w",
				field, sorted[i].pos, sorted[i].id, next, field, sorted[i-1].pos, prev, ErrOverlap)
		}
	}
}

// monthlyWindows checks that every validity window of a monthly key holds
// exactly the twelve months (ADR-0002 §3: the art. 106 tables are published
// one per month and never derived).
func (v *validator) monthlyWindows(s Spec) {
	type window struct {
		key      Key
		validity Validity
	}
	masks := map[window]uint16{}
	var order []window
	for _, t := range s.Tables {
		if t.Month < time.January || t.Month > time.December {
			continue
		}
		w := window{key: t.Key, validity: t.Validity}
		if _, seen := masks[w]; !seen {
			order = append(order, w)
		}
		masks[w] |= 1 << t.Month
	}
	for _, w := range order {
		if masks[w] != allMonths {
			v.report("tables %s %s: months %s: %w", w.key, w.validity, monthList(masks[w]), ErrMonthlyIncomplete)
		}
	}
}

func monthList(mask uint16) string {
	var months []string
	for m := time.January; m <= time.December; m++ {
		if mask&(1<<m) != 0 {
			months = append(months, fmt.Sprint(int(m)))
		}
	}
	return "[" + strings.Join(months, " ") + "]"
}

// tableID identifies the entries that must not overlap: a key, plus the
// month for a monthly table (0 for values and plain tables).
type tableID struct {
	key   Key
	month time.Month
}

func (id tableID) String() string {
	if id.month == 0 {
		return string(id.key)
	}
	return fmt.Sprintf("%s month %d", id.key, id.month)
}
