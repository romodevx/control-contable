import 'package:control_contable/app/app.dart';
import 'package:control_contable/core/theme/app_theme.dart';
import 'package:control_contable/core/theme/directions/confianza_sobria.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

/// Pumps [form] inside the app's theme (direction A) and es-MX copy, without
/// the router or providers: form primitives depend on neither.
Future<void> pumpForm(
  WidgetTester tester,
  Widget form, {
  Brightness brightness = Brightness.light,
}) async {
  final tokens = brightness == Brightness.light
      ? confianzaSobria.light
      : confianzaSobria.dark;
  await tester.pumpWidget(
    MaterialApp(
      theme: buildAppTheme(tokens, brightness),
      locale: appLocale,
      supportedLocales: AppLocalizations.supportedLocales,
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      home: Scaffold(body: form),
    ),
  );
  await tester.pumpAndSettle();
}
