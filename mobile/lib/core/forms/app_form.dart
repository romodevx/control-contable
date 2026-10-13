import 'dart:async';

import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_error_summary.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/forms/form_scope.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// A record form laid out by the house rules
/// (~/.claude/rules/ui-forms-and-modals.md): one column, the error summary
/// on top, [children] (`AppTextField`s) in table-column order, and a sticky
/// [FormFooter]. Ctrl+Enter or Cmd+Enter submits from any field.
///
/// [onSave] runs only when every field is valid, with the busy state on. It
/// owns its errors (show them, don't throw): a throw is reported as an
/// uncaught error after the busy state clears.
class AppForm extends StatelessWidget {
  /// A form over [controller].
  const new({
    required this.controller,
    required this.onSave,
    required this.children,
    this.onCancel,
    this.allowCreateAnother = false,
    this.primaryLabel,
    super.key,
  });

  /// The form's fields and submit state.
  final AppFormController controller;

  /// Persists the record. Server validation is the authority.
  final Future<void> Function() onSave;

  /// The fields, in the order of the table they feed.
  final List<Widget> children;

  /// Leaves without saving; no *Cancelar* when `null`.
  final VoidCallback? onCancel;

  /// Adds *Guardar y crear otro* (repetitive entry).
  final bool allowCreateAnother;

  /// The primary's label; *Guardar* by default.
  final String? primaryLabel;

  void _submit() => unawaited(controller.submit(onSave));

  Future<void> _saveAndCreateAnother() async {
    if (await controller.submit(onSave)) controller.prepareNext();
  }

  @override
  Widget build(BuildContext context) => AppFormScope(
    controller: controller,
    submit: _submit,
    child: CallbackShortcuts(
      bindings: {
        const SingleActivator(LogicalKeyboardKey.enter, control: true): _submit,
        const SingleActivator(LogicalKeyboardKey.enter, meta: true): _submit,
      },
      child: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(AppSpacing.gutter),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  FormErrorSummary(controller: controller),
                  for (final (index, field) in children.indexed) ...[
                    if (index > 0) const SizedBox(height: AppSpacing.lg),
                    field,
                  ],
                ],
              ),
            ),
          ),
          FormFooter(
            controller: controller,
            onPrimary: _submit,
            onCreateAnother: allowCreateAnother
                ? () => unawaited(_saveAndCreateAnother())
                : null,
            onCancel: onCancel,
            primaryLabel: primaryLabel,
          ),
        ],
      ),
    ),
  );
}
