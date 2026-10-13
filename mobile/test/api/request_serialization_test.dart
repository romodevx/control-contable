import 'dart:convert';

import 'package:control_contable/api/generated/models/document_acceptance.dart';
import 'package:control_contable/api/generated/models/legal_document_kind.dart';
import 'package:control_contable/api/generated/models/register_request.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/synthetic_password.dart';

/// The contract's request schemas are `additionalProperties: false`, and an
/// optional non-nullable property rejects `null` (Prism answers 400). An
/// unset optional field must be left out of the body, not sent as `null`.
void main() {
  Map<String, Object?> wire(Object request) =>
      jsonDecode(jsonEncode(request)) as Map<String, Object?>;

  const documents = [
    DocumentAcceptance(
      kind: LegalDocumentKind.privacyNoticeIntegral,
      version: '2026-10-01',
    ),
    DocumentAcceptance(kind: LegalDocumentKind.terms, version: '2026-10-01'),
  ];

  test('an unset optional field is left out of the body', () {
    final body = wire(
      RegisterRequest(
        email: 'ana@example.mx',
        password: syntheticPassword(),
        acceptedDocuments: documents,
      ),
    );

    expect(body.containsKey('name'), isFalse);
    expect(body['accepted_documents'], [
      {'kind': 'privacy_notice_integral', 'version': '2026-10-01'},
      {'kind': 'terms', 'version': '2026-10-01'},
    ]);
  });

  test('a set optional field is sent', () {
    final body = wire(
      RegisterRequest(
        email: 'ana@example.mx',
        password: syntheticPassword(),
        acceptedDocuments: documents,
        name: 'Ana',
      ),
    );

    expect(body['name'], 'Ana');
  });
}
