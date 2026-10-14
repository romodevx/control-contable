import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/features/auth/presentation/legal_document_link.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The two required consents, each with a link to the current version of
/// its document. The boxes are always there (the error summary can point
/// at them); the links appear once the documents load, with a retry when
/// they fail.
class ConsentsSection extends ConsumerWidget {
  /// Consents bound to [privacy] and [terms].
  const new({required this.privacy, required this.terms, super.key});

  /// Consent to the privacy notice.
  final AppCheckboxController privacy;

  /// Consent to the terms and conditions.
  final AppCheckboxController terms;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final documents = ref.watch(legalDocumentsProvider);
    final loaded = documents.value;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(
            l10n.authConsentsHeading,
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        if (documents.isLoading && loaded == null)
          LoadStatus(message: l10n.authDocumentsLoading, busy: true)
        else if (loaded == null)
          LoadFailed(
            message: l10n.authDocumentsFailed,
            onRetry: () => ref.invalidate(legalDocumentsProvider),
          ),
        AppCheckboxField(
          field: privacy,
          below: loaded == null
              ? null
              : LegalDocumentLink(
                  label: l10n.authReadPrivacy,
                  document: loaded.privacyNotice,
                ),
        ),
        AppCheckboxField(
          field: terms,
          below: loaded == null
              ? null
              : LegalDocumentLink(
                  label: l10n.authReadTerms,
                  document: loaded.terms,
                ),
        ),
      ],
    );
  }
}
