import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/settings_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Asks to delete [taxpayer]; answers `true` once it is deleted.
Future<bool?> showDeleteTaxpayerDialog(
  BuildContext context,
  Taxpayer taxpayer,
) => showDialog<bool>(
  context: context,
  builder: (_) => DeleteTaxpayerDialog(taxpayer: taxpayer),
);

/// The confirmation of the irreversible delete (US-T3c): it names the
/// taxpayer and says its incomes and expenses are deleted too, and points
/// to archiving for someone who only wants to hide it. A modal route
/// (inert background, Esc and the barrier close it, focus returns to the
/// trigger), labeled by its title. The destructive primary is never
/// disabled: busy while deleting, repeat taps ignored, not closable then.
class DeleteTaxpayerDialog extends ConsumerStatefulWidget {
  /// The confirmation for [taxpayer].
  const new({required this.taxpayer, super.key});

  /// The taxpayer to delete.
  final Taxpayer taxpayer;

  @override
  ConsumerState<DeleteTaxpayerDialog> createState() =>
      _DeleteTaxpayerDialogState();
}

class _DeleteTaxpayerDialogState extends ConsumerState<DeleteTaxpayerDialog> {
  bool _busy = false;
  String? _problem;

  Future<void> _delete() async {
    if (_busy) return;
    setState(() => _busy = true);
    final l10n = AppLocalizations.of(context);
    try {
      await ref
          .read(taxpayerSettingsProvider)
          .run(
            WriteAccess.always,
            (repository) => repository.delete(widget.taxpayer.id),
          );
      if (mounted) Navigator.of(context).pop(true);
    } on ApiFailure catch (failure) {
      if (!mounted) return;
      setState(() {
        _busy = false;
        _problem = settingsFailureMessage(failure, l10n);
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final error = Theme.of(context).colorScheme.error;
    final confirm = l10n.settingsDeleteConfirm;
    return PopScope(
      canPop: !_busy,
      child: AlertDialog(
        title: Text(l10n.settingsDeleteTitle(widget.taxpayer.displayName)),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(l10n.settingsDeleteBody),
              if (_problem case final problem?) ...[
                const SizedBox(height: AppSpacing.md),
                FormNotice(message: problem),
              ],
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: _busy ? null : () => Navigator.of(context).pop(false),
            child: Text(l10n.formCancel),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: error),
            onPressed: _delete,
            child: _busy
                ? BusyLabel(label: confirm, announcement: l10n.formSaving)
                : Text(confirm),
          ),
        ],
      ),
    );
  }
}
