import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/fiscal/iva_math.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/money_suggestions.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';

/// Prefills the IVA of an income from its rate, and the rate from the
/// defaults, while the user has not chosen their own: a value the user
/// typed (the CFDI's) always wins, and clearing it hands the field back to
/// the suggestion. Withholdings are never suggested nor typed: the server
/// derives them (ADR-0008).
class IncomeSuggestions {
  /// Suggestions for the fields of one form.
  new({required this.amount, required this.ivaAmount});

  /// Subtotal: the base of the IVA.
  final MoneyFieldController amount;

  /// IVA charged.
  final MoneyFieldController ivaAmount;

  final MoneySuggestions _suggested = MoneySuggestions();
  String? _suggestedRate = IvaRate.rate16.json;
  bool _isApplying = false;

  /// The rate last suggested.
  IvaRate? get suggestedRate =>
      _suggestedRate == null ? null : IvaRate.fromJson(_suggestedRate!);

  /// Whether [current] is the user's choice rather than a suggestion.
  bool rateChosenByUser(IvaRate current) => current.json != _suggestedRate;

  /// Suggests [rate] in [field].
  void suggestRate(ChoiceFieldController field, IvaRate rate) {
    _suggestedRate = rate.json;
    field.value = rate.json;
  }

  /// Recomputes the IVA from the amount and [rate].
  void recompute(IvaRate rate) {
    // Writing a suggestion notifies its field, which asks again.
    if (_isApplying) return;
    _isApplying = true;
    try {
      final base = amount.centavos;
      _suggested.suggest(
        ivaAmount,
        base == null ? null : ivaCentsFor(base, rate),
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
    if (ivaFollowsAmount) _suggested.adopt(ivaAmount);
  }

  /// A new form session: whatever is in the fields now is not the user's.
  void forget() => _suggested.forget();
}
