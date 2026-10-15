package usecase

import (
	"fmt"

	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

// dueDate is the kernel's monthly due date of p (fiscal.MonthlyDueDate:
// day 17 of the following month moved past weekends and holidays, CFF 12
// ⚠️ "sujeto a confirmación"). When every regime held is the bimonthly
// 621, the payment follows the bimester (fiscal.PaysBimonthly, §2.10).
func dueDate(p domain.Period, held []RegimeHold, holidays fiscal.HolidaySet) (fiscal.Date, error) {
	_, due, _, err := fiscal.MonthlyDueDate(p.Year(), p.Month(), fiscal.PaysBimonthly(heldCodes(held)), holidays)
	if err != nil {
		return fiscal.Date{}, fmt.Errorf("taxengine: due date of %s: %w", p, err)
	}
	return due, nil
}

func heldCodes(held []RegimeHold) []fiscal.RegimeCode {
	codes := make([]fiscal.RegimeCode, len(held))
	for i, h := range held {
		codes[i] = h.Code
	}
	return codes
}
