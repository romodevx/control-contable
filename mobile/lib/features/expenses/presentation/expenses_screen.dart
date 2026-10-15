import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/fiscal/regimes_held.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/records/month_switcher.dart';
import 'package:control_contable/core/records/record_filter_chips.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/expenses/presentation/expense_list_view.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_switcher.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// The active taxpayer's expenses, one month at a time (M4-09): month
/// switcher, regime filter, the legend, totals, rows with delete + Undo.
class ExpensesScreen extends ConsumerWidget {
  /// Creates the screen, opened on [initialPeriod] (`YYYY-MM`) when it
  /// names a month up to the current one, else on the current month.
  const new({this.initialPeriod, super.key});

  /// The month to open on, as the location carries it (unchecked).
  final String? initialPeriod;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final active = ref.watch(activeTaxpayerProvider);
    return Scaffold(
      appBar: AppBar(
        title: ActiveTaxpayerSwitcher(fallback: Text(l10n.navExpenses)),
      ),
      floatingActionButton: active.value == null
          ? null
          : FloatingActionButton.extended(
              icon: const Icon(Icons.add),
              label: Text(l10n.expensesRecord),
              onPressed: () => openExpenseForm(context),
            ),
      body: switch (active) {
        AsyncData(value: final Taxpayer taxpayer) => _MonthView(
          key: ValueKey(taxpayer.id),
          taxpayer: taxpayer,
          initialPeriod: initialPeriod,
        ),
        AsyncData() => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: Text(
            l10n.expensesNoTaxpayer,
            style: Theme.of(context).textTheme.bodyLarge,
          ),
        ),
        AsyncError() => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadFailed(
            message: l10n.expensesLoadFailed,
            onRetry: () => ref.invalidate(taxpayersProvider),
          ),
        ),
        _ => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadStatus(message: l10n.expensesLoading, busy: true),
        ),
      },
    );
  }
}

/// Opens the new-expense page.
void openExpenseForm(BuildContext context) =>
    unawaited(context.push<void>(ExpensesRoutes.create));

/// One month of [taxpayer]'s expenses, with its month and regime choice.
class _MonthView extends ConsumerStatefulWidget {
  const new({required this.taxpayer, this.initialPeriod, super.key});

  final Taxpayer taxpayer;
  final String? initialPeriod;

  @override
  ConsumerState<_MonthView> createState() => _MonthViewState();
}

class _MonthViewState extends ConsumerState<_MonthView> {
  late RecordMonth _month = RecordMonth.opening(
    widget.initialPeriod,
    ref.read(clockProvider)(),
  );
  String? _regime;

  List<ChoiceOption> get _regimeOptions => [
    for (final regime in regimesHeldIn(widget.taxpayer, _month.period))
      (value: regime.regimeCode, label: regime.name),
  ];

  void _showMonth(RecordMonth month) => setState(() {
    _month = month;
    if (!_regimeOptions.any((o) => o.value == _regime)) _regime = null;
  });

  @override
  Widget build(BuildContext context) {
    final options = _regimeOptions;
    return MonthSwitcher(
      month: _month,
      latest: RecordMonth.current(ref.watch(clockProvider)()),
      onChanged: _showMonth,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          RecordFilterChips(
            label: AppLocalizations.of(context).recordsFilterLabel,
            options: options,
            selected: _regime,
            onSelected: (regime) => setState(() => _regime = regime),
          ),
          if (options.length > 1) const SizedBox(height: AppSpacing.xs),
          Expanded(
            child: ExpenseListView(
              taxpayer: widget.taxpayer,
              month: _month,
              query: (
                taxpayerId: widget.taxpayer.id,
                period: _month.period,
                regime: _regime,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
