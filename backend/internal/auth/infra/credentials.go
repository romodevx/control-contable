package infra

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// readCredentialsSQL calls the login lookup (migration
// 20261013150000_auth_service): at most one row, since email_hash is unique.
const readCredentialsSQL = "SELECT id::text AS id, password_hash, status FROM auth_lookup_credentials(?)" //nolint:gosec // G101: a query, not a credential

// statusActive is the users.status that may sign in. Any other value,
// including one added later, keeps the account out (fail closed).
const statusActive = "active"

// CredentialStore implements usecase.CredentialLookup over the SECURITY
// DEFINER function auth_lookup_credentials, the only way cc_app reads another
// user's credentials under FORCE RLS (D1). It runs in the unit of work in ctx.
type CredentialStore struct{}

// NewCredentialStore returns the store.
func NewCredentialStore() CredentialStore { return CredentialStore{} }

var _ usecase.CredentialLookup = CredentialStore{}

type credentialRow struct {
	ID           string
	PasswordHash string
	Status       string
}

// Lookup implements usecase.CredentialLookup.
func (CredentialStore) Lookup(ctx context.Context, emailHash []byte) (usecase.Credentials, bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.Credentials{}, false, err
	}
	var rows []credentialRow
	if err := db.Raw(readCredentialsSQL, bytea(emailHash)).Scan(&rows).Error; err != nil {
		return usecase.Credentials{}, false, fmt.Errorf("auth infra: look up credentials: %w", err)
	}
	if len(rows) == 0 {
		return usecase.Credentials{}, false, nil
	}
	id, err := domain.ParseUserID(rows[0].ID)
	if err != nil {
		return usecase.Credentials{}, false, fmt.Errorf("auth infra: credentials user id: %w", err)
	}
	return usecase.Credentials{
		UserID:       id,
		PasswordHash: rows[0].PasswordHash,
		Active:       rows[0].Status == statusActive,
	}, true, nil
}
