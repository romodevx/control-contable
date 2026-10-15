package infra

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

var _ usecase.BackfillStore = Store{}

// underivedCondition selects the incomes the withholdings backfill (WH-08)
// re-derives: no facts (written before WH-06), no complete stored
// derivation, or another table version than the parameter. Deleted rows
// are included so a restore finds their facts.
const underivedCondition = "counterparty_kind IS NULL OR operation_kind IS NULL OR withholding_isr_rate IS NULL " +
	"OR withholding_table_version IS DISTINCT FROM ?"

// derivationColumns are the only columns the backfill writes: the facts
// and the stored derivation. They are cc_worker's UPDATE grant on incomes
// (migration 20261018100000).
var derivationColumns = []string{
	"counterparty_type", "counterparty_kind", "operation_kind", "residential_rent", "via_platform",
	"isr_withheld_cents", "iva_withheld_cents", "withholding_rule_ids", "withholding_table_version",
	"withholding_cpa_unverified", "withholding_isr_rate", "withholding_iva_basis", "withholding_withholder",
}

// Underived implements usecase.BackfillStore, ordered by id.
func (Store) Underived(ctx context.Context, version string, lock usecase.Lock) ([]usecase.StoredIncome, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	query := db.Table(table).Select(incomeColumns).Where(underivedCondition, version).Order("id")
	if lock == usecase.LockUpdate {
		query = query.Clauses(clause.Locking{Strength: "NO KEY UPDATE"})
	}
	var rows []incomeRow
	if err := query.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("incomes infra: find underived incomes: %w", err)
	}
	out := make([]usecase.StoredIncome, len(rows))
	for i, r := range rows {
		if out[i], err = r.toStored(); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// SaveDerivation implements usecase.BackfillStore: columnsOf restricted to
// derivationColumns.
func (Store) SaveDerivation(ctx context.Context, id string, f usecase.Fields, w fiscal.WithholdingRecord) error {
	all := columnsOf(f, w)
	columns := make(map[string]any, len(derivationColumns))
	for _, name := range derivationColumns {
		columns[name] = all[name]
	}
	return update(ctx, "save derivation", id, columns)
}
