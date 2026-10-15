package dashboardhttp

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/romodevx/control-contable/backend/adapter/http/accesshttp"
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/dashboard/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// The response body of getDashboard, field for field as the contract
// (Dashboard): every required field is sent, null where the module is OFF.
// The generated types omit nil fields, which the contract does not allow.

type monthJSON struct {
	IncomeCents  *int64 `json:"income_cents"`
	ExpenseCents *int64 `json:"expense_cents"`
}

type trendTaxJSON struct {
	IVAPayableRoundedCents   int64  `json:"iva_payable_rounded_cents"`
	ISRPayableRoundedCents   int64  `json:"isr_payable_rounded_cents"`
	WithholdingsToRemitCents int64  `json:"withholdings_to_remit_cents"`
	TotalPayableRoundedCents int64  `json:"total_payable_rounded_cents"`
	IVAOutcome               string `json:"iva_outcome"`
	Status                   string `json:"status"`
}

type trendPointJSON struct {
	Period       string        `json:"period"`
	IncomeCents  *int64        `json:"income_cents"`
	ExpenseCents *int64        `json:"expense_cents"`
	Tax          *trendTaxJSON `json:"tax"`
}

type dashboardJSON[D any] struct {
	TaxpayerID string            `json:"taxpayer_id"`
	Period     string            `json:"period"`
	IsEmpty    bool              `json:"is_empty"`
	Month      monthJSON         `json:"month"`
	Estimate   *D                `json:"estimate"`
	Trend      []trendPointJSON  `json:"trend"`
	Access     gen.AccessSummary `json:"access"`
	// Legend is present if and only if Estimate is (contract Dashboard).
	Legend string `json:"legend,omitempty"`
}

func toDashboard[D any](d usecase.Dashboard[D]) dashboardJSON[D] {
	trend := make([]trendPointJSON, 0, len(d.Trend))
	for _, p := range d.Trend {
		trend = append(trend, trendPointJSON{
			Period: p.Month.String(), IncomeCents: centsOf(p.Income), ExpenseCents: centsOf(p.Expense), Tax: toTrendTax(p.Tax),
		})
	}
	return dashboardJSON[D]{
		TaxpayerID: d.TaxpayerID,
		Period:     d.Period.String(),
		IsEmpty:    d.IsEmpty,
		Month:      monthJSON{IncomeCents: centsOf(d.Current.Income), ExpenseCents: centsOf(d.Current.Expense)},
		Estimate:   d.Estimate,
		Trend:      trend,
		Access:     accesshttp.Summary(d.Access),
		Legend:     d.Legend,
	}
}

func centsOf(c *money.Cents) *int64 {
	if c == nil {
		return nil
	}
	v := int64(*c)
	return &v
}

func toTrendTax(t *usecase.TrendTax) *trendTaxJSON {
	if t == nil {
		return nil
	}
	return &trendTaxJSON{
		IVAPayableRoundedCents:   int64(t.IVAPayable),
		ISRPayableRoundedCents:   int64(t.ISRPayable),
		WithholdingsToRemitCents: int64(t.WithholdingsToRemit),
		TotalPayableRoundedCents: int64(t.TotalPayable),
		IVAOutcome:               string(t.IVAOutcome),
		Status:                   string(t.Status),
	}
}

// okResponse is the 200 JSON body of getDashboard. Its Cache-Control:
// no-store is written by middleware.NoStore, as on every response.
type okResponse struct {
	body any
}

// VisitGetDashboardResponse implements gen.GetDashboardResponseObject. It
// encodes first, so a failure leaves the response unwritten for the strict
// handler's error path.
func (r okResponse) VisitGetDashboardResponse(w http.ResponseWriter) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(r.body); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := buf.WriteTo(w)
	return err
}
