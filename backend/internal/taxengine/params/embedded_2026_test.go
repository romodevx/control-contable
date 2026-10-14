package params_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

// The expected values below are copied from the official publications cited
// in data/2026/v1.yaml (Anexo 8 RMF 2026, DOF 28-12-2025; LISR; LIVA; RLIVA;
// LIF 2026; DOF UMA notices), not from the YAML file, so a transcription slip
// in either place fails here.

func embedded2026(t *testing.T) *taxparams.Params {
	t.Helper()
	catalog, err := params.Embedded()
	require.NoError(t, err)
	sel, err := catalog.Select(2026)
	require.NoError(t, err)
	require.Equal(t, 2026, sel.Params.FiscalYear())
	return sel.Params
}

func cents(t testing.TB, s string) money.Cents {
	t.Helper()
	c, err := money.ParseCents(strings.ReplaceAll(s, ",", "")) // as printed: 1,011.68
	require.NoError(t, err)
	return c
}

func rate(t testing.TB, s string) money.Rate {
	t.Helper()
	r, err := money.ParseRate(s)
	require.NoError(t, err)
	return r
}

// row is a published tariff row: lower, upper ("" = "En adelante"), fixed
// fee, and the percentage as printed.
type row struct{ lower, upper, fee, percent string }

func assertRow(t *testing.T, got taxparams.Row, want row) {
	t.Helper()
	assert.Equal(t, cents(t, want.lower), got.Lower, "lower")
	if want.upper == "" {
		assert.True(t, got.Unbounded, "En adelante")
	} else {
		assert.False(t, got.Unbounded)
		assert.Equal(t, cents(t, want.upper), got.Upper, "upper")
	}
	assert.Equal(t, cents(t, want.fee), got.FixedFee, "fixed fee")
	pct, err := money.ParseRate(want.percent)
	require.NoError(t, err)
	assert.Equal(t, pct/100, got.Rate, "rate")
}

func TestEmbeddedSelects2026v1(t *testing.T) {
	catalog, err := params.Embedded()
	require.NoError(t, err)
	assert.Equal(t, []taxparams.SetID{{FiscalYear: 2026, Version: 1}}, catalog.Sets())

	sel, err := catalog.Select(2026)
	require.NoError(t, err)
	assert.False(t, sel.Provisional, "2026/v1 is transcribed from official sources")
	assert.False(t, sel.Fallback)

	next, err := catalog.Select(2027)
	require.NoError(t, err)
	assert.True(t, next.Fallback, "2027 falls back to 2026 until it is transcribed")
}

// Every key a regime strategy or another module reads must exist with the
// right shape: the boot check (Params.Require) depends on it.
func TestEmbedded2026DefinesEveryKeyTheModulesRead(t *testing.T) {
	p := embedded2026(t)
	rates := []taxparams.Key{
		taxparams.Key(fiscal.KeyIVAGeneral), taxparams.Key(fiscal.KeyIVABorder), taxparams.Key(fiscal.KeyIVAZero),
		taxparams.Key(fiscal.KeyISRWithholdingFees), taxparams.Key(fiscal.KeyISRWithholdingRent),
		// incomes/domain/catalog.go and expenses/domain/catalog.go
		"isr.withholding.resico_pf", "isr.platform.transport_delivery", "isr.platform.lodging",
		"isr.platform.goods_services", "isr.platform.no_rfc", "isr.platform.persona_moral",
		"iva.withholding.freight_pm", "iva.withholding.scrap_pm", // MODULE-withholdings §3 (WH-05)
		"isr.platform.persona_moral_no_rfc", "isr.rate.persona_moral", "isr.withholding.interest_annual",
		"isr.rent.blind_deduction",
	}
	amounts := []taxparams.Key{
		"deduction.cash_payment_limit", // expenses/domain/catalog.go
		"uma.daily", "uma.monthly", "uma.annual",
		"regime.resico_pf.income_cap", "regime.resico_pm.income_cap", "regime.platform.definitive_option_cap",
	}
	tables := []taxparams.Key{
		"isr.art96.monthly", "isr.art116.monthly", "isr.art152.annual", "isr.rif.bimonthly",
		"isr.resico_pf.monthly", "isr.resico_pf.annual",
	}
	var reqs []taxparams.Requirement
	for _, k := range rates {
		reqs = append(reqs, taxparams.RequireRate(k))
	}
	for _, k := range amounts {
		reqs = append(reqs, taxparams.RequireMoney(k))
	}
	for _, k := range tables {
		reqs = append(reqs, taxparams.RequireTable(k))
	}
	reqs = append(reqs, taxparams.RequireMonthlyTables("isr.art106.accumulated"))
	require.NoError(t, p.Require(reqs...))
}

func TestEmbedded2026EverySourceIsOfficial(t *testing.T) {
	official := []string{
		"https://www.sat.gob.mx/", "https://dof.gob.mx/", "https://www.diputados.gob.mx/", "https://www.inegi.org.mx/",
	}
	for _, s := range embedded2026(t).Sources() {
		ok := false
		for _, prefix := range official {
			ok = ok || strings.HasPrefix(s.URL, prefix)
		}
		assert.True(t, ok, "%s: %s is not an official host", s.ID, s.URL)
	}
}

func TestEmbedded2026Art96MonthlyTariff(t *testing.T) {
	tariff, err := embedded2026(t).Table("isr.art96.monthly", date(t, "2026-06-30"))
	require.NoError(t, err)
	require.Equal(t, taxparams.TableMarginal, tariff.Kind)
	require.Len(t, tariff.Rows, 11)
	// Anexo 8 B.V
	assertRow(t, tariff.Rows[0], row{"0.01", "844.59", "0.00", "1.92"})
	assertRow(t, tariff.Rows[3], row{"12,598.03", "14,644.64", "1,011.68", "16.00"})
	assertRow(t, tariff.Rows[7], row{"55,736.69", "106,410.50", "10,457.09", "30.00"})
	assertRow(t, tariff.Rows[10], row{"425,642.00", "", "133,488.54", "35.00"})
}

func TestEmbedded2026Art106AccumulatedTariffs(t *testing.T) {
	p := embedded2026(t)
	// Anexo 8 B.VI, one spot row or more per published month.
	tests := []struct {
		month time.Month
		index int
		want  row
	}{
		{time.January, 1, row{"844.60", "7,168.51", "16.22", "6.40"}},
		{time.February, 10, row{"851,283.99", "", "266,977.08", "35.00"}},
		{time.March, 5, row{"52,600.93", "106,088.49", "5,570.52", "21.36"}},
		{time.June, 8, row{"638,463.01", "851,283.96", "153,955.38", "32.00"}},
		{time.October, 5, row{"175,336.41", "353,628.30", "18,568.40", "21.36"}},
		// December is published, not 12 x January: 194.59, not 194.64.
		{time.December, 1, row{"10,135.12", "86,022.11", "194.59", "6.40"}},
		{time.December, 10, row{"5,107,703.93", "", "1,601,862.46", "35.00"}},
	}
	for _, tt := range tests {
		t.Run(tt.month.String(), func(t *testing.T) {
			tariff, err := p.MonthlyTable("isr.art106.accumulated", tt.month, date(t, "2026-12-31"))
			require.NoError(t, err)
			require.Len(t, tariff.Rows, 11)
			assertRow(t, tariff.Rows[tt.index], tt.want)
		})
	}
}

func TestEmbedded2026AnnualTariffEqualsTheDecemberTable(t *testing.T) {
	p := embedded2026(t)
	annual, err := p.Table("isr.art152.annual", date(t, "2026-12-31"))
	require.NoError(t, err)
	december, err := p.MonthlyTable("isr.art106.accumulated", time.December, date(t, "2026-12-31"))
	require.NoError(t, err)
	assert.Equal(t, december.Rows, annual.Rows)
	// Anexo 8 C.II
	assertRow(t, annual.Rows[6], row{"424,353.98", "668,840.14", "67,981.92", "23.52"})
}

func TestEmbedded2026RentAndRIFTariffs(t *testing.T) {
	p := embedded2026(t)
	on := date(t, "2026-06-30")
	monthly, err := p.Table("isr.art96.monthly", on)
	require.NoError(t, err)
	rent, err := p.Table("isr.art116.monthly", on)
	require.NoError(t, err)
	assert.Equal(t, monthly.Rows, rent.Rows, "the art. 116 table is published identical to B.V")

	rif, err := p.Table("isr.rif.bimonthly", on)
	require.NoError(t, err)
	february, err := p.MonthlyTable("isr.art106.accumulated", time.February, on)
	require.NoError(t, err)
	assert.Equal(t, february.Rows, rif.Rows, "B.VII is published identical to the February table")
	assertRow(t, rif.Rows[5], row{"35,067.29", "70,725.66", "3,713.68", "21.36"})
}

func TestEmbedded2026RESICOPFTables(t *testing.T) {
	p := embedded2026(t)
	monthly, err := p.Table("isr.resico_pf.monthly", date(t, "2026-06-30"))
	require.NoError(t, err)
	assert.Equal(t, taxparams.TableFlat, monthly.Kind)
	// LISR 113-E, TABLA MENSUAL ("Hasta" amounts; each lower is the previous upper + 0.01)
	want := []struct{ upper, rate string }{
		{"25000.00", "0.01"}, {"50000.00", "0.011"}, {"83333.33", "0.015"},
		{"208333.33", "0.02"}, {"3500000.00", "0.025"},
	}
	require.Len(t, monthly.Rows, len(want))
	for i, w := range want {
		assert.Equal(t, cents(t, w.upper), monthly.Rows[i].Upper, "row %d", i)
		assert.Equal(t, rate(t, w.rate), monthly.Rows[i].Rate, "row %d", i)
	}
	assert.Equal(t, cents(t, "25000.01"), monthly.Rows[1].Lower)

	annual, err := p.Table("isr.resico_pf.annual", date(t, "2026-06-30"))
	require.NoError(t, err)
	// LISR 113-F, TABLA ANUAL
	assert.Equal(t, cents(t, "300000.00"), annual.Rows[0].Upper)
	assert.Equal(t, cents(t, "1000000.00"), annual.Rows[2].Upper)
	assert.Equal(t, rate(t, "0.015"), annual.Rows[2].Rate)
}

func TestEmbedded2026Rates(t *testing.T) {
	p := embedded2026(t)
	tests := []struct {
		key  taxparams.Key
		on   string
		want string
	}{
		{"iva.rate.general_16", "2026-06-30", "0.16"},         // LIVA 1
		{"iva.rate.zero", "2026-06-30", "0"},                  // LIVA 2-A
		{"iva.rate.border_8", "2026-01-01", "0.08"},           // 16 % less the 50 % credit
		{"iva.rate.border_8", "2026-12-31", "0.08"},           // decree in force until 31-12-2026
		{"iva.withholding.freight_pm", "2026-06-30", "0.04"},  // RLIVA 3 fr. II
		{"iva.withholding.scrap_pm", "2026-06-30", "1"},       // LIVA 1-A fr. II b), no RLIVA 3 reduction
		{"isr.rate.persona_moral", "2026-06-30", "0.30"},      // LISR 9
		{"isr.withholding.fees_pm", "2026-06-30", "0.10"},     // LISR 106
		{"isr.withholding.rent_pm", "2026-06-30", "0.10"},     // LISR 116
		{"isr.withholding.resico_pf", "2026-06-30", "0.0125"}, // LISR 113-J
		{"isr.withholding.interest_annual", "2026-06-30", "0.009"},
		{"isr.platform.transport_delivery", "2026-06-30", "0.021"}, // LISR 113-A fr. I
		{"isr.platform.lodging", "2026-06-30", "0.04"},             // LISR 113-A fr. II
		{"isr.platform.goods_services", "2026-06-30", "0.025"},     // LIF 2026 25-VI
		{"isr.platform.no_rfc", "2026-06-30", "0.20"},              // LISR 113-C fr. IV
		{"isr.platform.persona_moral", "2026-06-30", "0.025"},      // LIF 2026 25-VI
		{"isr.platform.persona_moral_no_rfc", "2026-06-30", "0.20"},
		{"isr.rent.blind_deduction", "2026-06-30", "0.35"}, // LISR 115
	}
	for _, tt := range tests {
		t.Run(string(tt.key)+"@"+tt.on, func(t *testing.T) {
			got, err := p.Rate(tt.key, date(t, tt.on))
			require.NoError(t, err)
			assert.Equal(t, rate(t, tt.want), got)
		})
	}
}

// The withholding rates MODULE-withholdings §3 derives from (WH-05) start
// with the set on January 1, still hold on December 31, and are absent the
// day before: a record dated outside a value's validity never gets a silent
// rate.
func TestEmbedded2026WithholdingRatesAtTheYearBoundaries(t *testing.T) {
	p := embedded2026(t)
	tests := []struct {
		key  taxparams.Key
		want string
	}{
		{"isr.withholding.fees_pm", "0.10"},     // LISR 106, honorarios
		{"isr.withholding.rent_pm", "0.10"},     // LISR 116, arrendamiento
		{"isr.withholding.resico_pf", "0.0125"}, // LISR 113-J
		{"iva.withholding.freight_pm", "0.04"},  // RLIVA 3 fr. II, of the price
		{"iva.withholding.scrap_pm", "1"},       // LIVA 1-A fr. II b), of the IVA
	}
	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			for _, on := range []string{"2026-01-01", "2026-12-31"} {
				got, err := p.Rate(tt.key, date(t, on))
				require.NoError(t, err, on)
				assert.Equal(t, rate(t, tt.want), got, on)
			}
			_, err := p.Rate(tt.key, date(t, "2025-12-31"))
			require.ErrorIs(t, err, taxparams.ErrNotInForce)
		})
	}
}

// Rates set by the annual LIF or by a decree with a sunset lapse with it.
func TestEmbedded2026RatesThatLapseAtYearEnd(t *testing.T) {
	p := embedded2026(t)
	for _, key := range []taxparams.Key{
		"iva.rate.border_8", "isr.platform.goods_services", "isr.platform.persona_moral",
		"isr.platform.persona_moral_no_rfc", "isr.withholding.interest_annual",
	} {
		_, err := p.Rate(key, date(t, "2027-01-01"))
		require.ErrorIs(t, err, taxparams.ErrNotInForce, key)
	}
}

func TestEmbedded2026UMA(t *testing.T) {
	p := embedded2026(t)
	tests := []struct {
		key  taxparams.Key
		on   string
		want string
	}{
		// January 2026 still uses the UMA published DOF 10-01-2025.
		{"uma.daily", "2026-01-31", "113.14"},
		{"uma.monthly", "2026-01-31", "3439.46"},
		{"uma.annual", "2026-01-31", "41273.52"},
		// From 1 February 2026, DOF 09-01-2026.
		{"uma.daily", "2026-02-01", "117.31"},
		{"uma.monthly", "2026-02-01", "3566.22"},
		{"uma.annual", "2026-12-31", "42794.64"},
	}
	for _, tt := range tests {
		t.Run(string(tt.key)+"@"+tt.on, func(t *testing.T) {
			got, err := p.Money(tt.key, date(t, tt.on))
			require.NoError(t, err)
			assert.Equal(t, cents(t, tt.want), got)
		})
	}
}

func TestEmbedded2026Thresholds(t *testing.T) {
	p := embedded2026(t)
	tests := []struct {
		key  taxparams.Key
		want string
	}{
		{"regime.resico_pf.income_cap", "3500000.00"},          // LISR 113-E
		{"regime.resico_pm.income_cap", "35000000.00"},         // LISR 206
		{"regime.platform.definitive_option_cap", "300000.00"}, // LISR 113-B fr. I
		{"deduction.cash_payment_limit", "2000.00"},            // LISR 27 fr. III
	}
	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			got, err := p.Money(tt.key, date(t, "2026-06-30"))
			require.NoError(t, err)
			assert.Equal(t, cents(t, tt.want), got)
		})
	}
}
