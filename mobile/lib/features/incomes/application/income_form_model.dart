import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/fiscal/housing_rent_rate.dart';
import 'package:control_contable/core/fiscal/iva_math.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/incomes/application/income_suggestions.dart';
import 'package:control_contable/features/incomes/application/income_withholding_facts.dart';
import 'package:control_contable/features/incomes/domain/income_basis.dart';
import 'package:control_contable/features/incomes/domain/income_math.dart';
import 'package:control_contable/features/incomes/domain/income_preview_query.dart';
import 'package:flutter/foundation.dart';

export 'package:control_contable/features/incomes/application/income_withholding_facts.dart'
    show platformRegime;

/// The regime option that lists a 625 taxpayer's activities.
const String _activitiesOption = 'activities';

/// The 625 option saying whether the RFC was given to the platform.
const String _rfcGivenOption = 'rfc_given_to_platform';

/// A folio fiscal: 8-4-4-4-12 hexadecimal digits (contract `CfdiUuid`).
final RegExp _cfdiUuid = RegExp(
  r'^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$',
);

/// The visible labels and copy the model needs (read from the ARB files
/// by the screen, so this layer has no `BuildContext`).
typedef IncomeFormCopy = ({
  String collectedOn,
  String invoicedOn,
  String description,
  String amount,
  String ivaRate,
  IncomeFactLabels facts,
  String regime,
  String activity,
  String ivaAmount,
  String cfdiUuid,
  String cfdiUuidInvalid,
});

/// What the income form page holds (M3-09, WH-11): its fields in reading
/// order, the withholding facts (the server derives the amounts,
/// ADR-0008), the defaults' suggestions, and the idempotency id of the
/// current form session (MODULE-incomes §2.3).
///
/// New income: context fields (dates, regime, the withholding facts, IVA
/// rate, activity) survive "Guardar y crear otro"; the rest is cleared and
/// a new session id starts. Edit ([editing]): every field starts at the
/// stored income, and the stored amounts are the user's (the CFDI's): a
/// defaults suggestion never replaces them.
class IncomeFormModel extends ChangeNotifier {
  /// The form for [regimes] (those in effect on the income's date) of a
  /// [isMoral] taxpayer, starting on [today], or on [editing] when it edits
  /// a stored income.
  new({
    required List<TaxpayerRegime> regimes,
    required DateTime today,
    required this.isMoral,
    required this.copy,
    required String Function() newId,
    this.editing,
  }) : _regimes = regimes,
       _newId = newId,
       basis = incomeBasisOf(editing?.regimeCode ?? regimes.first.regimeCode),
       _sessionId = newId() {
    _build(today);
    _suggestions = IncomeSuggestions(amount: amount, ivaAmount: ivaAmount);
    if (editing case final stored?) {
      _suggestions.adoptStored(
        ivaFollowsAmount:
            stored.ivaCents == ivaCentsFor(stored.amountCents, stored.ivaRate),
      );
    }
    for (final field in [amount, ivaAmount]) {
      field.text.addListener(_recompute);
    }
    ivaRate.addListener(_recompute);
    form.addListener(_followFacts);
  }

  /// The copy the fields were labeled with.
  final IncomeFormCopy copy;

  /// The stored income this form edits; `null` for a new income.
  final Income? editing;

  /// Whether the taxpayer is a persona moral.
  final bool isMoral;

  /// Whether this form edits a stored income.
  bool get isEditing => editing != null;

  /// Which date books the income (from the first regime).
  final IncomeBasis basis;

  final List<TaxpayerRegime> _regimes;
  final String Function() _newId;
  String _sessionId;
  IncomeDefaults? _defaults;
  late final IncomeSuggestions _suggestions;
  late final HousingRentRate _housing = HousingRentRate(
    ivaRate,
    () => _suggestions.suggestedRate ?? IvaRate.rate16,
  );

  /// The fields, in reading order.
  final AppFormController form = AppFormController();

  /// The booking date (collection, or invoice for an accrual regime).
  late final DateFieldController bookingDate;

  /// Subtotal before IVA.
  late final MoneyFieldController amount;

  /// The IVA rate (`IvaRate.json`); hidden (exempt) for housing rent.
  late final ChoiceFieldController ivaRate;

  /// Q1–Q4: what the withholdings are derived from.
  late final IncomeWithholdingFacts facts;

  /// The regime; shown only with more than one in effect.
  late final ChoiceFieldController regime;

  /// The 625 activity; shown only when 625 holds several.
  late final ChoiceFieldController activity;

  /// What the income was ("Más detalles").
  late final AppFieldController description;

  /// IVA charged; prefilled from the rate (the rate's IVA when empty).
  late final MoneyFieldController ivaAmount;

  /// The date that does not book the income.
  late final DateFieldController otherDate;

  /// The CFDI's folio fiscal.
  late final AppFieldController cfdiUuid;

  void _build(DateTime today) {
    final accrual = basis == IncomeBasis.accrual;
    final stored = editing;
    String day(String? date) =>
        date == null ? '' : formatDate(parseBusinessDate(date));
    String money(int? cents) => cents == null ? '' : formatCentavos(cents);
    bookingDate = form.dateField(
      label: accrual ? copy.invoicedOn : copy.collectedOn,
      keepForNext: true,
      initialText: stored == null
          ? formatDate(today)
          : day(accrual ? stored.invoicedOn : stored.collectedOn),
    );
    amount = form.moneyField(
      label: copy.amount,
      initialText: money(stored?.amountCents),
    );
    ivaRate = form.add(
      ChoiceFieldController(
        label: copy.ivaRate,
        keepForNext: true,
        initialValue: (stored?.ivaRate ?? IvaRate.rate16).json,
      ),
    );
    // Built before the facts (they read it), added after them (reading order).
    regime = ChoiceFieldController(
      label: copy.regime,
      keepForNext: true,
      initialValue: stored?.regimeCode ?? _regimes.first.regimeCode,
    );
    facts = IncomeWithholdingFacts(
      form: form,
      labels: copy.facts,
      isMoral: isMoral,
      regimeCode: () => regimeCode,
      stored: stored,
    );
    form.add(regime);
    activity = form.add(
      ChoiceFieldController(
        label: copy.activity,
        isRequired: false,
        keepForNext: true,
        initialValue: stored?.platformActivity?.json,
      ),
    );
    description = form.textField(
      label: copy.description,
      isRequired: false,
      initialText: stored?.description ?? '',
    );
    // Prefilled from the rate; left empty, the rate's IVA is sent.
    ivaAmount = form.moneyField(
      label: copy.ivaAmount,
      isRequired: false,
      initialText: money(stored?.ivaCents),
    );
    otherDate = form.dateField(
      label: accrual ? copy.collectedOn : copy.invoicedOn,
      isRequired: false,
      keepForNext: true,
      initialText: stored == null
          ? ''
          : day(accrual ? stored.collectedOn : stored.invoicedOn),
    );
    cfdiUuid = form.textField(
      label: copy.cfdiUuid,
      isRequired: false,
      validator: (value) => _cfdiUuid.hasMatch(value)
          ? null
          : CustomFieldError(copy.cfdiUuidInvalid),
      canonicalize: (value) => value.toUpperCase(),
      initialText: stored?.cfdiUuid ?? '',
    );
  }

  /// The idempotency id of this form session: every retry resends it.
  String get sessionId => _sessionId;

  /// The regimes the user can choose (more than one shows the field).
  List<TaxpayerRegime> get regimes => List.unmodifiable(_regimes);

  /// The chosen regime's code.
  String get regimeCode => regime.value ?? _regimes.first.regimeCode;

  TaxpayerRegime? get _heldPlatformRegime =>
      _regimes.where((r) => r.regimeCode == platformRegime).firstOrNull;

  /// The 625 activities held, when the chosen regime is 625.
  List<PlatformActivity> get activities {
    if (regimeCode != platformRegime) return const [];
    final held = _heldPlatformRegime?.options[_activitiesOption];
    return [
      if (held is List)
        for (final key in held) PlatformActivity.fromJson('$key'),
    ];
  }

  /// The activity the income is for: the chosen one, or the only one.
  PlatformActivity? get platformActivity {
    final chosen = activity.value;
    if (chosen != null) return PlatformActivity.fromJson(chosen);
    final held = activities;
    return held.length == 1 ? held.single : null;
  }

  /// Whether a 625 income through a platform comes from a taxpayer who did
  /// not give the platform their RFC (20 % ISR and all the IVA withheld).
  bool get platformWithoutRfc =>
      facts.isViaPlatform &&
      regimeCode == platformRegime &&
      _heldPlatformRegime?.options[_rfcGivenOption] == false;

  /// The booking date as the contract's `YYYY-MM-DD`, when valid.
  String? get bookingDay {
    final date = bookingDate.date;
    return date == null ? null : businessDate(date);
  }

  /// The withholding date: collection, else invoice (the payer withholds
  /// when it pays, MODULE-withholdings §4.1).
  String? get withholdingDay {
    final collected = basis == IncomeBasis.accrual ? otherDate : bookingDate;
    final date = collected.date ?? bookingDate.date;
    return date == null ? null : businessDate(date);
  }

  /// The last defaults answer, or `null` before one (or after a failure).
  IncomeDefaults? get defaults => _defaults;

  /// The IVA rates to offer: the allowed ones once known.
  List<IvaRate> get ivaRatesOffered =>
      _defaults?.ivaRatesAllowed ?? IvaRate.$valuesDefined;

  /// Whether the IVA rate is asked (housing rent is exempt by law).
  bool get asksIvaRate => !facts.isResidentialRent;

  /// The IVA the user wrote, else the one the rate gives.
  int get _ivaCents =>
      ivaAmount.centavos ??
      ivaCentsFor(amount.centavos ?? 0, IvaRate.fromJson(ivaRate.value ?? ''));

  /// The live amounts (the total under the amount).
  IncomeAmounts get amounts =>
      IncomeAmounts(amountCents: amount.centavos ?? 0, ivaCents: _ivaCents);

  /// The facts the server derives the withholding from, or `null` until
  /// the amount, the rate and who was billed are known.
  IncomePreviewQuery? get previewQuery {
    final cents = amount.centavos;
    final kind = facts.kind;
    final rate = IvaRate.fromJson(ivaRate.value ?? '');
    if (cents == null || cents <= 0 || kind == null) return null;
    if (rate == IvaRate.$unknown) return null;
    return (
      regime: regimeCode,
      on: withholdingDay,
      amountCents: cents,
      ivaRate: rate,
      ivaCents: _ivaCents,
      counterpartyKind: kind,
      operationKind: facts.operation,
      residentialRent: facts.residentialRentSent,
      viaPlatform: facts.viaPlatformSent,
      platformActivity: regimeCode == platformRegime ? platformActivity : null,
    );
  }

  /// Applies the defaults for the current regime and date: the suggested
  /// IVA rate (unless the user chose one), the allowed rates and, for a
  /// new income, the facts last used in the regime.
  void applyDefaults(IncomeDefaults? defaults) {
    _defaults = defaults;
    if (defaults != null) {
      final chosen = IvaRate.fromJson(ivaRate.value ?? '');
      if (asksIvaRate &&
          (!_suggestions.rateChosenByUser(chosen) ||
              !defaults.ivaRatesAllowed.contains(chosen))) {
        _suggestions.suggestRate(ivaRate, defaults.suggestedIvaRate);
      }
      if (defaults.lastUsed case final lastUsed? when !isEditing) {
        facts.applyLastUsed(lastUsed);
      }
    }
    _recompute();
  }

  /// Housing rent carries no IVA: the rate is held at exempt meanwhile.
  void _followFacts() {
    _housing.follow(isHousing: facts.isResidentialRent);
    notifyListeners();
  }

  /// Any amount or the rate changed: the IVA suggestion and the total
  /// follow.
  void _recompute() {
    _suggestions.recompute(IvaRate.fromJson(ivaRate.value ?? ''));
    notifyListeners();
  }

  /// "Guardar y crear otro" saved: a new session (a new id) starts. The
  /// form clears its non-context fields itself.
  void startNextSession() {
    _sessionId = _newId();
    _suggestions.forget();
  }

  @override
  void dispose() {
    form.dispose();
    super.dispose();
  }
}
