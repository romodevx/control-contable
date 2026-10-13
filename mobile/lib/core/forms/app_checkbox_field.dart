import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Width the checkbox column takes, so [AppCheckboxField.below] lines up
/// with the label text.
const double _boxColumn = AppSizes.minTarget + AppSpacing.xxs;

/// A labeled checkbox inside an `AppForm` (a consent, an opt-in). The whole
/// row toggles it and is at least a 44 px target; the label is the box's
/// accessible name. Its error shows under the row, and the error summary
/// lists it like any other field.
class AppCheckboxField extends StatelessWidget {
  /// A checkbox for [field], with [below] (e.g. a document link) under the
  /// label.
  const new({required this.field, this.below, super.key});

  /// Finds the tappable row (tests).
  static const ValueKey<String> rowKey = ValueKey('AppCheckboxField.row');

  /// The box's state and error.
  final AppCheckboxController field;

  /// Shown under the label, aligned with it.
  final Widget? below;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: field,
      builder: (context, _) {
        final error = field.error;
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            InkWell(
              key: rowKey,
              // The checkbox below carries the semantics; the row only
              // widens the target.
              excludeFromSemantics: true,
              onTap: () => field.checked = !field.checked,
              child: ConstrainedBox(
                constraints: const BoxConstraints(
                  minHeight: AppSizes.minTarget,
                ),
                child: Row(
                  children: [
                    SizedBox(
                      width: _boxColumn,
                      child: Checkbox(
                        value: field.checked,
                        focusNode: field.focusNode,
                        isError: error != null,
                        semanticLabel: field.label,
                        onChanged: (value) => field.checked = value ?? false,
                      ),
                    ),
                    Expanded(
                      child: ExcludeSemantics(
                        child: Text(
                          field.label,
                          style: theme.textTheme.bodyLarge,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            if (error != null)
              Padding(
                padding: const EdgeInsets.only(left: _boxColumn),
                child: Semantics(
                  liveRegion: true,
                  child: Text(
                    describeFieldError(error, l10n),
                    style: theme.textTheme.bodySmall?.copyWith(
                      color: context.tokens.colors.error,
                    ),
                  ),
                ),
              ),
            if (below case final below?)
              Padding(
                padding: const EdgeInsets.only(left: _boxColumn),
                child: Align(alignment: Alignment.centerLeft, child: below),
              ),
          ],
        );
      },
    );
  }
}
