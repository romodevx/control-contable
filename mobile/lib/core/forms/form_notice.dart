import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';

/// Width of the notice's leading rule (matches the error summary).
const _ruleWidth = 3.0;

/// Whether a [FormNotice] reports a problem or confirms something.
enum FormNoticeTone {
  /// Something stopped the action (offline, too many tries).
  error,

  /// The action went through (a code was sent).
  info,
}

/// A form-level message that belongs to no single field: a failure the
/// user can act on, or a confirmation. Announced when it appears; tone is
/// carried by the rule color *and* the words, never by color alone.
class FormNotice extends StatelessWidget {
  /// A notice reading [message].
  const new({
    required this.message,
    this.tone = FormNoticeTone.error,
    super.key,
  });

  /// What the user reads.
  final String message;

  /// Error or confirmation.
  final FormNoticeTone tone;

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    final rule = tone == FormNoticeTone.error ? colors.error : colors.primary;
    return Semantics(
      liveRegion: true,
      container: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.surface,
          border: Border(
            left: BorderSide(color: rule, width: _ruleWidth),
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.md),
          child: Text(message, style: Theme.of(context).textTheme.bodyMedium),
        ),
      ),
    );
  }
}
