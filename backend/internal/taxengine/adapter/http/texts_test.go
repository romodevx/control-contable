package taxenginehttp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func march(t *testing.T) domain.Period {
	t.Helper()
	p, err := domain.ParsePeriod("2026-03")
	require.NoError(t, err)
	return p
}

// The concepts of the contract's crossRegimeIva612 example, from the 612
// strategy's labels and args.
func TestConcept_rendersTheContractExample(t *testing.T) {
	tests := []struct {
		name string
		line domain.Line
		want string
	}{
		{"YTD income", domain.Line{Label: "isr.income_ytd", Args: []string{"2026-01", "2026-03"}}, "Ingresos cobrados acumulados ene–mar"},
		{"YTD deductions", domain.Line{Label: "isr.deductions_ytd", Args: []string{"2026-01", "2026-03"}}, "Deducciones pagadas acumuladas ene–mar"},
		{"base", domain.Line{Label: "isr.base"}, "Base gravable"},
		{"tariff row", domain.Line{Label: "isr.tariff_row", Args: []string{"2026-03"}, Row: 6}, "Cuota fija (tarifa acumulada de marzo, renglón 6)"},
		{"excess", domain.Line{Label: "isr.tariff_excess"}, "Excedente del límite inferior por la tasa"},
		{"caused", domain.Line{Label: "isr.caused"}, "ISR causado acumulado"},
		{"prior payments", domain.Line{Label: "isr.prior_payments"}, "Pagos provisionales anteriores (ene–feb)"},
		{"withheld", domain.Line{Label: "isr.withheld_by_moral"}, "Retenciones de ISR (10 %) acumuladas"},
		{"payable", domain.Line{Label: "isr.payable"}, "ISR a pagar"},
		{"IVA charged", domain.Line{Label: "iva.charged"}, "IVA trasladado cobrado"},
		{"IVA creditable", domain.Line{Label: "iva.creditable"}, "IVA acreditable pagado"},
		{"IVA withheld", domain.Line{Label: "iva.withheld.moral"}, "IVA que te retuvieron personas morales"},
		{"regime IVA", domain.Line{Label: "iva.regime_net"}, "IVA de este régimen"},
		{"consolidated", domain.Line{Label: "iva.consolidated"}, "IVA a cargo del mes (todos tus regímenes)"},
		{"rounding", domain.Line{Label: domain.LabelRoundingAdjustment}, "Ajuste por redondeo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, taxenginehttp.Concept(fiscal.RegimeBusiness, march(t), tt.line))
		})
	}
}

// The same step reads per regime; args that do not parse fall back to the
// plain concept; an unknown label never leaks its key.
func TestConcept_variants(t *testing.T) {
	p := march(t)
	assert.Equal(t, "Cuota fija (tarifa mensual, renglón 4)", taxenginehttp.Concept(fiscal.RegimeRent, p,
		domain.Line{Label: "isr.tariff_row", Row: 4}))
	assert.Equal(t, "Cuota fija (tarifa)", taxenginehttp.Concept("601", p, domain.Line{Label: "isr.tariff_row"}))
	assert.Equal(t, "Retención de ISR (1.25 %) de personas morales", taxenginehttp.Concept(fiscal.RegimeRESICO, p,
		domain.Line{Label: "isr.withheld_by_moral"}))
	assert.Equal(t, "ISR que te retuvieron personas morales", taxenginehttp.Concept("601", p, domain.Line{Label: "isr.withheld_by_moral"}))
	assert.Equal(t, "Deducción ciega (35 % de las rentas)", taxenginehttp.Concept(fiscal.RegimeRent, p,
		domain.Line{Label: "isr.blind_deduction", Args: []string{"0.35"}}))
	assert.Equal(t, "Rentas cobradas en marzo", taxenginehttp.Concept(fiscal.RegimeRent, p,
		domain.Line{Label: "isr.rent_month", Args: []string{"2026-03"}}))
	assert.Equal(t, "Reducción del RIF (año 3, 70 %)", taxenginehttp.Concept("621", p,
		domain.Line{Label: "isr.rif_reduction", Args: []string{"3", "0.7"}}))
	assert.Equal(t, "El RIF se paga por bimestre; consulta abril", taxenginehttp.Concept("621", p,
		domain.Line{Label: "isr.rif_bimester_pending", Args: []string{"2026-04"}}))
	assert.Equal(t, "Ingresos cobrados del bimestre mar–abr", taxenginehttp.Concept("621", p,
		domain.Line{Label: "isr.income_bimester", Args: []string{"2026-03", "2026-04"}}))
	assert.Equal(t, "Tasa de la tabla de RESICO (renglón 2)", taxenginehttp.Concept(fiscal.RegimeRESICO, p,
		domain.Line{Label: "isr.resico_rate", Args: []string{"2"}}))
	assert.Equal(t, "Ingresos acumulados del ejercicio", taxenginehttp.Concept("601", p, domain.Line{Label: "isr.income_ytd"}))
	assert.Equal(t, "Ingresos acumulados del ejercicio", taxenginehttp.Concept("601", p,
		domain.Line{Label: "isr.income_ytd", Args: []string{"bad", "args"}}))
	jan, err := domain.ParsePeriod("2026-01")
	require.NoError(t, err)
	assert.Equal(t, "Pagos provisionales anteriores", taxenginehttp.Concept("601", jan, domain.Line{Label: "isr.prior_payments"}))
	assert.Equal(t, "Paso del cálculo", taxenginehttp.Concept("601", p, domain.Line{Label: "isr.not_a_step"}))
	assert.Equal(t, "IVA retenido el mes pasado (acreditable este mes; supone que enteraste la retención a tiempo)",
		taxenginehttp.Concept("601", p, domain.Line{Label: "iva.creditable.withheld_prior_month"}))
	assert.Equal(t, "Retención mayor al ISR del mes: este mes pagas $0 y el resto queda a favor en tu declaración anual",
		taxenginehttp.Concept(fiscal.RegimeRESICO, p, domain.Line{Label: "isr.resico_excess_to_annual"}))
	assert.False(t, taxenginehttp.HasConcept("isr.not_a_step"))
	assert.True(t, taxenginehttp.HasConcept("isr.payable"))
}
