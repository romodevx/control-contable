import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:control_contable/core/forms/money.dart';

/// A peso amount field; its value is read as [centavos].
class MoneyFieldController extends AppFieldController {
  /// See [AppFieldController.new].
  new({
    required super.label,
    super.isRequired,
    super.keepForNext,
    super.initialText,
  }) : super(validator: _validate);

  static FieldError? _validate(String value) =>
      parseCentavos(value) == null ? const InvalidMoney() : null;

  /// The amount in centavos, or `null` while empty or invalid.
  int? get centavos => parseCentavos(value);

  @override
  String canonical(String value) => formatCentavos(parseCentavos(value)!);
}

/// A calendar date field; its value is read as [date].
class DateFieldController extends AppFieldController {
  /// See [AppFieldController.new].
  new({
    required super.label,
    super.isRequired,
    super.keepForNext,
    super.initialText,
  }) : super(validator: _validate);

  static FieldError? _validate(String value) =>
      parseFlexibleDate(value) == null ? const InvalidDate() : null;

  /// The date, or `null` while empty or invalid.
  DateTime? get date => parseFlexibleDate(value);

  @override
  String canonical(String value) => formatDate(parseFlexibleDate(value)!);
}
