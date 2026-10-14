import 'dart:async';

import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/fiscal/regimes_held.dart';
import 'package:control_contable/core/records/record_empty_state.dart';
import 'package:control_contable/core/records/record_list.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/records/record_page_footer.dart';
import 'package:control_contable/core/records/record_row.dart';
import 'package:control_contable/core/records/record_totals_card.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/expenses/application/expense_categories.dart';
import 'package:control_contable/features/expenses/application/expense_list.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/expenses/presentation/expense_delete_action.dart';
import 'package:control_contable/features/expenses/presentation/expenses_screen.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Room under the last row so the floating *Registrar gasto* never covers
/// its delete button.
const double _fabClearance = 88;

/// The expenses of [query]: the server's legend, totals, then the rows
/// newest first; the next page loads when the list reaches its end.
/// Handles loading, failure and an empty month (whose header still shows
/// the legend).
class ExpenseListView extends ConsumerWidget {
  /// The list of [query] for [taxpayer]'s [month].
  const new({
    required this.taxpayer,
    required this.month,
    required this.query,
    super.key,
  });

  /// Whose expenses (its regime names label the rows).
  final Taxpayer taxpayer;

  /// The month shown.
  final RecordMonth month;

  /// What the list loads.
  final RecordListQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final list = ref.watch(expenseListProvider(query));
    return switch (list) {
      AsyncValue(:final value?) when value.items.isEmpty && !value.hasMore =>
        ListView(
          children: [
            if (value.legend case final legend?) _Legend(legend),
            RecordEmptyState(
              title: l10n.expensesEmptyTitle(
                MaterialLocalizations.of(context)
                    .formatMonthYear(month.firstDay),
              ),
              body: l10n.expensesEmptyBody,
              actionLabel: l10n.expensesRecord,
              onAction: () => openExpenseForm(context),
            ),
          ],
        ),
      AsyncValue(:final value?) => _Loaded(
        state: value,
        taxpayer: taxpayer,
        query: query,
        categoryNames: _categoryNames(ref),
      ),
      AsyncError() => Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: LoadFailed(
          message: l10n.expensesLoadFailed,
          onRetry: () => ref.invalidate(expenseListProvider(query)),
        ),
      ),
      _ => Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: LoadStatus(message: l10n.expensesLoading, busy: true),
      ),
    };
  }

  /// Category names by code (they are the same for every regime); empty
  /// while the catalog loads or when it cannot, and rows show the code.
  Map<String, String> _categoryNames(WidgetRef ref) {
    final regime =
        query.regime ??
        regimesHeldIn(taxpayer, month.period).firstOrNull?.regimeCode;
    if (regime == null) return const {};
    final catalog = ref.watch(
      expenseCatalogProvider((
        taxpayerId: taxpayer.id,
        regime: regime,
        on: null,
      )),
    );
    return {
      for (final category in catalog.value?.items ?? const <ExpenseCategory>[])
        category.code: category.name,
    };
  }
}

class _Legend extends StatelessWidget {
  const new(this.text);

  final String text;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.fromLTRB(
      AppSpacing.gutter,
      0,
      AppSpacing.gutter,
      AppSpacing.md,
    ),
    child: EstimateLegend(text: text),
  );
}

class _Loaded extends ConsumerWidget {
  const new({
    required this.state,
    required this.taxpayer,
    required this.query,
    required this.categoryNames,
  });

  final ExpenseListState state;
  final Taxpayer taxpayer;
  final RecordListQuery query;
  final Map<String, String> categoryNames;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final items = state.items;
    final legend = state.legend;
    return ListView.builder(
      padding: const EdgeInsets.only(bottom: _fabClearance),
      itemCount: items.length + 2,
      itemBuilder: (context, index) {
        if (index == 0) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [if (legend != null) _Legend(legend), _totals(l10n)],
          );
        }
        if (index > items.length) {
          return RecordPageFooter(
            hasMore: state.hasMore,
            isLoading: state.isLoadingMore,
            failed: state.loadMoreFailed,
            onLoadMore: ref.read(expenseListProvider(query).notifier).loadMore,
            loadingMessage: l10n.expensesLoading,
            failedMessage: l10n.expensesLoadFailed,
          );
        }
        return _row(context, ref, items[index - 1]);
      },
    );
  }

  Widget _totals(AppLocalizations l10n) {
    final totals = state.totals;
    final withheld = totals.isrWithheldCents + totals.ivaWithheldCents;
    return Padding(
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.gutter,
        0,
        AppSpacing.gutter,
        AppSpacing.md,
      ),
      child: RecordTotalsCard(
        title: l10n.expensesTotalTitle,
        totalCents: totals.amountCents + totals.ivaCents,
        countLabel: l10n.expensesCount(totals.count),
        lines: [
          (label: l10n.expensesTotalSubtotal, cents: totals.amountCents),
          (label: l10n.expensesTotalIva, cents: totals.ivaCents),
          if (withheld > 0)
            (label: l10n.expensesTotalWithheld, cents: withheld),
        ],
      ),
    );
  }

  Widget _row(BuildContext context, WidgetRef ref, Expense expense) {
    final l10n = AppLocalizations.of(context);
    final title =
        expense.description ??
        categoryNames[expense.categoryCode] ??
        expense.categoryCode;
    final regime = taxpayer.regimes
        .where((r) => r.regimeCode == expense.regimeCode)
        .firstOrNull;
    return RecordRow(
      key: ValueKey(expense.id),
      date: MaterialLocalizations.of(context)
          .formatShortMonthDay(parseBusinessDate(expense.paidOn)),
      title: title,
      amountCents: expense.amountCents + expense.ivaCents,
      tag: regime?.name ?? expense.regimeCode,
      badge: expense.deductible ? null : l10n.expensesNotDeductible,
      deleteTooltip: l10n.expensesDeleteTooltip(title),
      openLabel: l10n.expensesEditLabel(title),
      onOpen: () => unawaited(
        context.push<void>(ExpensesRoutes.editOf(expense.id), extra: expense),
      ),
      onDelete: () => unawaited(deleteExpense(context, ref, query, expense)),
    );
  }
}
