import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flag_gate.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/flags/flags_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/fake_flags_repository.dart';

const _menuEntry = 'Ingresos';

/// A screen that reads flags only through the two sanctioned helpers.
class _Menu extends ConsumerWidget {
  const new();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final expensesOn = ref.useFlag(AppModule.expenses);
    return Column(
      children: [
        const FlagGate(module: AppModule.incomes, child: Text(_menuEntry)),
        Text(expensesOn ? 'gastos: on' : 'gastos: off'),
      ],
    );
  }
}

Future<void> _pumpMenu(WidgetTester tester, Set<AppModule> enabled) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        flagsRepositoryProvider.overrideWithValue(FakeFlagsRepository(enabled)),
      ],
      child: const MaterialApp(home: Scaffold(body: _Menu())),
    ),
  );
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('gated UI is absent when its module is OFF', (tester) async {
    await _pumpMenu(tester, const {});

    expect(find.text(_menuEntry), findsNothing);
    expect(find.text('gastos: off'), findsOneWidget);
  });

  testWidgets('gated UI renders when its module is ON', (tester) async {
    await _pumpMenu(tester, const {AppModule.incomes, AppModule.expenses});

    expect(find.text(_menuEntry), findsOneWidget);
    expect(find.text('gastos: on'), findsOneWidget);
  });

  testWidgets('gated UI is absent while the flags are still loading', (
    tester,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          flagsRepositoryProvider.overrideWithValue(
            FakeFlagsRepository.never(),
          ),
        ],
        child: const MaterialApp(home: Scaffold(body: _Menu())),
      ),
    );

    expect(find.text(_menuEntry), findsNothing);
  });

  for (final (enabled, shown) in [
    (<AppModule>{}, 'antes'),
    ({AppModule.dashboard}, 'tablero'),
  ]) {
    testWidgets('a fallback stands in while the module is OFF ($shown)', (
      tester,
    ) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            flagsRepositoryProvider.overrideWithValue(
              FakeFlagsRepository(enabled),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(
              body: FlagGate(
                module: AppModule.dashboard,
                fallback: Text('antes'),
                child: Text('tablero'),
              ),
            ),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text(shown), findsOneWidget);
      expect(find.text(shown == 'antes' ? 'tablero' : 'antes'), findsNothing);
    });
  }
}
