import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/domain/annual_years.dart';
import 'package:control_contable/features/annual/presentation/annual_states.dart';
import 'package:control_contable/features/annual/presentation/annual_view.dart';
import 'package:control_contable/features/annual/presentation/annual_year_picker.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_states.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The active taxpayer's annual estimate (ANN-17): a fiscal-year picker
/// (the previous year until 30 April, then the current one), then the
/// estimate. Pull down to read it again; it is derived from the current
/// records on every read.
class AnnualScreen extends ConsumerWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final now = ref.watch(clockProvider)();
    final defaultYear = defaultAnnualYear(now);
    final year = ref.watch(annualYearProvider) ?? defaultYear;
    final taxpayer = ref.watch(activeTaxpayerProvider);
    final taxpayerId = taxpayer.value?.id;
    final query = taxpayerId == null
        ? null
        : (taxpayerId: taxpayerId, fiscalYear: year);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.annualTitle)),
      body: RefreshIndicator(
        onRefresh: () => _refresh(ref, query),
        child: ListView(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          physics: const AlwaysScrollableScrollPhysics(),
          children: [
            AnnualYearPicker(
              years: annualYears(now),
              selected: year,
              onSelected: (chosen) => ref
                  .read(annualYearProvider.notifier)
                  .select(chosen, byDefault: defaultYear),
            ),
            const SizedBox(height: AppSpacing.xl),
            if (query != null)
              _AnnualContent(query: query)
            else if (taxpayer.isLoading)
              LoadStatus(message: l10n.annualLoading, busy: true)
            else
              EstimateMessage(message: l10n.annualNoTaxpayer),
          ],
        ),
      ),
    );
  }

  Future<void> _refresh(WidgetRef ref, AnnualQuery? query) async {
    if (query == null) return;
    try {
      final reloaded = ref.refresh(annualEstimateProvider(query).future);
      await reloaded;
    } on ApiFailure {
      // The provider holds the failure; the screen shows it with a retry.
    }
  }
}

class _AnnualContent extends ConsumerWidget {
  const new({required this.query});

  final AnnualQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final estimate = ref.watch(annualEstimateProvider(query));
    if (estimate.value case final loaded?) return AnnualView(estimate: loaded);
    if (estimate.error case final error?) {
      return AnnualFailure(
        error: error,
        onRetry: () => ref.invalidate(annualEstimateProvider(query)),
      );
    }
    return LoadStatus(
      message: AppLocalizations.of(context).annualLoading,
      busy: true,
    );
  }
}
