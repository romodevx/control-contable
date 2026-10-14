import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/regime_incompatibility.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_selection_field.dart';
import 'package:flutter_test/flutter_test.dart';

import 'taxpayers_fakes.dart';

const _missing = CustomFieldError('Elige al menos un régimen.');
const _conflict = CustomFieldError('Hay regímenes que no se pueden combinar.');

void main() {
  late AppFormController form;

  setUp(() => form = AppFormController());
  tearDown(() => form.dispose());

  RegimeSelectionField selection({
    bool single = false,
    List<Regime> initial = const [],
  }) => form.add(
    RegimeSelectionField(
      label: 'Regímenes',
      single: single,
      missingError: _missing,
      conflictError: _conflict,
      initial: initial,
    ),
  );

  test('nothing chosen fails on submit only', () async {
    final field = selection();
    expect(field.error, isNull);

    expect(await form.submit(() async {}), isFalse);
    expect(field.error, _missing);

    field.toggle(regime('606'));
    expect(field.error, isNull);
  });

  test('multi-select adds and drops in selection order', () {
    final field = selection()
      ..toggle(regime('612'))
      ..toggle(regime('606'));
    expect(field.selected.map((r) => r.code), ['612', '606']);

    field.toggle(regime('612'));
    expect(field.selected.map((r) => r.code), ['606']);
  });

  test('single-select keeps only the last choice', () {
    final field = selection(single: true)
      ..toggle(regime('601'))
      ..toggle(regime('626'))
      ..toggle(regime('626'));

    expect(field.selected.map((r) => r.code), ['626']);
  });

  test('an incompatible pair is an error on submit', () async {
    final field = selection()
      ..toggle(
        regime(
          '626',
          incompatibleWith: [
            const RegimeIncompatibility(code: '612', reason: 'No.'),
          ],
        ),
      )
      ..toggle(regime('612'));

    expect(field.conflicts.keys, ['612']);
    expect(await form.submit(() async {}), isFalse);
    expect(field.error, _conflict);
  });

  test('dirty against its initial choice; reset empties it', () {
    final field = selection(initial: [regime('606')]);
    expect(field.isDirty, isFalse);
    expect(field.isChosen(regime('606')), isTrue);

    field.reset();

    expect(field.selected, isEmpty);
    expect(field.isDirty, isTrue);
  });
}
