package usecase

import "context"

// Delete soft-deletes an expense (M4-06, MODULE-incomes §2.7): it leaves
// lists, totals and the estimate at once and can be restored until the
// purge. Deleting a deleted expense succeeds (idempotent). A read-only user
// may delete; a slot is never given back. An archived taxpayer refuses it
// (422 taxpayer_archived): archive freezes the data.
func (u *Expenses) Delete(ctx context.Context, userID, rawTaxpayerID, rawExpenseID string) error {
	var deleted StoredExpense
	err := u.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		e, _, err := u.lockForLifecycle(ctx, userID, rawTaxpayerID, rawExpenseID)
		if err != nil || e.Deleted {
			return err
		}
		deleted, err = u.deps.Expenses.SetDeleted(ctx, e.ID, true)
		return err
	})
	if err != nil {
		return present("delete expense", err)
	}
	if deleted.ID != "" {
		u.audit(ctx, "expense.deleted", userID, deleted.TaxpayerID, deleted.ID)
	}
	return nil
}

// Restore undoes a delete (M4-06). DECIDED (coordinator, MODULE-incomes
// §2.7): restore never consumes a trial slot (the slot was spent at
// creation and never freed) and a read-only user may restore their own
// records, so no access check runs. The regime window is re-checked (422
// regime_not_active), an archived taxpayer refuses it, and the warnings
// are computed as on create. Restoring a live expense returns it unchanged.
func (u *Expenses) Restore(ctx context.Context, userID, rawTaxpayerID, rawExpenseID string) (Expense, error) {
	var (
		out      Expense
		restored bool
	)
	err := u.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		e, profile, err := u.lockForLifecycle(ctx, userID, rawTaxpayerID, rawExpenseID)
		if err != nil {
			return err
		}
		if !e.Deleted {
			out = view(e, nil)
			return nil
		}
		out, err = u.restore(ctx, userID, e, profile)
		restored = err == nil
		return err
	})
	if err != nil {
		return Expense{}, present("restore expense", err)
	}
	if restored {
		u.audit(ctx, "expense.restored", userID, out.TaxpayerID, out.ID)
	}
	return out, nil
}

func (u *Expenses) restore(ctx context.Context, userID string, e StoredExpense, profile TaxpayerProfile) (Expense, error) {
	v, err := u.judge(ctx, subject{
		userID: userID, taxpayerID: e.TaxpayerID, expenseID: e.ID, profile: profile, regime: e.Regime, draft: draftOf(e),
	})
	if err != nil {
		return Expense{}, err
	}
	if _, err := u.deps.Expenses.Update(ctx, stored(e, e.Regime, draftOf(e), e.Description, v.assessment)); err != nil {
		return Expense{}, err
	}
	saved, err := u.deps.Expenses.SetDeleted(ctx, e.ID, false)
	if err != nil {
		return Expense{}, err
	}
	return view(saved, v.warnings), nil
}

// lockForLifecycle reads the path's taxpayer and locks its expense, live or
// deleted, refusing an archived taxpayer.
func (u *Expenses) lockForLifecycle(ctx context.Context, userID, rawTaxpayerID, rawExpenseID string) (StoredExpense, TaxpayerProfile, error) {
	taxpayerID, profile, err := u.profile(ctx, userID, rawTaxpayerID)
	if err != nil {
		return StoredExpense{}, TaxpayerProfile{}, err
	}
	e, err := u.findExpense(ctx, taxpayerID, rawExpenseID, LockUpdate)
	if err != nil {
		return StoredExpense{}, TaxpayerProfile{}, err
	}
	if profile.Archived {
		return StoredExpense{}, TaxpayerProfile{}, archivedError()
	}
	return e, profile, nil
}
