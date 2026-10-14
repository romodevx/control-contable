import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:flutter/foundation.dart';

/// One option of a choice field: the `value` sent to the server and the
/// `label` the user reads.
typedef ChoiceOption = ({String value, String label});

/// A field whose value is one option of a fixed list (a radio group). A
/// required choice left empty is an error on submit only; once an error
/// shows, every change re-checks it.
///
/// Added to a form with `AppFormController.add`.
class ChoiceFieldController extends AppFormFieldModel {
  /// A choice labeled [label], starting at [initialValue] (none by default).
  new({
    required super.label,
    super.isRequired,
    super.keepForNext,
    String? initialValue,
  }) : _initialValue = initialValue,
       _value = initialValue;

  final String? _initialValue;
  String? _value;

  /// The chosen option's value, or `null` while none is chosen.
  String? get value => _value;

  set value(String? value) {
    if (value == _value) return;
    _value = value;
    notifyListeners();
    if (error != null) setError(check());
  }

  @override
  bool get isDirty => _value != _initialValue;

  @override
  void reset() {
    setError(null);
    value = null;
  }

  @override
  @protected
  FieldError? check() =>
      isRequired && _value == null ? const MissingChoice() : null;
}

/// A field whose value is a set of options (a checkbox group). A required
/// set needs at least one option.
///
/// Added to a form with `AppFormController.add`.
class MultiChoiceFieldController extends AppFormFieldModel {
  /// A set labeled [label], starting with [initialValues].
  new({
    required super.label,
    super.isRequired,
    super.keepForNext,
    Set<String> initialValues = const {},
  }) : _initialValues = Set.unmodifiable(initialValues),
       _values = {...initialValues};

  final Set<String> _initialValues;
  final Set<String> _values;

  /// The chosen values.
  Set<String> get values => Set.unmodifiable(_values);

  /// Adds [value] when absent, removes it when present.
  void toggle(String value) {
    if (!_values.remove(value)) _values.add(value);
    notifyListeners();
    if (error != null) setError(check());
  }

  @override
  bool get isDirty => !setEquals(_values, _initialValues);

  @override
  void reset() {
    setError(null);
    _values.clear();
    notifyListeners();
  }

  @override
  @protected
  FieldError? check() =>
      isRequired && _values.isEmpty ? const MissingChoice() : null;
}
