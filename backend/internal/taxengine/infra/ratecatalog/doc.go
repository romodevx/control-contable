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
package ratecatalog
