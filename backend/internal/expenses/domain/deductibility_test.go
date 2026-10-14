package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

// diningCategory is the Spanish catalog code of MODULE-expenses §2.2 row 11.
const diningCategory domain.CategoryCode = "restaurantes" //nolint:misspell // a Spanish catalog code, not English prose

// Category rule rows of MODULE-expenses §2.2, as category_regime_rules
// serves them for one (category, regime, person type).
func rule(category domain.CategoryCode, d domain.Deductibility, defaultDeductible bool) domain.CategoryRule {
	return domain.CategoryRule{Category: category, Deductibility: d, DefaultDeductible: defaultDeductible}
}

var (
	softwareBusiness  = rule("software", domain.Deductible, true)
	diningBusiness    = rule(diningCategory, domain.PartiallyDeductible, false)
	personalBusiness  = rule("gastos_personales", domain.NotDeductible, false)
	mantenimiento606  = rule("mantenimiento", domain.Deductible, true)
	predial606        = rule(domain.CategoryPredial, domain.Deductible, true)
	softwareRESICOPF  = rule("software", domain.ISRNotApplicable, false)
	papeleriaBusiness = rule("papeleria", domain.Deductible, true)
	fuelBusiness      = domain.CategoryRule{
		Category: "combustible", Deductibility: domain.Deductible, DefaultDeductible: true, ElectronicPaymentOnly: true,
	}
	feesBusiness = rule(domain.CategoryProfessionalFees, domain.Deductible, true)
)

func payer(person fiscal.PersonType, regime fiscal.RegimeCode, opts domain.Options) domain.Payer {
	return domain.Payer{PersonType: person, Regime: regime, Options: opts}
}

var (
	pf612      = payer(fiscal.PersonFisica, "612", domain.Options{})
	pf626      = payer(fiscal.PersonFisica, "626", domain.Options{})
	pm626      = payer(fiscal.PersonMoral, "626", domain.Options{})
	pm601      = payer(fiscal.PersonMoral, "601", domain.Options{})
	pf606Blind = payer(fiscal.PersonFisica, "606", domain.Options{DeductionMode: domain.DeductionBlind35})
	pf606Real  = payer(fiscal.PersonFisica, "606", domain.Options{DeductionMode: domain.DeductionActual})
	pf625Def   = payer(fiscal.PersonFisica, "625", domain.Options{DefinitiveOption: true})
	pf625      = payer(fiscal.PersonFisica, "625", domain.Options{})
)

// TestResolveDeductibility is MODULE-expenses §2.3: the regime options
// change the catalog hint; nothing else does.
func TestResolveDeductibility(t *testing.T) {
	notApplicable := domain.Hint{Deductibility: domain.ISRNotApplicable}
	tests := []struct {
		name  string
		rule  domain.CategoryRule
		payer domain.Payer
		want  domain.Hint
	}{
		{"612 software: the catalog row", softwareBusiness, pf612, domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}},
		{"612 restaurants: partial, conservative default", diningBusiness, pf612, domain.Hint{Deductibility: domain.PartiallyDeductible}},
		{"US-E1: RESICO PF, any category", softwareRESICOPF, pf626, notApplicable},
		{"RESICO PF even if a row said deductible", softwareBusiness, pf626, notApplicable},
		{"US-E1b: RESICO PM keeps the business row", softwareBusiness, pm626, domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}},
		{"US-E1: 606 blind 35 %, maintenance does not count", mantenimiento606, pf606Blind, notApplicable},
		{"US-E1: 606 blind 35 %, predial is added (LISR 115 ¶2)", predial606, pf606Blind, domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}},
		{"606 actual deductions keep the row", mantenimiento606, pf606Real, domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}},
		{"625 definitive option (113-B)", softwareBusiness, pf625Def, notApplicable},
		{"625 without the definitive option", softwareBusiness, pf625, domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}},
		{"601 ignores the 606 option key", softwareBusiness, payer(fiscal.PersonMoral, "601", domain.Options{DeductionMode: domain.DeductionBlind35}),
			domain.Hint{Deductibility: domain.Deductible, DefaultDeductible: true}},
		{"a not-deductible row never defaults to true", rule("gastos_personales", domain.NotDeductible, true), pf612,
			domain.Hint{Deductibility: domain.NotDeductible}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.ResolveDeductibility(tt.rule, tt.payer))
		})
	}
}
