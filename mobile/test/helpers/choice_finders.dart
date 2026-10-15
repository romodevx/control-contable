import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

/// The radio group (an `AppChoiceField`) labeled [label].
Finder groupOf(String label) =>
    find.ancestor(of: find.text(label), matching: find.byType(Column)).first;

/// The value chosen in the group labeled [label], or `null`.
String? chosenIn(WidgetTester tester, String label) => tester
    .widget<RadioGroup<String>>(
      find.descendant(
        of: groupOf(label),
        matching: find.byType(RadioGroup<String>),
      ),
    )
    .groupValue;

/// Taps [option] inside the group labeled [label].
Future<void> chooseIn(WidgetTester tester, String label, String option) async {
  final target = find.descendant(
    of: groupOf(label),
    matching: find.text(option),
  );
  await tester.ensureVisible(target);
  await tester.pumpAndSettle();
  await tester.tap(target);
  await tester.pumpAndSettle();
}
