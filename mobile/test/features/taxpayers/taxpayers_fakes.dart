import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/regime_incompatibility.dart';
import 'package:control_contable/api/generated/models/regime_option_def.dart';
import 'package:control_contable/api/generated/models/regime_option_def_type.dart';
import 'package:control_contable/api/generated/models/tax_input_def.dart';
import 'package:control_contable/api/generated/models/tax_input_def_kind.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/api/generated/models/treatment.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';

/// A catalog regime for tests.
Regime regime(
  String code, {
  String? name,
  PersonType personType = PersonType.fisica,
  Treatment treatment = Treatment.full,
  String? explanation,
  List<RegimeOptionDef> optionDefs = const [],
  List<TaxInputDef> requiredInputs = const [],
  List<RegimeIncompatibility> incompatibleWith = const [],
}) => Regime(
  code: code,
  name: name ?? 'Régimen $code',
  personType: personType,
  treatment: treatment,
  explanation: explanation,
  optionDefs: optionDefs,
  requiredInputs: requiredInputs,
  incompatibleWith: incompatibleWith,
);

/// An option definition for tests.
RegimeOptionDef optionDef(
  String key, {
  required RegimeOptionDefType type,
  String? label,
  bool required = true,
  List<String>? allowed,
  String? min,
  String? max,
  Object? defaultValue,
}) => RegimeOptionDef(
  key: key,
  label: label ?? key,
  type: type,
  requiredValue: required,
  allowedValues: allowed,
  min: min,
  max: max,
  defaultValue: defaultValue,
);

/// A per-year tax input for tests.
TaxInputDef inputDef(
  String key, {
  required String label,
  TaxInputDefKind kind = TaxInputDefKind.money,
}) => TaxInputDef(key: key, label: label, kind: kind, requiredValue: false);

/// The persona física catalog of the onboarding tests (one regime per
/// support level, the 606/612/625/626 options of MODULE-taxpayers §2.4,
/// and the 626 × 612 incompatibility).
final List<Regime> fisicaCatalog = [
  regime(
    '606',
    name: 'Arrendamiento',
    optionDefs: [
      optionDef(
        'deduction_mode',
        label: 'Forma de deducir',
        type: RegimeOptionDefType.valueEnum,
        allowed: ['blind_35', 'actual'],
        defaultValue: 'blind_35',
      ),
    ],
  ),
  regime(
    '612',
    name: 'Actividades empresariales y profesionales',
    optionDefs: [
      optionDef(
        'iva_fully_taxable',
        label: 'Todo lo que cobro causa IVA',
        type: RegimeOptionDefType.boolean,
        defaultValue: true,
      ),
    ],
    incompatibleWith: [
      const RegimeIncompatibility(
        code: '626',
        reason: 'El RESICO sustituye a este régimen.',
      ),
    ],
  ),
  regime(
    '625',
    name: 'Plataformas tecnológicas',
    optionDefs: [
      optionDef(
        'activities',
        label: 'Actividades en plataformas',
        type: RegimeOptionDefType.multiEnum,
        allowed: ['transport_delivery', 'lodging', 'goods_services'],
      ),
      optionDef(
        'definitive_option',
        label: 'Pago definitivo',
        type: RegimeOptionDefType.boolean,
        defaultValue: false,
      ),
      optionDef(
        'rfc_given_to_platform',
        label: 'Di mi RFC a la plataforma',
        type: RegimeOptionDefType.boolean,
        defaultValue: true,
      ),
    ],
  ),
  regime(
    '626',
    name: 'RESICO',
    incompatibleWith: [
      const RegimeIncompatibility(
        code: '612',
        reason: 'El RESICO sustituye a este régimen.',
      ),
    ],
  ),
  regime(
    '621',
    name: 'Incorporación Fiscal',
    treatment: Treatment.partial,
    explanation: 'Régimen en extinción.',
    optionDefs: [
      optionDef(
        'rif_year',
        label: 'Año en el RIF',
        type: RegimeOptionDefType.valueEnum,
        allowed: ['1', '2', '3'],
      ),
    ],
  ),
  regime(
    '605',
    name: 'Sueldos y salarios',
    treatment: Treatment.informational,
    explanation: 'Tu patrón calcula y retiene tu ISR.',
  ),
];

/// The persona moral catalog of the onboarding tests.
final List<Regime> moralCatalog = [
  regime(
    '601',
    name: 'General de Ley',
    personType: PersonType.moral,
    optionDefs: [
      optionDef(
        'iva_fully_taxable',
        label: 'Todo lo que cobra causa IVA',
        type: RegimeOptionDefType.boolean,
        defaultValue: true,
      ),
    ],
    requiredInputs: [
      inputDef(
        'profit_coefficient',
        label: 'Coeficiente de utilidad',
        kind: TaxInputDefKind.rate,
      ),
    ],
  ),
  regime('626', name: 'RESICO', personType: PersonType.moral),
];

/// A created taxpayer as the server answers it.
Taxpayer taxpayer({
  String id = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e',
  PersonType personType = PersonType.fisica,
  String displayName = 'Ana',
  String rfcMasked = 'GOMA******AB6',
  List<Warning>? warnings,
}) => Taxpayer(
  id: id,
  personType: personType,
  displayName: displayName,
  rfcMasked: rfcMasked,
  regimes: const <TaxpayerRegime>[],
  archivedAt: null,
  createdAt: DateTime.utc(2026, 10, 13),
  warnings: warnings,
);

/// A validation detail as the server sends it.
api.FieldError detail(String field, String code) =>
    api.FieldError(field: field, code: code, message: 'rejected');

/// One recorded tax-inputs save.
typedef TaxInputsSave = ({
  String taxpayerId,
  String regimeCode,
  int fiscalYear,
  List<TaxInputValue> values,
});

/// A [TaxpayersRepository] the test scripts: each operation answers with
/// its scripted value or throws its scripted failure, and every call is
/// recorded.
class FakeTaxpayersRepository implements TaxpayersRepository {
  /// Catalogs by person type.
  final Map<PersonType, List<Regime>> catalogs = {
    PersonType.fisica: fisicaCatalog,
    PersonType.moral: moralCatalog,
  };

  /// Thrown by the next catalog load, then cleared.
  Exception? catalogFailure;

  /// Answers of `createTaxpayer`, in order: a [Taxpayer] or a failure.
  final List<Object> createAnswers = [];

  /// Answers of `correctRfc`, in order.
  final List<Object> correctAnswers = [];

  /// Failure of every tax-inputs save, if any.
  Exception? taxInputsFailure;

  /// Held open while set: a create stays in flight until completed.
  Completer<void>? createGate;

  final List<PersonType> catalogCalls = [];
  final List<TaxpayerCreate> creates = [];
  final List<({String taxpayerId, String rfc})> corrections = [];
  final List<TaxInputsSave> taxInputSaves = [];

  @override
  Future<List<Regime>> listRegimes(PersonType personType) async {
    catalogCalls.add(personType);
    final failure = catalogFailure;
    if (failure != null) {
      catalogFailure = null;
      throw failure;
    }
    return catalogs[personType]!;
  }

  @override
  Future<Taxpayer> createTaxpayer(TaxpayerCreate request) async {
    creates.add(request);
    await createGate?.future;
    return _answer(createAnswers, taxpayer);
  }

  @override
  Future<Taxpayer> correctRfc({
    required String taxpayerId,
    required String rfc,
  }) async {
    corrections.add((taxpayerId: taxpayerId, rfc: rfc));
    return _answer(correctAnswers, taxpayer);
  }

  @override
  Future<void> saveTaxInputs({
    required String taxpayerId,
    required String regimeCode,
    required int fiscalYear,
    required List<TaxInputValue> values,
  }) async {
    taxInputSaves.add((
      taxpayerId: taxpayerId,
      regimeCode: regimeCode,
      fiscalYear: fiscalYear,
      values: values,
    ));
    final failure = taxInputsFailure;
    if (failure != null) throw failure;
  }

  Taxpayer _answer(List<Object> answers, Taxpayer Function() fallback) {
    if (answers.isEmpty) return fallback();
    final answer = answers.removeAt(0);
    if (answer is Taxpayer) return answer;
    throw answer as Exception;
  }
}
