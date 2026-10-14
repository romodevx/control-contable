import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// A route whose screen is still being built (the breakdown, M5-26, and
/// the tax-inputs form, M5-27): the estimate's links already land here.
class ComingSoonScreen extends StatelessWidget {
  /// The placeholder titled [title].
  const new({required this.title, super.key});

  /// The app bar title of the future screen.
  final String title;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(title)),
    body: Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: Text(
        AppLocalizations.of(context).comingSoon,
        style: Theme.of(context).textTheme.bodyLarge,
      ),
    ),
  );
}
