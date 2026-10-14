import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/features/auth/presentation/login_fields.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Login (US-A2a): email and password on their own screen (2 fields; house
/// UI rules). Bad credentials read as one generic message, whatever the
/// cause (no account enumeration). On success the session starts and the
/// router's guard takes the user where they were going.
class LoginScreen extends ConsumerStatefulWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  LoginFields? _fields;
  String? _notice;

  LoginFields get fields => _fields!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Labels and validators need the localized copy, not there in initState.
    _fields ??= LoginFields(AppLocalizations.of(context));
  }

  @override
  void dispose() {
    _fields?.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    final l10n = AppLocalizations.of(context);
    _showNotice(null);
    try {
      final session = await ref
          .read(authRepositoryProvider)
          .login(email: fields.email.value, password: fields.password.value);
      // Lets the password manager offer to save the credentials.
      TextInput.finishAutofillContext();
      await ref.read(sessionProvider).start(session.tokens);
    } on ApiUnauthorized {
      _showNotice(l10n.authLoginRejected);
    } on ApiFailure catch (failure) {
      _showNotice(describeAuthFailure(failure, l10n));
    }
  }

  void _showNotice(String? message) {
    if (mounted) setState(() => _notice = message);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final notice = _notice;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.authLoginTitle)),
      body: AutofillGroup(
        child: AppForm(
          controller: fields.form,
          onSave: _login,
          primaryLabel: l10n.authLoginSubmit,
          children: [
            if (notice != null) FormNotice(message: notice),
            AppTextField(
              field: fields.email,
              keyboardType: TextInputType.emailAddress,
              textInputAction: TextInputAction.next,
              autofillHints: const [AutofillHints.email],
            ),
            AppPasswordField(
              field: fields.password,
              textInputAction: TextInputAction.done,
              autofillHints: const [AutofillHints.password],
            ),
            const _CreateAccountLink(),
          ],
        ),
      ),
    );
  }
}

class _CreateAccountLink extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Wrap(
      crossAxisAlignment: WrapCrossAlignment.center,
      spacing: AppSpacing.xs,
      children: [
        Text(
          l10n.authLoginNoAccount,
          style: Theme.of(context).textTheme.bodyMedium,
        ),
        TextButton(
          onPressed: () => context.push<void>(AuthRoutes.register),
          child: Text(l10n.authRegisterTitle),
        ),
      ],
    );
  }
}
