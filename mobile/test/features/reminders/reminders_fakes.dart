import 'dart:async';

import 'package:control_contable/api/generated/models/device_platform.dart';
import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/api/generated/models/tax_calendar.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';

import 'reminders_fixtures.dart';

/// One recorded device registration (the token kept as the value object).
typedef RegisterCall = ({
  String installationId,
  DevicePlatform platform,
  PushToken token,
});

/// A [RemindersRepository] the test scripts and observes.
class FakeRemindersRepository implements RemindersRepository {
  /// What the server holds; GET answers it, PUT replaces it.
  ReminderSettings settings = defaultSettings;

  /// Every registration, in order.
  final List<RegisterCall> registered = [];

  /// Every unregistered installation id, in order.
  final List<String> unregistered = [];

  /// Every PUT body, in order.
  final List<ReminderSettings> puts = [];

  /// Every calendar year asked, in order.
  final List<int> calendarYears = [];

  /// Calendars by year; a missing year throws [calendarError].
  final Map<int, TaxCalendar> calendars = {2026: taxCalendar(2026)};

  /// What a GET of a year not in [calendars] throws.
  Exception? calendarError;

  /// Thrown by the next GET of the settings, then cleared.
  Exception? getError;

  /// Thrown (or awaited, for a [Completer]) by the next PUT, then cleared.
  Object? putReply;

  /// Thrown by the next registration, then cleared.
  Exception? registerError;

  /// Thrown by the next unregistration, then cleared.
  Exception? unregisterError;

  @override
  Future<ReminderSettings> getSettings() async {
    if (getError case final Exception error) {
      getError = null;
      throw error;
    }
    return settings;
  }

  @override
  Future<ReminderSettings> putSettings(ReminderSettings next) async {
    puts.add(next);
    final reply = putReply;
    putReply = null;
    if (reply is Completer<void>) await reply.future;
    if (reply is Exception) throw reply;
    return settings = next;
  }

  @override
  Future<TaxCalendar> getCalendar(int year) async {
    calendarYears.add(year);
    final calendar = calendars[year];
    if (calendar != null) return calendar;
    if (calendarError case final Exception error) throw error;
    throw StateError('no calendar for $year');
  }

  @override
  Future<void> registerDevice({
    required String installationId,
    required DevicePlatform platform,
    required PushToken token,
    String? appVersion,
  }) async {
    if (registerError case final Exception error) {
      registerError = null;
      throw error;
    }
    registered.add((
      installationId: installationId,
      platform: platform,
      token: token,
    ));
  }

  @override
  Future<void> unregisterDevice(String installationId) async {
    unregistered.add(installationId);
    if (unregisterError case final Exception error) {
      unregisterError = null;
      throw error;
    }
  }
}

/// A [PushMessaging] the test drives: availability, the OS answer to the
/// permission prompt, the token, refreshes and notification taps.
class FakePushMessaging implements PushMessaging {
  /// Whether Firebase started.
  @override
  bool isAvailable = true;

  /// The permission before any prompt.
  PushPermission status = PushPermission.notDetermined;

  /// What the user answers to the OS prompt.
  PushPermission answer = PushPermission.granted;

  /// How many times the OS prompt was shown.
  int prompts = 0;

  /// The current token.
  PushToken? currentToken = const PushToken('fcm-token-1');

  /// The link of the notification that cold-started the app.
  String? launchLink;

  final StreamController<PushToken> _refreshes = StreamController.broadcast();
  final StreamController<String> _taps = StreamController.broadcast();

  /// FCM rotates the token to [token].
  void rotate(PushToken token) {
    currentToken = token;
    _refreshes.add(token);
  }

  /// The user taps a notification carrying [link].
  void tap(String link) => _taps.add(link);

  @override
  Future<PushPermission> permission() async => status;

  @override
  Future<PushPermission> requestPermission() async {
    if (status != PushPermission.notDetermined) return status;
    prompts++;
    return status = answer;
  }

  @override
  Future<PushToken?> token() async => currentToken;

  @override
  Stream<PushToken> get tokenRefreshes => _refreshes.stream;

  @override
  Future<String?> initialLink() async => launchLink;

  @override
  Stream<String> get openedLinks => _taps.stream;
}
