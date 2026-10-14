import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// Size of the lock icon.
const double _iconSize = 40;

/// What a read-only user (`402 entitlement_required`: trial over, no plan)
/// sees instead of the create: why, what still works, and the way out.
/// Nothing was created; no form is offered again. It replaces the step the
/// user was on, so it is announced (a live region alone stays silent on
/// first appearance for TalkBack and VoiceOver).
class ReadOnlyNotice extends StatefulWidget {
  /// The notice; [onLeave] closes the flow. [body] says what needs a plan
  /// (registering a taxpayer by default).
  const new({required this.onLeave, this.body, this.actionLabel, super.key});

  /// Leaves the flow.
  final VoidCallback onLeave;

  /// What still works and what needs a plan.
  final String? body;

  /// The way out (*Volver al inicio* by default).
  final String? actionLabel;

  @override
  State<ReadOnlyNotice> createState() => _ReadOnlyNoticeState();
}

class _ReadOnlyNoticeState extends State<ReadOnlyNotice> {
  String _body(AppLocalizations l10n) =>
      widget.body ?? l10n.onboardingPaywallBody;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      final l10n = AppLocalizations.of(context);
      unawaited(
        SemanticsService.sendAnnouncement(
          View.of(context),
          '${l10n.onboardingPaywallTitle}. ${_body(l10n)}',
          Directionality.of(context),
        ),
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Align(
            alignment: Alignment.centerLeft,
            child: Icon(
              Icons.lock_outline,
              size: _iconSize,
              color: context.tokens.colors.primary,
            ),
          ),
          const SizedBox(height: AppSpacing.md),
          Semantics(
            header: true,
            child: Text(
              l10n.onboardingPaywallTitle,
              style: theme.textTheme.headlineMedium,
            ),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(_body(l10n), style: theme.textTheme.bodyMedium),
          const SizedBox(height: AppSpacing.xl),
          FilledButton(
            onPressed: widget.onLeave,
            child: Text(widget.actionLabel ?? l10n.onboardingPaywallAction),
          ),
        ],
      ),
    );
  }
}
