package domain

import (
	"errors"
	"maps"
	"slices"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// OptionType is the type of a regime option (contract RegimeOptionDef.type).
type OptionType string

// The contract's option types.
const (
	// OptionTypeBoolean takes a JSON boolean.
	OptionTypeBoolean OptionType = "boolean"
	// OptionTypeEnum takes one string from AllowedValues.
	OptionTypeEnum OptionType = "enum"
	// OptionTypeMultiEnum takes a list of distinct strings from
	// AllowedValues; a required one needs at least one.
	OptionTypeMultiEnum OptionType = "multi_enum"
	// OptionTypeRate takes a decimal string ("0.1234", at most six
	// decimals, no exponent or percent sign) within [Min, Max].
	OptionTypeRate OptionType = "rate"
)

var optionTypes = []OptionType{OptionTypeBoolean, OptionTypeEnum, OptionTypeMultiEnum, OptionTypeRate}

// OptionDef defines one option of a regime, as the catalog's option_defs
// row does. It is plain data: NewOptionSet checks it.
type OptionDef struct {
	Key      string
	Type     OptionType
	Required bool
	// AllowedValues lists the values of an enum or multi_enum, in catalog
	// order. Other types must leave it empty.
	AllowedValues []string
	// Min and Max are a rate's inclusive bounds. A rate must set them
	// (Min <= Max); other types ignore them.
	Min, Max money.Rate
	// Default is used when no value is given. The zero OptionValue means
	// no default.
	Default OptionValue
}

// valueKind is the JSON shape of an OptionValue.
type valueKind uint8

const (
	kindNone valueKind = iota
	kindBool
	kindText
	kindList
)

// OptionValue is one option value as the contract carries it: a boolean, a
// string or a list of strings. A rate travels as a decimal string. The zero
// value holds nothing and matches no option type.
type OptionValue struct {
	kind valueKind
	flag bool
	text string
	list []string
}

// BoolOption returns a boolean value.
func BoolOption(b bool) OptionValue { return OptionValue{kind: kindBool, flag: b} }

// StringOption returns a string value (an enum value or a decimal rate).
func StringOption(s string) OptionValue { return OptionValue{kind: kindText, text: s} }

// ListOption returns a list value; items are copied.
func ListOption(items []string) OptionValue {
	return OptionValue{kind: kindList, list: slices.Clone(items)}
}

// Bool returns the boolean and true when v holds one.
func (v OptionValue) Bool() (value, ok bool) { return v.flag, v.kind == kindBool }

// Text returns the string and true when v holds one.
func (v OptionValue) Text() (string, bool) { return v.text, v.kind == kindText }

// List returns a copy of the list and true when v holds one.
func (v OptionValue) List() ([]string, bool) { return slices.Clone(v.list), v.kind == kindList }

// IsZero reports whether v holds no value.
func (v OptionValue) IsZero() bool { return v.kind == kindNone }

// OptionSet is a regime's checked option definitions. The zero value
// defines no options.
type OptionSet struct {
	defs []OptionDef
	keys map[string]struct{}
}

// NewOptionSet checks defs and returns them as a set. A definition with an
// empty or duplicate key, an unknown type, allowed values that do not fit
// its type, inverted rate bounds or a default it would itself reject is
// ErrInvalidOptionDef, wrapped in an *OptionError naming the key. Defaults
// are stored in canonical form, and defs are copied.
func NewOptionSet(defs ...OptionDef) (OptionSet, error) {
	set := OptionSet{defs: make([]OptionDef, 0, len(defs)), keys: make(map[string]struct{}, len(defs))}
	for _, def := range defs {
		if _, duplicate := set.keys[def.Key]; duplicate {
			return OptionSet{}, &OptionError{Key: def.Key, Err: ErrInvalidOptionDef}
		}
		checked, err := checkDef(def)
		if err != nil {
			return OptionSet{}, &OptionError{Key: def.Key, Err: err}
		}
		set.defs = append(set.defs, checked)
		set.keys[def.Key] = struct{}{}
	}
	return set, nil
}

// Validate checks values against the set and returns them in canonical
// form, with defaults filled in for the options not given: rates are
// reformatted ("0.500" becomes "0.5") and multi_enum lists follow catalog
// order. It reports every offending option, joined with errors.Join, each
// an *OptionError: unknown keys first, sorted, then the definitions in
// catalog order. values is not modified or retained.
func (s OptionSet) Validate(values map[string]OptionValue) (map[string]OptionValue, error) {
	var errs []error
	for _, key := range slices.Sorted(maps.Keys(values)) {
		if _, known := s.keys[key]; !known {
			errs = append(errs, &OptionError{Key: key, Err: ErrOptionUnknown})
		}
	}
	options := make(map[string]OptionValue, len(s.defs))
	for _, def := range s.defs {
		value, given := values[def.Key]
		if !given {
			if !def.Default.IsZero() {
				options[def.Key] = def.Default // canonical since NewOptionSet
			} else if def.Required {
				errs = append(errs, &OptionError{Key: def.Key, Err: ErrOptionMissing})
			}
			continue
		}
		canonical, err := def.check(value)
		if err != nil {
			errs = append(errs, &OptionError{Key: def.Key, Err: err})
			continue
		}
		options[def.Key] = canonical
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return options, nil
}

// checkDef returns def with its slices copied and its default canonical, or
// ErrInvalidOptionDef.
func checkDef(def OptionDef) (OptionDef, error) {
	if def.Key == "" || !slices.Contains(optionTypes, def.Type) || !allowedValuesFit(def) {
		return OptionDef{}, ErrInvalidOptionDef
	}
	if def.Type == OptionTypeRate && def.Min > def.Max {
		return OptionDef{}, ErrInvalidOptionDef
	}
	def.AllowedValues = slices.Clone(def.AllowedValues)
	if def.Default.IsZero() {
		return def, nil
	}
	canonical, err := def.check(def.Default)
	if err != nil {
		return OptionDef{}, ErrInvalidOptionDef
	}
	def.Default = canonical
	return def, nil
}

// allowedValuesFit reports whether def's allowed values suit its type: a
// non-empty list of distinct, non-empty strings for the enums, none for the
// rest.
func allowedValuesFit(def OptionDef) bool {
	if def.Type != OptionTypeEnum && def.Type != OptionTypeMultiEnum {
		return len(def.AllowedValues) == 0
	}
	if len(def.AllowedValues) == 0 || slices.Contains(def.AllowedValues, "") {
		return false
	}
	sorted := slices.Sorted(slices.Values(def.AllowedValues))
	return len(slices.Compact(sorted)) == len(sorted)
}

// check validates one value of def and returns it in canonical form, else
// ErrOptionType or ErrOptionValue. def.Type is one of optionTypes.
func (def OptionDef) check(value OptionValue) (OptionValue, error) {
	switch def.Type {
	case OptionTypeBoolean:
		return expectKind(value, kindBool)
	case OptionTypeEnum:
		if _, err := expectKind(value, kindText); err != nil {
			return OptionValue{}, err
		}
		if !slices.Contains(def.AllowedValues, value.text) {
			return OptionValue{}, ErrOptionValue
		}
		return value, nil
	case OptionTypeMultiEnum:
		if _, err := expectKind(value, kindList); err != nil {
			return OptionValue{}, err
		}
		return def.checkList(value.list)
	default: // OptionTypeRate
		if _, err := expectKind(value, kindText); err != nil {
			return OptionValue{}, err
		}
		return def.checkRate(value.text)
	}
}

func expectKind(value OptionValue, kind valueKind) (OptionValue, error) {
	if value.kind != kind {
		return OptionValue{}, ErrOptionType
	}
	return value, nil
}

// checkList accepts distinct allowed items (at least one when required) and
// returns them in catalog order. O(len(items) * len(AllowedValues)); both
// are a handful of catalog entries.
func (def OptionDef) checkList(items []string) (OptionValue, error) {
	if len(items) == 0 && def.Required {
		return OptionValue{}, ErrOptionValue
	}
	chosen := make([]bool, len(def.AllowedValues))
	for _, item := range items {
		position := slices.Index(def.AllowedValues, item)
		if position < 0 || chosen[position] {
			return OptionValue{}, ErrOptionValue
		}
		chosen[position] = true
	}
	ordered := make([]string, 0, len(items))
	for position, allowed := range def.AllowedValues {
		if chosen[position] {
			ordered = append(ordered, allowed)
		}
	}
	return OptionValue{kind: kindList, list: ordered}, nil
}

// checkRate parses text as a money.Rate within [Min, Max] and returns its
// canonical string. The parse error is dropped: it echoes the input.
func (def OptionDef) checkRate(text string) (OptionValue, error) {
	rate, err := money.ParseRate(text)
	if err != nil || rate < def.Min || rate > def.Max {
		return OptionValue{}, ErrOptionValue
	}
	return StringOption(rate.String()), nil
}
