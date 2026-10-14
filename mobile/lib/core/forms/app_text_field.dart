import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_scope.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// IME actions that mean "I am done": they submit the form.
const Set<TextInputAction> _submitActions = {
  TextInputAction.done,
  TextInputAction.go,
  TextInputAction.send,
};

/// A labeled text field that follows the house form rules: the label above
/// the field (never Material's floating label, never a placeholder), the
/// requirement marked in text, the error under the field and announced with
/// it. The keyboard type, IME action and autofill hints are required so no
/// screen forgets them.
///
/// A single-line field whose action is done/go/send submits the form; `next`
/// moves on. A multi-line field never submits on Enter (newline).
class AppTextField extends StatelessWidget {
  /// A field for [field] inside an `AppForm`.
  const new({
    required this.field,
    required this.keyboardType,
    required this.textInputAction,
    required this.autofillHints,
    this.maxLines = 1,
    this.hintText,
    this.prefixText,
    this.inputFormatters,
    this.helperText,
    this.obscureText = false,
    this.suffixIcon,
    this.onChanged,
    this.textCapitalization = TextCapitalization.none,
    this.autocorrect = true,
    super.key,
  });

  /// The field's text, focus and error.
  final AppFieldController field;

  /// The keyboard the platform opens.
  final TextInputType keyboardType;

  /// The IME action key.
  final TextInputAction textInputAction;

  /// Autofill hints (`AutofillHints`); empty when nothing applies.
  final Iterable<String> autofillHints;

  /// More than one makes it a multi-line field.
  final int maxLines;

  /// An example inside the empty field. Never a label.
  final String? hintText;

  /// Fixed text before the value, such as `$ `.
  final String? prefixText;

  /// Restricts what can be typed.
  final List<TextInputFormatter>? inputFormatters;

  /// A short rule under the field, such as the minimum length.
  final String? helperText;

  /// Hides the text (a password); also turns suggestions and autocorrect off
  /// so the keyboard never learns it.
  final bool obscureText;

  /// A control at the end of the field, such as a show-password toggle.
  final Widget? suffixIcon;

  /// Called with the raw text on every change (e.g. submit a complete code).
  final ValueChanged<String>? onChanged;

  /// Which letters the keyboard capitalizes (`characters` for a code such
  /// as an RFC).
  final TextCapitalization textCapitalization;

  /// Whether the keyboard may correct and suggest; `false` for codes and
  /// identifiers, which no dictionary knows.
  final bool autocorrect;

  @override
  Widget build(BuildContext context) {
    final scope = AppFormScope.of(context);
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    return ListenableBuilder(
      listenable: field,
      builder: (context, _) {
        final error = field.error;
        final label = _labelText(scope.controller.marks, l10n);
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Read once, as the field's own label (below).
            ExcludeSemantics(
              child: Text(
                label,
                style: theme.textTheme.bodyMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
            const SizedBox(height: AppSpacing.xs),
            Semantics(
              label: label,
              child: TextField(
                controller: field.text,
                focusNode: field.focusNode,
                keyboardType: keyboardType,
                textInputAction: textInputAction,
                autofillHints: autofillHints,
                maxLines: maxLines,
                inputFormatters: inputFormatters,
                obscureText: obscureText,
                textCapitalization: textCapitalization,
                enableSuggestions: autocorrect && !obscureText,
                autocorrect: autocorrect && !obscureText,
                onChanged: onChanged,
                style: theme.textTheme.bodyLarge,
                scrollPadding: const EdgeInsets.fromLTRB(
                  AppSpacing.lg,
                  AppSpacing.lg,
                  AppSpacing.lg,
                  formFooterClearance + AppSpacing.lg,
                ),
                decoration: InputDecoration(
                  hintText: hintText,
                  prefixText: prefixText,
                  helperText: helperText,
                  helperMaxLines: 3,
                  suffixIcon: suffixIcon,
                  errorText: error == null
                      ? null
                      : describeFieldError(error, l10n),
                  errorMaxLines: 3,
                ),
                onSubmitted: (_) {
                  if (maxLines == 1 &&
                      _submitActions.contains(textInputAction)) {
                    scope.submit();
                  }
                },
              ),
            ),
          ],
        );
      },
    );
  }

  String _labelText(RequirementMark marks, AppLocalizations l10n) {
    final mark = switch (marks) {
      RequirementMark.optionalFields when !field.isRequired =>
        l10n.formMarkOptional,
      RequirementMark.requiredFields when field.isRequired =>
        l10n.formMarkRequired,
      _ => null,
    };
    return mark == null ? field.label : '${field.label} $mark';
  }
}
