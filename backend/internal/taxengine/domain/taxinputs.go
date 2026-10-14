package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// InputKey names a tax input (contract TaxInputDef.key). Which keys a
// regime takes is catalog data (regimes.required_inputs); the constants name
// the ones a strategy reads.
type InputKey string

// The tax inputs of MODULE-tax-engine §2.8.
const (
	InputPriorProvisionalPayments InputKey = "prior_provisional_payments"
	InputYTDBeforeAppIncome       InputKey = "ytd_before_app_income"
	InputYTDBeforeAppDeductions   InputKey = "ytd_before_app_deductions"
	InputPTUPaid                  InputKey = "ptu_paid"
	InputPriorYearLosses          InputKey = "prior_year_losses"
	InputInterestWithholding      InputKey = "interest_withholding"
	InputCoeficienteUtilidad      InputKey = "coeficiente_utilidad"
	InputFirstFiscalYear          InputKey = "first_fiscal_year"
	InputPriorYearIncome          InputKey = "prior_year_income"
)

// InputKind is the kind of a tax input (contract TaxInputDef.kind).
type InputKind string

// The contract's kinds (ADR-0002 Amendment 1 adds money_monthly and boolean).
const (
	InputMoney        InputKind = "money"
	InputMoneyMonthly InputKind = "money_monthly"
	InputRate         InputKind = "rate"
	InputBoolean      InputKind = "boolean"
)

// Valid reports a contract kind.
func (k InputKind) Valid() bool {
	return slices.Contains([]InputKind{InputMoney, InputMoneyMonthly, InputRate, InputBoolean}, k)
}

// TaxInputValue is one stored tax input: the field its Kind names is the
// value, and Month (1-12) is set only for money_monthly.
type TaxInputValue struct {
	Key   InputKey
	Kind  InputKind
	Month time.Month
	Money money.Cents
	Rate  money.Rate
	Bool  bool
}

// inputSlot identifies a value: its key, and its month for money_monthly.
type inputSlot struct {
	key   InputKey
	month time.Month
}

// TaxInputs is the validated, immutable set of a regime's tax inputs for
// one fiscal year. The zero value holds nothing. A key that is absent is a
// missing input: strategies compute with zero for it (ADR-0002 §2), and
// RequiredInputs decides whether that makes the estimate incomplete.
type TaxInputs struct {
	values map[inputSlot]TaxInputValue
	kinds  map[InputKey]InputKind
}

// NewTaxInputs validates values: a key and a known kind, money >= 0, rates
// in [0, 1], a month 1-12 on money_monthly values and none elsewhere, and
// no key (and month) twice or under two kinds.
func NewTaxInputs(values ...TaxInputValue) (TaxInputs, error) {
	in := TaxInputs{values: make(map[inputSlot]TaxInputValue, len(values)), kinds: map[InputKey]InputKind{}}
	for _, v := range values {
		if err := checkInput(v); err != nil {
			return TaxInputs{}, err
		}
		slot := inputSlot{key: v.Key, month: v.Month}
		if kind, seen := in.kinds[v.Key]; seen && (kind != v.Kind || in.has(slot)) {
			return TaxInputs{}, fmt.Errorf("%s %s: %w", v.Key, v.Month, ErrDuplicateTaxInput)
		}
		in.kinds[v.Key] = v.Kind
		in.values[slot] = v
	}
	return in, nil
}

func (in TaxInputs) has(slot inputSlot) bool {
	_, ok := in.values[slot]
	return ok
}

func checkInput(v TaxInputValue) error {
	switch {
	case strings.TrimSpace(string(v.Key)) == "" || !v.Kind.Valid():
		return fmt.Errorf("key %q kind %q: %w", v.Key, v.Kind, ErrInvalidTaxInput)
	case v.Money < 0:
		return fmt.Errorf("%s: negative amount: %w", v.Key, ErrInvalidTaxInput)
	case v.Rate < 0 || v.Rate > money.RateScale:
		return fmt.Errorf("%s: rate %s outside [0, 1]: %w", v.Key, v.Rate, ErrInvalidTaxInput)
	case v.Kind == InputMoneyMonthly && !validMonth(v.Month),
		v.Kind != InputMoneyMonthly && v.Month != 0:
		return fmt.Errorf("%s: month %d on a %s input: %w", v.Key, v.Month, v.Kind, ErrInvalidMonth)
	}
	return nil
}

func validMonth(m time.Month) bool { return m >= time.January && m <= time.December }

func (in TaxInputs) lookup(key InputKey, kind InputKind, month time.Month) (TaxInputValue, bool) {
	v, ok := in.values[inputSlot{key: key, month: month}]
	return v, ok && v.Kind == kind
}

// Money returns a money input and whether it is present.
func (in TaxInputs) Money(key InputKey) (money.Cents, bool) {
	v, ok := in.lookup(key, InputMoney, 0)
	return v.Money, ok
}

// MonthlyMoney returns month's value of a money_monthly input and whether it
// is present.
func (in TaxInputs) MonthlyMoney(key InputKey, month time.Month) (money.Cents, bool) {
	v, ok := in.lookup(key, InputMoneyMonthly, month)
	return v.Money, ok
}

// Rate returns a rate input and whether it is present.
func (in TaxInputs) Rate(key InputKey) (money.Rate, bool) {
	v, ok := in.lookup(key, InputRate, 0)
	return v.Rate, ok
}

// Bool returns a boolean input and whether it is present.
func (in TaxInputs) Bool(key InputKey) (value, ok bool) {
	v, ok := in.lookup(key, InputBoolean, 0)
	return v.Bool, ok
}

// SumBefore returns the sum of a money_monthly input over the months before
// month (January..month-1), a missing month counting as zero: the prior
// provisional payments of period month (Amendment 1). It fails with
// ErrInvalidMonth or money.ErrOverflow.
func (in TaxInputs) SumBefore(key InputKey, month time.Month) (money.Cents, error) {
	if !validMonth(month) {
		return 0, fmt.Errorf("%s before month %d: %w", key, month, ErrInvalidMonth)
	}
	var sum money.Cents
	for m := time.January; m < month; m++ {
		v, _ := in.MonthlyMoney(key, m) // absent: zero
		var err error
		if sum, err = sum.Add(v); err != nil {
			return 0, fmt.Errorf("%s before %s: %w", key, month, err)
		}
	}
	return sum, nil
}
