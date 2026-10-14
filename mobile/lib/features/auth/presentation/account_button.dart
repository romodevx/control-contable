import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// The way into *Mi cuenta* from the app bar: an icon with its name as the
/// tooltip and the screen-reader label.
class AccountButton extends StatelessWidget {
  /// Creates the button.
  const new({super.key});

  @override
  Widget build(BuildContext context) => IconButton(
    tooltip: AppLocalizations.of(context).accountTitle,
    icon: const Icon(Icons.account_circle_outlined),
    onPressed: () => context.push<void>(AuthRoutes.profile),
  );
}
