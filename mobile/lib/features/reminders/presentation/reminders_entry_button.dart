import 'dart:async';

import 'package:control_contable/features/reminders/reminders_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Opens the reminder settings from the app bar. Gate it with
/// `FlagGate(module: AppModule.reminders, …)` (`useFlag`).
class RemindersEntryButton extends StatelessWidget {
  /// Creates the button.
  const new({super.key});

  @override
  Widget build(BuildContext context) => IconButton(
    tooltip: AppLocalizations.of(context).remindersTitle,
    icon: const Icon(Icons.notifications_outlined),
    onPressed: () => unawaited(context.push<void>(RemindersRoutes.settings)),
  );
}
