package infra

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// errBadOptionValue reports a stored option value the contract cannot carry
// (RegimeOptions: string, boolean or list of strings).
var errBadOptionValue = errors.New("taxpayers infra: option value is not a string, boolean or list of strings")

// encodeOptions renders options as the jsonb object taxpayer_regimes.options
// stores. The zero OptionValue is refused: options are validated (and
// defaulted) before they are written.
func encodeOptions(options map[string]domain.OptionValue) ([]byte, error) {
	out := make(map[string]any, len(options))
	for key, value := range options {
		encoded, ok := optionJSON(value)
		if !ok {
			return nil, fmt.Errorf("%w: key %q", errBadOptionValue, key)
		}
		out[key] = encoded
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("taxpayers infra: encode options: %w", err)
	}
	return raw, nil
}

func optionJSON(value domain.OptionValue) (any, bool) {
	if flag, ok := value.Bool(); ok {
		return flag, true
	}
	if text, ok := value.Text(); ok {
		return text, true
	}
	if list, ok := value.List(); ok {
		if list == nil {
			list = []string{}
		}
		return list, true
	}
	return nil, false
}

// decodeOptions parses a stored options object.
func decodeOptions(raw []byte) (map[string]domain.OptionValue, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("taxpayers infra: stored options are not a JSON object: %w", errors.Join(err, errBadOptionValue))
	}
	options := make(map[string]domain.OptionValue, len(fields))
	for key, value := range fields {
		decoded, err := decodeOptionValue(value)
		if err != nil {
			return nil, fmt.Errorf("taxpayers infra: stored option %q: %w", key, err)
		}
		options[key] = decoded
	}
	return options, nil
}

// decodeOptionValue reads one JSON string, boolean or list of strings,
// choosing by the first byte of the token.
func decodeOptionValue(raw json.RawMessage) (domain.OptionValue, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return domain.OptionValue{}, errBadOptionValue
	}
	switch raw[0] {
	case 't', 'f':
		var flag bool
		err := json.Unmarshal(raw, &flag)
		return domain.BoolOption(flag), err
	case '"':
		var text string
		err := json.Unmarshal(raw, &text)
		return domain.StringOption(text), err
	case '[':
		list := []string{}
		if err := json.Unmarshal(raw, &list); err != nil {
			return domain.OptionValue{}, errors.Join(errBadOptionValue, err)
		}
		return domain.ListOption(list), nil
	}
	return domain.OptionValue{}, errBadOptionValue
}

// optionDefJSON is one element of regimes.option_defs (contract
// RegimeOptionDef; label is display-only and not needed here).
type optionDefJSON struct {
	Key           string          `json:"key"`
	Type          string          `json:"type"`
	Required      bool            `json:"required"`
	AllowedValues []string        `json:"allowed_values"`
	Min           *string         `json:"min"`
	Max           *string         `json:"max"`
	Default       json.RawMessage `json:"default"`
}

// decodeOptionDefs parses regimes.option_defs into domain definitions. The
// catalog CHECK regime_option_defs_valid already guarantees the shape;
// domain.NewOptionSet checks the semantics.
func decodeOptionDefs(raw []byte) ([]domain.OptionDef, error) {
	var elements []optionDefJSON
	if err := json.Unmarshal(raw, &elements); err != nil {
		return nil, fmt.Errorf("taxpayers infra: decode option_defs: %w", err)
	}
	defs := make([]domain.OptionDef, 0, len(elements))
	for _, e := range elements {
		def := domain.OptionDef{Key: e.Key, Type: domain.OptionType(e.Type), Required: e.Required, AllowedValues: e.AllowedValues}
		var err error
		if def.Min, err = parseBound(e.Min); err != nil {
			return nil, fmt.Errorf("taxpayers infra: option %q min: %w", e.Key, err)
		}
		if def.Max, err = parseBound(e.Max); err != nil {
			return nil, fmt.Errorf("taxpayers infra: option %q max: %w", e.Key, err)
		}
		if def.Default, err = decodeDefault(e.Default); err != nil {
			return nil, fmt.Errorf("taxpayers infra: option %q default: %w", e.Key, err)
		}
		defs = append(defs, def)
	}
	return defs, nil
}

func parseBound(bound *string) (money.Rate, error) {
	if bound == nil {
		return 0, nil
	}
	return money.ParseRate(*bound)
}

// decodeDefault reads an absent or null default as none, else a string or
// a boolean.
func decodeDefault(raw json.RawMessage) (domain.OptionValue, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return domain.OptionValue{}, nil
	}
	value, err := decodeOptionValue(raw)
	if err != nil {
		return domain.OptionValue{}, err
	}
	if _, isList := value.List(); isList {
		return domain.OptionValue{}, errBadOptionValue
	}
	return value, nil
}
