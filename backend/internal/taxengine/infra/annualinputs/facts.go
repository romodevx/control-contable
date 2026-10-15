package annualinputs

import (
	"context"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// factsQuery reads one year through annual_inputs_pkey (user_id from RLS).
const factsQuery = `SELECT prior_year_accumulable_income_cents, other_exempt_income_cents, told_employer_will_file, updated_at
FROM annual_inputs WHERE taxpayer_id = ? AND fiscal_year = ?`

// upsertFactsSQL inserts or overwrites the row; an identical body changes
// nothing (updated_at stays: the same PUT twice leaves the same state).
const upsertFactsSQL = `INSERT INTO annual_inputs (user_id, taxpayer_id, fiscal_year, prior_year_accumulable_income_cents,
                           other_exempt_income_cents, told_employer_will_file)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT (user_id, taxpayer_id, fiscal_year) DO UPDATE SET
  prior_year_accumulable_income_cents = EXCLUDED.prior_year_accumulable_income_cents,
  other_exempt_income_cents = EXCLUDED.other_exempt_income_cents,
  told_employer_will_file = EXCLUDED.told_employer_will_file
WHERE (annual_inputs.prior_year_accumulable_income_cents, annual_inputs.other_exempt_income_cents,
       annual_inputs.told_employer_will_file)
  IS DISTINCT FROM (EXCLUDED.prior_year_accumulable_income_cents, EXCLUDED.other_exempt_income_cents,
                    EXCLUDED.told_employer_will_file)`

// taxInputMoneyQuery reads the yearly money values of some keys, through
// the taxpayer_tax_inputs key btree (user_id, taxpayer_id, regime_code,
// fiscal_year).
const taxInputMoneyQuery = `SELECT regime_code, input_key, value_cents FROM taxpayer_tax_inputs
WHERE taxpayer_id = ? AND fiscal_year = ? AND input_key IN ? AND month IS NULL AND value_cents IS NOT NULL`

type factsRow struct {
	PriorYearIncomeCents *int64    `gorm:"column:prior_year_accumulable_income_cents"`
	OtherExemptCents     int64     `gorm:"column:other_exempt_income_cents"`
	ToldEmployer         bool      `gorm:"column:told_employer_will_file"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

type taxInputMoneyRow struct {
	RegimeCode string `gorm:"column:regime_code"`
	InputKey   string `gorm:"column:input_key"`
	ValueCents int64  `gorm:"column:value_cents"`
}

// Facts implements annual.FactsStore.
func (Store) Facts(ctx context.Context, taxpayerID string, fiscalYear int) (annual.StoredFacts, bool, error) {
	var rows []factsRow
	if err := raw(ctx, "read annual inputs", &rows, factsQuery, taxpayerID, fiscalYear); err != nil {
		return annual.StoredFacts{}, false, err
	}
	if len(rows) == 0 {
		return annual.StoredFacts{}, false, nil
	}
	r := rows[0]
	f := annual.Facts{OtherExemptIncome: money.Cents(r.OtherExemptCents), ToldEmployerWillFile: r.ToldEmployer}
	if r.PriorYearIncomeCents != nil {
		prior := money.Cents(*r.PriorYearIncomeCents)
		f.PriorYearAccumulableIncome = &prior
	}
	return annual.StoredFacts{Facts: f, UpdatedAt: r.UpdatedAt}, true, nil
}

// UpsertFacts implements annual.FactsStore.
func (Store) UpsertFacts(ctx context.Context, userID, taxpayerID string, fiscalYear int, f annual.Facts) error {
	var prior *int64
	if f.PriorYearAccumulableIncome != nil {
		cents := int64(*f.PriorYearAccumulableIncome)
		prior = &cents
	}
	_, err := exec(ctx, "upsert annual inputs", upsertFactsSQL,
		userID, taxpayerID, fiscalYear, prior, int64(f.OtherExemptIncome), f.ToldEmployerWillFile)
	return err
}

// TaxInputMoney implements annual.FactsStore.
func (Store) TaxInputMoney(ctx context.Context, taxpayerID string, fiscalYear int, keys []domain.InputKey) (map[fiscal.RegimeCode]map[domain.InputKey]money.Cents, error) {
	out := map[fiscal.RegimeCode]map[domain.InputKey]money.Cents{}
	if len(keys) == 0 {
		return out, nil
	}
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = string(k)
	}
	var rows []taxInputMoneyRow
	if err := raw(ctx, "read tax inputs", &rows, taxInputMoneyQuery, taxpayerID, fiscalYear, names); err != nil {
		return nil, err
	}
	for _, r := range rows {
		code := fiscal.RegimeCode(r.RegimeCode)
		if out[code] == nil {
			out[code] = map[domain.InputKey]money.Cents{}
		}
		out[code][domain.InputKey(r.InputKey)] = money.Cents(r.ValueCents)
	}
	return out, nil
}
