import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/records/month_switcher.dart';
import 'package:control_contable/core/records/record_filter_chips.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/features/incomes/presentation/income_list_view.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_switcher.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// The active taxpayer's incomes, one month at a time (M3-08): month
/// switcher, regime filter, totals, rows with delete + Undo.
class IncomesScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final active = ref.watch(activeTaxpayerProvider);
    final taxpayer = active.value;
    return Scaffold(
      appBar: AppBar(
        title: ActiveTaxpayerSwitcher(fallback: Text(l10n.navIncome)),
      ),
      floatingActionButton: taxpayer == null
          ? null
          : FloatingActionButton.extended(
              icon: const Icon(Icons.add),
              label: Text(l10n.incomesRecord),
              onPressed: () => openIncomeForm(context),
            ),
      body: switch (active) {
        AsyncData(value: final Taxpayer taxpayer) => _MonthView(
          key: ValueKey(taxpayer.id),
          taxpayer: taxpayer,
        ),
        AsyncData() => _Message(l10n.incomesNoTaxpayer),
        AsyncError() => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadFailed(
            message: l10n.incomesLoadFailed,
            onRetry: () => ref.invalidate(taxpayersProvider),
          ),
        ),
        _ => Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadStatus(message: l10n.incomesLoading, busy: true),
        ),
      },
    );
  }
}

/// Opens the new-income page.
void openIncomeForm(BuildContext context) =>
    unawaited(context.push<void>(IncomesRoutes.create));

class _Message extends StatelessWidget {
  const new(this.text);

  final String text;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(AppSpacing.gutter),
    child: Text(text, style: Theme.of(context).textTheme.bodyLarge),
  );
}

/// One month of [taxpayer]'s incomes, with its month and regime choice.
class _MonthView extends ConsumerStatefulWidget {
  const new({required this.taxpayer, super.key});

  final Taxpayer taxpayer;

  @override
  ConsumerState<_MonthView> createState() => _MonthViewState();
}

class _MonthViewState extends ConsumerState<_MonthView> {
  late RecordMonth _month = RecordMonth.current(ref.read(clockProvider)());
  String? _regime;

  /// The regimes held at some point of [_month].
  List<ChoiceOption> get _regimeOptions {
    final period = _month.period;
    return [
      for (final regime in widget.taxpayer.regimes)
        if (regime.startPeriod.compareTo(period) <= 0 &&
            (regime.endPeriod == null ||
                regime.endPeriod!.compareTo(period) >= 0))
          (value: regime.regimeCode, label: regime.name),
    ];
  }

  void _showMonth(RecordMonth month) => setState(() {
    _month = month;
    final stillHeld = _regimeOptions.any((o) => o.value == _regime);
    if (!stillHeld) _regime = null;
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
            child: IncomeListView(
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
