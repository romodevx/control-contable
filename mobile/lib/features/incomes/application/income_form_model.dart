import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/incomes/application/income_suggestions.dart';
import 'package:control_contable/features/incomes/domain/income_basis.dart';
import 'package:control_contable/features/incomes/domain/income_math.dart';
import 'package:flutter/foundation.dart';

/// The 625 regime (platform incomes), the only one with an activity.
const String platformRegime = '625';

/// The regime option that lists a 625 taxpayer's activities.
const String _activitiesOption = 'activities';

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
  String counterparty,
  String regime,
  String activity,
  String isrWithheld,
  String ivaWithheld,
  String ivaAmount,
  String cfdiUuid,
  String cfdiUuidInvalid,
});

/// What the new-income page holds (M3-09): its fields in the list's
/// column order, the defaults' suggestions, and the idempotency id of the
/// current form session (MODULE-incomes §2.3).
///
/// Context fields (dates, regime, who paid, IVA rate, activity) survive
/// "Guardar y crear otro"; the rest is cleared and a new session id starts.
class IncomeFormModel extends ChangeNotifier {
  /// The form for [regimes] (those in effect today), starting on [today].
  new({
    required List<TaxpayerRegime> regimes,
    required DateTime today,
    required CounterpartyType counterparty,
    required this.copy,
    required String Function() newId,
  }) : _regimes = regimes,
       _newId = newId,
       basis = incomeBasisOf(regimes.first.regimeCode),
       _sessionId = newId() {
    _build(today, counterparty);
    _suggestions = IncomeSuggestions(
      amount: amount,
      ivaAmount: ivaAmount,
      isrWithheld: isrWithheld,
      ivaWithheld: ivaWithheld,
    );
    for (final field in [amount, ivaAmount, isrWithheld, ivaWithheld]) {
      field.text.addListener(_recompute);
    }
    ivaRate.addListener(_recompute);
    form.addListener(notifyListeners);
  }

  /// The copy the fields were labeled with.
  final IncomeFormCopy copy;

  /// Which date books the income (from the first regime).
  final IncomeBasis basis;

  final List<TaxpayerRegime> _regimes;
  final String Function() _newId;
  String _sessionId;
  IncomeDefaults? _defaults;
  late final IncomeSuggestions _suggestions;

  /// The fields, in reading order.
  final AppFormController form = AppFormController();

  /// The booking date (collection, or invoice for an accrual regime).
  late final DateFieldController bookingDate;

  /// What the income was.
  late final AppFieldController description;

  /// Subtotal before IVA.
  late final MoneyFieldController amount;

  /// The IVA rate (`IvaRate.json`).
  late final ChoiceFieldController ivaRate;

  /// Who paid (`CounterpartyType.json`).
  late final ChoiceFieldController counterparty;

  /// The regime; shown only with more than one in effect.
  late final ChoiceFieldController regime;

  /// The 625 activity; shown only when 625 holds several.
  late final ChoiceFieldController activity;

  /// ISR the payer withheld.
  late final MoneyFieldController isrWithheld;

  /// IVA the payer withheld.
  late final MoneyFieldController ivaWithheld;

  /// IVA charged; prefilled from the rate (the rate's IVA when empty).
  late final MoneyFieldController ivaAmount;

  /// The date that does not book the income.
  late final DateFieldController otherDate;

  /// The CFDI's folio fiscal.
  late final AppFieldController cfdiUuid;

  void _build(DateTime today, CounterpartyType payer) {
    final accrual = basis == IncomeBasis.accrual;
    bookingDate = form.dateField(
      label: accrual ? copy.invoicedOn : copy.collectedOn,
      keepForNext: true,
      initialText: formatDate(today),
    );
    description = form.textField(label: copy.description, isRequired: false);
    amount = form.moneyField(label: copy.amount);
    ivaRate = form.add(
      ChoiceFieldController(
        label: copy.ivaRate,
        keepForNext: true,
        initialValue: IvaRate.rate16.json,
      ),
    );
    counterparty = form.add(
      ChoiceFieldController(
        label: copy.counterparty,
        keepForNext: true,
        initialValue: payer.json,
      ),
    );
    regime = form.add(
      ChoiceFieldController(
        label: copy.regime,
        keepForNext: true,
        initialValue: _regimes.first.regimeCode,
      ),
    );
    activity = form.add(
      ChoiceFieldController(
        label: copy.activity,
        isRequired: false,
        keepForNext: true,
      ),
    );
    isrWithheld = form.moneyField(label: copy.isrWithheld, isRequired: false);
    ivaWithheld = form.moneyField(label: copy.ivaWithheld, isRequired: false);
    // Prefilled from the rate; left empty, the rate's IVA is sent.
    ivaAmount = form.moneyField(label: copy.ivaAmount, isRequired: false);
    otherDate = form.dateField(
      label: accrual ? copy.collectedOn : copy.invoicedOn,
      isRequired: false,
      keepForNext: true,
    );
    cfdiUuid = form.textField(
      label: copy.cfdiUuid,
      isRequired: false,
      validator: (value) => _cfdiUuid.hasMatch(value)
          ? null
          : CustomFieldError(copy.cfdiUuidInvalid),
      canonicalize: (value) => value.toUpperCase(),
    );
  }

  /// The idempotency id of this form session: every retry resends it.
  String get sessionId => _sessionId;

  /// The regimes the user can choose (more than one shows the field).
  List<TaxpayerRegime> get regimes => List.unmodifiable(_regimes);

  /// The chosen regime's code.
  String get regimeCode => regime.value ?? _regimes.first.regimeCode;

  /// Who paid.
  CounterpartyType get payer =>
      CounterpartyType.fromJson(counterparty.value ?? '');

  /// The 625 activities held, when the chosen regime is 625.
  List<PlatformActivity> get activities {
    if (regimeCode != platformRegime) return const [];
    final held = _regimes
        .firstWhere((r) => r.regimeCode == platformRegime)
        .options[_activitiesOption];
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

  /// The booking date as the contract's `YYYY-MM-DD`, when valid.
  String? get bookingDay {
    final date = bookingDate.date;
    return date == null ? null : businessDate(date);
  }

  /// The last defaults answer, or `null` before one (or after a failure).
  IncomeDefaults? get defaults => _defaults;

  /// The IVA rates to offer: the allowed ones once known.
  List<IvaRate> get ivaRatesOffered =>
      _defaults?.ivaRatesAllowed ?? IvaRate.$valuesDefined;

  /// Whether the payer may have withheld anything (unknown counts as yes:
  /// the server decides).
  bool get withholdingsApply {
    final known = _defaults;
    return known == null ||
        known.isrWithholdingAllowed ||
        known.ivaWithholdingAllowed;
  }

  /// The IVA the user wrote, else the one the rate gives.
  int get _ivaCents =>
      ivaAmount.centavos ??
      ivaCentsFor(amount.centavos ?? 0, IvaRate.fromJson(ivaRate.value ?? ''));

  /// The live amounts (what the summary under the amount reads).
  IncomeAmounts get amounts => IncomeAmounts(
    amountCents: amount.centavos ?? 0,
    ivaCents: _ivaCents,
    isrWithheldCents: withholdingsApply ? isrWithheld.centavos ?? 0 : 0,
    ivaWithheldCents: withholdingsApply ? ivaWithheld.centavos ?? 0 : 0,
  );

  /// Applies the defaults for the current regime, payer and date: the
  /// suggested IVA rate (unless the user chose one), the allowed rates, and
  /// the withholding suggestions.
  void applyDefaults(IncomeDefaults? defaults) {
    _defaults = defaults;
    if (defaults != null) {
      final chosen = IvaRate.fromJson(ivaRate.value ?? '');
      if (!_suggestions.rateChosenByUser(chosen) ||
          !defaults.ivaRatesAllowed.contains(chosen)) {
        _suggestions.suggestRate(ivaRate, defaults.suggestedIvaRate);
      }
    }
    _recompute();
  }

  /// Any amount or the rate changed: suggestions and the summary follow.
  void _recompute() {
    _suggestions.recompute(IvaRate.fromJson(ivaRate.value ?? ''), _defaults);
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
