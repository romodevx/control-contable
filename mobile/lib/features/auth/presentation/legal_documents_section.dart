import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/features/auth/presentation/legal_document_link.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// *Documentos legales* on the account screen: links to the current privacy
/// notice and terms (store requirement: reachable from inside the app).
class LegalDocumentsSection extends ConsumerWidget {
  /// Creates the section.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final documents = ref.watch(legalDocumentsProvider);
    final loaded = documents.value;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Semantics(
          header: true,
          child: Text(
            l10n.authConsentsHeading,
            style: Theme.of(context).textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        if (loaded != null) ...[
          LegalDocumentLink(
            label: l10n.authReadPrivacy,
            document: loaded.privacyNotice,
          ),
          LegalDocumentLink(label: l10n.authReadTerms, document: loaded.terms),
        ] else if (documents.isLoading)
          LoadStatus(message: l10n.authDocumentsLoading, busy: true)
        else
          LoadFailed(
            message: l10n.authDocumentsFailed,
            onRetry: () => ref.invalidate(legalDocumentsProvider),
          ),
      ],
    );
  }
}
