import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

export 'package:control_contable/core/forms/choice_controllers.dart'
    show ChoiceOption;

/// One option of a [ChoiceFieldController] per row (a radio group inside an
/// `AppForm`): the label above, every row a 44 px target, the error under
/// the group and in the form's error summary.
class AppChoiceField extends StatelessWidget {
  /// The group for [field] offering [options].
  const new({
    required this.field,
    required this.options,
    this.helperText,
    super.key,
  });

  /// The chosen value and its error.
  final ChoiceFieldController field;

  /// The options, in reading order.
  final List<ChoiceOption> options;

  /// A short explanation under the label.
  final String? helperText;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: field,
    builder: (context, _) => _ChoiceGroup(
      field: field,
      helperText: helperText,
      child: RadioGroup<String>(
        groupValue: field.value,
        onChanged: (value) => field.value = value,
        child: Column(
          children: [
            for (final (index, option) in options.indexed)
              RadioListTile<String>(
                value: option.value,
                focusNode: index == 0 ? field.focusNode : null,
                contentPadding: EdgeInsets.zero,
                title: Text(option.label),
              ),
          ],
        ),
      ),
    ),
  );
}

/// One checkbox per option of a [MultiChoiceFieldController] inside an
/// `AppForm`; the same layout as [AppChoiceField].
class AppMultiChoiceField extends StatelessWidget {
  /// The group for [field] offering [options].
  const new({
    required this.field,
    required this.options,
    this.helperText,
    super.key,
  });

  /// The chosen values and their error.
  final MultiChoiceFieldController field;

  /// The options, in reading order.
  final List<ChoiceOption> options;

  /// A short explanation under the label.
  final String? helperText;

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: field,
    builder: (context, _) => _ChoiceGroup(
      field: field,
      helperText: helperText,
      child: Column(
        children: [
          for (final (index, option) in options.indexed)
            CheckboxListTile(
              value: field.values.contains(option.value),
              focusNode: index == 0 ? field.focusNode : null,
              isError: field.error != null,
              controlAffinity: ListTileControlAffinity.leading,
              contentPadding: EdgeInsets.zero,
              title: Text(option.label),
              onChanged: (_) => field.toggle(option.value),
            ),
        ],
      ),
    ),
  );
}

/// The label, optional help, the options and the error of a choice group,
/// read by screen readers as one labeled group.
class _ChoiceGroup extends StatelessWidget {
  const new({
    required this.field,
    required this.helperText,
    required this.child,
  });

  final AppFormFieldModel field;
  final String? helperText;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context);
    final colors = context.tokens.colors;
    final error = field.error;
    return Semantics(
      container: true,
      explicitChildNodes: true,
      label: field.label,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          ExcludeSemantics(
            child: Text(
              field.label,
              style: theme.textTheme.bodyMedium?.copyWith(
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          if (helperText case final help?) ...[
            const SizedBox(height: AppSpacing.xxs),
            Text(
              help,
              style: theme.textTheme.bodySmall?.copyWith(color: colors.muted),
            ),
          ],
          const SizedBox(height: AppSpacing.xxs),
          child,
          if (error != null)
            Semantics(
              liveRegion: true,
              child: Text(
                describeFieldError(error, l10n),
                style: theme.textTheme.bodySmall?.copyWith(color: colors.error),
              ),
            ),
        ],
      ),
    );
  }
}
