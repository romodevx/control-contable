import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Placeholder home until the dashboard module (M4) ships.
class HomeScreen extends StatelessWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.navHome)),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: Text(
          l10n.homePlaceholder,
          style: Theme.of(context).textTheme.bodyMedium,
        ),
      ),
    );
  }
}
