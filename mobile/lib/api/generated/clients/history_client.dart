// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/fiscal_year.dart';
import '../models/history_viewed_request.dart';
import '../models/history_year.dart';
import '../models/period.dart';
import '../models/uuid.dart';

part 'history_client.g.dart';

@RestApi()
abstract class HistoryClient {
  factory HistoryClient(Dio dio, {String? baseUrl}) = _HistoryClient;

  /// Fiscal-year history of one taxpayer - every month's estimate, status and changed marker.
  ///
  /// For the taxpayer and `fiscal_year`, one row per month with the estimated totals, the taxpayer-level status and whether the displayed figures changed since the caller last viewed that month. Every figure is recalculated from the current records on each read (ADR-0002 §5); History stores only what the user saw, never a tax figure.
  ///
  /// **Months.** January .. December for a past year; January .. the current month (America/Mexico_City) for the current year. Ordered **newest first** (the current month, `is_current: true`, at the top). The list is always contiguous: a month in which the taxpayer held no regime comes back with `status: none` and `tax: null` ("Sin régimen en este mes"), never as a hole. Each month's figures are the same `totals` that `getTaxpayerEstimate` answers for that `period`, to the centavo; the client never recomputes or re-rounds them. `year_totals` is the sum of the returned months' estimates ("Suma de estimados mensuales"), **not** an annual return; a saldo a favor never subtracts from it.
  ///
  /// **Changed since viewed** (driven only by `figures_fingerprint`, OQ-H4). For a month with a stored view: same figures fingerprint → `changed_since_viewed: false`, whatever the inputs or the engine did; different, and the engine version or parameter set differs from the one viewed → `true`, `change_reason: engine`; different otherwise → `true`, `change_reason: records` (when both causes apply, `engine` wins). A never-viewed month (`viewed_at: null`, OQ-H1) and the current month (OQ-H3) are never marked. `change_reason` is non-null if and only if `changed_since_viewed` is `true`. Opening a month and calling `markHistoryViewed` clears its marker; the marker also clears by itself when the figures return to the viewed ones.
  ///
  /// **Flags (server-enforced).** `module_history` requires `module_tax_engine` (and, through it, `module_taxpayers`): with any of them OFF the module is effectively OFF and this route answers `404`, indistinguishable from an unknown taxpayer; no tax figure leaves the server. `module_incomes` / `module_expenses` OFF never change the figures (the engine still reads the records); `has_records` then counts only the enabled modules' records, and the app hides the matching "Ver ingresos" / "Ver gastos" link.
  ///
  /// **Access.** Open to read-only users (no `402`) and to archived taxpayers. Per-user budget `history_reads`, 600 per hour (burst 600, ADR-0005), separate from `estimate_reads`; spent → `429 rate_limited` with `Retry-After`. No partial mode: if any month's estimate fails the whole request fails (`500`).
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscal_year` → `out_of_range` (a year after the current one, America/Mexico_City). `422 business_rule_violation` details: `fiscal_year` → `parameters_unavailable` (2022..2025: no published tax parameters).
  ///
  /// [fiscalYear] - Fiscal year to list (2026 .. the current year; the app's year picker offers only those). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @GET('/v1/taxpayers/{taxpayerId}/history')
  Future<HistoryYear> getHistory({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('fiscal_year') required FiscalYear fiscalYear,
  });

  /// Record that the caller saw a month's estimate (clears its changed marker).
  ///
  /// Stores, for the caller, the taxpayer and `period`, the fingerprints and versions of the estimate the app **rendered**: the body is copied from the `TaxpayerEstimate` shown (`records_fingerprint`, `figures_fingerprint`, `engine_version`, `parameter_set`), not recomputed by the server, so a record changed between render and POST still marks the month on the next `getHistory`. Idempotent: one upsert per (user, taxpayer, period), the last view wins; repeating it answers `204` again.
  ///
  /// The app calls it whenever the estimate screen of a past month renders successfully, **from any entry point** (History row, dashboard "Ver detalle"; OQ-H7), only while `module_history` is ON; fire-and-forget (a failure never blocks the screen). The current month is accepted and stored but is never marked. The values are only ever compared with later estimates of the caller's own month (RLS); they never feed a figure.
  ///
  /// **Flags.** Same as `getHistory`: `module_history`, `module_tax_engine` or `module_taxpayers` OFF → `404`.
  ///
  /// **Access.** Allowed for read-only users and archived taxpayers (it creates no record: no `402`, no cap). Per-user budget `history_views`, 1,200 per hour (burst 120, ADR-0005); spent → `429 rate_limited` with `Retry-After`.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `period` → `out_of_range` (a future month, America/Mexico_City). `422 business_rule_violation` details: `period` → `parameters_unavailable` (before 2026-01).
  @POST('/v1/taxpayers/{taxpayerId}/history/{period}/viewed')
  Future<void> markHistoryViewed({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('period') required Period period,
    @Body() required HistoryViewedRequest body,
  });
}
