import 'dart:async';

import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/forms/form_scope.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// Saves the typed value; answers the field's error, or `null` when saved.
typedef InlineSave = Future<FieldError?> Function(String value);

/// One value edited in place (house rule: 1–4 fields → an inline row, no
/// dialog). Shows [label], [display] and an [editLabel] button; editing
/// swaps them for the field, *Guardar* (never disabled; busy while saving)
/// and *Cancelar* (also Esc). Enter saves.
class InlineTextEditor extends StatefulWidget {
  /// An editor of one text value.
  const new({
    required this.label,
    required this.display,
    required this.editLabel,
    required this.onSave,
    required this.fieldBuilder,
    this.initialText = '',
    this.validator,
    this.canonicalize,
    this.canOpen,
    super.key,
  });

  /// What the value is.
  final String label;

  /// The value as shown while not editing.
  final Widget display;

  /// The button that starts editing.
  final String editLabel;

  /// Persists the value.
  final InlineSave onSave;

  /// The field widget (keyboard, capitalization, help) for its controller.
  final Widget Function(AppFieldController field) fieldBuilder;

  /// The field's starting text.
  final String initialText;

  /// The client check (format errors on blur; the server decides).
  final FieldValidator? validator;

  /// Rewrites the text on blur (e.g. an RFC upper-cased).
  final String Function(String value)? canonicalize;

  /// Asked before editing starts; `false` keeps the row closed (the caller
  /// shows why, e.g. the paywall).
  final bool Function()? canOpen;

  @override
  State<InlineTextEditor> createState() => _InlineTextEditorState();
}

class _InlineTextEditorState extends State<InlineTextEditor> {
  AppFormController? _form;
  AppFieldController? _field;

  void _open() {
    if (widget.canOpen?.call() == false) return;
    final form = AppFormController();
    final field = form.textField(
      label: widget.label,
      initialText: widget.initialText,
      validator: widget.validator,
      canonicalize: widget.canonicalize,
    );
    setState(() {
      _form = form;
      _field = field;
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) field.focusNode.requestFocus();
    });
  }

  void _close() {
    final form = _form;
    setState(() {
      _form = null;
      _field = null;
    });
    WidgetsBinding.instance.addPostFrameCallback((_) => form?.dispose());
  }

  Future<void> _save() async {
    final field = _field!;
    final error = await widget.onSave(field.value);
    if (!mounted || _field != field) return;
    if (error == null) return _close();
    _form!.showServerErrors({field: error});
  }

  void _submit() => unawaited(_form!.submit(_save));

  @override
  void dispose() {
    _form?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final form = _form;
    final field = _field;
    if (form == null || field == null) return _closedRow(context);
    final l10n = AppLocalizations.of(context);
    return AppFormScope(
      controller: form,
      submit: _submit,
      child: CallbackShortcuts(
        bindings: {const SingleActivator(LogicalKeyboardKey.escape): _close},
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            widget.fieldBuilder(field),
            const SizedBox(height: AppSpacing.xs),
            Wrap(
              spacing: AppSpacing.xs,
              children: [
                ListenableBuilder(
                  listenable: form,
                  builder: (context, _) => FilledButton(
                    onPressed: _submit,
                    child: form.isSaving
                        ? BusyLabel(
                            label: l10n.formSave,
                            announcement: l10n.formSaving,
                          )
                        : Text(l10n.formSave),
                  ),
                ),
                TextButton(onPressed: _close, child: Text(l10n.formCancel)),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _closedRow(BuildContext context) => Row(
    children: [
      Expanded(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(widget.label, style: Theme.of(context).textTheme.bodySmall),
            widget.display,
          ],
        ),
      ),
      TextButton(onPressed: _open, child: Text(widget.editLabel)),
    ],
  );
}
