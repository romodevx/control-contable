import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// Digits in an emailed verification code (contract: `^[0-9]{6}$`).
const int codeLength = 6;

final RegExp _codeShape = RegExp('^[0-9]{$codeLength}\$');

/// The body paths `verifyEmail` reports field errors under.
const String _codePath = 'code';
const String _passwordPath = 'password';

/// The code step's fields: the code, and the password chosen at sign-up.
/// The password field holds the draft's password and is only shown when it
/// must be typed (again): the server rejected it, or the app went to the
/// background and forgot it. Owns its `AppFormController`.
class CodeStepFields {
  /// The fields for [draft], labeled with the localized copy.
  new(this._l10n, SignUpDraft draft) {
    code = form.textField(
      label: _l10n.authCodeLabel,
      validator: (value) => _codeShape.hasMatch(value)
          ? null
          : CustomFieldError(_l10n.authCodeFormat),
    );
    password = form.textField(
      label: _l10n.authPasswordLabel,
      verbatim: true,
      validator: (value) => checkNewPassword(value, _l10n),
      initialText: draft.password ?? '',
    );
  }

  final AppLocalizations _l10n;

  /// The form the fields belong to.
  final AppFormController form = AppFormController();

  /// The 6-digit code.
  late final AppFieldController code;

  /// The new account's password (D8).
  late final AppFieldController password;

  /// Whether a complete code can be sent on its own (the password is known).
  bool get isReadyToSend => password.value.isNotEmpty;

  /// The `verifyEmail` body for [draft].
  VerifyEmailRequest request(SignUpDraft draft) => VerifyEmailRequest(
    email: draft.email,
    code: code.value,
    password: password.value,
    name: draft.name,
  );

  /// Forgets the password typed or carried over (the app went to the
  /// background).
  void forgetPassword() => password.text.clear();

  /// Shows a rejected code (401: one generic message).
  void showCodeRejected() =>
      form.showServerErrors({code: CustomFieldError(_l10n.authCodeRejected)});

  /// Shows the server's field [details] on the code or password field. Says
  /// whether the password was rejected and whether some detail matched no
  /// field.
  ({bool password, bool unmatched}) showRejected(List<api.FieldError> details) {
    final match = matchServerFieldErrors(
      details,
      fieldsByPath: {_codePath: code, _passwordPath: password},
      describe: (detail) => describeSignUpDetail(detail, _l10n),
    );
    form.showServerErrors(match.matched);
    return (
      password: match.matched.containsKey(password),
      unmatched: match.unmatched.isNotEmpty,
    );
  }

  /// Releases the form and its fields.
  void dispose() => form.dispose();
}
