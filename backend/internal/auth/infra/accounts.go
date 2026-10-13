package infra

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm/clause"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/auth/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

const (
	// sqlstateUniqueViolation is PostgreSQL's unique_violation.
	sqlstateUniqueViolation = "23505"
	// usersEmailHashKey is the one-account-per-email constraint.
	usersEmailHashKey = "users_email_hash_key"
)

// emailAAD binds an email ciphertext to its user: a ciphertext copied onto
// another user's row fails authentication instead of decrypting.
func emailAAD(userID domain.UserID) []byte { return []byte("users.email_encrypted:" + userID.String()) }

// userRow is the users columns sign-up writes. CreatedAt is read back only
// (the database stamps it; cc_app has no INSERT grant on it).
type userRow struct {
	ID              string `gorm:"column:id;primaryKey"`
	EmailHash       []byte
	EmailEncrypted  []byte
	EmailKeyID      int16
	PasswordHash    string
	Name            *string
	EmailVerifiedAt time.Time
	CreatedAt       time.Time `gorm:"->"`
}

func (userRow) TableName() string { return "users" }

// usageRow is user_usage at sign-up: the counters start at their defaults.
type usageRow struct {
	UserID string `gorm:"column:user_id;primaryKey"`
}

func (usageRow) TableName() string { return "user_usage" }

// consentRow is one accepted legal document version.
type consentRow struct {
	UserID          string
	DocumentKind    string
	DocumentVersion string
	AcceptedAt      time.Time
}

func (consentRow) TableName() string { return "consents" }

// AccountStore implements usecase.Accounts: the users row (email sealed with
// the keyring), its trial-cap counters and its consents, in the unit of work
// of the new user (RLS admits only that user's rows).
type AccountStore struct {
	emails *crypto.Keyring
}

// NewAccountStore returns a store sealing emails with keyring.
func NewAccountStore(keyring *crypto.Keyring) *AccountStore { return &AccountStore{emails: keyring} }

var _ usecase.Accounts = (*AccountStore)(nil)

// Create implements usecase.Accounts.
func (s *AccountStore) Create(ctx context.Context, a usecase.NewAccount) (usecase.Account, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.Account{}, err
	}
	sealed, err := s.emails.Encrypt([]byte(a.Email.String()), emailAAD(a.ID))
	if err != nil {
		return usecase.Account{}, fmt.Errorf("auth infra: seal email: %w", err)
	}
	user := userRow{
		ID:              a.ID.String(),
		EmailHash:       a.EmailHash,
		EmailEncrypted:  sealed,
		EmailKeyID:      int16(s.emails.ActiveKeyID()), // #nosec G115 -- crypto.MaxKeyID fits int16
		PasswordHash:    a.PasswordHash,
		Name:            optionalName(a.Name),
		EmailVerifiedAt: a.VerifiedAt,
	}
	err = db.Clauses(clause.Returning{Columns: []clause.Column{{Name: "created_at"}}}).Create(&user).Error
	if isUniqueViolation(err, usersEmailHashKey) {
		return usecase.Account{}, usecase.ErrAccountExists
	}
	if err != nil {
		return usecase.Account{}, fmt.Errorf("auth infra: insert user: %w", err)
	}
	if err := db.Create(&usageRow{UserID: user.ID}).Error; err != nil {
		return usecase.Account{}, fmt.Errorf("auth infra: insert usage counters: %w", err)
	}
	if err := insertConsents(ctx, user.ID, a.Consents, a.AcceptedAt); err != nil {
		return usecase.Account{}, err
	}
	return usecase.Account{ID: a.ID, Email: a.Email, Name: a.Name, CreatedAt: user.CreatedAt}, nil
}

var _ usecase.LoginAccounts = (*AccountStore)(nil)

// Find implements usecase.LoginAccounts: the account as the API shows it,
// its email opened with the keyring. RLS hides every row but the unit of
// work's user, so another user's id is ErrAccountNotFound.
func (s *AccountStore) Find(ctx context.Context, id domain.UserID) (usecase.Account, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.Account{}, err
	}
	var rows []userRow
	if err := db.Where("id = ?", id.String()).Limit(1).Find(&rows).Error; err != nil {
		return usecase.Account{}, fmt.Errorf("auth infra: read user: %w", err)
	}
	if len(rows) == 0 {
		return usecase.Account{}, usecase.ErrAccountNotFound
	}
	return s.toAccount(id, rows[0])
}

func (s *AccountStore) toAccount(id domain.UserID, row userRow) (usecase.Account, error) {
	plain, err := s.emails.Decrypt(row.EmailEncrypted, emailAAD(id))
	if err != nil {
		return usecase.Account{}, fmt.Errorf("auth infra: open email: %w", err)
	}
	email, err := domain.NewEmail(string(plain))
	if err != nil {
		return usecase.Account{}, fmt.Errorf("auth infra: stored email: %w", err)
	}
	var name domain.Name
	if row.Name != nil {
		if name, err = domain.NewName(*row.Name); err != nil {
			return usecase.Account{}, fmt.Errorf("auth infra: stored name: %w", err)
		}
	}
	return usecase.Account{ID: id, Email: email, Name: name, CreatedAt: row.CreatedAt}, nil
}

// ReplacePasswordHash implements usecase.LoginAccounts: a compare-and-set on
// the stored hash, so a password changed since the login read it stays.
func (s *AccountStore) ReplacePasswordHash(ctx context.Context, id domain.UserID, oldHash, newHash string) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	err = db.Model(&userRow{}).
		Where("id = ? AND password_hash = ?", id.String(), oldHash).
		Update("password_hash", newHash).Error
	if err != nil {
		return fmt.Errorf("auth infra: replace password hash: %w", err)
	}
	return nil
}

func insertConsents(ctx context.Context, userID string, documents []domain.DocumentVersion, acceptedAt time.Time) error {
	if len(documents) == 0 {
		return nil
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	rows := make([]consentRow, 0, len(documents))
	for _, d := range documents {
		rows = append(rows, consentRow{UserID: userID, DocumentKind: string(d.Kind), DocumentVersion: d.Version, AcceptedAt: acceptedAt})
	}
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("auth infra: insert consents: %w", err)
	}
	return nil
}

func optionalName(n domain.Name) *string {
	if n.IsZero() {
		return nil
	}
	s := n.String()
	return &s
}

func isUniqueViolation(err error, constraint string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == sqlstateUniqueViolation && pgErr.ConstraintName == constraint
}
