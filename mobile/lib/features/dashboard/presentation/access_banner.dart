import 'dart:async';

import 'package:control_contable/api/generated/models/access_level.dart';
import 'package:control_contable/api/generated/models/access_summary.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The trial / access banner (MODULE-dashboard §2.6): days and records
/// left on a trial, the read-only notice otherwise; nothing for `pro` and
/// `comp`. Persistent, never a toast. "Ver planes" appears only with
/// `module_subscriptions` (a placeholder until the M7 paywall).
class AccessBanner extends StatelessWidget {
  /// The banner of [access].
  const new({required this.access, super.key});

  /// The caller's access summary (server-computed days left).
  final AccessSummary access;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final message = switch (access.access) {
      AccessLevel.trial => _trialText(l10n),
      AccessLevel.readOnly => l10n.dashboardReadOnlyBanner,
      _ => null,
    };
    if (message == null) return const SizedBox.shrink();
    final tokens = context.tokens;
    final colors = tokens.colors;
    final spoken = access.access == AccessLevel.trial
        ? '$message. ${l10n.dashboardTrialScope}'
        : message;
    return Semantics(
      key: const ValueKey('access-banner'),
      container: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.surface,
          border: Border.all(color: colors.outline),
          borderRadius: BorderRadius.circular(tokens.radii.card),
        ),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Semantics(
                label: spoken,
                excludeSemantics: true,
                child: Text(
                  message,
                  style: Theme.of(context).textTheme.bodyMedium
                      ?.copyWith(color: colors.text),
                ),
              ),
              FlagGate(
                module: AppModule.subscriptions,
                child: TextButton(
                  onPressed: () => unawaited(
                    showPaywallSheet(
                      context,
                      title: l10n.dashboardPlansTitle,
                      body: l10n.dashboardPlansBody,
                    ),
                  ),
                  child: Text(l10n.dashboardSeePlans),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _trialText(AppLocalizations l10n) {
    final days = l10n.dashboardTrialBanner(access.trialDaysLeft ?? 1);
    final incomes = access.incomesRemaining;
    final expenses = access.expensesRemaining;
    if (incomes == null || expenses == null) return days;
    return days + l10n.dashboardTrialRecords(incomes, expenses);
  }
}
