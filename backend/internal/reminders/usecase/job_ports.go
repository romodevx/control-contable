package usecase

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

// The module keys the reminder job checks on every run (MODULE-reminders §3).
const (
	// ModuleReminders gates the job: OFF (or a prerequisite OFF) → no-op.
	ModuleReminders = "module_reminders"
	// ModuleTaxEngine gates the amounts (§2.5): OFF → names only.
	ModuleTaxEngine = "module_tax_engine"
)

// Flags answers whether a module is effectively ON (its prerequisites
// included), failing closed (flags.Service).
type Flags interface {
	Enabled(ctx context.Context, module string) bool
}

// DueReminder is one reminder the database says is due now: a user, the
// period it is about and its kind.
type DueReminder struct {
	UserID string
	Period domain.Period
	Kind   domain.Kind
}

// DueReminderReader is the job's one cross-user read, the
// due_reminder_users definer function (§2.4 steps 2 and 5). It runs on the
// System unit of work in ctx (no user context).
type DueReminderReader interface {
	// DueReminders lists, ordered by user id after afterUserID ("" = from
	// the start), up to limit reminders of today due by now: kind due when
	// dueToday (today is the period's moved due date), else early. A user
	// whose reminder has a log row is left out, unless that row is pending
	// with an expired lease and attempts left (the retry).
	DueReminders(ctx context.Context, now time.Time, dueToday bool, afterUserID string, limit int) ([]DueReminder, error)
}

// LogStatus is a notification_log row's status.
type LogStatus string

// The notification_log statuses. Only pending changes, once.
const (
	LogPending LogStatus = "pending"
	LogSent    LogStatus = "sent"
	LogFailed  LogStatus = "failed"
	LogSkipped LogStatus = "skipped"
)

// LogClaim is the caller's ownership of a notification_log row: its id and
// the attempt it holds. A record of the row matches both, so a slow sender
// whose lease another worker took over cannot overwrite the newer attempt.
type LogClaim struct {
	ID       string
	Attempts int
}

// LogOutcome is how an attempt ended. A pending outcome keeps the row
// pending (a transient failure, retried after the lease) with ErrorClass.
type LogOutcome struct {
	Status        LogStatus
	SentAt        time.Time
	DevicesOK     int
	DevicesFailed int
	// ErrorClass is a machine class (^[a-z][a-z0-9_]{0,39}$), never a
	// message, a token or a value; "" for none.
	ErrorClass string
}

// ReminderLog is the notification_log table (M9-02) under the caller's
// RLS: every method runs on the per-user unit of work in ctx.
type ReminderLog interface {
	// ExpireStale marks failed, error class "expired", the caller's pending
	// rows claimed before claimedBefore (the start of today in Mexico
	// City): a reminder whose day passed is never sent late (§2.4 step 5,
	// OQ-R5). It returns how many rows it expired.
	ExpireStale(ctx context.Context, claimedBefore time.Time) (int, error)
	// Claim takes r for the caller: a new pending row (attempt 1), or else
	// a re-claim of its pending row whose lease started before leaseCutoff
	// with attempts left (attempts + 1). claimed is false when someone else
	// holds the row or it is final: the caller stops.
	Claim(ctx context.Context, r DueReminder, now, leaseCutoff time.Time) (c LogClaim, claimed bool, err error)
	// Finish records o on the row c holds. recorded is false when the row
	// moved on (re-claimed by another worker after the lease, or final).
	Finish(ctx context.Context, c LogClaim, o LogOutcome) (recorded bool, err error)
}

// HeldRegime is one regime a taxpayer holds or held, by its period window
// (End is the zero Period while it is open).
type HeldRegime struct {
	Code  string
	Start domain.Period
	End   domain.Period
}

// RecipientTaxpayer is one of the caller's live (non-archived) taxpayers
// as a reminder needs it: its id (for the estimate only, never in a
// payload), its display name and its regimes.
type RecipientTaxpayer struct {
	ID          string
	DisplayName string
	Regimes     []HeldRegime
}

// SealedTarget is one of the caller's devices as stored: its installation
// id and its sealed token (opened by a TokenOpener).
type SealedTarget struct {
	DeviceID   string
	Ciphertext []byte
}

// RecipientStore reads and cleans the caller's reminder audience under
// RLS, on the per-user unit of work in ctx.
type RecipientStore interface {
	// Taxpayers lists the caller's live taxpayers in the switcher's order
	// (created_at, id), each with its regimes.
	Taxpayers(ctx context.Context) ([]RecipientTaxpayer, error)
	// Targets lists the caller's devices with their sealed tokens.
	Targets(ctx context.Context) ([]SealedTarget, error)
	// DeleteDevices removes the caller's devices of ids (FCM UNREGISTERED,
	// §2.4) and returns how many it removed.
	DeleteDevices(ctx context.Context, ids []string) (int, error)
}

// TokenOpener opens a sealed push token of the row b names
// (pushtoken.Vault.Open).
type TokenOpener interface {
	Open(ciphertext []byte, b TokenBinding) (domain.PushToken, error)
}

// EstimateStatus is the taxpayer-level status of a monthly estimate.
type EstimateStatus int

// The estimate statuses a reminder distinguishes (§2.5).
const (
	// EstimateOther: informational, partial or no regime; listed by name.
	EstimateOther EstimateStatus = iota
	// EstimateComplete: every regime computed; its total may be shown.
	EstimateComplete
	// EstimateIncomplete: inputs missing; "estimado pendiente".
	EstimateIncomplete
)

// EstimateSummary is what a reminder shows of a taxpayer's estimate: its
// status and its rounded total payable, in centavos (the Estimator's,
// never recomputed).
type EstimateSummary struct {
	Status            EstimateStatus
	TotalPayableCents int64
}

// EstimateReader is the tax engine's monthly Estimator (M5-22) as the job
// reads it; the composition root adapts it. It opens its own read-only
// unit of work for userID.
type EstimateReader interface {
	MonthlyEstimate(ctx context.Context, userID, taxpayerID string, p domain.Period) (EstimateSummary, error)
}
