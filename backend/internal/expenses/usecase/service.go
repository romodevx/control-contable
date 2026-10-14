package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
)

// maxDescriptionLength is the contract's RecordDescription maxLength, in
// characters.
const maxDescriptionLength = 200

// dayLayout formats the Clock's instant as a fiscal.Date.
const dayLayout = "2006-01-02"

// Deps are the ports the expense use cases need; cmd/api wires them.
type Deps struct {
	Tx         tx.Manager
	Access     AccessPolicy
	Regimes    TaxpayerRegimeChecker
	Taxpayers  TaxpayerProfiles
	Expenses   ExpenseStore
	Categories CategoryCatalog
	Parameters Parameters
	Clock      clock.Clock
	// NewID generates an expense id when the client sends none; nil means
	// uuid v7 (data-model: time-ordered keys).
	NewID func() (string, error)
	// Logger receives the audit events (S-6: ids only); nil discards them.
	Logger *slog.Logger
}

// Expenses is the expenses module's use cases (M4-05, M4-06): create,
// update, get, list, soft delete, restore and the categories query. It is
// safe for concurrent use.
type Expenses struct {
	deps Deps
}

// NewExpenses wires the use cases over deps.
func NewExpenses(deps Deps) *Expenses {
	if deps.NewID == nil {
		deps.NewID = newUUIDv7
	}
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &Expenses{deps: deps}
}

func newUUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("expenses: generate id: %w", err)
	}
	return id.String(), nil
}

// canonicalID parses a UUID in any accepted spelling and returns its
// canonical lower-case form, so ids compare as strings.
func canonicalID(raw string) (string, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", false
	}
	return id.String(), true
}

// today is the current Mexico City day (Clock).
func (u *Expenses) today() (fiscal.Date, error) {
	d, err := fiscal.ParseDate(u.deps.Clock.Now().Format(dayLayout))
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("expenses: clock day: %w", err)
	}
	return d, nil
}

// audit logs one audit event with the ids only (S-5, S-6: never an amount,
// a description or a folio fiscal).
func (u *Expenses) audit(ctx context.Context, event, userID, taxpayerID, expenseID string) {
	u.deps.Logger.LogAttrs(ctx, slog.LevelInfo, event,
		slog.String("user_id", userID), slog.String("taxpayer_id", taxpayerID), slog.String("expense_id", expenseID))
}

// normalizeDescription trims and NFC-normalizes free text; "" is no
// description. Longer than the contract allows is a 400.
func normalizeDescription(raw string) (string, error) {
	text := norm.NFC.String(strings.TrimSpace(raw))
	if utf8.RuneCountInString(text) > maxDescriptionLength {
		return "", violationsError(fiscal.Violations{{Field: fieldDescription, Err: errTooLong}})
	}
	return text, nil
}

// errTooLong is the too_long detail of a free-text field.
var errTooLong = fiscal.ValidationError(DetailTooLong)

// profile reads the path's taxpayer; a malformed id is the same 404 as an
// unknown one.
func (u *Expenses) profile(ctx context.Context, userID, rawTaxpayerID string) (string, TaxpayerProfile, error) {
	id, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return "", TaxpayerProfile{}, taxpayerNotFound()
	}
	p, err := u.deps.Taxpayers.Profile(ctx, userID, id)
	return id, p, err
}

// findExpense reads the path's expense of the taxpayer, live or deleted. A
// malformed id, another taxpayer's or another user's expense is a 404.
func (u *Expenses) findExpense(ctx context.Context, taxpayerID, rawExpenseID string, lock Lock) (StoredExpense, error) {
	id, ok := canonicalID(rawExpenseID)
	if !ok {
		return StoredExpense{}, notFound()
	}
	e, err := u.deps.Expenses.Find(ctx, id, lock)
	if errors.Is(err, ErrExpenseNotFound) || (err == nil && e.TaxpayerID != taxpayerID) {
		return StoredExpense{}, notFound()
	}
	return e, err
}

// view is the stored expense as the client sees it.
func view(e StoredExpense, warnings []domain.Warning) Expense {
	source := domain.SourceCategoryDefault
	if e.Overridden {
		source = domain.SourceUser
	}
	if warnings == nil {
		warnings = []domain.Warning{}
	}
	return Expense{
		ID: e.ID, TaxpayerID: e.TaxpayerID, Regime: e.Regime, Category: e.Category,
		Amount: e.Amount, IVARate: e.IVARate, IVA: e.IVA, PaymentMethod: e.PaymentMethod, PaidOn: e.PaidOn,
		HasCFDI: e.HasCFDI, CFDIUUID: e.CFDIUUID, Withheld: e.Withheld, Description: e.Description,
		Deductible: e.Deductible, Source: source, Warnings: warnings, Legend: domain.Legend,
		CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}
