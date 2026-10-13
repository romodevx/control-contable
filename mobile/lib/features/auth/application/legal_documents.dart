import 'package:control_contable/api/generated/models/document_acceptance.dart';
import 'package:control_contable/api/generated/models/legal_document.dart';
import 'package:control_contable/api/generated/models/legal_document_kind.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Privacy notices in the order the sign-up prefers them: the full notice
/// first, the simplified one when that is all the server publishes.
const List<LegalDocumentKind> _privacyNoticePreference = [
  LegalDocumentKind.privacyNoticeIntegral,
  LegalDocumentKind.privacyNoticeSimplified,
];

/// The documents a person accepts to sign up: one privacy notice and the
/// terms, at the versions the server says are current (US-A5).
@immutable
class LegalDocumentSet {
  /// The set of [privacyNotice] and [terms].
  const new({required this.privacyNotice, required this.terms});

  /// The privacy notice shown and accepted.
  final LegalDocument privacyNotice;

  /// The terms and conditions shown and accepted.
  final LegalDocument terms;

  /// The set out of the server's [documents], or `null` when the terms or
  /// every privacy notice is missing (nothing valid could be accepted).
  static LegalDocumentSet? pick(List<LegalDocument> documents) {
    final byKind = {for (final document in documents) document.kind: document};
    final terms = byKind[LegalDocumentKind.terms];
    final privacyNotice = _privacyNoticePreference
        .map((kind) => byKind[kind])
        .nonNulls
        .firstOrNull;
    if (terms == null || privacyNotice == null) return null;
    return LegalDocumentSet(privacyNotice: privacyNotice, terms: terms);
  }

  /// What `register` sends back: exactly the versions shown.
  List<DocumentAcceptance> get acceptances => [
    for (final document in [privacyNotice, terms])
      DocumentAcceptance(kind: document.kind, version: document.version),
  ];
}

/// The server published no acceptable set (see [LegalDocumentSet.pick]).
class LegalDocumentsIncomplete implements Exception {
  /// The failure.
  const new();
}

/// The current documents to accept. Re-read with `ref.invalidate` when the
/// server says a version is stale.
final FutureProvider<LegalDocumentSet> legalDocumentsProvider =
    FutureProvider.autoDispose<LegalDocumentSet>(
      (ref) async {
        final documents = await ref
            .watch(authRepositoryProvider)
            .listLegalDocuments();
        return LegalDocumentSet.pick(documents) ??
            (throw const LegalDocumentsIncomplete());
      },
      name: 'legalDocumentsProvider',
      retry: _noAutomaticRetry,
    );

/// A failed load shows a retry button: the person decides when to retry.
Duration? _noAutomaticRetry(int retryCount, Object error) => null;
