import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/presentation/annual_breakdown_view.dart';
import 'package:control_contable/features/annual/presentation/annual_states.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// One annual section's computation (ANN-18): every step with its amount,
/// source and legal basis. Read from the server on every visit; the client
/// never recomputes.
class AnnualBreakdownScreen extends ConsumerWidget {
  /// The breakdown of [section] in [fiscalYear] for [taxpayerId].
  const new({
    required this.taxpayerId,
    required this.fiscalYear,
    required this.section,
    super.key,
  });

  /// Whose breakdown.
  final String taxpayerId;

  /// The fiscal year.
  final int fiscalYear;

  /// The computed section.
  final AnnualComputedSectionKind section;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final query = (
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
      section: section,
    );
    final breakdown = ref.watch(annualBreakdownProvider(query));
    return Scaffold(
      appBar: AppBar(title: Text(l10n.annualBreakdownTitle)),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        children: [
          if (breakdown.value case final loaded?)
            AnnualBreakdownView(breakdown: loaded)
          else if (breakdown.error case final error?)
            AnnualFailure(
              error: error,
              onRetry: () => ref.invalidate(annualBreakdownProvider(query)),
            )
          else
            LoadStatus(message: l10n.annualBreakdownLoading, busy: true),
        ],
      ),
    );
  }
}
