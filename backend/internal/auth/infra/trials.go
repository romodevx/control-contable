package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/auth/domain"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// entitlementRow is the trial entitlement sign-up writes.
type entitlementRow struct {
	UserID   string
	Kind     string
	Source   string
	StartsAt time.Time
	EndsAt   time.Time
}

func (entitlementRow) TableName() string { return "entitlements" }

// TrialStore implements usecase.TrialStarter.
type TrialStore struct{}

// NewTrialStore returns the store.
func NewTrialStore() TrialStore { return TrialStore{} }

// StartTrial implements usecase.TrialStarter: a trial entitlement of
// access.TrialDuration from startsAt. The RLS policy only admits a trial
// that starts now (within 5 minutes of the database clock).
func (TrialStore) StartTrial(ctx context.Context, userID domain.UserID, startsAt time.Time) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	row := entitlementRow{
		UserID: userID.String(), Kind: string(access.EntitlementTrial), Source: "registration",
		StartsAt: startsAt, EndsAt: startsAt.Add(access.TrialDuration),
	}
	if err := db.Create(&row).Error; err != nil {
		return fmt.Errorf("auth infra: start trial: %w", err)
	}
	return nil
}
