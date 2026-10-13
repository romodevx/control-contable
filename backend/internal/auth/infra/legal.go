package infra

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// currentLegalSQL picks each kind's current version: the latest one already
// in effect (UNIQUE (kind, effective_at) keeps it unambiguous).
const currentLegalSQL = `SELECT DISTINCT ON (kind) kind, version FROM legal_documents
WHERE effective_at <= now() ORDER BY kind, effective_at DESC`

// LegalDocumentStore implements usecase.LegalDocuments.
type LegalDocumentStore struct{}

// NewLegalDocumentStore returns the store.
func NewLegalDocumentStore() LegalDocumentStore { return LegalDocumentStore{} }

type legalRow struct {
	Kind    string
	Version string
}

// CurrentVersions implements usecase.LegalDocuments.
func (LegalDocumentStore) CurrentVersions(ctx context.Context) (map[domain.DocumentKind]string, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []legalRow
	if err := db.Raw(currentLegalSQL).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("auth infra: current legal documents: %w", err)
	}
	current := make(map[domain.DocumentKind]string, len(rows))
	for _, r := range rows {
		current[domain.DocumentKind(r.Kind)] = r.Version
	}
	return current, nil
}
