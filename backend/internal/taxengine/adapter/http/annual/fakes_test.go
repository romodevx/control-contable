package annualhttp_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/httperr"
	"github.com/romodevx/control-contable/backend/adapter/http/middleware"
	"github.com/romodevx/control-contable/backend/adapter/http/router"
	"github.com/romodevx/control-contable/backend/adapter/http/router/routertest"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/principal"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	annualhttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

const (
	userA      = "0199d4a2-0000-7000-8000-00000000000a"
	fisicaID   = "0199d4a2-1111-7000-8000-00000000000a"
	moralID    = "0199d4a2-2222-7000-8000-00000000000a"
	rowID      = "0199d4b1-2b3c-7d4e-8f5a-6b7c8d9e0f01"
	userHeader = "X-Test-User"
)

// principalAuthn admits the X-Test-User as the principal.
type principalAuthn struct{}

func (principalAuthn) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get(userHeader)
		if user == "" {
			_ = httperr.Write(w, r, apperr.Mapper{}, apperr.New(apperr.CodeUnauthorized, "Authentication required."))
			return
		}
		next.ServeHTTP(w, r.WithContext(principal.With(r.Context(), principal.Principal{UserID: user})))
	})
}

// server serves the annual handler; every other operation is the nil
// interface (never called here).
type server struct {
	*annualhttp.Handler
	others
}

type others struct{ gen.StrictServerInterface }

// harness is the annual operations behind the real router: flag gate,
// authentication, contract validation and per-user limits.
type harness struct {
	handler http.Handler
	logs    bytes.Buffer
}

func newHarness(t *testing.T, deps annualhttp.Deps, flags flagtest.Static, limits map[string]router.Limiter) *harness {
	t.Helper()
	spec, err := gen.GetSpec()
	require.NoError(t, err)
	h := &harness{}
	routes, err := router.New(server{Handler: annualhttp.NewHandler(deps)}, spec, router.Options{
		Flags: flags, Authenticators: routertest.Authenticators(principalAuthn{}), Logger: logging.New(&h.logs, nil), UserLimits: limits,
	})
	require.NoError(t, err)
	h.handler = middleware.Chain(routes, middleware.NoStore)
	return h
}

func request(method, target, body string) *http.Request {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, target, nil)
	} else {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set(userHeader, userA)
	return req
}

// annualOn is module_annual ON (the router reads the effective state the
// flags service computes; its prerequisites are the service's job).
var annualOn = flagtest.Static{"module_annual": true}

// estimatePorts serve one fixed year to a real usecaseannual.Estimator.
type estimatePorts struct {
	profiles map[string]usecase.TaxpayerProfile
	incomes  []usecase.IncomeMonthSums
	expenses []usecase.ExpenseMonthSums
	inputs   map[fiscal.RegimeCode][]domain.TaxInputValue
	year     usecaseannual.YearInputs
	filed    map[fiscal.RegimeCode]map[time.Month]money.Cents
}

type inlineTx struct{}

func (inlineTx) ReadWrite(ctx context.Context, _ string, work tx.Work) error { return work(ctx) }
func (inlineTx) ReadOnly(ctx context.Context, _ string, work tx.Work) error  { return work(ctx) }

func (p *estimatePorts) Profile(_ context.Context, _, id string) (usecase.TaxpayerProfile, error) {
	prof, ok := p.profiles[id]
	if !ok {
		return usecase.TaxpayerProfile{}, apperr.New(apperr.CodeNotFound, "Not found.")
	}
	return prof, nil
}

func (p *estimatePorts) IncomeSums(context.Context, usecase.LedgerQuery) ([]usecase.IncomeMonthSums, error) {
	return p.incomes, nil
}

func (p *estimatePorts) ExpenseSums(context.Context, usecase.LedgerQuery) ([]usecase.ExpenseMonthSums, error) {
	return p.expenses, nil
}

func (p *estimatePorts) TaxInputs(_ context.Context, _ string, first, _ int) (map[int]map[fiscal.RegimeCode]domain.TaxInputs, error) {
	out := map[fiscal.RegimeCode]domain.TaxInputs{}
	for code, vs := range p.inputs {
		in, err := domain.NewTaxInputs(vs...)
		if err != nil {
			return nil, err
		}
		out[code] = in
	}
	return map[int]map[fiscal.RegimeCode]domain.TaxInputs{first: out}, nil
}

func (p *estimatePorts) YearInputs(context.Context, string, int) (usecaseannual.YearInputs, error) {
	return p.year, nil
}

func (p *estimatePorts) FiledPayments(context.Context, string, int) (map[fiscal.RegimeCode]map[time.Month]money.Cents, error) {
	return p.filed, nil
}

func (p *estimatePorts) Holidays(context.Context) (fiscal.HolidaySet, error) {
	return fiscal.HolidaySet{}, nil
}

// EstimateRange answers 1,000.00 of ISR a month for every regime.
func (p *estimatePorts) EstimateRange(_ context.Context, _, id string, from, to domain.Period) ([]usecase.TaxpayerEstimate, error) {
	var out []usecase.TaxpayerEstimate
	for m := from.Month(); m <= to.Month(); m++ {
		period, err := domain.NewPeriod(from.Year(), m)
		if err != nil {
			return nil, err
		}
		est := usecase.TaxpayerEstimate{Period: period}
		for _, h := range p.profiles[id].Regimes {
			isr := domain.TaxResult{Amount: 1_000_00}
			est.Regimes = append(est.Regimes, usecase.RegimeResult{RegimeEstimate: domain.RegimeEstimate{Key: domain.StrategyKey{Regime: h.Code}, ISR: &isr}})
		}
		out = append(out, est)
	}
	return out, nil
}

// realEstimates is the annual estimate over a PF with 605, 606, 612 and 626
// (deductions that bind caps, a cash exclusion, a 612 loss) and a PM with
// 601 and 603, on a day of 2027.
func realEstimates(t *testing.T) *usecaseannual.Estimator {
	t.Helper()
	start, err := domain.ParsePeriod("2024-01")
	require.NoError(t, err)
	actual := domain.NewOptions(map[domain.OptionKey]domain.OptionValue{domain.OptionRentDeductionMode: domain.TextOption("actual")})
	p := &estimatePorts{
		profiles: map[string]usecase.TaxpayerProfile{
			fisicaID: {PersonType: fiscal.PersonFisica, Regimes: []usecase.RegimeHold{
				{Code: "605", Start: start}, {Code: "606", Start: start, Options: actual}, {Code: "612", Start: start}, {Code: "626", Start: start},
			}},
			moralID: {PersonType: fiscal.PersonMoral, Regimes: []usecase.RegimeHold{{Code: "601", Start: start}, {Code: "603", Start: start}}},
		},
		incomes: []usecase.IncomeMonthSums{
			{Regime: "606", Year: 2026, Month: time.March, Sums: domain.IncomeSums{Cash: 240_000_00}},
			{Regime: "612", Year: 2026, Month: time.March, Sums: domain.IncomeSums{Cash: 100_000_00, ISRWithheld: domain.Withheld{ByMoral: 10_000_00}}},
			{Regime: "626", Year: 2026, Month: time.March, Sums: domain.IncomeSums{Cash: 300_000_00}},
			{Regime: "601", Year: 2026, Month: time.March, Sums: domain.IncomeSums{Accrual: 5_000_000_00, Cash: 5_000_000_00}},
		},
		expenses: []usecase.ExpenseMonthSums{
			{Regime: "606", Year: 2026, Month: time.March, Sums: domain.ExpenseSums{DeductiblePaid: 40_000_00}},
			{Regime: "612", Year: 2026, Month: time.March, Sums: domain.ExpenseSums{DeductiblePaid: 150_000_00}},
			{Regime: "601", Year: 2026, Month: time.March, Sums: domain.ExpenseSums{DeductiblePaid: 4_000_000_00}},
		},
		inputs: map[fiscal.RegimeCode][]domain.TaxInputValue{
			"601": {{Key: domain.InputPTUPaid, Kind: domain.InputMoney, Money: 10_000_00}},
		},
		year: usecaseannual.YearInputs{
			SalaryStatements: []adomain.SalaryStatement{{Accumulable: 300_000_00, Exempt: 20_000_00, ISRWithheld: 40_000_00}},
			Deductions: []adomain.Deduction{
				{Category: adomain.CategoryMedical, Amount: 90_000_00, PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiarySelf},
				{Category: adomain.CategoryMedical, Amount: 1_000_00, PaymentMethod: adomain.PaymentCash, Beneficiary: adomain.BeneficiarySelf},
				{Category: adomain.CategoryTuition, Amount: 30_000_00, PaymentMethod: adomain.PaymentElectronic, Beneficiary: adomain.BeneficiaryDescendant,
					Student: "Ana", TuitionLevel: adomain.LevelPrimaria},
			},
		},
		filed: map[fiscal.RegimeCode]map[time.Month]money.Cents{"601": {time.December: 100_000_00}},
	}
	cat, err := params.Embedded()
	require.NoError(t, err)
	monthly, err := domain.NewRegistry(taxengine.Strategies()...)
	require.NoError(t, err)
	sections, err := usecaseannual.NewSectionRegistry(taxengine.AnnualSections(annuallimits.New(cat))...)
	require.NoError(t, err)
	uc, err := usecaseannual.NewEstimator(usecaseannual.EstimateDeps{
		Tx: inlineTx{}, Profiles: p, Ledger: p, TaxInputs: p, YearInputs: p, Filed: p, Monthly: p,
		MonthlyStrategies: monthly, Sections: sections, Holidays: p, Params: cat,
		Clock: clock.NewFixed(time.Date(2027, time.February, 10, 12, 0, 0, 0, clock.Location())),
	})
	require.NoError(t, err)
	return uc
}

// fakeRange is the estimate_reads bucket.
type fakeRange struct {
	allowed bool
	keys    []string
}

func (f *fakeRange) Take(_ context.Context, key string) (bool, time.Duration, error) {
	f.keys = append(f.keys, key)
	return f.allowed, 30 * time.Second, nil
}

var stamp = time.Date(2026, time.October, 14, 18, 0, 0, 0, time.UTC)

func mustDate(t *testing.T, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}
