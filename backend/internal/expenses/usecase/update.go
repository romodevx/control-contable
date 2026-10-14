package usecase

import (
	"context"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// Update edits a live expense with merge-patch semantics (M4-05,
// MODULE-expenses §2.4): every rule of Create is re-run on the merged
// expense, moving it to another regime or date re-checks the window, and a
// default deductible is recomputed while an override is kept (UseDefault
// clears it). A read-only user gets 402; an edit reserves no slot. Last
// write wins: the row is locked FOR NO KEY UPDATE while it is judged.
func (u *Expenses) Update(ctx context.Context, in UpdateInput) (Expense, error) {
	var description *string
	if in.Patch.Description != nil {
		text, err := normalizeDescription(*in.Patch.Description)
		if err != nil {
			return Expense{}, err
		}
		description = &text
	}
	var out Expense
	err := u.deps.Tx.ReadWrite(ctx, in.UserID, func(ctx context.Context) error {
		var err error
		out, err = u.update(ctx, in, description)
		return err
	})
	if err != nil {
		return Expense{}, present("update expense", err)
	}
	u.audit(ctx, "expense.updated", in.UserID, out.TaxpayerID, out.ID)
	return out, nil
}

func (u *Expenses) update(ctx context.Context, in UpdateInput, description *string) (Expense, error) {
	taxpayerID, profile, err := u.profile(ctx, in.UserID, in.TaxpayerID)
	if err != nil {
		return Expense{}, err
	}
	current, err := u.findExpense(ctx, taxpayerID, in.ExpenseID, LockUpdate)
	if err != nil {
		return Expense{}, err
	}
	if current.Deleted {
		return Expense{}, notFound()
	}
	if err := u.deps.Access.EnsureWritable(ctx, in.UserID); err != nil {
		return Expense{}, err
	}
	if profile.Archived {
		return Expense{}, archivedError()
	}
	regime, draft := merge(current, in.Patch)
	if description != nil {
		current.Description = *description
	}
	v, err := u.judge(ctx, subject{
		userID: in.UserID, taxpayerID: taxpayerID, expenseID: current.ID, profile: profile, regime: regime, draft: draft,
	})
	if err != nil {
		return Expense{}, err
	}
	saved, err := u.deps.Expenses.Update(ctx, stored(current, regime, draft, current.Description, v.assessment))
	if err != nil {
		return Expense{}, err
	}
	return view(saved, v.warnings), nil
}

// merge applies the patch to the stored expense's regime and draft.
func merge(e StoredExpense, p Patch) (regime fiscal.RegimeCode, d domain.Draft) {
	regime, d = e.Regime, draftOf(e)
	set(&regime, p.Regime)
	set(&d.Category, p.Category)
	set(&d.Amount, p.Amount)
	set(&d.IVARate, p.IVARate)
	set(&d.IVA, p.IVA)
	set(&d.PaymentMethod, p.PaymentMethod)
	set(&d.PaidOn, p.PaidOn)
	set(&d.HasCFDI, p.HasCFDI)
	set(&d.CFDIUUID, p.CFDIUUID)
	set(&d.Deductible, p.Deductible)
	set(&d.Withheld.ISR, p.ISRWithheld)
	set(&d.Withheld.IVA, p.IVAWithheld)
	return regime, d
}

// set overwrites *field with *patch when the patch carries a value.
func set[T any](field *T, patch *T) {
	if patch != nil {
		*field = *patch
	}
}
