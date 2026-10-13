import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:flutter_test/flutter_test.dart';

api.FieldError _detail(String field, [String code = 'invalid']) =>
    api.FieldError(field: field, code: code, message: 'server text');

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  FieldError describe(api.FieldError detail) =>
      CustomFieldError('${detail.field}/${detail.code}');

  test('a detail lands on the field with the same path', () {
    final password = form.textField(label: 'Contraseña');

    final result = matchServerFieldErrors(
      [_detail('password', 'password_policy')],
      fieldsByPath: {'password': password},
      describe: describe,
    );

    expect(result.unmatched, isEmpty);
    expect(
      result.matched[password],
      isA<CustomFieldError>().having(
        (e) => e.message,
        'message',
        'password/password_policy',
      ),
    );
  });

  test('a nested dotted path lands on its nearest mapped parent', () {
    final docs = form.checkboxField(label: 'Acepto');

    final result = matchServerFieldErrors(
      [_detail('accepted_documents.1.kind')],
      fieldsByPath: {'accepted_documents': docs},
      describe: describe,
    );

    expect(result.matched.keys, [docs]);
  });

  test('a parent never matches a mere prefix of a segment', () {
    final name = form.textField(label: 'Nombre');

    final result = matchServerFieldErrors(
      [_detail('name_extra')],
      fieldsByPath: {'name': name},
      describe: describe,
    );

    expect(result.matched, isEmpty);
    expect(result.unmatched.single.field, 'name_extra');
  });

  test('the first detail of a field wins; unknown paths are kept apart', () {
    final email = form.textField(label: 'Correo');

    final result = matchServerFieldErrors(
      [_detail('email', 'format'), _detail('email', 'other'), _detail('body')],
      fieldsByPath: {'email': email},
      describe: describe,
    );

    expect(
      (result.matched[email]! as CustomFieldError).message,
      'email/format',
    );
    expect(result.unmatched.map((d) => d.field), ['body']);
  });
}
