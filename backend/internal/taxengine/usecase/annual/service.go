package annual

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

// firstParamsYear is the first fiscal year with tax parameters: an earlier
// year is 422 parameters_unavailable (MODULE-annual §5).
const firstParamsYear = 2026

// createAttempts bounds the units of work of one create: a primary-key
// collision is retried once, so a create that still meets a committed row of
// its id (the id lock is the primary guard) sees it and replays it.
const createAttempts = 2

// Optional is a nullable merge-patch field: Set false is absent (unchanged);
// Set with a nil Value is null (cleared).
type Optional[T any] struct {
	Set   bool
	Value *T
}

// Value is a merge-patch field set to v.
func Value[T any](v T) Optional[T] { return Optional[T]{Set: true, Value: &v} }

// Null is a merge-patch field set to null.
func Null[T any]() Optional[T] { return Optional[T]{Set: true} }

// apply returns the field after the patch: current when absent.
func (o Optional[T]) apply(current *T) *T {
	if !o.Set {
		return current
	}
	return o.Value
}

// WarningCode is a contract WarningCode the annual writes raise.
type WarningCode string

// The annual warnings (contract 0.11.0 additions).
const (
	// WarningCashNotDeductible: a cash payment of a cash-restricted
	// category, stored but never applied.
	WarningCashNotDeductible WarningCode = "cash_not_deductible"
	// WarningSalaryRegimeNotHeld: a constancia for a year the taxpayer does
	// not hold 605 in (allowed, OQ-A2; the screen suggests adding it).
	WarningSalaryRegimeNotHeld WarningCode = "salary_regime_not_held"
)

// Warning is a non-blocking notice of a write, on one field.
type Warning struct {
	Code  WarningCode
	Field string
}

// base holds what every annual use case needs.
type base struct {
	clock  clock.Clock
	logger *slog.Logger
}

func newBase(c clock.Clock, logger *slog.Logger) base {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return base{clock: c, logger: logger}
}

// now is the clock in Mexico City (the business day of every rule).
func (b base) now() time.Time { return b.clock.Now().In(clock.Location()) }

// today is the current business day.
func (b base) today() (fiscal.Date, error) {
	d, err := fiscal.ParseDate(b.now().Format(time.DateOnly))
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("annual: today: %w", err)
	}
	return d, nil
}

// checkFutureYear refuses a fiscal year after the current one (400).
func (b base) checkFutureYear(year int, field string) error {
	if year > b.now().Year() {
		return futureYearError(field)
	}
	return nil
}

// checkYearFloor refuses a fiscal year before the first with parameters
// (422).
func checkYearFloor(year int, field string) error {
	if year < firstParamsYear {
		return parametersUnavailableError(field)
	}
	return nil
}

// audit logs one audit event with ids only (§5: never an amount or label).
func (b base) audit(ctx context.Context, event string, attrs ...slog.Attr) {
	b.logger.LogAttrs(ctx, slog.LevelInfo, event, attrs...)
}

// canonicalID returns the canonical (lower-case, hyphenated) form of a
// UUID, so "ABC…" and "abc…" name one row.
func canonicalID(raw string) (string, bool) {
	id, err := uuid.Parse(raw)
	if err != nil || len(raw) != len(id.String()) {
		return "", false
	}
	return id.String(), true
}

// taxpayerID canonicalizes a path taxpayer id: a malformed one is 404.
func taxpayerID(raw string) (string, error) {
	id, ok := canonicalID(raw)
	if !ok {
		return "", notFound(ErrTaxpayerNotFound)
	}
	return id, nil
}

// pathIDs canonicalizes the taxpayer and row path ids: a malformed one is
// 404, never 400, like an unknown one.
func pathIDs(rawTaxpayerID, rawRowID string, rowNotFound error) (string, string, error) {
	tid, err := taxpayerID(rawTaxpayerID)
	if err != nil {
		return "", "", err
	}
	id, ok := canonicalID(rawRowID)
	if !ok {
		return "", "", notFound(rowNotFound)
	}
	return tid, id, nil
}

// clientID returns the canonical client id, or a new one when absent; a
// malformed one is 400 id invalid.
func clientID(raw string, newID func() (string, error)) (string, error) {
	if raw == "" {
		return newID()
	}
	id, ok := canonicalID(raw)
	if !ok {
		return "", fieldErrors([]apperr.FieldError{{Field: FieldID, Code: DetailInvalid, Message: msgInvalidValue}})
	}
	return id, nil
}

func newUUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("annual: new id: %w", err)
	}
	return id.String(), nil
}

// normalizeLabel composes a free-text label to NFC (ADR-0004); it never
// trims: an untrimmed label is refused, not repaired (contract).
func normalizeLabel(raw string) string {
	if !utf8.ValidString(raw) {
		return raw
	}
	return norm.NFC.String(raw)
}

// labelError is the FieldError code and message of a label that is not
// 1 to maxRunes characters of valid, trimmed UTF-8 (invalid) or that holds
// an invisible character (invisible_characters); both empty for a valid
// label.
func labelError(label string, maxRunes int) (code, message string) {
	length := utf8.RuneCountInString(label)
	switch {
	case !utf8.ValidString(label) || length == 0 || length > maxRunes || label != strings.TrimSpace(label):
		return DetailInvalid, msgInvalidValue
	case usertext.HasInvisible(label):
		return DetailInvisibleCharacters, msgInvisibleCharacters
	}
	return "", ""
}

// ensureWritable checks the taxpayer of a write under lock: 404 when it is
// not the caller's, 422 when archived, and 422 for a persona moral when
// fisicaOnly.
func ensureWritable(ctx context.Context, store TaxpayerStore, taxpayerID string, fisicaOnly bool) (Taxpayer, error) {
	t, err := store.Taxpayer(ctx, taxpayerID, LockWrite)
	if err != nil {
		return Taxpayer{}, err
	}
	if t.Archived {
		return Taxpayer{}, archivedError()
	}
	if fisicaOnly && t.PersonType != fiscal.PersonFisica {
		return Taxpayer{}, personTypeError()
	}
	return t, nil
}
