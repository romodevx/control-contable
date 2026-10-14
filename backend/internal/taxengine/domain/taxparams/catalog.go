package taxparams

import (
	"fmt"
	"slices"
)

// maxFallbackYears is how far back Select falls while a fiscal year is not
// published: one year (the December-January transcription window, ADR-0002
// Consequences). Older parameters are an error, never a silent estimate.
const maxFallbackYears = 1

// SetID names one set: its fiscal year and version ("2026/v1").
type SetID struct {
	FiscalYear int
	Version    int
}

func (id SetID) String() string { return fmt.Sprintf("%d/v%d", id.FiscalYear, id.Version) }

func (id SetID) compare(o SetID) int {
	if id.FiscalYear != o.FiscalYear {
		return id.FiscalYear - o.FiscalYear
	}
	return id.Version - o.Version
}

// Catalog holds every published parameter set and selects the one an
// estimate uses.
type Catalog struct {
	latest map[int]*Params // per fiscal year, its highest version
	ids    []SetID         // every set, ascending
}

// Selection is the set an estimate of RequestedYear uses. Provisional means
// the estimate must carry the provisional_parameters disclaimer: the set is
// marked provisional, or it is the previous year's (Fallback) because
// RequestedYear is not published yet.
type Selection struct {
	Params        *Params
	RequestedYear int
	Provisional   bool
	Fallback      bool
}

// NewCatalog indexes sets, or reports ErrNoParameterSets (none, or a nil
// one) or ErrDuplicateVersion.
func NewCatalog(sets ...*Params) (*Catalog, error) {
	if len(sets) == 0 {
		return nil, ErrNoParameterSets
	}
	c := &Catalog{latest: map[int]*Params{}}
	seen := map[SetID]bool{}
	for _, p := range sets {
		if p == nil {
			return nil, fmt.Errorf("nil set: %w", ErrNoParameterSets)
		}
		id := SetID{FiscalYear: p.fiscalYear, Version: p.version}
		if seen[id] {
			return nil, fmt.Errorf("%s: %w", id, ErrDuplicateVersion)
		}
		seen[id] = true
		c.ids = append(c.ids, id)
		if cur, ok := c.latest[id.FiscalYear]; !ok || cur.version < id.Version {
			c.latest[id.FiscalYear] = p
		}
	}
	slices.SortFunc(c.ids, SetID.compare)
	return c, nil
}

// Sets returns every set's id, ascending by fiscal year and version.
func (c *Catalog) Sets() []SetID { return slices.Clone(c.ids) }

// Select returns the highest version published for fiscalYear. While that
// year is not published, it falls back to the previous year's highest version,
// flagged Provisional and Fallback; values still resolve by date, so a
// decree that lapsed at year end stays lapsed. No set within reach is
// ErrNoParameters; a later year's set is never used for an earlier year.
func (c *Catalog) Select(fiscalYear int) (Selection, error) {
	if p, ok := c.latest[fiscalYear]; ok {
		return Selection{Params: p, RequestedYear: fiscalYear, Provisional: p.provisional}, nil
	}
	for back := 1; back <= maxFallbackYears; back++ {
		if p, ok := c.latest[fiscalYear-back]; ok {
			return Selection{Params: p, RequestedYear: fiscalYear, Provisional: true, Fallback: true}, nil
		}
	}
	return Selection{}, fmt.Errorf("fiscal year %d: %w", fiscalYear, ErrNoParameters)
}
