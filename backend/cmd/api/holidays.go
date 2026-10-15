package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	remindersholidays "github.com/romodevx/control-contable/backend/internal/reminders/infra/holidays"
	taxholidays "github.com/romodevx/control-contable/backend/internal/taxengine/infra/holidays"
)

// errHolidaysWiring reports a boot read of the holidays without its ports.
var errHolidaysWiring = errors.New("holidays: the system unit of work and the module's reader are required")

// holidayReader is the shape of each module's holidays reader (the tax
// engine's and the reminders'; ADR-0007 Decision 6 keeps one per module).
type holidayReader interface {
	Holidays(ctx context.Context) (fiscal.HolidaySet, error)
}

// holidaySnapshot is the holidays read once at boot (loadHolidays): it
// implements every module's HolidayCalendar port from memory, so a request
// spends no statement and no unit of work on them. The tables are global,
// tiny and never written at run time (a new or corrected year is a
// migration, so a deploy, so a restart: migration 20261021100000).
type holidaySnapshot struct {
	set fiscal.HolidaySet
}

// Holidays returns the boot snapshot. HolidaySet is immutable, so sharing
// it across requests is safe.
func (s holidaySnapshot) Holidays(context.Context) (fiscal.HolidaySet, error) {
	return s.set, nil
}

// loadHolidays reads the holidays through r in one system unit of work
// (the tables are global: no user context; RLS lets cc_app SELECT them).
// A failed read fails the boot rather than serving weekends-only dates.
func loadHolidays(ctx context.Context, system tx.SystemManager, r holidayReader) (holidaySnapshot, error) {
	if system == nil || r == nil {
		return holidaySnapshot{}, errHolidaysWiring
	}
	var set fiscal.HolidaySet
	err := system.System(ctx, func(ctx context.Context) error {
		var err error
		set, err = r.Holidays(ctx)
		return err
	})
	if err != nil {
		return holidaySnapshot{}, fmt.Errorf("holidays: boot read: %w", err)
	}
	return holidaySnapshot{set: set}, nil
}

// holidaySnapshots are each module's boot snapshot, each read through the
// module's own reader so either module stays removable (ADR-0007 Decision
// 6): two statements per boot, none per request.
type holidaySnapshots struct {
	taxEngine holidaySnapshot
	reminders holidaySnapshot
}

// loadHolidaySnapshots reads the holidays at boot through the shared cc_app
// unit of work (its system transactions).
func loadHolidaySnapshots(ctx context.Context, system tx.SystemManager) (holidaySnapshots, error) {
	taxEngine, err := loadHolidays(ctx, system, taxholidays.NewReader())
	if err != nil {
		return holidaySnapshots{}, fmt.Errorf("taxengine: %w", err)
	}
	reminders, err := loadHolidays(ctx, system, remindersholidays.NewReader())
	if err != nil {
		return holidaySnapshots{}, fmt.Errorf("reminders: %w", err)
	}
	return holidaySnapshots{taxEngine: taxEngine, reminders: reminders}, nil
}
