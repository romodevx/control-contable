package annual

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"slices"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// MaxDeductionsPerYear bounds the live personal deductions of a taxpayer
// and fiscal year (ADR-0009): it keeps the unpaginated list and the
// O(rows) deduction pass inside the p95 target. Create, a move to another
// year and restore all re-check it.
const MaxDeductionsPerYear = 500

// maxStudentLabel is the contract StudentLabel maxLength, in characters.
const maxStudentLabel = 40

// firstDeductionDay is the earliest paid_on: the first day with parameters
// (contract "before 2026-01-01 → out_of_range", as BusinessDate).
var firstDeductionDay = mustParseDate("2026-01-01")

// The audit events (ids only).
const (
	eventDeductionCreated  = "personal_deduction.created"
	eventDeductionUpdated  = "personal_deduction.updated"
	eventDeductionDeleted  = "personal_deduction.deleted"
	eventDeductionRestored = "personal_deduction.restored"
)

// Exclusion is why a stored deduction is never applied (contract
// DeductionExclusion); "" when it can be.
type Exclusion string

// ExclusionCashNotDeductible is a cash payment of a cash-restricted category.
const ExclusionCashNotDeductible Exclusion = "cash_not_deductible"

// DeductionInput is a createPersonalDeduction body after the schema
// (enums, formats) and the path's taxpayer. A nil pointer is absent or
// null.
type DeductionInput struct {
	ID            string
	TaxpayerID    string
	Category      adomain.Category
	Amount        money.Cents
	PaidOn        fiscal.Date
	PaymentMethod adomain.PaymentMethod
	Beneficiary   adomain.Beneficiary
	DoneeType     *adomain.DoneeType
	StudentLabel  *string
	TuitionLevel  *adomain.TuitionLevel
	Scholarship   *money.Cents
	CFDIUUID      *string
}

// DeductionPatch is an updatePersonalDeduction merge-patch: a nil pointer
// or an unset Optional is unchanged.
type DeductionPatch struct {
	Category      *adomain.Category
	Amount        *money.Cents
	PaidOn        *fiscal.Date
	PaymentMethod *adomain.PaymentMethod
	Beneficiary   *adomain.Beneficiary
	DoneeType     Optional[adomain.DoneeType]
	StudentLabel  Optional[string]
	TuitionLevel  Optional[adomain.TuitionLevel]
	Scholarship   Optional[money.Cents]
	CFDIUUID      Optional[string]
}

// Deduction is a live personal deduction as the contract shows it.
type Deduction struct {
	ID         string
	TaxpayerID string
	FiscalYear int
	DeductionFields
	Excluded  Exclusion
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DeductionResult is a write's answer. Created is false for an idempotent
// retry (200) or any write that is not a create.
type DeductionResult struct {
	Deduction Deduction
	Warnings  []Warning
	Created   bool
}

// DeductionQuery selects the list of one taxpayer and fiscal year,
// optionally of one category.
type DeductionQuery struct {
	TaxpayerID string
	FiscalYear int
	Category   adomain.Category
}

// CategoryTotal are the registered amounts of one category (contract
// DeductionCategoryTotal): Excluded is part of Registered.
type CategoryTotal struct {
	Category   adomain.Category
	Registered money.Cents
	Excluded   money.Cents
	Count      int
}

// DeductionList is listPersonalDeductions.
type DeductionList struct {
	FiscalYear int
	Items      []Deduction
	Totals     []CategoryTotal
}

// DeductionsDeps are the ports the personal-deduction use cases need.
type DeductionsDeps struct {
	Tx        tx.Manager
	Access    AccessPolicy
	Taxpayers TaxpayerStore
	Store     DeductionStore
	Clock     clock.Clock
	// NewID generates the id of a create without a client id; nil is
	// UUIDv7.
	NewID func() (string, error)
	// Logger receives the audit events; nil discards them.
	Logger *slog.Logger
}

// Deductions are the personal-deduction use cases (MODULE-annual §2.6,
// US-A10; contract list/create/update/delete/restorePersonalDeduction).
type Deductions struct {
	base
	deps DeductionsDeps
}

// NewDeductions returns the use cases over deps.
func NewDeductions(deps DeductionsDeps) *Deductions {
	if deps.NewID == nil {
		deps.NewID = newUUIDv7
	}
	return &Deductions{base: newBase(deps.Clock, deps.Logger), deps: deps}
}

// List returns the live deductions of q (listPersonalDeductions): by
// category in article order, then paid_on descending, then id, with the
// registered totals per category. A persona moral has none. Reads stay
// open to a read-only user and an archived taxpayer.
func (uc *Deductions) List(ctx context.Context, userID string, q DeductionQuery) (DeductionList, error) {
	tid, err := taxpayerID(q.TaxpayerID)
	if err != nil {
		return DeductionList{}, err
	}
	if err := uc.checkFutureYear(q.FiscalYear, FieldFiscalYear); err != nil {
		return DeductionList{}, err
	}
	out := DeductionList{FiscalYear: q.FiscalYear, Items: []Deduction{}, Totals: []CategoryTotal{}}
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		t, err := uc.deps.Taxpayers.Taxpayer(ctx, tid, LockNone)
		if err != nil {
			return err
		}
		if err := checkYearFloor(q.FiscalYear, FieldFiscalYear); err != nil || t.PersonType != fiscal.PersonFisica {
			return err
		}
		rows, err := uc.deps.Store.ListDeductions(ctx, tid, q.FiscalYear, q.Category)
		if err != nil {
			return err
		}
		out.Items, out.Totals, err = listOf(rows)
		return err
	})
	if err != nil {
		return DeductionList{}, present("list personal deductions", err)
	}
	return out, nil
}

// listOf orders the rows and sums them per category. O(n log n), n ≤ 500.
func listOf(rows []StoredDeduction) ([]Deduction, []CategoryTotal, error) {
	sortDeductions(rows)
	items := make([]Deduction, 0, len(rows))
	totals := []CategoryTotal{}
	for _, row := range rows {
		d := toDeduction(row)
		items = append(items, d)
		if len(totals) == 0 || totals[len(totals)-1].Category != d.Category {
			totals = append(totals, CategoryTotal{Category: d.Category})
		}
		total := &totals[len(totals)-1]
		var err error
		if total.Registered, err = total.Registered.Add(d.Amount); err != nil {
			return nil, nil, err
		}
		if d.Excluded != "" {
			if total.Excluded, err = total.Excluded.Add(d.Amount); err != nil {
				return nil, nil, err
			}
		}
		total.Count++
	}
	return items, totals, nil
}

// sortDeductions orders rows by category in article order, then paid_on
// descending, then id: the list order, and a deterministic calculator
// input.
func sortDeductions(rows []StoredDeduction) {
	rank := map[adomain.Category]int{}
	for i, c := range adomain.Categories() {
		rank[c] = i
	}
	slices.SortFunc(rows, func(a, b StoredDeduction) int {
		return cmp.Or(cmp.Compare(rank[a.Category], rank[b.Category]), b.PaidOn.Compare(a.PaidOn), cmp.Compare(a.ID, b.ID))
	})
}

// Create records a personal deduction (createPersonalDeduction). Order:
// ids (404/400) → idempotency (200/409) → read-only (402) → taxpayer
// (404), archived, person type, list bound (422) → field rules (400) →
// insert.
func (uc *Deductions) Create(ctx context.Context, userID string, in DeductionInput) (DeductionResult, error) {
	tid, err := taxpayerID(in.TaxpayerID)
	if err != nil {
		return DeductionResult{}, err
	}
	id, err := clientID(in.ID, uc.deps.NewID)
	if err != nil {
		return DeductionResult{}, present("create personal deduction", err)
	}
	fromClient := in.ID != ""
	in.ID, in.TaxpayerID = id, tid
	var res DeductionResult
	for attempt := 1; ; attempt++ {
		err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
			var err error
			res, err = uc.create(ctx, userID, in, fromClient)
			return err
		})
		if !errors.Is(err, ErrIDTaken) || attempt == createAttempts {
			break
		}
	}
	if err != nil {
		return DeductionResult{}, present("create personal deduction", err)
	}
	return res, nil
}

func (uc *Deductions) create(ctx context.Context, userID string, in DeductionInput, fromClient bool) (DeductionResult, error) {
	fields := normalizeDeduction(in)
	if fromClient {
		if err := uc.deps.Store.LockDeductionID(ctx, in.ID); err != nil {
			return DeductionResult{}, err
		}
	}
	existing, err := uc.deps.Store.FindDeduction(ctx, in.ID, LockNone)
	switch {
	case err == nil:
		if existing.TaxpayerID != in.TaxpayerID || existing.Deleted || existing.DeductionFields != fields {
			return DeductionResult{}, ErrIDReused
		}
		return resultOf(existing, false), nil
	case !errors.Is(err, ErrDeductionNotFound):
		return DeductionResult{}, err
	}
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return DeductionResult{}, err
	}
	if _, err := ensureWritable(ctx, uc.deps.Taxpayers, in.TaxpayerID, true); err != nil {
		return DeductionResult{}, err
	}
	if err := uc.checkBound(ctx, in.TaxpayerID, fields.PaidOn.Year()); err != nil {
		return DeductionResult{}, err
	}
	if err := uc.checkFields(in); err != nil {
		return DeductionResult{}, err
	}
	if err := uc.deps.Store.InsertDeduction(ctx, NewDeduction{ID: in.ID, UserID: userID, TaxpayerID: in.TaxpayerID, DeductionFields: fields}); err != nil {
		return DeductionResult{}, err
	}
	stored, err := uc.deps.Store.FindDeduction(ctx, in.ID, LockNone)
	if err != nil {
		return DeductionResult{}, err
	}
	uc.auditRow(ctx, eventDeductionCreated, userID, stored)
	return resultOf(stored, true), nil
}

// Update edits a live deduction (updatePersonalDeduction, merge-patch),
// re-running every rule of a create on the merged row; a move to another
// fiscal year applies that year's bound.
func (uc *Deductions) Update(ctx context.Context, userID, rawTaxpayerID, rawID string, p DeductionPatch) (DeductionResult, error) {
	tid, id, err := pathIDs(rawTaxpayerID, rawID, ErrDeductionNotFound)
	if err != nil {
		return DeductionResult{}, err
	}
	var res DeductionResult
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.find(ctx, tid, id, false)
		if err != nil {
			return err
		}
		if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
			return err
		}
		if _, err := ensureWritable(ctx, uc.deps.Taxpayers, tid, true); err != nil {
			return err
		}
		in := p.apply(stored)
		fields := normalizeDeduction(in)
		if year := fields.PaidOn.Year(); year != stored.FiscalYear {
			if err := uc.checkBound(ctx, tid, year); err != nil {
				return err
			}
		}
		if err := uc.checkFields(in); err != nil {
			return err
		}
		if err := uc.deps.Store.UpdateDeduction(ctx, id, fields); err != nil {
			return err
		}
		updated, err := uc.deps.Store.FindDeduction(ctx, id, LockNone)
		if err != nil {
			return err
		}
		uc.auditRow(ctx, eventDeductionUpdated, userID, updated)
		res = resultOf(updated, false)
		return nil
	})
	if err != nil {
		return DeductionResult{}, present("update personal deduction", err)
	}
	return res, nil
}

// Delete soft-deletes a deduction (deletePersonalDeduction). Open to a
// read-only user; deleting a deleted one is a no-op; an archived taxpayer
// refuses it.
func (uc *Deductions) Delete(ctx context.Context, userID, rawTaxpayerID, rawID string) error {
	tid, id, err := pathIDs(rawTaxpayerID, rawID, ErrDeductionNotFound)
	if err != nil {
		return err
	}
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.find(ctx, tid, id, true)
		if err != nil {
			return err
		}
		if _, err := ensureWritable(ctx, uc.deps.Taxpayers, tid, false); err != nil {
			return err
		}
		if stored.Deleted {
			return nil
		}
		if err := uc.deps.Store.SetDeductionDeleted(ctx, id, true); err != nil {
			return err
		}
		uc.auditRow(ctx, eventDeductionDeleted, userID, stored)
		return nil
	})
	return present("delete personal deduction", err)
}

// Restore undoes a delete (restorePersonalDeduction). Open to a read-only
// user (no access check); the bound of the row's year is re-checked, so
// Undo is never a way past it; restoring a live row answers it unchanged.
func (uc *Deductions) Restore(ctx context.Context, userID, rawTaxpayerID, rawID string) (DeductionResult, error) {
	tid, id, err := pathIDs(rawTaxpayerID, rawID, ErrDeductionNotFound)
	if err != nil {
		return DeductionResult{}, err
	}
	var res DeductionResult
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.find(ctx, tid, id, true)
		if err != nil {
			return err
		}
		if _, err := ensureWritable(ctx, uc.deps.Taxpayers, tid, false); err != nil {
			return err
		}
		if !stored.Deleted {
			res = resultOf(stored, false)
			return nil
		}
		if err := uc.checkBound(ctx, tid, stored.FiscalYear); err != nil {
			return err
		}
		if err := uc.deps.Store.SetDeductionDeleted(ctx, id, false); err != nil {
			return err
		}
		restored, err := uc.deps.Store.FindDeduction(ctx, id, LockNone)
		if err != nil {
			return err
		}
		uc.auditRow(ctx, eventDeductionRestored, userID, restored)
		res = resultOf(restored, false)
		return nil
	})
	if err != nil {
		return DeductionResult{}, present("restore personal deduction", err)
	}
	return res, nil
}

// find locks the caller's deduction of the taxpayer; a row of another
// taxpayer, or a deleted one unless withDeleted, is 404.
func (uc *Deductions) find(ctx context.Context, taxpayerID, id string, withDeleted bool) (StoredDeduction, error) {
	stored, err := uc.deps.Store.FindDeduction(ctx, id, LockWrite)
	if err != nil {
		return StoredDeduction{}, err
	}
	if stored.TaxpayerID != taxpayerID || (stored.Deleted && !withDeleted) {
		return StoredDeduction{}, ErrDeductionNotFound
	}
	return stored, nil
}

// checkBound refuses a row that would make the year hold more than
// MaxDeductionsPerYear live deductions. The list lock makes the count and
// the write that follows one step against concurrent writers.
func (uc *Deductions) checkBound(ctx context.Context, taxpayerID string, fiscalYear int) error {
	if err := uc.deps.Taxpayers.LockList(ctx, taxpayerID, ListDeductions, fiscalYear); err != nil {
		return err
	}
	n, err := uc.deps.Store.CountLiveDeductions(ctx, taxpayerID, fiscalYear)
	if err != nil {
		return err
	}
	if n >= MaxDeductionsPerYear {
		return limitError(msgLimitDeductions)
	}
	return nil
}

// checkFields applies the field rules (400) to in.
func (uc *Deductions) checkFields(in DeductionInput) error {
	today, err := uc.today()
	if err != nil {
		return err
	}
	if fields := deductionFieldErrors(in, today); len(fields) > 0 {
		return fieldErrors(fields)
	}
	return nil
}

func (uc *Deductions) auditRow(ctx context.Context, event, userID string, d StoredDeduction) {
	uc.audit(ctx, event, slog.String("user_id", userID), slog.String("taxpayer_id", d.TaxpayerID),
		slog.String("deduction_id", d.ID))
}

// apply returns the stored row with p applied, as an input to check again.
func (p DeductionPatch) apply(stored StoredDeduction) DeductionInput {
	in := inputOf(stored)
	setIf(&in.Category, p.Category)
	setIf(&in.Amount, p.Amount)
	setIf(&in.PaidOn, p.PaidOn)
	setIf(&in.PaymentMethod, p.PaymentMethod)
	setIf(&in.Beneficiary, p.Beneficiary)
	in.DoneeType = p.DoneeType.apply(in.DoneeType)
	in.StudentLabel = p.StudentLabel.apply(in.StudentLabel)
	in.TuitionLevel = p.TuitionLevel.apply(in.TuitionLevel)
	in.Scholarship = p.Scholarship.apply(in.Scholarship)
	in.CFDIUUID = p.CFDIUUID.apply(in.CFDIUUID)
	return in
}

func setIf[T any](dst *T, v *T) {
	if v != nil {
		*dst = *v
	}
}

// inputOf is a stored row as an input: its null fields are nil.
func inputOf(d StoredDeduction) DeductionInput {
	in := DeductionInput{
		ID: d.ID, TaxpayerID: d.TaxpayerID, Category: d.Category, Amount: d.Amount, PaidOn: d.PaidOn,
		PaymentMethod: d.PaymentMethod, Beneficiary: d.Beneficiary,
		DoneeType: nonZero(d.DoneeType), StudentLabel: nonZero(d.StudentLabel), TuitionLevel: nonZero(d.TuitionLevel),
		CFDIUUID: nonZero(d.CFDIUUID),
	}
	if d.Scholarship.Valid {
		in.Scholarship = &d.Scholarship.Cents
	}
	return in
}

func nonZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func valueOr[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

// normalizeDeduction turns an input into the fields to store and compare:
// the label NFC, the folio in canonical upper case (an unparseable one is
// kept as sent, so it never equals a stored row and the field rules refuse
// it).
func normalizeDeduction(in DeductionInput) DeductionFields {
	f := DeductionFields{
		Category: in.Category, Amount: in.Amount, PaidOn: in.PaidOn, PaymentMethod: in.PaymentMethod,
		Beneficiary: in.Beneficiary, DoneeType: valueOr(in.DoneeType), TuitionLevel: valueOr(in.TuitionLevel),
		StudentLabel: normalizeLabel(valueOr(in.StudentLabel)), CFDIUUID: valueOr(in.CFDIUUID),
	}
	if in.Scholarship != nil {
		f.Scholarship = NullCents{Cents: *in.Scholarship, Valid: true}
	}
	if u, err := fiscal.ParseCFDIUUID(f.CFDIUUID); err == nil {
		f.CFDIUUID = u.String()
	}
	return f
}

// deductionFieldErrors are the field rules of MODULE-annual §2.6 and the
// contract, against the calculator's own vocabularies, in field order.
func deductionFieldErrors(in DeductionInput, today fiscal.Date) []apperr.FieldError {
	var out []apperr.FieldError
	add := func(field, code, message string) {
		out = append(out, apperr.FieldError{Field: field, Code: code, Message: message})
	}
	if in.Amount < 1 {
		add(FieldAmount, DetailInvalid, msgInvalidValue)
	}
	if !in.Category.Valid() {
		add(FieldCategory, DetailInvalid, msgInvalidValue)
	}
	if in.PaidOn.Before(firstDeductionDay) || in.PaidOn.After(today) {
		add(FieldPaidOn, DetailOutOfRange, msgDateRange)
	}
	if !in.PaymentMethod.Valid() {
		add(FieldPaymentMethod, DetailInvalid, msgInvalidValue)
	}
	if !in.Beneficiary.Valid() {
		add(FieldBeneficiary, DetailInvalid, msgInvalidValue)
	}
	donation, tuition := in.Category == adomain.CategoryDonation, in.Category == adomain.CategoryTuition
	categoryField(add, FieldDoneeType, donation, in.DoneeType, func(v adomain.DoneeType) bool { return v.Valid() })
	categoryField(add, FieldStudentLabel, tuition, in.StudentLabel, func(v string) bool { return validLabel(normalizeLabel(v), maxStudentLabel) })
	categoryField(add, FieldTuitionLevel, tuition, in.TuitionLevel, func(v adomain.TuitionLevel) bool { return v.Valid() })
	categoryField(add, FieldScholarship, tuition, in.Scholarship, func(v money.Cents) bool { return v >= 0 })
	if in.CFDIUUID != nil {
		if _, err := fiscal.ParseCFDIUUID(*in.CFDIUUID); err != nil {
			add(FieldCFDIUUID, DetailInvalid, msgInvalidValue)
		}
	}
	return out
}

// categoryField checks a field that exists for one category only: required
// (and valid) when applies, absent otherwise.
func categoryField[T any](add func(field, code, message string), field string, applies bool, v *T, valid func(T) bool) {
	switch {
	case applies && v == nil:
		add(field, DetailRequired, msgRequired)
	case !applies && v != nil:
		add(field, DetailNotApplicable, msgNotApplicable)
	case applies && !valid(*v):
		add(field, DetailInvalid, msgInvalidValue)
	}
}

// exclusionOf is why the fields can never be applied (§2.7 step 1).
func exclusionOf(f DeductionFields) Exclusion {
	if f.PaymentMethod == adomain.PaymentCash && f.Category.CashRestricted() {
		return ExclusionCashNotDeductible
	}
	return ""
}

func toDeduction(d StoredDeduction) Deduction {
	return Deduction{
		ID: d.ID, TaxpayerID: d.TaxpayerID, FiscalYear: d.FiscalYear, DeductionFields: d.DeductionFields,
		Excluded: exclusionOf(d.DeductionFields), CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

// resultOf is a write's answer for row, with the warnings of its fields.
func resultOf(row StoredDeduction, created bool) DeductionResult {
	d := toDeduction(row)
	var warnings []Warning
	if d.Excluded == ExclusionCashNotDeductible {
		warnings = []Warning{{Code: WarningCashNotDeductible, Field: FieldPaymentMethod}}
	}
	return DeductionResult{Deduction: d, Warnings: warnings, Created: created}
}

func mustParseDate(raw string) fiscal.Date {
	d, err := fiscal.ParseDate(raw)
	if err != nil {
		panic(err)
	}
	return d
}
