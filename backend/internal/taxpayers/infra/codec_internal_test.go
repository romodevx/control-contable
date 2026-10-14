package infra

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

func TestOptionsCodec_RoundTrip(t *testing.T) {
	options := map[string]domain.OptionValue{
		"deduction_mode": domain.StringOption("blind_35"),
		"iva":            domain.BoolOption(false),
		"activities":     domain.ListOption([]string{"lodging", "goods_services"}),
		"none":           domain.ListOption([]string{}),
	}

	encoded, err := encodeOptions(options)
	require.NoError(t, err)
	assert.JSONEq(t, `{"deduction_mode":"blind_35","iva":false,"activities":["lodging","goods_services"],"none":[]}`, string(encoded))
	decoded, err := decodeOptions(encoded)
	require.NoError(t, err)
	assert.Equal(t, options, decoded)

	empty, err := encodeOptions(nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{}`, string(empty))
}

func TestOptionsCodec_RefusesWhatTheContractCannotCarry(t *testing.T) {
	for name, raw := range map[string]string{
		"number":        `{"a": 1}`,
		"null":          `{"a": null}`,
		"object":        `{"a": {}}`,
		"list of bools": `{"a": [true]}`,
		"not an object": `[]`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := decodeOptions([]byte(raw))
			require.Error(t, err)
		})
	}
	_, err := encodeOptions(map[string]domain.OptionValue{"a": {}})
	require.Error(t, err, "the zero OptionValue is never stored")
}

func TestOptionDefs_ParseTheCatalogShape(t *testing.T) {
	raw := `[
	  {"key": "deduction_mode", "label": "x", "type": "enum", "allowed_values": ["blind_35", "actual"], "default": "blind_35", "required": true},
	  {"key": "activities", "label": "x", "type": "multi_enum", "allowed_values": ["a", "b"], "default": null, "required": true},
	  {"key": "iva", "label": "x", "type": "boolean", "default": true, "required": false},
	  {"key": "coef", "label": "x", "type": "rate", "min": "0", "max": "0.5", "default": "0.25", "required": false},
	  {"key": "plain", "label": "x", "type": "boolean", "required": false}
	]`

	defs, err := decodeOptionDefs([]byte(raw))

	require.NoError(t, err)
	half, err := money.ParseRate("0.5")
	require.NoError(t, err)
	assert.Equal(t, []domain.OptionDef{
		{Key: "deduction_mode", Type: domain.OptionTypeEnum, Required: true, AllowedValues: []string{"blind_35", "actual"}, Default: domain.StringOption("blind_35")},
		{Key: "activities", Type: domain.OptionTypeMultiEnum, Required: true, AllowedValues: []string{"a", "b"}},
		{Key: "iva", Type: domain.OptionTypeBoolean, Default: domain.BoolOption(true)},
		{Key: "coef", Type: domain.OptionTypeRate, Max: half, Default: domain.StringOption("0.25")},
		{Key: "plain", Type: domain.OptionTypeBoolean},
	}, defs)
	_, err = domain.NewOptionSet(defs...)
	require.NoError(t, err)

	_, err = decodeOptionDefs([]byte(`[{"key": "r", "type": "rate", "min": "x", "max": "1", "required": false}]`))
	require.Error(t, err)
	_, err = decodeOptionDefs([]byte(`[{"key": "d", "type": "boolean", "default": 3, "required": false}]`))
	require.Error(t, err)
}
