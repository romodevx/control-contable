// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

/// A record date on the America/Mexico_City calendar. The server refuses dates before 2026-01-01 or after today with `400 validation_failed`, `out_of_range` on that date: 2026 is the first fiscal year with a tax parameter set, so no IVA rate or withholding on an earlier date can be validated, and a record the server cannot validate is invalid input. The bounds move with the clock and the parameter catalog, so they are not schema keywords. A date outside the taxpayer's regime window answers `422 regime_not_active` first (the window is checked before the parameters). Not the same answer as the estimates: a month before 2026 there is a well-formed period the engine has no parameters for, `422 business_rule_violation` / `parameters_unavailable` (`getTaxpayerEstimate`).
typedef BusinessDate = String;
