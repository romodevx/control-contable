package taxengine_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/clock"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/platform/tx"
	"github.com/romodevx/control-contable/backend/internal/taxengine"
	"github.com/romodevx/control-contable/backend/internal/taxengine/annualgolden"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/annuallimits"
	"github.com/romodevx/control-contable/backend/internal/taxengine/usecase"
	usecaseannual "github.com/romodevx/control-contable/backend/internal/taxengine/usecase/annual"
)

// The annual golden gate runs every case end to end through the
// EstimateAnnual use case (ANN-14): the case's year becomes the ports'
// rows (each regime's figures in one month of the ledger, its tax inputs,
// the constancias, deductions and annual inputs), its payments the filed
// payments, the M5 tax input or the Estimator range by source, and the
// estimate is mapped back to the harness's view. The sections are the
// composition root's list (taxengine.AnnualSections).

const (
	goldenTaxpayer = "0199d4a2-9999-7000-8000-000000000001"
	goldenUser     = "0199d4a2-9999-7000-8000-0000000000aa"
	// goldenRecordMonth holds each regime's year in the ledger.
	goldenRecordMonth = time.June
)

// goldenAfterTheYear is a day of the year after the case's: its fiscal
// year is closed.
func goldenAfterTheYear(fiscalYear int) time.Time {
	return time.Date(fiscalYear+1, time.February, 10, 12, 0, 0, 0, clock.Location())
}

var errGoldenUnknownTaxInput = errors.New("annual golden: tax input the annual does not read")

// goldenAnnualKeys are the money tax inputs a case may set.
var goldenAnnualKeys = []domain.InputKey{
	annual.InputAnnualDepreciation, domain.InputPTUPaid, domain.InputPriorYearLosses, domain.InputInterestWithholding,
	domain.InputYTDBeforeAppIncome, domain.InputYTDBeforeAppDeductions,
}

// useCaseAnnual is the gate's annual engine.
type useCaseAnnual struct {
	catalog  *taxparams.Catalog
	monthly  domain.Registry
	sections usecaseannual.SectionRegistry
}

func newUseCaseAnnual(t testing.TB) useCaseAnnual {
	t.Helper()
	cat := catalog(t)
	sections, err := usecaseannual.NewSectionRegistry(taxengine.AnnualSections(annuallimits.New(cat))...)
	require.NoError(t, err)
	return useCaseAnnual{catalog: cat, monthly: registry(t), sections: sections}
}

// Registered: explanation sections need no strategy.
func (u useCaseAnnual) Registered(kind annual.SectionKind) bool {
	return kind == annual.SectionExplanation || u.sections.Registered(kind)
}

func (u useCaseAnnual) Estimate(in annualgolden.Input) (annualgolden.Result, error) {
	est, err := u.estimate(in)
	if err != nil {
		return annualgolden.Result{}, err
	}
	return goldenResultOf(est), nil
}

// estimate runs the use case on the case's year.
func (u useCaseAnnual) estimate(in annualgolden.Input) (usecaseannual.Estimate, error) {
	ports, err := goldenPortsOf(in)
	if err != nil {
		return usecaseannual.Estimate{}, err
	}
	uc, err := usecaseannual.NewEstimator(usecaseannual.EstimateDeps{
		Tx: goldenTx{}, Profiles: ports, Ledger: ports, TaxInputs: ports, YearInputs: ports, Filed: ports,
		Monthly: ports, MonthlyStrategies: u.monthly, Sections: u.sections, Holidays: ports,
		Params: u.catalog, Clock: clock.NewFixed(goldenAfterTheYear(in.FiscalYear)),
	})
	if err != nil {
		return usecaseannual.Estimate{}, err
	}
	return uc.Estimate(context.Background(), goldenUser, goldenTaxpayer, in.FiscalYear)
}

// goldenResultOf maps the estimate to the harness's view.
func goldenResultOf(est usecaseannual.Estimate) annualgolden.Result {
	out := annualgolden.Result{InformativeIncomeOver500k: est.InformativeIncomeOver500k}
	for _, d := range est.Disclaimers {
		out.Disclaimers = append(out.Disclaimers, d.Code)
	}
	for _, s := range est.Sections {
		got := annualgolden.SectionResult{
			Kind: s.Kind, Regimes: s.Regimes, Status: s.Status, Obligation: s.Obligation,
			MissingInputs: goldenMissing(s.MissingInputs), CPATopics: s.CPATopics,
		}
		if s.RESICOCheck != nil {
			got.Difference = &s.RESICOCheck.Difference
		}
		if s.Tax != nil {
			got.Tax = &annualgolden.TaxFigures{
				AccumulableIncome: s.Tax.AccumulableIncome, Base: s.Tax.Base, ISR: s.Tax.ISR, Payments: s.Tax.Payments,
				Withholdings: s.Tax.Withholdings, SalaryISR: s.Tax.SalaryISR,
				Balance: annualgolden.Balance{Kind: s.Tax.Balance.Kind, Amount: s.Tax.Balance.Amount},
			}
			if s.Tax.PersonalDeductions != nil {
				got.Tax.PersonalDeductions = *s.Tax.PersonalDeductions
			}
		}
		if s.PMExtras != nil {
			got.PMExtras = &annualgolden.PMExtrasResult{PTUBase: s.PMExtras.PTUBase}
			if s.PMExtras.PTUToDistribute != nil {
				got.PMExtras.PTUToDistribute = *s.PMExtras.PTUToDistribute
			}
			if s.PMExtras.SuggestedCU != nil {
				got.PMExtras.SuggestedCU = *s.PMExtras.SuggestedCU
			}
		}
		for _, l := range s.Lines {
			got.Lines = append(got.Lines, annualgolden.ResultLine{
				Line: annualgolden.Line{Code: l.Code, Op: l.Operation, Amount: l.Amount}, Cited: len(l.Basis) > 0,
			})
		}
		out.Sections = append(out.Sections, got)
	}
	return out
}

// goldenMissing renders the structured missing inputs back to the domain
// names the goldens use.
func goldenMissing(ms []usecaseannual.MissingInput) []string {
	var out []string
	for _, m := range ms {
		if len(m.Months) == 0 {
			out = append(out, m.Key)
			continue
		}
		for _, p := range m.Months {
			out = append(out, fmt.Sprintf("%s:%s:%s", m.Key, m.Regime, p))
		}
	}
	return out
}

// goldenTx runs the unit of work inline.
type goldenTx struct{}

func (goldenTx) ReadWrite(ctx context.Context, _ string, work tx.Work) error { return work(ctx) }
func (goldenTx) ReadOnly(ctx context.Context, _ string, work tx.Work) error  { return work(ctx) }

// goldenPorts serves one case's year through every read port.
type goldenPorts struct {
	profile  usecase.TaxpayerProfile
	incomes  []usecase.IncomeMonthSums
	expenses []usecase.ExpenseMonthSums
	inputs   map[fiscal.RegimeCode]domain.TaxInputs
	year     usecaseannual.YearInputs
	filed    map[fiscal.RegimeCode]map[time.Month]money.Cents
	app      map[fiscal.RegimeCode]map[time.Month]money.Cents
	period   int
}

func goldenPortsOf(in annualgolden.Input) (*goldenPorts, error) {
	p := &goldenPorts{
		profile: usecase.TaxpayerProfile{PersonType: in.PersonType},
		filed:   map[fiscal.RegimeCode]map[time.Month]money.Cents{},
		app:     map[fiscal.RegimeCode]map[time.Month]money.Cents{},
		inputs:  map[fiscal.RegimeCode]domain.TaxInputs{},
		period:  in.FiscalYear,
		year: usecaseannual.YearInputs{
			Deductions: in.PersonalDeductions,
			Facts: annual.PFAnnualInputs{
				PriorYearAccumulableIncome: in.AnnualInputs.PriorYearAccumulableIncome,
				OtherExemptIncome:          in.AnnualInputs.OtherExemptIncome,
				ToldEmployerWillFile:       in.AnnualInputs.ToldEmployerWillFile,
			},
		},
	}
	for _, s := range in.SalaryStatements {
		p.year.SalaryStatements = append(p.year.SalaryStatements, annual.SalaryStatement{
			Accumulable: s.Accumulable, Exempt: s.Exempt, ISRWithheld: s.ISRWithheld, Simultaneous: s.Simultaneous,
			LeftBeforeYearEnd: s.LeftBeforeYearEnd, PayerNotWithholding: s.PayerNotWithholding,
		})
	}
	start, err := domain.NewPeriod(in.FiscalYear, time.January)
	if err != nil {
		return nil, err
	}
	for _, r := range in.Regimes {
		if err := p.addRegime(in.FiscalYear, start, r, in.Payments); err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *goldenPorts) addRegime(fiscalYear int, start domain.Period, r annualgolden.RegimeInput, payments []annualgolden.Payment) error {
	options, err := goldenOptions(r.Options)
	if err != nil {
		return fmt.Errorf("%s: %w", r.Regime, err)
	}
	p.profile.Regimes = append(p.profile.Regimes, usecase.RegimeHold{Code: r.Regime, Options: options, Start: start})
	p.incomes = append(p.incomes, usecase.IncomeMonthSums{Regime: r.Regime, Year: fiscalYear, Month: goldenRecordMonth, Sums: domain.IncomeSums{
		Cash: r.Income, Accrual: r.Income, ISRWithheld: domain.Withheld{ByMoral: r.ISRWithheld},
	}})
	p.expenses = append(p.expenses, usecase.ExpenseMonthSums{Regime: r.Regime, Year: fiscalYear, Month: goldenRecordMonth, Sums: domain.ExpenseSums{
		DeductiblePaid: r.Deductions, PredialPaid: r.Predial,
	}})
	var values []domain.TaxInputValue
	for _, ti := range r.TaxInputs {
		if !slices.Contains(goldenAnnualKeys, domain.InputKey(ti.Key)) {
			return fmt.Errorf("%s %s: %w", r.Regime, ti.Key, errGoldenUnknownTaxInput)
		}
		values = append(values, domain.TaxInputValue{Key: domain.InputKey(ti.Key), Kind: domain.InputMoney, Money: ti.Money})
	}
	for _, pay := range payments {
		if pay.Regime != r.Regime {
			continue
		}
		byMonth := goldenMonths(pay)
		switch pay.Source {
		case annualgolden.PaymentFiled:
			p.filed[r.Regime] = byMonth
		case annualgolden.PaymentEstimatedByApp:
			p.app[r.Regime] = byMonth
		case annualgolden.PaymentTaxInput:
			for m, v := range byMonth {
				values = append(values, domain.TaxInputValue{Key: domain.InputPriorProvisionalPayments, Kind: domain.InputMoneyMonthly, Month: m, Money: v})
			}
		}
	}
	in, err := domain.NewTaxInputs(values...)
	if err != nil {
		return fmt.Errorf("%s: %w", r.Regime, err)
	}
	p.inputs[r.Regime] = in
	return nil
}

// goldenMonths spreads a payment over the year: its month (December for a
// year total), zero in every other month, so every month has its source.
func goldenMonths(pay annualgolden.Payment) map[time.Month]money.Cents {
	at := time.Month(pay.Month)
	if at == 0 {
		at = time.December
	}
	out := map[time.Month]money.Cents{}
	for m := time.January; m <= time.December; m++ {
		out[m] = 0
	}
	out[at] = pay.Amount
	return out
}

func goldenOptions(raw map[string]string) (domain.Options, error) {
	options := map[domain.OptionKey]domain.OptionValue{}
	for k, v := range raw {
		key := domain.OptionKey(k)
		if key != annual.OptionPlatformDefinitive {
			options[key] = domain.TextOption(v)
			continue
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			return domain.Options{}, err
		}
		options[key] = domain.BoolOption(b)
	}
	return domain.NewOptions(options), nil
}

func (p *goldenPorts) Profile(context.Context, string, string) (usecase.TaxpayerProfile, error) {
	return p.profile, nil
}

func (p *goldenPorts) IncomeSums(context.Context, usecase.LedgerQuery) ([]usecase.IncomeMonthSums, error) {
	return p.incomes, nil
}

func (p *goldenPorts) ExpenseSums(context.Context, usecase.LedgerQuery) ([]usecase.ExpenseMonthSums, error) {
	return p.expenses, nil
}

func (p *goldenPorts) TaxInputs(context.Context, string, int, int) (map[int]map[fiscal.RegimeCode]domain.TaxInputs, error) {
	return map[int]map[fiscal.RegimeCode]domain.TaxInputs{p.period: p.inputs}, nil
}

func (p *goldenPorts) YearInputs(context.Context, string, int) (usecaseannual.YearInputs, error) {
	return p.year, nil
}

func (p *goldenPorts) FiledPayments(context.Context, string, int) (map[fiscal.RegimeCode]map[time.Month]money.Cents, error) {
	return p.filed, nil
}

func (p *goldenPorts) Holidays(ctx context.Context) (fiscal.HolidaySet, error) {
	return seededHolidays{}.Holidays(ctx)
}

// EstimateRange answers the app's estimates the case gives (zero for a
// regime without any), month by month.
func (p *goldenPorts) EstimateRange(_ context.Context, _, _ string, from, to domain.Period) ([]usecase.TaxpayerEstimate, error) {
	var out []usecase.TaxpayerEstimate
	for m := from.Month(); m <= to.Month(); m++ {
		period, err := domain.NewPeriod(from.Year(), m)
		if err != nil {
			return nil, err
		}
		est := usecase.TaxpayerEstimate{Period: period}
		for _, h := range p.profile.Regimes {
			isr := domain.TaxResult{Amount: p.app[h.Code][m]}
			est.Regimes = append(est.Regimes, usecase.RegimeResult{RegimeEstimate: domain.RegimeEstimate{
				Key: domain.StrategyKey{Regime: h.Code, PersonType: p.profile.PersonType}, ISR: &isr,
			}})
		}
		out = append(out, est)
	}
	return out, nil
}

// TestGoldenAnnualAdapterRefusesAnUnknownTaxInput: a case setting a tax
// input the annual does not read fails instead of passing on a figure the
// sections never saw.
func TestGoldenAnnualAdapterRefusesAnUnknownTaxInput(t *testing.T) {
	_, err := goldenPortsOf(annualgolden.Input{FiscalYear: 2026, Regimes: []annualgolden.RegimeInput{{
		Regime: "601", TaxInputs: []annualgolden.TaxInput{{Key: "coeficiente_utilidad"}},
	}}})

	require.ErrorIs(t, err, errGoldenUnknownTaxInput)
}

// TestGoldenAnnualAdapterReadsEveryInput maps a full PM regime through the
// use case, so a dropped tax input or payment shows as another figure.
func TestGoldenAnnualAdapterReadsEveryInput(t *testing.T) {
	in := annualgolden.Input{
		FiscalYear: 2026, PersonType: fiscal.PersonMoral,
		Regimes: []annualgolden.RegimeInput{{
			Regime: "601", Income: 1_000_000_00, Deductions: 500_000_00,
			TaxInputs: []annualgolden.TaxInput{
				{Key: "annual_depreciation", Money: 100_000_00}, {Key: "ptu_paid", Money: 50_000_00},
				{Key: "prior_year_losses", Money: 10_000_00}, {Key: "interest_withholding", Money: 1_000_00},
			},
		}},
		Payments: []annualgolden.Payment{{Regime: "601", Month: 3, Amount: 20_000_00, Source: annualgolden.PaymentEstimatedByApp}},
	}

	got, err := newUseCaseAnnual(t).Estimate(in)

	require.NoError(t, err)
	require.Len(t, got.Sections, 1)
	// 1,000,000 − 500,000 − 100,000 − 50,000 − 10,000 = 340,000 × 30 % =
	// 102,000 − 20,000 − 1,000.
	assert.Equal(t, money.Cents(340_000_00), got.Sections[0].Tax.Base)
	assert.Equal(t, annualgolden.Balance{Kind: annual.BalanceACargo, Amount: 81_000_00}, got.Sections[0].Tax.Balance)
}
