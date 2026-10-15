import 'dart:math';

import 'package:control_contable/api/generated/models/tax_calendar_entry.dart';
import 'package:control_contable/api/generated/models/tax_calendar_moved_reason.dart';
import 'package:control_contable/api/generated/models/tax_obligation.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:control_contable/features/reminders/domain/reminder_deep_link.dart';
import 'package:control_contable/features/reminders/domain/upcoming_due_dates.dart';
import 'package:control_contable/features/reminders/domain/uuid_v7.dart';
import 'package:flutter_test/flutter_test.dart';

TaxCalendarEntry _entry(String period, String due, {bool moved = false}) =>
    TaxCalendarEntry(
      period: period,
      obligation: TaxObligation.monthlyIsrIva,
      nominalDueDate: due,
      dueDate: due,
      movedReason: moved ? TaxCalendarMovedReason.weekend : null,
      subjectToConfirmation: true,
    );

void main() {
  group('PushToken', () {
    const secret = 'fcm-token-under-test';

    test('never prints its value', () {
      const token = PushToken(secret);

      expect('$token', isNot(contains(secret)));
      expect(token.toString(), contains('redacted'));
      expect(token.value, secret);
    });

    test('compares by value', () {
      expect(const PushToken(secret), const PushToken(secret));
      expect(const PushToken(secret), isNot(const PushToken('other')));
    });
  });

  group('newUuidV7', () {
    final shape = RegExp(
      r'^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$',
    );

    test('is a version 7, RFC 9562 variant UUID', () {
      final id = newUuidV7(DateTime.utc(2026, 10, 15), Random(1));

      expect(id, matches(shape));
    });

    test('starts with the 48-bit Unix time in milliseconds', () {
      final now = DateTime.utc(2026, 10, 15, 12);
      final id = newUuidV7(now, Random(1));

      final millis = int.parse(
        id.replaceAll('-', '').substring(0, 12),
        radix: 16,
      );
      expect(millis, now.millisecondsSinceEpoch);
    });

    test('later instants sort later', () {
      final first = newUuidV7(DateTime.utc(2026), Random(9));
      final second = newUuidV7(DateTime.utc(2026, 1, 2), Random(1));

      expect(first.compareTo(second), isNegative);
    });
  });

  group('estimatePeriodOfLink', () {
    for (final (link, period) in [
      ('controlcontable://estimate?period=2026-04', '2026-04'),
      ('controlcontable://estimate?period=2026-4', null),
      ('controlcontable://estimate', null),
      ('controlcontable://history?period=2026-04', null),
      ('https://estimate?period=2026-04', null),
      ('not a link', null),
    ]) {
      test('"$link" → $period', () {
        expect(estimatePeriodOfLink(link), period);
      });
    }
  });

  group('upcomingDueDates', () {
    final entries = [
      _entry('2026-03', '2026-04-17'),
      _entry('2026-04', '2026-05-18', moved: true),
      _entry('2026-05', '2026-06-17'),
      _entry('2026-06', '2026-07-17'),
      _entry('2026-07', '2026-08-17'),
    ];

    test('the next three from today, today included', () {
      final next = upcomingDueDates(entries, today: DateTime(2026, 5, 18));

      expect(next.map((e) => e.period), ['2026-04', '2026-05', '2026-06']);
    });

    test('fewer when the calendar runs out', () {
      final next = upcomingDueDates(entries, today: DateTime(2026, 7, 20));

      expect(next.map((e) => e.period), ['2026-07']);
    });
  });
}
