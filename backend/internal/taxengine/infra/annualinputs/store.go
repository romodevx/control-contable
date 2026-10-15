// Package annualinputs is the GORM persistence of the annual estimate's
// inputs (ADR-0009, ANN-03 tables): personal_deductions, salary_statements,
// annual_inputs and annual_filed_payments, plus the taxpayer and regime
// reads the annual use cases need. It holds no pool: every method runs on
// the unit of work in ctx, as cc_app under RLS, so another user's rows do
// not exist for it.
package annualinputs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// Store implements every annual store and read port.
type Store struct{}

var (
	_ annual.TaxpayerStore       = Store{}
	_ annual.DeductionStore      = Store{}
	_ annual.SalaryStore         = Store{}
	_ annual.FactsStore          = Store{}
	_ annual.FiledStore          = Store{}
	_ annual.YearInputsReader    = Store{}
	_ annual.FiledPaymentsReader = Store{}
)

// NewStore returns a Store.
func NewStore() Store { return Store{} }

// The first keys of the transaction-scoped advisory locks (two-int form):
// they keep the annual locks apart from the incomes ("INC1") and expenses
// ("EXP1") id locks and from the worker's single-key leader lock.
const (
	listLockClass      int32 = 0x414E4E31 // "ANN1": a bounded list of one taxpayer and year
	deductionLockClass int32 = 0x414E4E32 // "ANN2": a personal-deduction client id
	statementLockClass int32 = 0x414E4E33 // "ANN3": a salary-constancia client id
)

const advisoryLockSQL = "SELECT pg_advisory_xact_lock(?, hashtext(?))"

const sqlstateUniqueViolation = "23505"

// Row locks (annual.LockWrite). cc_app holds UPDATE on every table locked,
// which a locking read requires.
const (
	lockShare     = " FOR SHARE"
	lockNoKeyUpdt = " FOR NO KEY UPDATE"
)

const taxpayerQuery = "SELECT person_type, archived_at IS NOT NULL AS archived FROM taxpayers WHERE id = ?"

// regimesQuery reads every regime the taxpayer holds or held with the
// definitions of its catalog row (code, person type: 626 PF and PM
// differ). The 625 definitive option defaults to false (catalog default).
const regimesQuery = `SELECT tr.regime_code, tr.start_period, tr.end_period,
       COALESCE((tr.options ->> 'definitive_option')::boolean, false) AS definitive_option,
       r.required_inputs::text AS required_inputs
FROM taxpayer_regimes tr
JOIN regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
WHERE tr.taxpayer_id = ?
ORDER BY tr.regime_code`

const lockRegimeQuery = "SELECT regime_code FROM taxpayer_regimes WHERE taxpayer_id = ? AND regime_code = ?" + lockNoKeyUpdt

type taxpayerRow struct {
	PersonType string `gorm:"column:person_type"`
	Archived   bool   `gorm:"column:archived"`
}

type regimeRow struct {
	RegimeCode       string  `gorm:"column:regime_code"`
	StartPeriod      string  `gorm:"column:start_period"`
	EndPeriod        *string `gorm:"column:end_period"`
	DefinitiveOption bool    `gorm:"column:definitive_option"`
	RequiredInputs   string  `gorm:"column:required_inputs"`
}

// inputDefJSON is one regimes.required_inputs element (TaxInputDef).
type inputDefJSON struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
}

// Taxpayer implements annual.TaxpayerStore.
func (Store) Taxpayer(ctx context.Context, id string, lock annual.Lock) (annual.Taxpayer, error) {
	query := taxpayerQuery
	if lock == annual.LockWrite {
		query += lockShare
	}
	var rows []taxpayerRow
	if err := raw(ctx, "read taxpayer", &rows, query, id); err != nil {
		return annual.Taxpayer{}, err
	}
	if len(rows) == 0 {
		return annual.Taxpayer{}, annual.ErrTaxpayerNotFound
	}
	return annual.Taxpayer{PersonType: fiscal.PersonType(rows[0].PersonType), Archived: rows[0].Archived}, nil
}

// Regimes implements annual.TaxpayerStore.
func (Store) Regimes(ctx context.Context, taxpayerID string) ([]annual.HeldRegime, error) {
	var rows []regimeRow
	if err := raw(ctx, "read regimes", &rows, regimesQuery, taxpayerID); err != nil {
		return nil, err
	}
	out := make([]annual.HeldRegime, 0, len(rows))
	for _, r := range rows {
		held, err := r.toHeld()
		if err != nil {
			return nil, fmt.Errorf("annual infra: regime %s: %w", r.RegimeCode, err)
		}
		out = append(out, held)
	}
	return out, nil
}

func (r regimeRow) toHeld() (annual.HeldRegime, error) {
	start, err := domain.ParsePeriod(r.StartPeriod)
	if err != nil {
		return annual.HeldRegime{}, err
	}
	var end domain.Period
	if r.EndPeriod != nil {
		if end, err = domain.ParsePeriod(*r.EndPeriod); err != nil {
			return annual.HeldRegime{}, err
		}
	}
	defs, err := decodeDefs(r.RequiredInputs)
	if err != nil {
		return annual.HeldRegime{}, err
	}
	return annual.HeldRegime{
		Code: fiscal.RegimeCode(r.RegimeCode), Start: start, End: end, DefinitiveOption: r.DefinitiveOption, Definitions: defs,
	}, nil
}

// decodeDefs parses regimes.required_inputs, keeping the catalog order.
func decodeDefs(rawJSON string) ([]domain.InputDef, error) {
	var elements []inputDefJSON
	if err := json.Unmarshal([]byte(rawJSON), &elements); err != nil {
		return nil, fmt.Errorf("required_inputs: %w", err)
	}
	defs := make([]domain.InputDef, 0, len(elements))
	for _, e := range elements {
		kind := domain.InputKind(e.Kind)
		if !kind.Valid() {
			return nil, fmt.Errorf("required_inputs %s kind %q: %w", e.Key, e.Kind, domain.ErrInvalidTaxInput)
		}
		defs = append(defs, domain.InputDef{Key: domain.InputKey(e.Key), Label: e.Label, Kind: kind, Required: e.Required})
	}
	return defs, nil
}

// LockRegime implements annual.TaxpayerStore.
func (Store) LockRegime(ctx context.Context, taxpayerID string, code fiscal.RegimeCode) error {
	var codes []string
	if err := raw(ctx, "lock regime", &codes, lockRegimeQuery, taxpayerID, string(code)); err != nil {
		return err
	}
	if len(codes) == 0 {
		return annual.ErrRegimeNotHeld
	}
	return nil
}

// LockList implements annual.TaxpayerStore: an advisory lock on (taxpayer,
// list, year), released at commit or rollback. A hash collision only
// serializes two unrelated writers.
func (Store) LockList(ctx context.Context, taxpayerID string, list annual.List, fiscalYear int) error {
	key := taxpayerID + "/" + strconv.Itoa(int(list)) + "/" + strconv.Itoa(fiscalYear)
	return advisoryLock(ctx, "lock list", listLockClass, key)
}

func advisoryLock(ctx context.Context, op string, class int32, key string) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Exec(advisoryLockSQL, class, key).Error; err != nil {
		return fmt.Errorf("annual infra: %s: %w", op, err)
	}
	return nil
}

// raw scans query into dest on the unit of work in ctx.
func raw(ctx context.Context, op string, dest any, query string, args ...any) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Raw(query, args...).Scan(dest).Error; err != nil {
		return fmt.Errorf("annual infra: %s: %w", op, err)
	}
	return nil
}

// exec runs a statement on the unit of work in ctx and returns the rows it
// affected.
func exec(ctx context.Context, op, statement string, args ...any) (int64, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return 0, err
	}
	result := db.Exec(statement, args...)
	if result.Error != nil {
		return 0, fmt.Errorf("annual infra: %s: %w", op, result.Error)
	}
	return result.RowsAffected, nil
}

// insert creates row; a primary-key collision is annual.ErrIDTaken (the
// key sees across RLS: another user's id).
func insert(ctx context.Context, op, primaryKey string, row any) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Create(row).Error; err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == sqlstateUniqueViolation && pgErr.ConstraintName == primaryKey {
			return fmt.Errorf("%w (annual infra: %s: %w)", annual.ErrIDTaken, op, err)
		}
		return fmt.Errorf("annual infra: %s: %w", op, err)
	}
	return nil
}

func wrap(op string, err error) error { return fmt.Errorf("annual infra: %s: %w", op, err) }

// dateOf parses a date column read as text ("YYYY-MM-DD").
func dateOf(rawDate string) (fiscal.Date, error) {
	d, err := fiscal.ParseDate(rawDate)
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("annual infra: stored date %q: %w", rawDate, err)
	}
	return d, nil
}

// latest is the most recent of times, nil for none.
func latest(times []time.Time) *time.Time {
	var out *time.Time
	for i := range times {
		if out == nil || times[i].After(*out) {
			out = &times[i]
		}
	}
	return out
}

func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// nullable is nil for the zero value: a NULL column.
func nullable[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
