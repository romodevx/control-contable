import 'dart:convert';

import 'package:control_contable/api/generated/models/tax_inputs.dart';

/// A `TaxInputDef` JSON.
Map<String, Object?> defJson(
  String key,
  String label,
  String kind, {
  bool required = false,
}) => {'key': key, 'label': label, 'kind': kind, 'required': required};

/// A stored `TaxInputValue` JSON (responses send every field).
Map<String, Object?> valueJson(
  String key, {
  int? month,
  int? cents,
  String? rate,
  bool? flag,
}) => {
  'key': key,
  'month': month,
  'value_cents': cents,
  'value_rate': rate,
  'value_bool': flag,
};

/// The monthly prior payments definition of 612, 626 PM and 601.
final Map<String, Object?> priorPaymentsDef = defJson(
  'prior_provisional_payments',
  'Pagos provisionales anteriores del ejercicio',
  'money_monthly',
  required: true,
);

/// The definitions of 612 PF (the contract's example).
final List<Map<String, Object?>> defs612 = [
  priorPaymentsDef,
  defJson(
    'ytd_before_app_income',
    'Ingresos del ejercicio antes de usar la app',
    'money',
  ),
  defJson('ptu_paid', 'PTU pagada en el ejercicio', 'money'),
];

/// The definitions of 601 PM.
final List<Map<String, Object?>> defs601 = [
  defJson(
    'coeficiente_utilidad',
    'Coeficiente de utilidad',
    'rate',
    required: true,
  ),
  defJson('first_fiscal_year', 'Primer ejercicio fiscal', 'boolean'),
  priorPaymentsDef,
];

/// A `TaxInputs` JSON (612, 2026, January paid by default).
Map<String, Object?> taxInputsJson({
  String regime = '612',
  int fiscalYear = 2026,
  List<Map<String, Object?>>? definitions,
  List<Map<String, Object?>>? values,
}) => {
  'regime_code': regime,
  'fiscal_year': fiscalYear,
  'definitions': definitions ?? defs612,
  'values':
      values ??
      [valueJson('prior_provisional_payments', month: 1, cents: 200000)],
};

/// [json] decoded the way the generated client decodes a response.
TaxInputs taxInputsFrom(Map<String, Object?> json) =>
    TaxInputs.fromJson(jsonDecode(jsonEncode(json)) as Map<String, Object?>);
