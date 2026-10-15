package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/reminders/domain"
)

// The reminder job's schedule and bounds (MODULE-reminders §2.4).
const (
	reminderJobName     = "reminders"
	reminderJobInterval = 5 * time.Minute

	// DefaultJobConcurrency is how many users one run serves at once.
	DefaultJobConcurrency = 8
	// DefaultJobBatchSize is how many due reminders one due_reminder_users
	// call returns.
	DefaultJobBatchSize = 100
	// MaxJobBatchSize mirrors the bound due_reminder_users enforces.
	MaxJobBatchSize = 1000
	// MinCollapseKeyBytes is the shortest collapse-id HMAC key accepted.
	MinCollapseKeyBytes = 32

	// leaseDuration is how long a claim is owned before another worker may
	// re-claim it (§2.4 step 5); far above the sender's 30 s deadline.
	leaseDuration = 10 * time.Minute
	// maxAttempts is the last attempt a transient failure is retried at.
	maxAttempts = 3
	// regimeBimonthly is the 621 regime: its payment follows the bimester.
	regimeBimonthly = "621"
)

// The error classes the job writes (notification_log.error_class); the
// sender's own classes pass through when well formed.
const (
	classExpired             = "expired"
	classNoDevice            = "no_device"
	classTokenUnreadable     = "token_unreadable"
	classPushDisabled        = "push_disabled"
	classInvalidNotification = "invalid_notification"
	classSendError           = "send_error"
)

// errorClassPattern is the notification_log CHECK on error_class.
var errorClassPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,39}$`)

// ErrReminderFailures reports a run in which some users' reminders failed
// with an error (each one logged; the others were served).
var ErrReminderFailures = errors.New("reminders: some reminders failed")

// JobDeps are the ports of the ReminderJob.
type JobDeps struct {
	// System runs the cross-user due_reminder_users read (no user context).
	System tx.SystemManager
	// Units runs each user's claim and record transactions (cc_worker,
	// app.current_user_id set per transaction).
	Units      tx.Manager
	Due        DueReminderReader
	Log        ReminderLog
	Settings   SettingsStore
	Recipients RecipientStore
	Opener     TokenOpener
	Push       PushSender
	Estimates  EstimateReader
	Flags      Flags
	Holidays   HolidayCalendar
	Alerts     Alerts
	Clock      clock.Clock
	Logger     *slog.Logger
	// CollapseKey keys the notifications' collapse ids (§2.4 step 5).
	CollapseKey []byte
	Concurrency int
	BatchSize   int
}

// ReminderJob is the worker job (worker.Job) that sends the monthly
// reminders (MODULE-reminders §2.4): every 5 minutes, on the leader, it
// selects the users whose reminder of today is due and, for each, claims
// the notification_log row, sends one push to their devices and records the
// outcome. The UNIQUE log key, not the scheduler, makes a reminder
// exactly-once under a leader split brain; a crash after the claim is
// retried after the lease with the same collapse id. It is safe for
// concurrent use (two instances may run it at once).
type ReminderJob struct {
	deps JobDeps
}

// NewReminderJob validates deps and builds the job.
func NewReminderJob(deps JobDeps) (*ReminderJob, error) {
	if err := validateJobDeps(deps); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMissingDependency, err)
	}
	return &ReminderJob{deps: deps}, nil
}

func validateJobDeps(d JobDeps) error {
	ports := []struct {
		name    string
		missing bool
	}{
		{"System", d.System == nil}, {"Units", d.Units == nil}, {"Due", d.Due == nil}, {"Log", d.Log == nil},
		{"Settings", d.Settings == nil}, {"Recipients", d.Recipients == nil}, {"Opener", d.Opener == nil},
		{"Push", d.Push == nil}, {"Estimates", d.Estimates == nil}, {"Flags", d.Flags == nil},
		{"Holidays", d.Holidays == nil}, {"Alerts", d.Alerts == nil}, {"Clock", d.Clock == nil}, {"Logger", d.Logger == nil},
	}
	for _, p := range ports {
		if p.missing {
			return fmt.Errorf("JobDeps.%s", p.name)
		}
	}
	switch {
	case len(d.CollapseKey) < MinCollapseKeyBytes:
		return fmt.Errorf("JobDeps.CollapseKey shorter than %d bytes", MinCollapseKeyBytes)
	case d.Concurrency < 1:
		return fmt.Errorf("JobDeps.Concurrency %d is not positive", d.Concurrency)
	case d.BatchSize < 1 || d.BatchSize > MaxJobBatchSize:
		return fmt.Errorf("JobDeps.BatchSize %d not in [1, %d]", d.BatchSize, MaxJobBatchSize)
	}
	return nil
}

// Name implements worker.Job.
func (j *ReminderJob) Name() string { return reminderJobName }

// Interval implements worker.Job.
func (j *ReminderJob) Interval() time.Duration { return reminderJobInterval }

// runPlan is what one run decides once, for every user.
type runPlan struct {
	now        time.Time
	today      fiscal.Date
	startOfDay time.Time
	holidays   fiscal.HolidaySet
	dueToday   bool
	amountsOn  bool // module_tax_engine effectively ON
}

// Run implements worker.Job. A run with module_reminders OFF does nothing.
// It returns ctx's error when cancelled (claimed reminders stay pending for
// the lease retry) and ErrReminderFailures when some users failed.
func (j *ReminderJob) Run(ctx context.Context) error {
	if !j.deps.Flags.Enabled(ctx, ModuleReminders) {
		j.deps.Logger.DebugContext(ctx, "reminders skipped: module off", slog.String("job", reminderJobName))
		return nil
	}
	plan, err := j.plan(ctx)
	if err != nil {
		return err
	}
	var counts runCounts
	err = j.serveAll(ctx, plan, &counts)
	j.report(ctx, &counts)
	if err != nil {
		return err
	}
	if counts.errors > 0 {
		return fmt.Errorf("%w: %d", ErrReminderFailures, counts.errors)
	}
	return nil
}

// plan reads the holidays and decides whether today is the due date of the
// period the reminders of today are about (the month before today's).
func (j *ReminderJob) plan(ctx context.Context) (runPlan, error) {
	now := j.deps.Clock.Now().In(clock.Location())
	var holidays fiscal.HolidaySet
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		holidays, err = j.deps.Holidays.Holidays(ctx)
		return err
	})
	if err != nil {
		return runPlan{}, fmt.Errorf("reminders: holidays: %w", err)
	}
	today, err := fiscal.ParseDate(now.Format(time.DateOnly))
	if err != nil {
		return runPlan{}, fmt.Errorf("reminders: today: %w", err)
	}
	if !holidays.Covers(today.Year()) {
		j.deps.Alerts.CalendarUnseeded(ctx, today.Year())
	}
	previous := now.AddDate(0, 0, -now.Day()+1).AddDate(0, -1, 0)
	_, due, _, err := fiscal.MonthlyDueDate(previous.Year(), previous.Month(), false, holidays)
	if err != nil {
		return runPlan{}, fmt.Errorf("reminders: due date: %w", err)
	}
	return runPlan{
		now: now, today: today, holidays: holidays, dueToday: due == today,
		startOfDay: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, clock.Location()),
		amountsOn:  j.deps.Flags.Enabled(ctx, ModuleTaxEngine),
	}, nil
}

// serveAll pages through the due reminders and serves each page with at
// most Concurrency users at once.
func (j *ReminderJob) serveAll(ctx context.Context, plan runPlan, counts *runCounts) error {
	after := ""
	for {
		batch, err := j.dueBatch(ctx, plan, after)
		if err != nil {
			return err
		}
		if err := j.serveBatch(ctx, plan, batch, counts); err != nil {
			return err
		}
		if len(batch) < j.deps.BatchSize {
			return nil
		}
		after = batch[len(batch)-1].UserID
	}
}

func (j *ReminderJob) dueBatch(ctx context.Context, plan runPlan, after string) ([]DueReminder, error) {
	var batch []DueReminder
	err := j.deps.System.System(ctx, func(ctx context.Context) error {
		var err error
		batch, err = j.deps.Due.DueReminders(ctx, plan.now, plan.dueToday, after, j.deps.BatchSize)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("reminders: due users: %w", err)
	}
	return batch, nil
}

// serveBatch serves every reminder of batch. A user's error is logged and
// counted, never stops the others; only cancellation ends the batch early.
func (j *ReminderJob) serveBatch(ctx context.Context, plan runPlan, batch []DueReminder, counts *runCounts) error {
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(j.deps.Concurrency)
	for _, r := range batch {
		g.Go(func() error {
			if err := gctx.Err(); err != nil {
				return err
			}
			result, err := j.serve(gctx, plan, r)
			if err != nil && gctx.Err() != nil {
				return gctx.Err()
			}
			counts.add(result, err)
			if err != nil {
				j.deps.Logger.ErrorContext(gctx, "reminder failed",
					slog.String("user_id", r.UserID), slog.String("period", r.Period.String()), slog.String("kind", string(r.Kind)),
					slog.Any("error", err))
			}
			return nil
		})
	}
	return g.Wait()
}

// userResult is how one user's reminder ended in this run.
type userResult struct {
	kind         domain.Kind
	claimed      bool
	status       LogStatus
	errorClass   string
	expired      int
	unregistered int
	lost         bool // the row moved on before the record
}

// claimed is what the claim transaction read for the send.
type claimedReminder struct {
	claim     LogClaim
	settings  domain.Settings
	taxpayers []RecipientTaxpayer
	due       fiscal.Date
	targets   []SealedTarget
}

// serve runs one user's reminder: claim (and read) in one transaction,
// build and send outside it, record in a second short one.
func (j *ReminderJob) serve(ctx context.Context, plan runPlan, r DueReminder) (userResult, error) {
	result := userResult{kind: r.Kind}
	var c claimedReminder
	err := j.deps.Units.ReadWrite(ctx, r.UserID, func(ctx context.Context) error {
		var err error
		result, c, err = j.claim(ctx, plan, r)
		return err
	})
	if err != nil || !result.claimed || result.status != LogPending {
		return result, err
	}
	notification, err := j.notification(ctx, plan, r, c)
	if err != nil {
		return result, err
	}
	outcome, invalid := j.send(ctx, plan, r, c, notification)
	if err := ctx.Err(); err != nil && outcome.Status != LogSent {
		return result, err // cut short: pending, retried after the lease
	}
	return j.record(ctx, r, c.claim, outcome, invalid, result)
}

// claim expires the user's stale pending rows, claims r and reads what the
// message needs. A reminder with nothing to say is recorded skipped here.
func (j *ReminderJob) claim(ctx context.Context, plan runPlan, r DueReminder) (userResult, claimedReminder, error) {
	result := userResult{kind: r.Kind}
	expired, err := j.deps.Log.ExpireStale(ctx, plan.startOfDay)
	if err != nil {
		return result, claimedReminder{}, fmt.Errorf("expire stale: %w", err)
	}
	result.expired = expired
	claim, ok, err := j.deps.Log.Claim(ctx, r, plan.now, plan.now.Add(-leaseDuration))
	if err != nil || !ok {
		return result, claimedReminder{}, wrapIf("claim", err)
	}
	result.claimed, result.status = true, LogPending
	c, err := j.read(ctx, plan, r, claim)
	if err != nil {
		return result, claimedReminder{}, err
	}
	if len(c.taxpayers) > 0 {
		return result, c, nil
	}
	result.status = LogSkipped
	if _, err := j.deps.Log.Finish(ctx, claim, LogOutcome{Status: LogSkipped}); err != nil {
		return result, claimedReminder{}, fmt.Errorf("record skipped: %w", err)
	}
	return result, c, nil
}

// read loads the user's settings, the taxpayers due on the reminder's date
// and the devices. Settings turned off since the selection leave no
// taxpayer (skipped).
func (j *ReminderJob) read(ctx context.Context, plan runPlan, r DueReminder, claim LogClaim) (claimedReminder, error) {
	c := claimedReminder{claim: claim}
	settings, found, err := j.deps.Settings.Settings(ctx, r.UserID)
	if err != nil {
		return c, fmt.Errorf("settings: %w", err)
	}
	if !found || !settings.Enabled {
		return c, nil
	}
	c.settings = settings
	taxpayers, err := j.deps.Recipients.Taxpayers(ctx)
	if err != nil {
		return c, fmt.Errorf("taxpayers: %w", err)
	}
	if c.taxpayers, c.due, err = dueTaxpayers(taxpayers, r, plan); err != nil {
		return c, err
	}
	if len(c.taxpayers) == 0 {
		return c, nil
	}
	if c.targets, err = j.deps.Recipients.Targets(ctx); err != nil {
		return c, fmt.Errorf("devices: %w", err)
	}
	return c, nil
}

// dueTaxpayers keeps the taxpayers whose due date of r's period is the
// reminder's (§2.4 step 6): today for the due reminder, a day of this month
// for the early one. A taxpayer holding no regime in the period owes
// nothing for it; one whose every regime is 621 pays by bimester.
func dueTaxpayers(all []RecipientTaxpayer, r DueReminder, plan runPlan) ([]RecipientTaxpayer, fiscal.Date, error) {
	var (
		out []RecipientTaxpayer
		due fiscal.Date
	)
	for _, t := range all {
		held := heldIn(t.Regimes, r.Period)
		if len(held) == 0 {
			continue
		}
		_, d, _, err := fiscal.MonthlyDueDate(r.Period.Year(), r.Period.Month(), onlyBimonthly(held), plan.holidays)
		if err != nil {
			return nil, fiscal.Date{}, fmt.Errorf("due date of %s: %w", r.Period, err)
		}
		if remindsOn(r.Kind, d, plan.today) {
			out, due = append(out, t), d
		}
	}
	return out, due, nil
}

func remindsOn(kind domain.Kind, due, today fiscal.Date) bool {
	if kind == domain.KindDue {
		return due == today
	}
	return due.Period() == today.Period() && !due.Before(today)
}

func heldIn(regimes []HeldRegime, p domain.Period) []HeldRegime {
	var out []HeldRegime
	for _, h := range regimes {
		if !after(h.Start, p) && (h.End.IsZero() || !after(p, h.End)) {
			out = append(out, h)
		}
	}
	return out
}

// after reports a later than b.
func after(a, b domain.Period) bool {
	return a.Year() > b.Year() || (a.Year() == b.Year() && a.Month() > b.Month())
}

func onlyBimonthly(held []HeldRegime) bool {
	for _, h := range held {
		if h.Code != regimeBimonthly {
			return false
		}
	}
	return len(held) > 0
}

// notification builds the message: names in the switcher's order and,
// when the user opted in and the engine is ON, the Estimator's totals.
func (j *ReminderJob) notification(ctx context.Context, plan runPlan, r DueReminder, c claimedReminder) (PushNotification, error) {
	listed := make([]domain.Listed, len(c.taxpayers))
	withAmounts := c.settings.AmountsShown(plan.amountsOn)
	for i, t := range c.taxpayers {
		listed[i] = domain.Listed{Name: t.DisplayName}
		if withAmounts {
			listed[i] = j.withEstimate(ctx, r, t, listed[i])
		}
	}
	dueDay, err := time.Parse(time.DateOnly, c.due.String())
	if err != nil {
		return PushNotification{}, fmt.Errorf("due date: %w", err)
	}
	msg, err := domain.BuildMessage(domain.MessageInput{Kind: r.Kind, Period: r.Period, DueDate: dueDay, Taxpayers: listed})
	if err != nil {
		return PushNotification{}, fmt.Errorf("message: %w", err)
	}
	return PushNotification{
		Title: msg.Title, Body: msg.Body, DeepLink: msg.DeepLink,
		CollapseID: domain.CollapseID(j.deps.CollapseKey, r.UserID, r.Period, r.Kind),
	}, nil
}

// withEstimate adds t's estimate to l (§2.5): its total when complete and
// positive, "estimado pendiente" when incomplete, else the name only. An
// estimator error never blocks the reminder: the name only, the class
// logged (never the error text, which could name a record).
func (j *ReminderJob) withEstimate(ctx context.Context, r DueReminder, t RecipientTaxpayer, l domain.Listed) domain.Listed {
	summary, err := j.deps.Estimates.MonthlyEstimate(ctx, r.UserID, t.ID, r.Period)
	if err != nil {
		j.deps.Logger.WarnContext(ctx, "reminder sent without amounts: estimate unavailable",
			slog.String("user_id", r.UserID), slog.String("period", r.Period.String()),
			slog.String("error_class", "estimate_unavailable"))
		return l
	}
	switch {
	case summary.Status == EstimateComplete && summary.TotalPayableCents > 0:
		l.Estimate, l.AmountCents = domain.EstimateAmount, summary.TotalPayableCents
	case summary.Status == EstimateIncomplete:
		l.Estimate = domain.EstimatePending
	}
	return l
}

// send opens the device tokens and sends n to them in batches of
// MaxPushTargets. It returns the attempt's outcome and the devices whose
// token the provider declared invalid.
func (j *ReminderJob) send(ctx context.Context, plan runPlan, r DueReminder, c claimedReminder, n PushNotification) (LogOutcome, []string) {
	targets, unreadable := j.open(r.UserID, c.targets)
	tally := deliveryTally{failed: unreadable}
	if unreadable > 0 {
		tally.class = classTokenUnreadable
	}
	if len(targets) == 0 {
		return tally.outcome(plan.now, c.claim.Attempts, classNoDevice), nil
	}
	for start := 0; start < len(targets); start += MaxPushTargets {
		chunk := targets[start:min(start+MaxPushTargets, len(targets))]
		deliveries, err := j.deps.Push.Send(ctx, n, chunk)
		if err != nil {
			tally.wholeCall(err, len(chunk))
			continue
		}
		tally.addAll(deliveries)
	}
	return tally.outcome(plan.now, c.claim.Attempts, classNoDevice), tally.invalid
}

// open decrypts each device's token; a token that does not open is counted
// as a failed device and skipped.
func (j *ReminderJob) open(userID string, sealed []SealedTarget) ([]PushTarget, int) {
	targets := make([]PushTarget, 0, len(sealed))
	unreadable := 0
	for _, s := range sealed {
		id, err := uuid.Parse(s.DeviceID)
		if err != nil {
			unreadable++
			continue
		}
		token, err := j.deps.Opener.Open(s.Ciphertext, TokenBinding{InstallationID: s.DeviceID, UserID: userID})
		if err != nil {
			unreadable++
			continue
		}
		targets = append(targets, PushTarget{DeviceID: id, Token: NewPushToken(token.Reveal())})
	}
	return targets, unreadable
}

// deliveryTally accumulates one attempt's per-device results.
type deliveryTally struct {
	ok, failed, transient int
	permanentCall         bool // a whole-call error that retrying cannot fix
	class                 string
	invalid               []string
}

func (t *deliveryTally) addAll(deliveries []PushDelivery) {
	for _, d := range deliveries {
		switch d.Outcome {
		case PushDelivered:
			t.ok++
			continue
		case PushTransient:
			t.transient++
		case PushTokenInvalid:
			t.invalid = append(t.invalid, d.DeviceID.String())
		case PushRejected:
		}
		t.failed++
		t.setClass(sanitizeClass(d.ErrorClass, d.Outcome.String()))
	}
}

func (t *deliveryTally) wholeCall(err error, devices int) {
	t.failed += devices
	switch {
	case errors.Is(err, ErrPushDisabled):
		t.permanentCall = true
		t.setClass(classPushDisabled)
	case errors.Is(err, ErrInvalidPushNotification), errors.Is(err, ErrTooManyPushTargets):
		t.permanentCall = true
		t.setClass(classInvalidNotification)
	default:
		t.transient += devices
		t.setClass(classSendError)
	}
}

func (t *deliveryTally) setClass(class string) {
	if t.class == "" {
		t.class = class
	}
}

// outcome applies §2.4 step 3: any device reached → sent; only transient
// failures with attempts left → still pending; else failed.
func (t *deliveryTally) outcome(now time.Time, attempts int, emptyClass string) LogOutcome {
	out := LogOutcome{DevicesOK: t.ok, DevicesFailed: t.failed, ErrorClass: t.class}
	switch {
	case t.ok > 0:
		out.Status, out.SentAt = LogSent, now
	case t.ok+t.failed == 0:
		out.Status, out.ErrorClass = LogFailed, emptyClass
	case t.transient > 0 && !t.permanentCall && attempts < maxAttempts:
		out.Status = LogPending
	default:
		out.Status = LogFailed
	}
	return out
}

// sanitizeClass keeps a sender class that fits the column, else fallback.
func sanitizeClass(class, fallback string) string {
	if errorClassPattern.MatchString(class) {
		return class
	}
	return fallback
}

// record deletes the invalid devices and records the outcome, in the
// user's second short transaction. It runs on a context that the run's
// cancellation does not cut: the send already happened.
func (j *ReminderJob) record(ctx context.Context, r DueReminder, claim LogClaim, o LogOutcome, invalid []string, result userResult) (userResult, error) {
	err := j.deps.Units.ReadWrite(context.WithoutCancel(ctx), r.UserID, func(ctx context.Context) error {
		if len(invalid) > 0 {
			n, err := j.deps.Recipients.DeleteDevices(ctx, invalid)
			if err != nil {
				return fmt.Errorf("delete invalid devices: %w", err)
			}
			result.unregistered = n
		}
		recorded, err := j.deps.Log.Finish(ctx, claim, o)
		if err != nil {
			return fmt.Errorf("record: %w", err)
		}
		result.lost = !recorded
		return nil
	})
	if err != nil {
		return userResult{kind: r.Kind, claimed: true, status: LogPending}, err
	}
	result.status, result.errorClass = o.Status, o.ErrorClass
	return result, nil
}

func wrapIf(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// runCounts are one run's metrics (§9): counts only, no PII.
type runCounts struct {
	mu                  sync.Mutex
	sent                map[domain.Kind]int
	failed              map[string]int
	pending             int
	skipped             int
	notClaimed          int
	expired             int
	devicesUnregistered int
	lost                int
	errors              int
}

func (c *runCounts) add(r userResult, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.expired += r.expired
	c.devicesUnregistered += r.unregistered
	switch {
	case err != nil:
		c.errors++
	case !r.claimed:
		c.notClaimed++
	case r.lost:
		c.lost++
	case r.status == LogSent:
		if c.sent == nil {
			c.sent = map[domain.Kind]int{}
		}
		c.sent[r.kind]++
	case r.status == LogSkipped:
		c.skipped++
	case r.status == LogPending:
		c.pending++
	case r.status == LogFailed:
		if c.failed == nil {
			c.failed = map[string]int{}
		}
		c.failed[r.errorClass]++
	}
}

// report logs the run's counts: reminders_sent_total{kind},
// reminders_failed_total{error_class}, reminders_skipped_total,
// devices_unregistered_total (MODULE-reminders §9).
func (j *ReminderJob) report(ctx context.Context, c *runCounts) {
	c.mu.Lock()
	defer c.mu.Unlock()
	failed := make([]any, 0, len(c.failed))
	total := 0
	for class, n := range c.failed {
		failed = append(failed, slog.Int(class, n))
		total += n
	}
	j.deps.Logger.InfoContext(ctx, "reminders run finished",
		slog.String("job", reminderJobName),
		slog.Int("sent", c.sent[domain.KindEarly]+c.sent[domain.KindDue]),
		slog.Group("sent_by_kind", slog.Int(string(domain.KindEarly), c.sent[domain.KindEarly]), slog.Int(string(domain.KindDue), c.sent[domain.KindDue])),
		slog.Int("failed", total), slog.Group("failed_by_class", failed...),
		slog.Int("pending_retry", c.pending), slog.Int("skipped", c.skipped), slog.Int("not_claimed", c.notClaimed),
		slog.Int("expired", c.expired), slog.Int("devices_unregistered", c.devicesUnregistered),
		slog.Int("lost_lease", c.lost), slog.Int("errors", c.errors))
}
