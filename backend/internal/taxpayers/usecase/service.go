package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// startFloor is the earliest start a regime may be given: January of the
// first fiscal year with tax parameters (taxengine/params/data/2026). A
// month before it could hold no record (no rate covers it) nor an estimate.
// The database CHECK taxpayer_regimes_start_period_valid stays at the
// catalog's 2022-01, so rows stored before this floor remain valid.
var startFloor = mustPeriod("2026-01")

// Deps are the ports the taxpayers use cases need. Each use case type uses
// the subset it needs; cmd/api wires one Deps for all of them.
type Deps struct {
	Tx        tx.Manager
	Access    AccessPolicy
	Taxpayers TaxpayerStore
	Regimes   RegimeStore
	Catalog   RegimeCatalog
	Directory RegimeDirectory
	Records   RegimeRecords
	RFC       RFCProtector
	Clock     clock.Clock
	// NewID generates a taxpayer id; nil means uuid v7 (data-model: time
	// ordered keys).
	NewID func() (string, error)
	// Logger receives the audit events (S-9); nil discards them.
	Logger *slog.Logger
}

// service is what the use case types share.
type service struct {
	deps Deps
}

func newService(deps Deps) service {
	if deps.NewID == nil {
		deps.NewID = newUUIDv7
	}
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return service{deps: deps}
}

func newUUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("taxpayers: generate id: %w", err)
	}
	return id.String(), nil
}

// currentPeriod is the current month in Mexico City (Clock).
func (s service) currentPeriod() domain.Period { return domain.PeriodOf(s.deps.Clock.Now()) }

// now is the current instant (Clock).
func (s service) now() time.Time { return s.deps.Clock.Now() }

// audit logs one audit event with the ids only (S-9: never the RFC).
func (s service) audit(ctx context.Context, event, userID, taxpayerID string, attrs ...slog.Attr) {
	all := append([]slog.Attr{slog.String("user_id", userID), slog.String("taxpayer_id", taxpayerID)}, attrs...)
	s.deps.Logger.LogAttrs(ctx, slog.LevelInfo, event, all...)
}

// load reads one taxpayer with its regimes.
func (s service) load(ctx context.Context, id string) (Taxpayer, error) {
	t, err := s.deps.Taxpayers.Find(ctx, id, LockNone)
	if err != nil {
		return Taxpayer{}, err
	}
	held, err := s.deps.Regimes.Regimes(ctx, id)
	if err != nil {
		return Taxpayer{}, err
	}
	return toTaxpayer(t, held), nil
}

// findWritable locks taxpayer id for a write and refuses an archived one.
func (s service) findWritable(ctx context.Context, id string) (StoredTaxpayer, error) {
	t, err := s.deps.Taxpayers.Find(ctx, id, LockUpdate)
	if err != nil {
		return StoredTaxpayer{}, err
	}
	if t.Archived() {
		return StoredTaxpayer{}, archivedError()
	}
	return t, nil
}

// startPeriod resolves a regime start: "" is the current month; otherwise
// "YYYY-MM" within [startFloor, current month]. On failure it returns the
// detail code and the cause.
func (s service) startPeriod(raw string) (domain.Period, string, error) {
	current := s.currentPeriod()
	if raw == "" {
		return current, "", nil
	}
	start, err := domain.ParsePeriod(raw)
	if err != nil {
		return domain.Period{}, DetailInvalid, err
	}
	if start.Before(startFloor) || start.After(current) {
		return domain.Period{}, DetailOutOfRange, ErrPeriodOutOfRange
	}
	return start, "", nil
}

// canonicalID returns the canonical (lower-case, hyphenated) form of a UUID,
// so "ABC…" and "abc…" name the same row.
func canonicalID(raw string) (string, bool) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return "", false
	}
	return id.String(), true
}

// rfcSeparators are the characters a typed RFC may carry between its parts
// (contract RfcInput: inner spaces and "-" are removed).
var rfcSeparators = strings.NewReplacer(" ", "", "-", "")

// parseRFC normalizes and validates a typed RFC for the declared person
// type (domain.NewRFCFor).
func parseRFC(personType domain.PersonType, raw string) (domain.RFC, error) {
	return domain.NewRFCFor(personType, rfcSeparators.Replace(raw))
}

// rfcWarnings returns the check-digit warning when the digit of a written
// RFC did not match (D-RFC); none for the zero RFC (nothing written).
func rfcWarnings(rfc domain.RFC) []Warning {
	if rfc.IsZero() || rfc.CheckDigitValid() {
		return nil
	}
	return []Warning{{Code: WarningRFCCheckDigitMismatch, Field: fieldRFC, Message: msgRFCCheckDigitMismatch}}
}

func toTaxpayer(t StoredTaxpayer, held []HeldRegime) Taxpayer {
	regimes := make([]TaxpayerRegime, 0, len(held))
	for _, h := range held {
		regimes = append(regimes, toRegime(h))
	}
	return Taxpayer{
		ID: t.ID, PersonType: t.PersonType, DisplayName: t.DisplayName, RFCMasked: t.RFCMasked,
		Regimes: regimes, ArchivedAt: t.ArchivedAt, CreatedAt: t.CreatedAt,
	}
}

func toRegime(h HeldRegime) TaxpayerRegime {
	return TaxpayerRegime{
		Code: h.Code, Name: h.Name, Treatment: h.Treatment, IncomeBasis: h.IncomeBasis,
		Options: h.Options, Window: h.Window,
	}
}

func mustPeriod(raw string) domain.Period {
	p, err := domain.ParsePeriod(raw)
	if err != nil {
		panic(err)
	}
	return p
}
