package incomeshttp

import (
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// withholders word the kernel's withholder on the income side: the
// receptor of the CFDI is the counterparty that pays the taxpayer.
var withholders = map[fiscal.Withholder]gen.Withholder{
	fiscal.WithholderNone:     gen.WithholderNone,
	fiscal.WithholderReceptor: gen.WithholderCounterparty,
	fiscal.WithholderPlatform: gen.WithholderPlatform,
}

// explained are the facts a withholding's lines and net need.
type explained struct {
	amount, iva money.Cents
	receptor    fiscal.ReceptorKind
	activity    domain.PlatformActivity
}

// toWithholding is the contract Withholding of a stored or previewed
// derivation (ADR-0008); nil for an income stored before WH-06 (the WH-08
// backfill derives it). The explanations come from the kernel.
func toWithholding(w fiscal.WithholdingRecord, e explained) (*gen.Withholding, error) {
	if !w.Derived() {
		return nil, nil
	}
	net, err := w.Net(e.amount, e.iva)
	if err != nil {
		return nil, err
	}
	lines := fiscal.Explain(fiscal.ExplainInput{
		Side: fiscal.SideIncome, Receptor: e.receptor, RuleIDs: w.RuleIDs, ISRRate: w.ISRRate, ISR: w.ISR, IVA: w.IVA,
		IVABasis: w.IVABasis, Withholder: w.Withholder, PlatformActivity: fiscal.PlatformActivity(e.activity),
	})
	out := &gen.Withholding{
		IsrCents: int64(w.ISR), IvaCents: int64(w.IVA), IsrRate: w.ISRRate.String(), IvaBasis: gen.WithholdingBasis(w.IVABasis),
		Withholder: withholders[w.Withholder], NetCents: int64(net), RemitDueOn: nil,
		RuleIds: make([]gen.WithholdingRuleId, 0, len(w.RuleIDs)), Explanations: make([]gen.WithholdingExplanation, 0, len(lines)),
		TableVersion: w.TableVersion, CpaUnverified: w.CPAUnverified,
	}
	for _, id := range w.RuleIDs {
		out.RuleIds = append(out.RuleIds, string(id))
	}
	for _, l := range lines {
		out.Explanations = append(out.Explanations, gen.WithholdingExplanation{RuleId: string(l.RuleID), Text: l.Text})
	}
	return out, nil
}

// optionalKind is nil (absent or JSON null) for a record without the fact.
func optionalKind(k fiscal.ReceptorKind) *gen.CounterpartyKind {
	if k == "" {
		return nil
	}
	out := gen.CounterpartyKind(k)
	return &out
}

func optionalOperation(o fiscal.OperationKind) *gen.OperationKind {
	if o == "" {
		return nil
	}
	out := gen.OperationKind(o)
	return &out
}
