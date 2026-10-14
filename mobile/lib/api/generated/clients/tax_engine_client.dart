// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/breakdown.dart';
import '../models/fiscal_year.dart';
import '../models/period.dart';
import '../models/regime_code.dart';
import '../models/tax_inputs.dart';
import '../models/tax_inputs_update.dart';
import '../models/taxpayer_estimate.dart';
import '../models/uuid.dart';

part 'tax_engine_client.g.dart';

@RestApi()
abstract class TaxEngineClient {
  factory TaxEngineClient(Dio dio, {String? baseUrl}) = _TaxEngineClient;

  /// Monthly IVA + provisional ISR estimate of one month, per regime and total.
  ///
  /// Informational estimate for the taxpayer and `period`, derived from the current records and tax inputs on every read (never stored, ADR-0002 §5). `period` is optional: absent means the current month in America/Mexico_City. Past months back to the first fiscal year with published parameters (2026) are allowed; the current, still-open month estimates with the records so far.
  ///
  /// **Content.** One `RegimeEstimate` per regime held in the period (archived regimes included for periods up to their `end_period`), in catalog order. A taxpayer with no regime in the period answers `200` with `regimes: []` and zero totals. Monthly IVA is consolidated per taxpayer in `totals.iva` (`Σ charged_collected − Σ definitive_excluded − Σ creditable_paid − Σ withheld_from_taxpayer` over every regime, OQ-T2); it is never presented per regime as an amount to pay. A negative result is `in_favor` (saldo a favor, not carried to the next month, OQ-T5) and is not counted in `total_payable_rounded_cents`. `totals.isr` sums the payable per-regime ISR only (favor balances stay per regime, OQ-T10). Withholdings the taxpayer made to its suppliers are `withholdings_to_remit_cents`, never netted against IVA (LIVA 1-A).
  ///
  /// **Status per regime** (precedence `informational` > `incomplete` > `partial` > `complete`): `incomplete` computes with zero for every missing input and lists them in `missing_inputs` (the CTA opens `getTaxInputs` for that regime and fiscal year at the first missing key); `partial` estimates only some taxes and explains the rest; `informational` carries the explanation and no amount. A missing input is never an error. Rules a CPA has not validated still run with their documented default, set `cpa_review_pending`, list `cpa_topics` and add the `cpa_review_pending` disclaimer (severity `strong`, OQ-T8).
  ///
  /// `legend` is the mandatory constant; the client renders the server field. Reads are allowed for a read-only user and for an archived taxpayer.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `period` → `out_of_range` (a future month, America/Mexico_City). `422 business_rule_violation` details: `period` → `parameters_unavailable` (before 2026, the first fiscal year with transcribed parameters, OQ-T9).
  ///
  /// [period] - Month to estimate (`YYYY-MM`). Absent → the current month (America/Mexico_City).
  @GET('/v1/taxpayers/{taxpayerId}/estimates')
  Future<TaxpayerEstimate> getTaxpayerEstimate({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') Period? period,
  });

  /// Step-by-step computation of one regime's estimate, with legal basis.
  ///
  /// The computation trace of `regime` in `period` (ADR-0002 §1: strategies compute only by appending lines, so the explanation cannot drift from the math). `period` is required here and follows the rules of `getTaxpayerEstimate` (no future month, nothing before 2026). `lines` are in `seq` order: the regime's ISR lines, its IVA component lines, the taxpayer's consolidated IVA lines (`tax: iva`, LIVA 5-D, with `needs_cpa_review` when two or more regimes contribute IVA) and its withholdings to remit (`tax: withholding`). Every line except `info` and `rounding_adjustment` cites at least one legal basis, and `source` says where its figure comes from. Lines are displayed rounded to the centavo; when the displayed lines of a result do not add up to its rounded value, one `rounding_adjustment` line ("Ajuste por redondeo") reconciles them: never a hidden centavo. An `informational` regime answers its explanation with `lines: []`. The header is `regime.parameter_set` and `engine_version`. Reads are allowed for a read-only user and an archived taxpayer.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `period` → `out_of_range` (a future month). `422 business_rule_violation` details: `period` → `parameters_unavailable` (before 2026); `regime` → `regime_not_active` (the taxpayer does not hold the regime in that period).
  @GET('/v1/taxpayers/{taxpayerId}/estimates/{period}/breakdown')
  Future<Breakdown> getEstimateBreakdown({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('period') required Period period,
    @Query('regime') required RegimeCode regime,
  });

  /// Tax inputs of one regime and fiscal year, with their definitions.
  ///
  /// The per-fiscal-year inputs the regime's estimate needs (`definitions`, exactly the catalog's `regimes.required_inputs`) and the stored `values`. A key with no value is absent from `values`. `TaxInputDef.required` says the key can be required in some period of the year; which keys and months are missing **for a given period** is `RegimeEstimate.missing_inputs` (e.g. prior provisional payments are never required for January). This is the target of the "Completa tus datos" CTA. Reads are allowed for a read-only user and for an archived taxpayer.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscal_year` → `out_of_range` (a future year). `422 business_rule_violation` details: `fiscal_year` → `parameters_unavailable` (before 2026); `regimeCode` → `regime_not_active` (the taxpayer does not hold the regime in any month of that fiscal year).
  ///
  /// [fiscalYear] - Fiscal year of the inputs; 2026 … the current year (America/Mexico_City).
  @GET('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs')
  Future<TaxInputs> getTaxInputs({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Query('fiscal_year') required FiscalYear fiscalYear,
  });

  /// Replace the tax inputs of one regime and fiscal year.
  ///
  /// Replaces **all** values of `(taxpayer, regime, fiscal_year)` with `values` (an empty array clears them). Idempotent: the same body twice leaves the same state. Values are validated against the regime's `required_inputs` (catalog data): the key must be one of them, and exactly the value field of its `kind` is set: `value_cents` for `money`, `value_cents` plus `month` (1-12) for `money_monthly`, `value_rate` for `rate`, `value_bool` for `boolean`; every other value field is absent or `null`. Money is `>= 0`; the coeficiente de utilidad is `0 <= cu <= 1` and used as given. At most one value per key (per key and month for `money_monthly`). A `money_monthly` value, e.g. a monthly prior provisional payment (OQ-T1), is the amount actually paid for that month: the app may prefill its own estimate, the server never assumes it.
  ///
  /// **Order of checks:** schema (`400`) → read-only access (`402`) → archived taxpayer, regime window and fiscal year (`422`) → value rules (`400`) → replace. Per-user write budget: 60 per hour (ADR-0005, `429`). Audited as `tax_inputs.updated` with user, taxpayer, regime and fiscal year only, never the values.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords (`field` is `values.<index>[.<property>]`): `unknown_input` (the key is not among the regime's inputs); `invalid` (the value field does not match the kind, both or neither are set, `month` on a non-monthly key or missing on a monthly one, or a repeated key or key and month); `out_of_range` (a rate outside `[0, 1]`); `fiscal_year` → `out_of_range` (a future year). `422 business_rule_violation` details: `taxpayerId` → `taxpayer_archived`; `regimeCode` → `regime_not_active`; `fiscal_year` → `parameters_unavailable`.
  ///
  /// [fiscalYear] - Fiscal year of the inputs; 2026 … the current year (America/Mexico_City).
  @PUT('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/tax-inputs')
  Future<TaxInputs> putTaxInputs({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Query('fiscal_year') required FiscalYear fiscalYear,
    @Body() required TaxInputsUpdate body,
  });
}
