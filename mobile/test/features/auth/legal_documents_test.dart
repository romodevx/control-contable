import 'package:control_contable/api/generated/models/legal_document_kind.dart';
import 'package:control_contable/features/auth/application/legal_documents.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'auth_fakes.dart';

void main() {
  group('LegalDocumentSet.pick', () {
    test('prefers the full privacy notice and takes the terms', () {
      final set = LegalDocumentSet.pick(currentDocuments)!;

      expect(set.privacyNotice.kind, LegalDocumentKind.privacyNoticeIntegral);
      expect(set.terms.kind, LegalDocumentKind.terms);
    });

    test('falls back to the simplified notice', () {
      final set = LegalDocumentSet.pick([
        for (final document in currentDocuments)
          if (document.kind != LegalDocumentKind.privacyNoticeIntegral)
            document,
      ])!;

      expect(set.privacyNotice.kind, LegalDocumentKind.privacyNoticeSimplified);
    });

    test('is null without terms or without any privacy notice', () {
      expect(
        LegalDocumentSet.pick([
          for (final document in currentDocuments)
            if (document.kind != LegalDocumentKind.terms) document,
        ]),
        isNull,
      );
      expect(
        LegalDocumentSet.pick([
          for (final document in currentDocuments)
            if (document.kind == LegalDocumentKind.terms) document,
        ]),
        isNull,
      );
    });

    test('accepts exactly the versions shown', () {
      final acceptances = LegalDocumentSet.pick(currentDocuments)!.acceptances;

      expect(acceptances.map((a) => (a.kind, a.version)), [
        (LegalDocumentKind.privacyNoticeIntegral, '2026-10-01'),
        (LegalDocumentKind.terms, '2026-10-01'),
      ]);
    });
  });

  group('legalDocumentsProvider', () {
    test('loads the current set from the repository', () async {
      final container = ProviderContainer.test(
        overrides: [
          authRepositoryProvider.overrideWithValue(FakeAuthRepository()),
        ],
      );

      final set = await container.read(legalDocumentsProvider.future);

      expect(set.terms.url, 'https://example.mx/legal/terminos');
    });

    test('fails when a required kind is missing', () async {
      final repository = FakeAuthRepository()..documents = const [];
      final container = ProviderContainer.test(
        overrides: [authRepositoryProvider.overrideWithValue(repository)],
      );
      final subscription = container.listen(legalDocumentsProvider, (_, _) {});
      addTearDown(subscription.close);

      await expectLater(
        container.read(legalDocumentsProvider.future),
        throwsA(isA<LegalDocumentsIncomplete>()),
      );
    });
  });
}
