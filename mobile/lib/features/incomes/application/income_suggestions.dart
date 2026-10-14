import 'package:control_contable/api/generated/models/decimal_rate.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/features/incomes/domain/income_math.dart';

/// Prefills the derived amounts of an income (IVA from the rate, the
/// withholdings from the defaults' rates) while the user has not written
/// their own: a value the user typed (the CFDI's) always wins, and
/// clearing it hands the field back to the suggestion.
class IncomeSuggestions {
  /// Suggestions for the fields of one form.
  new({
    required this.amount,
    required this.ivaAmount,
    required this.isrWithheld,
    required this.ivaWithheld,
  });

  /// Subtotal: the base of the IVA and of the ISR withholding.
  final MoneyFieldController amount;

  /// IVA charged: the base of the IVA withholding.
  final MoneyFieldController ivaAmount;

  /// ISR withheld.
  final MoneyFieldController isrWithheld;

  /// IVA withheld.
  final MoneyFieldController ivaWithheld;

  final Map<MoneyFieldController, String> _suggested = {};
  String? _suggestedRate = IvaRate.rate16.json;
  bool _isApplying = false;

  /// Whether [current] is the user's choice rather than a suggestion.
  bool rateChosenByUser(IvaRate current) => current.json != _suggestedRate;

  /// Suggests [rate] in [field].
  void suggestRate(ChoiceFieldController field, IvaRate rate) {
    _suggestedRate = rate.json;
    field.value = rate.json;
  }

  /// Recomputes every suggestion from the amount, [rate] and [defaults].
  void recompute(IvaRate rate, IncomeDefaults? defaults) {
    // Writing a suggestion notifies its field, which asks again.
    if (_isApplying) return;
    _isApplying = true;
    try {
      final base = amount.centavos;
      _suggest(ivaAmount, base == null ? null : ivaCentsFor(base, rate));
      _suggest(isrWithheld, _withheld(base, defaults?.isrWithholdingRate));
      _suggest(
        ivaWithheld,
        _withheld(ivaAmount.centavos, defaults?.ivaWithholdingRate),
      );
    } finally {
      _isApplying = false;
    }
  }

  /// An edit starts from a stored income: its rate and amounts are the
  /// user's, except the IVA when it is exactly the rate's IVA
  /// ([ivaFollowsAmount]): it then keeps following the amount.
  void adoptStored({required bool ivaFollowsAmount}) {
    _suggestedRate = null;
    if (ivaFollowsAmount) _suggested[ivaAmount] = ivaAmount.text.text;
  }

  /// A new form session: whatever is in the fields now is not the user's.
  void forget() => _suggested.clear();

  int? _withheld(int? base, DecimalRate? rate) =>
      base == null || rate == null ? null : withheldCents(base, rate);

  void _suggest(MoneyFieldController field, int? cents) {
    final suggestion = cents == null ? '' : formatCentavos(cents);
    final current = field.text.text;
    final isUsers = current.isNotEmpty && current != (_suggested[field] ?? '');
    if (isUsers) return;
    _suggested[field] = suggestion;
    if (current != suggestion) field.text.text = suggestion;
  }
}
