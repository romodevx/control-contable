import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_create.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime_create.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:control_contable/features/taxpayers/application/regime_wizard.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart'
    as periods;
import 'package:control_contable/features/taxpayers/domain/regime_compatibility.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:flutter/foundation.dart';

export 'package:control_contable/features/taxpayers/application/onboarding_failures.dart'
    show OnboardingNotice, RfcRejection;

/// `Warning.code` of a check-digit mismatch (D-RFC).
const WarningCode rfcCheckDigitWarning = WarningCode.rfcCheckDigitMismatch;

/// `Warning.code` of a regime pair still under CPA review (§2.5).
const WarningCode cpaReviewWarning =
    WarningCode.regimeCombinationNeedsCpaReview;

/// A regime pair the server allowed under CPA review (§2.5): the regimes
/// sent, in request order (empty when the warning names none), and the
/// server's reason.
typedef CpaReview = ({List<Regime> regimes, String reason});

/// `Warning.field` of a CPA-review warning on create: the index of the
/// later regime of the pair in the request.
final RegExp _sentRegimeField = RegExp(r'^regimes\.(\d+)\.regime_code$');

/// Mexico City's offset from UTC (no daylight saving time since 2022).
const Duration _mexicoCityOffset = Duration(hours: -6);

/// Where the onboarding is.
enum OnboardingStep {
  /// 1 · Persona física or moral (OD-3: nothing is asked before it).
  personType,

  /// 2 · RFC and display name.
  identity,

  /// 3 · Regimes of the chosen person type.
  regimes,

  /// 4 · Options of each chosen regime that has some, one page each.
  options,

  /// After create: non-blocking warnings to confirm.
  review,

  /// 5 · Optional tax inputs (`module_tax_engine` on), one page each.
  taxInputs,

  /// A read-only user (`402`): the paywall message, nothing created.
  readOnly,

  /// The user already holds the cap of taxpayers (OD-2): said before
  /// step 1, nothing to fill.
  limitReached,

  /// Finished: the taxpayer exists.
  done,
}

/// The onboarding of a new taxpayer (MODULE-taxpayers US-T1, US-T1b): its
/// draft, the step it is on, and what the server said. Screens read it and
/// call its actions; it never holds UI state.
///
/// One `POST /v1/taxpayers` creates the taxpayer with every regime and
/// option at the end of step 4 (or step 3 when no chosen regime has
/// options). Going back keeps every value. A server verdict sends the user
/// back to the step that owns the rejected field.
class OnboardingFlow extends ChangeNotifier implements RegimeWizard {
  /// A flow writing through `repository`. [newId] makes the client id of
  /// the create (stable across retries); `clock` dates the fiscal year.
  /// [knownReadOnly] (the server already answered `402` this session)
  /// starts at the paywall: nothing to fill that could not be saved.
  /// [capReached] (the user's list already holds the cap) starts at the
  /// limit, after the paywall as the server checks access first (§2.10).
  new({
    required this._repository,
    required String Function() newId,
    required this._clock,
    required this.taxEngineEnabled,
    bool knownReadOnly = false,
    bool capReached = false,
  }) : _clientId = newId(),
       _step = switch ((knownReadOnly, capReached)) {
         (true, _) => OnboardingStep.readOnly,
         (_, true) => OnboardingStep.limitReached,
         _ => OnboardingStep.personType,
       };

  final TaxpayersRepository _repository;
  final Clock _clock;
  final String _clientId;

  /// Whether `module_tax_engine` is on (adds the optional step 5). The
  /// screen keeps it in step with the flag; it notifies nobody.
  bool taxEngineEnabled;

  OnboardingStep _step;
  int _page = 0;
  PersonType? _personType;
  String _displayName = '';
  String _rfc = '';
  List<Regime> _regimes = const [];
  late final String _currentPeriod = periods.currentPeriod(_clock());
  late String _startPeriod = _currentPeriod;
  final Map<String, Map<String, Object>> _options = {};
  Taxpayer? _created;
  List<Warning> _warnings = const [];
  RfcRejection? _rfcRejection;
  bool _displayNameRejected = false;
  Map<String, Set<String>> _optionRejections = const {};
  OnboardingNotice? _notice;

  /// The current step.
  OnboardingStep get step => _step;

  @override
  RegimeWizardPurpose get purpose => RegimeWizardPurpose.createTaxpayer;

  /// The page within [OnboardingStep.options] or [OnboardingStep.taxInputs].
  @override
  int get page => _page;

  /// The chosen person type.
  @override
  PersonType? get personType => _personType;

  /// A persona moral holds one regime at a time (UX; the server's `422`
  /// decides).
  @override
  bool get singleRegime => _personType == PersonType.moral;

  /// A new taxpayer holds nothing yet.
  @override
  Set<String> get heldCodes => const {};

  /// The display name as typed.
  String get displayName => _displayName;

  /// The RFC as typed (normalized only when sent).
  String get rfc => _rfc;

  /// The chosen regimes, in selection order.
  @override
  List<Regime> get regimes => _regimes;

  /// The created taxpayer, once it exists.
  Taxpayer? get created => _created;

  /// Why the server refused the RFC, until the user submits it again.
  RfcRejection? get rfcRejection => _rfcRejection;

  /// Whether the server refused the display name.
  bool get displayNameRejected => _displayNameRejected;

  /// Options the server refused, by regime code.
  @override
  Map<String, Set<String>> get optionRejections => _optionRejections;

  /// A form-level message for the current screen.
  @override
  OnboardingNotice? get notice => _notice;

  /// Chosen regimes that have options (one page each in step 4).
  @override
  List<Regime> get optionRegimes => [
    for (final regime in _regimes)
      if (regime.optionDefs.isNotEmpty) regime,
  ];

  /// Chosen regimes with tax inputs to offer (step 5), only while the tax
  /// engine is on.
  List<Regime> get inputRegimes => [
    if (taxEngineEnabled)
      for (final regime in _regimes)
        if (regime.requiredInputs.isNotEmpty) regime,
  ];

  /// How many numbered steps the onboarding has.
  @override
  int get totalSteps => taxEngineEnabled ? 5 : 4;

  /// The current step's number, or `null` outside the numbered steps.
  @override
  int? get stepNumber => switch (_step) {
    OnboardingStep.personType => 1,
    OnboardingStep.identity => 2,
    OnboardingStep.regimes => 3,
    OnboardingStep.options => 4,
    OnboardingStep.taxInputs => 5,
    _ => null,
  };

  /// Whether the server warned that the RFC check digit does not match.
  bool get rfcWarning => _warnings.any((w) => w.code == rfcCheckDigitWarning);

  /// Every regime pair under CPA review the server warned about. The
  /// warning names the later regime sent; its partners are the earlier ones
  /// the catalog pairs it with (`review_pending_with`).
  List<CpaReview> get cpaReviews => [
    for (final warning in _warnings)
      if (warning.code == cpaReviewWarning)
        (regimes: _pairNamedBy(warning.field), reason: warning.message),
  ];

  List<Regime> _pairNamedBy(String field) {
    final index = int.tryParse(
      _sentRegimeField.firstMatch(field)?.group(1) ?? '',
    );
    if (index == null || index >= _regimes.length) return const [];
    final later = _regimes[index];
    final partners = reviewPendingPartners(later, _regimes.take(index));
    return partners.isEmpty ? const [] : [...partners, later];
  }

  /// The fiscal year the tax inputs belong to (America/Mexico_City).
  int get fiscalYear => _clock().toUtc().add(_mexicoCityOffset).year;

  /// Step 1: [type] chosen. Another type than before drops the regimes,
  /// which belong to the other catalog.
  void choosePersonType(PersonType type) {
    if (type != _personType) {
      _regimes = const [];
      _options.clear();
    }
    _personType = type;
    _go(OnboardingStep.identity);
  }

  /// Keeps the step 2 values as typed (no rebuild): going back keeps them.
  void rememberIdentity({required String displayName, required String rfc}) {
    _displayName = displayName;
    _rfc = rfc;
  }

  /// Step 2 done (the client checks passed).
  void submitIdentity({required String displayName, required String rfc}) {
    rememberIdentity(displayName: displayName, rfc: rfc);
    _rfcRejection = null;
    _displayNameRejected = false;
    _go(OnboardingStep.regimes);
  }

  @override
  String get startPeriod => _startPeriod;

  @override
  List<String> get startPeriodChoices =>
      periods.startPeriodChoices(_currentPeriod);

  @override
  void chooseStartPeriod(String period) => _startPeriod = period;

  /// Keeps the step 3 choice as picked (no rebuild): going back keeps it.
  @override
  void rememberRegimes(List<Regime> chosen) =>
      _regimes = List.unmodifiable(chosen);

  /// Step 3 done: options next, or the create when none has options.
  @override
  Future<void> submitRegimes(List<Regime> chosen) async {
    rememberRegimes(chosen);
    if (optionRegimes.isNotEmpty) return _go(OnboardingStep.options);
    await _create();
  }

  /// The options of [regime]: what the user set, else the catalog
  /// defaults (an option without a default starts empty).
  @override
  Map<String, Object> optionsFor(Regime regime) =>
      _options[regime.code] ?? catalogDefaults(regime);

  /// Keeps the options of [regimeCode] as set (no rebuild).
  @override
  void rememberOptions(String regimeCode, Map<String, Object> values) =>
      _options[regimeCode] = Map.unmodifiable(values);

  /// One options page done: the next one, or the create after the last.
  @override
  Future<void> submitOptions(
    String regimeCode,
    Map<String, Object> values,
  ) async {
    rememberOptions(regimeCode, values);
    _optionRejections = {..._optionRejections}..remove(regimeCode);
    if (_page < optionRegimes.length - 1) {
      return _go(OnboardingStep.options, page: _page + 1);
    }
    await _create();
  }

  /// Leaves the review: [rfc] is the RFC field's value; when it differs
  /// from what was sent, it is corrected first (`PATCH`, OQ-1).
  Future<void> confirmReview({required String rfc}) async {
    final corrected = normalizeRfc(rfc);
    if (corrected == normalizeRfc(_rfc)) return _afterReview();
    final taxpayerId = _created!.id;
    try {
      final updated = await _repository.correctRfc(
        taxpayerId: taxpayerId,
        rfc: corrected,
      );
      _rfc = corrected;
      _created = updated;
      _warnings = updated.warnings ?? const [];
      _rfcRejection = null;
      if (rfcWarning) return _go(OnboardingStep.review);
      _afterReview();
    } on ApiFailure catch (failure) {
      _rfcRejection = rfcRejectionOf(failure);
      _notice = _rfcRejection == null ? noticeFor(failure) : null;
      notifyListeners();
    }
  }

  /// Saves the current step 5 page; nothing typed is a skip.
  Future<void> saveTaxInputs(List<TaxInputValue> values) async {
    if (values.isNotEmpty) {
      try {
        await _repository.saveTaxInputs(
          taxpayerId: _created!.id,
          regimeCode: inputRegimes[_page].code,
          fiscalYear: fiscalYear,
          values: values,
        );
      } on ApiFailure catch (failure) {
        _notice = noticeFor(failure);
        notifyListeners();
        return;
      }
    }
    if (_page < inputRegimes.length - 1) {
      return _go(OnboardingStep.taxInputs, page: _page + 1);
    }
    _go(OnboardingStep.done);
  }

  /// One step back, keeping every value. `false` when there is no step to
  /// go back to (the first step, or the taxpayer already exists): the
  /// screen closes then.
  bool back() {
    final previous = switch (_step) {
      OnboardingStep.identity => OnboardingStep.personType,
      OnboardingStep.regimes => OnboardingStep.identity,
      OnboardingStep.options when _page > 0 => OnboardingStep.options,
      OnboardingStep.options => OnboardingStep.regimes,
      _ => null,
    };
    if (previous == null) return false;
    _go(previous, page: previous == OnboardingStep.options ? _page - 1 : 0);
    return true;
  }

  Future<void> _create() async {
    _notice = null;
    try {
      _created = await _repository.createTaxpayer(_request());
    } on ApiFailure catch (failure) {
      return _showCreateFailure(failure);
    }
    _warnings = _created!.warnings ?? const [];
    if (_warnings.isNotEmpty) return _go(OnboardingStep.review);
    _afterReview();
  }

  TaxpayerCreate _request() => TaxpayerCreate(
    id: _clientId,
    personType: _personType!,
    rfc: normalizeRfc(_rfc),
    displayName: _displayName.trim(),
    regimes: [
      for (final regime in _regimes)
        TaxpayerRegimeCreate(
          regimeCode: regime.code,
          options: regime.optionDefs.isEmpty ? null : optionsFor(regime),
          startPeriod: periods.startPeriodToSend(
            chosen: _startPeriod,
            current: _currentPeriod,
          ),
        ),
    ],
  );

  void _afterReview() => inputRegimes.isEmpty
      ? _go(OnboardingStep.done)
      : _go(OnboardingStep.taxInputs);

  void _showCreateFailure(ApiFailure failure) {
    switch (failure) {
      case ApiEntitlementRequired():
        _go(OnboardingStep.readOnly);
      case ApiRejected(code: ErrorCode.rfcAlreadyRegistered):
        _rfcRejection = RfcRejection.duplicate;
        _go(OnboardingStep.identity);
      case ApiRejected(code: ErrorCode.invalidRegimeCombination):
        _go(OnboardingStep.regimes);
        _notice = OnboardingNotice.combinationRejected;
      case ApiValidationFailure(:final details) when details.isNotEmpty:
        _routeDetails(details);
      default:
        _notice = noticeFor(failure);
    }
    notifyListeners();
  }

  /// Sends the user to the earliest step owning a rejected field.
  void _routeDetails(List<api.FieldError> details) {
    final verdict = sortDetails(details, sentRegimes: _regimes);
    _rfcRejection = verdict.rfc;
    _displayNameRejected = verdict.displayName;
    _optionRejections = verdict.options;
    if (verdict.rfc != null || verdict.displayName) {
      return _go(OnboardingStep.identity);
    }
    if (verdict.regime || verdict.startPeriod) {
      _go(OnboardingStep.regimes);
      _notice = verdict.regime
          ? OnboardingNotice.regimeRejected
          : OnboardingNotice.startPeriodRejected;
      return;
    }
    final firstRejected = optionRegimes.indexWhere(
      (regime) => verdict.options.containsKey(regime.code),
    );
    if (firstRejected >= 0) {
      return _go(OnboardingStep.options, page: firstRejected);
    }
    _notice = OnboardingNotice.unexpected;
  }

  void _go(OnboardingStep step, {int page = 0}) {
    _step = step;
    _page = page;
    _notice = null;
    notifyListeners();
  }
}

/// The catalog defaults of [regime]'s options (an option without a default
/// is left out).
Map<String, Object> catalogDefaults(Regime regime) => {
  for (final def in regime.optionDefs)
    if (def.defaultValue case final Object value) def.key: value,
};
