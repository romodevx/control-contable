import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// The month chosen on the estimate screen: `null` is the current month
/// (the default), asked without a period so the server's clock decides.
/// Forgotten when the screen closes.
class EstimatePeriodNotifier extends Notifier<String?> {
  @override
  String? build() => null;

  /// The chosen month; `null` is the current one.
  String? get period => state;

  /// Chooses [period]; `null` goes back to the current month.
  set period(String? period) => state = period;
}

/// See [EstimatePeriodNotifier].
final NotifierProvider<EstimatePeriodNotifier, String?> estimatePeriodProvider =
    NotifierProvider.autoDispose<EstimatePeriodNotifier, String?>(
      EstimatePeriodNotifier.new,
      name: 'estimatePeriodProvider',
    );

/// Which estimate to read: a taxpayer and a month (`null`: the current).
typedef EstimateQuery = ({String taxpayerId, String? period});

/// The estimate of one taxpayer and month, read on every visit (the server
/// derives it from the current records, never stores it). A failure shows
/// a retry: no automatic retry.
final FutureProviderFamily<TaxpayerEstimate, EstimateQuery> estimateProvider =
    FutureProvider.autoDispose.family<TaxpayerEstimate, EstimateQuery>(
      (ref, query) => ref
          .read(taxEngineRepositoryProvider)
          .getEstimate(taxpayerId: query.taxpayerId, period: query.period),
      name: 'estimateProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
