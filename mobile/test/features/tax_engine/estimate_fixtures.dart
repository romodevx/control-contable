import 'dart:convert';

import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';

/// The taxpayer every estimate fixture belongs to.
const String fixtureTaxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

/// The mandatory legend as the server sends it.
const String legendText =
    'Esto es un aproximado. No todos los gastos aplican para tu giro '
    'comercial.';

/// `not_tax_advice`, always first.
const Map<String, Object?> notTaxAdvice = {
  'code': 'not_tax_advice',
  'severity': 'info',
  'text':
      'Estimación informativa. No sustituye la declaración ante el SAT ni la '
      'asesoría de un contador.',
};

/// `incomplete_inputs` (warning).
const Map<String, Object?> incompleteInputsDisclaimer = {
  'code': 'incomplete_inputs',
  'severity': 'warning',
  'text':
      'Faltan datos para este cálculo; se usaron ceros. Complétalos para un '
      'resultado más preciso.',
};

/// `provisional_parameters` (warning).
const Map<String, Object?> provisionalParametersDisclaimer = {
  'code': 'provisional_parameters',
  'severity': 'warning',
  'text': 'Aún no se publican las tarifas de 2027; usamos las de 2026.',
};

/// `cpa_review_pending` (strong).
const Map<String, Object?> cpaReviewDisclaimer = {
  'code': 'cpa_review_pending',
  'severity': 'strong',
  'text':
      'Este cálculo usa reglas que un contador aún no valida: compensación '
      'del IVA entre regímenes. Tómalo con mayor reserva.',
};

/// A `TaxResult` JSON.
Map<String, Object?> taxResult(int cents) => {
  'amount_cents': cents,
  'payable_rounded_cents': cents > 0 ? cents : 0,
  'outcome': cents > 0 ? 'payable' : (cents < 0 ? 'in_favor' : 'zero'),
};

/// The IVA components of PF-612-IVA-01.
const Map<String, Object?> ivaComponents612 = {
  'charged_collected_cents': 480000,
  'creditable_paid_cents': 160000,
  'withheld_from_taxpayer_cents': 106667,
  'definitive_excluded_cents': 0,
};

/// A `RegimeEstimate` JSON.
Map<String, Object?> regimeJson(
  String code, {
  String? name,
  String treatment = 'full',
  String status = 'complete',
  List<Map<String, Object?>> missingInputs = const [],
  String? explanation,
  int? isrCents,
  Map<String, Object?>? iva = ivaComponents612,
  int withheldToRemitCents = 0,
  List<String> cpaTopics = const [],
}) => {
  'regime_code': code,
  'regime_name': name ?? 'Régimen $code',
  'treatment': treatment,
  'status': status,
  'missing_inputs': missingInputs,
  'explanation': explanation,
  'isr': isrCents == null ? null : taxResult(isrCents),
  'iva_components': iva,
  'withheld_to_remit_cents': withheldToRemitCents,
  'cpa_review_pending': cpaTopics.isNotEmpty,
  'cpa_topics': cpaTopics,
  'parameter_set': {'fiscal_year': 2026, 'version': 1},
};

/// A `TaxpayerEstimate` JSON (TX-612-TOTAL-01 by default).
Map<String, Object?> estimateJson({
  String period = '2026-03',
  String dueDate = '2026-04-17',
  List<Map<String, Object?>> disclaimers = const [notTaxAdvice],
  List<Map<String, Object?>>? regimes,
  int ivaCents = 213333,
  int isrCents = 15096,
  int withholdingsToRemitCents = 0,
  int? totalPayableCents,
}) => {
  'taxpayer_id': fixtureTaxpayerId,
  'period': period,
  'due_date': dueDate,
  'legend': legendText,
  'disclaimers': disclaimers,
  'regimes':
      regimes ??
      [
        regimeJson(
          '612',
          name: 'Actividades Empresariales y Profesionales',
          isrCents: 15096,
        ),
      ],
  'totals': {
    'iva': taxResult(ivaCents),
    'isr': taxResult(isrCents),
    'withholdings_to_remit_cents': withholdingsToRemitCents,
    'total_payable_rounded_cents':
        totalPayableCents ??
        (ivaCents > 0 ? ivaCents : 0) +
            (isrCents > 0 ? isrCents : 0) +
            withholdingsToRemitCents,
  },
  'parameter_set': {'fiscal_year': 2026, 'version': 1},
  'records_fingerprint':
      '3f6c2a9e8d7b1c4f5a0e9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2e1d0c9b8a',
  'figures_fingerprint':
      'cccccccccccccccccccccccccccccccccccccccccccccccccccccccccc202603',
  'engine_version': '1.0.0',
};

/// TX-612-TOTAL-01 as the wire sends it.
final Map<String, Object?> completeEstimateJson = estimateJson();

/// [json] decoded the way the generated client decodes a response.
TaxpayerEstimate estimateFrom(Map<String, Object?> json) =>
    TaxpayerEstimate.fromJson(
      jsonDecode(jsonEncode(json)) as Map<String, Object?>,
    );
