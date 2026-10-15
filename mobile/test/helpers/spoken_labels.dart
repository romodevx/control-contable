import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_test/flutter_test.dart';

/// How many times a screen reader says [text] as a whole line, across every
/// node of the app's semantics tree. Call with semantics enabled
/// (`tester.ensureSemantics()`).
///
/// A visible caption above a control that also carries the caption as its
/// `Semantics(label:)` is read twice: as its own node, or twice inside one
/// merged label ("Ejercicio\nEjercicio\n2026"). Both count here.
int timesSpoken(WidgetTester tester, String text) {
  var count = 0;
  void visit(SemanticsNode node) {
    count += node.label.split('\n').where((line) => line == text).length;
    node.visitChildren((child) {
      visit(child);
      return true;
    });
  }

  var root = tester.getSemantics(find.byType(MaterialApp).first);
  for (var up = root.parent; up != null; up = up.parent) {
    root = up;
  }
  visit(root);
  return count;
}
