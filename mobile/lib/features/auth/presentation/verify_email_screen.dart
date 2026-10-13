import 'dart:async';

import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/application/resend_cooldown.dart';
import 'package:control_contable/features/auth/application/session_store.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/features/auth/presentation/resend_code.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Digits in an emailed verification code (contract: `^[0-9]{6}$`).
const int _codeLength = 6;

final RegExp _codeShape = RegExp('^[0-9]{$_codeLength}\$');

/// Email verification (US-A1b, US-A1c): the 6-digit code, with one-time
/// code autofill, submitted as soon as it is complete; a resend after the
/// cooldown. Every rejection reads the same (no account enumeration).
class VerifyEmailScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final email = ref.watch(pendingSignUpProvider);
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.authVerifyTitle)),
      body: email == null ? const _NoSignUp() : _CodeForm(email: email),
    );
  }
}

class _CodeForm extends ConsumerStatefulWidget {
  const new({required this.email});

  final String email;

  @override
  ConsumerState<_CodeForm> createState() => _CodeFormState();
}

class _CodeFormState extends ConsumerState<_CodeForm> {
  final _form = AppFormController();
  AppFieldController? _code;
  late final _cooldown = ResendCooldown(ref.read(clockProvider))..start();
  ({String message, FormNoticeTone tone})? _notice;
  bool _isResending = false;

  AppFieldController get code => _code!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final l10n = AppLocalizations.of(context);
    _code ??= _form.textField(
      label: l10n.authCodeLabel,
      validator: (value) => _codeShape.hasMatch(value)
          ? null
          : CustomFieldError(l10n.authCodeFormat),
    );
  }

  @override
  void dispose() {
    _cooldown.dispose();
    _form.dispose();
    super.dispose();
  }

  void _submitWhenComplete(String text) {
    if (text.length == _codeLength) unawaited(_form.submit(_verify));
  }

  Future<void> _verify() async {
    final l10n = AppLocalizations.of(context);
    _setNotice(null);
    try {
      final session = await ref
          .read(authRepositoryProvider)
          .verifyEmail(
            VerifyEmailRequest(email: widget.email, code: code.value),
          );
      await ref.read(sessionStoreProvider).save(session);
    } on ApiUnauthorized {
      return _form.showServerErrors({
        code: CustomFieldError(l10n.authCodeRejected),
      });
    } on ApiValidationFailure {
      return _form.showServerErrors({
        code: CustomFieldError(l10n.authCodeFormat),
      });
    } on ApiFailure catch (failure) {
      return _setNotice(describeAuthFailure(failure, l10n));
    }
    if (!mounted) return;
    final pending = ref.read(pendingSignUpProvider.notifier);
    context.go(AppRoutes.home);
    pending.clear();
  }

  Future<void> _resend() async {
    if (_isResending) return;
    _isResending = true;
    final l10n = AppLocalizations.of(context);
    try {
      await ref.read(authRepositoryProvider).resendVerification(widget.email);
      _cooldown.start();
      _setNotice(l10n.authResendSent, FormNoticeTone.info);
    } on ApiRateLimited catch (failure) {
      _cooldown.start(failure.retryAfter);
      _setNotice(describeAuthFailure(failure, l10n));
    } on ApiFailure catch (failure) {
      _setNotice(describeAuthFailure(failure, l10n));
    } finally {
      _isResending = false;
    }
  }

  void _setNotice(
    String? message, [
    FormNoticeTone tone = FormNoticeTone.error,
  ]) {
    if (!mounted) return;
    setState(
      () => _notice = message == null ? null : (message: message, tone: tone),
    );
  }

  void _changeEmail() {
    if (context.canPop()) return context.pop();
    context.go(AuthRoutes.register);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final textTheme = Theme.of(context).textTheme;
    final notice = _notice;
    return AppForm(
      controller: _form,
      onSave: _verify,
      primaryLabel: l10n.authVerifySubmit,
      children: [
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.authVerifyIntro(widget.email),
              style: textTheme.bodyLarge,
            ),
            const SizedBox(height: AppSpacing.xs),
            Text(l10n.authVerifySpamHint, style: textTheme.bodyMedium),
          ],
        ),
        if (notice != null)
          FormNotice(message: notice.message, tone: notice.tone),
        AppTextField(
          field: code,
          keyboardType: TextInputType.number,
          textInputAction: TextInputAction.done,
          autofillHints: const [AutofillHints.oneTimeCode],
          inputFormatters: [
            FilteringTextInputFormatter.digitsOnly,
            LengthLimitingTextInputFormatter(_codeLength),
          ],
          onChanged: _submitWhenComplete,
        ),
        ResendCode(cooldown: _cooldown, onResend: () => unawaited(_resend())),
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton(
            onPressed: _changeEmail,
            child: Text(l10n.authChangeEmail),
          ),
        ),
      ],
    );
  }
}

class _NoSignUp extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            l10n.authVerifyNoPending,
            style: Theme.of(context).textTheme.bodyLarge,
          ),
          const SizedBox(height: AppSpacing.lg),
          FilledButton(
            onPressed: () => context.go(AuthRoutes.register),
            child: Text(l10n.authRegisterTitle),
          ),
        ],
      ),
    );
  }
}
