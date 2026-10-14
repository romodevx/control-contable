import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The new-income page (M3-09).
class IncomeFormScreen extends StatelessWidget {
  /// Creates the page.
  const new({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(AppLocalizations.of(context).incomeFormTitle)),
  );
}
