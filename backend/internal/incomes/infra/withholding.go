package infra

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/romodevx/control-contable/backend/internal/incomes/domain"
	"github.com/romodevx/control-contable/backend/internal/incomes/usecase"
	"github.com/romodevx/control-contable/backend/internal/kernel/fiscal"
	"github.com/romodevx/control-contable/backend/internal/platform/money"
)

// ruleIDSeparator joins the rule ids on the wire to the database: the ids
// never contain it (contract WithholdingRuleId pattern), and
// string_to_array/array_to_string keep text[] out of the driver.
const ruleIDSeparator = ","

// writableColumns are every column an insert or an update of an income
// writes: the cc_app column grants (M3-02, WH-03, 20261017100000) minus the
// keys and deleted_at. columnsOf builds exactly these keys, so no request
// field can reach a column by name (PR #136 review note 1).
var writableColumns = []string{
	"regime_code", "counterparty_type", "counterparty_kind", "operation_kind", "residential_rent", "via_platform",
	"amount_cents", "iva_rate", "iva_cents", "isr_withheld_cents", "iva_withheld_cents",
	"withholding_rule_ids", "withholding_table_version", "withholding_cpa_unverified",
	"withholding_isr_rate", "withholding_iva_basis", "withholding_withholder",
	"invoiced_on", "collected_on", "platform_activity", "cfdi_uuid", "description",
}

// columnsOf maps the facts and the derived withholding to the writable
// columns; nil is NULL. counterparty_type is the deprecated mirror
// (domain.LegacyCounterparty) until WH-08 drops it.
func columnsOf(f usecase.Fields, w fiscal.WithholdingRecord) map[string]any {
	columns := map[string]any{
		"regime_code":        string(f.Regime),
		"counterparty_type":  string(domain.LegacyCounterparty(f.Counterparty, f.ViaPlatform)),
		"counterparty_kind":  nullable(string(f.Counterparty)),
		"operation_kind":     nullable(string(f.Operation)),
		"residential_rent":   f.ResidentialRent,
		"via_platform":       f.ViaPlatform,
		"amount_cents":       int64(f.Amount),
		"iva_rate":           string(f.IVARate),
		"iva_cents":          int64(f.IVA),
		"isr_withheld_cents": int64(w.ISR),
		"iva_withheld_cents": int64(w.IVA),
		"withholding_rule_ids": gorm.Expr("string_to_array(?, ?)",
			strings.Join(ruleIDStrings(w.RuleIDs), ruleIDSeparator), ruleIDSeparator),
		"withholding_table_version":  nullable(w.TableVersion),
		"withholding_cpa_unverified": w.CPAUnverified,
		"withholding_isr_rate":       nil,
		"withholding_iva_basis":      nil,
		"withholding_withholder":     nil,
		"invoiced_on":                nullable(f.Dates.InvoicedOn.String()),
		"collected_on":               nullable(f.Dates.CollectedOn.String()),
		"platform_activity":          nullable(string(f.PlatformActivity)),
		"cfdi_uuid":                  nullable(f.CFDIUUID.String()),
		"description":                nullable(f.Description),
	}
	if w.Derived() {
		columns["withholding_isr_rate"] = gorm.Expr("CAST(? AS numeric)", w.ISRRate.String())
		columns["withholding_iva_basis"] = string(w.IVABasis)
		columns["withholding_withholder"] = string(w.Withholder)
	}
	return columns
}

func ruleIDStrings(ids []fiscal.RuleID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = string(id)
	}
	return out
}

// withholdingSelect reads the stored derivation in driver-neutral forms.
const withholdingSelect = "array_to_string(withholding_rule_ids, ',') AS withholding_rule_ids, " +
	"withholding_table_version, withholding_cpa_unverified, withholding_isr_rate::text AS withholding_isr_rate, " +
	"withholding_iva_basis, withholding_withholder"

// withholdingColumns are the stored derivation as read.
type withholdingColumns struct {
	RuleIDs       string  `gorm:"column:withholding_rule_ids"`
	TableVersion  *string `gorm:"column:withholding_table_version"`
	CPAUnverified bool    `gorm:"column:withholding_cpa_unverified"`
	ISRRate       *string `gorm:"column:withholding_isr_rate"`
	IVABasis      *string `gorm:"column:withholding_iva_basis"`
	Withholder    *string `gorm:"column:withholding_withholder"`
}

// record is the stored derivation with its amounts. A row without the
// detail columns (none is written so: the CHECKs set them with the version)
// reads as underived, so the API never serves a half derivation.
func (c withholdingColumns) record(isr, iva int64) (fiscal.WithholdingRecord, error) {
	w := fiscal.WithholdingRecord{Withholdings: fiscal.Withholdings{ISR: money.Cents(isr), IVA: money.Cents(iva)}}
	if c.TableVersion == nil || c.ISRRate == nil {
		return w, nil
	}
	r, err := money.ParseRate(*c.ISRRate)
	if err != nil {
		return fiscal.WithholdingRecord{}, fmt.Errorf("incomes infra: stored withholding_isr_rate: %w", err)
	}
	w.ISRRate, w.IVABasis, w.Withholder = r, fiscal.IVABasis(deref(c.IVABasis)), fiscal.Withholder(deref(c.Withholder))
	w.TableVersion, w.CPAUnverified = *c.TableVersion, c.CPAUnverified
	for _, id := range strings.Split(c.RuleIDs, ruleIDSeparator) {
		if id != "" {
			w.RuleIDs = append(w.RuleIDs, fiscal.RuleID(id))
		}
	}
	return w, nil
}
