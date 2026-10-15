package annual

import (
	"context"
	"errors"
	"log/slog"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// MaxSalaryStatementsPerYear bounds the constancias of a taxpayer and
// fiscal year (ADR-0009): create and a move to another year re-check it.
const MaxSalaryStatementsPerYear = 20

// maxEmployerLabel is the contract EmployerLabel maxLength, in characters.
const maxEmployerLabel = 60

// The audit events (ids only).
const (
	eventStatementCreated = "salary_statement.created"
	eventStatementUpdated = "salary_statement.updated"
	eventStatementDeleted = "salary_statement.deleted"
)

// Contract field names of the constancia figures.
const (
	FieldAccumulable = "accumulable_cents"
	FieldExempt      = "exempt_cents"
	FieldISRWithheld = "isr_withheld_cents"
)

// SalaryInput is a createSalaryStatement body after the schema; a nil
// EmployerLabel is absent.
type SalaryInput struct {
	ID                  string
	TaxpayerID          string
	FiscalYear          int
	EmployerLabel       *string
	Accumulable         money.Cents
	Exempt              money.Cents
	ISRWithheld         money.Cents
	Simultaneous        bool
	LeftBeforeYearEnd   bool
	PayerNotWithholding bool
}

// SalaryPatch is an updateSalaryStatement merge-patch: a nil pointer or an
// unset Optional is unchanged.
type SalaryPatch struct {
	FiscalYear          *int
	EmployerLabel       Optional[string]
	Accumulable         *money.Cents
	Exempt              *money.Cents
	ISRWithheld         *money.Cents
	Simultaneous        *bool
	LeftBeforeYearEnd   *bool
	PayerNotWithholding *bool
}

// SalaryResult is a write's answer. Created is false for an idempotent
// retry (200) or an edit.
type SalaryResult struct {
	Statement StoredStatement
	Warnings  []Warning
	Created   bool
}

// SalaryTotals sum a year's constancias (contract SalaryStatementTotals).
type SalaryTotals struct {
	Accumulable money.Cents
	Exempt      money.Cents
	ISRWithheld money.Cents
}

// SalaryList is listSalaryStatements.
type SalaryList struct {
	FiscalYear int
	Items      []StoredStatement
	Totals     SalaryTotals
}

// SalaryDeps are the ports the salary-constancia use cases need.
type SalaryDeps struct {
	Tx        tx.Manager
	Access    AccessPolicy
	Taxpayers TaxpayerStore
	Store     SalaryStore
	Clock     clock.Clock
	// NewID generates the id of a create without a client id; nil is
	// UUIDv7.
	NewID func() (string, error)
	// Logger receives the audit events; nil discards them.
	Logger *slog.Logger
}

// SalaryStatements are the salary-constancia use cases (owner decision 3,
// MODULE-annual §2.6; contract list/create/update/deleteSalaryStatement).
// The figures are the constancia's totals: nothing is recomputed.
type SalaryStatements struct {
	base
	deps SalaryDeps
}

// NewSalaryStatements returns the use cases over deps.
func NewSalaryStatements(deps SalaryDeps) *SalaryStatements {
	if deps.NewID == nil {
		deps.NewID = newUUIDv7
	}
	return &SalaryStatements{base: newBase(deps.Clock, deps.Logger), deps: deps}
}

// List returns the constancias of one taxpayer and year, with totals. A
// persona moral has none. Reads stay open to a read-only user and an
// archived taxpayer.
func (uc *SalaryStatements) List(ctx context.Context, userID, rawTaxpayerID string, fiscalYear int) (SalaryList, error) {
	tid, err := taxpayerID(rawTaxpayerID)
	if err != nil {
		return SalaryList{}, err
	}
	if err := uc.checkFutureYear(fiscalYear, FieldFiscalYear); err != nil {
		return SalaryList{}, err
	}
	out := SalaryList{FiscalYear: fiscalYear, Items: []StoredStatement{}}
	err = uc.deps.Tx.ReadOnly(ctx, userID, func(ctx context.Context) error {
		t, err := uc.deps.Taxpayers.Taxpayer(ctx, tid, LockNone)
		if err != nil {
			return err
		}
		if err := checkYearFloor(fiscalYear, FieldFiscalYear); err != nil || t.PersonType != fiscal.PersonFisica {
			return err
		}
		rows, err := uc.deps.Store.ListStatements(ctx, tid, fiscalYear)
		if err != nil {
			return err
		}
		out.Items = append(out.Items, rows...)
		out.Totals, err = totalsOf(rows)
		return err
	})
	if err != nil {
		return SalaryList{}, present("list salary statements", err)
	}
	return out, nil
}

func totalsOf(rows []StoredStatement) (SalaryTotals, error) {
	var t SalaryTotals
	for _, r := range rows {
		var errs [3]error
		t.Accumulable, errs[0] = t.Accumulable.Add(r.Accumulable)
		t.Exempt, errs[1] = t.Exempt.Add(r.Exempt)
		t.ISRWithheld, errs[2] = t.ISRWithheld.Add(r.ISRWithheld)
		if err := errors.Join(errs[:]...); err != nil {
			return SalaryTotals{}, err
		}
	}
	return t, nil
}

// Create records a constancia (createSalaryStatement). Order: ids
// (404/400) → idempotency (200/409) → read-only (402) → taxpayer (404),
// archived, person type, fiscal year floor, list bound (422) → field
// rules (400) → insert.
func (uc *SalaryStatements) Create(ctx context.Context, userID string, in SalaryInput) (SalaryResult, error) {
	tid, err := taxpayerID(in.TaxpayerID)
	if err != nil {
		return SalaryResult{}, err
	}
	id, err := clientID(in.ID, uc.deps.NewID)
	if err != nil {
		return SalaryResult{}, present("create salary statement", err)
	}
	fromClient := in.ID != ""
	in.ID, in.TaxpayerID = id, tid
	var res SalaryResult
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
		return SalaryResult{}, present("create salary statement", err)
	}
	return res, nil
}

func (uc *SalaryStatements) create(ctx context.Context, userID string, in SalaryInput, fromClient bool) (SalaryResult, error) {
	fields := normalizeSalary(in)
	if fromClient {
		if err := uc.deps.Store.LockStatementID(ctx, in.ID); err != nil {
			return SalaryResult{}, err
		}
	}
	existing, err := uc.deps.Store.FindStatement(ctx, in.ID, LockNone)
	switch {
	case err == nil:
		if existing.TaxpayerID != in.TaxpayerID || existing.SalaryFields != fields {
			return SalaryResult{}, ErrIDReused
		}
		return uc.result(ctx, existing, false)
	case !errors.Is(err, ErrStatementNotFound):
		return SalaryResult{}, err
	}
	if err := uc.deps.Access.EnsureWritable(ctx, userID); err != nil {
		return SalaryResult{}, err
	}
	if _, err := ensureWritable(ctx, uc.deps.Taxpayers, in.TaxpayerID, true); err != nil {
		return SalaryResult{}, err
	}
	if err := uc.checkYear(ctx, in.TaxpayerID, fields.FiscalYear); err != nil {
		return SalaryResult{}, err
	}
	if err := uc.checkFields(in); err != nil {
		return SalaryResult{}, err
	}
	if err := uc.deps.Store.InsertStatement(ctx, NewStatement{ID: in.ID, UserID: userID, TaxpayerID: in.TaxpayerID, SalaryFields: fields}); err != nil {
		return SalaryResult{}, err
	}
	stored, err := uc.deps.Store.FindStatement(ctx, in.ID, LockNone)
	if err != nil {
		return SalaryResult{}, err
	}
	uc.auditRow(ctx, eventStatementCreated, userID, stored)
	return uc.result(ctx, stored, true)
}

// Update edits a constancia (updateSalaryStatement, merge-patch), re-running
// every rule of a create; a move to another fiscal year applies that
// year's floor and bound.
func (uc *SalaryStatements) Update(ctx context.Context, userID, rawTaxpayerID, rawID string, p SalaryPatch) (SalaryResult, error) {
	tid, id, err := pathIDs(rawTaxpayerID, rawID, ErrStatementNotFound)
	if err != nil {
		return SalaryResult{}, err
	}
	var res SalaryResult
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.find(ctx, tid, id)
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
		fields := normalizeSalary(in)
		if fields.FiscalYear != stored.FiscalYear {
			if err := uc.checkYear(ctx, tid, fields.FiscalYear); err != nil {
				return err
			}
		}
		if err := uc.checkFields(in); err != nil {
			return err
		}
		if err := uc.deps.Store.UpdateStatement(ctx, id, fields); err != nil {
			return err
		}
		updated, err := uc.deps.Store.FindStatement(ctx, id, LockNone)
		if err != nil {
			return err
		}
		uc.auditRow(ctx, eventStatementUpdated, userID, updated)
		res, err = uc.result(ctx, updated, false)
		return err
	})
	if err != nil {
		return SalaryResult{}, present("update salary statement", err)
	}
	return res, nil
}

// Delete deletes a constancia for good (deleteSalaryStatement). Open to a
// read-only user; an archived taxpayer refuses it; an unknown one is 404.
func (uc *SalaryStatements) Delete(ctx context.Context, userID, rawTaxpayerID, rawID string) error {
	tid, id, err := pathIDs(rawTaxpayerID, rawID, ErrStatementNotFound)
	if err != nil {
		return err
	}
	err = uc.deps.Tx.ReadWrite(ctx, userID, func(ctx context.Context) error {
		stored, err := uc.find(ctx, tid, id)
		if err != nil {
			return err
		}
		if _, err := ensureWritable(ctx, uc.deps.Taxpayers, tid, false); err != nil {
			return err
		}
		if err := uc.deps.Store.DeleteStatement(ctx, id); err != nil {
			return err
		}
		uc.auditRow(ctx, eventStatementDeleted, userID, stored)
		return nil
	})
	return present("delete salary statement", err)
}

// find locks the caller's constancia of the taxpayer; any other is 404.
func (uc *SalaryStatements) find(ctx context.Context, taxpayerID, id string) (StoredStatement, error) {
	stored, err := uc.deps.Store.FindStatement(ctx, id, LockWrite)
	if err != nil {
		return StoredStatement{}, err
	}
	if stored.TaxpayerID != taxpayerID {
		return StoredStatement{}, ErrStatementNotFound
	}
	return stored, nil
}

// checkYear applies the fiscal-year floor (422) and the year's bound.
func (uc *SalaryStatements) checkYear(ctx context.Context, taxpayerID string, fiscalYear int) error {
	if err := checkYearFloor(fiscalYear, FieldFiscalYear); err != nil {
		return err
	}
	if err := uc.deps.Taxpayers.LockList(ctx, taxpayerID, ListSalaryStatements, fiscalYear); err != nil {
		return err
	}
	n, err := uc.deps.Store.CountStatements(ctx, taxpayerID, fiscalYear)
	if err != nil {
		return err
	}
	if n >= MaxSalaryStatementsPerYear {
		return limitError(msgLimitStatements)
	}
	return nil
}

// checkFields applies the field rules (400): a future year, the label and
// the figures (the schema's minimum, re-checked).
func (uc *SalaryStatements) checkFields(in SalaryInput) error {
	if err := uc.checkFutureYear(in.FiscalYear, FieldFiscalYear); err != nil {
		return err
	}
	var fields []apperr.FieldError
	if in.EmployerLabel != nil && !validLabel(normalizeLabel(*in.EmployerLabel), maxEmployerLabel) {
		fields = append(fields, apperr.FieldError{Field: FieldEmployerLabel, Code: DetailInvalid, Message: msgInvalidValue})
	}
	for _, f := range []struct {
		name  string
		value money.Cents
	}{{FieldAccumulable, in.Accumulable}, {FieldExempt, in.Exempt}, {FieldISRWithheld, in.ISRWithheld}} {
		if f.value < 0 {
			fields = append(fields, apperr.FieldError{Field: f.name, Code: DetailInvalid, Message: msgInvalidValue})
		}
	}
	if len(fields) > 0 {
		return fieldErrors(fields)
	}
	return nil
}

// result is a write's answer, with the salary_regime_not_held warning when
// the taxpayer holds no 605 in the constancia's year (OQ-A2).
func (uc *SalaryStatements) result(ctx context.Context, row StoredStatement, created bool) (SalaryResult, error) {
	regimes, err := uc.deps.Taxpayers.Regimes(ctx, row.TaxpayerID)
	if err != nil {
		return SalaryResult{}, err
	}
	res := SalaryResult{Statement: row, Created: created}
	if !holdsIn(regimes, adomain.RegimeSalary, row.FiscalYear) {
		res.Warnings = []Warning{{Code: WarningSalaryRegimeNotHeld, Field: FieldFiscalYear}}
	}
	return res, nil
}

func (uc *SalaryStatements) auditRow(ctx context.Context, event, userID string, s StoredStatement) {
	uc.audit(ctx, event, slog.String("user_id", userID), slog.String("taxpayer_id", s.TaxpayerID),
		slog.String("statement_id", s.ID))
}

// holdsIn reports a regime of code held in some month of fiscalYear.
func holdsIn(regimes []HeldRegime, code fiscal.RegimeCode, fiscalYear int) bool {
	for _, r := range regimes {
		if r.Code == code && r.ActiveIn(fiscalYear) {
			return true
		}
	}
	return false
}

// apply returns the stored row with p applied.
func (p SalaryPatch) apply(stored StoredStatement) SalaryInput {
	in := SalaryInput{
		ID: stored.ID, TaxpayerID: stored.TaxpayerID, FiscalYear: stored.FiscalYear,
		EmployerLabel: nonZero(stored.EmployerLabel), Accumulable: stored.Accumulable, Exempt: stored.Exempt,
		ISRWithheld: stored.ISRWithheld, Simultaneous: stored.Simultaneous, LeftBeforeYearEnd: stored.LeftBeforeYearEnd,
		PayerNotWithholding: stored.PayerNotWithholding,
	}
	setIf(&in.FiscalYear, p.FiscalYear)
	in.EmployerLabel = p.EmployerLabel.apply(in.EmployerLabel)
	setIf(&in.Accumulable, p.Accumulable)
	setIf(&in.Exempt, p.Exempt)
	setIf(&in.ISRWithheld, p.ISRWithheld)
	setIf(&in.Simultaneous, p.Simultaneous)
	setIf(&in.LeftBeforeYearEnd, p.LeftBeforeYearEnd)
	setIf(&in.PayerNotWithholding, p.PayerNotWithholding)
	return in
}

// normalizeSalary is the input as stored: the label NFC (ADR-0004).
func normalizeSalary(in SalaryInput) SalaryFields {
	return SalaryFields{
		FiscalYear: in.FiscalYear, EmployerLabel: normalizeLabel(valueOr(in.EmployerLabel)), Accumulable: in.Accumulable,
		Exempt: in.Exempt, ISRWithheld: in.ISRWithheld, Simultaneous: in.Simultaneous,
		LeftBeforeYearEnd: in.LeftBeforeYearEnd, PayerNotWithholding: in.PayerNotWithholding,
	}
}
