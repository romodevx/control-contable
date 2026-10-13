import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  Future<void> pumpPassword(WidgetTester tester) => pumpForm(
    tester,
    AppForm(
      controller: form,
      onSave: () async {},
      children: [
        AppPasswordField(
          field: form.textField(label: 'Contraseña'),
          textInputAction: TextInputAction.done,
          autofillHints: const [AutofillHints.newPassword],
          helperText: 'Al menos 12 caracteres.',
        ),
      ],
    ),
  );

  TextField input(WidgetTester tester) => tester.widget(find.byType(TextField));

  testWidgets('hides the text and turns off suggestions', (tester) async {
    await pumpPassword(tester);

    final field = input(tester);
    expect(field.obscureText, isTrue);
    expect(field.enableSuggestions, isFalse);
    expect(field.autocorrect, isFalse);
    expect(field.keyboardType, TextInputType.visiblePassword);
    expect(field.autofillHints, [AutofillHints.newPassword]);
  });

  testWidgets('shows the help under the field', (tester) async {
    await pumpPassword(tester);

    expect(input(tester).decoration!.helperText, 'Al menos 12 caracteres.');
  });

  testWidgets('a labeled toggle shows and hides the password', (tester) async {
    await pumpPassword(tester);

    await tester.tap(find.byTooltip('Mostrar contraseña'));
    await tester.pump();
    expect(input(tester).obscureText, isFalse);

    await tester.tap(find.byTooltip('Ocultar contraseña'));
    await tester.pump();
    expect(input(tester).obscureText, isTrue);
  });
}
