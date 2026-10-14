import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:flutter/foundation.dart';

/// What a regime wizard ends in.
enum RegimeWizardPurpose {
  /// The onboarding's create (`POST /v1/taxpayers`).
  createTaxpayer,

  /// One more regime for an existing taxpayer (`POST …/regimes`).
  addRegime,
}

/// What the regimes step (3) and the options step (4) need from the flow
/// they belong to: the onboarding, or adding a regime from settings. The
/// steps read it and call its actions; it never holds UI state.
abstract interface class RegimeWizard implements Listenable {
  /// What the wizard ends in (it names the last primary button).
  RegimeWizardPurpose get purpose;

  /// The catalog's person type.
  PersonType? get personType;

  /// Whether only one regime can be chosen (a persona moral, or adding).
  bool get singleRegime;

  /// Codes the taxpayer already holds open: not offered again, and checked
  /// for conflicts with the choice.
  Set<String> get heldCodes;

  /// The chosen regimes, in selection order.
  List<Regime> get regimes;

  /// Keeps the step 3 choice as picked (no rebuild).
  void rememberRegimes(List<Regime> chosen);

  /// Step 3 done: options next, or the write when none has options.
  Future<void> submitRegimes(List<Regime> chosen);

  /// The page within the options step.
  int get page;

  /// Chosen regimes that have options (one page each).
  List<Regime> get optionRegimes;

  /// The options of [regime]: what the user set, else the defaults.
  Map<String, Object> optionsFor(Regime regime);

  /// Keeps the options of [regimeCode] as set (no rebuild).
  void rememberOptions(String regimeCode, Map<String, Object> values);

  /// One options page done: the next one, or the write after the last.
  Future<void> submitOptions(String regimeCode, Map<String, Object> values);

  /// Options the server refused, by regime code.
  Map<String, Set<String>> get optionRejections;

  /// A form-level message for the current screen.
  OnboardingNotice? get notice;

  /// The current step's number, or `null` outside the numbered steps.
  int? get stepNumber;

  /// How many numbered steps there are.
  int get totalSteps;
}
