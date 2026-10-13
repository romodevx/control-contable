import 'dart:async';

import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/features/auth/application/pending_sign_up.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/features/auth/presentation/consents_section.dart';
import 'package:control_contable/features/auth/presentation/register_fields.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Sign-up (US-A1a): name (optional), email, password and the two required
/// consents, on its own screen (≤ 4 fields; ui-forms-and-modals.md). On
/// the server's generic 202 it opens the code screen; the answer is the
/// same whether or not the address already has an account.
class RegisterScreen extends ConsumerStatefulWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  RegisterFields? _fields;
  String? _notice;

  RegisterFields get fields => _fields!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Labels and validators need the localized copy, not there in initState.
    _fields ??= RegisterFields(AppLocalizations.of(context));
  }

  @override
  void dispose() {
    _fields?.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    final l10n = AppLocalizations.of(context);
    final documents = ref.read(legalDocumentsProvider).value;
    if (documents == null) return _showNotice(l10n.authDocumentsNotLoaded);
    _showNotice(null);
    try {
      await ref
          .read(authRepositoryProvider)
          .register(fields.request(documents));
    } on ApiValidationFailure catch (failure) {
      final rejected = fields.showRejected(failure.details);
      if (rejected.staleDocuments) ref.invalidate(legalDocumentsProvider);
      if (rejected.unmatched) _showNotice(l10n.authFormRejected);
      return;
    } on ApiFailure catch (failure) {
      return _showNotice(describeAuthFailure(failure, l10n));
    }
    if (!mounted) return;
    ref.read(pendingSignUpProvider.notifier).start(fields.email.value);
    unawaited(context.push<void>(AuthRoutes.verifyEmail));
  }

  void _showNotice(String? message) {
    if (mounted) setState(() => _notice = message);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final notice = _notice;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.authRegisterTitle)),
      body: AutofillGroup(
        child: AppForm(
          controller: fields.form,
          onSave: _register,
          primaryLabel: l10n.authRegisterSubmit,
          children: [
            Text(
              l10n.authRegisterIntro,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            if (notice != null) FormNotice(message: notice),
            AppTextField(
              field: fields.name,
              keyboardType: TextInputType.name,
              textInputAction: TextInputAction.next,
              autofillHints: const [AutofillHints.name],
            ),
            AppTextField(
              field: fields.email,
              keyboardType: TextInputType.emailAddress,
              textInputAction: TextInputAction.next,
              autofillHints: const [AutofillHints.email],
            ),
            AppPasswordField(
              field: fields.password,
              textInputAction: TextInputAction.done,
              autofillHints: const [AutofillHints.newPassword],
              helperText: l10n.authPasswordHelp,
            ),
            ConsentsSection(privacy: fields.privacy, terms: fields.terms),
          ],
        ),
      ),
    );
  }
}
