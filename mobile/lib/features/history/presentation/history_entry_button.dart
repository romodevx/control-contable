import 'dart:async';

import 'package:control_contable/features/history/history_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Opens History from the app bar. Gate it with
/// `FlagGate(module: AppModule.history, …)` (`useFlag`).
class HistoryEntryButton extends StatelessWidget {
  /// Creates the button; [onReturn] runs when the user comes back, so the
  /// screen below can read again what History's links may have changed.
  const new({this.onReturn, super.key});

  /// Called when History is popped (records may have been added or edited
  /// from it).
  final VoidCallback? onReturn;

  @override
  Widget build(BuildContext context) => IconButton(
    tooltip: AppLocalizations.of(context).historyTitle,
    icon: const Icon(Icons.history_rounded),
    onPressed: () => unawaited(
      context
          .push<void>(HistoryRoutes.list)
          .whenComplete(() => onReturn?.call()),
    ),
  );
}
