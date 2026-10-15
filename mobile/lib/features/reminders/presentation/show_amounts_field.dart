import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The amounts opt-in (default off) and the lock-screen privacy note.
///
/// The toggle carries its consequence in its own label, not in a tooltip
/// (MODULE-reminders §2.5, §7), and is hidden while `module_tax_engine` is
/// off: with no estimates there are no amounts to show. The note stays: the
/// notification still shows the taxpayers' names on the lock screen.
class ShowAmountsField extends StatelessWidget {
  /// The toggle showing [value].
  const new({required this.value, required this.onChanged, super.key});

  /// Whether amounts are shown.
  final bool value;

  /// The user flipped the toggle.
  final ValueChanged<bool> onChanged;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        FlagGate(
          module: AppModule.taxEngine,
          child: SwitchListTile(
            title: Text(l10n.remindersShowAmountsLabel),
            value: value,
            onChanged: onChanged,
          ),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.gutter),
          child: Text(
            l10n.remindersPrivacyNote,
            style: Theme.of(context).textTheme.bodySmall,
          ),
        ),
      ],
    );
  }
}
