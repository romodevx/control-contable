import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// `RegisterRequest` limits (docs/api/openapi.yaml).
const int _passwordMinLength = 12;
const int _passwordMaxLength = 128;
const int _emailMaxLength = 254;
const int _nameMaxLength = 120;

/// The body paths `register` reports field errors under.
const String _emailPath = 'email';
const String _passwordPath = 'password';
const String _namePath = 'name';
const String _documentsPath = 'accepted_documents';

/// Server `FieldError.code` values with their own copy.
const String _passwordPolicyCode = 'password_policy';
const String _tooShortCode = 'minLength';
const String _tooLongCode = 'maxLength';

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
      validator: _checkPassword,
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

  /// The `register` body for the [documents] shown and accepted.
  RegisterRequest request(LegalDocumentSet documents) => RegisterRequest(
    email: email.value,
    password: password.value,
    acceptedDocuments: documents.acceptances,
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
      fieldsByPath: {
        _emailPath: email,
        _passwordPath: password,
        _namePath: name,
        _documentsPath: privacy,
      },
      describe: _describe,
    );
    form.showServerErrors({...match.matched, terms: ?match.matched[privacy]});
    return (
      staleDocuments: staleDocuments,
      unmatched: match.unmatched.isNotEmpty,
    );
  }

  FieldError? _checkPassword(String value) {
    final length = value.runes.length;
    if (length < _passwordMinLength) {
      return CustomFieldError(_l10n.authPasswordTooShort);
    }
    if (length > _passwordMaxLength) {
      return CustomFieldError(_l10n.authPasswordTooLong);
    }
    return null;
  }

  FieldError _describe(api.FieldError detail) {
    final root = detail.field.split('.').first;
    final message = switch ((root, detail.code)) {
      (_passwordPath, _passwordPolicyCode) => _l10n.authPasswordPolicy,
      (_passwordPath, _tooShortCode) => _l10n.authPasswordTooShort,
      (_passwordPath, _tooLongCode) => _l10n.authPasswordTooLong,
      (_emailPath, _) => _l10n.authEmailInvalid,
      (_namePath, _) => _l10n.authNameTooLong,
      (_documentsPath, _) => _l10n.authDocumentsStale,
      _ => _l10n.authFieldRejected,
    };
    return CustomFieldError(message);
  }

  /// Releases the form and its fields.
  void dispose() => form.dispose();
}
