package usecase

import (
	"context"
	"errors"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// previewPayment stands in for the payment method a preview does not know:
// any non-cash method, so the cash rule never fires (it changes the
// deduction, never the withholding).
const previewPayment = domain.PaymentTransfer

// withholdingWarnings are the warning codes a preview reports
// (contract WithholdingPreview): the withholding ones only.
var withholdingWarnings = map[domain.WarningCode]bool{
	domain.WarningSupplierRegimeUnknown:       true,
	domain.WarningDeductionRequiresRemittance: true,
	domain.WarningCPAUnverified:               true,
	domain.WarningForeignSupplier:             true,
}

// Preview derives what a persona moral must withhold from a supplier for
// the facts the form holds, and stores nothing (previewExpenseWithholdings,
// ADR-0008 §4). It runs the create's checks on the fields it takes: the
// day ("" = today) within [fiscal.MinRecordDate, today] and covered by a
// parameter set (400 out_of_range, before the regime window), the regime
// held that day (422 regime_not_active), a known category (400
// unknown_category), the IVA and supplier rules. A read: allowed for a read-only user and an archived
// taxpayer, in a read-only unit of work.
func (u *Expenses) Preview(ctx context.Context, in PreviewInput) (Preview, error) {
	var out Preview
	err := u.deps.Tx.ReadOnly(ctx, in.UserID, func(ctx context.Context) error {
		_, profile, err := u.profile(ctx, in.UserID, in.TaxpayerID)
		if err != nil {
			return err
		}
		on, err := u.categoriesDate(CategoriesInput{Regime: in.Regime, On: in.On})
		if err != nil {
			return err
		}
		if _, err := u.deps.Parameters.On(on); errors.Is(err, ErrParametersUnavailable) {
			return invalidField(fieldOn, fiscal.ErrOutOfRange)
		} else if err != nil {
			return err
		}
		hold, ok := profile.holding(in.Regime, on.Period())
		if !ok {
			return regimeNotActiveError()
		}
		out, err = u.preview(ctx, in, profile.payer(hold), on)
		return err
	})
	if err != nil {
		return Preview{}, present("preview expense withholdings", err)
	}
	return out, nil
}

func (u *Expenses) preview(ctx context.Context, in PreviewInput, payer domain.Payer, on fiscal.Date) (Preview, error) {
	entry, err := u.deps.Categories.Rule(ctx, payer.Regime, payer.PersonType, in.Category)
	if errors.Is(err, ErrCategoryNotFound) {
		return Preview{}, invalidField(domain.FieldCategoryCode, domain.ErrUnknownCategory)
	}
	if err != nil {
		return Preview{}, err
	}
	today, err := u.today()
	if err != nil {
		return Preview{}, err
	}
	draft := domain.Draft{
		Category: in.Category, Amount: in.Amount, IVARate: in.IVARate, IVA: in.IVA, PaymentMethod: previewPayment,
		PaidOn: on, HasCFDI: true, Supplier: in.Supplier,
	}
	a, err := domain.Assess(draft, payer, entry.Rule, today, u.deps.Parameters, u.deps.Withholdings)
	if err != nil {
		return Preview{}, assessError(err)
	}
	warnings := []domain.Warning{}
	for _, w := range a.Warnings {
		if withholdingWarnings[w.Code] {
			warnings = append(warnings, w)
		}
	}
	return Preview{
		Withholding: a.Withholding, RemitDueOn: a.RemitDueOn, Subtotal: in.Amount, IVA: in.IVA, Warnings: warnings,
	}, nil
}
