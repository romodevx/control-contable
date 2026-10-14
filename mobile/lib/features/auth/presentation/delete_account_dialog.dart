import 'dart:async';

import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_error_summary.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/form_scope.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/time/countdown.dart';
import 'package:control_contable/features/auth/application/delete_account.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/features/auth/presentation/delete_busy_label.dart';
import 'package:control_contable/features/auth/presentation/retry_wait.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Opens the account-deletion confirmation over [context].
Future<void> showDeleteAccountDialog(BuildContext context) => showDialog<void>(
  context: context,
  builder: (_) => const DeleteAccountDialog(),
);

/// The confirmation of the one irreversible action in the app (house UI
/// rules: a confirmation dialog only for the irreversible). A modal route
/// (inert background, Esc and the barrier close it, focus returns to the
/// trigger), labeled by its title. It says what is lost, that the free
/// trial does not restart and that a store subscription is cancelled in the
/// store, and asks for the current password (D4).
///
/// The destructive primary is never disabled: while deleting it shows a
/// busy state and repeat taps are ignored; the dialog cannot be closed
/// then. A wrong password (`403 invalid_password`) clears the field, says
/// so on it, announces it and focuses it. Too many tries (`429`) count the
/// `Retry-After` down; a tap during the wait sends nothing and announces
/// how long is left.
/// On success the session is gone and the goodbye screen shows.
class DeleteAccountDialog extends ConsumerStatefulWidget {
  /// Creates the dialog.
  const new({super.key});

  @override
  ConsumerState<DeleteAccountDialog> createState() =>
      _DeleteAccountDialogState();
}

class _DeleteAccountDialogState extends ConsumerState<DeleteAccountDialog> {
  final AppFormController _form = AppFormController();
  AppFieldController? _passwordField;
  late final Countdown _retryWait = Countdown(ref.read(clockProvider));
  String? _notice;

  AppFieldController get _password => _passwordField!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _passwordField ??= _form.textField(
      label: AppLocalizations.of(context).accountDeletePasswordLabel,
      verbatim: true,
    );
  }

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _password.focusNode.requestFocus();
    });
  }

  @override
  void dispose() {
    _retryWait.dispose();
    _form.dispose();
    super.dispose();
  }

  void _submit() {
    if (_retryWait.isRunning) return _announceWait();
    unawaited(_form.submit(_delete));
  }

  Future<void> _delete() async {
    final l10n = AppLocalizations.of(context);
    final router = GoRouter.of(context);
    final deleteAccount = ref.read(deleteAccountProvider);
    _showNotice(null);
    try {
      await deleteAccount(_password.value);
    } on ApiInvalidPassword {
      _rejectPassword(l10n);
      return;
    } on ApiRateLimited catch (failure) {
      if (failure.retryAfter case final wait?) _retryWait.run(wait);
      _showNotice(describeAuthFailure(failure, l10n));
      return;
    } on ApiFailure catch (failure) {
      _showNotice(describeAuthFailure(failure, l10n));
      return;
    }
    // The session is gone: this dialog and the screen under it may already
    // be unmounted by the router's guard, so navigate through the router.
    router.go(AuthRoutes.goodbye);
  }

  /// Focus alone is silent for TalkBack and VoiceOver (it moves input
  /// focus, not the screen reader's), and an unchanged error is not read
  /// again: the verdict is announced on every wrong try.
  void _rejectPassword(AppLocalizations l10n) {
    if (!mounted) return;
    final message = l10n.accountDeleteWrongPassword;
    _password
      ..text.clear()
      ..showError(CustomFieldError(message));
    _password.focusNode.requestFocus();
    _announce(message);
  }

  void _announceWait() => announceRetryWait(context, _retryWait);

  void _announce(String message) => unawaited(
    SemanticsService.sendAnnouncement(
      View.of(context),
      message,
      Directionality.of(context),
      assertiveness: Assertiveness.assertive,
    ),
  );

  void _showNotice(String? message) {
    if (mounted) setState(() => _notice = message);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final body = Theme.of(context).textTheme.bodyMedium;
    final notice = _notice;
    return ListenableBuilder(
      listenable: _form,
      builder: (context, child) =>
          PopScope(canPop: !_form.isSaving, child: child!),
      child: AppFormScope(
        controller: _form,
        submit: _submit,
        child: AlertDialog(
          title: Text(l10n.accountDeleteTitle),
          scrollable: true,
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            spacing: AppSpacing.md,
            children: [
              FormErrorSummary(controller: _form),
              if (notice != null) FormNotice(message: notice),
              RetryWait(countdown: _retryWait),
              Text(l10n.accountDeleteWhat, style: body),
              Text(l10n.accountDeleteTrial, style: body),
              Text(l10n.accountDeleteSubscription, style: body),
              AppPasswordField(
                field: _password,
                textInputAction: TextInputAction.done,
                autofillHints: const [AutofillHints.password],
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).maybePop(),
              child: Text(l10n.formCancel),
            ),
            _ConfirmButton(form: _form, onPressed: _submit),
          ],
        ),
      ),
    );
  }
}

class _ConfirmButton extends StatelessWidget {
  const new({required this.form, required this.onPressed});

  final AppFormController form;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final scheme = Theme.of(context).colorScheme;
    return ListenableBuilder(
      listenable: form,
      builder: (context, _) => FilledButton(
        style: FilledButton.styleFrom(
          backgroundColor: scheme.error,
          foregroundColor: scheme.onError,
        ),
        onPressed: onPressed,
        child: form.isSaving
            ? DeleteBusyLabel(
                label: l10n.accountDeleteConfirm,
                announcement: l10n.accountDeleting,
              )
            : Text(l10n.accountDeleteConfirm),
      ),
    );
  }
}
