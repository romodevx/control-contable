package infra

import (
	"context"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

var (
	_ usecase.ProfileAccounts = (*AccountStore)(nil)
	_ usecase.AccountRemover  = (*AccountStore)(nil)
)

// Rename implements usecase.ProfileAccounts. RLS admits only the unit of
// work's own row, so another user's id updates nothing: ErrAccountNotFound.
func (s *AccountStore) Rename(ctx context.Context, id domain.UserID, name domain.Name) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	res := db.Model(&userRow{}).Where("id = ?", id.String()).Update("name", name.String())
	if res.Error != nil {
		return fmt.Errorf("auth infra: rename user: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return usecase.ErrAccountNotFound
	}
	return nil
}

// PasswordHash implements usecase.AccountRemover.
func (s *AccountStore) PasswordHash(ctx context.Context, id domain.UserID) (string, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return "", err
	}
	var hashes []string
	if err := db.Model(&userRow{}).Where("id = ?", id.String()).Limit(1).Pluck("password_hash", &hashes).Error; err != nil {
		return "", fmt.Errorf("auth infra: read password hash: %w", err)
	}
	if len(hashes) == 0 {
		return "", usecase.ErrAccountNotFound
	}
	return hashes[0], nil
}

// Delete implements usecase.AccountRemover: one hard DELETE of the users
// row, a compare-and-delete on the verified hash. Every user-owned table
// references users ON DELETE CASCADE (refresh tokens, so every session;
// consents; usage counters; entitlements; and each module's tables), and
// PostgreSQL runs the cascades below RLS, so cc_app needs no DELETE grant on
// them. trial_registry has no foreign key to users: it survives.
func (s *AccountStore) Delete(ctx context.Context, id domain.UserID, passwordHash string) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	res := db.Where("id = ? AND password_hash = ?", id.String(), passwordHash).Delete(&userRow{})
	if res.Error != nil {
		return false, fmt.Errorf("auth infra: delete user: %w", res.Error)
	}
	return res.RowsAffected == 1, nil
}
