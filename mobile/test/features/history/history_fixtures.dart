import 'dart:convert';

import 'package:control_contable/api/generated/models/history_year.dart';

import '../tax_engine/estimate_fixtures.dart';

/// The records fingerprint every history month carries (opaque to the app).
const String historyRecordsFingerprint =
    'e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2';

/// A `HistoryTax` JSON; the total is the sum of the three payable parts.
Map<String, Object?> historyTaxJson({
  int isr = 15096,
  int iva = 213333,
  String ivaOutcome = 'payable',
  int ivaInFavor = 0,
  int withholdings = 0,
}) => {
  'isr_payable_rounded_cents': isr,
  'iva_payable_rounded_cents': iva,
  'iva_outcome': ivaOutcome,
  'iva_in_favor_cents': ivaInFavor,
  'withholdings_to_remit_cents': withholdings,
  'total_payable_rounded_cents': isr + iva + withholdings,
};

/// A `HistoryMonth` JSON: complete, with records, never viewed. [tax]
/// `null` is a month with no figure (`none`, informational).
Map<String, Object?> historyMonthJson(
  String period, {
  String status = 'complete',
  bool hasRecords = true,
  Map<String, Object?>? tax = const {},
  bool isCurrent = false,
  String? viewedAt,
  String? changeReason,
  bool cpaReviewPending = false,
  String dueDate = '2026-04-17',
}) => {
  'period': period,
  'status': status,
  'has_records': hasRecords,
  'tax': tax != null && tax.isEmpty ? historyTaxJson() : tax,
  'due_date': dueDate,
  'cpa_review_pending': cpaReviewPending,
  'is_current': isCurrent,
  'records_fingerprint': historyRecordsFingerprint,
  'figures_fingerprint': fixtureFiguresFingerprint,
  'engine_version': '1.0.0',
  'parameter_set': {'fiscal_year': 2026, 'version': 1},
  'viewed_at': viewedAt,
  'changed_since_viewed': changeReason != null,
  'change_reason': changeReason,
};

/// March 2026 (current) down to January: February viewed and changed by a
/// late record, January viewed and unchanged.
List<Map<String, Object?>> threeMonthHistory() => [
  historyMonthJson('2026-03', isCurrent: true),
  historyMonthJson(
    '2026-02',
    tax: historyTaxJson(isr: 40000, iva: 310000),
    viewedAt: '2026-03-03T17:00:00Z',
    changeReason: 'records',
    dueDate: '2026-03-17',
  ),
  historyMonthJson(
    '2026-01',
    tax: historyTaxJson(isr: 12000, iva: 96000),
    viewedAt: '2026-02-14T17:00:00Z',
    dueDate: '2026-02-17',
  ),
];

/// A `HistoryYear` JSON for 2026 ([threeMonthHistory] by default).
Map<String, Object?> historyYearJson({
  List<Map<String, Object?>>? months,
  List<Map<String, Object?>> disclaimers = const [notTaxAdvice],
  int fiscalYear = 2026,
  String taxpayerId = fixtureTaxpayerId,
}) {
  final shown = months ?? threeMonthHistory();
  int sum(String key) => shown.fold(0, (total, month) {
    final tax = month['tax'] as Map<String, Object?>?;
    return total + ((tax?[key] as int?) ?? 0);
  });
  return {
    'taxpayer_id': taxpayerId,
    'fiscal_year': fiscalYear,
    'months': shown,
    'year_totals': {
      'isr_payable_rounded_cents': sum('isr_payable_rounded_cents'),
      'iva_payable_rounded_cents': sum('iva_payable_rounded_cents'),
      'withholdings_to_remit_cents': sum('withholdings_to_remit_cents'),
      'total_payable_rounded_cents': sum('total_payable_rounded_cents'),
    },
    'legend': legendText,
    'disclaimers': disclaimers,
  };
}

/// [json] decoded the way the generated client decodes a response.
HistoryYear historyFrom(Map<String, Object?> json) =>
    HistoryYear.fromJson(jsonDecode(jsonEncode(json)) as Map<String, Object?>);
