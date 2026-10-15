import 'package:control_contable/api/generated/models/annual_breakdown.dart';
import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/features/annual/data/annual_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// The fiscal year chosen on the annual screen: `null` is the default
/// (`defaultAnnualYear`, which follows the clock). Forgotten when the
/// screen closes.
class AnnualYearNotifier extends Notifier<int?> {
  @override
  int? build() => null;

  /// Chooses [year]; choosing [byDefault] goes back to `null`, so the
  /// default keeps following the clock.
  void select(int year, {required int byDefault}) =>
      state = year == byDefault ? null : year;
}

/// See [AnnualYearNotifier].
final NotifierProvider<AnnualYearNotifier, int?> annualYearProvider =
    NotifierProvider.autoDispose<AnnualYearNotifier, int?>(
      AnnualYearNotifier.new,
      name: 'annualYearProvider',
    );

/// Which annual estimate to read.
typedef AnnualQuery = ({String taxpayerId, int fiscalYear});

/// The annual estimate of one taxpayer and year, read on every visit (the
/// server derives it, never stores it). A failure shows a retry: no
/// automatic retry.
final FutureProviderFamily<AnnualEstimate, AnnualQuery> annualEstimateProvider =
    FutureProvider.autoDispose.family<AnnualEstimate, AnnualQuery>(
      (ref, query) => ref
          .read(annualRepositoryProvider)
          .getEstimate(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
          ),
      name: 'annualEstimateProvider',
      retry: _noAutomaticRetry,
    );

/// How many deductions and constancias the year holds ("Datos para tu
/// anual"). Counters only: a failure leaves them out, never blocks.
final FutureProviderFamily<AnnualInputCounts, AnnualQuery>
annualInputCountsProvider = FutureProvider.autoDispose
    .family<AnnualInputCounts, AnnualQuery>(
      (ref, query) => ref
          .read(annualRepositoryProvider)
          .getInputCounts(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
          ),
      name: 'annualInputCountsProvider',
      retry: _noAutomaticRetry,
    );

/// Which annual breakdown to read.
typedef AnnualBreakdownQuery = ({
  String taxpayerId,
  int fiscalYear,
  AnnualComputedSectionKind section,
});

/// The computation trace of one annual section, read on every visit.
final FutureProviderFamily<AnnualBreakdown, AnnualBreakdownQuery>
annualBreakdownProvider = FutureProvider.autoDispose
    .family<AnnualBreakdown, AnnualBreakdownQuery>(
      (ref, query) => ref
          .read(annualRepositoryProvider)
          .getBreakdown(
            taxpayerId: query.taxpayerId,
            fiscalYear: query.fiscalYear,
            section: query.section,
          ),
      name: 'annualBreakdownProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
