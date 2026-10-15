package domain_test

import (
	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
)

func payee(person fiscal.PersonType, regime fiscal.RegimeCode, basis domain.IncomeBasis) domain.Payee {
	return domain.Payee{PersonType: person, Regime: regime, Basis: basis, RFCGivenToPlatform: true}
}

var (
	pf612 = payee(fiscal.PersonFisica, "612", domain.BasisCash)
	pf606 = payee(fiscal.PersonFisica, "606", domain.BasisCash)
	pf626 = payee(fiscal.PersonFisica, "626", domain.BasisCash)
	pf625 = payee(fiscal.PersonFisica, "625", domain.BasisCash)
	pm601 = payee(fiscal.PersonMoral, "601", domain.BasisAccrual)
	pm626 = payee(fiscal.PersonMoral, "626", domain.BasisCash)

	pf625goods      = domain.Payee{PersonType: fiscal.PersonFisica, Regime: "625", Basis: domain.BasisCash, RFCGivenToPlatform: true, PlatformActivities: []domain.PlatformActivity{domain.ActivityGoodsServices}}
	pf625goodsNoRFC = domain.Payee{PersonType: fiscal.PersonFisica, Regime: "625", Basis: domain.BasisCash, PlatformActivities: []domain.PlatformActivity{domain.ActivityGoodsServices}}
)
