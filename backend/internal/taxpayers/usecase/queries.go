package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// TaxpayerRegimeChecker is the write-side question M3 (incomes) and M4
// (expenses) ask before storing a record: "does taxpayer T, owned by the
// caller, hold regime R in period P, and is T not archived?"
//
// It answers nil, or an *apperr.Error the consumer returns as is:
// 404 not_found (no such taxpayer for the caller), 422
// business_rule_violation with details taxpayer_archived or
// regime_not_active (field regime_code: R not held in P), 400
// validation_failed (field period: P is not YYYY-MM).
//
// Call it from inside the record's own unit of work (the same ctx): it then
// joins that transaction and its FOR SHARE locks on the taxpayer and the
// regime hold until the record commits, so a concurrent archive or
// end-regime waits and then sees the record (§2.6: no record outside the
// window).
//
// Consumers declare their own copy of this interface (archtest: modules
// never import each other); cmd/api wires *Queries into it.
type TaxpayerRegimeChecker interface {
	CheckRegime(ctx context.Context, userID, taxpayerID, regimeCode, period string) error
}

// TaxpayerProfileReader is what M5 (tax engine) reads about a taxpayer:
// person type, archived state and every regime with its options and period
// window. Archived taxpayers are readable (past estimates stay available).
// Errors are *apperr.Error (404 not_found). Consumers declare their own
// copy; cmd/api adapts the result to the consumer's types.
type TaxpayerProfileReader interface {
	Profile(ctx context.Context, userID, taxpayerID string) (TaxpayerProfile, error)
}

// TaxpayerProfile is a taxpayer as the tax engine sees it.
type TaxpayerProfile struct {
	TaxpayerID string
	PersonType domain.PersonType
	Archived   bool
	// Regimes are ordered by start period, then code.
	Regimes []RegimeProfile
}

// RegimeProfile is one regime of a TaxpayerProfile.
type RegimeProfile struct {
	Code      string
	Treatment string
	// Options carries every option, defaults included.
	Options map[string]domain.OptionValue
	Window  domain.Window
}

// Queries implements TaxpayerRegimeChecker and TaxpayerProfileReader.
type Queries struct {
	service
}

// NewQueries returns the read-side use cases over deps (Tx, Taxpayers and
// Regimes are used).
func NewQueries(deps Deps) *Queries { return &Queries{service: newService(deps)} }

var (
	_ TaxpayerRegimeChecker = (*Queries)(nil)
	_ TaxpayerProfileReader = (*Queries)(nil)
)

// CheckRegime implements TaxpayerRegimeChecker. It runs in a read-write unit
// of work because FOR SHARE is refused in a READ ONLY transaction; nested
// in the caller's it simply joins it.
func (q *Queries) CheckRegime(ctx context.Context, userID, rawTaxpayerID, regimeCode, rawPeriod string) error {
	id, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return notFound()
	}
	p, err := domain.ParsePeriod(rawPeriod)
	if err != nil {
		fields := &fieldErrors{}
		fields.addPeriod(fieldPeriod, DetailInvalid, err)
		return fields.err()
	}
	err = q.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		return q.checkRegime(ctx, id, regimeCode, p)
	})
	return present("check regime", err)
}

func (q *Queries) checkRegime(ctx context.Context, id, code string, p domain.Period) error {
	t, err := q.deps.Taxpayers.Find(ctx, id, LockShare)
	if err != nil {
		return err
	}
	if t.Archived() {
		return archivedError()
	}
	held, err := q.deps.Regimes.FindRegime(ctx, id, code, LockShare)
	if err == nil && held.Window.Contains(p) {
		return nil
	}
	if err != nil && err != ErrRegimeNotHeld { //nolint:errorlint // stores return the sentinel itself
		return err
	}
	return businessRule(ErrRegimeNotActive, msgNotActive, fieldRegimeCode, DetailRegimeNotActive, msgNotActive)
}

// Profile implements TaxpayerProfileReader.
func (q *Queries) Profile(ctx context.Context, userID, rawTaxpayerID string) (TaxpayerProfile, error) {
	id, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return TaxpayerProfile{}, notFound()
	}
	var out TaxpayerProfile
	err := q.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		t, err := q.deps.Taxpayers.Find(ctx, id, LockNone)
		if err != nil {
			return err
		}
		held, err := q.deps.Regimes.Regimes(ctx, id)
		if err != nil {
			return err
		}
		out = TaxpayerProfile{TaxpayerID: t.ID, PersonType: t.PersonType, Archived: t.Archived()}
		for _, h := range held {
			out.Regimes = append(out.Regimes, RegimeProfile{Code: h.Code, Treatment: h.Treatment, Options: h.Options, Window: h.Window})
		}
		return nil
	})
	if err != nil {
		return TaxpayerProfile{}, present("read taxpayer profile", err)
	}
	return out, nil
}
