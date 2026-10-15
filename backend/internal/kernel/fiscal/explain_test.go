package fiscal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// The contract examples (openapi 0.10.x, Withholding.explanations) are the
// oracle: the app shows these lines verbatim and never maps rule ids itself.
func TestExplain_contractExamples(t *testing.T) {
	tests := []struct {
		name string
		in   fiscal.ExplainInput
		want []fiscal.Explanation
	}{
		{
			name: "US-W1 RESICO services to a company",
			in: fiscal.ExplainInput{
				Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM, RuleIDs: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA6},
				ISRRate: rate(t, "0.0125"), ISR: 125000, IVA: 1066667, IVABasis: fiscal.IVABasisTwoThirds,
			},
			want: []fiscal.Explanation{
				{RuleID: fiscal.RuleISR4, Text: "La empresa te retiene 1.25 % de ISR porque estás en RESICO (LISR art. 113-J)."},
				{RuleID: fiscal.RuleIVA6, Text: "Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3)."},
			},
		},
		{
			name: "US-W2 nobody withholds for the general public",
			in: fiscal.ExplainInput{
				Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPFNoActivityOrPublic,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA0}, IVABasis: fiscal.IVABasisNone,
			},
			want: []fiscal.Explanation{
				{RuleID: fiscal.RuleISR0, Text: "Nadie te retiene ISR cuando le facturas a una persona sin negocio o al público en general."},
				{RuleID: fiscal.RuleIVA0, Text: "Tampoco te retienen IVA en esta operación."},
			},
		},
		{
			name: "US-W4 625 goods via a platform",
			in: fiscal.ExplainInput{
				Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM, RuleIDs: []fiscal.RuleID{fiscal.RuleISR3, fiscal.RuleIVA3},
				ISRRate: rate(t, "0.025"), ISR: 25000, IVA: 80000, IVABasis: fiscal.IVABasisHalfOfIVA,
				Withholder: fiscal.WithholderPlatform, PlatformActivity: fiscal.PlatformGoodsServices,
			},
			want: []fiscal.Explanation{
				{RuleID: fiscal.RuleISR3, Text: "La plataforma te retiene 2.5 % de ISR por venta de bienes y servicios (LISR art. 113-A)."},
				{RuleID: fiscal.RuleIVA3, Text: "Y la mitad del IVA (LIVA art. 18-J)."},
			},
		},
		{
			name: "612 fees to a company",
			in: fiscal.ExplainInput{
				Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM, RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6},
				ISRRate: rate(t, "0.10"), ISR: 100000, IVA: 106667, IVABasis: fiscal.IVABasisTwoThirds,
			},
			want: []fiscal.Explanation{
				{RuleID: fiscal.RuleISR5, Text: "La empresa te retiene 10 % de ISR por honorarios (LISR art. 106)."},
				{RuleID: fiscal.RuleIVA6, Text: "Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3)."},
			},
		},
		{
			name: "US-W6 a persona moral pays a 612 consultant",
			in: fiscal.ExplainInput{
				Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR5, fiscal.RuleIVA6},
				ISRRate: rate(t, "0.10"), ISR: 400000, IVA: 426667, IVABasis: fiscal.IVABasisTwoThirds,
			},
			want: []fiscal.Explanation{
				{RuleID: fiscal.RuleISR5, Text: "Debes retener 10 % de ISR por honorarios de persona física (LISR art. 106)."},
				{RuleID: fiscal.RuleIVA6, Text: "Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3)."},
			},
		},
		{
			name: "WH-06 freight from a persona moral carrier",
			in: fiscal.ExplainInput{
				Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA4},
				IVA: 40000, IVABasis: fiscal.IVABasisFourPercentOfPrice,
			},
			want: []fiscal.Explanation{
				{RuleID: fiscal.RuleISR2, Text: "No retienes ISR a una persona moral."},
				{RuleID: fiscal.RuleIVA4, Text: "Debes retener 4 % del precio por flete de mercancía por carretera (LIVA art. 1-A, RLIVA art. 3)."},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fiscal.Explain(tt.in))
		})
	}
}

// Every rule on each side reads as a sentence of its own: one line per rule
// id, same order, never empty, never a placeholder left unfilled.
func TestExplain_everyRuleHasALine(t *testing.T) {
	isrRules := []fiscal.RuleID{
		fiscal.RuleISR0, fiscal.RuleISR1, fiscal.RuleISR2, fiscal.RuleISR3,
		fiscal.RuleISR4, fiscal.RuleISR5, fiscal.RuleISR6, fiscal.RuleISR7,
	}
	ivaRules := []fiscal.RuleID{
		fiscal.RuleIVANoTax, fiscal.RuleIVA0, fiscal.RuleIVA1, fiscal.RuleIVA2, fiscal.RuleIVA3,
		fiscal.RuleIVA4, fiscal.RuleIVA5, fiscal.RuleIVA6, fiscal.RuleIVA7, fiscal.RuleIVA8,
	}
	receptors := []fiscal.ReceptorKind{
		fiscal.ReceptorPM, fiscal.ReceptorGovFederal, fiscal.ReceptorGovStateMunicipal,
		fiscal.ReceptorPFWithActivity, fiscal.ReceptorPFNoActivityOrPublic, fiscal.ReceptorForeign,
	}
	for _, side := range []fiscal.Side{fiscal.SideIncome, fiscal.SideExpense} {
		for _, receptor := range receptors {
			for _, isr := range isrRules {
				for _, iva := range ivaRules {
					for _, amounts := range [][2]money.Cents{{0, 0}, {100, 0}, {0, 100}, {100, 100}} {
						lines := fiscal.Explain(fiscal.ExplainInput{
							Side: side, Receptor: receptor, RuleIDs: []fiscal.RuleID{isr, iva},
							ISRRate: rate(t, "0.10"), ISR: amounts[0], IVA: amounts[1], IVABasis: fiscal.IVABasisTwoThirds,
						})
						if assert.Len(t, lines, 2) {
							assert.Equal(t, isr, lines[0].RuleID)
							assert.Equal(t, iva, lines[1].RuleID)
							for _, l := range lines {
								assert.NotEmpty(t, l.Text)
								assert.NotContains(t, l.Text, "%!")
							}
						}
					}
				}
			}
		}
	}
}

func TestExplain_wording(t *testing.T) {
	tests := []struct {
		name string
		in   fiscal.ExplainInput
		want []string
	}{
		{
			name: "a persona física with activity never withholds",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPFWithActivity,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR0, fiscal.RuleIVA1}},
			want: []string{
				"Nadie te retiene ISR cuando le facturas a una persona física.",
				"Tampoco te retienen IVA en esta operación.",
			},
		},
		{
			name: "foreign customer",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorForeign,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR1, fiscal.RuleIVA0}},
			want: []string{
				"Nadie te retiene ISR cuando le facturas a alguien en el extranjero.",
				"Tampoco te retienen IVA en esta operación.",
			},
		},
		{
			name: "the federal government withholds 2/3 without ISR",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorGovFederal,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA7}, IVA: 3200, IVABasis: fiscal.IVABasisTwoThirds},
			want: []string{
				"El gobierno federal no te retiene ISR en esta operación.",
				"El gobierno federal te retiene 2/3 del IVA por servicios de persona física al gobierno federal (LIVA art. 3, RLIVA art. 3).",
			},
		},
		{
			name: "a state government withholds ISR on rent, no IVA",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorGovStateMunicipal,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR6, fiscal.RuleIVA2}, ISRRate: rate(t, "0.1"), ISR: 300000},
			want: []string{
				"El gobierno estatal o municipal te retiene 10 % de ISR por arrendamiento (LISR art. 116).",
				"No te retienen IVA en esta operación.",
			},
		},
		{
			name: "residential rent has no IVA to withhold",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR6, fiscal.RuleIVANoTax}, ISRRate: rate(t, "0.1"), ISR: 300000},
			want: []string{
				"La empresa te retiene 10 % de ISR por arrendamiento (LISR art. 116).",
				"La operación no lleva IVA: no hay IVA que retener.",
			},
		},
		{
			name: "a persona moral selling via a platform",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM, Withholder: fiscal.WithholderPlatform,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA8}, ISRRate: rate(t, "0.025"), ISR: 25000},
			want: []string{
				"La plataforma te retiene 2.5 % de ISR (LISR art. 113-A; LIF art. 25 fr. VI).",
				"No te retienen IVA en esta operación.",
			},
		},
		{
			name: "a persona moral selling to a company",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR2, fiscal.RuleIVA4}, IVA: 40000, IVABasis: fiscal.IVABasisFourPercentOfPrice},
			want: []string{
				"Como persona moral, no te retienen ISR en esta operación.",
				"La empresa te retiene 4 % del precio por flete de mercancía por carretera (LIVA art. 1-A, RLIVA art. 3).",
			},
		},
		{
			name: "a 625 seller without RFC loses all the IVA",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPFNoActivityOrPublic,
				Withholder: fiscal.WithholderPlatform, PlatformActivity: fiscal.PlatformLodging,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR3, fiscal.RuleIVA3}, ISRRate: rate(t, "0.2"), ISR: 200000, IVA: 160000,
				IVABasis: fiscal.IVABasisAllOfIVA},
			want: []string{
				"La plataforma te retiene 20 % de ISR por hospedaje (LISR art. 113-A).",
				"Y todo el IVA porque no le diste tu RFC a la plataforma (LIVA art. 18-J).",
			},
		},
		{
			name: "a platform sale of transport, IVA only",
			in: fiscal.ExplainInput{Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPFNoActivityOrPublic,
				Withholder: fiscal.WithholderPlatform, PlatformActivity: fiscal.PlatformTransportDelivery,
				RuleIDs: []fiscal.RuleID{fiscal.RuleISR3, fiscal.RuleIVA3}, ISRRate: rate(t, "0.021"), IVA: 8000,
				IVABasis: fiscal.IVABasisHalfOfIVA},
			want: []string{
				"La plataforma te retiene 2.1 % de ISR por transporte o entrega (LISR art. 113-A).",
				"La plataforma te retiene la mitad del IVA (LIVA art. 18-J).",
			},
		},
		{
			name: "RESICO supplier paid by a persona moral, scrap",
			in: fiscal.ExplainInput{Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA5},
				ISRRate: rate(t, "0.0125"), ISR: 25000, IVA: 320000, IVABasis: fiscal.IVABasisAllOfIVA},
			want: []string{
				"Debes retener 1.25 % de ISR porque tu proveedor está en RESICO (LISR art. 113-J).",
				"Y todo el IVA por desperdicios y reciclables (LIVA art. 1-A).",
			},
		},
		{
			name: "rent paid to a persona física",
			in: fiscal.ExplainInput{Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR6, fiscal.RuleIVA6},
				ISRRate: rate(t, "0.1"), ISR: 300000, IVA: 320000, IVABasis: fiscal.IVABasisTwoThirds},
			want: []string{
				"Debes retener 10 % de ISR por arrendamiento de persona física (LISR art. 116).",
				"Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3).",
			},
		},
		{
			name: "goods bought from a persona física",
			in:   fiscal.ExplainInput{Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR7, fiscal.RuleIVA8}},
			want: []string{
				"No retienes ISR en esta operación.",
				"Tampoco retienes IVA en esta operación.",
			},
		},
		{
			name: "a foreign supplier is not calculated",
			in:   fiscal.ExplainInput{Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR1, fiscal.RuleIVA0}},
			want: []string{
				"No calculamos retenciones a proveedores en el extranjero.",
				"Tampoco retienes IVA en esta operación.",
			},
		},
		{
			name: "an RESICO supplier of goods: ISR only",
			in: fiscal.ExplainInput{Side: fiscal.SideExpense, RuleIDs: []fiscal.RuleID{fiscal.RuleISR4, fiscal.RuleIVA8},
				ISRRate: rate(t, "0.0125"), ISR: 62500},
			want: []string{
				"Debes retener 1.25 % de ISR porque tu proveedor está en RESICO (LISR art. 113-J).",
				"No retienes IVA en esta operación.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := fiscal.Explain(tt.in)
			texts := make([]string, len(lines))
			for i, l := range lines {
				texts[i] = l.Text
			}
			assert.Equal(t, tt.want, texts)
		})
	}
}

// A rule id a newer table adds still reads as a sentence until it gets its
// own wording.
func TestExplain_unknownRuleFallsBack(t *testing.T) {
	lines := fiscal.Explain(fiscal.ExplainInput{
		Side: fiscal.SideIncome, Receptor: fiscal.ReceptorPM, RuleIDs: []fiscal.RuleID{"ISR-99", "IVA-99"}, IVA: 100,
	})

	assert.Equal(t, []fiscal.Explanation{
		{RuleID: "ISR-99", Text: "La empresa no te retiene ISR en esta operación."},
		{RuleID: "IVA-99", Text: "La empresa te retiene parte del IVA (LIVA art. 1-A)."},
	}, lines)
}

func TestPercent(t *testing.T) {
	for raw, want := range map[string]string{"0": "0 %", "0.0125": "1.25 %", "0.1": "10 %", "0.021": "2.1 %", "1": "100 %"} {
		assert.Equal(t, want, fiscal.Percent(rate(t, raw)), raw)
	}
}

func TestExplain_noRulesNoLines(t *testing.T) {
	assert.Empty(t, fiscal.Explain(fiscal.ExplainInput{Side: fiscal.SideIncome}))
}

// A derivation explains itself with the facts it was derived from.
func TestDerivation_Explain(t *testing.T) {
	catalog := withholdingCatalog2026(t)
	d, err := fiscal.DeriveWithholdings(fiscal.WithholdingInput{
		EmisorPersona: fiscal.PersonFisica, EmisorRegime: fiscal.RegimeRESICO, Receptor: fiscal.ReceptorPM,
		Operation: fiscal.OperationProfessionalServices, Subtotal: 10_000_000, IVA: 1_600_000,
		IVARate: fiscal.IVARate16, On: date(t, "2026-07-10"),
	}, catalog)
	if err != nil {
		t.Fatal(err)
	}

	lines := d.Explain(fiscal.SideIncome, fiscal.ReceptorPM, "")

	assert.Equal(t, []fiscal.Explanation{
		{RuleID: fiscal.RuleISR4, Text: "La empresa te retiene 1.25 % de ISR porque estás en RESICO (LISR art. 113-J)."},
		{RuleID: fiscal.RuleIVA6, Text: "Y 2/3 del IVA por servicios de persona física (LIVA art. 1-A, RLIVA art. 3)."},
	}, lines)
}
