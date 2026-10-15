import 'package:control_contable/features/annual/presentation/annual_year_picker.dart';
import 'package:control_contable/features/history/presentation/history_year_picker.dart';
import 'package:control_contable/features/tax_engine/presentation/estimate_period_picker.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/pump_form.dart';
import '../helpers/spoken_labels.dart';

/// iOS QA walk (VoiceOver, 2026-10-15): a picker with a visible caption
/// above it and the same caption as its `Semantics(label:)` was announced
/// twice ("Ejercicio fiscal, Ejercicio fiscal, 2026"). The field carries the
/// name; the caption is not read again.
void main() {
  final pickers = <String, (String, Widget)>{
    'annual year': (
      'Ejercicio',
      AnnualYearPicker(years: const [2026], selected: 2026, onSelected: (_) {}),
    ),
    'history year': (
      'Ejercicio fiscal',
      HistoryYearPicker(
        years: const [2026],
        selected: 2026,
        onSelected: (_) {},
      ),
    ),
    'estimate month': (
      'Mes',
      EstimatePeriodPicker(
        current: '2026-10',
        selected: '2026-09',
        onSelected: (_) {},
      ),
    ),
  };

  for (final MapEntry(key: name, value: (label, picker)) in pickers.entries) {
    testWidgets('$name: "$label" is announced once', (tester) async {
      final semantics = tester.ensureSemantics();
      await pumpForm(tester, picker);

      expect(find.text(label), findsOneWidget, reason: 'still visible');
      expect(timesSpoken(tester, label), 1);
      semantics.dispose();
    });
  }
}
