import 'package:control_contable/core/forms/field_error.dart';
import 'package:flutter/widgets.dart';

/// What a form needs from any of its fields, whatever the input (text,
/// checkbox): a label, a focus target, an error, and the submit policy.
///
/// Created through `AppFormController`, which owns and disposes it.
abstract class AppFormFieldModel extends ChangeNotifier {
  /// A field labeled [label].
  new({required this.label, bool isRequired = true, this.keepForNext = false})
    : _isRequired = isRequired;

  /// The visible label, also used by the error summary.
  final String label;

  bool _isRequired;

  /// Whether an empty value is an error.
  bool get isRequired => _isRequired;

  /// Another choice can make the field required or optional (an income's
  /// dates follow its regime). A shown error is re-checked, so a
  /// "missing" error goes once the field turns optional; a field turning
  /// required shows nothing until the next submit.
  set isRequired(bool value) {
    if (value == _isRequired) return;
    _isRequired = value;
    if (_error != null) _error = check();
    notifyListeners();
  }

  /// Whether the value survives "Guardar y crear otro".
  final bool keepForNext;

  /// The field's focus; the error summary moves it here.
  final FocusNode focusNode = FocusNode();

  FieldError? _error;
  bool _isActive = true;

  /// Whether the field is part of the form now. A field another choice
  /// hides (a donation's donee when the category is not a donation) is
  /// inactive: never validated, never showing an error.
  bool get isActive => _isActive;

  set isActive(bool value) {
    if (value == _isActive) return;
    _isActive = value;
    if (!value) _error = null;
    notifyListeners();
  }

  /// The error the user sees, if any.
  FieldError? get error => _error;

  /// Whether the value differs from what the form started with.
  bool get isDirty;

  /// Shows any error the value has now; `true` when it has none.
  bool validateForSubmit() {
    setError(_isActive ? check() : null);
    return _error == null;
  }

  /// Shows [error], found outside the client checks (the server's verdict).
  /// Like any shown error, it is re-checked once the value changes.
  void showError(FieldError error) => setError(error);

  /// Clears the value and any error (used by "Guardar y crear otro").
  void reset();

  /// Makes the current value the one [isDirty] compares with (the next
  /// record of "Guardar y crear otro" starts here).
  void markPristine();

  /// The error the current value has, or `null`.
  @protected
  FieldError? check();

  /// Shows [error] (or none), notifying only on a change.
  @protected
  void setError(FieldError? error) {
    if (error == _error) return;
    _error = error;
    notifyListeners();
  }

  @override
  void dispose() {
    focusNode.dispose();
    super.dispose();
  }
}
