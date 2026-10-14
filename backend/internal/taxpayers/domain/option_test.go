package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
)

// The definitions below are shaped like the research catalog (606 deduction
// mode, 625 platform types and definitive withholding, 601 coeficiente) but
// are test data: the real catalog is seeded by migration (M2-02).
func deductionModeDef() domain.OptionDef {
	return domain.OptionDef{
		Key: "deduction_mode", Type: domain.OptionTypeEnum, Required: true,
		AllowedValues: []string{"blind_35", "actual"}, Default: domain.StringOption("blind_35"),
	}
}

func platformTypesDef() domain.OptionDef {
	return domain.OptionDef{
		Key: "platform_types", Type: domain.OptionTypeMultiEnum, Required: true,
		AllowedValues: []string{"goods", "services", "lodging", "transport"},
	}
}

func definitiveDef() domain.OptionDef {
	return domain.OptionDef{Key: "definitive_withholding", Type: domain.OptionTypeBoolean, Default: domain.BoolOption(false)}
}

func coeficienteDef() domain.OptionDef {
	return domain.OptionDef{Key: "coeficiente", Type: domain.OptionTypeRate, Min: 0, Max: money.RateScale}
}

func catalogSet(t *testing.T) domain.OptionSet {
	t.Helper()
	set, err := domain.NewOptionSet(deductionModeDef(), platformTypesDef(), definitiveDef(), coeficienteDef())
	require.NoError(t, err)
	return set
}

func TestOptionSetValidateAccepts(t *testing.T) {
	tests := []struct {
		name   string
		values map[string]domain.OptionValue
		want   map[string]domain.OptionValue
	}{
		{
			name:   "only required without default given; defaults applied",
			values: map[string]domain.OptionValue{"platform_types": domain.ListOption([]string{"goods"})},
			want: map[string]domain.OptionValue{
				"deduction_mode":         domain.StringOption("blind_35"),
				"platform_types":         domain.ListOption([]string{"goods"}),
				"definitive_withholding": domain.BoolOption(false),
			},
		},
		{
			name: "every option given",
			values: map[string]domain.OptionValue{
				"deduction_mode":         domain.StringOption("actual"),
				"platform_types":         domain.ListOption([]string{"services"}),
				"definitive_withholding": domain.BoolOption(true),
				"coeficiente":            domain.StringOption("0.1234"),
			},
			want: map[string]domain.OptionValue{
				"deduction_mode":         domain.StringOption("actual"),
				"platform_types":         domain.ListOption([]string{"services"}),
				"definitive_withholding": domain.BoolOption(true),
				"coeficiente":            domain.StringOption("0.1234"),
			},
		},
		{
			name: "multi enum kept in catalog order",
			values: map[string]domain.OptionValue{
				"platform_types": domain.ListOption([]string{"transport", "goods", "lodging"}),
			},
			want: map[string]domain.OptionValue{
				"deduction_mode":         domain.StringOption("blind_35"),
				"platform_types":         domain.ListOption([]string{"goods", "lodging", "transport"}),
				"definitive_withholding": domain.BoolOption(false),
			},
		},
		{
			name: "rate canonicalized and bounds inclusive (lower)",
			values: map[string]domain.OptionValue{
				"platform_types": domain.ListOption([]string{"goods"}),
				"coeficiente":    domain.StringOption("0.000000"),
			},
			want: map[string]domain.OptionValue{
				"deduction_mode":         domain.StringOption("blind_35"),
				"platform_types":         domain.ListOption([]string{"goods"}),
				"definitive_withholding": domain.BoolOption(false),
				"coeficiente":            domain.StringOption("0"),
			},
		},
		{
			name: "rate bounds inclusive (upper)",
			values: map[string]domain.OptionValue{
				"platform_types": domain.ListOption([]string{"goods"}),
				"coeficiente":    domain.StringOption("1.0"),
			},
			want: map[string]domain.OptionValue{
				"deduction_mode":         domain.StringOption("blind_35"),
				"platform_types":         domain.ListOption([]string{"goods"}),
				"definitive_withholding": domain.BoolOption(false),
				"coeficiente":            domain.StringOption("1"),
			},
		},
	}
	set := catalogSet(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := set.Validate(tt.values)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOptionSetValidateRejects(t *testing.T) {
	valid := func(overrides map[string]domain.OptionValue) map[string]domain.OptionValue {
		values := map[string]domain.OptionValue{"platform_types": domain.ListOption([]string{"goods"})}
		for key, value := range overrides {
			values[key] = value
		}
		return values
	}
	tests := []struct {
		name    string
		values  map[string]domain.OptionValue
		wantKey string
		wantErr error
	}{
		{name: "required option without default missing", values: nil, wantKey: "platform_types", wantErr: domain.ErrOptionMissing},
		{name: "required multi enum empty", values: valid(map[string]domain.OptionValue{"platform_types": domain.ListOption(nil)}), wantKey: "platform_types", wantErr: domain.ErrOptionValue},
		{name: "unknown key", values: valid(map[string]domain.OptionValue{"coefficient": domain.StringOption("0.2")}), wantKey: "coefficient", wantErr: domain.ErrOptionUnknown},
		{name: "enum value not allowed", values: valid(map[string]domain.OptionValue{"deduction_mode": domain.StringOption("blind_40")}), wantKey: "deduction_mode", wantErr: domain.ErrOptionValue},
		{name: "enum value is case sensitive", values: valid(map[string]domain.OptionValue{"deduction_mode": domain.StringOption("ACTUAL")}), wantKey: "deduction_mode", wantErr: domain.ErrOptionValue},
		{name: "enum given a boolean", values: valid(map[string]domain.OptionValue{"deduction_mode": domain.BoolOption(true)}), wantKey: "deduction_mode", wantErr: domain.ErrOptionType},
		{name: "enum given a list", values: valid(map[string]domain.OptionValue{"deduction_mode": domain.ListOption([]string{"actual"})}), wantKey: "deduction_mode", wantErr: domain.ErrOptionType},
		{name: "zero value given", values: valid(map[string]domain.OptionValue{"deduction_mode": {}}), wantKey: "deduction_mode", wantErr: domain.ErrOptionType},
		{name: "boolean given a string", values: valid(map[string]domain.OptionValue{"definitive_withholding": domain.StringOption("true")}), wantKey: "definitive_withholding", wantErr: domain.ErrOptionType},
		{name: "multi enum given a string", values: valid(map[string]domain.OptionValue{"platform_types": domain.StringOption("goods")}), wantKey: "platform_types", wantErr: domain.ErrOptionType},
		{name: "multi enum item not allowed", values: valid(map[string]domain.OptionValue{"platform_types": domain.ListOption([]string{"goods", "crypto"})}), wantKey: "platform_types", wantErr: domain.ErrOptionValue},
		{name: "multi enum duplicate item", values: valid(map[string]domain.OptionValue{"platform_types": domain.ListOption([]string{"goods", "goods"})}), wantKey: "platform_types", wantErr: domain.ErrOptionValue},
		{name: "rate given a boolean", values: valid(map[string]domain.OptionValue{"coeficiente": domain.BoolOption(true)}), wantKey: "coeficiente", wantErr: domain.ErrOptionType},
		{name: "rate not a decimal", values: valid(map[string]domain.OptionValue{"coeficiente": domain.StringOption("12%")}), wantKey: "coeficiente", wantErr: domain.ErrOptionValue},
		{name: "rate in exponent notation", values: valid(map[string]domain.OptionValue{"coeficiente": domain.StringOption("1e-1")}), wantKey: "coeficiente", wantErr: domain.ErrOptionValue},
		{name: "rate with more than six decimals", values: valid(map[string]domain.OptionValue{"coeficiente": domain.StringOption("0.1234567")}), wantKey: "coeficiente", wantErr: domain.ErrOptionValue},
		{name: "rate below the lower bound", values: valid(map[string]domain.OptionValue{"coeficiente": domain.StringOption("-0.000001")}), wantKey: "coeficiente", wantErr: domain.ErrOptionValue},
		{name: "rate above the upper bound", values: valid(map[string]domain.OptionValue{"coeficiente": domain.StringOption("1.000001")}), wantKey: "coeficiente", wantErr: domain.ErrOptionValue},
	}
	set := catalogSet(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := set.Validate(tt.values)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got)
			optionErr, ok := errors.AsType[*domain.OptionError](err)
			require.True(t, ok, "want an *OptionError in %v", err)
			assert.Equal(t, tt.wantKey, optionErr.Key)
		})
	}
}

// TestOptionSetValidateReportsEveryFieldInOrder: the adapter turns each
// *OptionError into one FieldError, so all of them are reported, unknown
// keys first (sorted) and then the catalog's definitions in order.
func TestOptionSetValidateReportsEveryFieldInOrder(t *testing.T) {
	set := catalogSet(t)
	_, err := set.Validate(map[string]domain.OptionValue{
		"zeta":           domain.BoolOption(true),
		"alpha":          domain.BoolOption(true),
		"deduction_mode": domain.StringOption("blind_40"),
		"coeficiente":    domain.StringOption("2"),
	})
	require.Error(t, err)
	joined, ok := errors.AsType[interface {
		error
		Unwrap() []error
	}](err)
	require.True(t, ok)
	var keys []string
	for _, each := range joined.Unwrap() {
		optionErr, isOption := errors.AsType[*domain.OptionError](each)
		require.True(t, isOption)
		keys = append(keys, optionErr.Key)
	}
	assert.Equal(t, []string{"alpha", "zeta", "deduction_mode", "platform_types", "coeficiente"}, keys)
}

func TestOptionSetValidateDoesNotAliasTheInput(t *testing.T) {
	set := catalogSet(t)
	items := []string{"goods"}
	got, err := set.Validate(map[string]domain.OptionValue{"platform_types": domain.ListOption(items)})
	require.NoError(t, err)
	items[0] = "crypto"
	list, ok := got["platform_types"].List()
	require.True(t, ok)
	assert.Equal(t, []string{"goods"}, list)
	list[0] = "crypto"
	again, _ := got["platform_types"].List()
	assert.Equal(t, []string{"goods"}, again)
}

func TestEmptyOptionSetAcceptsOnlyNoOptions(t *testing.T) {
	set, err := domain.NewOptionSet()
	require.NoError(t, err)
	got, err := set.Validate(nil)
	require.NoError(t, err)
	assert.Empty(t, got)
	_, err = set.Validate(map[string]domain.OptionValue{"deduction_mode": domain.StringOption("actual")})
	require.ErrorIs(t, err, domain.ErrOptionUnknown)
}

func TestNewOptionSetRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name string
		defs []domain.OptionDef
	}{
		{name: "empty key", defs: []domain.OptionDef{{Type: domain.OptionTypeBoolean}}},
		{name: "duplicate key", defs: []domain.OptionDef{definitiveDef(), definitiveDef()}},
		{name: "unknown type", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionType("decimal")}}},
		{name: "enum without allowed values", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeEnum}}},
		{name: "multi enum without allowed values", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeMultiEnum}}},
		{name: "enum with an empty allowed value", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeEnum, AllowedValues: []string{"a", ""}}}},
		{name: "enum with a duplicate allowed value", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeEnum, AllowedValues: []string{"a", "a"}}}},
		{name: "boolean with allowed values", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeBoolean, AllowedValues: []string{"a"}}}},
		{name: "rate with allowed values", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeRate, AllowedValues: []string{"a"}, Max: 1}}},
		{name: "rate with inverted bounds", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeRate, Min: 2, Max: 1}}},
		{name: "default outside the enum", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeEnum, AllowedValues: []string{"a"}, Default: domain.StringOption("b")}}},
		{name: "default of the wrong type", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeBoolean, Default: domain.StringOption("false")}}},
		{name: "default rate out of bounds", defs: []domain.OptionDef{{Key: "x", Type: domain.OptionTypeRate, Max: money.RateScale, Default: domain.StringOption("2")}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewOptionSet(tt.defs...)
			require.ErrorIs(t, err, domain.ErrInvalidOptionDef)
		})
	}
}

func TestNewOptionSetCanonicalizesDefaults(t *testing.T) {
	set, err := domain.NewOptionSet(domain.OptionDef{
		Key: "coeficiente", Type: domain.OptionTypeRate, Max: money.RateScale, Default: domain.StringOption("0.500"),
	})
	require.NoError(t, err)
	got, err := set.Validate(nil)
	require.NoError(t, err)
	assert.Equal(t, map[string]domain.OptionValue{"coeficiente": domain.StringOption("0.5")}, got)
}

func TestNewOptionSetDoesNotAliasTheDefinitions(t *testing.T) {
	def := deductionModeDef()
	set, err := domain.NewOptionSet(def)
	require.NoError(t, err)
	def.AllowedValues[1] = "blind_40"
	_, err = set.Validate(map[string]domain.OptionValue{"deduction_mode": domain.StringOption("actual")})
	require.NoError(t, err)
}

func TestOptionValueAccessors(t *testing.T) {
	b, isBool := domain.BoolOption(true).Bool()
	assert.True(t, b)
	assert.True(t, isBool)
	_, isBool = domain.StringOption("x").Bool()
	assert.False(t, isBool)

	s, isText := domain.StringOption("x").Text()
	assert.Equal(t, "x", s)
	assert.True(t, isText)
	_, isText = domain.BoolOption(true).Text()
	assert.False(t, isText)

	list, isList := domain.ListOption([]string{"a"}).List()
	assert.Equal(t, []string{"a"}, list)
	assert.True(t, isList)
	_, isList = domain.StringOption("x").List()
	assert.False(t, isList)

	assert.True(t, domain.OptionValue{}.IsZero())
	assert.False(t, domain.BoolOption(false).IsZero())
}

func TestOptionErrorText(t *testing.T) {
	err := &domain.OptionError{Key: "deduction_mode", Err: domain.ErrOptionValue}
	assert.Equal(t, "taxpayers: option deduction_mode: taxpayers: option value not allowed", err.Error())
	assert.ErrorIs(t, err, domain.ErrOptionValue)
}
