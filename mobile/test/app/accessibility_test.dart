import 'package:control_contable/core/router/app_router.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/pump_app.dart';

/// DIRECTIONS.md: layouts must survive 200 % system text.
const _maxSupportedTextScale = 2.0;

Future<void> _openNotFound(WidgetTester tester) async {
  final container = ProviderScope.containerOf(
    tester.element(find.byType(MaterialApp)),
  );
  container.read(appRouterProvider).go('/no-existe');
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('screens meet the tap-target, labeling and contrast guidelines', (
    tester,
  ) async {
    final semantics = tester.ensureSemantics();
    await pumpControlContableApp(tester);
    await _openNotFound(tester);

    await expectLater(tester, meetsGuideline(androidTapTargetGuideline));
    await expectLater(tester, meetsGuideline(iOSTapTargetGuideline));
    await expectLater(tester, meetsGuideline(labeledTapTargetGuideline));
    await expectLater(tester, meetsGuideline(textContrastGuideline));
    semantics.dispose();
  });

  testWidgets('screens render at 200 % text scale without overflow', (
    tester,
  ) async {
    tester.platformDispatcher.textScaleFactorTestValue = _maxSupportedTextScale;
    addTearDown(tester.platformDispatcher.clearTextScaleFactorTestValue);

    await pumpControlContableApp(tester);
    expect(tester.takeException(), isNull);

    await _openNotFound(tester);
    expect(tester.takeException(), isNull);
  });
}
