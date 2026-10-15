import 'dart:async';
import 'dart:developer' as developer;

import 'package:control_contable/api/generated/models/history_viewed_request.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/history/application/history.dart';
import 'package:control_contable/features/history/data/history_repository.dart';
import 'package:control_contable/features/history/domain/history_viewed.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const _logName = 'history';

/// Marks a past month viewed whenever its estimate is shown, from any
/// entry point (MODULE-history §2.5, OQ-H7), while `module_history` is on.
///
/// Fire-and-forget: the post never blocks the estimate screen; a failure
/// is logged (never the fingerprints) and the next open posts again. Once
/// it settles, History is read again so its marker reflects the view.
class HistoryViewMarker {
  /// The marker reading flags, clock and server through [_ref].
  new(this._ref);

  final Ref _ref;

  /// Posts "seen" for [estimate] when History is on and the month is past.
  void onEstimateRendered(TaxpayerEstimate estimate) {
    if (!_ref.read(moduleEnabledProvider(AppModule.history))) return;
    final current = currentPeriod(_ref.read(clockProvider)());
    final request = viewedRequestOf(estimate, current: current);
    if (request == null) return;
    unawaited(_post(estimate, request));
  }

  Future<void> _post(
    TaxpayerEstimate estimate,
    HistoryViewedRequest request,
  ) async {
    try {
      await _ref
          .read(historyRepositoryProvider)
          .markViewed(
            taxpayerId: estimate.taxpayerId,
            period: estimate.period,
            request: request,
          );
    } on ApiFailure catch (failure) {
      developer.log(
        'markHistoryViewed ${estimate.period} failed: '
        '${failure.runtimeType}',
        name: _logName,
      );
    }
    if (_ref.mounted) _ref.invalidate(historyYearProvider);
  }
}
