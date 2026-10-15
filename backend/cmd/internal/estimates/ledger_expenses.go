package estimates

import (
	"context"

	expensesinfra "github.com/romodevx/control-contable/backend/internal/expenses/infra"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// ExpenseSumsReader is the expenses module's aggregate reader
// (expensesinfra.LedgerReader).
type ExpenseSumsReader interface {
	ExpenseSums(ctx context.Context, q expensesinfra.LedgerQuery) ([]expensesinfra.ExpenseMonthSums, error)
}

// ExpenseLedger adapts the expenses aggregate reader to the tax engine's
// ExpenseLedgerReader port (M4-08). Modules never import each other
// (ADR-0001), so the composition root, the one place that sees both,
// translates the row types; M5-22 wires it into the estimate use case.
type ExpenseLedger struct{ reader ExpenseSumsReader }

var _ taxusecase.ExpenseLedgerReader = ExpenseLedger{}

// NewExpenseLedger adapts reader.
func NewExpenseLedger(reader ExpenseSumsReader) ExpenseLedger { return ExpenseLedger{reader: reader} }

// ExpenseSums implements taxusecase.ExpenseLedgerReader on the unit of work
// in ctx.
func (l ExpenseLedger) ExpenseSums(ctx context.Context, q taxusecase.LedgerQuery) ([]taxusecase.ExpenseMonthSums, error) {
	rows, err := l.reader.ExpenseSums(ctx, expensesinfra.LedgerQuery{
		TaxpayerID: q.TaxpayerID, FirstYear: q.FirstYear, FiscalYear: q.FiscalYear, Through: q.Through,
		PriorDecember: q.PriorDecember,
	})
	if err != nil {
		return nil, err
	}
	out := make([]taxusecase.ExpenseMonthSums, 0, len(rows))
	for _, r := range rows {
		out = append(out, taxusecase.ExpenseMonthSums{Regime: r.Regime, Year: r.Year, Month: r.Month, Sums: expenseSumsOf(r)})
	}
	return out, nil
}

// expenseSumsOf is one row's sums as the tax engine's domain holds them.
func expenseSumsOf(r expensesinfra.ExpenseMonthSums) domain.ExpenseSums {
	return domain.ExpenseSums{
		DeductiblePaid:    r.DeductiblePaid,
		NonDeductiblePaid: r.NonDeductiblePaid,
		CashOverLimitPaid: r.CashOverLimitPaid,
		IVAPaid: domain.IVAByRate{
			Rate16:     domain.RateSums{Base: r.IVAPaid.Rate16.Base, IVA: r.IVAPaid.Rate16.IVA},
			Rate8:      domain.RateSums{Base: r.IVAPaid.Rate8.Base, IVA: r.IVAPaid.Rate8.IVA},
			Rate0Base:  r.IVAPaid.Rate0Base,
			ExemptBase: r.IVAPaid.ExemptBase,
		},
		CreditableIVAPaid:     r.CreditableIVAPaid,
		CreditableIVAWithheld: r.CreditableIVAWithheld,
		PredialPaid:           r.PredialPaid,
		ISRWithheldToRemit:    r.ISRWithheldToRemit,
		IVAWithheldToRemit:    r.IVAWithheldToRemit,
	}
}
