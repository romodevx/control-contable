import 'package:control_contable/api/generated/models/breakdown.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which breakdown to read: a taxpayer, a month (`YYYY-MM`) and a regime.
typedef BreakdownQuery = ({String taxpayerId, String period, String regime});

/// The breakdown of one regime and month, read on every visit (derived
/// from the current records, never stored). A failure shows a retry: no
/// automatic retry.
final FutureProviderFamily<Breakdown, BreakdownQuery> breakdownProvider =
    FutureProvider.autoDispose.family<Breakdown, BreakdownQuery>(
      (ref, query) => ref
          .read(taxEngineRepositoryProvider)
          .getBreakdown(
            taxpayerId: query.taxpayerId,
            period: query.period,
            regime: query.regime,
          ),
      name: 'breakdownProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
