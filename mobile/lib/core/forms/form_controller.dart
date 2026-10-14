import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:flutter/foundation.dart';

/// Which fields carry a requirement mark in their label: the fewer kind
/// (ui-forms-and-modals.md, "Mark required fields (or the optional ones —
/// whichever is fewer) in text").
enum RequirementMark {
  /// Optional fields read "(opcional)": most fields are required.
  optionalFields,

  /// Required fields read "(obligatorio)": most fields are optional.
  requiredFields,
}

/// A form's fields, in order, plus its submit state: the failed-submit
/// count the error summary reacts to and the busy state of the primary
/// button. Owns and disposes the fields it creates.
class AppFormController extends ChangeNotifier {
  /// A form that marks fields per [marks].
  new({this.marks = RequirementMark.optionalFields});

  /// Which fields carry a requirement mark.
  final RequirementMark marks;

  final List<AppFormFieldModel> _fields = [];
  bool _isSaving = false;
  bool _isDisposed = false;
  int _submitFailures = 0;

  /// The fields, in reading order.
  List<AppFormFieldModel> get fields => List.unmodifiable(_fields);

  /// Fields showing an error, in reading order.
  List<AppFormFieldModel> get invalidFields => [
    for (final field in _fields)
      if (field.error != null) field,
  ];

  /// Whether a save is in flight (the primary button shows busy).
  bool get isSaving => _isSaving;

  /// How many submits failed validation; each one moves focus to the
  /// error summary.
  int get submitFailures => _submitFailures;

  /// Whether any field differs from what the form started with.
  bool get isDirty => _fields.any((field) => field.isDirty);

  /// Adds a text field (see [AppFieldController.new]).
  AppFieldController textField({
    required String label,
    bool isRequired = true,
    bool keepForNext = false,
    FieldValidator? validator,
    bool verbatim = false,
    String Function(String value)? canonicalize,
    String initialText = '',
  }) => add(
    AppFieldController(
      label: label,
      isRequired: isRequired,
      keepForNext: keepForNext,
      validator: validator,
      verbatim: verbatim,
      canonicalize: canonicalize,
      initialText: initialText,
    ),
  );

  /// Adds a peso amount field.
  MoneyFieldController moneyField({
    required String label,
    bool isRequired = true,
    bool keepForNext = false,
    String initialText = '',
  }) => add(
    MoneyFieldController(
      label: label,
      isRequired: isRequired,
      keepForNext: keepForNext,
      initialText: initialText,
    ),
  );

  /// Adds a calendar date field.
  DateFieldController dateField({
    required String label,
    bool isRequired = true,
    bool keepForNext = false,
    String initialText = '',
  }) => add(
    DateFieldController(
      label: label,
      isRequired: isRequired,
      keepForNext: keepForNext,
      initialText: initialText,
    ),
  );

  /// Adds a checkbox (see [AppCheckboxController.new]).
  AppCheckboxController checkboxField({
    required String label,
    bool isRequired = true,
    bool keepForNext = false,
  }) => add(
    AppCheckboxController(
      label: label,
      isRequired: isRequired,
      keepForNext: keepForNext,
    ),
  );

  /// Validates every field and, when all pass, runs [save] with the busy
  /// state on. Returns whether [save] ran to completion.
  ///
  /// A submit while saving is ignored (the button stays enabled; repeat
  /// taps do nothing). An exception from [save] propagates after the busy
  /// state is cleared.
  Future<bool> submit(Future<void> Function() save) async {
    if (_isSaving) return false;
    final allValid = _fields
        .map((field) => field.validateForSubmit())
        .fold(true, (all, valid) => all && valid);
    if (!allValid) {
      _submitFailures++;
      notifyListeners();
      return false;
    }
    _setSaving(true);
    try {
      await save();
      return true;
    } finally {
      _setSaving(false);
    }
  }

  /// Shows the server's verdict (it is the authority): each field in
  /// [errors] shows its error, and the summary reacts as to a failed submit.
  /// An error clears once the user changes that field's value.
  void showServerErrors(Map<AppFormFieldModel, FieldError> errors) {
    if (errors.isEmpty) return;
    for (final MapEntry(key: field, value: error) in errors.entries) {
      assert(_fields.contains(field), '${field.label} is not in this form.');
      field.showError(error);
    }
    _submitFailures++;
    notifyListeners();
  }

  /// Readies the form for the next record ("Guardar y crear otro"): keeps
  /// context fields, clears the rest, and focuses the first cleared field.
  void prepareNext() {
    final cleared = [
      for (final field in _fields)
        if (!field.keepForNext) field,
    ];
    for (final field in cleared) {
      field.reset();
    }
    if (cleared.isNotEmpty) cleared.first.focusNode.requestFocus();
  }

  /// Adds [field], of any kind (a choice, a set of choices), under the
  /// form's policy; the form then owns and disposes it.
  T add<T extends AppFormFieldModel>(T field) {
    _fields.add(field);
    field.addListener(notifyListeners);
    return field;
  }

  void _setSaving(bool saving) {
    // A save can outlive its screen (the user leaves while it is in flight).
    if (_isDisposed) return;
    _isSaving = saving;
    notifyListeners();
  }

  @override
  void dispose() {
    _isDisposed = true;
    for (final field in _fields) {
      field.dispose();
    }
    super.dispose();
  }
}
