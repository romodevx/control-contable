import 'package:control_contable/api/generated/models/dashboard_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Income collected vs expenses paid this month (MODULE-dashboard US-D2),
/// before IVA. A module that is OFF comes back `null` and its row is not
/// shown; "Diferencia" (never "utilidad") needs both. Nothing at all when
/// both are OFF.
class MonthCard extends StatelessWidget {
  /// The card of [month].
  const new({required this.month, super.key});

  /// This month's figures.
  final DashboardMonth month;

  @override
  Widget build(BuildContext context) {
    final income = month.incomeCents;
    final expense = month.expenseCents;
    if (income == null && expense == null) return const SizedBox.shrink();
    final l10n = AppLocalizations.of(context);
    final tokens = context.tokens;
    final text = Theme.of(context).textTheme;
    return DecoratedBox(
      key: const ValueKey('dashboard-month'),
      decoration: BoxDecoration(
        color: tokens.colors.surface,
        border: Border.all(color: tokens.colors.border),
        borderRadius: BorderRadius.circular(tokens.radii.card),
      ),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Semantics(
              header: true,
              child: Text(l10n.dashboardMonthTitle, style: text.titleMedium),
            ),
            const SizedBox(height: AppSpacing.xs),
            if (income != null)
              MoneyRow(label: l10n.dashboardIncomeCollected, centavos: income),
            if (expense != null)
              MoneyRow(label: l10n.dashboardExpensePaid, centavos: expense),
            if (income != null && expense != null) ...[
              Divider(color: tokens.colors.border),
              MoneyRow(
                label: l10n.dashboardDifference,
                centavos: income - expense,
              ),
            ],
            const SizedBox(height: AppSpacing.xxs),
            Text(
              l10n.dashboardBeforeIva,
              style: text.bodyMedium?.copyWith(color: tokens.colors.muted),
            ),
          ],
        ),
      ),
    );
  }
}
