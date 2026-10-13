import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// The login form's fields. The password gets no length check: the server
/// decides, and a client rule could lock out a password set under an older
/// policy. Owns its `AppFormController`.
class LoginFields {
  /// The fields, labeled and validated with the localized copy.
  new(this._l10n) {
    email = form.textField(
      label: _l10n.authEmailLabel,
      validator: (value) => checkEmail(value, _l10n),
    );
    password = form.textField(label: _l10n.authPasswordLabel, verbatim: true);
  }

  final AppLocalizations _l10n;

  /// The form the fields belong to.
  final AppFormController form = AppFormController();

  /// The account's email (trimmed).
  late final AppFieldController email;

  /// The current password (sent exactly as typed).
  late final AppFieldController password;

  /// Releases the form and its fields.
  void dispose() => form.dispose();
}
