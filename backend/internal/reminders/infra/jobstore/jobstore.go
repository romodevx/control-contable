// Package jobstore is the reminder job's GORM adapter (M9-06) of the
// reminders tables (migration 20261021101000, M9-02), as cc_worker. It holds
// no pool: every method runs on the unit of work in ctx. DueReminders runs
// on the System unit of work (the due_reminder_users definer function, the
// job's only cross-user read); every other method runs on a per-user unit of
// work (app.current_user_id set per transaction), so RLS scopes each
// statement to that user's rows.
package jobstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

const (
	dueSQL = `SELECT user_id::text AS user_id, to_char(period, 'YYYY-MM') AS period, kind
FROM due_reminder_users(CAST(? AS timestamptz), CAST(? AS boolean), CAST(? AS uuid), CAST(? AS integer))`

	expireSQL = `UPDATE notification_log SET status = 'failed', error_class = 'expired'
WHERE status = 'pending' AND claimed_at < CAST(? AS timestamptz)`

	// claimSQL takes a fresh reminder: the UNIQUE key decides who sends.
	claimSQL = `INSERT INTO notification_log (user_id, period, kind, channel, status, attempts, claimed_at)
VALUES (CAST(? AS uuid), CAST(? AS date), ?, 'push', 'pending', 1, CAST(? AS timestamptz))
ON CONFLICT (user_id, period, kind, channel) DO NOTHING
RETURNING id::text AS id, attempts`

	// reclaimSQL re-claims a pending row whose lease expired, atomically:
	// of two workers racing, the second finds claimed_at moved and gets no row.
	reclaimSQL = `UPDATE notification_log SET attempts = attempts + 1, claimed_at = CAST(? AS timestamptz)
WHERE user_id = CAST(? AS uuid) AND period = CAST(? AS date) AND kind = ? AND channel = 'push'
  AND status = 'pending' AND claimed_at < CAST(? AS timestamptz) AND attempts < 3
RETURNING id::text AS id, attempts`

	// finishSQL records an attempt on the row and attempt the caller holds.
	finishSQL = `UPDATE notification_log
SET status = ?, sent_at = CAST(? AS timestamptz), devices_ok = ?, devices_failed = ?, error_class = ?
WHERE id = CAST(? AS uuid) AND attempts = ? AND status = 'pending'`

	taxpayersSQL = `SELECT t.id::text AS id, t.display_name, tr.regime_code, tr.start_period, tr.end_period
FROM taxpayers t
LEFT JOIN taxpayer_regimes tr ON tr.taxpayer_id = t.id
WHERE t.archived_at IS NULL
ORDER BY t.created_at, t.id, tr.start_period, tr.regime_code`

	targetsSQL = `SELECT id::text AS id, token_ciphertext FROM devices ORDER BY id`

	deleteDevicesSQL = `DELETE FROM devices WHERE id IN ?`

	periodLayout = "2006-01"
	firstOfMonth = "-01"
)

// Store implements the reminder job's database ports.
type Store struct{}

var (
	_ usecase.DueReminderReader = Store{}
	_ usecase.ReminderLog       = Store{}
	_ usecase.RecipientStore    = Store{}
)

// New returns the store.
func New() Store { return Store{} }

type dueRow struct {
	UserID string `gorm:"column:user_id"`
	Period string `gorm:"column:period"`
	Kind   string `gorm:"column:kind"`
}

// DueReminders implements usecase.DueReminderReader.
func (Store) DueReminders(ctx context.Context, now time.Time, dueToday bool, afterUserID string, limit int) ([]usecase.DueReminder, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var after *string
	if afterUserID != "" {
		after = &afterUserID
	}
	var rows []dueRow
	if err := db.Raw(dueSQL, now, dueToday, after, limit).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("reminders jobstore: due users: %w", err)
	}
	out := make([]usecase.DueReminder, 0, len(rows))
	for _, r := range rows {
		p, err := parsePeriod(r.Period)
		if err != nil {
			return nil, err
		}
		kind, err := domain.ParseKind(r.Kind)
		if err != nil {
			return nil, fmt.Errorf("reminders jobstore: due users: %w", err)
		}
		out = append(out, usecase.DueReminder{UserID: r.UserID, Period: p, Kind: kind})
	}
	return out, nil
}

// ExpireStale implements usecase.ReminderLog.
func (Store) ExpireStale(ctx context.Context, claimedBefore time.Time) (int, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	res := db.Exec(expireSQL, claimedBefore)
	if res.Error != nil {
		return 0, fmt.Errorf("reminders jobstore: expire stale: %w", res.Error)
	}
	return int(res.RowsAffected), nil
}

type claimRow struct {
	ID       string `gorm:"column:id"`
	Attempts int    `gorm:"column:attempts"`
}

// Claim implements usecase.ReminderLog.
func (Store) Claim(ctx context.Context, r usecase.DueReminder, now, leaseCutoff time.Time) (usecase.LogClaim, bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.LogClaim{}, false, err
	}
	period := periodDate(r.Period)
	var rows []claimRow
	if err := db.Raw(claimSQL, r.UserID, period, string(r.Kind), now).Scan(&rows).Error; err != nil {
		return usecase.LogClaim{}, false, fmt.Errorf("reminders jobstore: claim: %w", err)
	}
	if len(rows) == 0 {
		err = db.Raw(reclaimSQL, now, r.UserID, period, string(r.Kind), leaseCutoff).Scan(&rows).Error
		if err != nil {
			return usecase.LogClaim{}, false, fmt.Errorf("reminders jobstore: re-claim: %w", err)
		}
	}
	if len(rows) == 0 {
		return usecase.LogClaim{}, false, nil
	}
	return usecase.LogClaim{ID: rows[0].ID, Attempts: rows[0].Attempts}, true, nil
}

// Finish implements usecase.ReminderLog.
func (Store) Finish(ctx context.Context, c usecase.LogClaim, o usecase.LogOutcome) (bool, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return false, err
	}
	var sentAt *time.Time
	if o.Status == usecase.LogSent {
		sentAt = &o.SentAt
	}
	var class *string
	if o.ErrorClass != "" {
		class = &o.ErrorClass
	}
	res := db.Exec(finishSQL, string(o.Status), sentAt, o.DevicesOK, o.DevicesFailed, class, c.ID, c.Attempts)
	if res.Error != nil {
		return false, fmt.Errorf("reminders jobstore: finish: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

type taxpayerRow struct {
	ID          string  `gorm:"column:id"`
	DisplayName string  `gorm:"column:display_name"`
	RegimeCode  *string `gorm:"column:regime_code"`
	StartPeriod *string `gorm:"column:start_period"`
	EndPeriod   *string `gorm:"column:end_period"`
}

// Taxpayers implements usecase.RecipientStore: one statement, one row per
// (taxpayer, regime), folded in the switcher's order.
func (Store) Taxpayers(ctx context.Context) ([]usecase.RecipientTaxpayer, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []taxpayerRow
	if err := db.Raw(taxpayersSQL).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("reminders jobstore: taxpayers: %w", err)
	}
	var out []usecase.RecipientTaxpayer
	for _, r := range rows {
		if len(out) == 0 || out[len(out)-1].ID != r.ID {
			out = append(out, usecase.RecipientTaxpayer{ID: r.ID, DisplayName: r.DisplayName})
		}
		if r.RegimeCode == nil {
			continue
		}
		held, err := heldRegime(*r.RegimeCode, r.StartPeriod, r.EndPeriod)
		if err != nil {
			return nil, err
		}
		last := &out[len(out)-1]
		last.Regimes = append(last.Regimes, held)
	}
	return out, nil
}

func heldRegime(code string, start, end *string) (usecase.HeldRegime, error) {
	held := usecase.HeldRegime{Code: code}
	if start == nil {
		return held, fmt.Errorf("reminders jobstore: regime %s without a start", code)
	}
	var err error
	if held.Start, err = parsePeriod(*start); err != nil {
		return held, err
	}
	if end != nil {
		if held.End, err = parsePeriod(*end); err != nil {
			return held, err
		}
	}
	return held, nil
}

type targetRow struct {
	ID         string `gorm:"column:id"`
	Ciphertext []byte `gorm:"column:token_ciphertext"`
}

// Targets implements usecase.RecipientStore.
func (Store) Targets(ctx context.Context) ([]usecase.SealedTarget, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []targetRow
	if err := db.Raw(targetsSQL).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("reminders jobstore: devices: %w", err)
	}
	out := make([]usecase.SealedTarget, len(rows))
	for i, r := range rows {
		out[i] = usecase.SealedTarget{DeviceID: r.ID, Ciphertext: r.Ciphertext}
	}
	return out, nil
}

// DeleteDevices implements usecase.RecipientStore. Ids that are not uuids
// match nothing.
func (Store) DeleteDevices(ctx context.Context, ids []string) (int, error) {
	valid := make([]string, 0, len(ids))
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			valid = append(valid, u.String())
		}
	}
	if len(valid) == 0 {
		return 0, nil
	}
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	res := db.Exec(deleteDevicesSQL, valid)
	if res.Error != nil {
		return 0, fmt.Errorf("reminders jobstore: delete devices: %w", res.Error)
	}
	return int(res.RowsAffected), nil
}

// parsePeriod parses "YYYY-MM" (char(7) columns and to_char output).
func parsePeriod(raw string) (domain.Period, error) {
	t, err := time.Parse(periodLayout, strings.TrimSpace(raw))
	if err != nil {
		return domain.Period{}, fmt.Errorf("reminders jobstore: period %q: %w", raw, err)
	}
	return domain.PeriodOf(t)
}

// periodDate is the period as notification_log.period: its first day.
func periodDate(p domain.Period) string { return p.String() + firstOfMonth }
