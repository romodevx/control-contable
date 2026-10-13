import 'dart:async';

import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
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
import 'package:control_contable/features/auth/presentation/code_step_fields.dart';
import 'package:control_contable/features/auth/presentation/resend_code.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// The code step for the sign-up of [email]: the code (one-time-code
/// autofill, sent as soon as it is complete), the password field when it
/// must be typed, the resend countdown and a way back to the form.
class CodeStepForm extends ConsumerStatefulWidget {
  /// The code step for [email].
  const new({required this.email, super.key});

  /// Where the code was sent.
  final String email;

  @override
  ConsumerState<CodeStepForm> createState() => _CodeStepFormState();
}

class _CodeStepFormState extends ConsumerState<CodeStepForm> {
  CodeStepFields? _fields;
  late final _cooldown = ResendCooldown(ref.read(clockProvider))..start();
  ({String message, FormNoticeTone tone})? _notice;
  bool _isResending = false;
  bool _asksPassword = false;

  CodeStepFields get fields => _fields!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_fields != null) return;
    final draft = ref.read(pendingSignUpProvider)!;
    _fields = CodeStepFields(AppLocalizations.of(context), draft);
    _asksPassword = draft.password == null;
  }

  @override
  void dispose() {
    _cooldown.dispose();
    _fields?.dispose();
    super.dispose();
  }

  void _submitWhenComplete(String text) {
    ref.read(pendingSignUpProvider.notifier).touch();
    if (text.length == codeLength && fields.isReadyToSend) {
      unawaited(fields.form.submit(_verify));
    }
  }

  Future<void> _verify() async {
    final l10n = AppLocalizations.of(context);
    final pending = ref.read(pendingSignUpProvider.notifier);
    _setNotice(null);
    try {
      final draft = ref.read(pendingSignUpProvider)!;
      final session = await ref
          .read(authRepositoryProvider)
          .verifyEmail(fields.request(draft));
      await ref.read(sessionStoreProvider).save(session);
    } on ApiUnauthorized {
      return fields.showCodeRejected();
    } on ApiValidationFailure catch (failure) {
      final rejected = fields.showRejected(failure.details);
      if (rejected.password) setState(() => _asksPassword = true);
      if (rejected.unmatched) _setNotice(l10n.authFormRejected);
      return;
    } on ApiFailure catch (failure) {
      return _setNotice(describeAuthFailure(failure, l10n));
    }
    if (!mounted) return;
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

  void _handleDraftChanged(SignUpDraft? previous, SignUpDraft? next) {
    if (next == null || next.password != null) return;
    fields.forgetPassword();
    setState(() => _asksPassword = true);
  }

  void _changeEmail() {
    if (context.canPop()) return context.pop();
    ref.read(pendingSignUpProvider.notifier).clear();
    context.go(AuthRoutes.register);
  }

  @override
  Widget build(BuildContext context) {
    ref.listen<SignUpDraft?>(pendingSignUpProvider, _handleDraftChanged);
    final l10n = AppLocalizations.of(context);
    final textTheme = Theme.of(context).textTheme;
    final notice = _notice;
    return AppForm(
      controller: fields.form,
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
          field: fields.code,
          keyboardType: TextInputType.number,
          textInputAction: TextInputAction.done,
          autofillHints: const [AutofillHints.oneTimeCode],
          inputFormatters: [
            FilteringTextInputFormatter.digitsOnly,
            LengthLimitingTextInputFormatter(codeLength),
          ],
          onChanged: _submitWhenComplete,
        ),
        if (_asksPassword)
          AppPasswordField(
            field: fields.password,
            textInputAction: TextInputAction.done,
            autofillHints: const [AutofillHints.newPassword],
            helperText: l10n.authPasswordHelp,
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
