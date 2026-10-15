import 'package:control_contable/api/generated/models/deduction_beneficiary.dart';
import 'package:control_contable/api/generated/models/deduction_category.dart';
import 'package:control_contable/api/generated/models/deduction_payment_method.dart';
import 'package:control_contable/api/generated/models/donee_type.dart';
import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/tuition_level.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/choice_finders.dart';
import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import '../taxpayers/pump_onboarding.dart' show inputLabeled, typeInto;
import '../taxpayers/taxpayers_fakes.dart' show detail;
import 'pump_annual_inputs.dart';

final String _list = AnnualRoutes.personalDeductions(fixtureTaxpayerId, 2026);
final String _form = AnnualRoutes.newPersonalDeduction(fixtureTaxpayerId, 2026);

const String _category = 'Tipo de deducción';
const String _medical = 'Honorarios médicos y hospitalarios';
const String _electronic = 'Pagué con tarjeta, transferencia o cheque';

/// ANN-19: the category-driven personal deduction form (its own screen).
void main() {
  late AnnualInputsHarness harness;

  setUp(() => harness = AnnualInputsHarness());

  Future<void> openForm(WidgetTester tester) async {
    await harness.pump(tester, location: _list);
    await tester.tap(find.text('Agregar deducción'));
    await tester.pumpAndSettle();
  }

  Future<void> save(WidgetTester tester) async {
    await tester.tap(find.widgetWithText(FilledButton, 'Guardar'));
    await tester.pumpAndSettle();
  }

  testWidgets('category first: nothing else until one is chosen', (
    tester,
  ) async {
    await openForm(tester);

    expect(find.text('Nueva deducción'), findsOneWidget);
    expect(groupOf(_category), findsOneWidget);
    expect(find.text('Monto'), findsNothing);

    await chooseIn(tester, _category, _medical);

    expect(inputLabeled('Monto'), findsOneWidget);
    expect(inputLabeled('Fecha de pago'), findsOneWidget);
    expect(find.text(_electronic), findsOneWidget);
    expect(groupOf('¿Para quién fue?'), findsOneWidget);
    expect(find.text('Escuela'), findsNothing);
  });

  testWidgets('a medical fee saves with today, electronic and self by '
      'default, and goes back to the list', (tester) async {
    await openForm(tester);
    await chooseIn(tester, _category, _medical);
    await typeInto(tester, 'Monto', '1,234.5');

    expect(
      tester.widget<TextField>(inputLabeled('Fecha de pago')).controller!.text,
      '14/10/2026',
    );
    await save(tester);

    final sent = harness.inputs.createdDeductions.single;
    expect(sent.id, 'client-1');
    expect(sent.category, DeductionCategory.medical);
    expect(sent.amountCents, 123450);
    expect(sent.paidOn, '2026-10-14');
    expect(sent.paymentMethod, DeductionPaymentMethod.electronic);
    expect(sent.beneficiary, DeductionBeneficiary.self);
    expect(sent.cfdiUuid, isNull);
    expect(harness.location, _list);
    expect(find.text('Deducción guardada'), findsOneWidget);
  });

  testWidgets('paid in cash: warns it is not deductible and saves cash', (
    tester,
  ) async {
    await openForm(tester);
    await chooseIn(tester, _category, _medical);
    await _tapText(tester, _electronic);

    expect(
      find.text(
        'En efectivo no es deducible: la guardamos, pero la '
        'estimación la deja fuera.',
      ),
      findsOneWidget,
    );
    await typeInto(tester, 'Monto', '500');
    await save(tester);

    expect(
      harness.inputs.createdDeductions.single.paymentMethod,
      DeductionPaymentMethod.cash,
    );
  });

  testWidgets('mortgage interest: no cash question, dated 31 December', (
    tester,
  ) async {
    await openForm(tester);
    await chooseIn(
      tester,
      _category,
      'Intereses reales de crédito hipotecario',
    );

    expect(find.text(_electronic), findsNothing);
    expect(find.text('¿Para quién fue?'), findsNothing);
    await typeInto(tester, 'Monto', '8000');
    await save(tester);

    final sent = harness.inputs.createdDeductions.single;
    expect(sent.paidOn, '2026-12-31');
    expect(sent.paymentMethod, DeductionPaymentMethod.electronic);
    expect(sent.beneficiary, DeductionBeneficiary.self);
  });

  testWidgets('donation asks for the donee', (tester) async {
    await openForm(tester);
    await chooseIn(tester, _category, 'Donativos');
    await typeInto(tester, 'Monto', '1000');
    await save(tester);

    expect(harness.inputs.createdDeductions, isEmpty);
    expect(find.text('Elige una opción.'), findsWidgets);

    await chooseIn(tester, '¿A quién donaste?', 'Gobierno');
    await save(tester);

    expect(
      harness.inputs.createdDeductions.single.doneeType,
      DoneeType.government,
    );
  });

  testWidgets('tuition asks for the student, level and scholarship; a '
      'dependant needs the income attestation', (tester) async {
    await openForm(tester);
    await chooseIn(tester, _category, 'Colegiaturas');
    await typeInto(tester, 'Monto', '20000');
    await typeInto(tester, 'Alumno', 'Sofía');
    await chooseIn(tester, 'Nivel escolar', 'Primaria');
    await chooseIn(tester, '¿Para quién fue?', 'Descendiente (hijos, nietos)');
    await save(tester);

    expect(harness.inputs.createdDeductions, isEmpty);
    expect(find.text('Marca esta casilla para continuar.'), findsWidgets);

    await _tapText(
      tester,
      'Confirmo que esta persona tuvo ingresos menores '
      'a una UMA anual en 2026.',
    );
    await save(tester);

    final sent = harness.inputs.createdDeductions.single;
    expect(sent.studentLabel, 'Sofía');
    expect(sent.tuitionLevel, TuitionLevel.primaria);
    expect(sent.scholarshipCents, 0);
    expect(sent.beneficiary, DeductionBeneficiary.descendant);
  });

  testWidgets('the CFDI folio is optional, under "Más datos"', (tester) async {
    await openForm(tester);
    await chooseIn(tester, _category, _medical);
    await typeInto(tester, 'Monto', '100');
    await _tapText(tester, 'Más datos');
    await typeInto(
      tester,
      'Folio fiscal (UUID) del CFDI (opcional)',
      'AB3F0C4E-7B1E-4C3A-9D2E-6F1A2B3C4D5E',
    );
    await save(tester);

    expect(
      harness.inputs.createdDeductions.single.cfdiUuid,
      'AB3F0C4E-7B1E-4C3A-9D2E-6F1A2B3C4D5E',
    );
  });

  testWidgets('submitting empty shows the summary, never sends', (
    tester,
  ) async {
    await openForm(tester);
    await save(tester);

    expect(harness.inputs.createdDeductions, isEmpty);
    expect(find.textContaining('Revisa'), findsWidgets);
  });

  testWidgets('the 500 bound refuses with its message', (tester) async {
    harness.inputs.failures['createDeduction'] = ApiRejected(
      code: ErrorCode.businessRuleViolation,
      details: [detail('body', 'limit_reached')],
    );
    await openForm(tester);
    await chooseIn(tester, _category, _medical);
    await typeInto(tester, 'Monto', '100');
    await save(tester);

    expect(
      find.text(
        'Llegaste al límite de 500 deducciones de este año. Para '
        'agregar otra, elimina alguna.',
      ),
      findsOneWidget,
    );
    expect(harness.location, _form);
  });

  testWidgets('a date outside the year is the server verdict on the field', (
    tester,
  ) async {
    harness.inputs.failures['createDeduction'] = ApiValidationFailure([
      detail('paid_on', 'out_of_range'),
    ]);
    await openForm(tester);
    await chooseIn(tester, _category, _medical);
    await typeInto(tester, 'Monto', '100');
    await save(tester);

    expect(find.text('La fecha debe ser de 2026.'), findsWidgets);
  });

  testWidgets('read-only (402): the paywall over the form, data kept', (
    tester,
  ) async {
    harness.inputs.failures['createDeduction'] = const ApiEntitlementRequired();
    await openForm(tester);
    await chooseIn(tester, _category, _medical);
    await typeInto(tester, 'Monto', '100');
    await save(tester);

    expect(find.text('Tu cuenta es de solo consulta'), findsOneWidget);
    expect(harness.location, _form);
  });
}

Future<void> _tapText(WidgetTester tester, String text) async {
  final target = find.text(text);
  await tester.ensureVisible(target);
  await tester.pumpAndSettle();
  await tester.tap(target);
  await tester.pumpAndSettle();
}
