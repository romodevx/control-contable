package expenseshttp

import (
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/expenses/domain"
	"github.com/romodevx/control-contable/backend/internal/expenses/usecase"
)

// warningMessages are the es-MX texts of each warning code
// (MODULE-expenses §2.5, §2.8). They never echo an amount, a description or
// a folio fiscal. The $2,000 is the spec's copy of the 2026 cash limit
// (LISR 27 fr. III); the rule itself reads the dated parameter.
var warningMessages = map[domain.WarningCode]string{
	domain.WarningCashOverThreshold:           "Los pagos en efectivo de más de $2,000 no son deducibles.",
	domain.WarningIVANotCreditableCashPayment: "El IVA de un pago en efectivo de más de $2,000 no es acreditable.",
	domain.WarningNoCFDI:                      "Sin factura (CFDI) este gasto no es deducible.",
	domain.WarningNotISRDeductibleInRegime:    "En tu régimen este gasto no reduce tu ISR.",
	domain.WarningCategoryNotDeductible:       "Esta categoría normalmente no es deducible. Confírmalo con tu contador.",
	domain.WarningDuplicateCFDIUUID:           "Ya tienes otro registro con este folio fiscal.",
	usecase.WarningProvisionalParameters: "Aún no se publican los parámetros fiscales de ese año; " +
		"usamos los del año anterior.",
}

// msgWarningFallback is the text of a code without its own (the contract's
// clients show `message` for a code they do not know).
const msgWarningFallback = "Revisa este dato."

// toWarnings is never nil: the contract's warnings is [] when there are
// none.
func toWarnings(warnings []domain.Warning) []gen.Warning {
	out := make([]gen.Warning, 0, len(warnings))
	for _, w := range warnings {
		message, ok := warningMessages[w.Code]
		if !ok {
			message = msgWarningFallback
		}
		out = append(out, gen.Warning{Code: gen.WarningCode(w.Code), Field: w.Field, Message: message})
	}
	return out
}
