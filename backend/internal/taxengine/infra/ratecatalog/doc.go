// Package ratecatalog is the tax engine's RateCatalog adapter (M5-08,
// ADR-0002 §3, ADR-0007 §6): it serves the dated rates and amounts of the
// versioned parameter files to the income and expense domains.
//
// One Catalog value satisfies both consumer-side ports: fiscal.RateCatalog
// (incomes reads it as is) and the expenses module's RateCatalog, which
// embeds it and adds Amount. The tax engine may not import another module,
// so the amount key type is a type parameter: the composition root
// instantiates New[expenses domain.AmountKey] and hands the same value to
// both modules.
//
// Every value resolves by the record's date, never by fiscal year alone. The
// set is the highest version of the date's fiscal year; while that year is
// not published, the previous year's set is used and the result is flagged
// Provisional (the provisional_parameters disclaimer). A date no set covers
// is ErrDateNotCovered, never a silent value. The boolean port methods
// (Rate, Amount) report every such failure as "no value in force"; the
// error-returning methods (RateOn, AmountOn, AllowedIVARates,
// CashPaymentLimit, UMA) tell the cases apart and carry the flag.
//
// Withholding rates (WH-05, for fiscal.DeriveWithholdings of WH-04): one
// accessor per dated rate, each (money.Rate, error) on the booking date so a
// kernel consumer interface can declare them with kernel types only:
//
//	ISRWithholdingFees     isr.withholding.fees_pm     10 %   of the subtotal  LISR 106
//	ISRWithholdingRent     isr.withholding.rent_pm     10 %   of the subtotal  LISR 116
//	ISRWithholdingRESICO   isr.withholding.resico_pf   1.25 % of the subtotal  LISR 113-J
//	IVAWithholdingFreight  iva.withholding.freight_pm  4 %    of the subtotal  LIVA 1-A II c; RLIVA 3 II
//	IVAWithholdingScrap    iva.withholding.scrap_pm    100 %  of the IVA       LIVA 1-A II b
//
// plus IVAWithholdingFederalThreshold (iva.withholding.federal_threshold,
// $2,000.00, RMF 2026 rule 4.1.3, lapses with its RMF). Withholdings
// returns the WithholdingView that implements fiscal.WithholdingCatalog
// whatever amount key the Catalog was built with.
//
// New refuses a served set lacking any of them (ErrMissingWithholdingRate).
// The 2/3 IVA share (RLIVA 3 fr. I) and the platform 50 %/100 % IVA shares
// (LIVA 18-J) stay in code as fiscal.ShareTwoThirds / ShareHalf /
// ShareWhole; the platform ISR rates are read with Rate and the existing
// isr.platform.* keys.
package ratecatalog
