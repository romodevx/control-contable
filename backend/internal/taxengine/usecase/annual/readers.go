package annual

import (
	"slices"

	adomain "github.com/romodevx/control-contable/backend/internal/taxengine/domain/annual"
)

// YearInputsOf assembles what YearInputsReader answers from the stored live
// deductions, the constancias (in their list order) and the facts: the
// deductions in the list order, so the calculator's notes are
// deterministic (MODULE-annual §2.11).
func YearInputsOf(deductions []StoredDeduction, statements []StoredStatement, facts Facts) YearInputs {
	sorted := slices.Clone(deductions)
	sortDeductions(sorted)
	out := YearInputs{
		Deductions:       make([]adomain.Deduction, 0, len(sorted)),
		SalaryStatements: make([]adomain.SalaryStatement, 0, len(statements)),
		Facts:            facts.Domain(),
	}
	for _, d := range sorted {
		out.Deductions = append(out.Deductions, d.Domain())
	}
	for _, s := range statements {
		out.SalaryStatements = append(out.SalaryStatements, s.Domain())
	}
	return out
}
