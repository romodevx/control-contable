package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// periodLayout is a list period: "YYYY-MM".
const periodLayout = "2006-01"

// firstDay completes a period into its first day.
const firstDay = "-01"

// cursorSeparator splits a cursor's booking date from its id.
const cursorSeparator = "|"

// Get returns a live income (getIncome). A soft-deleted, foreign or unknown
// income is 404.
func (uc *Incomes) Get(ctx context.Context, userID, rawTaxpayerID, rawIncomeID string) (Income, error) {
	taxpayerID, id, err := pathIDs(rawTaxpayerID, rawIncomeID)
	if err != nil {
		return Income{}, err
	}
	var out Income
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.findLive(ctx, userID, taxpayerID, id, LockNone)
		out = toIncome(stored)
		return err
	})
	if err != nil {
		return Income{}, present("get income", err)
	}
	return out, nil
}

// List returns one page of a month's live incomes (listIncomes): booked in
// the period, newest first by (booked_on, id), with the totals of the
// whole filtered period. Archived taxpayers are listed.
func (uc *Incomes) List(ctx context.Context, userID string, q ListQuery) (Page, error) {
	taxpayerID, ok := canonicalID(q.TaxpayerID)
	if !ok {
		return Page{}, taxpayerNotFound()
	}
	filter, after, limit, err := parseList(taxpayerID, q)
	if err != nil {
		return Page{}, err
	}
	var page Page
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		profile, err := uc.deps.Taxpayers.Profile(ctx, userID, taxpayerID)
		if err != nil {
			return err
		}
		rows, err := uc.deps.Store.List(ctx, filter, after, limit+1)
		if err != nil {
			return err
		}
		page.Totals, err = uc.deps.Store.Totals(ctx, filter)
		if err != nil {
			return err
		}
		if len(rows) > limit {
			rows = rows[:limit]
			last := rows[limit-1]
			page.NextCursor = encodeCursor(Keyset{BookedOn: last.BookedOn, ID: last.ID})
		}
		page.Items = make([]Income, len(rows))
		for i, row := range rows {
			if row, err = factsFor(row, profile.PersonType); err != nil {
				return err
			}
			page.Items[i] = toIncome(row)
		}
		return nil
	})
	if err != nil {
		return Page{}, present("list incomes", err)
	}
	return page, nil
}

// Delete soft-deletes an income (deleteIncome). Every user may delete, a
// read-only one included; deleting a deleted income is a no-op. An
// archived taxpayer refuses it (OQ-5: archive freezes the data).
func (uc *Incomes) Delete(ctx context.Context, userID, rawTaxpayerID, rawIncomeID string) error {
	taxpayerID, id, err := pathIDs(rawTaxpayerID, rawIncomeID)
	if err != nil {
		return err
	}
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.findOwned(ctx, userID, taxpayerID, id)
		if err != nil {
			return err
		}
		if err := uc.ensureNotArchived(ctx, userID, taxpayerID); err != nil {
			return err
		}
		if stored.Deleted {
			return nil
		}
		if err := uc.deps.Store.SetDeleted(ctx, id, true); err != nil {
			return err
		}
		uc.audit(ctx, "income.deleted", userID, taxpayerID, id)
		return nil
	})
	return present("delete income", err)
}

// Restore undoes a delete (restoreIncome). It re-checks the archive and the
// regime window, never reserves a trial slot (OQ-10: the slot it took at
// creation never left the count) and needs no write access: a read-only
// user may undo a delete (coordinator decision). Restoring a live income is
// a no-op.
func (uc *Incomes) Restore(ctx context.Context, userID, rawTaxpayerID, rawIncomeID string) (WriteResult, error) {
	taxpayerID, id, err := pathIDs(rawTaxpayerID, rawIncomeID)
	if err != nil {
		return WriteResult{}, err
	}
	var res WriteResult
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.findOwned(ctx, userID, taxpayerID, id)
		if err != nil {
			return err
		}
		if err := uc.ensureNotArchived(ctx, userID, taxpayerID); err != nil {
			return err
		}
		if !stored.Deleted {
			res = WriteResult{Income: toIncome(stored)}
			return nil
		}
		res, err = uc.restore(ctx, userID, stored)
		return err
	})
	if err != nil {
		return WriteResult{}, present("restore income", err)
	}
	return res, nil
}

func (uc *Incomes) restore(ctx context.Context, userID string, stored StoredIncome) (WriteResult, error) {
	today, err := uc.today()
	if err != nil {
		return WriteResult{}, err
	}
	if err := uc.checkWindow(ctx, userID, stored.TaxpayerID, stored.Fields, today); err != nil {
		return WriteResult{}, err
	}
	if err := uc.deps.Store.SetDeleted(ctx, stored.ID, false); err != nil {
		return WriteResult{}, err
	}
	derived, err := uc.rederive(ctx, userID, stored)
	if err != nil {
		return WriteResult{}, err
	}
	duplicate, err := uc.duplicateFolio(ctx, stored.TaxpayerID, stored.CFDIUUID, stored.ID)
	if err != nil {
		return WriteResult{}, err
	}
	warnings := slices.Concat(derived, duplicate)
	restored, err := uc.deps.Store.Find(ctx, stored.ID, LockNone)
	if err != nil {
		return WriteResult{}, err
	}
	uc.audit(ctx, "income.restored", userID, stored.TaxpayerID, stored.ID)
	return WriteResult{Income: toIncome(restored), Warnings: warnings}, nil
}

// rederive derives the restored income's withholding again with today's
// table and the taxpayer's current 625 options, stores it, and returns its
// warnings (ADR-0008: create, update and restore derive). A record stored
// before WH-06 derives on the facts a read serves (withFacts), as the WH-08
// backfill would.
func (uc *Incomes) rederive(ctx context.Context, userID string, stored StoredIncome) ([]domain.Warning, error) {
	profile, err := uc.deps.Taxpayers.Profile(ctx, userID, stored.TaxpayerID)
	if err != nil {
		return nil, err
	}
	held, ok := findRegime(profile, stored.Regime)
	if !ok {
		return nil, regimeNotHeldError()
	}
	payee, err := uc.payee(ctx, profile.PersonType, held)
	if err != nil {
		return nil, err
	}
	derivation, err := domain.Derive(draftOf(stored.Fields), payee, stored.PlatformActivity, uc.deps.Withholdings)
	if err != nil {
		return nil, err
	}
	if err := uc.deps.Store.Update(ctx, stored.ID, stored.Fields, derivation.Record()); err != nil {
		return nil, err
	}
	return domain.WarningsOf(derivation), nil
}

// findOwned locks the caller's income of the taxpayer, deleted or not, with
// its facts (withFacts); any other id is 404.
func (uc *Incomes) findOwned(ctx context.Context, userID, taxpayerID, id string) (StoredIncome, error) {
	stored, err := uc.deps.Store.Find(ctx, id, LockUpdate)
	if errors.Is(err, ErrIncomeNotFound) || (err == nil && stored.TaxpayerID != taxpayerID) {
		return StoredIncome{}, incomeNotFound()
	}
	if err != nil {
		return StoredIncome{}, err
	}
	return uc.withFacts(ctx, userID, stored)
}

// findLive is findOwned for a live income: a deleted one is 404.
func (uc *Incomes) findLive(ctx context.Context, userID, taxpayerID, id string, lock Lock) (StoredIncome, error) {
	stored, err := uc.deps.Store.Find(ctx, id, lock)
	if errors.Is(err, ErrIncomeNotFound) || (err == nil && (stored.TaxpayerID != taxpayerID || stored.Deleted)) {
		return StoredIncome{}, incomeNotFound()
	}
	if err != nil {
		return StoredIncome{}, err
	}
	return uc.withFacts(ctx, userID, stored)
}

// ensureNotArchived refuses a write on an archived taxpayer (422); an
// unknown one is the profile's 404.
func (uc *Incomes) ensureNotArchived(ctx context.Context, userID, taxpayerID string) error {
	profile, err := uc.deps.Taxpayers.Profile(ctx, userID, taxpayerID)
	if err != nil {
		return err
	}
	if profile.Archived {
		return archivedError()
	}
	return nil
}

// pathIDs canonicalizes the path ids: a malformed one is 404, never 400
// (QA O3).
func pathIDs(rawTaxpayerID, rawIncomeID string) (string, string, error) {
	taxpayerID, ok := canonicalID(rawTaxpayerID)
	if !ok {
		return "", "", taxpayerNotFound()
	}
	id, ok := canonicalID(rawIncomeID)
	if !ok {
		return "", "", incomeNotFound()
	}
	return taxpayerID, id, nil
}

// parseList validates the list query: the period's [first day, first day of
// the next month), the keyset of the cursor and the page size.
func parseList(taxpayerID string, q ListQuery) (ListFilter, Keyset, int, error) {
	if q.Period == "" {
		return ListFilter{}, Keyset{}, 0, fieldError(ErrInvalidPeriod, fieldPeriod, fiscal.ErrRequired, msgRequired)
	}
	month, err := time.Parse(periodLayout, q.Period)
	from, fromErr := fiscal.ParseDate(q.Period + firstDay)
	until, untilErr := fiscal.ParseDate(month.AddDate(0, 1, 0).Format(time.DateOnly))
	if err != nil || fromErr != nil || untilErr != nil {
		return ListFilter{}, Keyset{}, 0, fieldError(ErrInvalidPeriod, fieldPeriod, fiscal.ErrInvalid, msgInvalidField)
	}
	limit := q.Limit
	if limit == 0 {
		limit = DefaultLimit
	}
	if limit < 1 || limit > MaxLimit {
		return ListFilter{}, Keyset{}, 0, fieldError(ErrInvalidLimit, fieldLimit, fiscal.ErrOutOfRange, msgLimit)
	}
	after, err := decodeCursor(q.Cursor)
	if err != nil {
		return ListFilter{}, Keyset{}, 0, fieldError(err, fieldCursor, fiscal.ErrInvalid, msgInvalidField)
	}
	filter := ListFilter{
		TaxpayerID: taxpayerID, Regime: q.Regime,
		From: from, Until: until,
	}
	return filter, after, limit, nil
}

// encodeCursor is the opaque cursor after k.
func encodeCursor(k Keyset) string {
	return base64.RawURLEncoding.EncodeToString([]byte(k.BookedOn.String() + cursorSeparator + k.ID))
}

// decodeCursor reads a cursor encodeCursor issued; "" is the first page.
func decodeCursor(raw string) (Keyset, error) {
	if raw == "" {
		return Keyset{}, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Keyset{}, ErrInvalidCursor
	}
	booked, rawID, _ := strings.Cut(string(b), cursorSeparator)
	on, err := fiscal.ParseDate(booked)
	if err != nil {
		return Keyset{}, ErrInvalidCursor
	}
	id, ok := canonicalID(rawID)
	if !ok {
		return Keyset{}, ErrInvalidCursor
	}
	return Keyset{BookedOn: on, ID: id}, nil
}
