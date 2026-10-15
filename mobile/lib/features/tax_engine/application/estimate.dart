import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:control_contable/core/fiscal/tax_data_revision.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// The month chosen on the estimate screen: `null` is the current month
/// (the default), asked without a period so the server's clock decides.
/// Starts at the month the screen was opened on (History opens a past
/// one); forgotten when the screen closes.
class EstimatePeriodNotifier extends Notifier<String?> {
  /// The choice, starting at [_opened] (`null`: the current month).
  new(this._opened);

  final String? _opened;

  @override
  String? build() => _opened;

  /// Chooses [period]; choosing [current] goes back to `null`, so the
  /// current month keeps following the server's clock.
  void select(String period, {required String current}) =>
      state = period == current ? null : period;
}

/// See [EstimatePeriodNotifier], keyed by the month the screen opened on.
final NotifierProviderFamily<EstimatePeriodNotifier, String?, String?>
estimatePeriodProvider = NotifierProvider.autoDispose
    .family<EstimatePeriodNotifier, String?, String?>(
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
      (ref, query) {
        ref.watch(taxDataRevisionProvider);
        return ref
            .read(taxEngineRepositoryProvider)
            .getEstimate(taxpayerId: query.taxpayerId, period: query.period);
      },
      name: 'estimateProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
