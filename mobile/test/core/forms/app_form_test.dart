import 'dart:async';

import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_error_summary.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_form.dart';

const _missing = 'Escribe este dato.';

void main() {
  late AppFormController form;
  late AppFieldController name;
  late DateFieldController date;
  late MoneyFieldController amount;
  late int saves;
  late int cancels;

  setUp(() {
    form = AppFormController();
    date = form.dateField(label: 'Fecha', keepForNext: true);
    name = form.textField(label: 'Concepto');
    amount = form.moneyField(label: 'Monto');
    saves = 0;
    cancels = 0;
  });
  tearDown(() => form.dispose());

  Future<void> pumpIncomeForm(
    WidgetTester tester, {
    Future<void> Function()? onSave,
  }) => pumpForm(
    tester,
    AppForm(
      controller: form,
      onSave: onSave ?? () async => saves++,
      onCancel: () => cancels++,
      allowCreateAnother: true,
      children: [
        AppDateField(field: date, textInputAction: TextInputAction.next),
        AppTextField(
          field: name,
          keyboardType: TextInputType.text,
          textInputAction: TextInputAction.next,
          autofillHints: const [],
        ),
        AppMoneyField(field: amount, textInputAction: TextInputAction.done),
      ],
    ),
  );

  Finder primary() => find.widgetWithText(FilledButton, 'Guardar');

  group('a failed submit', () {
    testWidgets('shows a summary at the top that takes focus', (tester) async {
      await pumpIncomeForm(tester);

      await tester.tap(primary());
      await tester.pumpAndSettle();

      expect(find.text('Revisa 3 datos antes de guardar'), findsOneWidget);
      final summary = tester.getRect(find.byType(FormErrorSummary));
      final firstField = tester.getRect(find.byType(TextField).first);
      expect(summary.bottom, lessThanOrEqualTo(firstField.top));
      expect(
        Focus.of(tester.element(find.text('Revisa 3 datos antes de guardar')))
            .hasFocus,
        isTrue,
      );
      expect(find.text(_missing), findsWidgets, reason: 'inline errors stay');
      expect(saves, 0);
    });

    testWidgets('every failed submit is announced to screen readers, '
        'even when the errors are the same', (tester) async {
      await pumpIncomeForm(tester);
      tester.takeAnnouncements();

      for (var attempt = 0; attempt < 2; attempt++) {
        await tester.tap(primary());
        await tester.pumpAndSettle();
      }

      final announced = tester.takeAnnouncements().map((a) => a.message);
      expect(announced, [
        'Revisa 3 datos antes de guardar',
        'Revisa 3 datos antes de guardar',
      ]);
    });

    testWidgets('server errors fill the summary like a failed submit', (
      tester,
    ) async {
      await pumpIncomeForm(tester);

      form.showServerErrors({
        amount: const CustomFieldError('Excede el saldo'),
      });
      await tester.pumpAndSettle();

      expect(find.text('Monto: Excede el saldo'), findsOneWidget);
      expect(find.text('Excede el saldo'), findsOneWidget, reason: 'inline');
    });

    testWidgets('each summary item moves focus to its field', (tester) async {
      await pumpIncomeForm(tester);
      await tester.tap(primary());
      await tester.pumpAndSettle();

      await tester.tap(find.text('Monto: $_missing'));
      await tester.pumpAndSettle();

      expect(amount.focusNode.hasFocus, isTrue);
    });

    testWidgets('the summary goes away once every error is fixed', (
      tester,
    ) async {
      await pumpIncomeForm(tester);
      await tester.tap(primary());
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextField).at(0), '01/02/2026');
      await tester.enterText(find.byType(TextField).at(1), 'Honorarios');
      await tester.enterText(find.byType(TextField).at(2), '1,234.50');
      await tester.pump();

      expect(find.byType(FormErrorSummary), findsOneWidget);
      expect(find.textContaining('Revisa'), findsNothing);
    });
  });

  group('the primary button', () {
    Future<void> fillValid(WidgetTester tester) async {
      date.text.text = '01/02/2026';
      name.text.text = 'Honorarios';
      amount.text.text = '1234.5';
      await tester.pump();
    }

    testWidgets('is never disabled', (tester) async {
      await pumpIncomeForm(tester);

      expect(tester.widget<FilledButton>(primary()).enabled, isTrue);
    });

    testWidgets('stays enabled and busy while saving, ignoring repeats', (
      tester,
    ) async {
      final semantics = tester.ensureSemantics();
      final pending = Completer<void>();
      await pumpIncomeForm(
        tester,
        onSave: () {
          saves++;
          return pending.future;
        },
      );
      await fillValid(tester);

      await tester.tap(primary());
      await tester.pump();
      await tester.tap(primary());
      await tester.pump();

      expect(saves, 1);
      expect(tester.widget<FilledButton>(primary()).enabled, isTrue);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.bySemanticsLabel('Guardando…'), findsOneWidget);

      pending.complete();
      await tester.pumpAndSettle();
      expect(find.byType(CircularProgressIndicator), findsNothing);
      semantics.dispose();
    });

    testWidgets('Enter in a single-line done field submits', (tester) async {
      await pumpIncomeForm(tester);
      await fillValid(tester);

      await tester.showKeyboard(find.byType(TextField).last);
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      expect(saves, 1);
    });

    testWidgets('Ctrl+Enter and Cmd+Enter submit from any field', (
      tester,
    ) async {
      await pumpIncomeForm(tester);
      await fillValid(tester);
      await tester.showKeyboard(find.byType(TextField).first);

      await tester.sendKeyDownEvent(LogicalKeyboardKey.controlLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.controlLeft);
      await tester.pumpAndSettle();
      await tester.sendKeyDownEvent(LogicalKeyboardKey.metaLeft);
      await tester.sendKeyEvent(LogicalKeyboardKey.enter);
      await tester.sendKeyUpEvent(LogicalKeyboardKey.metaLeft);
      await tester.pumpAndSettle();

      expect(saves, 2);
    });
  });

  testWidgets('"Guardar y crear otro" keeps context, clears the rest, '
      'and refocuses', (tester) async {
    await pumpIncomeForm(tester);
    date.text.text = '01/02/2026';
    name.text.text = 'Honorarios';
    amount.text.text = '1,234.50';

    await tester.tap(
      find.widgetWithText(OutlinedButton, 'Guardar y crear otro'),
    );
    await tester.pumpAndSettle();

    expect(saves, 1);
    expect(date.text.text, '01/02/2026');
    expect(name.text.text, isEmpty);
    expect(amount.text.text, isEmpty);
    expect(name.focusNode.hasFocus, isTrue);
  });

  testWidgets('"Guardar y crear otro" uses its own save when given, and '
      'keeps every field when that save did not go through', (tester) async {
    var saved = false;
    var nextSaves = 0;
    await pumpForm(
      tester,
      AppForm(
        controller: form,
        onSave: () async => saves++,
        onSaveForNext: () async {
          nextSaves++;
          return saved;
        },
        allowCreateAnother: true,
        children: [
          AppTextField(
            field: name,
            keyboardType: TextInputType.text,
            textInputAction: TextInputAction.done,
            autofillHints: const [],
          ),
        ],
      ),
    );
    date.text.text = '01/02/2026';
    amount.text.text = '1,234.50';
    name.text.text = 'Honorarios';
    final another = find.widgetWithText(OutlinedButton, 'Guardar y crear otro');

    await tester.tap(another);
    await tester.pumpAndSettle();
    expect(nextSaves, 1);
    expect(saves, 0);
    expect(name.text.text, 'Honorarios');

    saved = true;
    await tester.tap(another);
    await tester.pumpAndSettle();
    expect(name.text.text, isEmpty);
  });

  testWidgets('Cancelar is a text button', (tester) async {
    await pumpIncomeForm(tester);

    await tester.tap(find.widgetWithText(TextButton, 'Cancelar'));

    expect(cancels, 1);
  });

  testWidgets('the form meets the tap-target and contrast guidelines, '
      'light and dark, and survives 200 % text', (tester) async {
    final semantics = tester.ensureSemantics();
    for (final brightness in Brightness.values) {
      await pumpForm(
        tester,
        AppForm(
          controller: form,
          onSave: () async {},
          onCancel: () {},
          allowCreateAnother: true,
          children: [
            AppMoneyField(field: amount, textInputAction: TextInputAction.done),
          ],
        ),
        brightness: brightness,
      );
      await tester.tap(primary());
      await tester.pumpAndSettle();

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
    }
    semantics.dispose();

    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
