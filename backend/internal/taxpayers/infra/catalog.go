package infra

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// Catalog implements usecase.RegimeCatalog over the global regimes and
// regime_compatibility tables (M2-02), which cc_app may only read. Both hold
// a few dozen rows; their primary keys serve every lookup.
type Catalog struct{}

var _ usecase.RegimeCatalog = Catalog{}

// NewCatalog returns a Catalog. It holds no pool.
func NewCatalog() Catalog { return Catalog{} }

type catalogRow struct {
	Code       string
	PersonType string
	Name       string
	Treatment  string
	OptionDefs string
}

// Regime implements usecase.RegimeCatalog.
func (Catalog) Regime(ctx context.Context, personType domain.PersonType, code string) (usecase.CatalogRegime, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.CatalogRegime{}, err
	}
	var rows []catalogRow
	err = db.Raw(`SELECT code, person_type, name, treatment, option_defs::text AS option_defs
FROM regimes WHERE code = ? AND person_type = ?`, code, string(personType)).Scan(&rows).Error
	if err != nil {
		return usecase.CatalogRegime{}, translate("read regime catalog", err)
	}
	if len(rows) == 0 {
		return usecase.CatalogRegime{}, usecase.ErrRegimeNotForPersonType
	}
	defs, err := decodeOptionDefs([]byte(rows[0].OptionDefs))
	if err != nil {
		return usecase.CatalogRegime{}, err
	}
	options, err := domain.NewOptionSet(defs...)
	if err != nil {
		return usecase.CatalogRegime{}, fmt.Errorf("taxpayers infra: regime %s option_defs: %w", code, err)
	}
	return usecase.CatalogRegime{
		Code: rows[0].Code, PersonType: personType, Name: rows[0].Name, Treatment: rows[0].Treatment, Options: options,
	}, nil
}

type compatibilityRow struct {
	RegimeA        string
	RegimeB        string
	Allowed        bool
	NeedsCPAReview bool `gorm:"column:needs_cpa_review"`
	Reason         string
}

// Compatibility implements usecase.RegimeCatalog: every stored pair among
// codes, in one query.
func (Catalog) Compatibility(ctx context.Context, personType domain.PersonType, codes []string) (map[usecase.RegimePair]usecase.Compatibility, error) {
	out := make(map[usecase.RegimePair]usecase.Compatibility)
	if len(codes) < 2 {
		return out, nil
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []compatibilityRow
	err = db.Raw(`SELECT regime_a, regime_b, allowed, needs_cpa_review, reason
FROM regime_compatibility
WHERE person_type = ? AND regime_a IN ? AND regime_b IN ?`, string(personType), codes, codes).Scan(&rows).Error
	if err != nil {
		return nil, translate("read regime compatibility", err)
	}
	for _, r := range rows {
		out[usecase.PairOf(r.RegimeA, r.RegimeB)] = usecase.Compatibility{
			Allowed: r.Allowed, NeedsCPAReview: r.NeedsCPAReview, Reason: r.Reason,
		}
	}
	return out, nil
}
