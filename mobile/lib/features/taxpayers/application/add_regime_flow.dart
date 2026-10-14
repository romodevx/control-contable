import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/application/regime_wizard.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:flutter/foundation.dart';

/// Where adding a regime is.
enum AddRegimeStep {
  /// Wizard step 3: pick one regime.
  regimes,

  /// Wizard step 4: its options, when it has some.
  options,

  /// A read-only user (`402`): the paywall message, nothing added.
  readOnly,

  /// Added.
  done,
}

/// Adds one regime to an existing taxpayer from settings (US-T3), reusing
/// the onboarding's regimes and options steps: regimes already held are
/// not offered, a pick that conflicts with a held one is flagged at once,
/// and one `POST …/regimes` writes it starting this month. Re-adding an
/// ended regime reopens it (OQ-5).
class AddRegimeFlow extends ChangeNotifier implements RegimeWizard {
  /// The flow adding to [taxpayerId] (of [personType], holding
  /// [heldCodes] open) through [_settings]. [knownReadOnly] starts at the
  /// paywall.
  new({
    required this._settings,
    required this.taxpayerId,
    required this.personType,
    required this.heldCodes,
    bool knownReadOnly = false,
  }) : _step = knownReadOnly ? AddRegimeStep.readOnly : AddRegimeStep.regimes;

  final TaxpayerSettings _settings;

  /// The taxpayer the regime is added to.
  final String taxpayerId;

  @override
  final PersonType personType;

  @override
  final Set<String> heldCodes;

  AddRegimeStep _step;
  List<Regime> _regimes = const [];
  final Map<String, Map<String, Object>> _options = {};
  Map<String, Set<String>> _optionRejections = const {};
  OnboardingNotice? _notice;
  TaxpayerRegime? _added;

  /// The current step.
  AddRegimeStep get step => _step;

  /// The regime as the server added it (with its warnings), once done.
  TaxpayerRegime? get added => _added;

  @override
  RegimeWizardPurpose get purpose => RegimeWizardPurpose.addRegime;

  @override
  bool get singleRegime => true;

  @override
  List<Regime> get regimes => _regimes;

  @override
  int get page => 0;

  @override
  List<Regime> get optionRegimes => [
    for (final regime in _regimes)
      if (regime.optionDefs.isNotEmpty) regime,
  ];

  @override
  Map<String, Set<String>> get optionRejections => _optionRejections;

  @override
  OnboardingNotice? get notice => _notice;

  @override
  int? get stepNumber => switch (_step) {
    AddRegimeStep.regimes => 1,
    AddRegimeStep.options => 2,
    _ => null,
  };

  @override
  int get totalSteps => 2;

  @override
  void rememberRegimes(List<Regime> chosen) =>
      _regimes = List.unmodifiable(chosen);

  @override
  Future<void> submitRegimes(List<Regime> chosen) async {
    rememberRegimes(chosen);
    if (optionRegimes.isNotEmpty) return _go(AddRegimeStep.options);
    await _add();
  }

  @override
  Map<String, Object> optionsFor(Regime regime) =>
      _options[regime.code] ?? catalogDefaults(regime);

  @override
  void rememberOptions(String regimeCode, Map<String, Object> values) =>
      _options[regimeCode] = Map.unmodifiable(values);

  @override
  Future<void> submitOptions(
    String regimeCode,
    Map<String, Object> values,
  ) async {
    rememberOptions(regimeCode, values);
    await _add();
  }

  /// One step back; `false` from the first step (the screen closes).
  bool back() {
    if (_step != AddRegimeStep.options) return false;
    _go(AddRegimeStep.regimes);
    return true;
  }

  Future<void> _add() async {
    final regime = _regimes.single;
    _notice = null;
    try {
      _added = await _settings.run(
        WriteAccess.needsPlan,
        (repository) => repository.addRegime(
          taxpayerId: taxpayerId,
          regimeCode: regime.code,
          options: regime.optionDefs.isEmpty ? null : optionsFor(regime),
        ),
      );
      _go(AddRegimeStep.done);
    } on ApiFailure catch (failure) {
      _showFailure(failure, regime.code);
    }
  }

  void _showFailure(ApiFailure failure, String code) {
    switch (failure) {
      case ApiEntitlementRequired():
        return _go(AddRegimeStep.readOnly);
      case ApiRejected(code: ErrorCode.invalidRegimeCombination):
        _go(AddRegimeStep.regimes);
        _notice = OnboardingNotice.combinationRejected;
      case ApiValidationFailure(:final details)
          when _rejectedOptions(details).isNotEmpty:
        _optionRejections = {code: _rejectedOptions(details)};
        _go(AddRegimeStep.options);
      case ApiValidationFailure() || ApiRejected(code: ErrorCode.conflict):
        _go(AddRegimeStep.regimes);
        _notice = OnboardingNotice.regimeRejected;
      default:
        _notice = noticeFor(failure);
    }
    notifyListeners();
  }

  /// Option keys named by `options.<key>` details.
  Set<String> _rejectedOptions(List<api.FieldError> details) => {
    for (final detail in details)
      if (detail.field.split('.') case ['options', final key, ...]) key,
  };

  void _go(AddRegimeStep step) {
    _step = step;
    _notice = null;
    notifyListeners();
  }
}
