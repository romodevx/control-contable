import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Shown for any location no route matches (including a disabled module's).
class NotFoundScreen extends StatelessWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.notFoundTitle)),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: Align(
          alignment: Alignment.topLeft,
          child: FilledButton(
            onPressed: () => context.go(AppRoutes.home),
            child: Text(l10n.notFoundAction),
          ),
        ),
      ),
    );
  }
}
