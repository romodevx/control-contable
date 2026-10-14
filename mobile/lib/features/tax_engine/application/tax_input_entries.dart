import 'package:control_contable/api/generated/models/tax_input_def.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';

/// The editable state of one `TaxInputDef`, by its kind.
sealed class TaxInputEntry {
  const new(this.def);

  /// The definition (key, label, kind, whether an estimate can need it).
  final TaxInputDef def;
}

/// `money`: one amount for the year.
final class MoneyEntry extends TaxInputEntry {
  /// The entry of [def] edited in [field].
  const new(super.def, this.field);

  /// The amount.
  final MoneyFieldController field;
}

/// `rate`: a decimal string in `[0, 1]`.
final class RateEntry extends TaxInputEntry {
  /// The entry of [def] edited in [field].
  const new(super.def, this.field);

  /// The rate as typed.
  final AppFieldController field;
}

/// `boolean`: yes or no.
final class FlagEntry extends TaxInputEntry {
  /// The entry of [def] edited in [field]; [wasStored] when a value exists.
  const new(super.def, this.field, {required this.wasStored});

  /// The box.
  final AppCheckboxController field;

  /// Whether the server holds a value (an unchecked box then saves `false`
  /// instead of leaving the key absent).
  final bool wasStored;
}

/// `money_monthly`: one amount per month that can already be known.
final class MonthlyEntry extends TaxInputEntry {
  /// The entry of [def] with one field per month in [months].
  const new(super.def, this.months);

  /// Month (1-12) → its amount, in month order.
  final Map<int, MoneyFieldController> months;
}

/// A kind this app version does not edit: its stored values are kept as
/// they are (the save replaces every value of the regime and year).
final class KeptEntry extends TaxInputEntry {
  /// The entry of [def] keeping [stored].
  const new(super.def, this.stored);

  /// The values the server holds for it.
  final List<TaxInputValue> stored;
}
