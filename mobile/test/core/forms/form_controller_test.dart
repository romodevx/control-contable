import 'dart:async';

import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  group('validation policy (ui-forms-and-modals.md)', () {
    test('a required-but-empty field shows nothing until submit', () async {
      final name = form.textField(label: 'Nombre')..handleBlur();
      expect(name.error, isNull, reason: 'never on blur');

      await form.submit(() async {});
      expect(name.error, isA<MissingValue>());
    });

    test('a format error shows on blur once the field has a value', () {
      final amount = form.moneyField(label: 'Monto');

      amount.text.text = '12,34';
      expect(amount.error, isNull, reason: 'never while typing');

      amount.handleBlur();
      expect(amount.error, isA<InvalidMoney>());
    });

    test('after an error shows, every change re-checks it', () {
      final amount = form.moneyField(label: 'Monto');
      amount.text.text = '12,34';
      amount.handleBlur();

      amount.text.text = '12,345';
      expect(amount.error, isNull, reason: 'cleared the moment it is valid');

      amount.text.text = '12,3456';
      expect(amount.error, isNull, reason: 'a fresh value waits for blur');
    });

    test('emptying a field with an error shows the required error', () {
      final amount = form.moneyField(label: 'Monto')
        ..text.text = '12,34'
        ..handleBlur()
        ..text.text = '';
      expect(amount.error, isA<MissingValue>());
    });

    test('a valid value is rewritten in canonical form on blur', () {
      final amount = form.moneyField(label: 'Monto');
      final date = form.dateField(label: 'Fecha');

      amount
        ..text.text = '1234.5'
        ..handleBlur();
      date
        ..text.text = '1-2-26'
        ..handleBlur();

      expect(amount.text.text, '1,234.50');
      expect(amount.centavos, 123450);
      expect(date.text.text, '01/02/2026');
      expect(date.date, DateTime(2026, 2));
    });

    test('an empty optional field is valid', () async {
      final note = form.textField(label: 'Nota', isRequired: false);

      expect(await form.submit(() async {}), isTrue);
      expect(note.error, isNull);
    });
  });

  group('submit', () {
    test('lists the invalid fields in order and does not save', () async {
      final name = form.textField(label: 'Nombre');
      form.textField(label: 'Nota', isRequired: false);
      final amount = form.moneyField(label: 'Monto');
      var saves = 0;

      final saved = await form.submit(() async => saves++);

      expect(saved, isFalse);
      expect(saves, 0);
      expect(form.invalidFields, [name, amount]);
      expect(form.submitFailures, 1);
    });

    test('is busy while saving and ignores repeat submits', () async {
      form.textField(label: 'Nombre').text.text = 'Ana';
      final pending = Completer<void>();
      var saves = 0;

      final first = form.submit(() {
        saves++;
        return pending.future;
      });
      expect(form.isSaving, isTrue);
      expect(await form.submit(() async => saves++), isFalse);

      pending.complete();
      expect(await first, isTrue);
      expect(form.isSaving, isFalse);
      expect(saves, 1);
    });

    test('clears the busy state when the save throws', () async {
      form.textField(label: 'Nombre').text.text = 'Ana';

      await expectLater(
        form.submit(() async => throw Exception('500')),
        throwsException,
      );
      expect(form.isSaving, isFalse);
    });

    test(
      'a save that outlives its form does not notify after dispose',
      () async {
        final screen = AppFormController();
        screen.textField(label: 'Nombre').text.text = 'Ana';
        final pending = Completer<void>();

        final saving = screen.submit(() => pending.future);
        screen.dispose(); // the user leaves while the request is in flight
        pending.complete();

        expect(await saving, isTrue);
      },
    );
  });

  group('verbatim fields (passwords)', () {
    test('keep surrounding spaces through blur and in their value', () {
      final password = form.textField(label: 'Contraseña', verbatim: true)
        ..text.text = ' ${syntheticPassword()} '
        ..handleBlur();

      expect(password.text.text, ' ${syntheticPassword()} ');
      expect(password.value, ' ${syntheticPassword()} ');
    });
  });

  group('server errors (the server is the authority)', () {
    test('survive leaving the field while the value is unchanged', () {
      final password = form.textField(label: 'Contraseña')
        ..text.text = syntheticPassword();
      form.showServerErrors({password: const CustomFieldError('Filtrada')});

      // The error summary takes focus: the field blurs.
      password.handleBlur();

      expect(password.error, isA<CustomFieldError>());
    });

    test('show on their fields and count as a failed submit', () {
      final name = form.textField(label: 'Nombre')..text.text = 'Ana';
      final amount = form.moneyField(label: 'Monto')..text.text = '10';

      form.showServerErrors({
        amount: const CustomFieldError('Excede el saldo'),
      });

      expect(
        amount.error,
        isA<CustomFieldError>().having(
          (e) => e.message,
          'message',
          'Excede el saldo',
        ),
      );
      expect(name.error, isNull);
      expect(form.invalidFields, [amount]);
      expect(form.submitFailures, 1);
    });

    test('clear once the user changes the value', () {
      final amount = form.moneyField(label: 'Monto')..text.text = '10';
      form.showServerErrors({
        amount: const CustomFieldError('Excede el saldo'),
      });

      amount.text.text = '9';

      expect(amount.error, isNull);
    });
  });

  test('prepareNext keeps context fields and clears the rest', () {
    final date = form.dateField(label: 'Fecha', keepForNext: true)
      ..text.text = '01/02/2026';
    final amount = form.moneyField(label: 'Monto')
      ..text.text = '12,34'
      ..handleBlur();

    form.prepareNext();

    expect(date.text.text, '01/02/2026');
    expect(amount.text.text, isEmpty);
    expect(amount.error, isNull);
  });

  test('isDirty is true once any field differs from its start', () {
    final name = form.textField(label: 'Nombre', initialText: 'Ana');
    expect(form.isDirty, isFalse);

    name.text.text = 'Ana María';
    expect(form.isDirty, isTrue);
  });
}
