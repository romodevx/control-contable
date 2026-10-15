import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_last_used.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/yes_no.dart';

/// The 625 regime (platform incomes).
const String platformRegime = '625';

/// The payers that can withhold, so what was sold matters (§2.1 Q2).
const Set<CounterpartyKind> _withholdingPayers = {
  CounterpartyKind.pm,
  CounterpartyKind.govFederal,
  CounterpartyKind.govStateMunicipal,
};

/// The labels of the four questions.
typedef IncomeFactLabels = ({
  String counterpartyKind,
  String operationKind,
  String residentialRent,
  String viaPlatform,
});

/// The facts the withholdings are derived from (MODULE-withholdings §2.1,
/// Q1–Q4), each asked only when its answer can change a withholding, and
/// preselected from what was last used in the regime (`last_used`) until
/// the user answers. The figures come from the server (ADR-0008).
class IncomeWithholdingFacts {
  /// The four questions added to [form] for a [isMoral] taxpayer, starting
  /// at [stored] when editing, else empty until the server's `last_used`
  /// for [regimeCode] arrives ([applyLastUsed]).
  new({
    required AppFormController form,
    required IncomeFactLabels labels,
    required this.isMoral,
    required this.regimeCode,
    Income? stored,
  }) {
    final regime = regimeCode();
    counterpartyKind = form.add(
      ChoiceFieldController(
        label: labels.counterpartyKind,
        keepForNext: true,
        initialValue: _storedKind(stored)?.json,
      ),
    );
    operationKind = form.add(
      ChoiceFieldController(
        label: labels.operationKind,
        isRequired: false,
        keepForNext: true,
        // The regime default is the server's (`last_used`), never a copy
        // here: until it arrives (or if it cannot), Q2 is unanswered and
        // omitted, so the server applies the same default.
        initialValue: stored?.operationKind?.json,
      ),
    );
    residentialRent = form.add(
      ChoiceFieldController(
        label: labels.residentialRent,
        isRequired: false,
        keepForNext: true,
        initialValue: (stored?.residentialRent ?? false).choice,
      ),
    );
    viaPlatform = form.add(
      ChoiceFieldController(
        label: labels.viaPlatform,
        isRequired: false,
        keepForNext: true,
        initialValue:
            (stored?.viaPlatform ??
                    (_wasPlatform(stored) ||
                        (stored == null && regime == platformRegime)))
                .choice,
      ),
    );
    for (final field in _all) {
      _suggested[field] = field.value;
    }
  }

  /// Whether the taxpayer is a persona moral (Q4 then lives under "Más
  /// detalles").
  final bool isMoral;

  /// The record's regime, as the form currently says.
  final String Function() regimeCode;

  final Map<ChoiceFieldController, String?> _suggested = {};

  /// Q1 "¿A quién le facturaste?" (`CounterpartyKind.json`).
  late final ChoiceFieldController counterpartyKind;

  /// Q2 "¿Qué le vendiste…?" (`OperationKind.json`).
  late final ChoiceFieldController operationKind;

  /// Q3 "¿Es una casa o departamento para vivir?" ([yes] / [no]).
  late final ChoiceFieldController residentialRent;

  /// Q4 "¿Lo cobraste a través de una plataforma digital…?" ([yes] / [no]).
  late final ChoiceFieldController viaPlatform;

  List<ChoiceFieldController> get _all => [
    counterpartyKind,
    operationKind,
    residentialRent,
    viaPlatform,
  ];

  /// Whether any answer differs from what the form started with.
  bool get isDirty => _all.any((field) => field.isDirty);

  /// The questions, for the server verdicts by path.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    'counterparty_kind': counterpartyKind,
    'operation_kind': operationKind,
    'residential_rent': residentialRent,
    'via_platform': viaPlatform,
  };

  static bool _wasPlatform(Income? stored) =>
      // ignore: deprecated_member_use_from_same_package, legacy rows until WH-08 backfills them.
      stored?.counterpartyType == CounterpartyType.plataforma;

  static CounterpartyKind? _storedKind(Income? stored) {
    if (stored == null) return null;
    if (stored.counterpartyKind case final kind?) return kind;
    // ignore: deprecated_member_use_from_same_package, legacy rows until WH-08 backfills them.
    return switch (stored.counterpartyType) {
      CounterpartyType.personaFisica => CounterpartyKind.pfWithActivity,
      CounterpartyType.publicoGeneral => CounterpartyKind.pfNoActivityOrPublic,
      CounterpartyType.personaMoral ||
      CounterpartyType.plataforma => CounterpartyKind.pm,
      CounterpartyType.$unknown => null,
    };
  }

  /// Who was billed, once answered.
  CounterpartyKind? get kind {
    final value = counterpartyKind.value;
    return value == null ? null : CounterpartyKind.fromJson(value);
  }

  /// Whether Q4 is asked at all: 625, or any persona moral regime.
  bool get asksViaPlatform => regimeCode() == platformRegime || isMoral;

  /// Whether Q4 is asked under "Más detalles" (a persona moral) rather
  /// than with the other questions (625).
  bool get viaPlatformIsSecondary => isMoral;

  /// Whether the income was collected through a platform.
  bool get isViaPlatform => asksViaPlatform && viaPlatform.value == yes;

  /// Whether Q2 is asked: the payer can withhold and no platform did.
  bool get asksOperation => _withholdingPayers.contains(kind) && !isViaPlatform;

  /// What was sold, stored even when not asked (its default); `null` via a
  /// platform (the server derives `platform`).
  OperationKind? get operation {
    final value = operationKind.value;
    if (isViaPlatform || value == null) return null;
    return OperationKind.fromJson(value);
  }

  /// Whether Q3 is asked: the rent of a property.
  bool get asksResidentialRent => operation == OperationKind.rentRealEstate;

  /// Whether the income is housing rent (exempt, LIVA 20 fr. II).
  bool get isResidentialRent =>
      asksResidentialRent && residentialRent.value == yes;

  /// `residential_rent` as sent: only where it was asked.
  bool? get residentialRentSent =>
      asksResidentialRent ? isResidentialRent : null;

  /// `via_platform` as sent: only where it may be (625 and persona moral).
  bool? get viaPlatformSent => asksViaPlatform ? isViaPlatform : null;

  /// The deprecated `counterparty_type` 0.10.0 still requires (WH-06 drops
  /// it), mapped from the new facts.
  CounterpartyType get legacyType {
    if (isViaPlatform) return CounterpartyType.plataforma;
    return switch (kind) {
      CounterpartyKind.pm ||
      CounterpartyKind.govFederal ||
      CounterpartyKind.govStateMunicipal => CounterpartyType.personaMoral,
      CounterpartyKind.pfWithActivity => CounterpartyType.personaFisica,
      _ => CounterpartyType.publicoGeneral,
    };
  }

  /// The server's `last_used` for the regime: each question the user has
  /// not answered moves to it.
  void applyLastUsed(IncomeLastUsed lastUsed) {
    _suggest(counterpartyKind, lastUsed.counterpartyKind?.json);
    _suggest(operationKind, lastUsed.operationKind.json);
    _suggest(residentialRent, lastUsed.residentialRent.choice);
    _suggest(viaPlatform, lastUsed.viaPlatform.choice);
  }

  /// Moves [field] to [value] unless the user answered it; a suggestion
  /// is not an unsaved change.
  void _suggest(ChoiceFieldController field, String? value) {
    if (field.value != _suggested[field]) return;
    final wasPristine = !field.isDirty;
    _suggested[field] = value;
    field.value = value;
    if (wasPristine) field.markPristine();
  }
}
