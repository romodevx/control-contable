import 'dart:convert';

import 'package:control_contable/api/generated/api_client.dart';
import 'package:control_contable/api/generated/models/device_platform.dart';
import 'package:control_contable/api/generated/models/reminder_settings.dart';
import 'package:control_contable/core/diagnostics/error_reporter.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/reminders/data/reminders_repository.dart';
import 'package:control_contable/features/reminders/domain/push_token.dart';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';
import 'reminders_fixtures.dart';

const String _installationId = '01920000-0000-7000-8000-000000000001';
const String _tokenValue = 'fcm-token-under-test';

void main() {
  late RecordingAdapter adapter;
  late ApiRemindersRepository repository;

  setUp(() {
    adapter = RecordingAdapter(
      replies: {
        '/v1/devices': (status: 201, body: deviceJson(_installationId)),
        '/v1/devices/$_installationId': (status: 204, body: null),
        '/v1/reminder-settings': (status: 200, body: defaultSettingsJson),
        '/v1/tax-calendar': (status: 200, body: taxCalendarJson(2026)),
      },
    );
    final dio = Dio(BaseOptions(baseUrl: 'http://localhost:3001'))
      ..httpClientAdapter = adapter;
    repository = ApiRemindersRepository(ApiClient(dio).reminders);
  });

  test(
    'registers the installation with its token as the closed body',
    () async {
      await repository.registerDevice(
        installationId: _installationId,
        platform: DevicePlatform.android,
        token: const PushToken(_tokenValue),
        appVersion: '0.1.0',
      );

      final request = adapter.requests.single;
      expect(request.method, 'POST');
      expect(request.uri.path, '/v1/devices');
      expect(jsonDecode(jsonEncode(request.data)), {
        'id': _installationId,
        'platform': 'android',
        'push_token': _tokenValue,
        'app_version': '0.1.0',
      });
    },
  );

  test('unregisters by installation id', () async {
    await repository.unregisterDevice(_installationId);

    final request = adapter.requests.single;
    expect(request.method, 'DELETE');
    expect(request.uri.path, '/v1/devices/$_installationId');
  });

  test('reads and replaces the settings, early_day null kept', () async {
    final read = await repository.getSettings();
    expect(read, defaultSettings);

    const next = ReminderSettings(
      enabled: true,
      earlyDay: null,
      onDueDate: true,
      sendTime: '08:30',
      showAmounts: false,
    );
    await repository.putSettings(next);

    final put = adapter.requests.last;
    expect(put.method, 'PUT');
    expect(jsonDecode(jsonEncode(put.data)), {
      'enabled': true,
      'early_day': null,
      'on_due_date': true,
      'send_time': '08:30',
      'show_amounts': false,
    });
  });

  test('reads the calendar of a year', () async {
    final calendar = await repository.getCalendar(2026);

    expect(adapter.requests.single.uri.queryParameters, {'year': '2026'});
    expect(calendar.entries, hasLength(12));
  });

  test('an unseeded year is an ApiRejected', () async {
    adapter.reply('/v1/tax-calendar', (
      status: 422,
      body: errorJson('business_rule_violation', 'calendar_unavailable'),
    ));

    final thrown = await failureOf(() => repository.getCalendar(2028));

    expect(thrown, isA<ApiRejected>());
  });

  test('a failed registration reports no token (M9-09 carry-over)', () async {
    adapter.reply(
      '/v1/devices',
      DioException.connectionError(
        requestOptions: RequestOptions(path: '/v1/devices'),
        reason: 'offline',
      ),
    );

    final thrown = await failureOf(
      () => repository.registerDevice(
        installationId: _installationId,
        platform: DevicePlatform.ios,
        token: const PushToken(_tokenValue),
      ),
    );

    expect(thrown, isA<ApiFailure>());
    expect('$thrown', isNot(contains(_tokenValue)));
    expect(
      describeAbsorbedError('push.register', thrown!),
      isNot(contains(_tokenValue)),
    );
  });
}
