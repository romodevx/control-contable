package usecase

import (
	"context"
	"errors"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/access"
)

// Create records an expense of the caller's taxpayer (M4-05,
// MODULE-expenses §2, MODULE-incomes §2.3/§2.10). In one unit of work:
//
//  1. Idempotency: a client id is locked, then looked up. The same request
//     already stored returns it (Created false, no slot used, even at the
//     trial cap, with no warnings: they are computed on the write that
//     stored it); other content, another taxpayer or a deleted expense is
//     409 id_reused; an id another user holds is 409 without a detail.
//  2. The read-only check (402), then the archived check (422).
//  3. The write rules (judge): regime window, category rule, domain rules,
//     duplicate-folio and provisional-parameter warnings.
//  4. ReserveRecord(expense) (402 expenses_remaining at the trial cap): last,
//     so a refused or failed create never spends a slot; a failed insert
//     rolls the slot back with the transaction.
//  5. Insert, then the expense.created audit event (ids only).
func (u *Expenses) Create(ctx context.Context, in CreateInput) (CreateResult, error) {
	description, err := normalizeDescription(in.Description)
	if err != nil {
		return CreateResult{}, err
	}
	id, err := u.expenseID(in.ID)
	if err != nil {
		return CreateResult{}, err
	}
	var out CreateResult
	err = u.deps.Tx.ReadWrite(ctx, in.UserID, func(ctx context.Context) error {
		var err error
		out, err = u.create(ctx, in, id, in.ID != "", description)
		return err
	})
	if err != nil {
		return CreateResult{}, present("create expense", err)
	}
	if out.Created {
		u.audit(ctx, "expense.created", in.UserID, out.Expense.TaxpayerID, out.Expense.ID)
	}
	return out, nil
}

// expenseID is the canonical client id, or a new server id.
func (u *Expenses) expenseID(raw string) (string, error) {
	if raw == "" {
		return u.deps.NewID()
	}
	id, ok := canonicalID(raw)
	if !ok {
		return "", invalidField(fieldID, fiscal.ErrInvalid)
	}
	return id, nil
}

func (u *Expenses) create(ctx context.Context, in CreateInput, id string, fromClient bool, description string) (CreateResult, error) {
	taxpayerID, profile, err := u.profile(ctx, in.UserID, in.TaxpayerID)
	if err != nil {
		return CreateResult{}, err
	}
	if fromClient {
		replay, found, err := u.replay(ctx, id, taxpayerID, in, description)
		if found || err != nil {
			return replay, err
		}
	}
	if err := u.deps.Access.EnsureWritable(ctx, in.UserID); err != nil {
		return CreateResult{}, err
	}
	if profile.Archived {
		return CreateResult{}, archivedError()
	}
	v, err := u.judge(ctx, subject{userID: in.UserID, taxpayerID: taxpayerID, profile: profile, regime: in.Regime, draft: in.Draft})
	if err != nil {
		return CreateResult{}, err
	}
	if err := u.deps.Access.ReserveRecord(ctx, in.UserID, access.RecordExpense); err != nil {
		return CreateResult{}, err
	}
	row := stored(StoredExpense{ID: id, TaxpayerID: taxpayerID}, in.Regime, in.Draft, description, v.assessment)
	saved, err := u.deps.Expenses.Insert(ctx, in.UserID, row)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Expense: view(saved, v.warnings), Created: true}, nil
}

// replay answers a client id the caller already used (MODULE-incomes §2.3):
// found reports that the id is taken by the caller, with the stored expense
// when the request is the same, else 409 id_reused.
func (u *Expenses) replay(ctx context.Context, id, taxpayerID string, in CreateInput, description string) (CreateResult, bool, error) {
	if err := u.deps.Expenses.LockID(ctx, id); err != nil {
		return CreateResult{}, false, err
	}
	existing, err := u.deps.Expenses.Find(ctx, id, LockNone)
	if errors.Is(err, ErrExpenseNotFound) {
		return CreateResult{}, false, nil
	}
	if err != nil {
		return CreateResult{}, false, err
	}
	if !sameRequest(existing, taxpayerID, in, description) {
		return CreateResult{}, true, idReusedError()
	}
	return CreateResult{Expense: view(existing, nil)}, true, nil
}

// sameRequest compares the persisted fields of a live expense with a
// create request, after normalization. A default deductible is derived,
// so only an override is compared.
func sameRequest(e StoredExpense, taxpayerID string, in CreateInput, description string) bool {
	d := in.Draft
	return !e.Deleted && e.TaxpayerID == taxpayerID && e.Regime == in.Regime && e.Category == d.Category &&
		e.Amount == d.Amount && e.IVARate == d.IVARate && e.IVA == d.IVA && e.PaymentMethod == d.PaymentMethod &&
		e.PaidOn == d.PaidOn && e.HasCFDI == d.HasCFDI && e.CFDIUUID == d.CFDIUUID && e.Withheld == d.Withheld &&
		e.Description == description && overrideOf(e) == d.Deductible
}
