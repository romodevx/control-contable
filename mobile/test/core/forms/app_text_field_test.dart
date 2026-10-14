import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  Future<void> pumpFields(WidgetTester tester, List<Widget> fields) => pumpForm(
    tester,
    AppForm(controller: form, onSave: () async {}, children: fields),
  );

  TextField textFieldOf(WidgetTester tester, String label) => tester.widget(
    find.descendant(
      of: find.byKey(ValueKey(label)),
      matching: find.byType(TextField),
    ),
  );

  testWidgets('the label sits above the field, never inside it', (
    tester,
  ) async {
    final rfc = form.textField(label: 'RFC');
    await pumpFields(tester, [
      AppTextField(
        key: const ValueKey('RFC'),
        field: rfc,
        keyboardType: TextInputType.text,
        textInputAction: TextInputAction.done,
        autofillHints: const [],
      ),
    ]);

    final label = tester.getRect(find.text('RFC'));
    final input = tester.getRect(find.byType(TextField));
    expect(label.bottom, lessThanOrEqualTo(input.top));
    final decoration = textFieldOf(tester, 'RFC').decoration!;
    expect(decoration.labelText, isNull);
    expect(decoration.label, isNull);
  });

  testWidgets('the fewer kind is marked in text: optional by default', (
    tester,
  ) async {
    final name = form.textField(label: 'Nombre');
    final note = form.textField(label: 'Nota', isRequired: false);
    await pumpFields(tester, [_plain(name), _plain(note)]);

    expect(find.text('Nombre'), findsOneWidget);
    expect(find.text('Nota (opcional)'), findsOneWidget);
  });

  testWidgets('required fields are marked when they are the fewer kind', (
    tester,
  ) async {
    form.dispose();
    form = AppFormController(marks: RequirementMark.requiredFields);
    final name = form.textField(label: 'Nombre');
    await pumpFields(tester, [_plain(name)]);

    expect(find.text('Nombre (obligatorio)'), findsOneWidget);
  });

  testWidgets('keyboard, action and autofill reach the platform field', (
    tester,
  ) async {
    final email = form.textField(label: 'Correo');
    await pumpFields(tester, [
      AppTextField(
        key: const ValueKey('Correo'),
        field: email,
        keyboardType: TextInputType.emailAddress,
        textInputAction: TextInputAction.next,
        autofillHints: const [AutofillHints.email],
      ),
    ]);

    final input = textFieldOf(tester, 'Correo');
    expect(input.keyboardType, TextInputType.emailAddress);
    expect(input.textInputAction, TextInputAction.next);
    expect(input.autofillHints, [AutofillHints.email]);
    expect(input.textCapitalization, TextCapitalization.none);
    expect(input.autocorrect, isTrue);
  });

  testWidgets('a code field capitalizes and never autocorrects', (
    tester,
  ) async {
    final rfc = form.textField(label: 'RFC');
    await pumpFields(tester, [
      AppTextField(
        key: const ValueKey('RFC'),
        field: rfc,
        keyboardType: TextInputType.text,
        textInputAction: TextInputAction.next,
        autofillHints: const [],
        textCapitalization: TextCapitalization.characters,
        autocorrect: false,
      ),
    ]);

    final input = textFieldOf(tester, 'RFC');
    expect(input.textCapitalization, TextCapitalization.characters);
    expect(input.autocorrect, isFalse);
    expect(input.enableSuggestions, isFalse);
  });

  testWidgets('input text is at least 16 px and the target 44 px', (
    tester,
  ) async {
    final name = form.textField(label: 'Nombre');
    await pumpFields(tester, [_plain(name)]);

    final editable = tester.widget<EditableText>(find.byType(EditableText));
    expect(editable.style.fontSize, greaterThanOrEqualTo(16));
    expect(
      tester.getSize(find.byType(TextField)).height,
      greaterThanOrEqualTo(AppSizes.minTarget),
    );
  });

  testWidgets('the error shows under the field and is announced with it', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    final name = form.textField(label: 'Nombre');
    await pumpFields(tester, [_plain(name)]);

    await form.submit(() async {});
    await tester.pump();

    expect(find.text('Escribe este dato.'), findsWidgets);
    final node = tester.getSemantics(find.byType(EditableText));
    expect(node.flagsCollection.isTextField, isTrue);
    expect(node.label, 'Nombre');
    expect(node.hint, contains('Escribe este dato.'));
    semantics.dispose();
  });

  testWidgets('a format error appears only after leaving the field', (
    tester,
  ) async {
    final amount = form.moneyField(label: 'Monto');
    final name = form.textField(label: 'Nombre');
    await pumpFields(tester, [
      AppMoneyField(field: amount, textInputAction: TextInputAction.next),
      _plain(name),
    ]);

    await tester.enterText(find.byType(TextField).first, '12,34');
    await tester.pump();
    expect(find.text('Escribe un monto, por ejemplo 1,234.50.'), findsNothing);

    name.focusNode.requestFocus();
    await tester.pump();
    expect(find.text('Escribe un monto, por ejemplo 1,234.50.'), findsWidgets);
  });

  testWidgets('money and date fields open numeric keyboards', (tester) async {
    final amount = form.moneyField(label: 'Monto');
    final date = form.dateField(label: 'Fecha');
    await pumpFields(tester, [
      AppMoneyField(
        key: const ValueKey('Monto'),
        field: amount,
        textInputAction: TextInputAction.next,
      ),
      AppDateField(
        key: const ValueKey('Fecha'),
        field: date,
        textInputAction: TextInputAction.done,
      ),
    ]);

    expect(
      textFieldOf(tester, 'Monto').keyboardType,
      const TextInputType.numberWithOptions(decimal: true),
    );
    expect(textFieldOf(tester, 'Fecha').keyboardType, TextInputType.datetime);
    expect(find.text('dd/mm/aaaa'), findsOneWidget);
  });
}

/// A free-text field: the three keyboard parameters are required on purpose.
AppTextField _plain(AppFieldController field) => AppTextField(
  field: field,
  keyboardType: TextInputType.text,
  textInputAction: TextInputAction.next,
  autofillHints: const [],
);
