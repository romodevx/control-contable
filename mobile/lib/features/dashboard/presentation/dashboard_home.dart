import 'package:control_contable/api/generated/models/access_level.dart';
import 'package:control_contable/api/generated/models/dashboard.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/dashboard/application/dashboard.dart';
import 'package:control_contable/features/dashboard/domain/dashboard_trend.dart';
import 'package:control_contable/features/dashboard/presentation/access_banner.dart';
import 'package:control_contable/features/dashboard/presentation/dashboard_hero.dart';
import 'package:control_contable/features/dashboard/presentation/dashboard_states.dart';
import 'package:control_contable/features/dashboard/presentation/month_card.dart';
import 'package:control_contable/features/dashboard/presentation/quick_actions.dart';
import 'package:control_contable/features/dashboard/presentation/trend_section.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/presentation/no_regime_prompt.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_launcher.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Home for the active taxpayer (M6): pull to refresh over the dashboard
/// of the taxpayer chosen in the header. Place it behind
/// `FlagGate(module: AppModule.dashboard, …)`.
class DashboardHome extends ConsumerWidget {
  /// Creates Home.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final taxpayer = ref.watch(activeTaxpayerProvider);
    final taxpayerId = taxpayer.value?.id;
    return RefreshIndicator(
      onRefresh: () => _refresh(ref, taxpayerId),
      child: ListView(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        physics: const AlwaysScrollableScrollPhysics(),
        children: [
          const FlagGate(
            module: AppModule.taxpayers,
            child: OnboardingLauncher(),
          ),
          if (taxpayerId != null)
            _DashboardContent(taxpayerId: taxpayerId)
          else if (taxpayer.isLoading)
            const DashboardSkeleton()
          else
            const DashboardNoTaxpayer(),
        ],
      ),
    );
  }

  static Future<void> _refresh(WidgetRef ref, String? taxpayerId) async {
    if (taxpayerId == null) return;
    try {
      final reloaded = ref.refresh(dashboardProvider(taxpayerId).future);
      await reloaded;
    } on ApiFailure {
      // The provider holds the failure; Home shows it with a retry.
    }
  }
}

class _DashboardContent extends ConsumerWidget {
  const new({required this.taxpayerId});

  final String taxpayerId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dashboard = ref.watch(dashboardProvider(taxpayerId));
    void reload() => ref.invalidate(dashboardProvider(taxpayerId));
    if (dashboard.value case final loaded?) {
      return Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        spacing: AppSpacing.xl,
        children: [
          for (final block in _blocks(loaded, reload))
            Semantics(container: true, child: block),
        ],
      );
    }
    if (dashboard.error case final error? when !dashboard.isLoading) {
      return DashboardFailure(error: error, onRetry: reload);
    }
    return const DashboardSkeleton();
  }

  /// Home in reading order (MODULE-dashboard §6): access banner → strong
  /// CPA banner + hero + legend → month card → chart → quick actions.
  static List<Widget> _blocks(Dashboard dashboard, VoidCallback reload) {
    final status = heroStatusOf(dashboard);
    final estimate = dashboard.estimate;
    final legend = dashboard.legend ?? estimate?.legend;
    final hasTrend = trendViewsOf(dashboard.trend).isNotEmpty;
    return [
      if (dashboard.access.access
          case AccessLevel.trial || AccessLevel.readOnly)
        AccessBanner(access: dashboard.access),
      if (dashboard.isEmpty)
        const DashboardEmpty()
      else ...[
        if (estimate != null && status == TaxpayerEstimateStatus.none)
          const FlagGate(module: AppModule.taxpayers, child: NoRegimePrompt())
        else if (estimate != null && legend != null)
          DashboardHero(estimate: estimate, status: status, legend: legend),
        if (dashboard.month.incomeCents != null ||
            dashboard.month.expenseCents != null)
          MonthCard(month: dashboard.month),
        if (hasTrend) TrendSection(trend: dashboard.trend, legend: legend),
      ],
      QuickActions(
        readOnly: dashboard.access.access == AccessLevel.readOnly,
        onReturn: reload,
      ),
    ];
  }
}
