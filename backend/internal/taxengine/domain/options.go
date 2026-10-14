package domain

import (
	"maps"
	"slices"
)

// OptionKey names a regime option (contract RegimeOptionDef.key), such as
// "deduction_mode" (606) or "definitive_option" (625). The taxpayers module
// validates options against the catalog; the engine only reads them.
type OptionKey string

// valueKind is the JSON shape of an OptionValue; zero holds nothing.
type valueKind uint8

const (
	kindBool valueKind = iota + 1
	kindText
	kindList
)

// OptionValue is one option value: a boolean, a string (an enum value) or a
// list of strings (a multi_enum). The zero value holds nothing.
type OptionValue struct {
	kind valueKind
	flag bool
	text string
	list []string
}

// BoolOption returns a boolean value.
func BoolOption(b bool) OptionValue { return OptionValue{kind: kindBool, flag: b} }

// TextOption returns a string value.
func TextOption(s string) OptionValue { return OptionValue{kind: kindText, text: s} }

// ListOption returns a list value; items are copied.
func ListOption(items []string) OptionValue {
	return OptionValue{kind: kindList, list: slices.Clone(items)}
}

// Options are a regime's option values for the period, immutable. The zero
// value holds none.
type Options struct {
	values map[OptionKey]OptionValue
}

// NewOptions copies values into Options.
func NewOptions(values map[OptionKey]OptionValue) Options {
	return Options{values: maps.Clone(values)}
}

func (o Options) value(key OptionKey, kind valueKind) (OptionValue, bool) {
	v := o.values[key]
	return v, v.kind == kind
}

// Bool returns a boolean option and whether it is set.
func (o Options) Bool(key OptionKey) (value, ok bool) {
	v, ok := o.value(key, kindBool)
	return v.flag, ok
}

// Text returns a string option and whether it is set.
func (o Options) Text(key OptionKey) (string, bool) {
	v, ok := o.value(key, kindText)
	return v.text, ok
}

// List returns a copy of a list option and whether it is set.
func (o Options) List(key OptionKey) ([]string, bool) {
	v, ok := o.value(key, kindList)
	return slices.Clone(v.list), ok
}
