import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/tax_engine/application/estimate.dart';
import 'package:control_contable/features/tax_engine/application/estimate_rendered.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_period_picker.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_states.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_view.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The active taxpayer's monthly estimate (M5-25): a month picker (the
/// current month by default, or [initialPeriod]), then the estimate. Pull
/// down to read it again; it is derived from the current records on every
/// read. Each estimate shown is told to the app's listener (History marks
/// the month viewed).
class EstimateScreen extends ConsumerWidget {
  /// Creates the screen, opened on [initialPeriod] (`YYYY-MM`; `null`: the
  /// current month).
  const new({this.initialPeriod, super.key});

  /// The month the screen opens on; `null` is the current month.
  final String? initialPeriod;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final current = currentPeriod(ref.watch(clockProvider)());
    final chosen = ref.watch(estimatePeriodProvider(initialPeriod));
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
              onSelected: (period) => ref
                  .read(estimatePeriodProvider(initialPeriod).notifier)
                  .select(period, current: current),
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
    // Each new load that succeeds is told once (never on a mere rebuild).
    ref.listen(estimateProvider(query), (_, next) {
      if (next case AsyncData(:final value)) {
        ref.read(estimateRenderedListenerProvider)(value);
      }
    });
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
