package taxinputs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// Store implements usecase.TaxInputStore, usecase.TaxInputReader and
// usecase.RegimeTextReader. It holds no pool: every method runs on the unit
// of work in ctx.
type Store struct{}

var (
	_ usecase.TaxInputStore    = Store{}
	_ usecase.TaxInputReader   = Store{}
	_ usecase.RegimeTextReader = Store{}
)

// NewStore returns a Store.
func NewStore() Store { return Store{} }

// Row locks (usecase.LockWrite). cc_app holds UPDATE on both tables, which
// a locking read requires.
const (
	lockTaxpayer = " FOR SHARE"
	lockRegime   = " FOR NO KEY UPDATE OF tr"
)

const taxpayerQuery = "SELECT archived_at IS NOT NULL AS archived FROM taxpayers WHERE id = ?"

// regimeQuery reads the held regime and the definitions of its catalog row
// (code, person type): 626 PF and 626 PM differ.
const regimeQuery = `SELECT tr.start_period, tr.end_period, r.required_inputs::text AS required_inputs
FROM taxpayer_regimes tr
JOIN regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
WHERE tr.taxpayer_id = ? AND tr.regime_code = ?`

// valueColumns are what a value read returns; value_rate as text keeps the
// stored scale ("0.1500").
const valueColumns = "regime_code, input_key, kind, month, value_cents, value_rate::text AS value_rate, value_bool"

// readerColumns are a reader row: the value columns and its fiscal year.
const readerColumns = valueColumns + ", fiscal_year"

// valuesQuery and readerQuery are served by the taxpayer_tax_inputs_key
// btree (user_id from RLS, taxpayer_id, regime_code, fiscal_year).
const (
	valuesQuery = "SELECT " + valueColumns + " FROM taxpayer_tax_inputs WHERE taxpayer_id = ? AND regime_code = ? AND fiscal_year = ?"
	readerQuery = "SELECT " + readerColumns + " FROM taxpayer_tax_inputs WHERE taxpayer_id = ? AND fiscal_year BETWEEN ? AND ?" +
		" ORDER BY fiscal_year, regime_code, input_key, month"
	deleteSQL = "DELETE FROM taxpayer_tax_inputs WHERE taxpayer_id = ? AND regime_code = ? AND fiscal_year = ?"
)

type taxpayerRow struct {
	Archived bool `gorm:"column:archived"`
}

type regimeRow struct {
	StartPeriod    string  `gorm:"column:start_period"`
	EndPeriod      *string `gorm:"column:end_period"`
	RequiredInputs string  `gorm:"column:required_inputs"`
}

// valueRow is one taxpayer_tax_inputs value as read.
type valueRow struct {
	RegimeCode string  `gorm:"column:regime_code"`
	InputKey   string  `gorm:"column:input_key"`
	Kind       string  `gorm:"column:kind"`
	Month      *int16  `gorm:"column:month"`
	ValueCents *int64  `gorm:"column:value_cents"`
	ValueRate  *string `gorm:"column:value_rate"`
	ValueBool  *bool   `gorm:"column:value_bool"`
	// FiscalYear is set by readerQuery only.
	FiscalYear int `gorm:"column:fiscal_year"`
}

// newValueRow is exactly the columns cc_app may INSERT: kind is derived by
// the taxpayer_tax_inputs_derive_kind trigger, never written.
type newValueRow struct {
	UserID     string  `gorm:"column:user_id"`
	TaxpayerID string  `gorm:"column:taxpayer_id"`
	RegimeCode string  `gorm:"column:regime_code"`
	FiscalYear int16   `gorm:"column:fiscal_year"`
	InputKey   string  `gorm:"column:input_key"`
	Month      *int16  `gorm:"column:month"`
	ValueCents *int64  `gorm:"column:value_cents"`
	ValueRate  *string `gorm:"column:value_rate"`
	ValueBool  *bool   `gorm:"column:value_bool"`
}

func (newValueRow) TableName() string { return "taxpayer_tax_inputs" }

// inputDefJSON is one regimes.required_inputs element (TaxInputDef).
type inputDefJSON struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	Required bool   `json:"required"`
}

// Taxpayer implements usecase.TaxInputStore.
func (Store) Taxpayer(ctx context.Context, id string, lock usecase.Lock) (usecase.TaxpayerState, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.TaxpayerState{}, err
	}
	query := taxpayerQuery
	if lock == usecase.LockWrite {
		query += lockTaxpayer
	}
	var rows []taxpayerRow
	if err := db.Raw(query, id).Scan(&rows).Error; err != nil {
		return usecase.TaxpayerState{}, fmt.Errorf("taxengine infra: read taxpayer: %w", err)
	}
	if len(rows) == 0 {
		return usecase.TaxpayerState{}, usecase.ErrTaxpayerNotFound
	}
	return usecase.TaxpayerState{Archived: rows[0].Archived}, nil
}

// Regime implements usecase.TaxInputStore.
func (Store) Regime(ctx context.Context, taxpayerID string, code fiscal.RegimeCode, lock usecase.Lock) (usecase.HeldRegime, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.HeldRegime{}, err
	}
	query := regimeQuery
	if lock == usecase.LockWrite {
		query += lockRegime
	}
	var rows []regimeRow
	if err := db.Raw(query, taxpayerID, string(code)).Scan(&rows).Error; err != nil {
		return usecase.HeldRegime{}, fmt.Errorf("taxengine infra: read regime: %w", err)
	}
	if len(rows) == 0 {
		return usecase.HeldRegime{}, usecase.ErrRegimeNotHeld
	}
	held, err := rows[0].toHeld()
	if err != nil {
		return usecase.HeldRegime{}, fmt.Errorf("taxengine infra: regime %s: %w", code, err)
	}
	return held, nil
}

func (r regimeRow) toHeld() (usecase.HeldRegime, error) {
	start, err := domain.ParsePeriod(r.StartPeriod)
	if err != nil {
		return usecase.HeldRegime{}, err
	}
	var end domain.Period
	if r.EndPeriod != nil {
		if end, err = domain.ParsePeriod(*r.EndPeriod); err != nil {
			return usecase.HeldRegime{}, err
		}
	}
	defs, err := decodeDefs(r.RequiredInputs)
	if err != nil {
		return usecase.HeldRegime{}, err
	}
	return usecase.HeldRegime{Start: start, End: end, Definitions: defs}, nil
}

// regimeTextsQuery reads the catalog row of every regime the taxpayer holds
// or held: the taxpayer_regimes primary key (user_id from RLS, taxpayer_id)
// and the regimes primary key, a handful of rows.
const regimeTextsQuery = `SELECT r.code, r.name, coalesce(r.explanation, '') AS explanation,
       r.required_inputs::text AS required_inputs
FROM taxpayer_regimes tr
JOIN regimes r ON r.code = tr.regime_code AND r.person_type = tr.person_type
WHERE tr.taxpayer_id = ?`

type regimeTextRow struct {
	Code           string `gorm:"column:code"`
	Name           string `gorm:"column:name"`
	Explanation    string `gorm:"column:explanation"`
	RequiredInputs string `gorm:"column:required_inputs"`
}

// RegimeTexts implements usecase.RegimeTextReader.
func (Store) RegimeTexts(ctx context.Context, taxpayerID string) (map[fiscal.RegimeCode]usecase.RegimeText, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []regimeTextRow
	if err := db.Raw(regimeTextsQuery, taxpayerID).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("taxengine infra: read regime texts: %w", err)
	}
	out := make(map[fiscal.RegimeCode]usecase.RegimeText, len(rows))
	for _, r := range rows {
		defs, err := decodeDefs(r.RequiredInputs)
		if err != nil {
			return nil, fmt.Errorf("taxengine infra: regime %s: %w", r.Code, err)
		}
		labels := make(map[domain.InputKey]string, len(defs))
		for _, d := range defs {
			labels[d.Key] = d.Label
		}
		out[fiscal.RegimeCode(r.Code)] = usecase.RegimeText{Name: r.Name, Explanation: r.Explanation, InputLabels: labels}
	}
	return out, nil
}

// decodeDefs parses regimes.required_inputs, keeping the catalog order.
func decodeDefs(raw string) ([]domain.InputDef, error) {
	var elements []inputDefJSON
	if err := json.Unmarshal([]byte(raw), &elements); err != nil {
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

// Values implements usecase.TaxInputStore.
func (Store) Values(ctx context.Context, key usecase.InputsKey) ([]usecase.StoredInput, error) {
	rows, err := readValues(ctx, valuesQuery, key.TaxpayerID, string(key.Regime), key.FiscalYear)
	if err != nil {
		return nil, err
	}
	out := make([]usecase.StoredInput, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.toStored())
	}
	return out, nil
}

func readValues(ctx context.Context, query string, args ...any) ([]valueRow, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []valueRow
	if err := db.Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("taxengine infra: read tax inputs: %w", err)
	}
	return rows, nil
}

func (r valueRow) toStored() usecase.StoredInput {
	s := usecase.StoredInput{Key: domain.InputKey(r.InputKey), Kind: domain.InputKind(r.Kind)}
	if r.Month != nil {
		s.Month = time.Month(*r.Month)
	}
	if r.ValueCents != nil {
		s.Cents = money.Cents(*r.ValueCents)
	}
	if r.ValueRate != nil {
		s.Rate = *r.ValueRate
	}
	if r.ValueBool != nil {
		s.Bool = *r.ValueBool
	}
	return s
}

// Replace implements usecase.TaxInputStore: one DELETE of the set, then one
// multi-row INSERT. The caller holds the regime row lock (LockWrite), so
// no concurrent replace of the same set can interleave.
func (Store) Replace(ctx context.Context, userID string, key usecase.InputsKey, values []usecase.StoredInput) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	if err := db.Exec(deleteSQL, key.TaxpayerID, string(key.Regime), key.FiscalYear).Error; err != nil {
		return fmt.Errorf("taxengine infra: clear tax inputs: %w", err)
	}
	if len(values) == 0 {
		return nil
	}
	rows := make([]newValueRow, 0, len(values))
	for _, v := range values {
		rows = append(rows, toNewRow(userID, key, v))
	}
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("taxengine infra: store tax inputs: %w", err)
	}
	return nil
}

func toNewRow(userID string, key usecase.InputsKey, v usecase.StoredInput) newValueRow {
	row := newValueRow{
		UserID: userID, TaxpayerID: key.TaxpayerID, RegimeCode: string(key.Regime),
		FiscalYear: int16(key.FiscalYear), // #nosec G115 -- contract FiscalYear is 2022..2100
		InputKey:   string(v.Key),
	}
	switch v.Kind {
	case domain.InputMoneyMonthly:
		month := int16(v.Month) // #nosec G115 -- a validated month, 1..12
		row.Month = &month
		fallthrough
	case domain.InputMoney:
		cents := int64(v.Cents)
		row.ValueCents = &cents
	case domain.InputRate:
		rate := v.Rate
		row.ValueRate = &rate
	case domain.InputBoolean:
		flag := v.Bool
		row.ValueBool = &flag
	}
	return row
}

// TaxInputs implements usecase.TaxInputReader: every regime's inputs of
// each of the taxpayer's fiscal years firstYear..lastYear in one query.
func (Store) TaxInputs(ctx context.Context, taxpayerID string, firstYear, lastYear int) (map[int]map[fiscal.RegimeCode]domain.TaxInputs, error) {
	rows, err := readValues(ctx, readerQuery, taxpayerID, firstYear, lastYear)
	if err != nil {
		return nil, err
	}
	type yearRegime struct {
		year int
		code fiscal.RegimeCode
	}
	grouped := map[yearRegime][]domain.TaxInputValue{}
	for _, r := range rows {
		v, err := r.toDomain()
		if err != nil {
			return nil, fmt.Errorf("taxengine infra: tax input %d/%s/%s: %w", r.FiscalYear, r.RegimeCode, r.InputKey, err)
		}
		key := yearRegime{year: r.FiscalYear, code: fiscal.RegimeCode(r.RegimeCode)}
		grouped[key] = append(grouped[key], v)
	}
	out := map[int]map[fiscal.RegimeCode]domain.TaxInputs{}
	for key, values := range grouped {
		in, err := domain.NewTaxInputs(values...)
		if err != nil {
			return nil, fmt.Errorf("taxengine infra: tax inputs of %d/%s: %w", key.year, key.code, err)
		}
		if out[key.year] == nil {
			out[key.year] = map[fiscal.RegimeCode]domain.TaxInputs{}
		}
		out[key.year][key.code] = in
	}
	return out, nil
}

func (r valueRow) toDomain() (domain.TaxInputValue, error) {
	s := r.toStored()
	v := domain.TaxInputValue{Key: s.Key, Kind: s.Kind, Month: s.Month, Money: s.Cents, Bool: s.Bool}
	if s.Kind == domain.InputRate {
		rate, err := money.ParseRate(s.Rate)
		if err != nil {
			return domain.TaxInputValue{}, err
		}
		v.Rate = rate
	}
	return v, nil
}
