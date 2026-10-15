import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/history/domain/history_viewed.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../tax_engine/estimate_fixtures.dart';
import 'pump_history.dart';

void main() {
  group('viewedRequestOf', () {
    test('copies what the estimate rendered, verbatim', () {
      final request = viewedRequestOf(
        estimateFrom(estimateJson(period: '2026-02')),
        current: '2026-03',
      );

      expect(request?.figuresFingerprint, fixtureFiguresFingerprint);
      expect(
        request?.recordsFingerprint,
        completeEstimateJson['records_fingerprint'],
      );
      expect(request?.engineVersion, '1.0.0');
      expect(request?.parameterSet.fiscalYear, 2026);
      expect(request?.parameterSet.version, 1);
    });

    test('the current month is live: never marked', () {
      expect(
        viewedRequestOf(estimateFrom(estimateJson()), current: '2026-03'),
        isNull,
      );
    });
  });

  group('marking from any entry point', () {
    late HistoryHarness harness;

    setUp(() {
      harness = HistoryHarness();
      harness.estimates.fallback = estimateFrom(
        estimateJson(period: '2026-02', dueDate: '2026-03-17'),
      );
    });

    testWidgets("the estimate screen's own month picker marks too", (
      tester,
    ) async {
      harness.estimates.replies.add(estimateFrom(completeEstimateJson));
      await harness.pump(tester, location: TaxEngineRoutes.estimate);

      await tester.tap(find.byKey(const ValueKey('estimate-period')));
      await tester.pumpAndSettle();
      await tester.tap(find.text('febrero de 2026').last);
      await tester.pumpAndSettle();

      expect(harness.repository.viewed.single.period, '2026-02');
    });

    testWidgets('with module_history OFF nothing is posted', (tester) async {
      await harness.pump(
        tester,
        modules: historyModules.difference({AppModule.history}),
        location: TaxEngineRoutes.estimateOf('2026-02'),
      );

      expect(find.text('Total a pagar'), findsOneWidget);
      expect(harness.repository.viewed, isEmpty);
    });

    testWidgets('a failed post never blocks the screen; the next open '
        'tries again', (tester) async {
      harness.repository.viewedReplies.add(const ApiUnreachable());
      await harness.pump(
        tester,
        location: TaxEngineRoutes.estimateOf('2026-02'),
      );

      expect(find.text('Total a pagar'), findsOneWidget);
      expect(tester.takeException(), isNull);
      await harness.back(tester);
      await harness.push(tester, TaxEngineRoutes.estimateOf('2026-02'));

      expect(harness.repository.viewed, hasLength(2));
    });
  });
}
