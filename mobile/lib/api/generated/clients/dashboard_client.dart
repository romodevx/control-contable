// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/dashboard.dart';
import '../models/period.dart';
import '../models/uuid.dart';

part 'dashboard_client.g.dart';

@RestApi()
abstract class DashboardClient {
  factory DashboardClient(Dio dio, {String? baseUrl}) = _DashboardClient;

  /// Home of one taxpayer - this month's estimate, income vs expense, trend and access.
  ///
  /// Read-only Home for the taxpayer, derived on every read (no cache, ADR-0002 §5). `period` is optional: absent means the current month in America/Mexico_City (the app always omits it). The trend window is `[max(period - (months - 1), 2026-01), period]`, oldest first, one point per month with gaps filled with zeros, so the last point is always `period` (in October 2026 the default window has 10 points).
  ///
  /// **Figures.** Income is the sum of `amount_cents` (before IVA) of live incomes *collected* in the month (`collected_on`, cash basis, "Ingresos cobrados"); expense the sum of `amount_cents` (before IVA) of live expenses *paid* in the month (`paid_on`, deductible or not, "Gastos pagados"). Informational regimes count in income and expense, never in tax figures. `estimate` is the same `TaxpayerEstimate` that `getTaxpayerEstimate` answers for `period` (same totals to the centavo); the client renders it as is and links to the estimate screen with its `period`. Each `trend[].tax` carries the taxpayer-level status (`informational` > `incomplete` > `partial` > `complete`; `none` when no regime is held that month); the hero chip is the status of the last point.
  ///
  /// **Flags (server-enforced; nothing is computed for a module that is OFF).** `module_tax_engine` OFF: `estimate` is `null`, every `trend[].tax` is `null` and `legend` is absent; no tax figure leaves the server. `module_incomes` OFF: `month.income_cents` and every `trend[].income_cents` are `null` (same for `module_expenses` and the expense fields); tax figures are unaffected. `is_empty` counts only the enabled modules' records (`false` when both are OFF). `module_dashboard` OFF, or `module_taxpayers` OFF (prerequisite), answers `404`.
  ///
  /// **Access.** Open to read-only users (no `402`) and to archived taxpayers. `access` is the caller's `AccessSummary` (per user, across taxpayers), with the server-computed `trial_days_left`. Per-user budget `dashboard_reads`, 600 per hour (burst 600, ADR-0005), separate from `estimate_reads`; spent → `429 rate_limited` with `Retry-After`. If any part fails the whole request fails (`500`); a Home with silently missing figures is never returned.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `period` → `out_of_range` (a future month, America/Mexico_City). `422 business_rule_violation` details: `period` → `parameters_unavailable` (before 2026-01).
  ///
  /// [period] - Month the Home is for (`YYYY-MM`). Absent → the current month (America/Mexico_City).
  ///
  /// [months] - Trend window length in months, ending at `period`. Months before 2026-01 are never included, so the window may hold fewer points.
  @GET('/v1/taxpayers/{taxpayerId}/dashboard')
  Future<Dashboard> getDashboard({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('months') int? months = 12,
    @Query('period') Period? period,
  });
}
