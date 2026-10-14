import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';

/// Amounts a form prefills (an IVA from its rate, a withholding from its
/// rate) while the user has not written their own: a value the user typed
/// (the CFDI's) always wins, and clearing it hands the field back to the
/// suggestion.
class MoneySuggestions {
  final Map<MoneyFieldController, String> _suggested = {};

  /// Writes [cents] (empty when `null`) into [field], unless the user wrote
  /// a value of their own there.
  void suggest(MoneyFieldController field, int? cents) {
    final suggestion = cents == null ? '' : formatCentavos(cents);
    final current = field.text.text;
    final isUsers = current.isNotEmpty && current != (_suggested[field] ?? '');
    if (isUsers) return;
    _suggested[field] = suggestion;
    if (current != suggestion) field.text.text = suggestion;
  }

  /// Takes what [field] holds now as a suggestion: it keeps following.
  void adopt(MoneyFieldController field) => _suggested[field] = field.text.text;

  /// A new form session: whatever is in the fields now is not the user's.
  void forget() => _suggested.clear();
}
