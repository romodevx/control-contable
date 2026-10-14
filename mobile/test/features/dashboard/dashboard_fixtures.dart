import 'dart:convert';

import 'package:control_contable/api/generated/models/dashboard.dart';

import '../tax_engine/estimate_fixtures.dart';

/// A `TrendTax` JSON; the total is the sum of the three payable parts.
Map<String, Object?> trendTaxJson({
  int iva = 0,
  int isr = 0,
  int withholdings = 0,
  String ivaOutcome = 'payable',
  String status = 'complete',
}) => {
  'iva_payable_rounded_cents': iva,
  'isr_payable_rounded_cents': isr,
  'withholdings_to_remit_cents': withholdings,
  'total_payable_rounded_cents': iva + isr + withholdings,
  'iva_outcome': ivaOutcome,
  'status': status,
};

/// A `TrendPoint` JSON. [tax] `null` is the engine OFF.
Map<String, Object?> trendPointJson(
  String period, {
  int? income = 0,
  int? expense = 0,
  Map<String, Object?>? tax = const {},
}) => {
  'period': period,
  'income_cents': income,
  'expense_cents': expense,
  'tax': tax != null && tax.isEmpty ? trendTaxJson() : tax,
};

/// January..March 2026 with March as the current month (TX-612-TOTAL-01).
List<Map<String, Object?>> threeMonthTrend({bool engine = true}) => [
  trendPointJson(
    '2026-01',
    income: 1000000,
    expense: 400000,
    tax: engine ? trendTaxJson(iva: 96000, isr: 12000) : null,
  ),
  trendPointJson(
    '2026-02',
    income: 2500000,
    expense: 600000,
    tax: engine
        ? trendTaxJson(iva: 310000, isr: 40000, status: 'incomplete')
        : null,
  ),
  trendPointJson(
    '2026-03',
    income: 3000000,
    expense: 1000000,
    tax: engine ? trendTaxJson(iva: 213333, isr: 15096) : null,
  ),
];

/// An `AccessSummary` JSON (`pro` by default: no banner).
Map<String, Object?> accessJson({
  String access = 'pro',
  int? daysLeft,
  int? incomesRemaining,
  int? expensesRemaining,
}) => {
  'access': access,
  'trial_ends_at': daysLeft == null ? null : '2026-03-23T06:00:00Z',
  'trial_days_left': daysLeft,
  'incomes_remaining': incomesRemaining,
  'expenses_remaining': expensesRemaining,
};

/// A `Dashboard` JSON for March 2026, engine ON, both modules ON.
Map<String, Object?> dashboardJson({
  bool isEmpty = false,
  int? monthIncome = 3000000,
  int? monthExpense = 1000000,
  Map<String, Object?>? estimate,
  bool engine = true,
  List<Map<String, Object?>>? trend,
  Map<String, Object?>? access,
}) => {
  'taxpayer_id': fixtureTaxpayerId,
  'period': '2026-03',
  'is_empty': isEmpty,
  'month': {'income_cents': monthIncome, 'expense_cents': monthExpense},
  'estimate': engine ? (estimate ?? estimateJson()) : null,
  'trend': trend ?? threeMonthTrend(engine: engine),
  'access': access ?? accessJson(),
  if (engine) 'legend': legendText,
};

/// [json] decoded the way the generated client decodes a response.
Dashboard dashboardFrom(Map<String, Object?> json) =>
    Dashboard.fromJson(jsonDecode(jsonEncode(json)) as Map<String, Object?>);
