import 'dart:convert';

import 'package:control_contable/api/generated/models/breakdown.dart';

import 'estimate_fixtures.dart';

/// A `Citation` JSON.
Map<String, Object?> citation(String law, String article, [String? note]) => {
  'law': law,
  'article': article,
  'note': note,
};

/// A `BreakdownLine` JSON.
Map<String, Object?> lineJson(
  int seq,
  String concept, {
  String tax = 'isr',
  String operation = 'input',
  int amountCents = 0,
  String? rate,
  List<Map<String, Object?>> legalBasis = const [],
  String source = 'computed',
  bool needsCpaReview = false,
}) => {
  'seq': seq,
  'tax': tax,
  'concept': concept,
  'operation': operation,
  'amount_cents': amountCents,
  'rate': rate,
  'legal_basis': legalBasis,
  'source': source,
  'needs_cpa_review': needsCpaReview,
};

/// The rounding adjustment line ("Ajuste por redondeo").
Map<String, Object?> roundingLine(int seq, {int cents = 1}) => lineJson(
  seq,
  'Ajuste por redondeo',
  operation: 'rounding_adjustment',
  amountCents: cents,
);

/// The ISR lines of PF-612 for March 2026 (the contract's example).
final List<Map<String, Object?>> isrLines612 = [
  lineJson(
    1,
    'Ingresos cobrados acumulados ene–mar',
    amountCents: 9000000,
    legalBasis: [citation('LISR', '102')],
    source: 'records',
  ),
  lineJson(
    2,
    'Deducciones pagadas acumuladas ene–mar',
    operation: 'subtract',
    amountCents: 3000000,
    legalBasis: [citation('LISR', '103'), citation('LISR', '105')],
    source: 'records',
  ),
  lineJson(
    3,
    'Base gravable',
    operation: 'result',
    amountCents: 6000000,
    legalBasis: [citation('LISR', '106', 'párrafo primero')],
  ),
  lineJson(
    4,
    'Cuota fija (tarifa acumulada de marzo, renglón 6)',
    operation: 'lookup',
    amountCents: 557052,
    legalBasis: [citation('Anexo 8 RMF 2026', 'B.VI')],
    source: 'params:isr.art106.accumulated',
  ),
  lineJson(
    5,
    'Excedente del límite inferior por la tasa',
    operation: 'multiply',
    amountCents: 158044,
    rate: '0.2136',
    legalBasis: [citation('Anexo 8 RMF 2026', 'B.VI')],
    source: 'params:isr.art106.accumulated',
  ),
  lineJson(
    6,
    'ISR causado acumulado',
    operation: 'result',
    amountCents: 715096,
    legalBasis: [citation('LISR', '106', 'párrafo primero')],
  ),
  lineJson(
    7,
    'Pagos provisionales anteriores (ene–feb)',
    operation: 'subtract',
    amountCents: 400000,
    legalBasis: [citation('LISR', '106', 'párrafo cuarto')],
    source: 'tax_input:prior_provisional_payments',
  ),
  lineJson(
    8,
    'Retenciones de ISR (10 %) acumuladas',
    operation: 'subtract',
    amountCents: 300000,
    legalBasis: [citation('LISR', '106', 'párrafo quinto')],
    source: 'records',
  ),
  lineJson(
    9,
    'ISR a pagar',
    operation: 'result',
    amountCents: 15096,
    legalBasis: [citation('LISR', '106')],
  ),
];

/// The IVA lines of PF-612 for March 2026; [consolidated] marks the month's
/// IVA as netted across two or more regimes (OQ-T2).
List<Map<String, Object?>> ivaLines612({bool consolidated = false}) => [
  lineJson(
    10,
    'IVA trasladado cobrado',
    tax: 'iva',
    amountCents: 480000,
    legalBasis: [citation('LIVA', '1'), citation('LIVA', '1-B')],
    source: 'records',
  ),
  lineJson(
    11,
    'IVA acreditable pagado',
    tax: 'iva',
    operation: 'subtract',
    amountCents: 160000,
    legalBasis: [citation('LIVA', '5', 'fracciones I a V')],
    source: 'records',
  ),
  lineJson(
    12,
    'IVA a cargo del mes (todos tus regímenes)',
    tax: 'iva',
    operation: 'result',
    amountCents: 320000,
    legalBasis: [citation('LIVA', '5-D')],
    needsCpaReview: consolidated,
  ),
];

/// Withholdings the taxpayer made and must remit.
final List<Map<String, Object?>> withholdingLines = [
  lineJson(
    13,
    'ISR retenido a tus trabajadores',
    tax: 'withholding',
    amountCents: 125000,
    legalBasis: [citation('LISR', '96')],
    source: 'records',
  ),
];

/// A `Breakdown` JSON (612 of TX-612-TOTAL-01, March 2026, by default).
Map<String, Object?> breakdownJson({
  Map<String, Object?>? regime,
  List<Map<String, Object?>>? lines,
  List<Map<String, Object?>> disclaimers = const [notTaxAdvice],
  String legend = legendText,
}) => {
  'taxpayer_id': fixtureTaxpayerId,
  'period': '2026-03',
  'regime':
      regime ??
      regimeJson(
        '612',
        name: 'Actividades Empresariales y Profesionales',
        isrCents: 15096,
      ),
  'lines': lines ?? [...isrLines612, ...ivaLines612()],
  'legend': legend,
  'disclaimers': disclaimers,
  'engine_version': '1.0.0',
};

/// [json] decoded the way the generated client decodes a response.
Breakdown breakdownFrom(Map<String, Object?> json) =>
    Breakdown.fromJson(jsonDecode(jsonEncode(json)) as Map<String, Object?>);
