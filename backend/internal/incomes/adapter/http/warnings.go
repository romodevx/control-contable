package incomeshttp

import (
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
)

// warningMessages are the es-MX texts of each warning code
// (MODULE-incomes §2.8). They never echo an amount, a description or a
// folio fiscal.
var warningMessages = map[domain.WarningCode]string{
	domain.WarningDuplicateCFDIUUID: "Ya tienes otro registro con este folio fiscal.",
	domain.WarningWithholdingDiffersFromRate: "La retención no coincide con la tasa sugerida; " +
		"guardamos la de tu factura.",
	domain.WarningProvisionalParameters: "Aún no se publican los parámetros fiscales de ese año; " +
		"usamos los del año anterior.",
}

// msgWarningFallback is the text of a code without its own (the contract's
// clients show `message` for a code they do not know).
const msgWarningFallback = "Revisa este dato."

// toWarnings is never nil: the contract's warnings is [] when there are
// none (always on reads).
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
