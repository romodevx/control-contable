import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:flutter/foundation.dart';

/// A checkbox field (a consent, an opt-in). A required box left unchecked is
/// an error on submit only; once an error shows, every change re-checks it.
///
/// Created through `AppFormController.checkboxField`.
class AppCheckboxController extends AppFormFieldModel {
  /// A box labeled [label], unchecked unless [initialValue].
  new({
    required super.label,
    super.isRequired,
    super.keepForNext,
    bool initialValue = false,
  }) : _initialValue = initialValue,
       _checked = initialValue;

  bool _initialValue;
  bool _checked;

  /// Whether the box is checked.
  bool get checked => _checked;

  set checked(bool value) {
    if (value == _checked) return;
    _checked = value;
    notifyListeners();
    if (error != null) setError(check());
  }

  @override
  bool get isDirty => _checked != _initialValue;

  @override
  void markPristine() => _initialValue = _checked;

  @override
  void reset() {
    setError(null);
    checked = false;
  }

  @override
  @protected
  FieldError? check() => isRequired && !_checked ? const NotAccepted() : null;
}
