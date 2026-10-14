import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// Size of the leading icon.
const double _iconSize = 40;

/// What a flow shows instead of a form that could not be saved: why, what
/// still works, and the way out. It replaces the step the user was on, so
/// it is announced (a live region alone stays silent on first appearance
/// for TalkBack and VoiceOver).
class StopNotice extends StatefulWidget {
  /// The notice titled [title], explaining [body]; [actionLabel] runs
  /// [onLeave].
  const new({
    required this.icon,
    required this.title,
    required this.body,
    required this.actionLabel,
    required this.onLeave,
    super.key,
  });

  /// The leading icon.
  final IconData icon;

  /// Why the flow stops.
  final String title;

  /// What still works and what to do.
  final String body;

  /// The way out.
  final String actionLabel;

  /// Leaves the flow.
  final VoidCallback onLeave;

  @override
  State<StopNotice> createState() => _StopNoticeState();
}

class _StopNoticeState extends State<StopNotice> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      unawaited(
        SemanticsService.sendAnnouncement(
          View.of(context),
          '${widget.title}. ${widget.body}',
          Directionality.of(context),
        ),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Align(
            alignment: Alignment.centerLeft,
            child: Icon(
              widget.icon,
              size: _iconSize,
              color: context.tokens.colors.primary,
            ),
          ),
          const SizedBox(height: AppSpacing.md),
          Semantics(
            header: true,
            child: Text(widget.title, style: theme.textTheme.headlineMedium),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(widget.body, style: theme.textTheme.bodyMedium),
          const SizedBox(height: AppSpacing.xl),
          FilledButton(
            onPressed: widget.onLeave,
            child: Text(widget.actionLabel),
          ),
        ],
      ),
    );
  }
}
