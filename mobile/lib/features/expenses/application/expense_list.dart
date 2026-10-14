import 'dart:async';

import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/core/records/record_list.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// The loaded pages of one month of expenses, the totals and the legend.
typedef ExpenseListState = RecordListState<Expense>;

/// One month of a taxpayer's expenses (MODULE-expenses US-E2), by
/// `paid_on`. A `404` means the taxpayer is gone (or the module is off):
/// the taxpayer list reloads so the app moves on.
class ExpenseListNotifier extends RecordListNotifier<Expense> {
  /// The list of [query].
  new(super.query);

  @override
  Future<RecordPage<Expense>> fetchPage(String? cursor) async {
    final page = await ref
        .read(expensesRepositoryProvider)
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
      legend: page.legend,
    );
  }

  @override
  String idOf(Expense record) => record.id;

  @override
  RecordAmounts amountsOf(Expense record) => (
    amountCents: record.amountCents,
    ivaCents: record.ivaCents,
    isrWithheldCents: record.isrWithheldCents,
    ivaWithheldCents: record.ivaWithheldCents,
  );

  @override
  void onNotFound() => unawaited(ref.read(taxpayersProvider.notifier).reload());
}

/// The expenses of one [RecordListQuery] (see [ExpenseListNotifier]).
final AsyncNotifierProviderFamily<
  ExpenseListNotifier,
  ExpenseListState,
  RecordListQuery
>
expenseListProvider = AsyncNotifierProvider.autoDispose
    .family<ExpenseListNotifier, ExpenseListState, RecordListQuery>(
      ExpenseListNotifier.new,
      name: 'expenseListProvider',
      retry: noAutomaticRetry,
    );
