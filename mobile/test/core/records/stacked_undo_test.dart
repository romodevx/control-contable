import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

/// PR #85 follow-up: a later plain toast must not close a pending Undo.
void main() {
  late ScaffoldMessengerState messenger;

  Future<void> pump(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Builder(
            builder: (context) {
              messenger = ScaffoldMessenger.of(context);
              return const SizedBox.shrink();
            },
          ),
        ),
      ),
    );
  }

  void push(StackedUndo<String> undo, List<List<String>> undone) => undo.push(
    'a',
    messenger: messenger,
    message: (count) => '$count borrados',
    undoLabel: 'Deshacer',
    onUndo: undone.add,
  );

  testWidgets('a plain toast waits behind a pending Undo, which stays', (
    tester,
  ) async {
    await pump(tester);
    final undo = StackedUndo<String>();
    final undone = <List<String>>[];
    push(undo, undone);
    await tester.pumpAndSettle();

    undo.notify(messenger, 'No se pudo eliminar.');
    await tester.pumpAndSettle();

    expect(find.text('1 borrados'), findsOneWidget);
    expect(find.text('No se pudo eliminar.'), findsNothing);
    expect(undo.pending, ['a']);

    await tester.tap(find.text('Deshacer'));
    await tester.pumpAndSettle();

    expect(undone, [
      ['a'],
    ]);
    expect(find.text('No se pudo eliminar.'), findsOneWidget);
  });

  testWidgets('with nothing pending, a plain toast replaces the current one', (
    tester,
  ) async {
    await pump(tester);
    final undo = StackedUndo<String>();
    messenger.showSnackBar(
      const SnackBar(content: Text('Antes'), persist: true),
    );
    await tester.pumpAndSettle();

    undo.notify(messenger, 'Después');
    await tester.pumpAndSettle();

    expect(find.text('Antes'), findsNothing);
    expect(find.text('Después'), findsOneWidget);
  });

  test('restoreEach tries every item and reports what came back', () async {
    final tried = <String>[];
    final report = await restoreEach<String, Exception>(['a', 'b', 'c'], (
      item,
    ) async {
      tried.add(item);
      if (item == 'b') throw Exception('refused');
    });

    expect(tried, ['a', 'b', 'c']);
    expect(report.restored, 2);
    expect(report.refusals, hasLength(1));
  });
}
