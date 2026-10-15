package usecase

import (
	"cmp"
	"context"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// OptionDefView is one regimes.option_defs element as the contract shows it
// (RegimeOptionDef): the stored definition, label included.
type OptionDefView struct {
	Key   string
	Label string
	Type  domain.OptionType
	// AllowedValues is set for enum and multi_enum.
	AllowedValues []string
	// Min and Max are the stored decimal bounds of a rate; "" when absent.
	Min, Max string
	// Default is the zero OptionValue when the option has none (null).
	Default  domain.OptionValue
	Required bool
}

// TaxInputDefView is one regimes.required_inputs element (TaxInputDef).
type TaxInputDefView struct {
	Key      string
	Label    string
	Kind     string
	Required bool
}

// ListedRegime is one regimes catalog row for a person type, as listed.
type ListedRegime struct {
	Code       string
	PersonType domain.PersonType
	Name       string
	Treatment  string
	// IncomeBasis is regimes.income_basis ("cash" or "accrual"): which
	// date books an income of the regime (MODULE-incomes §2.2).
	IncomeBasis string
	// Explanation is "" when the regime has none (null).
	Explanation    string
	OptionDefs     []OptionDefView
	RequiredInputs []TaxInputDefView
}

// PairVerdict is one regime_compatibility row.
type PairVerdict struct {
	Pair RegimePair
	Compatibility
}

// RegimeDirectory lists the global regime catalog (M2-02) for
// listRegimes. Both reads run in the unit of work in ctx.
type RegimeDirectory interface {
	// ListRegimes returns the person type's regimes ordered by code.
	ListRegimes(ctx context.Context, personType domain.PersonType) ([]ListedRegime, error)
	// ListCompatibility returns every stored pair of the person type.
	ListCompatibility(ctx context.Context, personType domain.PersonType) ([]PairVerdict, error)
}

// RegimeLink names another regime of a compatibility pair and the pair's
// user-facing reason (contract RegimeIncompatibility).
type RegimeLink struct {
	Code   string
	Reason string
}

// CatalogEntry is a regime as listRegimes shows it.
type CatalogEntry struct {
	ListedRegime
	// IncompatibleWith are the regimes it cannot overlap in time.
	IncompatibleWith []RegimeLink
	// ReviewPendingWith are the regimes it may overlap with a warning: the
	// pair is allowed but not yet verified by a CPA
	// (regime_combination_needs_cpa_review), so the client can warn before
	// submitting (PR #56 follow-up).
	ReviewPendingWith []RegimeLink
}

// Catalog lists the regime catalog (listRegimes). Global, read-only data:
// a read-only user may list it, so there is no access check.
type Catalog struct {
	service
}

// NewCatalog returns the use case over deps (Tx and Directory are used).
func NewCatalog(deps Deps) *Catalog { return &Catalog{service: newService(deps)} }

// List returns the regimes selectable for personType, ordered by code, each
// with the regimes it pairs with: refused pairs in IncompatibleWith, pairs
// under CPA review in ReviewPendingWith, both ordered by code.
func (c *Catalog) List(ctx context.Context, userID, rawPersonType string) ([]CatalogEntry, error) {
	personType, err := domain.ParsePersonType(rawPersonType)
	if err != nil {
		fields := &fieldErrors{}
		fields.add(fieldPersonType, DetailInvalid, msgInvalidField, err)
		return nil, fields.err()
	}
	var (
		regimes []ListedRegime
		pairs   []PairVerdict
	)
	err = c.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		var err error
		if regimes, err = c.deps.Directory.ListRegimes(ctx, personType); err != nil {
			return err
		}
		pairs, err = c.deps.Directory.ListCompatibility(ctx, personType)
		return err
	})
	if err != nil {
		return nil, present("list regimes", err)
	}
	return linkPairs(regimes, pairs), nil
}

// linkPairs attaches every pair to both of its regimes. O(r + p log p).
func linkPairs(regimes []ListedRegime, pairs []PairVerdict) []CatalogEntry {
	incompatible := make(map[string][]RegimeLink, len(regimes))
	pending := make(map[string][]RegimeLink, len(regimes))
	for _, p := range pairs {
		var into map[string][]RegimeLink
		switch {
		case !p.Allowed:
			into = incompatible
		case p.NeedsCPAReview:
			into = pending
		default:
			continue
		}
		into[p.Pair.A] = append(into[p.Pair.A], RegimeLink{Code: p.Pair.B, Reason: p.Reason})
		into[p.Pair.B] = append(into[p.Pair.B], RegimeLink{Code: p.Pair.A, Reason: p.Reason})
	}
	out := make([]CatalogEntry, 0, len(regimes))
	for _, r := range regimes {
		out = append(out, CatalogEntry{
			ListedRegime:      r,
			IncompatibleWith:  sortedLinks(incompatible[r.Code]),
			ReviewPendingWith: sortedLinks(pending[r.Code]),
		})
	}
	return out
}

func sortedLinks(links []RegimeLink) []RegimeLink {
	if links == nil {
		return []RegimeLink{}
	}
	slices.SortFunc(links, func(a, b RegimeLink) int { return cmp.Compare(a.Code, b.Code) })
	return links
}
