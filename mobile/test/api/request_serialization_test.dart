import 'dart:convert';

import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/income_update.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
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

  // `BusinessDate` is a calendar date: `YYYY-MM-DD` on the wire, never a
  // date-time (the server and Prism refuse `2026-09-10T00:00:00.000`).
  test('a record date travels as YYYY-MM-DD', () {
    final body = wire(
      const IncomeCreate(
        regimeCode: '612',
        counterpartyKind: CounterpartyKind.pm,
        amountCents: 1000000,
        ivaRate: IvaRate.rate16,
        ivaCents: 160000,
        collectedOn: '2026-09-10',
      ),
    );

    expect(body['collected_on'], '2026-09-10');
    expect(body.containsKey('invoiced_on'), isFalse);
  });

  // A required nullable field keeps `null` on the wire: PUT
  // /v1/reminder-settings replaces the whole object, so the early reminder is
  // turned off with an explicit `early_day: null` and an omission is a 400
  // (contract 0.13.0, MODULE-reminders §10 item 3).
  test('a required nullable field keeps an explicit null', () {
    final body = wire(
      const ReminderSettings(
        enabled: true,
        earlyDay: null,
        onDueDate: true,
        sendTime: '10:00',
        showAmounts: false,
      ),
    );

    expect(body.containsKey('early_day'), isTrue);
    expect(body['early_day'], isNull);
    expect(body.keys, {
      'enabled',
      'early_day',
      'on_due_date',
      'send_time',
      'show_amounts',
    });
  });
}
