import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Size of the busy spinner inside the primary button.
const _spinnerSize = 20.0;
const _spinnerStroke = 2.0;

/// A form's sticky footer: one primary (*Guardar*), an optional secondary
/// (*Guardar y crear otro*) and *Cancelar* as a text button.
///
/// The primary is **never disabled** (that would also break Enter to
/// submit): while saving it shows a busy state and repeat taps are ignored
/// by `AppFormController.submit`.
class FormFooter extends StatelessWidget {
  /// The footer of [controller]'s form.
  const new({
    required this.controller,
    required this.onPrimary,
    this.onCreateAnother,
    this.onCancel,
    this.primaryLabel,
    super.key,
  });

  /// The form; its busy state drives the primary.
  final AppFormController controller;

  /// Validates and saves.
  final VoidCallback onPrimary;

  /// Saves and readies the next record; no secondary button when `null`.
  final VoidCallback? onCreateAnother;

  /// Leaves without saving; no cancel button when `null`.
  final VoidCallback? onCancel;

  /// The primary's label; *Guardar* by default.
  final String? primaryLabel;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.tokens.colors;
    final label = primaryLabel ?? l10n.formSave;
    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.surface,
        border: Border(top: BorderSide(color: colors.border)),
      ),
      child: SafeArea(
        top: false,
        minimum: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            ListenableBuilder(
              listenable: controller,
              builder: (context, _) => FilledButton(
                onPressed: onPrimary,
                child: controller.isSaving
                    ? BusyLabel(label: label, announcement: l10n.formSaving)
                    : Text(label),
              ),
            ),
            if (onCreateAnother != null || onCancel != null)
              Padding(
                padding: const EdgeInsets.only(top: AppSpacing.xs),
                child: Wrap(
                  alignment: WrapAlignment.spaceBetween,
                  spacing: AppSpacing.xs,
                  runSpacing: AppSpacing.xs,
                  children: [
                    if (onCancel != null)
                      TextButton(
                        onPressed: onCancel,
                        child: Text(l10n.formCancel),
                      ),
                    if (onCreateAnother != null)
                      OutlinedButton(
                        onPressed: onCreateAnother,
                        child: Text(l10n.formSaveAndCreateAnother),
                      ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// A primary button's busy content: a spinner beside [label], read as
/// [announcement].
class BusyLabel extends StatelessWidget {
  /// The busy state of a button labeled [label].
  const new({required this.label, required this.announcement, super.key});

  /// The button's label.
  final String label;

  /// What a screen reader says while busy.
  final String announcement;

  @override
  Widget build(BuildContext context) => Semantics(
    label: announcement,
    excludeSemantics: true,
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox.square(
          dimension: _spinnerSize,
          child: CircularProgressIndicator(
            strokeWidth: _spinnerStroke,
            color: Theme.of(context).colorScheme.onPrimary,
          ),
        ),
        const SizedBox(width: AppSpacing.xs),
        Text(label),
      ],
    ),
  );
}
