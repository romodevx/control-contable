package usecase

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// List page sizes (contract Limit).
const (
	defaultLimit = 50
	maxLimit     = 100
)

// periodLayout is a "YYYY-MM" tax period.
const periodLayout = "2006-01"

// cursorSeparator splits a cursor's paid_on from its id.
const cursorSeparator = "|"

// Get returns a live expense of the caller's taxpayer (M4-06). A deleted,
// purged, foreign or unknown expense, or a malformed id, is 404. Reads
// carry no warnings; an archived taxpayer stays readable.
func (u *Expenses) Get(ctx context.Context, userID, rawTaxpayerID, rawExpenseID string) (Expense, error) {
	var out Expense
	err := u.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		taxpayerID, _, err := u.profile(ctx, userID, rawTaxpayerID)
		if err != nil {
			return err
		}
		e, err := u.findExpense(ctx, taxpayerID, rawExpenseID, LockNone)
		if err != nil {
			return err
		}
		if e.Deleted {
			return notFound()
		}
		out = view(e, nil)
		return nil
	})
	if err != nil {
		return Expense{}, present("get expense", err)
	}
	return out, nil
}

// List returns one keyset page of the live expenses paid in a month
// (M4-06, MODULE-incomes §2.9), newest first (paid_on, then id, both
// descending), optionally of one regime, with the whole period's totals.
// The taxpayer is checked before the query parameters, so a malformed or
// foreign taxpayer is 404 whatever else is wrong.
func (u *Expenses) List(ctx context.Context, in ListInput) (Page, error) {
	var out Page
	err := u.deps.Tx.ReadOnly(ctx, in.UserID, func(ctx context.Context) error {
		taxpayerID, _, err := u.profile(ctx, in.UserID, in.TaxpayerID)
		if err != nil {
			return err
		}
		q, err := listQuery(taxpayerID, in)
		if err != nil {
			return err
		}
		out, err = u.page(ctx, q)
		return err
	})
	if err != nil {
		return Page{}, present("list expenses", err)
	}
	return out, nil
}

func (u *Expenses) page(ctx context.Context, q ListQuery) (Page, error) {
	limit := q.Limit
	q.Limit = limit + 1 // one more tells whether a next page exists
	rows, err := u.deps.Expenses.List(ctx, q)
	if err != nil {
		return Page{}, err
	}
	totals, err := u.deps.Expenses.Totals(ctx, q.PeriodQuery)
	if err != nil {
		return Page{}, err
	}
	page := Page{Items: make([]Expense, 0, min(len(rows), limit)), Totals: totals, Legend: domain.Legend}
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[limit-1]
		page.NextCursor = encodeCursor(Cursor{PaidOn: last.PaidOn, ID: last.ID})
	}
	for _, e := range rows {
		page.Items = append(page.Items, view(e, nil))
	}
	return page, nil
}

// listQuery validates the list parameters, reporting every bad one.
func listQuery(taxpayerID string, in ListInput) (ListQuery, error) {
	var vs fiscal.Violations
	from, until, problem, ok := monthRange(in.Period)
	if !ok {
		vs.Add(fieldPeriod, problem)
	}
	q := ListQuery{
		PeriodQuery: PeriodQuery{TaxpayerID: taxpayerID, From: from, Until: until, Regime: in.Regime},
		Limit:       in.Limit,
	}
	switch {
	case in.Limit == 0:
		q.Limit = defaultLimit
	case in.Limit < 1 || in.Limit > maxLimit:
		vs.Add(fieldLimit, fiscal.ErrOutOfRange)
	}
	if in.Cursor != "" {
		c, ok := decodeCursor(in.Cursor)
		if !ok {
			vs.Add(fieldCursor, fiscal.ErrInvalid)
		}
		q.After = &c
	}
	if len(vs) > 0 {
		return ListQuery{}, violationsError(vs)
	}
	return q, nil
}

// monthRange returns [first day of period, first day of the next month),
// or the period's problem (required, invalid) and false.
func monthRange(period string) (from, until fiscal.Date, problem fiscal.Error, ok bool) {
	if period == "" {
		return from, until, fiscal.ErrRequired, false
	}
	month, err := time.Parse(periodLayout, period)
	if err != nil || len(period) != len(periodLayout) {
		return from, until, fiscal.ErrInvalid, false
	}
	from, errFrom := fiscal.ParseDate(month.Format(dayLayout))
	until, errUntil := fiscal.ParseDate(month.AddDate(0, 1, 0).Format(dayLayout))
	if errFrom != nil || errUntil != nil {
		return fiscal.Date{}, fiscal.Date{}, fiscal.ErrInvalid, false
	}
	return from, until, problem, true
}

// encodeCursor is the opaque form of a keyset position.
func encodeCursor(c Cursor) string {
	return base64.RawURLEncoding.EncodeToString([]byte(c.PaidOn.String() + cursorSeparator + c.ID))
}

// decodeCursor parses a cursor encodeCursor made; anything else is invalid.
func decodeCursor(raw string) (Cursor, bool) {
	text, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Cursor{}, false
	}
	day, rawID, ok := strings.Cut(string(text), cursorSeparator)
	if !ok {
		return Cursor{}, false
	}
	paidOn, err := fiscal.ParseDate(day)
	if err != nil {
		return Cursor{}, false
	}
	id, ok := canonicalID(rawID)
	return Cursor{PaidOn: paidOn, ID: id}, ok
}

// Categories returns the category catalog resolved for a regime the
// taxpayer holds on a date (M4-06 query for listExpenseCategories,
// MODULE-expenses §2.2-§2.3, US-E1b): each category's hint and default
// after the regime's options, its note and basis, plus the IVA rates
// allowed and the cash-payment limit on that date, and the legend.
// Allowed for a read-only user and an archived taxpayer.
func (u *Expenses) Categories(ctx context.Context, in CategoriesInput) (CategoryList, error) {
	var out CategoryList
	err := u.deps.Tx.ReadOnly(ctx, in.UserID, func(ctx context.Context) error {
		_, profile, err := u.profile(ctx, in.UserID, in.TaxpayerID)
		if err != nil {
			return err
		}
		on, err := u.categoriesDate(in)
		if err != nil {
			return err
		}
		hold, ok := profile.holding(in.Regime, on.Period())
		if !ok {
			return regimeNotActiveError()
		}
		out, err = u.categories(ctx, profile, hold, on)
		return err
	})
	if err != nil {
		return CategoryList{}, present("list expense categories", err)
	}
	return out, nil
}

// categoriesDate validates the regime and the date ("" = today), which
// must lie in [fiscal.MinRecordDate, today].
func (u *Expenses) categoriesDate(in CategoriesInput) (fiscal.Date, error) {
	today, err := u.today()
	if err != nil {
		return fiscal.Date{}, err
	}
	var vs fiscal.Violations
	if in.Regime == "" {
		vs.Add(fieldRegime, fiscal.ErrRequired)
	}
	on := today
	if in.On != "" {
		on, err = fiscal.ParseDate(in.On)
		switch {
		case err != nil:
			vs.Add(fieldOn, fiscal.ErrInvalid)
		case on.Before(fiscal.MinRecordDate) || on.After(today):
			vs.Add(fieldOn, fiscal.ErrOutOfRange)
		}
	}
	if len(vs) > 0 {
		return fiscal.Date{}, violationsError(vs)
	}
	return on, nil
}

func (u *Expenses) categories(ctx context.Context, profile TaxpayerProfile, hold RegimeHold, on fiscal.Date) (CategoryList, error) {
	dated, err := u.deps.Parameters.On(on)
	if errors.Is(err, ErrParametersUnavailable) {
		return CategoryList{}, invalidField(fieldOn, fiscal.ErrOutOfRange)
	}
	if err != nil {
		return CategoryList{}, err
	}
	entries, err := u.deps.Categories.Rules(ctx, hold.Code, profile.PersonType)
	if err != nil {
		return CategoryList{}, err
	}
	payer := profile.payer(hold)
	items := make([]Category, 0, len(entries))
	for _, entry := range entries {
		hint := domain.ResolveDeductibility(entry.Rule, payer)
		items = append(items, Category{
			Code: entry.Rule.Category, Name: entry.Name, Deductibility: hint.Deductibility,
			DefaultDeductible: hint.DefaultDeductible, Note: entry.Note, LegalBasis: entry.LegalBasis,
		})
	}
	return CategoryList{
		Items: items, IVARatesAllowed: dated.IVARatesAllowed, CashPaymentLimit: dated.CashPaymentLimit, Legend: domain.Legend,
	}, nil
}
