package estimates

import (
	"context"
	"errors"

	incomesinfra "github.com/romodevx/control-contable/backend/internal/incomes/infra"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	taxusecase "github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// IncomeSumsReader is the incomes module's aggregate reader
// (incomesinfra.LedgerReader).
type IncomeSumsReader interface {
	IncomeSums(ctx context.Context, q incomesinfra.LedgerQuery) ([]incomesinfra.IncomeMonthSums, error)
}

// errIncomePriorDecember reports a PriorDecember income query: only the
// expenses carry January's withheld IVA credit (LIVA 5 fr. IV), so the
// incomes reader has no such range and a caller asking for it is a bug.
var errIncomePriorDecember = errors.New("taxengine ledger: the incomes reader has no previous-December range")

// IncomeLedger adapts the incomes aggregate reader to the tax engine's
// IncomeLedgerReader port (M3-07). Modules never import each other
// (ADR-0001), so the composition root, the one place that sees both,
// translates the row types; M5-22 wires it into the estimate use case.
type IncomeLedger struct{ reader IncomeSumsReader }

var _ taxusecase.IncomeLedgerReader = IncomeLedger{}

// NewIncomeLedger adapts reader.
func NewIncomeLedger(reader IncomeSumsReader) IncomeLedger { return IncomeLedger{reader: reader} }

// IncomeSums implements taxusecase.IncomeLedgerReader on the unit of work in
// ctx.
func (l IncomeLedger) IncomeSums(ctx context.Context, q taxusecase.LedgerQuery) ([]taxusecase.IncomeMonthSums, error) {
	if q.PriorDecember {
		return nil, errIncomePriorDecember
	}
	rows, err := l.reader.IncomeSums(ctx, incomesinfra.LedgerQuery{
		TaxpayerID: q.TaxpayerID, FirstYear: q.FirstYear, FiscalYear: q.FiscalYear, Through: q.Through,
	})
	if err != nil {
		return nil, err
	}
	out := make([]taxusecase.IncomeMonthSums, 0, len(rows))
	for _, r := range rows {
		out = append(out, taxusecase.IncomeMonthSums{Regime: r.Regime, Year: r.Year, Month: r.Month, Sums: incomeSumsOf(r)})
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
