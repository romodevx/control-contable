import 'package:control_contable/api/generated/models/history_month.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/history/presentation/history_month_labels.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

const double _markerIconSize = 20;

/// One month of History (direction A): the month and its right-aligned
/// total (wrapping under it at large text), ISR · IVA, withholdings when
/// any, the status chip, the due date, then the changed marker or when it
/// was seen. One target, read as one sentence ([HistoryMonthLabels.spoken]),
/// followed by "Ver ingresos" / "Ver gastos" as separate targets, each
/// shown only while its module is on (US-H6).
class HistoryMonthTile extends StatelessWidget {
  /// The row of [month], opened by [onOpen]; [onOpenRecords] opens a
  /// record list location.
  const new({
    required this.month,
    required this.onOpen,
    required this.onOpenRecords,
    super.key,
  });

  /// The month shown.
  final HistoryMonth month;

  /// Opens the month's estimate.
  final VoidCallback onOpen;

  /// Opens a record list location of the month.
  final ValueChanged<String> onOpenRecords;

  @override
  Widget build(BuildContext context) {
    final labels = HistoryMonthLabels(context, month);
    final colors = context.tokens.colors;
    final text = Theme.of(context).textTheme;
    final muted = text.bodyMedium?.copyWith(color: colors.muted);
    final tax = month.tax;
    final status = labels.status;
    final gap = labels.gap;
    final marker = labels.marker;
    final viewed = labels.viewed;
    return DecoratedBox(
      decoration: BoxDecoration(
        border: Border(bottom: BorderSide(color: colors.border)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Semantics(
            button: true,
            label: labels.spoken,
            onTap: onOpen,
            excludeSemantics: true,
            child: InkWell(
              onTap: onOpen,
              child: Padding(
                padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  spacing: AppSpacing.xxs,
                  children: [
                    _Heading(labels: labels, isCurrent: month.isCurrent),
                    if (tax != null) ...[
                      Text(labels.taxLine(tax), style: muted),
                      if (tax.withholdingsToRemitCents > 0)
                        Text(labels.withholdings(tax), style: muted),
                    ],
                    if (status != null)
                      Align(
                        alignment: AlignmentDirectional.centerStart,
                        child: StatusChip(label: status),
                      ),
                    if (gap != null) Text(gap, style: muted),
                    if (tax != null) Text(labels.dueDate, style: muted),
                    if (marker != null) _ChangedMarker(text: marker),
                    if (viewed != null) Text(viewed, style: muted),
                  ],
                ),
              ),
            ),
          ),
          _RecordLinks(labels: labels, onOpen: onOpenRecords),
        ],
      ),
    );
  }
}

/// "Ver ingresos" / "Ver gastos" of the month, each behind its module's
/// flag; nothing when both are off. A navigation aid, not a
/// reconciliation: the lists select by booking date (MODULE-history §1.1).
class _RecordLinks extends StatelessWidget {
  const new({required this.labels, required this.onOpen});

  final HistoryMonthLabels labels;
  final ValueChanged<String> onOpen;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final period = labels.month.period;
    return Wrap(
      spacing: AppSpacing.xs,
      children: [
        FlagGate(
          module: AppModule.incomes,
          child: TextButton(
            onPressed: () => onOpen(IncomesRoutes.listOf(period)),
            child: Text(
              l10n.historyViewIncomes,
              semanticsLabel: labels.incomesLink,
            ),
          ),
        ),
        FlagGate(
          module: AppModule.expenses,
          child: TextButton(
            onPressed: () => onOpen(ExpensesRoutes.listOf(period)),
            child: Text(
              l10n.historyViewExpenses,
              semanticsLabel: labels.expensesLink,
            ),
          ),
        ),
      ],
    );
  }
}

/// The month (and "Mes en curso") with its total: beside it when it fits,
/// under it otherwise (200 % text).
class _Heading extends StatelessWidget {
  const new({required this.labels, required this.isCurrent});

  final HistoryMonthLabels labels;
  final bool isCurrent;

  @override
  Widget build(BuildContext context) {
    final title = Theme.of(context).textTheme.titleMedium;
    final total = labels.month.tax?.totalPayableRoundedCents;
    return Wrap(
      alignment: WrapAlignment.spaceBetween,
      crossAxisAlignment: WrapCrossAlignment.center,
      spacing: AppSpacing.xs,
      runSpacing: AppSpacing.xxs,
      children: [
        Wrap(
          spacing: AppSpacing.xs,
          runSpacing: AppSpacing.xxs,
          crossAxisAlignment: WrapCrossAlignment.center,
          children: [
            Text(labels.name, style: title),
            if (isCurrent)
              StatusChip(
                label: AppLocalizations.of(context).historyCurrentMonth,
                color: context.tokens.colors.muted,
              ),
          ],
        ),
        if (total != null)
          Text(
            formatMoney(total),
            style: title?.copyWith(
              fontFeatures: const [FontFeature.tabularFigures()],
            ),
          ),
      ],
    );
  }
}

/// "Cambió desde que lo viste: …" in the legend's warning tint, with an
/// icon: never color alone (MODULE-history §6).
class _ChangedMarker extends StatelessWidget {
  const new({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    final tokens = context.tokens;
    final ink = tokens.colors.legendText;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: tokens.colors.legendBg,
        borderRadius: BorderRadius.circular(tokens.radii.control),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xs),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(Icons.update_rounded, size: _markerIconSize, color: ink),
            const SizedBox(width: AppSpacing.xs),
            Expanded(
              child: Text(
                text,
                style: Theme.of(context).textTheme.bodyMedium
                    ?.copyWith(color: ink),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
