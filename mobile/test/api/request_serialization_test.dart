import 'dart:convert';

import 'package:control_contable/api/generated/models/income_update.dart';
import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/api/generated/models/verify_email_request.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/synthetic_password.dart';

/// The contract's request schemas are `additionalProperties: false`, and an
/// optional non-nullable property rejects `null` (Prism answers 400). An
/// unset optional field must be left out of the body, not sent as `null`.
void main() {
  Map<String, Object?> wire(Object request) =>
      jsonDecode(jsonEncode(request)) as Map<String, Object?>;

  test('an unset optional field is left out of the body', () {
    final body = wire(
      VerifyEmailRequest(
        email: 'ana@example.mx',
        code: '482913',
        password: syntheticPassword(),
      ),
    );

    expect(body.containsKey('name'), isFalse);
    expect(body.keys, {'email', 'code', 'password'});
  });

  test('a set optional field is sent', () {
    final body = wire(
      VerifyEmailRequest(
        email: 'ana@example.mx',
        code: '482913',
        password: syntheticPassword(),
        name: 'Ana',
      ),
    );

    expect(body['name'], 'Ana');
  });

  // ADR-0003 "Null on the wire": PATCH bodies are partial, so an unset
  // field is absent (a `null` would be read as a value).
  test('a PATCH body carries only the fields that were set', () {
    final body = wire(const IncomeUpdate(description: 'Diseño de logotipo'));

    expect(body, {'description': 'Diseño de logotipo'});
  });

  // A nullable field with a contract default is generated as required, so
  // `null` stays on the wire: PUT /v1/reminder-settings turns the early
  // reminder off with `early_day: null`.
  test('a nullable field with a default keeps an explicit null', () {
    final body = wire(const ReminderSettings(earlyDay: null));

    expect(body.containsKey('early_day'), isTrue);
    expect(body['early_day'], isNull);
  });
}
