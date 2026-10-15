import 'dart:async';

import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/synthetic_password.dart';

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  group('inactive fields (a choice hides them)', () {
    test('are skipped on submit and lose their error', () async {
      final donee = form.add(ChoiceFieldController(label: 'Donataria'));
      await form.submit(() async {});
      expect(donee.error, isA<MissingChoice>());

      donee.isActive = false;
      expect(donee.error, isNull);
      var saved = false;
      await form.submit(() async => saved = true);

      expect(saved, isTrue);
      expect(form.invalidFields, isEmpty);
    });

    test('count again once active', () async {
      final donee = form.add(ChoiceFieldController(label: 'Donataria'))
        ..isActive = false;
      var notified = 0;
      form.addListener(() => notified++);

      donee.isActive = true;
      final saved = await form.submit(() async {});

      expect(notified, greaterThan(0));
      expect(saved, isFalse);
      expect(donee.isActive, isTrue);
    });
  });

  group('requiredness that follows another choice', () {
    test('turning optional clears a shown "missing" error', () async {
      final collected = form.dateField(label: 'Fecha de cobro');
      await form.submit(() async {});
      expect(collected.error, isA<MissingValue>());
      var notified = 0;
      collected.addListener(() => notified++);

      collected.isRequired = false;

      expect(collected.isRequired, isFalse);
      expect(collected.error, isNull);
      expect(notified, 1, reason: 'the label mark changes');
      var saved = false;
      await form.submit(() async => saved = true);
      expect(saved, isTrue);
    });

    test('turning required shows nothing until the next submit', () async {
      final invoiced = form.dateField(
        label: 'Fecha de factura',
        isRequired: false,
      )..isRequired = true;

      expect(invoiced.error, isNull);
      await form.submit(() async {});
      expect(invoiced.error, isA<MissingValue>());
    });
  });

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

  group('custom fields', () {
    test('canonicalize rewrites a valid text field on blur', () {
      final rfc =
          form.textField(
              label: 'RFC',
              canonicalize: (value) => value.toUpperCase(),
            )
            ..text.text = 'goma800101ab6'
            ..handleBlur();

      expect(rfc.text.text, 'GOMA800101AB6');
    });

    test('canonicalize never rewrites an invalid value', () {
      final rfc =
          form.textField(
              label: 'RFC',
              validator: (_) => const CustomFieldError('mal'),
              canonicalize: (value) => value.toUpperCase(),
            )
            ..text.text = 'abc'
            ..handleBlur();

      expect(rfc.text.text, 'abc');
    });

    test('add puts a field of any kind under the form policy', () async {
      final kind = form.add(ChoiceFieldController(label: 'Tipo'));

      expect(form.fields, [kind]);
      expect(await form.submit(() async {}), isFalse);
      expect(form.invalidFields, [kind]);
    });
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

  // QA M3/M4 D-8: the next record starts where prepareNext leaves it, so
  // kept context values (any kind of field) are not unsaved changes.
  test('after prepareNext the form is pristine until a value changes', () {
    final date = form.dateField(label: 'Fecha', keepForNext: true)
      ..text.text = '01/02/2026';
    final amount = form.moneyField(label: 'Monto')..text.text = '12,34';
    final payer = form.add(
      ChoiceFieldController(label: 'Pagador', keepForNext: true),
    )..value = 'moral';
    final box = form.checkboxField(label: 'Con factura', keepForNext: true)
      ..checked = true;
    final set = form.add(
      MultiChoiceFieldController(label: 'Actividades', keepForNext: true),
    )..toggle('a');
    final cleared = form.add(ChoiceFieldController(label: 'Tasa'))
      ..value = '16';
    expect(form.isDirty, isTrue);

    form.prepareNext();
    expect(form.isDirty, isFalse);

    for (final change in <void Function()>[
      () => date.text.text = '02/02/2026',
      () => amount.text.text = '1',
      () => payer.value = 'fisica',
      () => box.checked = false,
      () => set.toggle('b'),
      () => cleared.value = '8',
    ]) {
      change();
      expect(form.isDirty, isTrue);
      form.prepareNext();
    }
  });

  test('isDirty is true once any field differs from its start', () {
    final name = form.textField(label: 'Nombre', initialText: 'Ana');
    expect(form.isDirty, isFalse);

    name.text.text = 'Ana María';
    expect(form.isDirty, isTrue);
  });
}
