package infra

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
)

// cashAnyPayment is category_regime_rules.cash_payment_rule for a category
// never deductible when paid in cash (fuel, LISR 27 fr. III ⚠️ CPA).
const cashAnyPayment = "any_cash"

// CategoryCatalog implements usecase.CategoryCatalog over the global
// expense_categories and category_regime_rules tables (SELECT only).
type CategoryCatalog struct{}

var _ usecase.CategoryCatalog = CategoryCatalog{}

// NewCategoryCatalog returns a CategoryCatalog. It holds no pool.
func NewCategoryCatalog() CategoryCatalog { return CategoryCatalog{} }

// ruleRow is one category joined with its rule for a regime.
type ruleRow struct {
	Code              string  `gorm:"column:code"`
	Name              string  `gorm:"column:name"`
	Deductibility     string  `gorm:"column:deductibility"`
	DefaultDeductible bool    `gorm:"column:default_deductible"`
	IVACreditable     bool    `gorm:"column:iva_creditable_default"`
	CashPaymentRule   string  `gorm:"column:cash_payment_rule"`
	Note              *string `gorm:"column:note"`
	LegalBasis        string  `gorm:"column:legal_basis"`
	DefaultOperation  *string `gorm:"column:default_operation_kind"`
}

func (r ruleRow) toEntry() (usecase.CategoryEntry, error) {
	deductibility, err := domain.ParseDeductibility(r.Deductibility)
	if err != nil {
		return usecase.CategoryEntry{}, fmt.Errorf("expenses infra: rule of %q: deductibility %q: %w", r.Code, r.Deductibility, err)
	}
	entry := usecase.CategoryEntry{
		Rule: domain.CategoryRule{
			Category: domain.CategoryCode(r.Code), Deductibility: deductibility, DefaultDeductible: r.DefaultDeductible,
			IVACreditable: r.IVACreditable, ElectronicPaymentOnly: r.CashPaymentRule == cashAnyPayment,
			DefaultOperation: fiscal.OperationKind(deref(r.DefaultOperation)),
		},
		Name: r.Name, LegalBasis: r.LegalBasis,
	}
	if r.Note != nil {
		entry.Note = *r.Note
	}
	return entry, nil
}

// rules is the regime's rules joined with their categories, in catalog
// order, over the (regime_code, person_type, category_code) primary key.
func rules(db *gorm.DB, regime fiscal.RegimeCode, person fiscal.PersonType) *gorm.DB {
	return db.Table("category_regime_rules AS r").
		Select("c.code, c.name, r.deductibility, r.default_deductible, r.iva_creditable_default, r.cash_payment_rule, r.note, r.legal_basis, "+
			"r.default_operation_kind").
		Joins("JOIN expense_categories AS c ON c.code = r.category_code").
		Where("r.regime_code = ? AND r.person_type = ?", string(regime), string(person)).
		Order("c.sort_order")
}

// Rules implements usecase.CategoryCatalog.
func (CategoryCatalog) Rules(ctx context.Context, regime fiscal.RegimeCode, person fiscal.PersonType) ([]usecase.CategoryEntry, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return nil, err
	}
	var rows []ruleRow
	if err := rules(db, regime, person).Scan(&rows).Error; err != nil {
		return nil, translate("read category rules", err)
	}
	out := make([]usecase.CategoryEntry, 0, len(rows))
	for _, r := range rows {
		entry, err := r.toEntry()
		if err != nil {
			return nil, err
		}
		out = append(out, entry)
	}
	return out, nil
}

// Rule implements usecase.CategoryCatalog.
func (CategoryCatalog) Rule(ctx context.Context, regime fiscal.RegimeCode, person fiscal.PersonType, category domain.CategoryCode) (usecase.CategoryEntry, error) {
	db, err := pgdb.Conn(ctx)
	if err != nil {
		return usecase.CategoryEntry{}, err
	}
	var rows []ruleRow
	if err := rules(db, regime, person).Where("r.category_code = ?", string(category)).Scan(&rows).Error; err != nil {
		return usecase.CategoryEntry{}, translate("read category rule", err)
	}
	if len(rows) == 0 {
		return usecase.CategoryEntry{}, usecase.ErrCategoryNotFound
	}
	return rows[0].toEntry()
}
