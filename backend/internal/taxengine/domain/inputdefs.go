package domain

import (
	"strconv"
	"time"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// InputDef is one regimes.required_inputs element (contract TaxInputDef):
// a per-fiscal-year input a regime's estimate may need. Which keys a regime
// takes is catalog data; Required says the key can be required in some
// period of the year (MissingInputs decides per period).
type InputDef struct {
	Key      InputKey
	Label    string
	Kind     InputKind
	Required bool
}

// InputDefPositions maps each key to its position in defs: the order of
// the stored values in a TaxInputs response (by key position, then month).
func InputDefPositions(defs []InputDef) map[InputKey]int {
	position := make(map[InputKey]int, len(defs))
	for i, d := range defs {
		position[d.Key] = i
	}
	return position
}

// SubmittedInput is one value as the client sends it (contract
// TaxInputValue): a nil field is absent or null.
type SubmittedInput struct {
	Key   InputKey
	Month *int
	Cents *int64
	// Rate is the decimal string as sent (DecimalRate).
	Rate *string
	Bool *bool
}

// InputProperty is the TaxInputValue property a violation names.
type InputProperty string

// The TaxInputValue properties.
const (
	PropertyKey        InputProperty = "key"
	PropertyMonth      InputProperty = "month"
	PropertyValueCents InputProperty = "value_cents"
	PropertyValueRate  InputProperty = "value_rate"
	PropertyValueBool  InputProperty = "value_bool"
)

// InputViolation is one submitted value that does not fit its definition:
// the value's index, the property at fault and why (ErrUnknownInput,
// ErrInvalidTaxInput, ErrDuplicateTaxInput or ErrInputOutOfRange).
type InputViolation struct {
	Index    int
	Property InputProperty
	Err      error
}

// Field is the violation's dotted field under prefix: "values.2.month".
func (v InputViolation) Field(prefix string) string {
	return prefix + "." + strconv.Itoa(v.Index) + "." + string(v.Property)
}

func (v InputViolation) Error() string { return v.Field("values") + ": " + v.Err.Error() }

func (v InputViolation) Unwrap() error { return v.Err }

// valueProperty is the value field each kind uses.
var valueProperty = map[InputKind]InputProperty{
	InputMoney:        PropertyValueCents,
	InputMoneyMonthly: PropertyValueCents,
	InputRate:         PropertyValueRate,
	InputBoolean:      PropertyValueBool,
}

// CheckSubmittedInputs validates a replacement set of values against the
// regime's definitions (PUT tax-inputs, MODULE-tax-engine §2.8 and §5): the
// key is defined; exactly the value field of its kind is set; a month 1-12
// on money_monthly values and none elsewhere; money in [0,
// fiscal.MaxCents]; a rate in [0, 1]; no key (and month) twice. It returns
// the values in submission order, each with its catalog kind, or nil and
// every violation, ordered by index. O(n) for n values.
func CheckSubmittedInputs(defs []InputDef, submitted []SubmittedInput) ([]TaxInputValue, []InputViolation) {
	kinds := make(map[InputKey]InputKind, len(defs))
	for _, d := range defs {
		kinds[d.Key] = d.Kind
	}
	values := make([]TaxInputValue, 0, len(submitted))
	seen := make(map[inputSlot]bool, len(submitted))
	var violations []InputViolation
	for i, s := range submitted {
		kind, defined := kinds[s.Key]
		if !defined {
			violations = append(violations, InputViolation{Index: i, Property: PropertyKey, Err: ErrUnknownInput})
			continue
		}
		v, refused := checkSubmitted(i, kind, s)
		if len(refused) > 0 {
			violations = append(violations, refused...)
			continue
		}
		slot := inputSlot{key: v.Key, month: v.Month}
		if seen[slot] {
			violations = append(violations, InputViolation{Index: i, Property: duplicateProperty(kind), Err: ErrDuplicateTaxInput})
			continue
		}
		seen[slot] = true
		values = append(values, v)
	}
	if len(violations) > 0 {
		return nil, violations
	}
	return values, nil
}

func duplicateProperty(kind InputKind) InputProperty {
	if kind == InputMoneyMonthly {
		return PropertyMonth
	}
	return PropertyKey
}

// checkSubmitted checks one defined value of kind: its fields, then its
// month, then the value itself.
func checkSubmitted(i int, kind InputKind, s SubmittedInput) (TaxInputValue, []InputViolation) {
	refuse := func(p InputProperty, err error) InputViolation {
		return InputViolation{Index: i, Property: p, Err: err}
	}
	var violations []InputViolation
	for _, p := range fieldsOf(s) {
		if p != valueProperty[kind] {
			violations = append(violations, refuse(p, ErrInvalidTaxInput))
		}
	}
	if !hasField(s, valueProperty[kind]) {
		violations = append(violations, refuse(valueProperty[kind], ErrInvalidTaxInput))
	}
	month, err := submittedMonth(kind, s.Month)
	if err != nil {
		violations = append(violations, refuse(PropertyMonth, err))
	}
	if len(violations) > 0 {
		return TaxInputValue{}, violations
	}
	v := TaxInputValue{Key: s.Key, Kind: kind, Month: month}
	switch kind {
	case InputMoney, InputMoneyMonthly:
		v.Money, err = submittedMoney(*s.Cents)
	case InputRate:
		v.Rate, err = submittedRate(*s.Rate)
	case InputBoolean:
		v.Bool = *s.Bool
	}
	if err != nil {
		return TaxInputValue{}, []InputViolation{refuse(valueProperty[kind], err)}
	}
	return v, nil
}

// fieldsOf lists the value fields s sets, in contract order.
func fieldsOf(s SubmittedInput) []InputProperty {
	var set []InputProperty
	for _, p := range []InputProperty{PropertyValueCents, PropertyValueRate, PropertyValueBool} {
		if hasField(s, p) {
			set = append(set, p)
		}
	}
	return set
}

func hasField(s SubmittedInput, p InputProperty) bool {
	switch p {
	case PropertyValueCents:
		return s.Cents != nil
	case PropertyValueRate:
		return s.Rate != nil
	case PropertyValueBool:
		return s.Bool != nil
	}
	return false
}

// submittedMonth is the month of a money_monthly value, and none elsewhere.
func submittedMonth(kind InputKind, month *int) (time.Month, error) {
	switch {
	case kind != InputMoneyMonthly && month == nil:
		return 0, nil
	case kind != InputMoneyMonthly, month == nil:
		return 0, ErrInvalidTaxInput
	case !validMonth(time.Month(*month)):
		return 0, ErrInputOutOfRange
	}
	return time.Month(*month), nil
}

func submittedMoney(c int64) (money.Cents, error) {
	if c < 0 || money.Cents(c) > fiscal.MaxCents {
		return 0, ErrInputOutOfRange
	}
	return money.Cents(c), nil
}

func submittedRate(raw string) (money.Rate, error) {
	r, err := money.ParseRate(raw)
	if err != nil {
		return 0, ErrInvalidTaxInput
	}
	if r < 0 || r > money.RateScale {
		return 0, ErrInputOutOfRange
	}
	return r, nil
}
