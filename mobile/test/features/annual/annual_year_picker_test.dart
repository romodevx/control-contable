import 'package:control_contable/features/annual/presentation/annual_year_picker.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

/// Every label a screen reader reads inside the picker.
List<String> _spokenLabels(WidgetTester tester) {
  final labels = <String>[];
  void visit(SemanticsNode node) {
    if (node.label.isNotEmpty) labels.add(node.label);
    node.visitChildren((child) {
      visit(child);
      return true;
    });
  }

  visit(tester.getSemantics(find.byType(AnnualYearPicker)));
  return labels;
}

void main() {
  // iOS QA walk (VoiceOver, 2026-10-15): the picker was read as
  // "Ejercicio, Ejercicio, 2026": the visible label above the field and the
  // field's own Semantics label were both announced.
  testWidgets('the year picker names its field once', (tester) async {
    final semantics = tester.ensureSemantics();
    await pumpForm(
      tester,
      AnnualYearPicker(years: const [2026], selected: 2026, onSelected: (_) {}),
    );

    final spoken = _spokenLabels(tester).join('\n');

    expect(
      RegExp('Ejercicio').allMatches(spoken),
      hasLength(1),
      reason: spoken,
    );
    expect(
      tester.getSemantics(find.byKey(const ValueKey('annual-year'))),
      containsSemantics(label: 'Ejercicio\n2026', isButton: true),
    );
    semantics.dispose();
  });
}
