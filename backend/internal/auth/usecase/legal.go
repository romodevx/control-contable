package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// ErrNoLegalDocuments reports a catalog with no document in effect: nobody
// could accept the current versions, so sign-up is broken (a server fault).
var ErrNoLegalDocuments = errors.New("auth: no legal document is in effect")

// LegalDocument is the current version of one legal document kind
// (contract LegalDocument). PublishedAt is when it took effect.
type LegalDocument struct {
	Kind        domain.DocumentKind
	Version     string
	URL         string
	PublishedAt time.Time
}

// LegalCatalog reads the legal documents in effect, one per kind.
type LegalCatalog interface {
	Current(ctx context.Context) ([]LegalDocument, error)
}

// ListLegalDocuments lists the current legal documents (US-A5): public, part
// of the sign-up path, so it reads the global catalog with no user context.
type ListLegalDocuments struct {
	system  tx.SystemManager
	catalog LegalCatalog
}

// NewListLegalDocuments returns the use case.
func NewListLegalDocuments(system tx.SystemManager, catalog LegalCatalog) *ListLegalDocuments {
	return &ListLegalDocuments{system: system, catalog: catalog}
}

// Execute returns the current version of each kind.
func (uc *ListLegalDocuments) Execute(ctx context.Context) ([]LegalDocument, error) {
	var docs []LegalDocument
	err := uc.system.System(ctx, func(ctx context.Context) error {
		var err error
		docs, err = uc.catalog.Current(ctx)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("auth: list legal documents: %w", err)
	}
	if len(docs) == 0 {
		return nil, ErrNoLegalDocuments
	}
	return docs, nil
}
