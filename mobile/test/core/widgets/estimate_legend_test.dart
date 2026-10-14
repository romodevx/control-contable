import 'package:control_contable/core/theme/directions/confianza_sobria.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

void main() {
  for (final brightness in Brightness.values) {
    testWidgets('the legend reads in full on its tinted block '
        '(${brightness.name})', (tester) async {
      await pumpForm(tester, const EstimateLegend(), brightness: brightness);

      expect(
        find.text(
          'Esto es un aproximado. No todos los gastos aplican para tu giro '
          'comercial.',
        ),
        findsOneWidget,
      );
      final tokens = brightness == Brightness.light
          ? confianzaSobria.light
          : confianzaSobria.dark;
      final box = tester.widget<DecoratedBox>(
        find.descendant(
          of: find.byType(EstimateLegend),
          matching: find.byType(DecoratedBox),
        ),
      );
      expect((box.decoration as BoxDecoration).color, tokens.colors.legendBg);
      await expectLater(tester, meetsGuideline(textContrastGuideline));
    });
  }
}
