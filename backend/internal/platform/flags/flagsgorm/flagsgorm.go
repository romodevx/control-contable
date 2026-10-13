// Package flagsgorm is the PostgreSQL adapter of the flag service (GORM,
// ADR-0001). Reader serves the API gate and GET /v1/flags and runs as the
// runtime role cc_app, which has SELECT on feature_flags only. Store backs the
// owner's admin CLI (cmd/flags) and runs as cc_owner: feature_flags,
// flag_presets and flag_preset_flags accept writes from no other role.
//
// The flag tables are global configuration with no user_id, so no RLS session
// variable is set (baseline migration, P0-10).
package flagsgorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/romodevx/control-contable/backend/internal/platform/flags"
)

// DB is a GORM handle over an existing *sql.DB pool.
type DB struct {
	gorm *gorm.DB
}

// Open wraps db (a pgx database/sql pool, the same one goose uses) in GORM. It
// does not ping and does not take ownership: the caller closes db.
func Open(db *sql.DB) (*DB, error) {
	g, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{
		Logger:                 logger.Discard, // SQL is never logged: it can carry PII elsewhere
		SkipDefaultTransaction: true,           // multi-writes open their own transaction
		DisableAutomaticPing:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("flagsgorm: open: %w", err)
	}
	return &DB{gorm: g}, nil
}

// flagRow maps feature_flags. created_at/updated_at are set by the database.
type flagRow struct {
	Key         string `gorm:"primaryKey"`
	Enabled     bool
	Description string
}

func (flagRow) TableName() string { return "feature_flags" }

func (r flagRow) toDomain() flags.Flag {
	return flags.Flag{Key: r.Key, Enabled: r.Enabled, Description: r.Description}
}

type presetRow struct {
	Name        string `gorm:"primaryKey"`
	Description string
	Members     []presetFlagRow `gorm:"foreignKey:PresetName;references:Name"`
}

func (presetRow) TableName() string { return "flag_presets" }

type presetFlagRow struct {
	PresetName string `gorm:"primaryKey"`
	FlagKey    string `gorm:"primaryKey"`
}

func (presetFlagRow) TableName() string { return "flag_preset_flags" }

// Reader implements flags.Reader.
type Reader struct {
	db *gorm.DB
}

// NewReader returns a Reader; db should connect as cc_app.
func NewReader(db *DB) Reader { return Reader{db: db.gorm} }

// EnabledKeys returns the keys of the flags that are ON, ordered by key.
func (r Reader) EnabledKeys(ctx context.Context) ([]string, error) {
	keys := []string{}
	err := r.db.WithContext(ctx).Model(&flagRow{}).
		Where("enabled").Order("key").Pluck("key", &keys).Error
	if err != nil {
		return nil, fmt.Errorf("flagsgorm: read enabled flags: %w", err)
	}
	return keys, nil
}

// Store implements flags.AdminStore.
type Store struct {
	db *gorm.DB
}

// NewStore returns a Store; db must connect as cc_owner.
func NewStore(db *DB) Store { return Store{db: db.gorm} }

// List returns every flag, ordered by key.
func (s Store) List(ctx context.Context) ([]flags.Flag, error) {
	return list(s.db.WithContext(ctx))
}

func list(db *gorm.DB) ([]flags.Flag, error) {
	var rows []flagRow
	if err := db.Order("key").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("flagsgorm: list flags: %w", err)
	}
	out := make([]flags.Flag, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toDomain())
	}
	return out, nil
}

// SetEnabled turns key ON or OFF and returns its previous state. The row is
// locked first, so the previous state reported is the one replaced, and an
// unchanged flag is not written (its updated_at keeps the last real change).
func (s Store) SetEnabled(ctx context.Context, key string, enabled bool) (bool, error) {
	var previous bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row flagRow
		err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Where("key = ?", key).Take(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %q", flags.ErrUnknownFlag, key)
		}
		if err != nil {
			return fmt.Errorf("flagsgorm: lock flag %s: %w", key, err)
		}
		previous = row.Enabled
		if previous == enabled {
			return nil
		}
		if err := tx.Model(&row).Update("enabled", enabled).Error; err != nil {
			return fmt.Errorf("flagsgorm: update flag %s: %w", key, err)
		}
		return nil
	})
	return previous, err
}

// applyPresetSQL sets every flag to "is a member of the preset", writing only
// the rows whose state changes.
const applyPresetSQL = `
UPDATE feature_flags AS f
SET enabled = m.member
FROM (
  SELECT ff.key,
         EXISTS (SELECT 1 FROM flag_preset_flags p WHERE p.preset_name = ? AND p.flag_key = ff.key) AS member
  FROM feature_flags ff
) AS m
WHERE f.key = m.key AND f.enabled IS DISTINCT FROM m.member`

// ApplyPreset atomically turns the preset's flags ON and every other flag
// OFF, and returns the resulting flags. The preset row is share-locked so its
// membership cannot change under the update.
func (s Store) ApplyPreset(ctx context.Context, name string) ([]flags.Flag, error) {
	var result []flags.Flag
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var preset presetRow
		err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).
			Where("name = ?", name).Take(&preset).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %q", flags.ErrUnknownPreset, name)
		}
		if err != nil {
			return fmt.Errorf("flagsgorm: lock preset %s: %w", name, err)
		}
		if err := tx.Exec(applyPresetSQL, name).Error; err != nil {
			return fmt.Errorf("flagsgorm: apply preset %s: %w", name, err)
		}
		result, err = list(tx)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Presets returns every preset with its flag keys; presets ordered by name,
// keys by key. Two queries (preset + one Preload), never one per preset.
func (s Store) Presets(ctx context.Context) ([]flags.Preset, error) {
	var rows []presetRow
	err := s.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Order("flag_key") }).
		Order("name").Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("flagsgorm: list presets: %w", err)
	}
	out := make([]flags.Preset, 0, len(rows))
	for _, r := range rows {
		keys := make([]string, 0, len(r.Members))
		for _, m := range r.Members {
			keys = append(keys, m.FlagKey)
		}
		out = append(out, flags.Preset{Name: r.Name, Description: r.Description, Flags: keys})
	}
	return out, nil
}
