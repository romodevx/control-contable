import 'dart:async';

import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/records/record_empty_state.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/records/record_page_footer.dart';
import 'package:control_contable/core/records/record_row.dart';
import 'package:control_contable/core/records/record_totals_card.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/incomes/application/income_list.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/features/incomes/presentation/income_delete_action.dart';
import 'package:control_contable/features/incomes/presentation/income_labels.dart';
import 'package:control_contable/features/incomes/presentation/incomes_screen.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Room under the last row so the floating *Registrar ingreso* never
/// covers its delete button.
const double _fabClearance = 88;

/// The incomes of [query]: totals, then the rows newest first; the next
/// page loads when the list reaches its end. Handles loading, failure and
/// an empty month.
class IncomeListView extends ConsumerWidget {
  /// The list of [query] for [taxpayer]'s [month].
  const new({
    required this.taxpayer,
    required this.month,
    required this.query,
    super.key,
  });

  /// Whose incomes (its regime names label the rows).
  final Taxpayer taxpayer;

  /// The month shown.
  final RecordMonth month;

  /// What the list loads.
  final IncomeListQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final list = ref.watch(incomeListProvider(query));
    return switch (list) {
      AsyncValue(:final value?) when value.items.isEmpty && !value.hasMore =>
        RecordEmptyState(
          title: l10n.incomesEmptyTitle(
            MaterialLocalizations.of(context).formatMonthYear(month.firstDay),
          ),
          body: l10n.incomesEmptyBody,
          actionLabel: l10n.incomesRecord,
          onAction: () => openIncomeForm(context),
        ),
      AsyncValue(:final value?) => _Loaded(
        state: value,
        taxpayer: taxpayer,
        query: query,
      ),
      AsyncError() => Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: LoadFailed(
          message: l10n.incomesLoadFailed,
          onRetry: () => ref.invalidate(incomeListProvider(query)),
        ),
      ),
      _ => Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: LoadStatus(message: l10n.incomesLoading, busy: true),
      ),
    };
  }
}

class _Loaded extends ConsumerWidget {
  const new({required this.state, required this.taxpayer, required this.query});

  final IncomeListState state;
  final Taxpayer taxpayer;
  final IncomeListQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final totals = state.totals;
    final items = state.items;
    return ListView.builder(
      padding: const EdgeInsets.only(bottom: _fabClearance),
      itemCount: items.length + 2,
      itemBuilder: (context, index) {
        if (index == 0) {
          return Padding(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.gutter,
              0,
              AppSpacing.gutter,
              AppSpacing.md,
            ),
            child: RecordTotalsCard(
              title: l10n.incomesTotalTitle,
              totalCents: totals.amountCents + totals.ivaCents,
              countLabel: l10n.incomesCount(totals.count),
              lines: [
                (label: l10n.incomesTotalSubtotal, cents: totals.amountCents),
                (label: l10n.incomesTotalIva, cents: totals.ivaCents),
                if (totals.isrWithheldCents + totals.ivaWithheldCents > 0)
                  (
                    label: l10n.incomesTotalWithheld,
                    cents: -(totals.isrWithheldCents + totals.ivaWithheldCents),
                  ),
              ],
            ),
          );
        }
        if (index > items.length) {
          return RecordPageFooter(
            hasMore: state.hasMore,
            isLoading: state.isLoadingMore,
            failed: state.loadMoreFailed,
            onLoadMore: ref.read(incomeListProvider(query).notifier).loadMore,
            loadingMessage: l10n.incomesLoading,
            failedMessage: l10n.incomesLoadFailed,
          );
        }
        return _row(context, ref, items[index - 1]);
      },
    );
  }

  Widget _row(BuildContext context, WidgetRef ref, Income income) {
    final l10n = AppLocalizations.of(context);
    final title =
        // ignore: deprecated_member_use_from_same_package, deprecated until WH-11 (ADR-0008).
        income.description ?? counterpartyLabel(income.counterpartyType, l10n);
    final regime = taxpayer.regimes
        .where((r) => r.regimeCode == income.regimeCode)
        .firstOrNull;
    return RecordRow(
      key: ValueKey(income.id),
      date: MaterialLocalizations.of(context)
          .formatShortMonthDay(bookingDateOf(income)),
      title: title,
      amountCents: income.amountCents + income.ivaCents,
      tag: regime?.name ?? income.regimeCode,
      badge: income.pendingCollection ? l10n.incomesPendingCollection : null,
      deleteTooltip: l10n.incomesDeleteTooltip(title),
      openLabel: l10n.incomesEditLabel(title),
      onOpen: () => unawaited(
        context.push<void>(IncomesRoutes.editOf(income.id), extra: income),
      ),
      onDelete: () => unawaited(deleteIncome(context, ref, query, income)),
    );
  }
}
