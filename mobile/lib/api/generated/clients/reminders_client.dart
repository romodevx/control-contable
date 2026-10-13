// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/device.dart';
import '../models/device_create.dart';
import '../models/fiscal_year.dart';
import '../models/reminder_settings.dart';
import '../models/tax_calendar.dart';
import '../models/uuid.dart';

part 'reminders_client.g.dart';

@RestApi()
abstract class RemindersClient {
  factory RemindersClient(Dio dio, {String? baseUrl}) = _RemindersClient;

  /// Register or refresh a push token
  @POST('/v1/devices')
  Future<Device> registerDevice({@Body() required DeviceCreate body});

  /// Remove a push token (logout / opt-out)
  @DELETE('/v1/devices/{deviceId}')
  Future<void> unregisterDevice({@Path('deviceId') required Uuid deviceId});

  /// Reminder preferences (defaults if never set)
  @GET('/v1/reminder-settings')
  Future<ReminderSettings> getReminderSettings();

  /// Replace reminder preferences
  @PUT('/v1/reminder-settings')
  Future<ReminderSettings> putReminderSettings({
    @Body() required ReminderSettings body,
  });

  /// Monthly due dates for a year (day 17 moved past weekends/holidays)
  @GET('/v1/tax-calendar')
  Future<TaxCalendar> getTaxCalendar({@Query('year') required FiscalYear year});
}
