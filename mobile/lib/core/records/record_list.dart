import 'dart:developer' as developer;

import 'package:control_contable/api/generated/models/period_totals.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Which records a month list shows: one taxpayer's month (`YYYY-MM`), all
/// regimes or one.
typedef RecordListQuery = ({String taxpayerId, String period, String? regime});

/// The money of one record, in centavos (what leaves the totals when it is
/// hidden).
typedef RecordAmounts = ({
  int amountCents,
  int ivaCents,
  int isrWithheldCents,
  int ivaWithheldCents,
});

/// One page of a month list as the server answers it. `legend` is the
/// mandatory text some lists carry in their header (expenses).
typedef RecordPage<T> = ({
  List<T> items,
  PeriodTotals totals,
  String? nextCursor,
  String? legend,
});

/// The loaded pages of one [RecordListQuery] and the month's totals.
@immutable
class RecordListState<T> {
  /// The state after loading [items] with [totals].
  const new({
    required this.items,
    required this.totals,
    required this.nextCursor,
    this.legend,
    this.isLoadingMore = false,
    this.loadMoreFailed = false,
  });

  /// The records loaded so far, newest first.
  final List<T> items;

  /// Totals of the whole filtered month (not only the loaded pages).
  final PeriodTotals totals;

  /// The next page's cursor; `null` once every page is loaded.
  final String? nextCursor;

  /// The server's header legend, when the list has one.
  final String? legend;

  /// Whether the next page is loading.
  final bool isLoadingMore;

  /// Whether the last attempt at the next page failed.
  final bool loadMoreFailed;

  /// Whether more pages exist.
  bool get hasMore => nextCursor != null;

  /// A copy with the given fields replaced.
  RecordListState<T> copyWith({
    List<T>? items,
    PeriodTotals? totals,
    String? nextCursor,
    bool clearCursor = false,
    bool? isLoadingMore,
    bool? loadMoreFailed,
  }) => RecordListState<T>(
    items: items ?? this.items,
    totals: totals ?? this.totals,
    nextCursor: clearCursor ? null : nextCursor ?? this.nextCursor,
    legend: legend,
    isLoadingMore: isLoadingMore ?? this.isLoadingMore,
    loadMoreFailed: loadMoreFailed ?? this.loadMoreFailed,
  );
}

/// One month of a taxpayer's records (MODULE-incomes §2.9, shared by
/// expenses): the first page on build, the next ones on demand
/// ([loadMore]), and [hide] while a delete is in flight.
///
/// A `404` on the first page calls [onNotFound] (the taxpayer is gone, or
/// the module is off) and fails. No automatic retry: the screen offers one.
abstract class RecordListNotifier<T> extends AsyncNotifier<RecordListState<T>> {
  /// The list of [query].
  new(this.query);

  /// What this list shows.
  final RecordListQuery query;

  /// The page after [cursor] (the first one when `null`).
  Future<RecordPage<T>> fetchPage(String? cursor);

  /// The id of [record].
  String idOf(T record);

  /// The money of [record].
  RecordAmounts amountsOf(T record);

  /// The first page answered `404`.
  void onNotFound() {}

  @override
  Future<RecordListState<T>> build() async {
    try {
      final page = await fetchPage(null);
      return RecordListState<T>(
        items: page.items,
        totals: page.totals,
        nextCursor: page.nextCursor,
        legend: page.legend,
      );
    } on ApiNotFound {
      onNotFound();
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
      final page = await fetchPage(cursor);
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
        'next page failed (${failure.runtimeType})',
        name: 'records',
      );
      if (!ref.mounted) return;
      state = AsyncData(
        current.copyWith(isLoadingMore: false, loadMoreFailed: true),
      );
    }
  }

  /// Takes [record] off the list and its amounts off the totals at once,
  /// while the delete is in flight (the server answer follows).
  void hide(T record) {
    final current = state.value;
    if (current == null) return;
    final id = idOf(record);
    if (!current.items.any((r) => idOf(r) == id)) return;
    final t = current.totals;
    final gone = amountsOf(record);
    state = AsyncData(
      current.copyWith(
        items: [
          for (final r in current.items)
            if (idOf(r) != id) r,
        ],
        totals: PeriodTotals(
          count: t.count - 1,
          amountCents: t.amountCents - gone.amountCents,
          ivaCents: t.ivaCents - gone.ivaCents,
          isrWithheldCents: t.isrWithheldCents - gone.isrWithheldCents,
          ivaWithheldCents: t.ivaWithheldCents - gone.ivaWithheldCents,
        ),
      ),
    );
  }
}

/// No automatic retry for a record list: the screen offers one.
Duration? noAutomaticRetry(int retryCount, Object error) => null;
