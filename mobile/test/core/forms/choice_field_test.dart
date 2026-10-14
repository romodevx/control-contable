import 'package:control_contable/core/forms/app_choice_fields.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

const _choose = 'Elige una opción.';

const List<ChoiceOption> _modes = [
  (value: 'blind_35', label: 'Deducción ciega'),
  (value: 'actual', label: 'Gastos reales'),
];

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  Future<void> pumpFields(WidgetTester tester, List<Widget> fields) => pumpForm(
    tester,
    AppForm(controller: form, onSave: () async {}, children: fields),
  );

  group('ChoiceFieldController', () {
    test('a required choice left empty fails only on submit', () async {
      final mode = form.add(ChoiceFieldController(label: 'Modo'));
      expect(mode.error, isNull);

      expect(await form.submit(() async {}), isFalse);
      expect(mode.error, isA<MissingChoice>());
    });

    test('choosing clears a shown error', () async {
      final mode = form.add(ChoiceFieldController(label: 'Modo'));
      await form.submit(() async {});

      mode.value = 'actual';

      expect(mode.error, isNull);
    });

    test('an initial value is not dirty; a change is; reset empties', () {
      final mode = form.add(
        ChoiceFieldController(label: 'Modo', initialValue: 'blind_35'),
      );
      expect(mode.isDirty, isFalse);

      mode.value = 'actual';
      expect(mode.isDirty, isTrue);

      mode.reset();
      expect(mode.value, isNull);
    });

    test('an optional choice may stay empty', () async {
      form.add(ChoiceFieldController(label: 'Modo', isRequired: false));

      expect(await form.submit(() async {}), isTrue);
    });
  });

  group('MultiChoiceFieldController', () {
    test('a required set needs at least one value', () async {
      final kinds = form.add(MultiChoiceFieldController(label: 'Actividades'));

      expect(await form.submit(() async {}), isFalse);
      expect(kinds.error, isA<MissingChoice>());

      kinds.toggle('lodging');
      expect(kinds.error, isNull);
      expect(kinds.values, {'lodging'});
      expect(kinds.isDirty, isTrue);

      kinds.toggle('lodging');
      expect(kinds.values, isEmpty);
      expect(kinds.error, isNull, reason: 'a cleared error waits for submit');
    });

    test('reset empties it', () {
      final kinds = form.add(
        MultiChoiceFieldController(label: 'A', initialValues: {'x'}),
      )..reset();

      expect(kinds.values, isEmpty);
      expect(kinds.isDirty, isTrue);
    });
  });

  group('AppChoiceField', () {
    testWidgets('tapping an option row selects it (44 px rows)', (
      tester,
    ) async {
      final mode = form.add(ChoiceFieldController(label: 'Modo'));
      await pumpFields(tester, [AppChoiceField(field: mode, options: _modes)]);

      expect(find.text('Modo'), findsOneWidget);
      await tester.tap(find.text('Gastos reales'));
      await tester.pumpAndSettle();

      expect(mode.value, 'actual');
      final row = tester.getSize(
        find.ancestor(
          of: find.text('Gastos reales'),
          matching: find.byType(RadioListTile<String>),
        ),
      );
      expect(row.height, greaterThanOrEqualTo(AppSizes.minTarget));
    });

    testWidgets('a failed submit shows the error inline and in the summary', (
      tester,
    ) async {
      final mode = form.add(ChoiceFieldController(label: 'Modo'));
      await pumpFields(tester, [AppChoiceField(field: mode, options: _modes)]);

      await tester.tap(find.text('Guardar'));
      await tester.pumpAndSettle();

      expect(find.text(_choose), findsOneWidget);
      expect(find.textContaining('Modo: $_choose'), findsOneWidget);
    });

    testWidgets('shows a help text under the label', (tester) async {
      final mode = form.add(ChoiceFieldController(label: 'Modo'));
      await pumpFields(tester, [
        AppChoiceField(field: mode, options: _modes, helperText: 'Ayuda'),
      ]);

      expect(find.text('Ayuda'), findsOneWidget);
    });
  });

  group('AppMultiChoiceField', () {
    testWidgets('each option toggles its value', (tester) async {
      final kinds = form.add(MultiChoiceFieldController(label: 'Actividades'));
      await pumpFields(tester, [
        AppMultiChoiceField(field: kinds, options: _modes),
      ]);

      await tester.tap(find.text('Deducción ciega'));
      await tester.tap(find.text('Gastos reales'));
      await tester.pumpAndSettle();
      expect(kinds.values, {'blind_35', 'actual'});

      await tester.tap(find.text('Gastos reales'));
      await tester.pumpAndSettle();
      expect(kinds.values, {'blind_35'});
    });

    testWidgets('a failed submit shows the error under the group', (
      tester,
    ) async {
      final kinds = form.add(MultiChoiceFieldController(label: 'Actividades'));
      await pumpFields(tester, [
        AppMultiChoiceField(field: kinds, options: _modes),
      ]);

      await tester.tap(find.text('Guardar'));
      await tester.pumpAndSettle();

      expect(find.text(_choose), findsOneWidget);
    });
  });
}
