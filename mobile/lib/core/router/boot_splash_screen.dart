import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Shown while a location inside a module waits for the first flags answer
/// (see `RouterFlags`). The router leaves it on its own once flags resolve.
class BootSplashScreen extends StatelessWidget {
  /// Creates the splash.
  const new({super.key});

  @override
  Widget build(BuildContext context) {
    final label = AppLocalizations.of(context).bootLoading;
    return Scaffold(
      body: Center(
        child: Semantics(
          liveRegion: true,
          label: label,
          child: const ExcludeSemantics(child: CircularProgressIndicator()),
        ),
      ),
    );
  }
}
