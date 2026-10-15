package domain

import "github.com/romodevx/control-contable/backend/internal/kernel/fiscal"

// Payee is the taxpayer receiving the income, as the taxpayers module
// reports it (person type, regime and its catalog basis, 625 options).
type Payee struct {
	PersonType         fiscal.PersonType
	Regime             fiscal.RegimeCode
	Basis              IncomeBasis
	PlatformActivities []PlatformActivity // 625 option "activities"
	RFCGivenToPlatform bool               // 625 option "rfc_given_to_platform"
}
