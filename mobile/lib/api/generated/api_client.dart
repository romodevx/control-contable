// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';

import 'clients/platform_client.dart';
import 'clients/auth_client.dart';
import 'clients/taxpayers_client.dart';
import 'clients/incomes_client.dart';
import 'clients/expenses_client.dart';
import 'clients/tax_engine_client.dart';
import 'clients/dashboard_client.dart';
import 'clients/subscriptions_client.dart';
import 'clients/history_client.dart';
import 'clients/reminders_client.dart';

/// Control Contable API `v0.2.0`.
///
/// Monthly IVA and provisional ISR estimates for Mexican taxpayers.
///
/// Spec-first contract (single source of truth). Server stubs (oapi-codegen),.
/// request validation, contract tests and the Dart client are generated from it.
///
/// Conventions:.
/// - Money is `int64` centavos (`*_cents`). Rates travel as decimal strings. No floats.
/// - Periods are `YYYY-MM`. Business dates are `date` (America/Mexico_City calendar).
/// - Every error uses the single `ErrorResponse` envelope.
/// - Every operation declares `x-module`. When that module's flag is OFF the server.
///   answers `404 not_found`, indistinguishable from a missing resource.
/// - Operations without authentication declare `security: []` AND `x-public: true`.
/// - Nodes marked `x-frozen: sha256:<hash>` are frozen; changing them is governed by.
///   `docs/api/CONTRACT.md` (lint fails when a frozen node drifts from its hash).
/// - Resources of other users answer `404` (RLS), never `403`.
/// - Create operations accept a client-generated UUID `id`; retrying the same POST.
///   returns the existing resource (idempotent against double submit).
///   Reusing an `id` with a different body, or an `id` that already exists outside.
///   the caller's scope, answers `409 conflict` (never reveals the other row).
///
/// Status: DRAFT 0.1.0 (Phase 2 Design). Regime-specific shapes (regime options,.
/// tax-input keys, breakdown concepts) are generic until.
/// `docs/research/tax-regimes-mx-2026.md` lands.
///
class ApiClient {
  ApiClient(Dio dio, {String? baseUrl}) : _dio = dio, _baseUrl = baseUrl;

  final Dio _dio;
  final String? _baseUrl;

  static String get version => '0.2.0';

  PlatformClient? _platform;
  AuthClient? _auth;
  TaxpayersClient? _taxpayers;
  IncomesClient? _incomes;
  ExpensesClient? _expenses;
  TaxEngineClient? _taxEngine;
  DashboardClient? _dashboard;
  SubscriptionsClient? _subscriptions;
  HistoryClient? _history;
  RemindersClient? _reminders;

  PlatformClient get platform =>
      _platform ??= PlatformClient(_dio, baseUrl: _baseUrl);

  AuthClient get auth => _auth ??= AuthClient(_dio, baseUrl: _baseUrl);

  TaxpayersClient get taxpayers =>
      _taxpayers ??= TaxpayersClient(_dio, baseUrl: _baseUrl);

  IncomesClient get incomes =>
      _incomes ??= IncomesClient(_dio, baseUrl: _baseUrl);

  ExpensesClient get expenses =>
      _expenses ??= ExpensesClient(_dio, baseUrl: _baseUrl);

  TaxEngineClient get taxEngine =>
      _taxEngine ??= TaxEngineClient(_dio, baseUrl: _baseUrl);

  DashboardClient get dashboard =>
      _dashboard ??= DashboardClient(_dio, baseUrl: _baseUrl);

  SubscriptionsClient get subscriptions =>
      _subscriptions ??= SubscriptionsClient(_dio, baseUrl: _baseUrl);

  HistoryClient get history =>
      _history ??= HistoryClient(_dio, baseUrl: _baseUrl);

  RemindersClient get reminders =>
      _reminders ??= RemindersClient(_dio, baseUrl: _baseUrl);
}
