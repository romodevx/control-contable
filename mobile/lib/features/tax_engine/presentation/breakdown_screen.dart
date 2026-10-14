import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/tax_engine/application/breakdown.dart';
import 'package:control_contable/features/tax_engine/presentation/breakdown_view.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_states.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// One regime's breakdown for a month (M5-26): every step of the
/// computation with its amount, source and legal basis. Read from the
/// server on every visit; the client never recomputes.
class BreakdownScreen extends ConsumerWidget {
  /// The breakdown of [regime] in [period] for [taxpayerId].
  const new({
    required this.taxpayerId,
    required this.period,
    required this.regime,
    super.key,
  });

  /// Whose breakdown.
  final String taxpayerId;

  /// The month (`YYYY-MM`).
  final String period;

  /// The regime code.
  final String regime;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final query = (taxpayerId: taxpayerId, period: period, regime: regime);
    final breakdown = ref.watch(breakdownProvider(query));
    return Scaffold(
      appBar: AppBar(title: Text(l10n.breakdownTitle)),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        children: [
          if (breakdown.value case final loaded?)
            BreakdownView(breakdown: loaded)
          else if (breakdown.error case final error?)
            EstimateFailure(
              error: error,
              onRetry: () => ref.invalidate(breakdownProvider(query)),
            )
          else
            LoadStatus(message: l10n.breakdownLoading, busy: true),
        ],
      ),
    );
  }
}
