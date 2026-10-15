import 'dart:async';
import 'dart:convert';

import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/income_last_used.dart';
import 'package:control_contable/api/generated/models/income_page.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/period_totals.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/withholding.dart';
import 'package:control_contable/api/generated/models/withholding_preview.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/incomes/domain/income_basis.dart';
import 'package:control_contable/features/incomes/domain/income_preview_query.dart';

import '../../helpers/withholding_fixtures.dart';

/// The taxpayer every incomes test records for.
const String incomesTaxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

int _ids = 0;

/// A fresh income id.
String nextIncomeId() =>
    '0192f0c4-0000-4000-8000-${(++_ids).toString().padLeft(12, '0')}';

/// The derivation of an income nobody withholds from (a persona moral
/// billing a persona moral): ISR-2, IVA-8.
Withholding _nothingWithheld(int netCents) => withholding(
  netCents: netCents,
  explanations: const {
    'ISR-2': 'Como persona moral, no te retienen ISR en esta operación.',
    'IVA-8': 'Tampoco te retienen IVA en esta operación.',
  },
);

/// An income as the server answers it: by default billed to a persona
/// moral, with nothing withheld (the derivation the facts carry is the
/// caller's when it matters). `derived: false` is a record stored before
/// automatic derivation that the server has not derived yet (`withholding:
/// null`, contract 0.12.0).
Income income({
  String? id,
  String regime = '612',
  IncomeBasis basis = IncomeBasis.cash,
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
  CounterpartyKind counterpartyKind = CounterpartyKind.pm,
  OperationKind operationKind = OperationKind.professionalServices,
  bool residentialRent = false,
  bool viaPlatform = false,
  Withholding? withholding,
  bool derived = true,
}) => Income(
  id: id ?? nextIncomeId(),
  taxpayerId: incomesTaxpayerId,
  regimeCode: regime,
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
  pendingCollection: basis == IncomeBasis.accrual && collectedOn == null,
  warnings: warnings,
  createdAt: DateTime.utc(2026, 10, 13),
  updatedAt: DateTime.utc(2026, 10, 13),
  counterpartyKind: counterpartyKind,
  operationKind: operationKind,
  residentialRent: residentialRent,
  viaPlatform: viaPlatform,
  withholding: !derived
      ? null
      : withholding ??
            _nothingWithheld(
              amountCents + ivaCents - isrWithheldCents - ivaWithheldCents,
            ),
);

/// The withholding facts of the last income in a regime
/// (`IncomeLastUsed`); by default a returning user who last billed the
/// public (nothing withheld, no follow-up question).
IncomeLastUsed lastUsed({
  CounterpartyKind? counterpartyKind = CounterpartyKind.pfNoActivityOrPublic,
  OperationKind operationKind = OperationKind.professionalServices,
  bool residentialRent = false,
  bool viaPlatform = false,
}) => IncomeLastUsed(
  counterpartyKind: counterpartyKind,
  operationKind: operationKind,
  residentialRent: residentialRent,
  viaPlatform: viaPlatform,
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
  platformActivity: null,
  lastUsed: IncomeLastUsed(
    counterpartyKind: CounterpartyKind.pm,
    operationKind: OperationKind.professionalServices,
    residentialRent: false,
    viaPlatform: false,
  ),
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
  platformActivity: null,
  lastUsed: IncomeLastUsed(
    counterpartyKind: CounterpartyKind.pfNoActivityOrPublic,
    operationKind: OperationKind.professionalServices,
    residentialRent: false,
    viaPlatform: false,
  ),
);

/// One recorded defaults request.
typedef DefaultsRequest = ({
  String regime,
  String? on,
  PlatformActivity? activity,
});

/// An in-memory incomes server: lists by booking month and regime with
/// keyset-like pages of [pageSize], totals over the whole filtered month,
/// soft delete and restore. Every call is recorded; failures are scripted.
class FakeIncomesRepository implements IncomesRepository {
  /// Every income the server holds, live or deleted.
  final List<Income> incomes = [];

  /// Regimes the server books on the invoice date (`income_basis:
  /// accrual`); every other regime books on the collection date.
  final Set<String> accrualRegimes = {};

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

  /// The defaults answered for each regime (else [defaultsAnswer]).
  final Map<String, IncomeDefaults> defaultsBy = {};

  /// The defaults answered for a regime not in [defaultsBy].
  IncomeDefaults defaultsAnswer = noWithholdingDefaults;

  /// Every defaults request.
  final List<DefaultsRequest> defaultsRequests = [];

  /// Every withholding preview request, in order.
  final List<IncomePreviewQuery> previewRequests = [];

  /// The preview answered for a query (else nothing withheld).
  WithholdingPreview Function(IncomePreviewQuery query)? previewAnswer;

  /// Held open while set: a preview stays in flight until completed.
  Completer<void>? previewGate;

  void _call(String operation, String argument) {
    calls.add('$operation:$argument');
    final failure = failNext.remove(operation) ?? failAlways[operation];
    if (failure != null) throw failure;
  }

  String _bookedOn(Income i) =>
      (accrualRegimes.contains(i.regimeCode) ? i.invoicedOn : i.collectedOn) ??
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
      counterpartyKind: income.counterpartyKind,
      operationKind: income.operationKind ?? OperationKind.professionalServices,
      residentialRent: income.residentialRent ?? false,
      viaPlatform: income.viaPlatform ?? false,
      amountCents: income.amountCents,
      ivaRate: income.ivaRate,
      ivaCents: income.ivaCents,
      isrWithheldCents: 0,
      ivaWithheldCents: 0,
      withholding: _nothingWithheld(income.amountCents + income.ivaCents),
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
    String? on,
    PlatformActivity? platformActivity,
  }) async {
    defaultsRequests.add((regime: regime, on: on, activity: platformActivity));
    _call('defaults', regime);
    return defaultsBy[regime] ?? defaultsAnswer;
  }

  @override
  Future<WithholdingPreview> previewWithholdings({
    required String taxpayerId,
    required IncomePreviewQuery query,
  }) async {
    previewRequests.add(query);
    await previewGate?.future;
    _call('preview', query.counterpartyKind.json!);
    return previewAnswer?.call(query) ??
        nothingWithheld(query.amountCents + query.ivaCents);
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
    // Through JSON text: nested objects (`withholding`) become maps too.
    final stored =
        jsonDecode(jsonEncode(incomes[index])) as Map<String, Object?>;
    final merged = {...stored, ...patch};
    final updated = Income.fromJson(merged);
    incomes[index] = updated;
    return updated;
  }
}
