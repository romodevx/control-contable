import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart' show fixtureTaxpayerId;
import 'annual_inputs_fixtures.dart';
import 'pump_annual_inputs.dart';

/// ANN-19..21 screens meet the tap-target, label and contrast guidelines
/// and survive 200 % text without overflow.
/// A small phone, where 200 % text has the least room.
const Size _phone = Size(360, 740);

void main() {
  final screens = {
    'deductions': AnnualRoutes.personalDeductions(fixtureTaxpayerId, 2026),
    'deduction form': AnnualRoutes.newPersonalDeduction(
      fixtureTaxpayerId,
      2026,
    ),
    'constancias': AnnualRoutes.salaryStatements(fixtureTaxpayerId, 2026),
    'filed payments': AnnualRoutes.filedPayments(
      fixtureTaxpayerId,
      2026,
      '612',
    ),
    'year facts': AnnualRoutes.annualInputs(fixtureTaxpayerId, 2026),
  };

  for (final MapEntry(key: name, value: location) in screens.entries) {
    testWidgets('$name: guidelines and 200 % text', (tester) async {
      tester.view
        ..physicalSize = _phone
        ..devicePixelRatio = 1;
      addTearDown(tester.view.reset);
      final semantics = tester.ensureSemantics();
      final harness = AnnualInputsHarness();
      harness.inputs
        ..deductions.add(deductionJson())
        ..statements.add(statementJson());
      await harness.pump(tester, location: location);

      await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
      await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
      await expectLater(tester, meetsGuideline(textContrastGuideline));
      semantics.dispose();

      tester.platformDispatcher.textScaleFactorTestValue = 2;
      addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('constancia sheet: guidelines and 200 % text', (tester) async {
    tester.view
      ..physicalSize = _phone
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    final semantics = tester.ensureSemantics();
    final harness = AnnualInputsHarness();
    await harness.pump(
      tester,
      location: AnnualRoutes.salaryStatements(fixtureTaxpayerId, 2026),
    );
    await tester.tap(find.text('Agregar constancia'));
    await tester.pumpAndSettle();

    await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
    await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    semantics.dispose();

    tester.platformDispatcher.textScaleFactorTestValue = 2;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
