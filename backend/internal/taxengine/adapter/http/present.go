package taxenginehttp

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
)

// taxInputValueJSON is contract TaxInputValue as responses send it: every
// field, null when not set. The generated gen.TaxInputValue omits nil
// fields (omitempty), which the contract allows on requests only.
type taxInputValueJSON struct {
	Key        string  `json:"key"`
	Month      *int    `json:"month"`
	ValueCents *int64  `json:"value_cents"`
	ValueRate  *string `json:"value_rate"`
	ValueBool  *bool   `json:"value_bool"`
}

// taxInputsJSON is contract TaxInputs with taxInputValueJSON values.
type taxInputsJSON struct {
	RegimeCode  gen.RegimeCode      `json:"regime_code"`
	FiscalYear  gen.FiscalYear      `json:"fiscal_year"`
	Definitions []gen.TaxInputDef   `json:"definitions"`
	Values      []taxInputValueJSON `json:"values"`
}

// taxInputsResponse is the 200 of getTaxInputs and putTaxInputs.
type taxInputsResponse struct {
	body taxInputsJSON
}

// VisitGetTaxInputsResponse implements gen.GetTaxInputsResponseObject.
func (r taxInputsResponse) VisitGetTaxInputsResponse(w http.ResponseWriter) error { return r.write(w) }

// VisitPutTaxInputsResponse implements gen.PutTaxInputsResponseObject.
func (r taxInputsResponse) VisitPutTaxInputsResponse(w http.ResponseWriter) error { return r.write(w) }

func (r taxInputsResponse) write(w http.ResponseWriter) error { return writeOK(w, r.body) }

// okResponse is the 200 JSON body of the estimate and the breakdown. Its
// Cache-Control: no-store (S-4) is written by middleware.NoStore, as on
// every response.
type okResponse struct {
	body any
}

// VisitGetTaxpayerEstimateResponse implements gen.GetTaxpayerEstimateResponseObject.
func (r okResponse) VisitGetTaxpayerEstimateResponse(w http.ResponseWriter) error {
	return writeOK(w, r.body)
}

// VisitGetEstimateBreakdownResponse implements gen.GetEstimateBreakdownResponseObject.
func (r okResponse) VisitGetEstimateBreakdownResponse(w http.ResponseWriter) error {
	return writeOK(w, r.body)
}

// writeOK mirrors the generated 200 writers: encode first, so a failure
// leaves the response unwritten for the strict handler's error path.
func writeOK(w http.ResponseWriter, body any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := buf.WriteTo(w)
	return err
}

func toTaxInputs(v usecase.TaxInputsView) taxInputsJSON {
	defs := make([]gen.TaxInputDef, 0, len(v.Definitions))
	for _, d := range v.Definitions {
		defs = append(defs, gen.TaxInputDef{Key: string(d.Key), Label: d.Label, Kind: gen.TaxInputDefKind(d.Kind), Required: d.Required})
	}
	values := make([]taxInputValueJSON, 0, len(v.Values))
	for _, s := range v.Values {
		values = append(values, toValue(s))
	}
	return taxInputsJSON{RegimeCode: string(v.Regime), FiscalYear: v.FiscalYear, Definitions: defs, Values: values}
}

// toValue sets exactly the field of the value's kind (and month for
// money_monthly).
func toValue(s usecase.StoredInput) taxInputValueJSON {
	out := taxInputValueJSON{Key: string(s.Key)}
	switch s.Kind {
	case domain.InputMoneyMonthly:
		month := int(s.Month)
		out.Month = &month
		fallthrough
	case domain.InputMoney:
		cents := int64(s.Cents)
		out.ValueCents = &cents
	case domain.InputRate:
		rate := s.Rate
		out.ValueRate = &rate
	case domain.InputBoolean:
		flag := s.Bool
		out.ValueBool = &flag
	}
	return out
}
