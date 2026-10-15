package annual

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// eventFiledUpdated is the audit event of putFiledPayments (ids only).
const eventFiledUpdated = "annual_filed_payments.updated"

// FiledRequest names the filed payments of one taxpayer, regime and fiscal
// year as the client sent them (ids not validated yet).
type FiledRequest struct {
	TaxpayerID string
	RegimeCode string
	FiscalYear int
}

// FiledEntry is one month of a putFiledPayments body.
type FiledEntry struct {
	Month  int
	Amount money.Cents
}

// FiledMonthView is one month of the year as stored: whether the regime is
// held in it, whether it is closed (only a closed month can carry a
// payment) and the amount the user confirmed (nil: not confirmed). The
// tax-input and app-estimate columns and the precedence are the annual
// estimate's (ANN-14).
type FiledMonthView struct {
	Month    time.Month
	Period   domain.Period
	InWindow bool
	Closed   bool
	Filed    *money.Cents
}

// FiledView is the 12 months of a regime and year, January first;
// UpdatedAt is nil when nothing is confirmed.
type FiledView struct {
	Regime     fiscal.RegimeCode
	FiscalYear int
	Months     []FiledMonthView
	UpdatedAt  *time.Time
}

// FiledDeps are the ports the filed-payments use cases need.
type FiledDeps struct {
	Tx        tx.Manager
	Access    AccessPolicy
	Taxpayers TaxpayerStore
	Store     FiledStore
	Clock     clock.Clock
	// Logger receives the audit events; nil discards them.
	Logger *slog.Logger
}

// FiledPayments stores the provisional payments the user actually filed
// (MODULE-annual §2.5 source 1; getFiledPayments, putFiledPayments). The
// screen's "Confirmar todos" is a PUT of the prefilled months: it stores
// them as filed like any typed value.
type FiledPayments struct {
	base
	deps FiledDeps
}

// NewFiledPayments returns the use cases over deps.
func NewFiledPayments(deps FiledDeps) *FiledPayments {
	return &FiledPayments{base: newBase(deps.Clock, deps.Logger), deps: deps}
}

// Get returns the stored months of req. Order: taxpayer id (404) → future
// year (400) → taxpayer (404) → regime held in the year (422) → year floor
// (422) → regime with provisional ISR (422). Open to a read-only user and
// an archived taxpayer.
func (uc *FiledPayments) Get(ctx context.Context, userID string, req FiledRequest) (FiledView, error) {
	key, err := uc.resolve(req)
	if err != nil {
		return FiledView{}, err
	}
	var out FiledView
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		if _, err := uc.deps.Taxpayers.Taxpayer(ctx, key.TaxpayerID, LockNone); err != nil {
			return err
		}
		regime, err := uc.payingRegime(ctx, key)
		if err != nil {
			return err
		}
		out, err = uc.view(ctx, key, regime)
		return err
	})
	if err != nil {
		return FiledView{}, present("get filed payments", err)
	}
	return out, nil
}

// Put replaces every confirmed month of req with entries (an empty set
// clears them). Order: taxpayer id (404) → future year (400) → read-only
// (402) → taxpayer (404), archived (422) → regime, year floor, provisional
// ISR (422) → months outside the window (422) → month rules (400) →
// replace. The regime row stays locked until commit: two replaces of one
// regime serialize.
func (uc *FiledPayments) Put(ctx context.Context, userID string, req FiledRequest, entries []FiledEntry) (FiledView, error) {
	key, err := uc.resolve(req)
	if err != nil {
		return FiledView{}, err
	}
	var out FiledView
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
			return err
		}
		if _, err := ensureWritable(ctx, uc.deps.Taxpayers, key.TaxpayerID, false); err != nil {
			return err
		}
		if err := uc.deps.Taxpayers.LockRegime(ctx, key.TaxpayerID, key.Regime); err != nil {
			return err
		}
		regime, err := uc.payingRegime(ctx, key)
		if err != nil {
			return err
		}
		months, err := uc.checkEntries(regime, key.FiscalYear, entries)
		if err != nil {
			return err
		}
		if err := uc.deps.Store.ReplaceFiled(ctx, userID, key, months); err != nil {
			return err
		}
		out, err = uc.view(ctx, key, regime)
		return err
	})
	if err != nil {
		return FiledView{}, present("put filed payments", err)
	}
	uc.audit(ctx, eventFiledUpdated, slog.String("user_id", userID), slog.String("taxpayer_id", key.TaxpayerID),
		slog.String("regime_code", string(key.Regime)), slog.Int("fiscal_year", key.FiscalYear))
	return out, nil
}

func (uc *FiledPayments) resolve(req FiledRequest) (FiledKey, error) {
	tid, err := taxpayerID(req.TaxpayerID)
	if err != nil {
		return FiledKey{}, err
	}
	if err := uc.checkFutureYear(req.FiscalYear, FieldFiscalYearPath); err != nil {
		return FiledKey{}, err
	}
	return FiledKey{TaxpayerID: tid, Regime: fiscal.RegimeCode(req.RegimeCode), FiscalYear: req.FiscalYear}, nil
}

// payingRegime finds the regime of key held in its year, after which the
// year floor and the provisional-ISR rule apply.
func (uc *FiledPayments) payingRegime(ctx context.Context, key FiledKey) (HeldRegime, error) {
	regimes, err := uc.deps.Taxpayers.Regimes(ctx, key.TaxpayerID)
	if err != nil {
		return HeldRegime{}, err
	}
	for _, r := range regimes {
		if r.Code != key.Regime {
			continue
		}
		switch {
		case !r.ActiveIn(key.FiscalYear):
			return HeldRegime{}, regimeNotActiveError(FieldRegimeCode, msgRegimeNotActive)
		case key.FiscalYear < firstParamsYear:
			return HeldRegime{}, parametersUnavailableError(FieldFiscalYearPath)
		case !adomain.PaysProvisionalISR(r.Code, r.DefinitiveOption):
			return HeldRegime{}, paymentsNotApplicableError()
		}
		return r, nil
	}
	return HeldRegime{}, ErrRegimeNotHeld
}

// checkEntries refuses months outside the regime's window (422), then a
// repeated, unknown or not yet closed month or a negative amount (400),
// naming months.<index>[.<property>].
func (uc *FiledPayments) checkEntries(regime HeldRegime, fiscalYear int, entries []FiledEntry) ([]FiledMonth, error) {
	var window, rules []apperr.FieldError
	seen := map[int]bool{}
	months := make([]FiledMonth, 0, len(entries))
	for i, e := range entries {
		monthField := entryField(i, fieldMonthProperty)
		known := e.Month >= 1 && e.Month <= monthsPerYear
		switch {
		case known && !regime.Holds(fiscalYear, time.Month(e.Month)):
			window = append(window, apperr.FieldError{Field: monthField, Code: DetailRegimeNotActive, Message: msgRegimeNotActiveMonth})
		case !known || seen[e.Month]:
			rules = append(rules, apperr.FieldError{Field: monthField, Code: DetailInvalid, Message: msgMonthRepeated})
		case !uc.closed(fiscalYear, time.Month(e.Month)):
			rules = append(rules, apperr.FieldError{Field: monthField, Code: DetailOutOfRange, Message: msgMonthNotClosed})
		}
		if e.Amount < 0 {
			rules = append(rules, apperr.FieldError{Field: entryField(i, fieldAmountProperty), Code: DetailInvalid, Message: msgInvalidValue})
		}
		seen[e.Month] = true
		months = append(months, FiledMonth{Month: time.Month(e.Month), Amount: e.Amount})
	}
	if len(window) > 0 {
		return nil, apperr.Wrap(ErrRegimeNotActive, apperr.CodeBusinessRuleViolation, msgBusinessRule).WithFields(window...)
	}
	if len(rules) > 0 {
		return nil, fieldErrors(rules)
	}
	return months, nil
}

func entryField(index int, property string) string {
	return fieldMonths + fieldIndexSeparation + strconv.Itoa(index) + fieldIndexSeparation + property
}

// closed reports a month that has ended: every month of a past year, the
// months before the current one in the current year.
func (uc *FiledPayments) closed(fiscalYear int, month time.Month) bool {
	now := uc.now()
	return fiscalYear < now.Year() || (fiscalYear == now.Year() && month < now.Month())
}

// view reads the stored months of key into the 12 months of the year.
func (uc *FiledPayments) view(ctx context.Context, key FiledKey, regime HeldRegime) (FiledView, error) {
	stored, updatedAt, err := uc.deps.Store.FiledMonths(ctx, key)
	if err != nil {
		return FiledView{}, err
	}
	filed := make(map[time.Month]money.Cents, len(stored))
	for _, m := range stored {
		filed[m.Month] = m.Amount
	}
	out := FiledView{Regime: key.Regime, FiscalYear: key.FiscalYear, Months: make([]FiledMonthView, 0, monthsPerYear), UpdatedAt: updatedAt}
	for m := time.January; m <= time.December; m++ {
		period, err := domain.NewPeriod(key.FiscalYear, m)
		if err != nil {
			return FiledView{}, err
		}
		month := FiledMonthView{Month: m, Period: period, InWindow: regime.Holds(key.FiscalYear, m), Closed: uc.closed(key.FiscalYear, m)}
		if v, ok := filed[m]; ok {
			month.Filed = &v
		}
		out.Months = append(out.Months, month)
	}
	return out, nil
}
