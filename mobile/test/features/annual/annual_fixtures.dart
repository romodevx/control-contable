import 'dart:convert';

import 'package:control_contable/api/generated/models/annual_breakdown.dart';
import 'package:control_contable/api/generated/models/annual_estimate.dart';

import '../tax_engine/breakdown_fixtures.dart' show citation;
import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;

/// The mandatory annual legend as the server sends it.
const String annualLegendText =
    'Esto es un aproximado; tu contador te dará las cifras reales.';

/// `annual_estimate`, always first.
const Map<String, Object?> annualEstimateDisclaimer = {
  'code': 'annual_estimate',
  'severity': 'info',
  'text':
      'Estimación informativa de la declaración anual 2026. No sustituye la '
      'declaración que presenta tu contador ante el SAT.',
};

/// `no_annual_iva`, always present.
const Map<String, Object?> noAnnualIvaDisclaimer = {
  'code': 'no_annual_iva',
  'severity': 'info',
  'text': 'El IVA no tiene declaración anual: se paga mes con mes.',
};

/// `year_in_progress` (warning).
const Map<String, Object?> yearInProgressDisclaimer = {
  'code': 'year_in_progress',
  'severity': 'warning',
  'text':
      'El año aún no termina: calculamos con tus registros hasta septiembre '
      'de 2026.',
};

/// `payments_estimated` (warning).
const Map<String, Object?> paymentsEstimatedDisclaimer = {
  'code': 'payments_estimated',
  'severity': 'warning',
  'text':
      'Usamos lo que la app estimó para 3 meses; confírmalo con tus acuses '
      'para un resultado más preciso.',
};

/// `incomplete_inputs` (warning).
const Map<String, Object?> annualIncompleteDisclaimer = {
  'code': 'incomplete_inputs',
  'severity': 'warning',
  'text':
      'Faltan datos; usamos ceros o dejamos fuera lo que no pudimos '
      'verificar.',
};

/// `cpa_review_pending` (strong) naming the RESICO relief renewal.
const Map<String, Object?> annualCpaDisclaimer = {
  'code': 'cpa_review_pending',
  'severity': 'strong',
  'text':
      'Este cálculo usa reglas que un contador aún no valida: renovación de '
      'la facilidad de RESICO en la RMF 2027.',
};

const Map<String, Object?> _parameterSet = {'fiscal_year': 2026, 'version': 1};

/// An `AnnualTax` JSON; AN-01 by default.
Map<String, Object?> annualTaxJson({
  int accumulableCents = 50000000,
  int? personalDeductionsCents = 4000000,
  int baseCents = 46000000,
  int isrCents = 7636586,
  int paymentsCents = 5577386,
  int withholdingsCents = 3000000,
  int salaryIsrCents = 0,
  int filedMonths = 12,
  int estimatedMonths = 0,
  String balanceKind = 'a_favor',
  int balanceCents = 940800,
}) => {
  'accumulable_income_cents': accumulableCents,
  'personal_deductions_cents': personalDeductionsCents,
  'base_cents': baseCents,
  'isr_cents': isrCents,
  'credits': {
    'payments_cents': paymentsCents,
    'withholdings_cents': withholdingsCents,
    'salary_isr_cents': salaryIsrCents,
    'total_cents': paymentsCents + withholdingsCents + salaryIsrCents,
    'payment_sources': {
      'filed_months': filedMonths,
      'tax_input_months': 0,
      'estimated_by_app_months': estimatedMonths,
    },
  },
  'balance': {'kind': balanceKind, 'amount_cents': balanceCents},
};

/// An `AnnualSection` JSON; AN-01's `pf_art152` by default.
Map<String, Object?> sectionJson({
  String kind = 'pf_art152',
  List<String> regimes = const ['612'],
  String status = 'complete',
  List<Map<String, Object?>> missingInputs = const [],
  String? obligation = 'obligated',
  List<String> obligationReasons = const [
    'Tuviste ingresos por actividad empresarial o profesional.',
  ],
  String? deadline = '2027-04-30',
  Map<String, Object?>? tax,
  bool noTax = false,
  Map<String, Object?>? pmExtras,
  Map<String, Object?>? resicoCheck,
  String? explanation,
  List<Map<String, Object?>> notes = const [],
  List<String> cpaTopics = const [],
}) => {
  'kind': kind,
  'regimes': regimes,
  'status': status,
  'missing_inputs': missingInputs,
  'obligation': obligation,
  'obligation_reasons': obligationReasons,
  'deadline': deadline,
  'tax': noTax ? null : (tax ?? annualTaxJson()),
  'pm_extras': pmExtras,
  'resico_check': resicoCheck,
  'explanation': explanation,
  'notes': notes,
  'cpa_review_pending': cpaTopics.isNotEmpty,
  'cpa_topics': cpaTopics,
  'parameter_set': _parameterSet,
};

/// An `AnnualMissingInput` JSON.
Map<String, Object?> missingJson(
  String key,
  String label,
  String target, {
  String? regimeCode,
  List<String> months = const [],
}) => {
  'key': key,
  'label': label,
  'target': target,
  'regime_code': regimeCode,
  'months': months,
};

/// RESICO PF's informative check (AN-03 case a).
Map<String, Object?> resicoSectionJson() => sectionJson(
  kind: 'pf_resico_check',
  regimes: ['626'],
  status: 'informational',
  obligation: 'relieved',
  obligationReasons: [],
  deadline: null,
  noTax: true,
  resicoCheck: {
    'income_cents': 48000000,
    'rate': '0.0110',
    'annual_isr_cents': 528000,
    'monthly_payments_cents': 660000,
    'withheld_cents': 0,
    'difference_cents': -132000,
  },
  explanation:
      'Tus pagos mensuales de RESICO son definitivos; no presentas anual por '
      'estos ingresos (regla 3.13.7 RMF 2026).',
  cpaTopics: ['resico_pf_relief_renewal'],
);

/// PM 601 (AN-06): saldo a cargo, PTU and suggested CU.
Map<String, Object?> pm601SectionJson() => sectionJson(
  kind: 'pm_601',
  regimes: ['601'],
  obligationReasons: [],
  deadline: '2027-03-31',
  tax: annualTaxJson(
    accumulableCents: 1000000000,
    personalDeductionsCents: null,
    baseCents: 118000000,
    isrCents: 35400000,
    paymentsCents: 33000000,
    withholdingsCents: 200000,
    balanceKind: 'a_cargo',
    balanceCents: 2200000,
  ),
  pmExtras: {
    'ptu_base_cents': 150000000,
    'ptu_to_distribute_cents': 15000000,
    'suggested_cu_next_year': '0.1380',
  },
  notes: [
    {
      'code': 'cu_suggestion',
      'text': 'Sugerido para tus pagos de 2027; confírmalo con tu contador.',
    },
  ],
  cpaTopics: ['cu_rounding', 'ptu_rate'],
);

/// An explanation-only regime (603).
Map<String, Object?> explanationSectionJson() => sectionJson(
  kind: 'explanation',
  regimes: ['603'],
  status: 'informational',
  obligation: null,
  obligationReasons: [],
  deadline: '2027-02-15',
  noTax: true,
  explanation:
      'Presentas una declaración informativa a más tardar el 15 de febrero '
      'de 2027; tu contador la prepara.',
);

/// An `AnnualEstimate` JSON; AN-01 (PF 612, closed 2026, saldo a favor).
Map<String, Object?> annualEstimateJson({
  int fiscalYear = 2026,
  String personType = 'fisica',
  List<Map<String, Object?>>? disclaimers,
  String yearStatus = 'closed',
  String? dataThrough = '2026-12',
  String? deadline = '2027-04-30',
  List<String> headline = const ['pf_art152'],
  List<Map<String, Object?>>? sections,
}) => {
  'taxpayer_id': fixtureTaxpayerId,
  'fiscal_year': fiscalYear,
  'person_type': personType,
  'legend': annualLegendText,
  'disclaimers':
      disclaimers ?? [annualEstimateDisclaimer, noAnnualIvaDisclaimer],
  'year_status': yearStatus,
  'data_through': dataThrough,
  'deadline': deadline,
  'headline_sections': headline,
  'sections': sections ?? [sectionJson()],
  'informative_income_over_500k': false,
  'parameter_set': _parameterSet,
  'engine_version': '1.1.0',
};

/// [json] decoded the way the generated client decodes it.
AnnualEstimate annualEstimateFrom(Map<String, Object?> json) =>
    AnnualEstimate.fromJson(_wire(json));

/// An `AnnualBreakdownLine` JSON.
Map<String, Object?> annualLineJson(
  int seq,
  String concept, {
  String? regimeCode,
  String operation = 'input',
  int amountCents = 0,
  String? rate,
  List<Map<String, Object?>>? legalBasis,
  String source = 'computed',
  bool needsCpaReview = false,
}) => {
  'seq': seq,
  'regime_code': regimeCode,
  'concept': concept,
  'operation': operation,
  'amount_cents': amountCents,
  'rate': rate,
  'legal_basis': legalBasis ?? [citation('LISR', '152')],
  'source': source,
  'needs_cpa_review': needsCpaReview,
};

/// The AN-01 trace, abridged: income, a capped deduction, the tariff, one
/// payment per source and the balance.
final List<Map<String, Object?>> an01Lines = [
  annualLineJson(
    1,
    'Ingresos acumulables de actividad empresarial',
    regimeCode: '612',
    amountCents: 50000000,
    source: 'records',
  ),
  annualLineJson(
    2,
    r'Tope aplicado: colegiaturas, de $30,000.00 a $12,900.00',
    operation: 'info',
    amountCents: 1290000,
    legalBasis: const [],
    source: 'personal_deductions',
  ),
  annualLineJson(
    3,
    'Deducciones personales',
    operation: 'subtract',
    amountCents: 4000000,
    source: 'personal_deductions',
  ),
  annualLineJson(
    4,
    'Tasa sobre el excedente',
    operation: 'multiply',
    amountCents: 2000000,
    rate: '0.3000',
    source: 'params:isr.art152.annual',
  ),
  annualLineJson(
    5,
    'Pagos provisionales de ISR (enero a junio)',
    regimeCode: '612',
    operation: 'subtract',
    amountCents: 3000000,
    source: 'filed_payment',
  ),
  annualLineJson(
    6,
    'Pagos provisionales de ISR (julio a septiembre, estimados por la app)',
    regimeCode: '612',
    operation: 'subtract',
    amountCents: 1560000,
    source: 'estimated_by_app',
  ),
  annualLineJson(
    7,
    'ISR retenido por tus patrones',
    operation: 'subtract',
    amountCents: 500000,
    source: 'salary_statements',
  ),
  annualLineJson(
    8,
    'Saldo a favor',
    operation: 'result',
    amountCents: -940800,
    needsCpaReview: true,
  ),
];

/// An `AnnualBreakdown` JSON of AN-01's `pf_art152`.
Map<String, Object?> annualBreakdownJson({
  Map<String, Object?>? section,
  List<Map<String, Object?>>? lines,
  List<Map<String, Object?>>? disclaimers,
}) => {
  'taxpayer_id': fixtureTaxpayerId,
  'fiscal_year': 2026,
  'section': section ?? sectionJson(),
  'lines': lines ?? an01Lines,
  'legend': annualLegendText,
  'disclaimers':
      disclaimers ?? [annualEstimateDisclaimer, noAnnualIvaDisclaimer],
  'engine_version': '1.1.0',
};

/// [json] decoded the way the generated client decodes it.
AnnualBreakdown annualBreakdownFrom(Map<String, Object?> json) =>
    AnnualBreakdown.fromJson(_wire(json));

/// [json] after a trip through the wire (nested maps typed as decoded).
Map<String, Object?> _wire(Map<String, Object?> json) =>
    jsonDecode(jsonEncode(json)) as Map<String, Object?>;
