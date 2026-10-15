import 'dart:async';

import 'package:control_contable/core/diagnostics/error_reporter.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';

/// The data key of the reminder's deep link (backend `fcm.dataKeyLink`).
const String _linkKey = 'link';

/// How long boot waits for Firebase before going on without push.
const Duration firebaseStartTimeout = Duration(seconds: 3);

/// Starts Firebase and answers the push provider for this run: the
/// Firebase adapter when it started, else [UnavailablePushMessaging].
///
/// Guarded on purpose (OB9): until the owner creates the Firebase project
/// there is no `google-services.json` / `GoogleService-Info.plist`, so
/// `Firebase.initializeApp()` fails (or, worst case, never answers). The
/// app then boots normally, the reminders screen says push is not
/// available, and no device is registered. The failure is reported by type
/// only. [initialize] is a seam for tests.
Future<PushMessaging> startPushMessaging({
  required ErrorReporter reporter,
  Future<void> Function()? initialize,
  Duration timeout = firebaseStartTimeout,
}) async {
  try {
    await (initialize ?? _initializeFirebase)().timeout(timeout);
    return FirebasePushMessaging(FirebaseMessaging.instance);
  } on Object catch (error) {
    // Any failure (a PlatformException, a FirebaseException, a timeout,
    // a missing plugin) means "no push in this run"; never a crash.
    reporter.report('push.start', error);
    return const UnavailablePushMessaging();
  }
}

Future<void> _initializeFirebase() async {
  await Firebase.initializeApp();
}

/// [PushMessaging] over Firebase Cloud Messaging (APNs goes through FCM).
class FirebasePushMessaging implements PushMessaging {
  /// The adapter over a started [_messaging].
  new(this._messaging);

  final FirebaseMessaging _messaging;

  @override
  bool get isAvailable => true;

  @override
  Future<PushPermission> permission() async => pushPermissionOf(
    (await _messaging.getNotificationSettings()).authorizationStatus,
  );

  @override
  Future<PushPermission> requestPermission() async {
    final settings = await _messaging.requestPermission();
    final permission = pushPermissionOf(settings.authorizationStatus);
    // Auto-init is off in the manifests until the user opts in, so no
    // token exists for someone who never turned reminders on.
    if (permission == PushPermission.granted) {
      await _messaging.setAutoInitEnabled(true);
    }
    return permission;
  }

  @override
  Future<PushToken?> token() async {
    final value = await _messaging.getToken();
    return value == null ? null : PushToken(value);
  }

  @override
  Stream<PushToken> get tokenRefreshes =>
      _messaging.onTokenRefresh.map(PushToken.new);

  @override
  Future<String?> initialLink() async {
    final message = await _messaging.getInitialMessage();
    return message == null ? null : deepLinkOf(message);
  }

  @override
  Stream<String> get openedLinks => FirebaseMessaging.onMessageOpenedApp
      .map(deepLinkOf)
      .where((link) => link != null)
      .cast<String>();
}

/// The app's permission for an FCM [status]. Provisional (iOS quiet
/// delivery) counts as granted: the notification still arrives.
PushPermission pushPermissionOf(AuthorizationStatus status) => switch (status) {
  AuthorizationStatus.authorized ||
  AuthorizationStatus.provisional => PushPermission.granted,
  AuthorizationStatus.denied ||
  AuthorizationStatus.deniedPermanently => PushPermission.denied,
  AuthorizationStatus.notDetermined => PushPermission.notDetermined,
};

/// The deep link a reminder [message] carries, if any.
String? deepLinkOf(RemoteMessage message) => message.data[_linkKey] as String?;
