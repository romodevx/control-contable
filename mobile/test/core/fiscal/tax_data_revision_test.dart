import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/fiscal/tax_data_revision.dart';
import 'package:control_contable/features/dashboard/application/dashboard.dart';
import 'package:control_contable/features/dashboard/data/dashboard_repository.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/history/application/history.dart';
import 'package:control_contable/features/history/data/history_repository.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/tax_engine/application/breakdown.dart';
import 'package:control_contable/features/tax_engine/application/estimate.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../features/dashboard/pump_dashboard.dart';
import '../../features/expenses/expenses_fakes.dart';
import '../../features/history/pump_history.dart';
import '../../features/incomes/incomes_fakes.dart';
import '../../features/tax_engine/pump_estimate.dart';

/// One app-wide signal for "the data the tax figures derive from changed"
/// (M8-07 device walk F1–F3): the record writes bump it, and every derived
/// read watches it, so a screen on top or below reads again whichever
/// screen the write (or its Undo) came from.
void main() {
  late ProviderContainer container;
  late FakeIncomesRepository incomes;
  late FakeExpensesRepository expenses;
  late FakeHistoryRepository history;
  late FakeDashboardRepository dashboard;
  late FakeTaxEngineRepository engine;

  setUp(() {
    incomes = FakeIncomesRepository();
    expenses = FakeExpensesRepository();
    history = FakeHistoryRepository();
    dashboard = FakeDashboardRepository();
    engine = FakeTaxEngineRepository();
    container = ProviderContainer.test(
      overrides: [
        incomesRepositoryProvider.overrideWithValue(incomes),
        expensesRepositoryProvider.overrideWithValue(expenses),
        historyRepositoryProvider.overrideWithValue(history),
        dashboardRepositoryProvider.overrideWithValue(dashboard),
        taxEngineRepositoryProvider.overrideWithValue(engine),
      ],
    );
  });

  int revision() => container.read(taxDataRevisionProvider);

  group('writes bump the revision', () {
    test('an income created', () async {
      await container
          .read(incomeRecordsProvider)
          .create(
            incomesTaxpayerId,
            const IncomeCreate(
              regimeCode: '612',
              counterpartyKind: CounterpartyKind.pm,
              amountCents: 1000000,
              ivaRate: IvaRate.rate16,
              ivaCents: 160000,
              collectedOn: '2026-10-05',
            ),
          );
      expect(revision(), 1);
    });

    test('an income edited', () async {
      final saved = income();
      incomes.incomes.add(saved);
      await container.read(incomeRecordsProvider).update(
        incomesTaxpayerId,
        saved.id,
        const {'description': 'Otro'},
      );
      expect(revision(), 1);
    });

    test('incomes restored by Undo (F3)', () async {
      final deleted = income();
      incomes.incomes.add(deleted);
      await container.read(incomeRecordsProvider).restore([deleted]);
      expect(revision(), 1);
    });

    test('expenses restored by Undo (F3)', () async {
      final deleted = expense();
      expenses.expenses.add(deleted);
      await container.read(expenseRecordsProvider).restore([deleted]);
      expect(revision(), 1);
    });
  });

  group('derived reads follow the revision', () {
    test('History, Home, the estimate and the breakdown read again', () async {
      const taxpayerId = incomesTaxpayerId;
      final reads = [
        historyYearProvider((taxpayerId: taxpayerId, fiscalYear: 2026)),
        dashboardProvider(taxpayerId),
        estimateProvider((taxpayerId: taxpayerId, period: null)),
        breakdownProvider((
          taxpayerId: taxpayerId,
          period: '2026-03',
          regime: '612',
        )),
      ];
      for (final read in reads) {
        container.listen(read, (_, _) {});
      }
      await Future<void>.delayed(Duration.zero);
      int calls() =>
          history.calls.length +
          dashboard.calls.length +
          engine.calls.length +
          engine.breakdownCalls.length;
      expect(calls(), reads.length);

      container.read(taxDataRevisionProvider.notifier).bump();
      await Future<void>.delayed(Duration.zero);

      expect(calls(), reads.length * 2);
    });
  });
}
