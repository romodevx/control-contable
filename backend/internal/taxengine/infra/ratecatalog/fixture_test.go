package ratecatalog_test

import (
	"fmt"
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

// syntheticSet is a minimal valid set of fiscalYear holding the general IVA
// rate (open-ended) and a border rate that lapses on December 31.
func syntheticSet(t testing.TB, fiscalYear int, provisional bool) *taxparams.Params {
	t.Helper()
	first := date(t, fmtDate(fiscalYear, "01-01"))
	last := date(t, fmtDate(fiscalYear, "12-31"))
	basis := []taxparams.Citation{{Source: syntheticSource, Article: "Art. 1"}}
	set, err := taxparams.NewParams(taxparams.Spec{
		FiscalYear:  fiscalYear,
		Version:     1,
		PublishedOn: first,
		Provisional: provisional,
		Sources: []taxparams.Source{{
			ID: syntheticSource, Title: "Synthetic", URL: "https://example.org/dof",
			DOFDate: first, Reference: "Art. 1",
		}},
		Values: []taxparams.Value{
			{
				Key: taxparams.Key(fiscal.KeyIVAGeneral), Kind: taxparams.KindRate, Rate: rate(t, "0.16"),
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
		},
	})
	require.NoError(t, err)
	return set
}

func fmtDate(year int, monthDay string) string { return fmt.Sprintf("%04d-%s", year, monthDay) }

// synthetic is the adapter over the given sets.
func synthetic(t testing.TB, sets ...*taxparams.Params) *catalog {
	t.Helper()
	c, err := taxparams.NewCatalog(sets...)
	require.NoError(t, err)
	rc, err := ratecatalog.New[expenses.AmountKey](c)
	require.NoError(t, err)
	return rc
}
