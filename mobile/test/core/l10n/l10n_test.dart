import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';

void main() {
  testWidgets('the app runs in es-MX with Spanish framework strings', (
    tester,
  ) async {
    await pumpControlContableApp(tester);
    final context = tester.element(find.byType(HomeScreen));

    expect(Localizations.localeOf(context), const Locale('es', 'MX'));
    expect(MaterialLocalizations.of(context).cancelButtonLabel, 'Cancelar');
  });

  testWidgets('the app strings load from the es-MX ARB', (tester) async {
    await pumpControlContableApp(tester);
    final l10n = AppLocalizations.of(tester.element(find.byType(HomeScreen)));

    expect(l10n.appTitle, 'Control Contable');
    expect(l10n.navHome, 'Inicio');
    expect(
      l10n.estimateLegend,
      'Esto es un aproximado. No todos los gastos aplican para tu giro '
      'comercial.',
      reason: 'the mandatory legend is fixed copy (DIRECTIONS.md)',
    );
  });
}
