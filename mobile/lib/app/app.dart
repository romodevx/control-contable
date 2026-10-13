import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/theme/app_theme.dart';
import 'package:control_contable/core/theme/design_direction.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The app's only locale.
const appLocale = Locale('es', 'MX');

/// Root widget: router, light/dark theme from the active design direction,
/// es-MX localization. Follows the system brightness.
class ControlContableApp extends ConsumerWidget {
  /// Creates the root widget.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    final direction = ref.watch(designDirectionProvider);
    return MaterialApp.router(
      onGenerateTitle: (context) => AppLocalizations.of(context).appTitle,
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(direction.light, Brightness.light),
      darkTheme: buildAppTheme(direction.dark, Brightness.dark),
      locale: appLocale,
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      routerConfig: router,
    );
  }
}
