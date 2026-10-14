import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/tax_engine/application/estimate.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_period_picker.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_states.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_view.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The active taxpayer's monthly estimate (M5-25): a month picker (the
/// current month by default), then the estimate. Pull down to read it
/// again; it is derived from the current records on every read.
class EstimateScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final current = currentPeriod(ref.watch(clockProvider)());
    final chosen = ref.watch(estimatePeriodProvider);
    final taxpayer = ref.watch(activeTaxpayerProvider);
    final taxpayerId = taxpayer.value?.id;
    final query = taxpayerId == null
        ? null
        : (taxpayerId: taxpayerId, period: chosen);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.estimateTitle)),
      body: RefreshIndicator(
        onRefresh: () => _refresh(ref, query),
        child: ListView(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            EstimatePeriodPicker(
              current: current,
              selected: chosen ?? current,
              onSelected: (period) =>
                  ref.read(estimatePeriodProvider.notifier).period =
                      period == current ? null : period,
            ),
            const SizedBox(height: AppSpacing.xl),
            if (query != null)
              _EstimateContent(query: query)
            else if (taxpayer.isLoading)
              const EstimateSkeleton()
            else
              EstimateMessage(message: l10n.estimateNoTaxpayer),
          ],
        ),
      ),
    );
  }

  Future<void> _refresh(WidgetRef ref, EstimateQuery? query) async {
    if (query == null) return;
    try {
      final reloaded = ref.refresh(estimateProvider(query).future);
      await reloaded;
    } on ApiFailure {
      // The provider holds the failure; the screen shows it with a retry.
    }
  }
}

class _EstimateContent extends ConsumerWidget {
  const new({required this.query});

  final EstimateQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final estimate = ref.watch(estimateProvider(query));
    if (estimate.value case final loaded?) {
      return EstimateView(estimate: loaded);
    }
    if (estimate.error case final error?) {
      return EstimateFailure(
        error: error,
        onRetry: () => ref.invalidate(estimateProvider(query)),
      );
    }
    return const EstimateSkeleton();
  }
}
