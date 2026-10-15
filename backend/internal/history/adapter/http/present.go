package historyhttp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/romodevx/control-contable/backend/internal/history/domain"
	"github.com/romodevx/control-contable/backend/internal/history/usecase"
)

// The response body of getHistory, field for field as the contract
// (HistoryYear): every required field is sent, null where the contract
// says so, and lists are never null. The generated types omit nil fields,
// which the contract does not allow here.

type taxJSON struct {
	ISRPayableRoundedCents   int64  `json:"isr_payable_rounded_cents"`
	IVAPayableRoundedCents   int64  `json:"iva_payable_rounded_cents"`
	IVAOutcome               string `json:"iva_outcome"`
	IVAInFavorCents          int64  `json:"iva_in_favor_cents"`
	WithholdingsToRemitCents int64  `json:"withholdings_to_remit_cents"`
	TotalPayableRoundedCents int64  `json:"total_payable_rounded_cents"`
}

type parameterSetJSON struct {
	FiscalYear int `json:"fiscal_year"`
	Version    int `json:"version"`
}

type monthJSON struct {
	Period             string           `json:"period"`
	Status             string           `json:"status"`
	HasRecords         bool             `json:"has_records"`
	Tax                *taxJSON         `json:"tax"`
	DueDate            string           `json:"due_date"`
	CPAReviewPending   bool             `json:"cpa_review_pending"`
	IsCurrent          bool             `json:"is_current"`
	RecordsFingerprint string           `json:"records_fingerprint"`
	FiguresFingerprint string           `json:"figures_fingerprint"`
	EngineVersion      string           `json:"engine_version"`
	ParameterSet       parameterSetJSON `json:"parameter_set"`
	ViewedAt           *string          `json:"viewed_at"`
	ChangedSinceViewed bool             `json:"changed_since_viewed"`
	ChangeReason       *string          `json:"change_reason"`
}

type yearTotalsJSON struct {
	ISRPayableRoundedCents   int64 `json:"isr_payable_rounded_cents"`
	IVAPayableRoundedCents   int64 `json:"iva_payable_rounded_cents"`
	WithholdingsToRemitCents int64 `json:"withholdings_to_remit_cents"`
	TotalPayableRoundedCents int64 `json:"total_payable_rounded_cents"`
}

type disclaimerJSON struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Text     string `json:"text"`
}

type yearJSON struct {
	TaxpayerID  string           `json:"taxpayer_id"`
	FiscalYear  int              `json:"fiscal_year"`
	Months      []monthJSON      `json:"months"`
	YearTotals  yearTotalsJSON   `json:"year_totals"`
	Legend      string           `json:"legend"`
	Disclaimers []disclaimerJSON `json:"disclaimers"`
}

func toYear(y usecase.Year) yearJSON {
	months := make([]monthJSON, 0, len(y.Months))
	for _, m := range y.Months {
		months = append(months, toMonth(m))
	}
	disclaimers := make([]disclaimerJSON, 0, len(y.Disclaimers))
	for _, d := range y.Disclaimers {
		disclaimers = append(disclaimers, disclaimerJSON{Code: d.Code, Severity: string(d.Severity), Text: d.Text})
	}
	t := y.Totals
	return yearJSON{
		TaxpayerID: y.TaxpayerID, FiscalYear: y.FiscalYear, Months: months,
		YearTotals: yearTotalsJSON{
			ISRPayableRoundedCents: int64(t.ISRPayable), IVAPayableRoundedCents: int64(t.IVAPayable),
			WithholdingsToRemitCents: int64(t.WithholdingsToRemit), TotalPayableRoundedCents: int64(t.TotalPayable),
		},
		Legend: y.Legend, Disclaimers: disclaimers,
	}
}

func toMonth(m usecase.Month) monthJSON {
	out := monthJSON{
		Period: m.Month.String(), Status: string(m.Status), HasRecords: m.HasRecords, Tax: toTax(m.Tax),
		DueDate: m.DueDate, CPAReviewPending: m.CPAReviewPending, IsCurrent: m.IsCurrent,
		RecordsFingerprint: m.Seen.RecordsFingerprint, FiguresFingerprint: m.Seen.FiguresFingerprint,
		EngineVersion: m.Seen.EngineVersion,
		ParameterSet:  parameterSetJSON{FiscalYear: m.ParameterSet.FiscalYear, Version: m.ParameterSet.Version},
	}
	if m.ViewedAt != nil {
		at := m.ViewedAt.UTC().Format(time.RFC3339)
		out.ViewedAt = &at
	}
	if m.Change != "" {
		reason := string(m.Change)
		out.ChangedSinceViewed, out.ChangeReason = true, &reason
	}
	return out
}

func toTax(t *domain.Tax) *taxJSON {
	if t == nil {
		return nil
	}
	return &taxJSON{
		ISRPayableRoundedCents: int64(t.ISRPayable), IVAPayableRoundedCents: int64(t.IVAPayable), IVAOutcome: t.IVAOutcome,
		IVAInFavorCents: int64(t.IVAInFavor), WithholdingsToRemitCents: int64(t.WithholdingsToRemit),
		TotalPayableRoundedCents: int64(t.TotalPayable),
	}
}

// okResponse is the 200 JSON body of getHistory. Its Cache-Control:
// no-store is written by middleware.NoStore, as on every response.
type okResponse struct {
	body yearJSON
}

// VisitGetHistoryResponse implements gen.GetHistoryResponseObject. It
// encodes first, so a failure leaves the response unwritten for the strict
// handler's error path.
func (r okResponse) VisitGetHistoryResponse(w http.ResponseWriter) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(r.body); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := buf.WriteTo(w)
	return err
}
