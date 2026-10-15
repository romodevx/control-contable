import 'dart:convert';

import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/personal_deduction.dart';
import 'package:control_contable/api/generated/models/personal_deduction_list.dart';
import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/salary_statement_list.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;

const String _stamp = '2026-10-13T16:00:00Z';

/// A `PersonalDeduction` JSON: a $1,500.00 medical fee paid electronically
/// for the taxpayer herself by default.
Map<String, Object?> deductionJson({
  String id = 'ded-1',
  String category = 'medical',
  int amountCents = 150000,
  String paidOn = '2026-03-10',
  String paymentMethod = 'electronic',
  String beneficiary = 'self',
  String? doneeType,
  String? studentLabel,
  String? tuitionLevel,
  int? scholarshipCents,
  String? cfdiUuid,
  String? excluded,
}) => {
  'id': id,
  'taxpayer_id': fixtureTaxpayerId,
  'fiscal_year': 2026,
  'category': category,
  'amount_cents': amountCents,
  'paid_on': paidOn,
  'payment_method': paymentMethod,
  'beneficiary': beneficiary,
  'donee_type': doneeType,
  'student_label': studentLabel,
  'tuition_level': tuitionLevel,
  'scholarship_cents': scholarshipCents,
  'cfdi_uuid': cfdiUuid,
  'excluded': excluded,
  'warnings': <Object>[],
  'created_at': _stamp,
  'updated_at': _stamp,
};

/// [json] decoded as the client decodes it.
PersonalDeduction deductionFrom(Map<String, Object?> json) =>
    PersonalDeduction.fromJson(_wire(json));

/// A `PersonalDeductionList` of [items], with totals per category.
PersonalDeductionList deductionListOf(List<Map<String, Object?>> items) {
  final totals = <String, Map<String, Object?>>{};
  for (final item in items) {
    final category = item['category']! as String;
    final amount = item['amount_cents']! as int;
    final total = totals.putIfAbsent(
      category,
      () => {
        'category': category,
        'registered_cents': 0,
        'excluded_cents': 0,
        'count': 0,
      },
    );
    total['registered_cents'] = (total['registered_cents']! as int) + amount;
    total['count'] = (total['count']! as int) + 1;
    if (item['excluded'] != null) {
      total['excluded_cents'] = (total['excluded_cents']! as int) + amount;
    }
  }
  return PersonalDeductionList.fromJson(
    _wire({
      'fiscal_year': 2026,
      'items': items,
      'totals': totals.values.toList(),
    }),
  );
}

/// A `SalaryStatement` JSON: one employer's 2026 constancia.
Map<String, Object?> statementJson({
  String id = 'sal-1',
  String? employerLabel = 'Empresa A',
  int accumulableCents = 45000000,
  int exemptCents = 1500000,
  int isrWithheldCents = 8000000,
  bool simultaneous = false,
  bool leftBeforeYearEnd = false,
  bool payerNotWithholding = false,
}) => {
  'id': id,
  'taxpayer_id': fixtureTaxpayerId,
  'fiscal_year': 2026,
  'employer_label': employerLabel,
  'accumulable_cents': accumulableCents,
  'exempt_cents': exemptCents,
  'isr_withheld_cents': isrWithheldCents,
  'simultaneous': simultaneous,
  'left_before_year_end': leftBeforeYearEnd,
  'payer_not_withholding': payerNotWithholding,
  'warnings': <Object>[],
  'created_at': _stamp,
  'updated_at': _stamp,
};

/// [json] decoded as the client decodes it.
SalaryStatement statementFrom(Map<String, Object?> json) =>
    SalaryStatement.fromJson(_wire(json));

/// A `SalaryStatementList` of [items] with their totals.
SalaryStatementList statementListOf(List<Map<String, Object?>> items) {
  int sum(String key) =>
      items.fold(0, (total, item) => total + (item[key]! as int));
  return SalaryStatementList.fromJson(
    _wire({
      'fiscal_year': 2026,
      'items': items,
      'totals': {
        'accumulable_cents': sum('accumulable_cents'),
        'exempt_cents': sum('exempt_cents'),
        'isr_withheld_cents': sum('isr_withheld_cents'),
      },
    }),
  );
}

/// One `FiledPaymentMonth` JSON: by default an in-window month the app
/// estimated at $1,000.00 and the user has not confirmed.
Map<String, Object?> paymentMonthJson(
  int month, {
  bool inWindow = true,
  int? filedCents,
  int? taxInputCents,
  int? appEstimateCents = 100000,
  bool appEstimateIncomplete = false,
  String? source,
}) {
  final effective = filedCents ?? taxInputCents ?? appEstimateCents;
  return {
    'month': month,
    'period': '2026-${month.toString().padLeft(2, '0')}',
    'in_window': inWindow,
    'filed_cents': filedCents,
    'tax_input_cents': taxInputCents,
    'app_estimate_cents': appEstimateCents,
    'app_estimate_incomplete': appEstimateIncomplete,
    'effective_cents': inWindow ? effective : null,
    'source':
        source ??
        switch ((inWindow, filedCents, taxInputCents, appEstimateCents)) {
          (false, _, _, _) => 'none',
          (_, final int _, _, _) => 'filed',
          (_, _, final int _, _) => 'tax_input',
          (_, _, _, final int _) => 'estimated_by_app',
          _ => 'none',
        },
  };
}

/// The filed payments of [regime] in 2026 over [months] (twelve app
/// estimates by default).
FiledPayments filedPaymentsOf({
  String regime = '612',
  List<Map<String, Object?>>? months,
  bool saved = false,
}) {
  final rows = months ?? [for (var m = 1; m <= 12; m++) paymentMonthJson(m)];
  final total = rows.fold(
    0,
    (sum, row) => sum + ((row['effective_cents'] as int?) ?? 0),
  );
  return FiledPayments.fromJson(
    _wire({
      'regime_code': regime,
      'fiscal_year': 2026,
      'months': rows,
      'total_effective_cents': total,
      'updated_at': saved ? _stamp : null,
    }),
  );
}

/// One `AnnualRegimeTaxInput` JSON.
Map<String, Object?> regimeTaxInputJson(
  String key,
  String label, {
  String regime = '612',
  int? valueCents,
}) => {
  'regime_code': regime,
  'key': key,
  'label': label,
  'value_cents': valueCents,
};

/// The annual facts of 2026: never saved by default.
AnnualInputs annualFactsOf({
  int? priorYearCents,
  int otherExemptCents = 0,
  bool toldEmployer = false,
  List<Map<String, Object?>>? regimeInputs,
  bool saved = false,
}) => AnnualInputs.fromJson(
  _wire({
    'fiscal_year': 2026,
    'prior_year_accumulable_income_cents': priorYearCents,
    'other_exempt_income_cents': otherExemptCents,
    'told_employer_will_file': toldEmployer,
    'regime_tax_inputs':
        regimeInputs ??
        [
          regimeTaxInputJson(
            'prior_year_losses',
            'Pérdidas fiscales de años anteriores',
          ),
          regimeTaxInputJson('ptu_paid', 'PTU pagada', valueCents: 1200000),
          regimeTaxInputJson('annual_depreciation', 'Depreciación del año'),
        ],
    'updated_at': saved ? _stamp : null,
  }),
);

Map<String, Object?> _wire(Map<String, Object?> json) =>
    jsonDecode(jsonEncode(json)) as Map<String, Object?>;
