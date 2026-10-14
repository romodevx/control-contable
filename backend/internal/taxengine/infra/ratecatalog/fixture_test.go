package ratecatalog_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	expenses "github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/infra/ratecatalog"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

// catalog is the adapter as the composition root instantiates it: with the
// expenses module's amount key, so one value serves both consumer ports.
type catalog = ratecatalog.Catalog[expenses.AmountKey]

func date(t testing.TB, raw string) fiscal.Date {
	t.Helper()
	d, err := fiscal.ParseDate(raw)
	require.NoError(t, err)
	return d
}

func rate(t testing.TB, raw string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(raw)
	require.NoError(t, err)
	return r
}

func cents(t testing.TB, raw string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(raw)
	require.NoError(t, err)
	return c
}

// embedded is the adapter over the parameter files compiled into the binary
// (params/data/2026/v1.yaml): the values every test below expects are the
// verified 2026 ones.
func embedded(t testing.TB) *catalog {
	t.Helper()
	sets, err := params.Embedded()
	require.NoError(t, err)
	c, err := ratecatalog.New[expenses.AmountKey](sets)
	require.NoError(t, err)
	return c
}

const syntheticSource = "dof-synthetic"

// syntheticWithholdingRates are the withholding rates every set must hold
// (WH-05), open-ended. The values differ from the real ones on purpose, so a
// test reading them proves it went through the synthetic set.
var syntheticWithholdingRates = map[fiscal.RateKey]string{
	fiscal.KeyISRWithholdingFees:         "0.11",
	fiscal.KeyISRWithholdingRent:         "0.12",
	ratecatalog.KeyISRWithholdingRESICO:  "0.013",
	ratecatalog.KeyIVAWithholdingFreight: "0.05",
	ratecatalog.KeyIVAWithholdingScrap:   "0.9",
}

// syntheticSet is a minimal valid set of fiscalYear holding the IVA rates
// every set must hold (general 16 % and zero, open-ended), a border rate that
// lapses on December 31, the cash payment limit and the withholding rates.
func syntheticSet(t testing.TB, fiscalYear int, provisional bool) *taxparams.Params {
	t.Helper()
	return syntheticSetWithout(t, fiscalYear, provisional)
}

// syntheticSetWithout is syntheticSet without the values of omit.
func syntheticSetWithout(t testing.TB, fiscalYear int, provisional bool, omit ...taxparams.Key) *taxparams.Params {
	t.Helper()
	return syntheticVersion(t, fiscalYear, 1, provisional, omit...)
}

// syntheticVersion is version of syntheticSetWithout.
func syntheticVersion(t testing.TB, fiscalYear, version int, provisional bool, omit ...taxparams.Key) *taxparams.Params {
	t.Helper()
	values := syntheticValues(t, fiscalYear)
	values = slices.DeleteFunc(values, func(v taxparams.Value) bool { return slices.Contains(omit, v.Key) })
	return newSyntheticParams(t, fiscalYear, version, provisional, values)
}

// syntheticSetStartingOn is syntheticSet with key in force only from day
// from on, as when a reform starts a rate mid-year.
func syntheticSetStartingOn(t testing.TB, fiscalYear int, key fiscal.RateKey, from fiscal.Date) *taxparams.Params {
	t.Helper()
	values := syntheticValues(t, fiscalYear)
	for i := range values {
		if values[i].Key == taxparams.Key(key) {
			values[i].Validity = taxparams.Validity{From: from}
		}
	}
	return newSyntheticParams(t, fiscalYear, 1, false, values)
}

// syntheticValues are the values of syntheticSet.
func syntheticValues(t testing.TB, fiscalYear int) []taxparams.Value {
	t.Helper()
	first := date(t, fmtDate(fiscalYear, "01-01"))
	last := date(t, fmtDate(fiscalYear, "12-31"))
	basis := []taxparams.Citation{{Source: syntheticSource, Article: "Art. 1"}}
	values := []taxparams.Value{
		{
			Key: taxparams.Key(fiscal.KeyIVAGeneral), Kind: taxparams.KindRate, Rate: rate(t, "0.16"),
			Validity: taxparams.Validity{From: first}, Basis: basis,
		},
		{
			Key: taxparams.Key(fiscal.KeyIVAZero), Kind: taxparams.KindRate, Rate: rate(t, "0"),
			Validity: taxparams.Validity{From: first}, Basis: basis,
		},
		{
			Key: taxparams.Key(fiscal.KeyIVABorder), Kind: taxparams.KindRate, Rate: rate(t, "0.08"),
			Validity: taxparams.Validity{From: first, To: last}, Basis: basis,
		},
		{
			Key: taxparams.Key(expenses.KeyCashPaymentLimit), Kind: taxparams.KindMoney, Money: cents(t, "2000.00"),
			Validity: taxparams.Validity{From: first}, Basis: basis,
		},
	}
	for _, key := range slices.Sorted(maps.Keys(syntheticWithholdingRates)) {
		values = append(values, taxparams.Value{
			Key: taxparams.Key(key), Kind: taxparams.KindRate, Rate: rate(t, syntheticWithholdingRates[key]),
			Validity: taxparams.Validity{From: first}, Basis: basis,
		})
	}
	return values
}

func newSyntheticParams(t testing.TB, fiscalYear, version int, provisional bool, values []taxparams.Value) *taxparams.Params {
	t.Helper()
	first := date(t, fmtDate(fiscalYear, "01-01"))
	set, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear:  fiscalYear,
		Version:     version,
		PublishedOn: first,
		Provisional: provisional,
		Sources: []taxparams.Source{{
			ID: syntheticSource, Title: "Synthetic", URL: "https://example.org/dof",
			DOFDate: first, Reference: "Art. 1",
		}},
		Values: values,
	})
	require.NoError(t, err)
	return set
}

func fmtDate(year int, monthDay string) string { return fmt.Sprintf("%04d-%s", year, monthDay) }

// synthetic is the adapter over the given sets.
func synthetic(t testing.TB, sets ...*taxparams.Params) *catalog {
	t.Helper()
	rc, err := newSynthetic(t, sets...)
	require.NoError(t, err)
	return rc
}

// newSynthetic is ratecatalog.New over the given sets, with its error.
func newSynthetic(t testing.TB, sets ...*taxparams.Params) (*catalog, error) {
	t.Helper()
	c, err := taxparams.NewCatalog(sets...)
	require.NoError(t, err)
	return ratecatalog.New[expenses.AmountKey](c)
}
