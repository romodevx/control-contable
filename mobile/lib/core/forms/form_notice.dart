import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

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
/// user can act on, or a confirmation. Announced to screen readers when it
/// appears and whenever its message changes (assertively for an error,
/// politely for a confirmation): a live region alone is silent on VoiceOver
/// and when the text repeats. Tone is carried by the rule color *and* the
/// words, never by color alone.
class FormNotice extends StatefulWidget {
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
  State<FormNotice> createState() => _FormNoticeState();
}

class _FormNoticeState extends State<FormNotice> {
  @override
  void initState() {
    super.initState();
    _announceAfterFrame();
  }

  @override
  void didUpdateWidget(FormNotice oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.message != oldWidget.message) _announceAfterFrame();
  }

  void _announceAfterFrame() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(
        SemanticsService.sendAnnouncement(
          View.of(context),
          widget.message,
          Directionality.of(context),
          assertiveness: widget.tone == FormNoticeTone.error
              ? Assertiveness.assertive
              : Assertiveness.polite,
        ),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.tokens.colors;
    final tone = widget.tone;
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
          child: Text(
            widget.message,
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ),
      ),
    );
  }
}
