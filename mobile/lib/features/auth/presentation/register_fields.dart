import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// `RegisterRequest` limits (docs/api/openapi.yaml).
const int _emailMaxLength = 254;
const int _nameMaxLength = 120;

/// The body paths `register` reports field errors under.
const String _emailPath = 'email';
const String _documentsPath = 'accepted_documents';

/// A plausible address: something@something.tld, no spaces. The server
/// decides; this only catches typos on blur.
final RegExp _emailShape = RegExp(r'^[^@\s]+@[^@\s]+\.[^@\s]+$');

/// The sign-up form's fields, in reading order, and how the server's
/// verdict maps onto them. Owns its `AppFormController`.
class RegisterFields {
  /// The fields, labeled and validated with the localized copy.
  new(this._l10n) {
    name = form.textField(
      label: _l10n.authNameLabel,
      isRequired: false,
      validator: (value) => value.length > _nameMaxLength
          ? CustomFieldError(_l10n.authNameTooLong)
          : null,
    );
    email = form.textField(
      label: _l10n.authEmailLabel,
      validator: (value) =>
          value.length > _emailMaxLength || !_emailShape.hasMatch(value)
          ? CustomFieldError(_l10n.authEmailInvalid)
          : null,
    );
    password = form.textField(
      label: _l10n.authPasswordLabel,
      validator: (value) => checkNewPassword(value, _l10n),
      verbatim: true,
    );
    privacy = form.checkboxField(label: _l10n.authConsentPrivacy);
    terms = form.checkboxField(label: _l10n.authConsentTerms);
  }

  final AppLocalizations _l10n;

  /// The form the fields belong to.
  final AppFormController form = AppFormController();

  /// Optional display name.
  late final AppFieldController name;

  /// The sign-up email.
  late final AppFieldController email;

  /// The new password (sent exactly as typed, never trimmed).
  late final AppFieldController password;

  /// Consent to the privacy notice.
  late final AppCheckboxController privacy;

  /// Consent to the terms and conditions.
  late final AppCheckboxController terms;

  /// The `register` body for the [documents] shown and accepted: the email
  /// and the consents only. The password and name wait in [draft] for the
  /// code step (D8).
  RegisterRequest request(LegalDocumentSet documents) => RegisterRequest(
    email: email.value,
    acceptedDocuments: documents.acceptances,
  );

  /// What the code step sends with the code.
  SignUpDraft draft() => SignUpDraft(
    email: email.value,
    password: password.value,
    name: name.value.isEmpty ? null : name.value,
  );

  /// Shows the server's field [details] on their fields. A stale document
  /// version unchecks both consents (the new versions must be read and
  /// accepted again). Says whether the documents were stale and whether
  /// some detail matched no field.
  ({bool staleDocuments, bool unmatched}) showRejected(
    List<api.FieldError> details,
  ) {
    final staleDocuments = details.any(
      (detail) => detail.field.startsWith(_documentsPath),
    );
    if (staleDocuments) {
      privacy.checked = false;
      terms.checked = false;
    }
    final match = matchServerFieldErrors(
      details,
      fieldsByPath: {_emailPath: email, _documentsPath: privacy},
      describe: (detail) => describeSignUpDetail(detail, _l10n),
    );
    form.showServerErrors({...match.matched, terms: ?match.matched[privacy]});
    return (
      staleDocuments: staleDocuments,
      unmatched: match.unmatched.isNotEmpty,
    );
  }

  /// Releases the form and its fields.
  void dispose() => form.dispose();
}
