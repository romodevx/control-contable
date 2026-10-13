import 'package:control_contable/core/forms/field_error.dart';
import 'package:flutter/widgets.dart';

/// What a form needs from any of its fields, whatever the input (text,
/// checkbox): a label, a focus target, an error, and the submit policy.
///
/// Created through `AppFormController`, which owns and disposes it.
abstract class AppFormFieldModel extends ChangeNotifier {
  /// A field labeled [label].
  new({required this.label, this.isRequired = true, this.keepForNext = false});

  /// The visible label, also used by the error summary.
  final String label;

  /// Whether an empty value is an error.
  final bool isRequired;

  /// Whether the value survives "Guardar y crear otro".
  final bool keepForNext;

  /// The field's focus; the error summary moves it here.
  final FocusNode focusNode = FocusNode();

  FieldError? _error;

  /// The error the user sees, if any.
  FieldError? get error => _error;

  /// Whether the value differs from what the form started with.
  bool get isDirty;

  /// Shows any error the value has now; `true` when it has none.
  bool validateForSubmit() {
    setError(check());
    return _error == null;
  }

  /// Shows [error], found outside the client checks (the server's verdict).
  /// Like any shown error, it is re-checked once the value changes.
  void showError(FieldError error) => setError(error);

  /// Clears the value and any error (used by "Guardar y crear otro").
  void reset();

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
