package expenseshttp

import (
	"fmt"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// errNotDerived reports a preview without a derivation, or a derivation
// naming an unknown withholder: a server fault, never a client error.
var errNotDerived = fmt.Errorf("expenses http: without a derivation: %w", fiscal.ErrInvalidParameter)

// storedWithholding is a stored expense's Withholding, nil (JSON null) for
// an expense stored before WH-07 that the WH-08 backfill or an edit has not
// derived yet (0.12.0, ADR-0008 Amendment 2): served, never a 500.
func storedWithholding(w fiscal.WithholdingRecord, subtotal, iva money.Cents, due fiscal.Date) (*gen.Withholding, error) {
	if !w.Derived() {
		return nil, nil
	}
	out, err := toWithholding(w, subtotal, iva, due)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// supplierOf maps the contract's optional supplier facts.
func supplierOf(kind *gen.SupplierKind, regime *gen.SupplierRegime, operation *gen.OperationKind, residential *bool) domain.Supplier {
	var s domain.Supplier
	if kind != nil {
		s.Kind = domain.SupplierKind(*kind)
	}
	if regime != nil {
		s.Regime = domain.SupplierRegime(*regime)
	}
	if operation != nil {
		s.Operation = fiscal.OperationKind(*operation)
	}
	if residential != nil {
		s.ResidentialRent = *residential
	}
	return s
}

// withholders are the kernel's withholders as an expense words them.
var withholders = map[fiscal.Withholder]gen.Withholder{
	fiscal.WithholderReceptor: gen.WithholderYou,
	fiscal.WithholderPlatform: gen.WithholderPlatform,
	fiscal.WithholderNone:     gen.WithholderNone,
}

// toWithholding is the contract's Withholding of a stored or previewed
// derivation (ADR-0008); errNotDerived without one (0.12.0 requires it).
func toWithholding(w fiscal.WithholdingRecord, subtotal, iva money.Cents, due fiscal.Date) (gen.Withholding, error) {
	withholder, known := withholders[w.Withholder]
	if !w.Derived() || !known {
		return gen.Withholding{}, errNotDerived
	}
	net, err := w.Net(subtotal, iva)
	if err != nil {
		return gen.Withholding{}, err
	}
	remit, err := contractDate(due)
	if err != nil {
		return gen.Withholding{}, err
	}
	ids := make([]gen.WithholdingRuleId, 0, len(w.RuleIDs))
	for _, id := range w.RuleIDs {
		ids = append(ids, string(id))
	}
	lines := fiscal.Explain(fiscal.ExplainInput{
		Side: fiscal.SideExpense, RuleIDs: w.RuleIDs, ISRRate: w.ISRRate, ISR: w.ISR, IVA: w.IVA,
		IVABasis: w.IVABasis, Withholder: w.Withholder,
	})
	explanations := make([]gen.WithholdingExplanation, 0, len(lines))
	for _, l := range lines {
		explanations = append(explanations, gen.WithholdingExplanation{RuleId: string(l.RuleID), Text: l.Text})
	}
	return gen.Withholding{
		IsrCents: int64(w.ISR), IvaCents: int64(w.IVA), IsrRate: w.ISRRate.String(),
		IvaBasis: gen.WithholdingBasis(w.IVABasis), Withholder: withholder, NetCents: int64(net),
		RemitDueOn: remit, RuleIds: ids, Explanations: explanations, TableVersion: w.TableVersion,
		CpaUnverified: w.CPAUnverified,
	}, nil
}

// contractDate is nil (JSON null) for the zero Date.
func contractDate(d fiscal.Date) (*openapi_types.Date, error) {
	if d.IsZero() {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, d.String())
	if err != nil {
		return nil, fmt.Errorf("expenses http: date %s: %w", d, err)
	}
	return &openapi_types.Date{Time: t}, nil
}

// supplierFields are the contract's supplier facts of a stored expense,
// nil (null) where absent.
func supplierFields(s domain.Supplier) (*gen.SupplierKind, *gen.SupplierRegime, *gen.OperationKind) {
	var kind *gen.SupplierKind
	if s.Kind != "" {
		k := gen.SupplierKind(s.Kind)
		kind = &k
	}
	var regime *gen.SupplierRegime
	if s.Regime != "" {
		r := gen.SupplierRegime(s.Regime)
		regime = &r
	}
	var operation *gen.OperationKind
	if s.Operation != "" {
		o := gen.OperationKind(s.Operation)
		operation = &o
	}
	return kind, regime, operation
}

// toPreviewInput decodes previewExpenseWithholdings' query.
func toPreviewInput(userID string, req gen.PreviewExpenseWithholdingsRequestObject) usecase.PreviewInput {
	p := req.Params
	in := usecase.PreviewInput{
		UserID: userID, TaxpayerID: req.TaxpayerId.String(), Regime: regimeOf(p.Regime),
		Category: domain.CategoryCode(p.CategoryCode), Amount: money.Cents(p.AmountCents),
		IVARate: fiscal.IVARate(p.IvaRate), IVA: money.Cents(p.IvaCents),
		Supplier: supplierOf(p.SupplierKind, p.SupplierRegime, p.OperationKind, p.ResidentialRent),
	}
	if p.On != nil {
		in.On = p.On.String()
	}
	return in
}

// toPreview is the contract's WithholdingPreview.
func toPreview(p usecase.Preview) (gen.WithholdingPreview, error) {
	w, err := toWithholding(p.Withholding, p.Subtotal, p.IVA, p.RemitDueOn)
	if err != nil {
		return gen.WithholdingPreview{}, fmt.Errorf("expenses http: preview: %w", err)
	}
	return gen.WithholdingPreview{Withholding: w, Warnings: toWarnings(p.Warnings)}, nil
}
