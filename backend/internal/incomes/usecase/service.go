package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/platform/usertext"
)

// maxDescription is the longest description, in characters (contract
// RecordDescription.maxLength, the incomes_description_valid CHECK).
const maxDescription = 200

// Deps are the ports the incomes use cases need; cmd/api wires them.
type Deps struct {
	Tx        tx.Manager
	Store     Store
	Taxpayers Taxpayers
	Access    AccessPolicy
	Rates     RateCatalog
	// Withholdings are the dated parameters of the withholding decision
	// table (ADR-0008), the same parameter sets as Rates.
	Withholdings fiscal.WithholdingCatalog
	Clock        clock.Clock
	// NewID generates an income id when the client sends none; nil means
	// uuid v7 (data-model: time-ordered keys).
	NewID func() (string, error)
	// Logger receives the audit events (S-6); nil discards them.
	Logger *slog.Logger
}

// Incomes are the incomes use cases. They are safe for concurrent use.
type Incomes struct {
	deps Deps
}

// NewIncomes returns the use cases over deps.
func NewIncomes(deps Deps) *Incomes {
	if deps.NewID == nil {
		deps.NewID = newUUIDv7
	}
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.DiscardHandler)
	}
	return &Incomes{deps: deps}
}

func newUUIDv7() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("incomes: generate id: %w", err)
	}
	return id.String(), nil
}

// today is the current Mexico City day (Clock).
func (uc *Incomes) today() (fiscal.Date, error) {
	d, err := fiscal.ParseDate(uc.deps.Clock.Now().In(clock.Location()).Format(time.DateOnly))
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("incomes: today: %w", err)
	}
	return d, nil
}

// audit logs one audit event with the ids only (S-6).
func (uc *Incomes) audit(ctx context.Context, event, userID, taxpayerID, incomeID string) {
	uc.deps.Logger.LogAttrs(ctx, slog.LevelInfo, event, slog.String("user_id", userID),
		slog.String("taxpayer_id", taxpayerID), slog.String("income_id", incomeID))
}

// canonicalID returns the canonical (lower-case, hyphenated) form of a UUID,
// so "ABC…" and "abc…" name the same row.
func canonicalID(raw string) (string, bool) {
	id, err := uuid.Parse(raw)
	if err != nil || len(raw) != len(id.String()) {
		return "", false
	}
	return id.String(), true
}

// request is a normalized write before the defaults of its hidden
// questions are resolved (resolveDefaults): an empty Operation, and a nil
// viaPlatform, take theirs.
type request struct {
	fields      Fields
	viaPlatform *bool
}

// normalize turns an input into persisted fields: the CFDI UUID upper case,
// the description trimmed and NFC (blank is null). Its failures are
// validation violations.
func normalize(in IncomeInput) (request, error) {
	var vs fiscal.Violations
	f := Fields{
		Regime: in.Regime, Counterparty: in.Counterparty, Operation: in.Operation, ResidentialRent: in.ResidentialRent,
		Amount: in.Amount, IVARate: in.IVARate, IVA: in.IVA, Dates: in.Dates, PlatformActivity: in.PlatformActivity,
	}
	if in.CFDIUUID != "" {
		u, err := fiscal.ParseCFDIUUID(in.CFDIUUID)
		if err != nil {
			vs.Add(domain.FieldCFDIUUID, fiscal.ErrInvalid)
		}
		f.CFDIUUID = u
	}
	f.Description = norm.NFC.String(strings.TrimSpace(in.Description))
	switch {
	case utf8.RuneCountInString(f.Description) > maxDescription:
		vs.Add(fieldDescription, ErrTooLong)
	case usertext.HasInvisible(f.Description):
		vs.Add(fieldDescription, ErrInvisibleCharacters)
	}
	if len(vs) > 0 {
		return request{}, vs
	}
	return request{fields: f, viaPlatform: in.ViaPlatform}, nil
}

// stored is the request a stored income makes: every default resolved.
func stored(f Fields) request {
	via := f.ViaPlatform
	return request{fields: f, viaPlatform: &via}
}

// sameFields reports whether a retry carries the stored facts. A question
// the retry omits matches the stored answer: the create resolved its
// default (the 625 activity, the operation, via a platform).
func sameFields(stored Fields, retry request) bool {
	f := retry.fields
	if f.PlatformActivity == "" && stored.Regime == fiscal.RegimePlatforms {
		f.PlatformActivity = stored.PlatformActivity
	}
	if f.Operation == "" {
		f.Operation = stored.Operation
	}
	f.ViaPlatform = stored.ViaPlatform
	if retry.viaPlatform != nil {
		f.ViaPlatform = *retry.viaPlatform
	}
	return stored == f
}
