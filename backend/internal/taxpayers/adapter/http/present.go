package taxpayershttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

const (
	// optionsField is the request field of a regime's options; a create
	// prefixes it with "regimes.N.".
	optionsField = "options"
	// msgInvalidInput is the router's wording for every validation_failed.
	msgInvalidInput = "One or more fields are invalid."
	// msgInvalidOptionValue answers an option value the contract cannot
	// carry (contract validation refuses it first; this is the backstop).
	msgInvalidOptionValue = "This value is not allowed for the option."
)

// errOptionValue reports an option value that is not a string, a boolean
// or a list of strings (contract RegimeOptions).
var errOptionValue = errors.New("taxpayers http: option value is not a string, boolean or list of strings")

// toAppError returns a use case error for the router: an *apperr.Error as
// is, a domain error ErrorMapper knows with its contract code, and anything
// else unchanged, so the router logs it and answers the generic 500.
func toAppError(err error) error {
	if _, ok := errors.AsType[*apperr.Error](err); ok {
		return err
	}
	if resolved := ErrorMapper.Resolve(err); resolved.Code != apperr.CodeInternalError {
		return resolved
	}
	return err
}

func toCreateInput(body *gen.TaxpayerCreate) (usecase.CreateInput, error) {
	if body == nil {
		return usecase.CreateInput{}, nil
	}
	in := usecase.CreateInput{PersonType: string(body.PersonType), RFC: body.Rfc, DisplayName: body.DisplayName}
	if body.Id != nil {
		in.ID = body.Id.String()
	}
	for i := range body.Regimes {
		regime, err := toRegimeInput(&body.Regimes[i], fmt.Sprintf("regimes.%d.%s", i, optionsField))
		if err != nil {
			return usecase.CreateInput{}, err
		}
		in.Regimes = append(in.Regimes, regime)
	}
	return in, nil
}

// toRegimeInput maps one TaxpayerRegimeCreate; field names its options in
// a field error.
func toRegimeInput(body *gen.TaxpayerRegimeCreate, field string) (usecase.RegimeInput, error) {
	if body == nil {
		return usecase.RegimeInput{}, nil
	}
	in := usecase.RegimeInput{Code: body.RegimeCode}
	if body.StartPeriod != nil {
		in.Start = *body.StartPeriod
	}
	var err error
	in.Options, err = fromOptions(body.Options, field)
	return in, err
}

// fromOptions decodes the RegimeOptions union, choosing each value's kind
// by its first JSON byte. nil stays nil (no options sent).
func fromOptions(options *gen.RegimeOptions, field string) (map[string]domain.OptionValue, error) {
	if options == nil {
		return nil, nil
	}
	out := make(map[string]domain.OptionValue, len(*options))
	var fields []apperr.FieldError
	for key, value := range *options {
		raw, err := value.MarshalJSON()
		if err == nil {
			out[key], err = decodeOptionValue(raw)
		}
		if err != nil {
			fields = append(fields, apperr.FieldError{Field: field + "." + key, Code: usecase.DetailInvalidOption, Message: msgInvalidOptionValue})
		}
	}
	if len(fields) > 0 {
		return nil, apperr.Wrap(errOptionValue, apperr.CodeValidationFailed, msgInvalidInput).WithFields(fields...)
	}
	return out, nil
}

func decodeOptionValue(raw []byte) (domain.OptionValue, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return domain.OptionValue{}, errOptionValue
	}
	switch raw[0] {
	case 't', 'f':
		var flag bool
		if err := json.Unmarshal(raw, &flag); err != nil {
			return domain.OptionValue{}, errors.Join(errOptionValue, err)
		}
		return domain.BoolOption(flag), nil
	case '"':
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return domain.OptionValue{}, errors.Join(errOptionValue, err)
		}
		return domain.StringOption(text), nil
	case '[':
		list := []string{}
		if err := json.Unmarshal(raw, &list); err != nil {
			return domain.OptionValue{}, errors.Join(errOptionValue, err)
		}
		return domain.ListOption(list), nil
	}
	return domain.OptionValue{}, errOptionValue
}

// toOptions encodes options for the RegimeOptions union. A zero value (never
// stored: options are validated and defaulted) is skipped.
func toOptions(options map[string]domain.OptionValue) gen.RegimeOptions {
	out := make(gen.RegimeOptions, len(options))
	for key, value := range options {
		var union gen.RegimeOptions_AdditionalProperties
		if setOption(&union, value) == nil {
			out[key] = union
		}
	}
	return out
}

func setOption(union *gen.RegimeOptions_AdditionalProperties, value domain.OptionValue) error {
	if flag, ok := value.Bool(); ok {
		return union.FromRegimeOptions1(flag)
	}
	if text, ok := value.Text(); ok {
		return union.FromRegimeOptions0(text)
	}
	if list, ok := value.List(); ok {
		if list == nil {
			list = []string{}
		}
		return union.FromRegimeOptions2(list)
	}
	return errOptionValue
}

// toTaxpayer maps a taxpayer; its id comes from a uuid column, so a parse
// failure is a server fault.
func toTaxpayer(t usecase.Taxpayer) (gen.Taxpayer, error) {
	id, err := uuid.Parse(t.ID)
	if err != nil {
		return gen.Taxpayer{}, fmt.Errorf("taxpayers http: taxpayer id: %w", err)
	}
	regimes := make([]gen.TaxpayerRegime, 0, len(t.Regimes))
	for _, r := range t.Regimes {
		regime := toTaxpayerRegime(r)
		regime.Warnings = nil // never inside Taxpayer.regimes
		regimes = append(regimes, regime)
	}
	out := gen.Taxpayer{
		Id: id, PersonType: gen.PersonType(t.PersonType), DisplayName: t.DisplayName,
		RfcMasked: t.RFCMasked, Regimes: regimes, CreatedAt: t.CreatedAt.UTC(), Warnings: toWarnings(t.Warnings),
	}
	if !t.ArchivedAt.IsZero() {
		archived := t.ArchivedAt.UTC()
		out.ArchivedAt = &archived
	}
	return out, nil
}

func toTaxpayerRegime(r usecase.TaxpayerRegime) gen.TaxpayerRegime {
	out := gen.TaxpayerRegime{
		RegimeCode: r.Code, Name: r.Name, Treatment: gen.Treatment(r.Treatment), IncomeBasis: gen.IncomeBasis(r.IncomeBasis),
		Options: toOptions(r.Options), StartPeriod: r.Window.Start.String(), Warnings: toWarnings(r.Warnings),
	}
	if !r.Window.IsOpen() {
		end := r.Window.End.String()
		out.EndPeriod = &end
	}
	return out
}

// toWarnings is nil for none: warnings are present only when non-empty.
func toWarnings(warnings []usecase.Warning) *[]gen.Warning {
	if len(warnings) == 0 {
		return nil
	}
	out := make([]gen.Warning, 0, len(warnings))
	for _, w := range warnings {
		out = append(out, gen.Warning{Code: gen.WarningCode(w.Code), Field: w.Field, Message: w.Message})
	}
	return &out
}

func toRegimeList(entries []usecase.CatalogEntry) gen.RegimeList {
	items := make([]gen.Regime, 0, len(entries))
	for _, e := range entries {
		pending := toLinks(e.ReviewPendingWith)
		regime := gen.Regime{
			Code: e.Code, Name: e.Name, PersonType: gen.PersonType(e.PersonType), Treatment: gen.Treatment(e.Treatment),
			IncomeBasis: gen.IncomeBasis(e.IncomeBasis),
			OptionDefs:  toOptionDefs(e.OptionDefs), RequiredInputs: toTaxInputDefs(e.RequiredInputs),
			IncompatibleWith: toLinks(e.IncompatibleWith), ReviewPendingWith: &pending,
		}
		if e.Explanation != "" {
			explanation := e.Explanation
			regime.Explanation = &explanation
		}
		items = append(items, regime)
	}
	return gen.RegimeList{Items: items}
}

func toLinks(links []usecase.RegimeLink) []gen.RegimeIncompatibility {
	out := make([]gen.RegimeIncompatibility, 0, len(links))
	for _, l := range links {
		out = append(out, gen.RegimeIncompatibility{Code: l.Code, Reason: l.Reason})
	}
	return out
}

func toOptionDefs(defs []usecase.OptionDefView) []gen.RegimeOptionDef {
	out := make([]gen.RegimeOptionDef, 0, len(defs))
	for _, d := range defs {
		def := gen.RegimeOptionDef{
			Key: d.Key, Label: d.Label, Type: gen.RegimeOptionDefType(d.Type), Required: d.Required,
			Min: optionalText(d.Min), Max: optionalText(d.Max), Default: defaultValue(d.Default),
		}
		if d.AllowedValues != nil {
			allowed := d.AllowedValues
			def.AllowedValues = &allowed
		}
		out = append(out, def)
	}
	return out
}

// defaultValue is the RegimeOptionDef.default: the value, or an explicit
// null when the option has none (the stored shape).
func defaultValue(value domain.OptionValue) *any {
	out := new(any)
	if flag, ok := value.Bool(); ok {
		*out = flag
	} else if text, ok := value.Text(); ok {
		*out = text
	}
	return out
}

func optionalText(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func toTaxInputDefs(inputs []usecase.TaxInputDefView) []gen.TaxInputDef {
	out := make([]gen.TaxInputDef, 0, len(inputs))
	for _, in := range inputs {
		out = append(out, gen.TaxInputDef{Key: in.Key, Label: in.Label, Kind: gen.TaxInputDefKind(in.Kind), Required: in.Required})
	}
	return out
}
