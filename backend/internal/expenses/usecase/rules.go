package usecase

import (
	"context"
	"errors"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// WarningProvisionalParameters is the use case's warning on paid_on: the
// payment date's fiscal year has no published tax parameters yet, so the
// IVA and cash rules used the previous year's (ADR-0002 §3).
const WarningProvisionalParameters domain.WarningCode = "provisional_parameters"

// subject is one expense the write rules judge: its taxpayer, regime and
// the domain draft, plus the id a duplicate-folio lookup must skip.
type subject struct {
	userID     string
	taxpayerID string
	expenseID  string // "" on create: nothing to skip
	profile    TaxpayerProfile
	regime     fiscal.RegimeCode
	draft      domain.Draft
}

// verdict is what the write rules decided for a subject.
type verdict struct {
	assessment domain.Assessment
	warnings   []domain.Warning
}

// judge runs every write rule on s, in the order of MODULE-incomes §2.10:
// the regime window (CheckRegime, which also refuses an archived taxpayer),
// the category rule of the regime and person type, the domain rules on the
// dated parameters, then the warnings that need stored data. Validation
// failures (400) of every field are reported together and win over the
// business rules (422).
func (u *Expenses) judge(ctx context.Context, s subject) (verdict, error) {
	today, err := u.today()
	if err != nil {
		return verdict{}, err
	}
	var extra fiscal.Violations
	if s.regime == "" {
		extra.Add(fieldRegimeCode, fiscal.ErrRequired)
	}
	dated, err := u.checkDated(ctx, s, today, &extra)
	if err != nil {
		return verdict{}, err
	}
	rule, err := u.ruleFor(ctx, s, &extra)
	if err != nil {
		return verdict{}, err
	}
	assessment, err := domain.Assess(s.draft, u.payerOf(s), rule, today, u.deps.Parameters)
	if len(extra) > 0 {
		return verdict{}, mergeValidation(extra, err)
	}
	if err != nil {
		return verdict{}, assessError(err)
	}
	warnings, err := u.storedWarnings(ctx, s, assessment.Warnings, dated)
	if err != nil {
		return verdict{}, err
	}
	return verdict{assessment: assessment, warnings: warnings}, nil
}

// checkDated runs the checks that need a paid_on the domain accepts: the
// regime window on its period and the parameters in force that day. A
// date the domain refuses is left to Assess (400 out_of_range).
func (u *Expenses) checkDated(ctx context.Context, s subject, today fiscal.Date, extra *fiscal.Violations) (DatedParameters, error) {
	on := s.draft.PaidOn
	if on.IsZero() || on.Before(fiscal.MinRecordDate) || on.After(today) {
		return DatedParameters{}, nil
	}
	if s.regime != "" {
		if err := u.deps.Regimes.CheckRegime(ctx, s.userID, s.taxpayerID, string(s.regime), on.Period()); err != nil {
			return DatedParameters{}, err
		}
	}
	dated, err := u.deps.Parameters.On(on)
	if errors.Is(err, ErrParametersUnavailable) {
		extra.Add(domain.FieldPaidOn, fiscal.ErrOutOfRange)
		return DatedParameters{}, nil
	}
	return dated, err
}

// ruleFor reads the category's rule for the subject's regime and person
// type. A missing or unknown category is recorded in extra and judged
// under a placeholder that never lowers ISR, so the draft's other fields
// are still validated.
func (u *Expenses) ruleFor(ctx context.Context, s subject, extra *fiscal.Violations) (domain.CategoryRule, error) {
	placeholder := domain.CategoryRule{Category: s.draft.Category, Deductibility: domain.NotDeductible}
	if s.draft.Category == "" || s.regime == "" {
		return placeholder, nil
	}
	entry, err := u.deps.Categories.Rule(ctx, s.regime, s.profile.PersonType, s.draft.Category)
	if errors.Is(err, ErrCategoryNotFound) {
		extra.Add(domain.FieldCategoryCode, domain.ErrUnknownCategory)
		return placeholder, nil
	}
	return entry.Rule, err
}

// payerOf is the domain's payer: the taxpayer under the regime window that
// holds paid_on (its options), or the bare regime when none does yet.
func (u *Expenses) payerOf(s subject) domain.Payer {
	if hold, ok := s.profile.holding(s.regime, s.draft.PaidOn.Period()); ok {
		return s.profile.payer(hold)
	}
	return domain.Payer{PersonType: s.profile.PersonType, Regime: s.regime}
}

// storedWarnings adds the warnings that need more than the draft: a folio
// fiscal another live expense of the taxpayer carries, and provisional
// parameters.
func (u *Expenses) storedWarnings(ctx context.Context, s subject, warnings []domain.Warning, dated DatedParameters) ([]domain.Warning, error) {
	out := append([]domain.Warning{}, warnings...)
	if !s.draft.CFDIUUID.IsZero() {
		inUse, err := u.deps.Expenses.CFDIInUse(ctx, s.taxpayerID, s.expenseID, s.draft.CFDIUUID)
		if err != nil {
			return nil, err
		}
		if inUse {
			out = append(out, domain.Warning{Code: domain.WarningDuplicateCFDIUUID, Field: domain.FieldCFDIUUID})
		}
	}
	if dated.Provisional {
		out = append(out, domain.Warning{Code: WarningProvisionalParameters, Field: domain.FieldPaidOn})
	}
	return out, nil
}

// mergeValidation joins the use case's own validation failures with the
// domain's: a domain validation pass adds its fields; a business-rule pass
// (or a parameter fault caused by the same bad input) yields to them.
func mergeValidation(extra fiscal.Violations, err error) error {
	if vs, ok := errors.AsType[fiscal.Violations](err); ok && vs.Class() == fiscal.ClassValidation {
		extra = append(extra, vs...)
	}
	return violationsError(extra)
}

// assessError presents a domain failure: violations as their client error,
// anything else (a parameter fault) as a server error.
func assessError(err error) error {
	if vs, ok := errors.AsType[fiscal.Violations](err); ok {
		return violationsError(vs)
	}
	return err
}

// stored applies an assessed draft to e.
func stored(e StoredExpense, regime fiscal.RegimeCode, d domain.Draft, description string, a domain.Assessment) StoredExpense {
	e.Regime, e.Category, e.Amount, e.IVARate, e.IVA = regime, d.Category, d.Amount, d.IVARate, d.IVA
	e.PaymentMethod, e.PaidOn, e.HasCFDI, e.CFDIUUID = d.PaymentMethod, d.PaidOn, d.HasCFDI, d.CFDIUUID
	e.Withheld, e.Description = d.Withheld, description
	e.Deductible, e.Overridden = a.Deductible, a.Source == domain.SourceUser
	e.CashOverLimit, e.CreditableIVA = a.CashOverLimit, a.CreditableIVA
	return e
}

// draftOf is the stored expense as a domain draft; its override is the
// stored claim.
func draftOf(e StoredExpense) domain.Draft {
	return domain.Draft{
		Category: e.Category, Amount: e.Amount, IVARate: e.IVARate, IVA: e.IVA, PaymentMethod: e.PaymentMethod,
		PaidOn: e.PaidOn, HasCFDI: e.HasCFDI, CFDIUUID: e.CFDIUUID, Deductible: overrideOf(e), Withheld: e.Withheld,
	}
}

func overrideOf(e StoredExpense) domain.Override {
	switch {
	case !e.Overridden:
		return domain.UseDefault
	case e.Deductible:
		return domain.ClaimDeductible
	}
	return domain.ClaimNotDeductible
}
