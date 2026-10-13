import 'dart:async';

import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// Width of the error rule on the summary's leading edge.
const _ruleWidth = 3.0;

/// The error summary at the top of a form: after a failed submit it lists
/// every field with an error, takes focus, is announced to screen readers
/// (every time: input focus does not move TalkBack/VoiceOver focus, and a
/// live region stays silent when the count repeats), and each item moves
/// focus to its field. It hides once every error is
/// fixed and comes back, focused, on the next failed submit.
class FormErrorSummary extends StatefulWidget {
  /// The summary of [controller]'s errors.
  const new({required this.controller, super.key});

  /// The form whose errors are listed.
  final AppFormController controller;

  @override
  State<FormErrorSummary> createState() => _FormErrorSummaryState();
}

class _FormErrorSummaryState extends State<FormErrorSummary> {
  final _focusNode = FocusNode(debugLabel: 'FormErrorSummary');
  late int _seenFailures = widget.controller.submitFailures;

  @override
  void initState() {
    super.initState();
    widget.controller.addListener(_handleFormChanged);
  }

  void _handleFormChanged() {
    final failures = widget.controller.submitFailures;
    if (failures == _seenFailures) return;
    _seenFailures = failures;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      _announce();
      _focusNode.requestFocus();
      final summary = _focusNode.context;
      if (summary != null) Scrollable.ensureVisible(summary);
    });
  }

  void _announce() {
    final invalid = widget.controller.invalidFields.length;
    if (invalid == 0) return;
    unawaited(
      SemanticsService.sendAnnouncement(
        View.of(context),
        AppLocalizations.of(context).formErrorSummaryTitle(invalid),
        Directionality.of(context),
        assertiveness: Assertiveness.assertive,
      ),
    );
  }

  @override
  void dispose() {
    widget.controller.removeListener(_handleFormChanged);
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => ListenableBuilder(
    listenable: widget.controller,
    builder: (context, _) {
      final invalid = widget.controller.invalidFields;
      if (widget.controller.submitFailures == 0 || invalid.isEmpty) {
        return const SizedBox.shrink();
      }
      final l10n = AppLocalizations.of(context);
      final colors = context.tokens.colors;
      return Padding(
        padding: const EdgeInsets.only(bottom: AppSpacing.xl),
        child: Focus(
          focusNode: _focusNode,
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: colors.surface,
              border: Border(
                left: BorderSide(color: colors.error, width: _ruleWidth),
              ),
            ),
            child: Padding(
              padding: const EdgeInsets.all(AppSpacing.md),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Semantics(
                    header: true,
                    child: Text(
                      l10n.formErrorSummaryTitle(invalid.length),
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  const SizedBox(height: AppSpacing.xs),
                  for (final field in invalid)
                    TextButton(
                      style: TextButton.styleFrom(
                        alignment: Alignment.centerLeft,
                        foregroundColor: colors.error,
                        padding: EdgeInsets.zero,
                      ),
                      onPressed: field.focusNode.requestFocus,
                      child: Text(
                        l10n.formErrorSummaryItem(
                          field.label,
                          describeFieldError(field.error!, l10n),
                        ),
                        style: const TextStyle(
                          decoration: TextDecoration.underline,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ),
      );
    },
  );
}
