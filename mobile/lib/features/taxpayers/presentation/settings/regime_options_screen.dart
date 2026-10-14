import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/application/regime_catalog.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/regime_options_form.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Edits the options of the regime [regimeCode] of [taxpayerId] (US-T3b),
/// rendered from the catalog's `option_defs` and starting from the
/// options the taxpayer holds.
class RegimeOptionsScreen extends ConsumerWidget {
  /// The options screen of one held regime.
  const new({required this.taxpayerId, required this.regimeCode, super.key});

  /// The taxpayer's id (from the location).
  final String taxpayerId;

  /// The regime's code (from the location).
  final String regimeCode;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final taxpayer = ref
        .watch(taxpayersProvider)
        .value
        ?.where((t) => t.id == taxpayerId)
        .firstOrNull;
    final held = taxpayer?.regimes
        .where((r) => r.regimeCode == regimeCode && r.endPeriod == null)
        .firstOrNull;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.onboardingOptionsTitle)),
      body: taxpayer == null || held == null
          ? _Status(LoadStatus(message: l10n.settingsNotFound))
          : _CatalogGate(taxpayer: taxpayer, held: held),
    );
  }
}

/// Loads the person type's catalog for the regime's option definitions.
class _CatalogGate extends ConsumerWidget {
  const new({required this.taxpayer, required this.held});

  final Taxpayer taxpayer;
  final TaxpayerRegime held;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final provider = regimeCatalogProvider(taxpayer.personType);
    return switch (ref.watch(provider)) {
      AsyncData(:final value) => switch (_regimeIn(value)) {
        final regime? => RegimeOptionsForm(
          taxpayerId: taxpayer.id,
          regime: regime,
          current: held.options,
        ),
        null => _Status(LoadStatus(message: l10n.settingsNotFound)),
      },
      AsyncError() => _Status(
        LoadFailed(
          message: l10n.onboardingRegimesLoadFailed,
          onRetry: () => ref.invalidate(provider),
        ),
      ),
      _ => _Status(
        LoadStatus(message: l10n.onboardingRegimesLoading, busy: true),
      ),
    };
  }

  Regime? _regimeIn(List<Regime> catalog) =>
      catalog.where((r) => r.code == held.regimeCode).firstOrNull;
}

class _Status extends StatelessWidget {
  const new(this.child);

  final Widget child;

  @override
  Widget build(BuildContext context) =>
      Padding(padding: const EdgeInsets.all(AppSpacing.gutter), child: child);
}
