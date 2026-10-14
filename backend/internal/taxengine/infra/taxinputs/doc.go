// Package taxinputs is the tax engine's taxpayer_tax_inputs repository
// (M5-20, M5-21). Store implements usecase.TaxInputStore (getTaxInputs,
// putTaxInputs) and usecase.TaxInputReader (the estimate's port, M5-02) with
// GORM on the unit of work in ctx, as cc_app under RLS. It never begins,
// commits or rolls back, and never sets session state (P0-30 D-3).
//
// It also reads the taxpayers, taxpayer_regimes and regimes rows the inputs
// hang from, in the same transaction: the replace must lock the taxpayer
// (FOR SHARE: an archive waits) and the regime row (FOR NO KEY UPDATE: two
// replaces serialize, as the M5-20 migration prescribes), and the input
// definitions are those of the held regime's (code, person type) catalog
// row, exactly what the taxpayer_tax_inputs_derive_kind trigger checks. A
// port through the taxpayers module could not take those locks inside this
// unit of work; the read is SQL only, so no Go code of that module is
// imported (archtest cross-module).
package taxinputs
