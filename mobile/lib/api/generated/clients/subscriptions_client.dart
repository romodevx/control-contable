// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/revenue_cat_webhook.dart';
import '../models/subscription_status.dart';

part 'subscriptions_client.g.dart';

@RestApi()
abstract class SubscriptionsClient {
  factory SubscriptionsClient(Dio dio, {String? baseUrl}) =
      _SubscriptionsClient;

  /// Access status (trial, pro, comp, read-only) and plan
  @GET('/v1/subscription')
  Future<SubscriptionStatus> getSubscription();

  /// Pull current state from RevenueCat (post-purchase / restore)
  @POST('/v1/subscription/sync')
  Future<SubscriptionStatus> syncSubscription();

  /// RevenueCat webhook receiver (idempotent by event id).
  ///
  /// Authenticated by the shared secret configured in RevenueCat, sent as.
  /// `Authorization: Bearer <secret>`. Duplicate `event.id` returns 200 with no effect.
  /// State is re-fetched from RevenueCat, so out-of-order delivery is safe.
  @POST('/v1/webhooks/revenuecat')
  Future<void> receiveRevenueCatWebhook({
    @Body() required RevenueCatWebhook body,
  });
}
