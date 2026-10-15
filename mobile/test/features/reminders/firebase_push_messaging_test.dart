import 'dart:async';

import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/firebase_push_messaging.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_error_reporter.dart';

void main() {
  group('startPushMessaging (OB9: no Firebase config yet)', () {
    test(
      'a failed Firebase start is "not available", reported by type',
      () async {
        final reporter = RecordingErrorReporter();

        final messaging = await startPushMessaging(
          initialize: () async => throw PlatformException(
            code: 'not-initialized',
            message: 'Failed to load FirebaseOptions from resource.',
          ),
          reporter: reporter,
        );

        expect(messaging, isA<UnavailablePushMessaging>());
        expect(messaging.isAvailable, isFalse);
        expect(reporter.reports.single.operation, 'push.start');
      },
    );

    test('a Firebase start that hangs gives up: the app still boots', () async {
      final messaging = await startPushMessaging(
        initialize: () => Completer<void>().future,
        reporter: RecordingErrorReporter(),
        timeout: Duration.zero,
      );

      expect(messaging.isAvailable, isFalse);
    });
  });

  group('pushPermissionOf', () {
    for (final (status, permission) in [
      (AuthorizationStatus.authorized, PushPermission.granted),
      (AuthorizationStatus.provisional, PushPermission.granted),
      (AuthorizationStatus.denied, PushPermission.denied),
      (AuthorizationStatus.deniedPermanently, PushPermission.denied),
      (AuthorizationStatus.notDetermined, PushPermission.notDetermined),
    ]) {
      test('$status → $permission', () {
        expect(pushPermissionOf(status), permission);
      });
    }
  });

  group('deepLinkOf', () {
    test('reads the "link" data key the reminder job sends', () {
      const message = RemoteMessage(
        data: {'link': 'controlcontable://estimate?period=2026-04'},
      );

      expect(deepLinkOf(message), 'controlcontable://estimate?period=2026-04');
    });

    test('no link: null', () {
      expect(deepLinkOf(const RemoteMessage()), isNull);
    });
  });
}
