import 'dart:async';
import 'dart:developer' as developer;

import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/period_totals.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which incomes a list shows: one taxpayer's month (`YYYY-MM`), all
/// regimes or one.
typedef IncomeListQuery = ({String taxpayerId, String period, String? regime});

/// The loaded pages of one [IncomeListQuery] and the month's totals.
@immutable
class IncomeListState {
  /// The state after loading [items] with [totals].
  const new({
    required this.items,
    required this.totals,
    required this.nextCursor,
    this.isLoadingMore = false,
    this.loadMoreFailed = false,
  });

  /// The incomes loaded so far, newest first.
  final List<Income> items;

  /// Totals of the whole filtered month (not only the loaded pages).
  final PeriodTotals totals;

  /// The next page's cursor; `null` once every page is loaded.
  final String? nextCursor;

  /// Whether the next page is loading.
  final bool isLoadingMore;

  /// Whether the last attempt at the next page failed.
  final bool loadMoreFailed;

  /// Whether more pages exist.
  bool get hasMore => nextCursor != null;

  /// A copy with the given fields replaced.
  IncomeListState copyWith({
    List<Income>? items,
    PeriodTotals? totals,
    String? nextCursor,
    bool clearCursor = false,
    bool? isLoadingMore,
    bool? loadMoreFailed,
  }) => IncomeListState(
    items: items ?? this.items,
    totals: totals ?? this.totals,
    nextCursor: clearCursor ? null : nextCursor ?? this.nextCursor,
    isLoadingMore: isLoadingMore ?? this.isLoadingMore,
    loadMoreFailed: loadMoreFailed ?? this.loadMoreFailed,
  );
}

/// One month of a taxpayer's incomes (MODULE-incomes §2.9): the first page
/// on build, the next ones on demand ([loadMore]).
///
/// A `404` means the taxpayer is gone (deleted or archived elsewhere, or
/// the module is off): the taxpayer list reloads so the app moves on
/// (PR #61 follow-up). No automatic retry: the screen offers one.
class IncomeListNotifier extends AsyncNotifier<IncomeListState> {
  /// The list of [query].
  new(this.query);

  /// What this list shows.
  final IncomeListQuery query;

  IncomesRepository get _repository => ref.read(incomesRepositoryProvider);

  @override
  Future<IncomeListState> build() async {
    try {
      final page = await _repository.list(
        taxpayerId: query.taxpayerId,
        period: query.period,
        regime: query.regime,
      );
      return IncomeListState(
        items: page.items,
        totals: page.totals,
        nextCursor: page.nextCursor,
      );
    } on ApiNotFound {
      unawaited(ref.read(taxpayersProvider.notifier).reload());
      rethrow;
    }
  }

  /// Loads the next page; a no-op while one loads or when none is left.
  Future<void> loadMore() async {
    if (!ref.mounted) return;
    final current = state.value;
    final cursor = current?.nextCursor;
    if (current == null || cursor == null || current.isLoadingMore) return;
    state = AsyncData(
      current.copyWith(isLoadingMore: true, loadMoreFailed: false),
    );
    try {
      final page = await _repository.list(
        taxpayerId: query.taxpayerId,
        period: query.period,
        regime: query.regime,
        cursor: cursor,
      );
      if (!ref.mounted) return;
      state = AsyncData(
        current.copyWith(
          items: [...current.items, ...page.items],
          totals: page.totals,
          nextCursor: page.nextCursor,
          clearCursor: page.nextCursor == null,
          isLoadingMore: false,
        ),
      );
    } on ApiFailure catch (failure) {
      developer.log(
        'next income page failed (${failure.runtimeType})',
        name: 'incomes',
      );
      if (!ref.mounted) return;
      state = AsyncData(
        current.copyWith(isLoadingMore: false, loadMoreFailed: true),
      );
    }
  }

  /// Takes [income] off the list and its amounts off the totals at once,
  /// while the delete is in flight (the server answer follows).
  void hide(Income income) {
    final current = state.value;
    if (current == null) return;
    if (!current.items.any((i) => i.id == income.id)) return;
    final t = current.totals;
    state = AsyncData(
      current.copyWith(
        items: [
          for (final i in current.items)
            if (i.id != income.id) i,
        ],
        totals: PeriodTotals(
          count: t.count - 1,
          amountCents: t.amountCents - income.amountCents,
          ivaCents: t.ivaCents - income.ivaCents,
          isrWithheldCents: t.isrWithheldCents - income.isrWithheldCents,
          ivaWithheldCents: t.ivaWithheldCents - income.ivaWithheldCents,
        ),
      ),
    );
  }
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
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
