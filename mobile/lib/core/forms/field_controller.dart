import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:flutter/widgets.dart';

/// One text field's text, focus and error, and the house validation policy
/// (~/.claude/rules/ui-forms-and-modals.md, "Validation and errors"):
///
/// - required-but-empty shows only on submit ([validateForSubmit]);
/// - a format error shows on blur once the field has a value;
/// - once an error shows, every change re-checks it and clears it the moment
///   the value is valid; never while a fresh value is being typed.
///
/// Created through `AppFormController`, which owns and disposes it.
class AppFieldController extends AppFormFieldModel {
  /// A field labeled [label]; [validator] checks non-empty values. With
  /// [keepForNext], "Guardar y crear otro" keeps its value (a context field).
  new({
    required super.label,
    super.isRequired,
    super.keepForNext,
    this.validator,
    String initialText = '',
  }) : _initialText = initialText,
       _lastChecked = initialText,
       text = TextEditingController(text: initialText) {
    text.addListener(_handleChanged);
    focusNode.addListener(_handleFocusChanged);
  }

  /// Checks a non-empty value.
  final FieldValidator? validator;

  /// The text being edited.
  final TextEditingController text;

  final String _initialText;
  String _lastChecked;

  /// The trimmed value.
  String get value => text.text.trim();

  @override
  bool get isDirty => text.text != _initialText;

  /// Policy on leaving the field: a non-empty value is checked and, when
  /// valid, rewritten in its canonical form.
  void handleBlur() {
    if (value.isEmpty) return;
    setError(check());
    if (error == null) _normalize();
  }

  @override
  void reset() {
    setError(null);
    text.clear();
  }

  /// The canonical text of a valid value; the value itself by default.
  @protected
  String canonical(String value) => value;

  @override
  @protected
  FieldError? check() {
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
    if (error != null) setError(check());
  }

  void _handleFocusChanged() {
    if (!focusNode.hasFocus) handleBlur();
  }

  @override
  void dispose() {
    text.dispose();
    super.dispose();
  }
}
