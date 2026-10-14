package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	taxengine "github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// Every disclaimer code the tax engine can attach must be in the contract's
// append-only DisclaimerCode enum (docs/api/openapi.yaml). A strategy adding a
// code without the contract fails here, not in the client's unknown fallback.
func TestTaxEngineDisclaimerCodesAreInTheContract(t *testing.T) {
	codes := []taxengine.DisclaimerCode{
		taxengine.DisclaimerNotTaxAdvice,
		taxengine.DisclaimerIncompleteInputs,
		taxengine.DisclaimerProvisionalParameters,
		taxengine.DisclaimerInformationalRegime,
		taxengine.DisclaimerPartialRegime,
		taxengine.DisclaimerDeductibilityClaimed,
		taxengine.DisclaimerNoCoefficientFirstYear,
		taxengine.DisclaimerIVAMixedActivity,
		taxengine.DisclaimerCPAReviewPending,
		taxengine.DisclaimerRESICOCapExceeded,
	}
	for _, code := range codes {
		t.Run(string(code), func(t *testing.T) {
			assert.True(t, code.Valid(), "give %q a severity in the domain", code)
			assert.True(t, gen.DisclaimerCode(code).Valid(), "add %q to DisclaimerCode in docs/api/openapi.yaml", code)
		})
	}
}
