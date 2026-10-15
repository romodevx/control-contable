import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';

/// The detail code of a user-text field holding an invisible character
/// (control, format, line or paragraph separator; ADR-0004 Amendment 1).
const String invisibleCharactersCode = 'invisible_characters';

/// Separator of the server's field paths (`accepted_documents.1.kind`;
/// docs/api/CONTRACT.md).
const _pathSeparator = '.';

/// Where each server field error goes: `matched` onto form fields (ready
/// for `AppFormController.showServerErrors`), `unmatched` for the screen to
/// report as a form-level message.
typedef ServerFieldErrorMatch = ({
  Map<AppFormFieldModel, FieldError> matched,
  List<api.FieldError> unmatched,
});

/// Matches the `details` of an `ErrorResponse` to form fields by their
/// dotted path relative to the request body.
///
/// A detail lands on the field mapped to its own path or, failing that, to
/// its nearest parent path (`accepted_documents.1.kind` →
/// `accepted_documents`); a parent matches whole segments only. The first
/// detail for a field wins. [describe] turns a detail into the copy the user
/// reads (the server's `message` is English and generic).
ServerFieldErrorMatch matchServerFieldErrors(
  List<api.FieldError> details, {
  required Map<String, AppFormFieldModel> fieldsByPath,
  required FieldError Function(api.FieldError detail) describe,
}) {
  final matched = <AppFormFieldModel, FieldError>{};
  final unmatched = <api.FieldError>[];
  for (final detail in details) {
    final field = _nearestField(detail.field, fieldsByPath);
    if (field == null) {
      unmatched.add(detail);
    } else {
      matched.putIfAbsent(field, () => describe(detail));
    }
  }
  return (matched: matched, unmatched: unmatched);
}

AppFormFieldModel? _nearestField(
  String path,
  Map<String, AppFormFieldModel> fieldsByPath,
) {
  final segments = path.split(_pathSeparator);
  for (var length = segments.length; length > 0; length--) {
    final candidate = segments.take(length).join(_pathSeparator);
    final field = fieldsByPath[candidate];
    if (field != null) return field;
  }
  return null;
}
