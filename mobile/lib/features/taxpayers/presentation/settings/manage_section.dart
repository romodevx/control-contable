import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/delete_taxpayer_dialog.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/settings_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Archive or unarchive (reversible: done at once, then Undo in a toast
/// that does not auto-dismiss) and delete (irreversible: a confirmation
/// dialog first). Archive and delete work for a read-only user; unarchive
/// (and so Undo) needs a plan (OQ-8).
class ManageSection extends ConsumerWidget {
  /// The actions on [taxpayer].
  const new({required this.taxpayer, super.key});

  /// The taxpayer.
  final Taxpayer taxpayer;

  Future<void> _archive(BuildContext context, WidgetRef ref) async {
    // Undo may be tapped after this screen is gone: it keeps what it needs
    // and never reads [context] before checking it is still mounted.
    final feedback = _Feedback.of(context);
    final settings = ref.read(taxpayerSettingsProvider);
    if (!await _run(
      context,
      feedback,
      settings,
      WriteAccess.always,
      _archiveCall,
    )) {
      return;
    }
    feedback.toast(
      feedback.l10n.settingsArchived(taxpayer.displayName),
      action: SnackBarAction(
        label: feedback.l10n.settingsUndo,
        onPressed: () => unawaited(_unarchiveWith(context, feedback, settings)),
      ),
    );
  }

  Future<void> _unarchive(BuildContext context, WidgetRef ref) =>
      _unarchiveWith(
        context,
        _Feedback.of(context),
        ref.read(taxpayerSettingsProvider),
      );

  Future<void> _unarchiveWith(
    BuildContext context,
    _Feedback feedback,
    TaxpayerSettings settings,
  ) async {
    final done = await _run(
      context,
      feedback,
      settings,
      WriteAccess.needsPlan,
      _unarchiveCall,
    );
    if (done)
      feedback.toast(feedback.l10n.settingsUnarchived(taxpayer.displayName));
  }

  Future<void> _delete(BuildContext context) async {
    // The list reloads without this taxpayer before the dialog closes, so
    // this section is gone by then: keep what the follow-up needs.
    final messenger = ScaffoldMessenger.of(context);
    final router = GoRouter.of(context);
    final message = AppLocalizations.of(context)
        .settingsDeleted(taxpayer.displayName);
    final deleted = await showDeleteTaxpayerDialog(context, taxpayer);
    if (deleted != true) return;
    showSettingsToast(messenger, message);
    router.go(AppRoutes.home);
  }

  Future<Taxpayer> _archiveCall(TaxpayersRepository r) =>
      r.archive(taxpayer.id);

  Future<Taxpayer> _unarchiveCall(TaxpayersRepository r) =>
      r.unarchive(taxpayer.id);

  /// Runs one write; `false` when it failed (the paywall or a toast says
  /// why). The paywall opens over [context] only while it is mounted.
  Future<bool> _run(
    BuildContext context,
    _Feedback feedback,
    TaxpayerSettings settings,
    WriteAccess access,
    Future<Taxpayer> Function(TaxpayersRepository) write,
  ) async {
    try {
      await settings.run(access, write);
      return true;
    } on ApiEntitlementRequired {
      if (context.mounted) {
        unawaited(showPaywall(context));
      } else {
        feedback.toast(feedback.l10n.paywallWriteBody);
      }
    } on ApiFailure catch (failure) {
      feedback.toast(settingsFailureMessage(failure, feedback.l10n));
    }
    return false;
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final archived = taxpayer.archivedAt != null;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (!archived) ...[
          Text(l10n.settingsArchiveHelp),
          const SizedBox(height: AppSpacing.xs),
          OutlinedButton(
            onPressed: () => unawaited(_archive(context, ref)),
            child: Text(l10n.settingsArchive),
          ),
        ] else
          FilledButton(
            onPressed: () => unawaited(_unarchive(context, ref)),
            child: Text(l10n.settingsUnarchive),
          ),
        const SizedBox(height: AppSpacing.lg),
        Text(l10n.settingsDeleteHelp),
        const SizedBox(height: AppSpacing.xs),
        OutlinedButton(
          style: OutlinedButton.styleFrom(
            foregroundColor: Theme.of(context).colorScheme.error,
          ),
          onPressed: () => unawaited(_delete(context)),
          child: Text(l10n.settingsDelete),
        ),
      ],
    );
  }
}

/// What a write's follow-up needs, read while the screen is mounted: the
/// toast outlives the screen.
final class _Feedback {
  _Feedback.of(BuildContext context)
    : messenger = ScaffoldMessenger.of(context),
      l10n = AppLocalizations.of(context);

  final ScaffoldMessengerState messenger;
  final AppLocalizations l10n;

  void toast(String message, {SnackBarAction? action}) =>
      showSettingsToast(messenger, message, action: action);
}
