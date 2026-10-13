import 'dart:async';

import 'package:control_contable/api/generated/models/legal_document.dart';
import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/links/url_opener.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const double _spinnerSize = 20;
const double _spinnerStroke = 2;

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
          _Status(message: l10n.authDocumentsLoading, busy: true)
        else if (loaded == null)
          _LoadFailed(onRetry: () => ref.invalidate(legalDocumentsProvider)),
        AppCheckboxField(
          field: privacy,
          below: loaded == null
              ? null
              : _DocumentLink(
                  label: l10n.authReadPrivacy,
                  document: loaded.privacyNotice,
                ),
        ),
        AppCheckboxField(
          field: terms,
          below: loaded == null
              ? null
              : _DocumentLink(
                  label: l10n.authReadTerms,
                  document: loaded.terms,
                ),
        ),
      ],
    );
  }
}

class _DocumentLink extends ConsumerWidget {
  const new({required this.label, required this.document});

  final String label;
  final LegalDocument document;

  @override
  Widget build(BuildContext context, WidgetRef ref) => TextButton(
    onPressed: () => unawaited(_open(context, ref)),
    child: Text(label),
  );

  Future<void> _open(BuildContext context, WidgetRef ref) async {
    final failed = AppLocalizations.of(context).authDocumentOpenFailed;
    final messenger = ScaffoldMessenger.of(context);
    final url = Uri.tryParse(document.url);
    final opened = url != null && await ref.read(urlOpenerProvider)(url);
    if (!opened) messenger.showSnackBar(SnackBar(content: Text(failed)));
  }
}

class _LoadFailed extends StatelessWidget {
  const new({required this.onRetry});

  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Row(
      children: [
        Expanded(child: _Status(message: l10n.authDocumentsFailed)),
        TextButton(onPressed: onRetry, child: Text(l10n.authRetry)),
      ],
    );
  }
}

class _Status extends StatelessWidget {
  const new({required this.message, this.busy = false});

  final String message;
  final bool busy;

  @override
  Widget build(BuildContext context) => Semantics(
    liveRegion: true,
    child: Row(
      children: [
        if (busy) ...[
          const SizedBox.square(
            dimension: _spinnerSize,
            child: ExcludeSemantics(
              child: CircularProgressIndicator(strokeWidth: _spinnerStroke),
            ),
          ),
          const SizedBox(width: AppSpacing.xs),
        ],
        Expanded(
          child: Text(message, style: Theme.of(context).textTheme.bodyMedium),
        ),
      ],
    ),
  );
}
