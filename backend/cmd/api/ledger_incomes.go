package main

import (
	"context"

	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// incomeSumsReader is the incomes module's aggregate reader
// (incomesinfra.LedgerReader).
type incomeSumsReader interface {
	IncomeSums(ctx context.Context, q incomesinfra.LedgerQuery) ([]incomesinfra.IncomeMonthSums, error)
}

// incomeLedger adapts the incomes aggregate reader to the tax engine's
// IncomeLedgerReader port (M3-07). Modules never import each other
// (ADR-0001), so the composition root, the one place that sees both,
// translates the row types; M5-22 wires it into the estimate use case.
type incomeLedger struct{ reader incomeSumsReader }

var _ taxusecase.IncomeLedgerReader = incomeLedger{}

// newIncomeLedger adapts reader.
func newIncomeLedger(reader incomeSumsReader) incomeLedger { return incomeLedger{reader: reader} }

// IncomeSums implements taxusecase.IncomeLedgerReader on the unit of work in
// ctx.
func (l incomeLedger) IncomeSums(ctx context.Context, q taxusecase.LedgerQuery) ([]taxusecase.IncomeMonthSums, error) {
	rows, err := l.reader.IncomeSums(ctx, incomesinfra.LedgerQuery{
		TaxpayerID: q.TaxpayerID, FiscalYear: q.FiscalYear, Through: q.Through,
	})
	if err != nil {
		return nil, err
	}
	out := make([]taxusecase.IncomeMonthSums, 0, len(rows))
	for _, r := range rows {
		out = append(out, taxusecase.IncomeMonthSums{Regime: r.Regime, Month: r.Month, Sums: incomeSumsOf(r)})
	}
	return out, nil
}

// incomeSumsOf is one row's sums as the tax engine's domain holds them.
func incomeSumsOf(r incomesinfra.IncomeMonthSums) domain.IncomeSums {
	return domain.IncomeSums{
		Cash:    r.Cash,
		Accrual: r.Accrual,
		ByRate: domain.IVAByRate{
			Rate16:     domain.RateSums{Base: r.ByRate.Rate16.Base, IVA: r.ByRate.Rate16.IVA},
			Rate8:      domain.RateSums{Base: r.ByRate.Rate8.Base, IVA: r.ByRate.Rate8.IVA},
			Rate0Base:  r.ByRate.Rate0Base,
			ExemptBase: r.ByRate.ExemptBase,
		},
		IVAWithheld: domain.Withheld{ByMoral: r.IVAWithheld.ByMoral, ByPlatform: r.IVAWithheld.ByPlatform},
		ISRWithheld: domain.Withheld{ByMoral: r.ISRWithheld.ByMoral, ByPlatform: r.ISRWithheld.ByPlatform},
	}
}
