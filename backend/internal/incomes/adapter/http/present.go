package incomeshttp

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// Client-safe messages of a field the adapter cannot decode; they never
// echo the value (the router's and the use cases' wording).
const (
	msgInvalidInput = "One or more fields are invalid."
	msgFieldInvalid = "This value is not valid."
)

// dateLayout is a contract date (BusinessDate) and fiscal.Date's String.
const dateLayout = "2006-01-02"

// invalidField is the 400 naming one field the adapter could not decode
// (contract validation runs first, so this is a backstop).
func invalidField(field string, cause error) error {
	return apperr.Wrap(cause, apperr.CodeValidationFailed, msgInvalidInput).
		WithFields(apperr.FieldError{Field: field, Code: fiscal.ErrInvalid.Detail(), Message: msgFieldInvalid})
}

// toCreateInput decodes a create body. A nil body is an empty income: the
// use case reports every required field. The folio and the description go
// as raw text (the use case normalizes them).
func toCreateInput(taxpayerID string, body *gen.IncomeCreate) (usecase.CreateInput, error) {
	in := usecase.CreateInput{TaxpayerID: taxpayerID}
	if body == nil {
		return in, nil
	}
	invoiced, err := optionalDate(domain.FieldInvoicedOn, body.InvoicedOn)
	if err != nil {
		return usecase.CreateInput{}, err
	}
	collected, err := optionalDate(domain.FieldCollectedOn, body.CollectedOn)
	if err != nil {
		return usecase.CreateInput{}, err
	}
	if body.Id != nil {
		in.ID = body.Id.String()
	}
	in.Income = usecase.IncomeInput{
		Regime: fiscal.RegimeCode(body.RegimeCode), Counterparty: fiscal.ReceptorKind(body.CounterpartyKind),
		Operation: operationOrEmpty(body.OperationKind), ResidentialRent: body.ResidentialRent != nil && *body.ResidentialRent,
		ViaPlatform: body.ViaPlatform, Amount: money.Cents(body.AmountCents), IVARate: fiscal.IVARate(body.IvaRate),
		IVA:      money.Cents(body.IvaCents),
		Dates:    domain.Dates{InvoicedOn: invoiced, CollectedOn: collected},
		CFDIUUID: textOrEmpty(body.CfdiUuid), Description: textOrEmpty(body.Description),
	}
	if body.PlatformActivity != nil {
		in.Income.PlatformActivity = domain.PlatformActivity(*body.PlatformActivity)
	}
	return in, nil
}

// toPatch decodes a merge-patch: an absent field stays nil; null clears a
// date, the 625 activity, the folio or the description (their zero values).
func toPatch(body *gen.IncomeUpdate) (usecase.Patch, error) {
	var p usecase.Patch
	if body == nil {
		return p, nil
	}
	invoiced, err := patchDate(domain.FieldInvoicedOn, body.InvoicedOn)
	if err != nil {
		return usecase.Patch{}, err
	}
	collected, err := patchDate(domain.FieldCollectedOn, body.CollectedOn)
	if err != nil {
		return usecase.Patch{}, err
	}
	p.InvoicedOn, p.CollectedOn = invoiced, collected
	p.Regime = convert(body.RegimeCode, func(s string) fiscal.RegimeCode { return fiscal.RegimeCode(s) })
	p.Counterparty = convert(body.CounterpartyKind, func(k gen.CounterpartyKind) fiscal.ReceptorKind { return fiscal.ReceptorKind(k) })
	p.Operation = convert(body.OperationKind, func(o gen.OperationKind) fiscal.OperationKind { return fiscal.OperationKind(o) })
	p.ResidentialRent, p.ViaPlatform = body.ResidentialRent, body.ViaPlatform
	p.Amount = convert(body.AmountCents, toCents)
	p.IVARate = convert(body.IvaRate, func(r gen.IvaRate) fiscal.IVARate { return fiscal.IVARate(r) })
	p.IVA = convert(body.IvaCents, toCents)
	p.PlatformActivity = patchValue(body.PlatformActivity, func(a gen.PlatformActivity) domain.PlatformActivity {
		return domain.PlatformActivity(a)
	})
	p.CFDIUUID = patchValue(body.CfdiUuid, func(s string) string { return s })
	p.Description = patchValue(body.Description, func(s string) string { return s })
	return p, nil
}

// patchValue is nil when absent, the zero value (clear) for null, else the
// mapped value.
func patchValue[From, To any](field nullable.Nullable[From], f func(From) To) *To {
	if !field.IsSpecified() {
		return nil
	}
	var out To
	if !field.IsNull() {
		out = f(field.MustGet())
	}
	return &out
}

// patchDate is patchValue for a date, which may fail to decode.
func patchDate(name string, field nullable.Nullable[openapi_types.Date]) (*fiscal.Date, error) {
	if !field.IsSpecified() {
		return nil, nil
	}
	var out fiscal.Date
	if !field.IsNull() {
		d, err := toDate(name, field.MustGet())
		if err != nil {
			return nil, err
		}
		out = d
	}
	return &out, nil
}

// optionalDate maps an optional contract date; nil is the zero Date.
func optionalDate(name string, d *openapi_types.Date) (fiscal.Date, error) {
	if d == nil {
		return fiscal.Date{}, nil
	}
	return toDate(name, *d)
}

// toDate maps a contract date; the zero value is an absent date.
func toDate(name string, d openapi_types.Date) (fiscal.Date, error) {
	if d.IsZero() {
		return fiscal.Date{}, nil
	}
	out, err := fiscal.ParseDate(d.String())
	if err != nil {
		return fiscal.Date{}, invalidField(name, err)
	}
	return out, nil
}

func toCents(v int64) money.Cents { return money.Cents(v) }

func operationOrEmpty(o *gen.OperationKind) fiscal.OperationKind {
	if o == nil {
		return ""
	}
	return fiscal.OperationKind(*o)
}

func textOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// convert maps an optional value, keeping nil.
func convert[From, To any](v *From, f func(From) To) *To {
	if v == nil {
		return nil
	}
	out := f(*v)
	return &out
}

func toListQuery(req gen.ListIncomesRequestObject) usecase.ListQuery {
	q := usecase.ListQuery{TaxpayerID: req.TaxpayerId.String(), Period: req.Params.Period}
	if req.Params.Regime != nil {
		q.Regime = fiscal.RegimeCode(*req.Params.Regime)
	}
	if req.Params.Cursor != nil {
		q.Cursor = *req.Params.Cursor
	}
	if req.Params.Limit != nil {
		q.Limit = *req.Params.Limit
	}
	return q
}

func toDefaultsQuery(req gen.GetIncomeDefaultsRequestObject) usecase.DefaultsQuery {
	q := usecase.DefaultsQuery{TaxpayerID: req.TaxpayerId.String(), Regime: fiscal.RegimeCode(req.Params.Regime)}
	if req.Params.On != nil {
		q.On = req.Params.On.String()
	}
	if req.Params.PlatformActivity != nil {
		q.PlatformActivity = domain.PlatformActivity(*req.Params.PlatformActivity)
	}
	return q
}

// toIncome maps a stored income with the write's warnings (none on reads).
// Its ids and dates come from the database's own types, so a value that
// fails to parse is a server fault (500), never a client error. So is an
// income without facts: the use case fills a legacy one's (0.12.0,
// ADR-0008 Amendment 2). A legacy income not derived yet has a null
// withholding.
func toIncome(i usecase.Income, warnings []domain.Warning) (gen.Income, error) {
	id, err := uuid.Parse(i.ID)
	if err != nil {
		return gen.Income{}, fmt.Errorf("incomes http: income id: %w", err)
	}
	taxpayerID, err := uuid.Parse(i.TaxpayerID)
	if err != nil {
		return gen.Income{}, fmt.Errorf("incomes http: taxpayer id of %s: %w", i.ID, err)
	}
	invoiced, err := contractDate(i.Dates.InvoicedOn)
	if err != nil {
		return gen.Income{}, fmt.Errorf("incomes http: invoiced_on of %s: %w", i.ID, err)
	}
	collected, err := contractDate(i.Dates.CollectedOn)
	if err != nil {
		return gen.Income{}, fmt.Errorf("incomes http: collected_on of %s: %w", i.ID, err)
	}
	if i.Counterparty == "" || i.Operation == "" {
		return gen.Income{}, fmt.Errorf("incomes http: facts of %s: %w", i.ID, errNoFacts)
	}
	withholding, err := storedWithholding(i.Withholding, explained{
		amount: i.Amount, iva: i.IVA, receptor: i.Counterparty, activity: i.PlatformActivity,
	})
	if err != nil {
		return gen.Income{}, fmt.Errorf("incomes http: withholding of %s: %w", i.ID, err)
	}
	return gen.Income{
		Id: id, TaxpayerId: taxpayerID, RegimeCode: string(i.Regime), AmountCents: int64(i.Amount),
		CounterpartyKind: gen.CounterpartyKind(i.Counterparty), OperationKind: gen.OperationKind(i.Operation),
		ResidentialRent: i.ResidentialRent, ViaPlatform: i.ViaPlatform, Withholding: withholding,
		IvaRate: gen.IvaRate(i.IVARate), IvaCents: int64(i.IVA),
		IsrWithheldCents: int64(i.Withholding.ISR), IvaWithheldCents: int64(i.Withholding.IVA),
		InvoicedOn: invoiced, CollectedOn: collected, PendingCollection: i.PendingCollection,
		PlatformActivity: nullableActivity(i.PlatformActivity), CfdiUuid: nullableText(i.CFDIUUID.String()),
		Description: nullableText(i.Description), Warnings: toWarnings(warnings),
		CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt,
	}, nil
}

// contractDate is nil (JSON null) for the zero Date.
func contractDate(d fiscal.Date) (*openapi_types.Date, error) {
	if d.IsZero() {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, d.String())
	if err != nil {
		return nil, err
	}
	return &openapi_types.Date{Time: t}, nil
}

// nullableText is nil (JSON null) for "".
func nullableText(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableActivity(a domain.PlatformActivity) *gen.PlatformActivity {
	if a == "" {
		return nil
	}
	out := gen.PlatformActivity(a)
	return &out
}

func toPage(p usecase.Page) (gen.IncomePage, error) {
	items := make([]gen.Income, 0, len(p.Items))
	for _, i := range p.Items {
		item, err := toIncome(i, nil)
		if err != nil {
			return gen.IncomePage{}, err
		}
		items = append(items, item)
	}
	return gen.IncomePage{
		Items: items, NextCursor: nullableText(p.NextCursor),
		Totals: gen.PeriodTotals{
			Count: int(p.Totals.Count), AmountCents: int64(p.Totals.Amount), IvaCents: int64(p.Totals.IVA),
			IsrWithheldCents: int64(p.Totals.ISRWithheld), IvaWithheldCents: int64(p.Totals.IVAWithheld),
		},
	}, nil
}

func toDefaults(d usecase.Defaults) gen.IncomeDefaults {
	rates := make([]gen.IvaRate, 0, len(d.IVARatesAllowed))
	for _, r := range d.IVARatesAllowed {
		rates = append(rates, gen.IvaRate(r))
	}
	return gen.IncomeDefaults{
		IvaRatesAllowed: rates, SuggestedIvaRate: gen.IvaRate(d.SuggestedIVARate),
		PlatformActivity: nullableActivity(d.PlatformActivity),
		LastUsed: gen.IncomeLastUsed{
			CounterpartyKind: optionalKind(d.LastUsed.Counterparty), OperationKind: gen.OperationKind(d.LastUsed.Operation),
			ResidentialRent: d.LastUsed.ResidentialRent, ViaPlatform: d.LastUsed.ViaPlatform,
		},
	}
}

func toPreviewQuery(req gen.PreviewIncomeWithholdingsRequestObject) usecase.PreviewQuery {
	p := req.Params
	q := usecase.PreviewQuery{
		TaxpayerID: req.TaxpayerId.String(), Regime: fiscal.RegimeCode(p.Regime), Amount: money.Cents(p.AmountCents),
		IVARate: fiscal.IVARate(p.IvaRate), IVA: money.Cents(p.IvaCents), Counterparty: fiscal.ReceptorKind(p.CounterpartyKind),
		Operation: operationOrEmpty(p.OperationKind), ResidentialRent: p.ResidentialRent != nil && *p.ResidentialRent,
		ViaPlatform: p.ViaPlatform,
	}
	if p.On != nil {
		q.On = p.On.String()
	}
	if p.PlatformActivity != nil {
		q.PlatformActivity = domain.PlatformActivity(*p.PlatformActivity)
	}
	return q
}

// toPreview is the contract WithholdingPreview: a preview always derives.
func toPreview(p usecase.Preview) (gen.WithholdingPreview, error) {
	w, err := toWithholding(p.Withholding, explained{
		amount: p.Amount, iva: p.IVA, receptor: p.Counterparty, activity: p.PlatformActivity,
	})
	if err != nil {
		return gen.WithholdingPreview{}, fmt.Errorf("incomes http: preview withholding: %w", err)
	}
	return gen.WithholdingPreview{Withholding: w, Warnings: toWarnings(p.Warnings)}, nil
}
