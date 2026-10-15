package params_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/money"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain/taxparams"
	"github.com/romodevx/control-contable/backend/internal/taxengine/params"
)

// The annual-return parameters (ANN-04). As in embedded_2026_test.go, the
// expected values are copied from the official texts (LISR 76, 86, 98, 150,
// 151 as published by the Cámara de Diputados, last reform DOF 01-04-2024;
// decree DOF 26-12-2013 art. 1.10; Anexo 8 RMF 2026 C.II; INEGI UMA notice
// DOF 09-01-2026), never from the YAML file.

// level stands in for annual.TuitionLevel (PR #127).
type level string

func TestEmbedded2026DefinesTheAnnualKeys(t *testing.T) {
	p := embedded2026(t)
	reqs := taxparams.DeductionLimitRequirements()
	reqs = append(reqs,
		taxparams.RequireMoney(taxparams.KeySalaryAnnualThreshold),
		taxparams.RequireMoney(taxparams.KeyInformativeIncomeThreshold),
		taxparams.RequireDate(taxparams.KeyDeadlinePF),
		taxparams.RequireDate(taxparams.KeyDeadlinePM),
		taxparams.RequireDate(taxparams.KeyDeadlineNonprofitInformative),
		taxparams.RequireTable("isr.art152.annual"),
		taxparams.RequireTable("isr.resico_pf.annual"),
		taxparams.RequireRate("isr.rate.persona_moral"),
	)
	require.NoError(t, p.Require(reqs...))
}

func TestEmbedded2026DeductionLimits(t *testing.T) {
	catalog, err := params.Embedded()
	require.NoError(t, err)
	got, err := taxparams.ReadDeductionLimits[level](catalog, 2026)
	require.NoError(t, err)

	assert.Equal(t, cents(t, "42,794.64"), got.AnnualUMA, "UMA valor anual from 1 Feb 2026")
	assert.Equal(t, int64(5), got.GlobalCapUMAs, "151 last ¶: cinco veces el valor anual de la UMA")
	assert.Equal(t, rate(t, "0.15"), got.GlobalCapIncomeShare, "151 last ¶: 15% del total de los ingresos")
	assert.Equal(t, int64(1), got.FuneralCapUMAs, "151 fr. II: salario mínimo elevado al año")
	assert.Equal(t, rate(t, "0.07"), got.DonationShare, "151 fr. III: 7% de los ingresos acumulables")
	assert.Equal(t, rate(t, "0.04"), got.GovernmentDonationShare, "151 fr. III: no podrá exceder del 4%")
	assert.Equal(t, rate(t, "0.10"), got.RetirementIncomeShare, "151 fr. V: hasta el 10%")
	assert.Equal(t, int64(5), got.RetirementCapUMAs, "151 fr. V: cinco salarios mínimos elevados al año")
	// Decree DOF 26-12-2013, art. 1.10 table "Límite anual de deducción".
	assert.Equal(t, map[level]money.Cents{
		"preescolar":          cents(t, "14,200.00"),
		"primaria":            cents(t, "12,900.00"),
		"secundaria":          cents(t, "19,900.00"),
		"profesional_tecnico": cents(t, "17,100.00"),
		"bachillerato":        cents(t, "24,500.00"),
	}, got.TuitionCaps)
}

func TestEmbedded2026AnnualThresholds(t *testing.T) {
	p := embedded2026(t)
	tests := []struct {
		key  taxparams.Key
		want string
	}{
		{taxparams.KeySalaryAnnualThreshold, "400,000.00"},      // LISR 98 fr. III e)
		{taxparams.KeyInformativeIncomeThreshold, "500,000.00"}, // LISR 150 ¶3
	}
	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			got, err := p.Money(tt.key, date(t, "2026-12-31"))
			require.NoError(t, err)
			assert.Equal(t, cents(t, tt.want), got)
		})
	}
}

// The deadlines belong to fiscal year 2026: in force through its last day,
// absent for 2027 until that year is transcribed (a 2027 estimate on the
// fallback set must not show a 2027-04-30 deadline for the wrong year).
func TestEmbedded2026AnnualDeadlines(t *testing.T) {
	p := embedded2026(t)
	tests := []struct {
		key  taxparams.Key
		want string
	}{
		{taxparams.KeyDeadlinePF, "2027-04-30"},                   // LISR 150: "en el mes de abril del año siguiente"
		{taxparams.KeyDeadlinePM, "2027-03-31"},                   // LISR 76 fr. V: "dentro de los tres meses siguientes"
		{taxparams.KeyDeadlineNonprofitInformative, "2027-02-15"}, // LISR 86 fr. III: "a más tardar el día 15 de febrero"
	}
	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			got, err := p.Date(tt.key, date(t, "2026-12-31"))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.String())
			_, err = p.Date(tt.key, date(t, "2027-01-01"))
			require.ErrorIs(t, err, taxparams.ErrNotInForce)
		})
	}
}

// Anexo 8 C.II, the 2026 annual tariff (arts. 97 and 152), spot rows at both
// ends and the middle, as printed in the DOF edition of 28-12-2025.
func TestEmbedded2026AnnualTariffSpotRows(t *testing.T) {
	annual, err := embedded2026(t).Table("isr.art152.annual", date(t, "2026-12-31"))
	require.NoError(t, err)
	require.Equal(t, taxparams.TableMarginal, annual.Kind)
	require.Len(t, annual.Rows, 11)
	assertRow(t, annual.Rows[0], row{"0.01", "10,135.11", "0.00", "1.92"})
	assertRow(t, annual.Rows[1], row{"10,135.12", "86,022.11", "194.59", "6.40"})
	assertRow(t, annual.Rows[5], row{"210,403.70", "424,353.97", "22,282.14", "21.36"})
	assertRow(t, annual.Rows[9], row{"1,702,567.98", "5,107,703.92", "444,116.23", "34.00"})
	assertRow(t, annual.Rows[10], row{"5,107,703.93", "", "1,601,862.46", "35.00"})
}
