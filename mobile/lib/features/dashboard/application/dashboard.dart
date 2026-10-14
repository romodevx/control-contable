import 'package:control_contable/api/generated/models/dashboard.dart';
import 'package:control_contable/features/dashboard/data/dashboard_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Home of one taxpayer, keyed by its id: switching taxpayer reads a new
/// key, so the previous taxpayer's figures are never shown while the new
/// one loads. Derived on every read (no cache, ADR-0002 §5); a failure
/// shows a retry, never an automatic one.
final FutureProviderFamily<Dashboard, String> dashboardProvider = FutureProvider
    .autoDispose
    .family<Dashboard, String>(
      (ref, taxpayerId) => ref
          .read(dashboardRepositoryProvider)
          .getDashboard(taxpayerId: taxpayerId),
      name: 'dashboardProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
