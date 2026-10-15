package infra

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// expenseIDLockClass is the first key of the client-id advisory locks
// (two-int form, objsubid 2): it keeps them apart from the worker's
// single-key leader lock and from any other module's id locks.
const expenseIDLockClass int32 = 0x45585031 // "EXP1"

// dateLayout is a paid_on as the stores pass and read it.
const dateLayout = "2006-01-02"

// Store implements usecase.ExpenseStore with GORM on the unit of work in
// ctx (pgdb.Conn), as cc_app: RLS shows only the context user's rows.
type Store struct{}

var _ usecase.ExpenseStore = Store{}

// NewStore returns a Store. It holds no pool.
func NewStore() Store { return Store{} }

// expenseRow is an expenses row as the stores read it.
type expenseRow struct {
	ID                   string     `gorm:"column:id"`
	TaxpayerID           string     `gorm:"column:taxpayer_id"`
	RegimeCode           string     `gorm:"column:regime_code"`
	CategoryCode         string     `gorm:"column:category_code"`
	Description          *string    `gorm:"column:description"`
	AmountCents          int64      `gorm:"column:amount_cents"`
	IVARate              string     `gorm:"column:iva_rate"`
	IVACents             int64      `gorm:"column:iva_cents"`
	PaymentMethod        string     `gorm:"column:payment_method"`
	HasCFDI              bool       `gorm:"column:has_cfdi"`
	CFDIUUID             *string    `gorm:"column:cfdi_uuid"`
	Deductible           bool       `gorm:"column:deductible"`
	DeductibleOverridden bool       `gorm:"column:deductible_overridden"`
	CashOverLimit        bool       `gorm:"column:cash_over_limit"`
	CreditableIVACents   int64      `gorm:"column:creditable_iva_cents"`
	ISRWithheldCents     int64      `gorm:"column:isr_withheld_cents"`
	IVAWithheldCents     int64      `gorm:"column:iva_withheld_cents"`
	PaidOn               time.Time  `gorm:"column:paid_on"`
	SupplierKind         *string    `gorm:"column:supplier_kind"`
	SupplierRegime       *string    `gorm:"column:supplier_regime"`
	OperationKind        *string    `gorm:"column:operation_kind"`
	ResidentialRent      bool       `gorm:"column:residential_rent"`
	CreditableWithheld   int64      `gorm:"column:creditable_iva_withheld_cents"`
	RuleIDs              string     `gorm:"column:withholding_rule_ids"`
	TableVersion         *string    `gorm:"column:withholding_table_version"`
	CPAUnverified        bool       `gorm:"column:withholding_cpa_unverified"`
	ISRRate              *string    `gorm:"column:withholding_isr_rate"`
	IVABasis             *string    `gorm:"column:withholding_iva_basis"`
	Withholder           *string    `gorm:"column:withholding_withholder"`
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (expenseRow) TableName() string { return "expenses" }

// columns lists every column of an expenseRow. The rule ids travel as one
// comma-separated text and the ISR rate as its decimal text.
const columns = `id, taxpayer_id, regime_code, category_code, description, amount_cents, iva_rate, iva_cents,
	payment_method, has_cfdi, cfdi_uuid, deductible, deductible_overridden, cash_over_limit, creditable_iva_cents,
	isr_withheld_cents, iva_withheld_cents, paid_on, supplier_kind, supplier_regime, operation_kind,
	residential_rent, creditable_iva_withheld_cents,
	array_to_string(withholding_rule_ids, ',') AS withholding_rule_ids, withholding_table_version,
	withholding_cpa_unverified, withholding_isr_rate::text AS withholding_isr_rate, withholding_iva_basis,
	withholding_withholder, deleted_at, created_at, updated_at`

// returning hands back the written row.
const returning = " RETURNING " + columns

// ruleIDSeparator joins the rule ids to and from Postgres.
const ruleIDSeparator = ","

// toStored maps a row to the use case's type. The CHECKs guarantee the
// enums, dates and folio formats; a value that still fails to parse is a
// server fault, never a client error.
func (r expenseRow) toStored() (usecase.StoredExpense, error) {
	paidOn, err := fiscal.ParseDate(r.PaidOn.Format(dateLayout))
	if err != nil {
		return usecase.StoredExpense{}, fmt.Errorf("expenses infra: paid_on of %s: %w", r.ID, err)
	}
	var folio fiscal.CFDIUUID
	if r.CFDIUUID != nil {
		if folio, err = fiscal.ParseCFDIUUID(*r.CFDIUUID); err != nil {
			return usecase.StoredExpense{}, fmt.Errorf("expenses infra: cfdi_uuid of %s: %w", r.ID, err)
		}
	}
	withholding, err := r.withholding()
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	return usecase.StoredExpense{
		ID: r.ID, TaxpayerID: r.TaxpayerID, Regime: fiscal.RegimeCode(r.RegimeCode),
		Category: domain.CategoryCode(r.CategoryCode), Amount: money.Cents(r.AmountCents),
		IVARate: fiscal.IVARate(r.IVARate), IVA: money.Cents(r.IVACents),
		PaymentMethod: domain.PaymentMethod(r.PaymentMethod), PaidOn: paidOn, HasCFDI: r.HasCFDI, CFDIUUID: folio,
		Deductible: r.Deductible, Overridden: r.DeductibleOverridden,
		CashOverLimit: r.CashOverLimit, CreditableIVA: money.Cents(r.CreditableIVACents),
		CreditableIVAWithheld: money.Cents(r.CreditableWithheld),
		Supplier: domain.Supplier{
			Kind: domain.SupplierKind(deref(r.SupplierKind)), Regime: domain.SupplierRegime(deref(r.SupplierRegime)),
			Operation: fiscal.OperationKind(deref(r.OperationKind)), ResidentialRent: r.ResidentialRent,
		},
		Withholding: withholding, Description: deref(r.Description),
		Deleted: r.DeletedAt != nil, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}, nil
}

// withholding reads the stored derivation. The typed amounts of a record
// written before WH-07 have no table version: they come back underived.
func (r expenseRow) withholding() (fiscal.WithholdingRecord, error) {
	w := fiscal.WithholdingRecord{
		Withholdings: fiscal.Withholdings{ISR: money.Cents(r.ISRWithheldCents), IVA: money.Cents(r.IVAWithheldCents)},
		TableVersion: deref(r.TableVersion), CPAUnverified: r.CPAUnverified,
		IVABasis: fiscal.IVABasis(deref(r.IVABasis)), Withholder: fiscal.Withholder(deref(r.Withholder)),
	}
	for id := range strings.SplitSeq(r.RuleIDs, ruleIDSeparator) {
		if id != "" {
			w.RuleIDs = append(w.RuleIDs, fiscal.RuleID(id))
		}
	}
	if r.ISRRate != nil {
		rate, err := money.ParseRate(*r.ISRRate)
		if err != nil {
			return fiscal.WithholdingRecord{}, fmt.Errorf("expenses infra: withholding_isr_rate of %s: %w", r.ID, err)
		}
		w.ISRRate = rate
	}
	return w, nil
}

// ruleIDsArray is the rule ids as a Postgres text[] literal ('{}' when
// none): every id matches the contract pattern, so none needs quoting.
func ruleIDsArray(ids []fiscal.RuleID) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = string(id)
	}
	return "{" + strings.Join(parts, ruleIDSeparator) + "}"
}

// derivedDetail is the ISR rate, IVA basis and withholder columns of a
// derivation, all NULL on an underived record (the detail CHECKs).
func derivedDetail(w fiscal.WithholdingRecord) (rate, basis, withholder *string) {
	if !w.Derived() {
		return nil, nil, nil
	}
	r := w.ISRRate.String()
	return &r, nullable(string(w.IVABasis)), nullable(string(w.Withholder))
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nullable is "" as NULL.
func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// folioOf is a folio fiscal as its column value (NULL when absent).
func folioOf(u fiscal.CFDIUUID) *string { return nullable(u.String()) }

// LockID implements usecase.ExpenseStore: a transaction-scoped advisory
// lock on the id, released when the unit of work ends.
func (Store) LockID(ctx context.Context, id string) error {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return err
	}
	return translate("lock expense id", db.Exec("SELECT pg_advisory_xact_lock(?, hashtext(?))", expenseIDLockClass, id).Error)
}

// Find implements usecase.ExpenseStore.
func (Store) Find(ctx context.Context, id string, lock usecase.Lock) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	query := "SELECT " + columns + " FROM expenses WHERE id = ?"
	if lock == usecase.LockUpdate {
		query += " FOR NO KEY UPDATE"
	}
	var rows []expenseRow
	if err := db.Raw(query, id).Scan(&rows).Error; err != nil {
		return usecase.StoredExpense{}, translate("read expense", err)
	}
	if len(rows) == 0 {
		return usecase.StoredExpense{}, usecase.ErrExpenseNotFound
	}
	return rows[0].toStored()
}

// editableColumn is one column the use cases write and its placeholder.
type editableColumn struct{ name, placeholder string }

// editableColumns are exactly the columns the use cases write, in the order
// editable returns their values. It is the allowlist every INSERT and
// UPDATE is built from: no request field reaches a column not named here
// (PR #136 note 1); cc_app's column grants are the backstop.
var editableColumns = []editableColumn{
	{"regime_code", "?"}, {"category_code", "?"}, {"description", "?"}, {"amount_cents", "?"},
	{"iva_rate", "?"}, {"iva_cents", "?"}, {"payment_method", "?"}, {"has_cfdi", "?"}, {"cfdi_uuid", "?"},
	{"deductible", "?"}, {"deductible_overridden", "?"}, {"cash_over_limit", "?"}, {"creditable_iva_cents", "?"},
	{"isr_withheld_cents", "?"}, {"iva_withheld_cents", "?"}, {"paid_on", "?::date"},
	{"supplier_kind", "?"}, {"supplier_regime", "?"}, {"operation_kind", "?"}, {"residential_rent", "?"},
	{"creditable_iva_withheld_cents", "?"}, {"withholding_rule_ids", "?::text[]"},
	{"withholding_table_version", "?"}, {"withholding_cpa_unverified", "?"},
	{"withholding_isr_rate", "?::numeric"}, {"withholding_iva_basis", "?"}, {"withholding_withholder", "?"},
}

// insertStatement names id, user_id, taxpayer_id and the editable columns.
func insertStatement() string {
	names := []string{"id", "user_id", "taxpayer_id"}
	values := []string{"?", "?", "?"}
	for _, c := range editableColumns {
		names = append(names, c.name)
		values = append(values, c.placeholder)
	}
	return "INSERT INTO expenses (" + strings.Join(names, ", ") + ") VALUES (" + strings.Join(values, ", ") + ")" + returning
}

// updateStatement sets every editable column of one expense.
func updateStatement() string {
	sets := make([]string, 0, len(editableColumns))
	for _, c := range editableColumns {
		sets = append(sets, c.name+" = "+c.placeholder)
	}
	return "UPDATE expenses SET " + strings.Join(sets, ", ") + " WHERE id = ?" + returning
}

// Insert implements usecase.ExpenseStore.
func (Store) Insert(ctx context.Context, userID string, e usecase.StoredExpense) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	args := []any{e.ID, userID, e.TaxpayerID}
	return write(db.Raw(insertStatement(), append(args, editable(e)...)...), "insert expense")
}

// Update implements usecase.ExpenseStore. It writes every editable column.
func (Store) Update(ctx context.Context, e usecase.StoredExpense) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	return write(db.Raw(updateStatement(), append(editable(e), e.ID)...), "update expense")
}

// editable is the values of editableColumns for e, in order.
func editable(e usecase.StoredExpense) []any {
	w := e.Withholding
	rate, basis, withholder := derivedDetail(w)
	return []any{
		string(e.Regime), string(e.Category), nullable(e.Description), int64(e.Amount), string(e.IVARate),
		int64(e.IVA), string(e.PaymentMethod), e.HasCFDI, folioOf(e.CFDIUUID), e.Deductible, e.Overridden,
		e.CashOverLimit, int64(e.CreditableIVA), int64(w.ISR), int64(w.IVA), e.PaidOn.String(),
		nullable(string(e.Supplier.Kind)), nullable(string(e.Supplier.Regime)), nullable(string(e.Supplier.Operation)),
		e.Supplier.ResidentialRent, int64(e.CreditableIVAWithheld), ruleIDsArray(w.RuleIDs),
		nullable(w.TableVersion), w.CPAUnverified, rate, basis, withholder,
	}
}

// SetDeleted implements usecase.ExpenseStore: deleted_at is the
// database's now() (soft delete) or NULL (restore).
func (Store) SetDeleted(ctx context.Context, id string, deleted bool) (usecase.StoredExpense, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.StoredExpense{}, err
	}
	statement := "UPDATE expenses SET deleted_at = CASE WHEN ? THEN now() END WHERE id = ?" + returning
	return write(db.Raw(statement, deleted, id), "set expense deleted")
}
