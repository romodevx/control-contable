import 'dart:async';

import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/core/records/record_list.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which incomes a list shows: one taxpayer's month (`YYYY-MM`), all
/// regimes or one.
typedef IncomeListQuery = RecordListQuery;

/// The loaded pages of one [IncomeListQuery] and the month's totals.
typedef IncomeListState = RecordListState<Income>;

/// One month of a taxpayer's incomes (MODULE-incomes §2.9).
///
/// A `404` means the taxpayer is gone (deleted or archived elsewhere, or
/// the module is off): the taxpayer list reloads so the app moves on
/// (PR #61 follow-up).
class IncomeListNotifier extends RecordListNotifier<Income> {
  /// The list of [query].
  new(super.query);

  @override
  Future<RecordPage<Income>> fetchPage(String? cursor) async {
    final page = await ref
        .read(incomesRepositoryProvider)
        .list(
          taxpayerId: query.taxpayerId,
          period: query.period,
          regime: query.regime,
          cursor: cursor,
        );
    return (
      items: page.items,
      totals: page.totals,
      nextCursor: page.nextCursor,
      legend: null,
    );
  }

  @override
  String idOf(Income record) => record.id;

  @override
  RecordAmounts amountsOf(Income record) => (
    amountCents: record.amountCents,
    ivaCents: record.ivaCents,
    isrWithheldCents: record.isrWithheldCents,
    ivaWithheldCents: record.ivaWithheldCents,
  );

  @override
  void onNotFound() => unawaited(ref.read(taxpayersProvider.notifier).reload());
}

/// The incomes of one [IncomeListQuery] (see [IncomeListNotifier]).
final AsyncNotifierProviderFamily<
  IncomeListNotifier,
  IncomeListState,
  IncomeListQuery
>
incomeListProvider = AsyncNotifierProvider.autoDispose
    .family<IncomeListNotifier, IncomeListState, IncomeListQuery>(
      IncomeListNotifier.new,
      name: 'incomeListProvider',
      retry: noAutomaticRetry,
    );
