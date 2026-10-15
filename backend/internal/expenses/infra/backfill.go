package infra

import (
	"context"
	"slices"
	"strings"

	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

var _ usecase.BackfillStore = Store{}

// underivedQuery selects the expenses the withholdings backfill (WH-08)
// re-derives: no complete stored derivation (written before WH-07) or
// another table version than the parameter. Deleted rows are included so a
// restore finds their supplier facts.
const underivedQuery = "SELECT " + columns + ` FROM expenses
WHERE withholding_isr_rate IS NULL OR withholding_table_version IS DISTINCT FROM ?
ORDER BY id`

// derivationColumnNames are the only columns the backfill writes: the
// supplier facts, the derivation and the creditable IVA split. They are
// cc_worker's UPDATE grant on expenses (migration 20261018100000).
var derivationColumnNames = []string{
	"creditable_iva_cents", "isr_withheld_cents", "iva_withheld_cents", "supplier_kind", "supplier_regime",
	"operation_kind", "residential_rent", "creditable_iva_withheld_cents", "withholding_rule_ids",
	"withholding_table_version", "withholding_cpa_unverified", "withholding_isr_rate", "withholding_iva_basis",
	"withholding_withholder",
}

// derivationStatement sets derivationColumnNames of one expense, built from
// the write allowlist so the placeholders and casts stay the same.
func derivationStatement() string {
	var sets []string
	for _, c := range editableColumns {
		if slices.Contains(derivationColumnNames, c.name) {
			sets = append(sets, c.name+" = "+c.placeholder)
		}
	}
	return "UPDATE expenses SET " + strings.Join(sets, ", ") + " WHERE id = ?"
}

// derivationArgs are derivationStatement's values for e, then its id.
func derivationArgs(e usecase.StoredExpense) []any {
	values := editable(e)
	var args []any
	for i, c := range editableColumns {
		if slices.Contains(derivationColumnNames, c.name) {
			args = append(args, values[i])
		}
	}
	return append(args, e.ID)
}

// Underived implements usecase.BackfillStore, ordered by id.
func (Store) Underived(ctx context.Context, version string, lock usecase.Lock) ([]usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	query := underivedQuery
	if lock == usecase.LockUpdate {
		query += " FOR NO KEY UPDATE"
	}
	var rows []expenseRow
	if err := db.Raw(query, version).Scan(&rows).Error; err != nil {
		return nil, translate("find underived expenses", err)
	}
	out := make([]usecase.StoredExpense, len(rows))
	for i, r := range rows {
		if out[i], err = r.toStored(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// SaveDerivation implements usecase.BackfillStore.
func (Store) SaveDerivation(ctx context.Context, e usecase.StoredExpense) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	result := db.Exec(derivationStatement(), derivationArgs(e)...)
	if result.Error != nil {
		return translate("save expense derivation", result.Error)
	}
	if result.RowsAffected == 0 {
		return usecase.ErrExpenseNotFound
	}
	return nil
}
