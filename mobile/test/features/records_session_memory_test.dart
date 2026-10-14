import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../helpers/recording_error_reporter.dart';
import '../helpers/session_fakes.dart';

const _income = IncomeCreate(
  regimeCode: '612',
  counterpartyType: CounterpartyType.personaMoral,
  amountCents: 1000000,
  ivaRate: IvaRate.rate16,
  ivaCents: 160000,
  collectedOn: '2026-10-05',
);

const _expense = ExpenseCreate(
  regimeCode: '612',
  categoryCode: 'software',
  amountCents: 50000,
  ivaRate: IvaRate.rate16,
  ivaCents: 8000,
  paymentMethod: PaymentMethod.cash,
  paidOn: '2026-10-05',
  hasCfdi: true,
);

/// QA M3/M4 D4: what the record forms remember "in this session" (the last
/// payer, the last payment method, the most used categories, the Undo
/// stacks) belongs to the signed-in user. Signing out forgets it, so the
/// next account on the device starts from the defaults, not from the
/// previous user's choices (seen on the emulator: Bea's first income
/// started on Ana's last payer, persona moral).
void main() {
  late Session session;
  late ProviderContainer container;

  setUp(() async {
    session = Session(
      store: InMemoryRefreshTokenStore(),
      rotate: (_) async => tokenPair(access: 'a', refresh: 'r'),
      reporter: RecordingErrorReporter(),
    );
    addTearDown(session.dispose);
    await session.start(tokenPair(access: 'a1', refresh: 'r1'));
    container = ProviderContainer(
      overrides: [sessionProvider.overrideWithValue(session)],
    );
    addTearDown(container.dispose);
  });

  test(
    'signing out forgets the last payer, payment method and categories',
    () async {
      container
          .read(lastCounterpartyProvider.notifier)
          .rememberPayerOf(_income);
      container
          .read(lastPaymentMethodProvider.notifier)
          .rememberMethodOf(_expense);
      container.read(categoryUsageProvider.notifier).used('software');
      expect(
        container.read(lastCounterpartyProvider),
        CounterpartyType.personaMoral,
      );

      await session.end();

      expect(
        container.read(lastCounterpartyProvider),
        CounterpartyType.personaFisica,
      );
      expect(
        container.read(lastPaymentMethodProvider),
        PaymentMethod.cardDebit,
      );
      expect(container.read(categoryUsageProvider), isEmpty);
    },
  );

  test('signing out drops the Undo stacks', () async {
    final incomes = container.read(incomesUndoProvider);
    final expenses = container.read(expensesUndoProvider);

    await session.end();

    expect(container.read(incomesUndoProvider), isNot(same(incomes)));
    expect(container.read(expensesUndoProvider), isNot(same(expenses)));
  });

  test('a token refresh keeps what the session remembers', () async {
    container.read(lastCounterpartyProvider.notifier).rememberPayerOf(_income);

    await session.refresh();

    expect(
      container.read(lastCounterpartyProvider),
      CounterpartyType.personaMoral,
    );
  });
}
