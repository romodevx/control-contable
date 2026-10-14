import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/end_regime_panel.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_tile.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// The taxpayer's regimes, active and ended, each with its period. An
/// active regime offers its options and ending it (not the last active
/// one: the taxpayer is archived instead); a regime can be added. An
/// archived taxpayer shows them read-only.
class RegimesSection extends StatelessWidget {
  /// The regimes of [taxpayer].
  const new({required this.taxpayer, super.key});

  /// The taxpayer.
  final Taxpayer taxpayer;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final editable = taxpayer.archivedAt == null;
    final activeCount = taxpayer.regimes
        .where((r) => r.endPeriod == null)
        .length;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        for (final regime in taxpayer.regimes)
          _RegimeCard(
            taxpayerId: taxpayer.id,
            regime: regime,
            canEdit: editable && regime.endPeriod == null,
            canEnd: editable && regime.endPeriod == null && activeCount > 1,
          ),
        if (editable && activeCount == 1) ...[
          Text(l10n.settingsLastRegime),
          const SizedBox(height: AppSpacing.xs),
        ],
        if (editable)
          OutlinedButton.icon(
            icon: const Icon(Icons.add),
            label: Text(l10n.settingsAddRegime),
            onPressed: () => unawaited(
              context.push<void>(TaxpayersRoutes.addRegime(taxpayer.id)),
            ),
          ),
      ],
    );
  }
}

class _RegimeCard extends StatefulWidget {
  const new({
    required this.taxpayerId,
    required this.regime,
    required this.canEdit,
    required this.canEnd,
  });

  final String taxpayerId;
  final TaxpayerRegime regime;
  final bool canEdit;
  final bool canEnd;

  @override
  State<_RegimeCard> createState() => _RegimeCardState();
}

class _RegimeCardState extends State<_RegimeCard> {
  bool _ending = false;

  String _period(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final months = MaterialLocalizations.of(context);
    String month(String period) => months.formatMonthYear(periodMonth(period));
    final start = month(widget.regime.startPeriod);
    return switch (widget.regime.endPeriod) {
      final end? => l10n.settingsRegimeRange(start, month(end)),
      null => l10n.settingsRegimeSince(start),
    };
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final regime = widget.regime;
    final theme = Theme.of(context);
    return Card(
      margin: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              '${regime.regimeCode} · ${regime.name}',
              style: theme.textTheme.titleMedium,
            ),
            Wrap(
              spacing: AppSpacing.xs,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                Text(_period(context)),
                if (regime.endPeriod != null)
                  TextBadge(label: l10n.settingsRegimeEnded),
              ],
            ),
            if (_ending)
              Padding(
                padding: const EdgeInsets.only(top: AppSpacing.sm),
                child: EndRegimePanel(
                  taxpayerId: widget.taxpayerId,
                  regime: regime,
                  onClose: () => setState(() => _ending = false),
                ),
              )
            else
              Wrap(
                spacing: AppSpacing.xs,
                children: [
                  if (widget.canEdit && regime.options.isNotEmpty)
                    TextButton(
                      onPressed: () => unawaited(
                        context.push<void>(
                          TaxpayersRoutes.regimeOptions(
                            widget.taxpayerId,
                            regime.regimeCode,
                          ),
                        ),
                      ),
                      child: Text(l10n.settingsEditOptions),
                    ),
                  if (widget.canEnd)
                    TextButton(
                      onPressed: () => setState(() => _ending = true),
                      child: Text(l10n.settingsEndRegime),
                    ),
                ],
              ),
          ],
        ),
      ),
    );
  }
}
