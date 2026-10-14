import 'dart:async';

import 'package:control_contable/api/generated/models/legal_document.dart';
import 'package:control_contable/core/links/url_opener.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// A text button that opens the current version of [document] in the
/// in-app browser view; a snack bar says so when it cannot be opened.
class LegalDocumentLink extends ConsumerWidget {
  /// A link reading [label] to [document].
  const new({required this.label, required this.document, super.key});

  /// What the button says (*Leer el aviso de privacidad*).
  final String label;

  /// The document it opens.
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
