import 'dart:async';

import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/income_page.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/period_totals.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/incomes/domain/income_basis.dart';

/// The taxpayer every incomes test records for.
const String incomesTaxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

int _ids = 0;

/// A fresh income id.
String nextIncomeId() =>
    '0192f0c4-0000-4000-8000-${(++_ids).toString().padLeft(12, '0')}';

/// An income as the server answers it.
Income income({
  String? id,
  String regime = '612',
  CounterpartyType counterparty = CounterpartyType.personaMoral,
  int amountCents = 1000000,
  IvaRate ivaRate = IvaRate.rate16,
  int ivaCents = 160000,
  int isrWithheldCents = 0,
  int ivaWithheldCents = 0,
  String? invoicedOn,
  String? collectedOn = '2026-10-05',
  String? description = 'Diseño de logotipo',
  String? cfdiUuid,
  List<Warning> warnings = const [],
}) => Income(
  id: id ?? nextIncomeId(),
  taxpayerId: incomesTaxpayerId,
  regimeCode: regime,
  counterpartyType: counterparty,
  amountCents: amountCents,
  ivaRate: ivaRate,
  ivaCents: ivaCents,
  isrWithheldCents: isrWithheldCents,
  ivaWithheldCents: ivaWithheldCents,
  invoicedOn: invoicedOn,
  collectedOn: collectedOn,
  platformActivity: null,
  cfdiUuid: cfdiUuid,
  description: description,
  pendingCollection:
      incomeBasisOf(regime) == IncomeBasis.accrual && collectedOn == null,
  warnings: warnings,
  createdAt: DateTime.utc(2026, 10, 13),
  updatedAt: DateTime.utc(2026, 10, 13),
);

/// The defaults of a 612 income from a persona moral (US-I1).
const IncomeDefaults moralDefaults612 = IncomeDefaults(
  ivaRatesAllowed: [
    IvaRate.rate16,
    IvaRate.rate8,
    IvaRate.rate0,
    IvaRate.exempt,
  ],
  suggestedIvaRate: IvaRate.rate16,
  isrWithholdingAllowed: true,
  isrWithholdingRate: '0.1',
  ivaWithholdingAllowed: true,
  ivaWithholdingRate: '0.666667',
  platformActivity: null,
);

/// The defaults when nothing may be withheld (a persona física payer).
const IncomeDefaults noWithholdingDefaults = IncomeDefaults(
  ivaRatesAllowed: [
    IvaRate.rate16,
    IvaRate.rate8,
    IvaRate.rate0,
    IvaRate.exempt,
  ],
  suggestedIvaRate: IvaRate.rate16,
  isrWithholdingAllowed: false,
  isrWithholdingRate: null,
  ivaWithholdingAllowed: false,
  ivaWithholdingRate: null,
  platformActivity: null,
);

/// One recorded defaults request.
typedef DefaultsRequest = ({
  String regime,
  CounterpartyType counterparty,
  String? on,
  PlatformActivity? activity,
});

/// An in-memory incomes server: lists by booking month and regime with
/// keyset-like pages of [pageSize], totals over the whole filtered month,
/// soft delete and restore. Every call is recorded; failures are scripted.
class FakeIncomesRepository implements IncomesRepository {
  /// Every income the server holds, live or deleted.
  final List<Income> incomes = [];

  /// Ids of the soft-deleted incomes.
  final Set<String> deleted = {};

  /// Items per page.
  int pageSize = 50;

  /// Every call, as `operation:argument`.
  final List<String> calls = [];

  /// Thrown by the next call of the named operation (`list`, `create`,
  /// `defaults`, `get`, `update`, `delete`, `restore`), then cleared.
  final Map<String, Exception> failNext = {};

  /// Thrown by every call of the named operation while set.
  final Map<String, Exception> failAlways = {};

  /// Every create body, retries included.
  final List<IncomeCreate> creates = [];

  /// Every merge-patch body sent to update, in order.
  final List<Map<String, Object?>> patches = [];

  /// Held open while set: a create stays in flight until completed.
  Completer<void>? createGate;

  /// Held open while set: a list load stays in flight until completed.
  Completer<void>? listGate;

  /// Warnings of the next create answer.
  List<Warning> warnNext = const [];

  /// The defaults answered for each counterparty (else [moralDefaults612]).
  final Map<CounterpartyType, IncomeDefaults> defaultsBy = {
    CounterpartyType.personaFisica: noWithholdingDefaults,
    CounterpartyType.publicoGeneral: noWithholdingDefaults,
  };

  /// Every defaults request.
  final List<DefaultsRequest> defaultsRequests = [];

  void _call(String operation, String argument) {
    calls.add('$operation:$argument');
    final failure = failNext.remove(operation) ?? failAlways[operation];
    if (failure != null) throw failure;
  }

  static String _bookedOn(Income i) =>
      (incomeBasisOf(i.regimeCode) == IncomeBasis.accrual
          ? i.invoicedOn
          : i.collectedOn) ??
      '';

  @override
  Future<IncomePage> list({
    required String taxpayerId,
    required String period,
    String? regime,
    String? cursor,
  }) async {
    await listGate?.future;
    _call('list', '$period ${regime ?? '*'} ${cursor ?? '-'}');
    final month = [
      for (final i in incomes)
        if (!deleted.contains(i.id) &&
            _bookedOn(i).startsWith(period) &&
            (regime == null || i.regimeCode == regime))
          i,
    ]..sort((a, b) => _bookedOn(b).compareTo(_bookedOn(a)));
    final start = cursor == null ? 0 : int.parse(cursor);
    final end = (start + pageSize).clamp(0, month.length);
    return IncomePage(
      items: month.sublist(start, end),
      nextCursor: end < month.length ? '$end' : null,
      totals: PeriodTotals(
        count: month.length,
        amountCents: month.fold(0, (sum, i) => sum + i.amountCents),
        ivaCents: month.fold(0, (sum, i) => sum + i.ivaCents),
        isrWithheldCents: month.fold(0, (sum, i) => sum + i.isrWithheldCents),
        ivaWithheldCents: month.fold(0, (sum, i) => sum + i.ivaWithheldCents),
      ),
    );
  }

  @override
  Future<Income> create({
    required String taxpayerId,
    required IncomeCreate income,
  }) async {
    creates.add(income);
    await createGate?.future;
    _call('create', income.id ?? '-');
    final stored = incomes.where((i) => i.id == income.id).firstOrNull;
    if (stored != null) return stored;
    final created = Income(
      id: income.id ?? nextIncomeId(),
      taxpayerId: taxpayerId,
      regimeCode: income.regimeCode,
      counterpartyType: income.counterpartyType,
      amountCents: income.amountCents,
      ivaRate: income.ivaRate,
      ivaCents: income.ivaCents,
      isrWithheldCents: income.isrWithheldCents ?? 0,
      ivaWithheldCents: income.ivaWithheldCents ?? 0,
      invoicedOn: income.invoicedOn,
      collectedOn: income.collectedOn,
      platformActivity: income.platformActivity,
      cfdiUuid: income.cfdiUuid,
      description: income.description,
      pendingCollection: false,
      warnings: const [],
      createdAt: DateTime.utc(2026, 10, 13),
      updatedAt: DateTime.utc(2026, 10, 13),
    );
    incomes.add(created);
    final warnings = warnNext;
    warnNext = const [];
    return created.copyWith(warnings: warnings);
  }

  @override
  Future<IncomeDefaults> defaults({
    required String taxpayerId,
    required String regime,
    required CounterpartyType counterparty,
    String? on,
    PlatformActivity? platformActivity,
  }) async {
    defaultsRequests.add((
      regime: regime,
      counterparty: counterparty,
      on: on,
      activity: platformActivity,
    ));
    _call('defaults', '$regime ${counterparty.json}');
    return defaultsBy[counterparty] ?? moralDefaults612;
  }

  @override
  Future<void> delete({
    required String taxpayerId,
    required String incomeId,
  }) async {
    _call('delete', incomeId);
    deleted.add(incomeId);
  }

  @override
  Future<Income> restore({
    required String taxpayerId,
    required String incomeId,
  }) async {
    _call('restore', incomeId);
    deleted.remove(incomeId);
    return incomes.firstWhere((i) => i.id == incomeId);
  }

  @override
  Future<Income> get({
    required String taxpayerId,
    required String incomeId,
  }) async {
    _call('get', incomeId);
    return incomes.firstWhere((i) => i.id == incomeId);
  }

  @override
  Future<Income> update({
    required String taxpayerId,
    required String incomeId,
    required Map<String, Object?> patch,
  }) async {
    patches.add(patch);
    _call('update', incomeId);
    final index = incomes.indexWhere((i) => i.id == incomeId);
    final merged = {...incomes[index].toJson(), ...patch};
    final updated = Income.fromJson(merged);
    incomes[index] = updated;
    return updated;
  }
}
