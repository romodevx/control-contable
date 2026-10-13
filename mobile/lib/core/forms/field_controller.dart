import 'package:control_contable/core/forms/field_error.dart';
import 'package:flutter/widgets.dart';

/// One form field's text, focus and error, and the house validation policy
/// (~/.claude/rules/ui-forms-and-modals.md, "Validation and errors"):
///
/// - required-but-empty shows only on submit ([validateForSubmit]);
/// - a format error shows on blur once the field has a value;
/// - once an error shows, every change re-checks it and clears it the moment
///   the value is valid; never while a fresh value is being typed.
///
/// Created through `AppFormController`, which owns and disposes it.
class AppFieldController extends ChangeNotifier {
  /// A field labeled [label]; [validator] checks non-empty values. With
  /// [keepForNext], "Guardar y crear otro" keeps its value (a context field).
  new({
    required this.label,
    this.isRequired = true,
    this.keepForNext = false,
    this.validator,
    String initialText = '',
  }) : _initialText = initialText,
       _lastChecked = initialText,
       text = TextEditingController(text: initialText) {
    text.addListener(_handleChanged);
    focusNode.addListener(_handleFocusChanged);
  }

  /// The visible label, also used by the error summary.
  final String label;

  /// Whether an empty value is an error.
  final bool isRequired;

  /// Whether the value survives "Guardar y crear otro".
  final bool keepForNext;

  /// Checks a non-empty value.
  final FieldValidator? validator;

  /// The text being edited.
  final TextEditingController text;

  /// The field's focus; the error summary moves it here.
  final FocusNode focusNode = FocusNode();

  final String _initialText;
  FieldError? _error;
  String _lastChecked;

  /// The error the user sees, if any.
  FieldError? get error => _error;

  /// The trimmed value.
  String get value => text.text.trim();

  /// Whether the text differs from what the form started with.
  bool get isDirty => text.text != _initialText;

  /// Shows any error the value has now; `true` when it has none.
  bool validateForSubmit() {
    _setError(_check());
    return _error == null;
  }

  /// Policy on leaving the field: a non-empty value is checked and, when
  /// valid, rewritten in its canonical form.
  void handleBlur() {
    if (value.isEmpty) return;
    _setError(_check());
    if (_error == null) _normalize();
  }

  /// Shows [error], found outside the client checks (the server's verdict).
  /// Like any shown error, it is re-checked once the value changes.
  void showError(FieldError error) => _setError(error);

  /// Clears the value and any error (used by "Guardar y crear otro").
  void reset() {
    _setError(null);
    text.clear();
  }

  /// The canonical text of a valid value; the value itself by default.
  @protected
  String canonical(String value) => value;

  FieldError? _check() {
    if (value.isEmpty) return isRequired ? const MissingValue() : null;
    return validator?.call(value);
  }

  void _normalize() {
    final normalized = canonical(value);
    if (normalized != text.text) text.text = normalized;
  }

  void _handleChanged() {
    // Selection moves also notify; only a new value re-checks.
    if (text.text == _lastChecked) return;
    _lastChecked = text.text;
    if (_error != null) _setError(_check());
  }

  void _handleFocusChanged() {
    if (!focusNode.hasFocus) handleBlur();
  }

  void _setError(FieldError? error) {
    if (error == _error) return;
    _error = error;
    notifyListeners();
  }

  @override
  void dispose() {
    text.dispose();
    focusNode.dispose();
    super.dispose();
  }
}
