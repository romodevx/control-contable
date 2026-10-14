import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/data_section.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/manage_section.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/regimes_section.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// A taxpayer's settings (M2-10, US-T3, US-T3c): its data (rename, RFC
/// correction), its regimes (add, options, end) and archive or delete. Read
/// from the taxpayer list, which every write reloads.
class TaxpayerSettingsScreen extends ConsumerWidget {
  /// The settings of the taxpayer [taxpayerId].
  const new({required this.taxpayerId, super.key});

  /// The taxpayer's id (from the location).
  final String taxpayerId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final list = ref.watch(taxpayersProvider);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.settingsTitle)),
      body: switch (list) {
        AsyncData(:final value) => switch (_find(value)) {
          final taxpayer? => _Sections(taxpayer: taxpayer),
          null => _Padded(LoadStatus(message: l10n.settingsNotFound)),
        },
        AsyncError() => _Padded(
          LoadFailed(
            message: l10n.settingsLoadFailed,
            onRetry: () => ref.invalidate(taxpayersProvider),
          ),
        ),
        _ => _Padded(LoadStatus(message: l10n.settingsLoading, busy: true)),
      },
    );
  }

  Taxpayer? _find(List<Taxpayer> taxpayers) =>
      taxpayers.where((t) => t.id == taxpayerId).firstOrNull;
}

class _Padded extends StatelessWidget {
  const new(this.child);

  final Widget child;

  @override
  Widget build(BuildContext context) =>
      Padding(padding: const EdgeInsets.all(AppSpacing.gutter), child: child);
}

class _Sections extends StatelessWidget {
  const new({required this.taxpayer});

  final Taxpayer taxpayer;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    // A few short sections (at most ~10 regimes): built eagerly, so every
    // section is in the semantics tree.
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          if (taxpayer.archivedAt != null) ...[
            FormNotice(
              tone: FormNoticeTone.info,
              message: l10n.settingsArchivedBanner,
            ),
            const SizedBox(height: AppSpacing.lg),
          ],
          _Heading(l10n.settingsDataTitle),
          DataSection(taxpayer: taxpayer),
          const SizedBox(height: AppSpacing.xl),
          _Heading(l10n.settingsRegimesTitle),
          RegimesSection(taxpayer: taxpayer),
          const SizedBox(height: AppSpacing.xl),
          _Heading(l10n.settingsManageTitle),
          ManageSection(taxpayer: taxpayer),
        ],
      ),
    );
  }
}

class _Heading extends StatelessWidget {
  const new(this.text);

  final String text;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: AppSpacing.sm),
    child: Semantics(
      header: true,
      child: Text(text, style: Theme.of(context).textTheme.titleLarge),
    ),
  );
}
