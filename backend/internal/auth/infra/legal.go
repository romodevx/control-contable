package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// currentLegalDocumentsSQL picks each kind's current version: the latest one
// already in effect (UNIQUE (kind, effective_at) keeps it unambiguous).
const currentLegalDocumentsSQL = `SELECT DISTINCT ON (kind) kind, version, url, effective_at FROM legal_documents
WHERE effective_at <= now() ORDER BY kind, effective_at DESC`

// LegalDocumentStore implements usecase.LegalDocuments and
// usecase.LegalCatalog.
type LegalDocumentStore struct{}

// NewLegalDocumentStore returns the store.
func NewLegalDocumentStore() LegalDocumentStore { return LegalDocumentStore{} }

// CurrentVersions implements usecase.LegalDocuments.
func (store LegalDocumentStore) CurrentVersions(ctx context.Context) (map[domain.DocumentKind]string, error) {
	docs, err := store.Current(ctx)
	if err != nil {
		return nil, err
	}
	current := make(map[domain.DocumentKind]string, len(docs))
	for _, d := range docs {
		current[d.Kind] = d.Version
	}
	return current, nil
}

type legalDocumentRow struct {
	Kind        string
	Version     string
	URL         string `gorm:"column:url"`
	EffectiveAt time.Time
}

var _ usecase.LegalCatalog = LegalDocumentStore{}

// Current implements usecase.LegalCatalog: the document in effect of each
// kind (the database clock decides), ordered by kind.
func (LegalDocumentStore) Current(ctx context.Context) ([]usecase.LegalDocument, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []legalDocumentRow
	if err := db.Raw(currentLegalDocumentsSQL).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("auth infra: current legal documents: %w", err)
	}
	docs := make([]usecase.LegalDocument, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, usecase.LegalDocument{
			Kind: domain.DocumentKind(r.Kind), Version: r.Version, URL: r.URL, PublishedAt: r.EffectiveAt,
		})
	}
	return docs, nil
}
