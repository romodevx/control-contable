import 'dart:async';

import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

/// Opens a paywall message (`402 entitlement_required`): [title] says what
/// stopped the write, [body] what still works and the way out. Announced
/// when it opens. What is behind it (a form) keeps its data.
Future<void> showPaywallSheet(
  BuildContext context, {
  required String title,
  required String body,
}) => showModalBottomSheet<void>(
  context: context,
  showDragHandle: true,
  useSafeArea: true,
  builder: (_) => _PaywallSheet(title: title, body: body),
);

class _PaywallSheet extends StatefulWidget {
  const new({required this.title, required this.body});

  final String title;
  final String body;

  @override
  State<_PaywallSheet> createState() => _PaywallSheetState();
}

class _PaywallSheetState extends State<_PaywallSheet> {
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
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.gutter,
        0,
        AppSpacing.gutter,
        AppSpacing.gutter,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Semantics(
            header: true,
            child: Text(widget.title, style: theme.textTheme.headlineSmall),
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(widget.body, style: theme.textTheme.bodyMedium),
          const SizedBox(height: AppSpacing.xl),
          FilledButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(l10n.paywallClose),
          ),
        ],
      ),
    );
  }
}
