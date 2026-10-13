import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

const _terms = 'Acepto los términos';
const _notAccepted = 'Marca esta casilla para continuar.';

void main() {
  late AppFormController form;
  late AppFieldController name;
  late AppCheckboxController terms;
  late int saves;

  setUp(() {
    form = AppFormController();
    name = form.textField(label: 'Nombre', isRequired: false);
    terms = form.checkboxField(label: _terms);
    saves = 0;
  });
  tearDown(() => form.dispose());

  group('AppCheckboxController', () {
    test('a required box left unchecked fails only on submit', () async {
      expect(terms.error, isNull);

      final saved = await form.submit(() async => saves++);

      expect(saved, isFalse);
      expect(terms.error, isA<NotAccepted>());
      expect(form.invalidFields, [terms]);
    });

    test('checking it clears the error at once', () async {
      await form.submit(() async {});

      terms.checked = true;

      expect(terms.error, isNull);
      expect(terms.isDirty, isTrue);
    });

    test('an optional box never blocks a submit', () async {
      final news = form.checkboxField(label: 'Noticias', isRequired: false);
      terms.checked = true;

      expect(await form.submit(() async => saves++), isTrue);
      expect(news.error, isNull);
      expect(saves, 1);
    });

    test('reset unchecks it and drops the error', () async {
      await form.submit(() async {});
      terms
        ..checked = true
        ..reset();

      expect(terms.checked, isFalse);
      expect(terms.error, isNull);
    });

    test('a server error shows until the box changes', () {
      terms.checked = true;
      form.showServerErrors({terms: const CustomFieldError('Vencido')});
      expect(terms.error, isA<CustomFieldError>());

      terms.checked = false;
      expect(terms.error, isA<NotAccepted>(), reason: 're-checked');
    });
  });

  group('AppCheckboxField', () {
    Future<void> pumpConsentForm(WidgetTester tester) => pumpForm(
      tester,
      AppForm(
        controller: form,
        onSave: () async => saves++,
        children: [
          AppTextField(
            field: name,
            keyboardType: TextInputType.name,
            textInputAction: TextInputAction.next,
            autofillHints: const [],
          ),
          AppCheckboxField(
            field: terms,
            below: TextButton(onPressed: () {}, child: const Text('Leer')),
          ),
        ],
      ),
    );

    testWidgets('tapping the label toggles the box', (tester) async {
      await pumpConsentForm(tester);

      await tester.tap(find.text(_terms));
      await tester.pump();
      expect(terms.checked, isTrue);

      await tester.tap(find.text(_terms));
      await tester.pump();
      expect(terms.checked, isFalse);
    });

    testWidgets('renders the slot below it (a document link)', (tester) async {
      await pumpConsentForm(tester);

      expect(find.widgetWithText(TextButton, 'Leer'), findsOneWidget);
    });

    testWidgets('the row is at least a 44 px target', (tester) async {
      await pumpConsentForm(tester);

      final row = tester.getSize(find.byKey(AppCheckboxField.rowKey));
      expect(row.height, greaterThanOrEqualTo(AppSizes.minTarget));
    });

    testWidgets('is one checkbox node labeled by its text', (tester) async {
      final semantics = tester.ensureSemantics();
      await pumpConsentForm(tester);

      expect(tester.getSemantics(find.byType(Checkbox)).label, _terms);
      expect(find.bySemanticsLabel(_terms), findsOneWidget, reason: 'once');
      semantics.dispose();
    });

    testWidgets('a failed submit lists it in the summary and inline', (
      tester,
    ) async {
      await pumpConsentForm(tester);

      await tester.tap(find.widgetWithText(FilledButton, 'Guardar'));
      await tester.pumpAndSettle();

      expect(find.text('$_terms: $_notAccepted'), findsOneWidget);
      expect(find.text(_notAccepted), findsOneWidget, reason: 'inline');
      expect(saves, 0);

      await tester.tap(find.text('$_terms: $_notAccepted'));
      await tester.pumpAndSettle();
      expect(terms.focusNode.hasFocus, isTrue);
    });
  });
}
