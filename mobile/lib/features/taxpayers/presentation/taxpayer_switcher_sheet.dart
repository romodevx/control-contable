import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_tile.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Opens the taxpayer switcher over [context].
Future<void> showTaxpayerSwitcher(BuildContext context) =>
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      useSafeArea: true,
      builder: (_) => const TaxpayerSwitcherSheet(),
    );

/// The user's taxpayers (MODULE-taxpayers §2.9, US-T4): picking one makes
/// it active and closes the sheet; each row also opens its settings.
/// Archived taxpayers are hidden until "Mostrar archivados" (they cannot be
/// active; tapping one opens its settings, where it can be unarchived).
class TaxpayerSwitcherSheet extends ConsumerStatefulWidget {
  /// Creates the sheet.
  const new({super.key});

  @override
  ConsumerState<TaxpayerSwitcherSheet> createState() =>
      _TaxpayerSwitcherSheetState();
}

class _TaxpayerSwitcherSheetState extends ConsumerState<TaxpayerSwitcherSheet> {
  bool _showArchived = false;

  /// Closes the sheet, then goes to [location] over the screen below.
  void _closeAndOpen(String location) {
    final router = GoRouter.of(context);
    Navigator.of(context).pop();
    unawaited(router.push<void>(location));
  }

  Future<void> _pick(Taxpayer taxpayer) async {
    if (taxpayer.archivedAt != null) {
      return _closeAndOpen(TaxpayersRoutes.settings(taxpayer.id));
    }
    final navigator = Navigator.of(context);
    await ref.read(activeTaxpayerProvider.notifier).select(taxpayer.id);
    if (navigator.mounted) navigator.pop();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final list = ref.watch(taxpayersProvider);
    final activeId = ref.watch(activeTaxpayerProvider).value?.id;
    return Padding(
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.gutter,
        0,
        AppSpacing.gutter,
        AppSpacing.gutter,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Semantics(
            header: true,
            child: Text(
              l10n.switcherTitle,
              style: Theme.of(context).textTheme.headlineSmall,
            ),
          ),
          SwitchListTile(
            contentPadding: EdgeInsets.zero,
            title: Text(l10n.switcherShowArchived),
            value: _showArchived,
            onChanged: (value) => setState(() => _showArchived = value),
          ),
          Flexible(
            child: switch (list) {
              AsyncData(:final value) => _Rows(
                taxpayers: [
                  for (final t in value)
                    if (_showArchived || t.archivedAt == null) t,
                ],
                activeId: activeId,
                onPick: (t) => unawaited(_pick(t)),
                onSettings: (t) =>
                    _closeAndOpen(TaxpayersRoutes.settings(t.id)),
              ),
              AsyncError() => LoadFailed(
                message: l10n.settingsLoadFailed,
                onRetry: () => ref.invalidate(taxpayersProvider),
              ),
              _ => LoadStatus(message: l10n.settingsLoading, busy: true),
            },
          ),
          const SizedBox(height: AppSpacing.md),
          OutlinedButton.icon(
            icon: const Icon(Icons.person_add_alt_outlined),
            label: Text(l10n.homeAddTaxpayer),
            onPressed: () => _closeAndOpen(TaxpayersRoutes.onboarding),
          ),
        ],
      ),
    );
  }
}

class _Rows extends StatelessWidget {
  const new({
    required this.taxpayers,
    required this.activeId,
    required this.onPick,
    required this.onSettings,
  });

  final List<Taxpayer> taxpayers;
  final String? activeId;
  final ValueChanged<Taxpayer> onPick;
  final ValueChanged<Taxpayer> onSettings;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListView(
      shrinkWrap: true,
      children: [
        for (final taxpayer in taxpayers)
          ListTile(
            contentPadding: EdgeInsets.zero,
            selected: taxpayer.id == activeId,
            leading: ExcludeSemantics(
              child: Icon(
                taxpayer.id == activeId
                    ? Icons.radio_button_checked
                    : Icons.radio_button_unchecked,
              ),
            ),
            title: Text(taxpayer.displayName),
            subtitle: TaxpayerSummary(taxpayer: taxpayer),
            onTap: () => onPick(taxpayer),
            trailing: IconButton(
              icon: const Icon(Icons.settings_outlined),
              tooltip: l10n.taxpayerSettingsTooltip(taxpayer.displayName),
              onPressed: () => onSettings(taxpayer),
            ),
          ),
      ],
    );
  }
}
