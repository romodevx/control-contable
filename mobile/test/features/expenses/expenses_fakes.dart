import 'dart:async';

import 'package:control_contable/api/generated/models/deductibility.dart';
import 'package:control_contable/api/generated/models/deductible_source.dart';
import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/expense_category_list.dart';
import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/api/generated/models/expense_page.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/period_totals.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';

/// The taxpayer every expenses test records for.
const String expensesTaxpayerId = '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e';

/// The mandatory legend, exactly as the contract's `Legend` const.
const String expenseLegend =
    'Esto es un aproximado. No todos los gastos aplican para tu giro '
    'comercial.';

int _ids = 0;

/// A fresh expense id.
String nextExpenseId() =>
    '0192f0c5-0000-4000-8000-${(++_ids).toString().padLeft(12, '0')}';

/// An expense as the server answers it.
Expense expense({
  String? id,
  String regime = '612',
  String category = 'software',
  int amountCents = 50000,
  IvaRate ivaRate = IvaRate.rate16,
  int ivaCents = 8000,
  PaymentMethod paymentMethod = PaymentMethod.cardDebit,
  String paidOn = '2026-10-05',
  bool hasCfdi = true,
  String? cfdiUuid,
  int isrWithheldCents = 0,
  int ivaWithheldCents = 0,
  String? description = 'Suscripción de diseño',
  bool deductible = true,
  DeductibleSource source = DeductibleSource.categoryDefault,
  List<Warning> warnings = const [],
}) => Expense(
  id: id ?? nextExpenseId(),
  taxpayerId: expensesTaxpayerId,
  regimeCode: regime,
  categoryCode: category,
  amountCents: amountCents,
  ivaRate: ivaRate,
  ivaCents: ivaCents,
  paymentMethod: paymentMethod,
  paidOn: paidOn,
  hasCfdi: hasCfdi,
  cfdiUuid: cfdiUuid,
  isrWithheldCents: isrWithheldCents,
  ivaWithheldCents: ivaWithheldCents,
  description: description,
  deductible: deductible,
  deductibleSource: source,
  warnings: warnings,
  legend: expenseLegend,
  createdAt: DateTime.utc(2026, 10, 13),
  updatedAt: DateTime.utc(2026, 10, 13),
);

ExpenseCategory _category(
  String code,
  String name,
  Deductibility deductibility, {
  bool? defaultDeductible,
  String? note,
  String? basis = 'LISR 103',
}) => ExpenseCategory(
  code: code,
  name: name,
  deductibility: deductibility,
  defaultDeductible:
      defaultDeductible ?? deductibility == Deductibility.deductible,
  note: note,
  legalBasis: basis,
);

/// The catalog of a business regime (612), a sample of MODULE-expenses §2.2.
List<ExpenseCategory> businessCategories() => [
  _category(
    'mercancia',
    'Mercancía y materias primas',
    Deductibility.deductible,
  ),
  _category(
    'servicios_profesionales',
    'Honorarios y servicios profesionales',
    Deductibility.deductible,
    note: 'Si eres persona moral, revisa las retenciones',
  ),
  _category('software', 'Software y suscripciones', Deductibility.deductible),
  _category(
    'papeleria',
    'Papelería y artículos de oficina',
    Deductibility.deductible,
  ),
  _category(
    'combustible',
    'Combustibles',
    Deductibility.deductible,
    note:
        'Sólo deducible si lo pagas con tarjeta, transferencia o monedero, '
        'sin importar el monto',
    basis: 'LISR 27',
  ),
  _category(
    'restaurantes',
    'Consumo en restaurantes',
    Deductibility.partiallyDeductible,
    defaultDeductible: false,
    note: 'Sólo una parte es deducible',
    basis: 'LISR 28',
  ),
  _category(
    'gastos_personales',
    'Gastos personales (médicos, colegiaturas, donativos)',
    Deductibility.notDeductible,
    note: 'Se aplican en tu declaración anual, no en el pago mensual',
    basis: 'LISR 151',
  ),
];

/// The same catalog resolved for RESICO PF: nothing lowers ISR.
List<ExpenseCategory> resicoCategories() => [
  for (final c in businessCategories())
    c.copyWith(
      deductibility: Deductibility.isrNotApplicable,
      defaultDeductible: false,
    ),
];

/// A categories answer.
ExpenseCategoryList categoryList(List<ExpenseCategory> items) =>
    ExpenseCategoryList(
      items: items,
      ivaRatesAllowed: const [
        IvaRate.rate16,
        IvaRate.rate8,
        IvaRate.rate0,
        IvaRate.exempt,
      ],
      cashPaymentLimitCents: 200000,
      legend: expenseLegend,
    );

/// One recorded categories request.
typedef CategoriesRequest = ({String regime, String? on});

/// An in-memory expenses server: lists by `paid_on` month and regime with
/// pages of [pageSize], totals over the whole filtered month, soft delete
/// and restore. Every call is recorded; failures are scripted.
class FakeExpensesRepository implements ExpensesRepository {
  /// Every expense the server holds, live or deleted.
  final List<Expense> expenses = [];

  /// Ids of the soft-deleted expenses.
  final Set<String> deleted = {};

  /// Items per page.
  int pageSize = 50;

  /// Every call, as `operation:argument`.
  final List<String> calls = [];

  /// Thrown by the next call of the named operation (`list`, `categories`,
  /// `create`, `get`, `update`, `delete`, `restore`), then cleared.
  final Map<String, Exception> failNext = {};

  /// Thrown by every call of the named operation while set.
  final Map<String, Exception> failAlways = {};

  /// Thrown by the restore of the named expense id while set.
  final Map<String, Exception> failRestoreOf = {};

  /// Every create body, retries included.
  final List<ExpenseCreate> creates = [];

  /// Every merge-patch body sent to update, in order.
  final List<Map<String, Object?>> patches = [];

  /// Warnings of the next create or update answer.
  List<Warning> warnNext = const [];

  /// The categories answered per regime (else [businessCategories]).
  final Map<String, List<ExpenseCategory>> categoriesBy = {
    '626': resicoCategories(),
  };

  /// Every categories request.
  final List<CategoriesRequest> categoryRequests = [];

  /// Held open while set: a create stays in flight until completed.
  Completer<void>? createGate;

  void _call(String operation, String argument) {
    calls.add('$operation:$argument');
    final failure = failNext.remove(operation) ?? failAlways[operation];
    if (failure != null) throw failure;
  }

  @override
  Future<ExpensePage> list({
    required String taxpayerId,
    required String period,
    String? regime,
    String? cursor,
  }) async {
    _call('list', '$period ${regime ?? '*'} ${cursor ?? '-'}');
    final month = [
      for (final e in expenses)
        if (!deleted.contains(e.id) &&
            e.paidOn.startsWith(period) &&
            (regime == null || e.regimeCode == regime))
          e,
    ]..sort((a, b) => b.paidOn.compareTo(a.paidOn));
    final start = cursor == null ? 0 : int.parse(cursor);
    final end = (start + pageSize).clamp(0, month.length);
    return ExpensePage(
      items: month.sublist(start, end),
      nextCursor: end < month.length ? '$end' : null,
      totals: PeriodTotals(
        count: month.length,
        amountCents: month.fold(0, (sum, e) => sum + e.amountCents),
        ivaCents: month.fold(0, (sum, e) => sum + e.ivaCents),
        isrWithheldCents: month.fold(0, (sum, e) => sum + e.isrWithheldCents),
        ivaWithheldCents: month.fold(0, (sum, e) => sum + e.ivaWithheldCents),
      ),
      legend: expenseLegend,
    );
  }

  @override
  Future<ExpenseCategoryList> categories({
    required String taxpayerId,
    required String regime,
    String? on,
  }) async {
    categoryRequests.add((regime: regime, on: on));
    _call('categories', '$regime ${on ?? '-'}');
    return categoryList(categoriesBy[regime] ?? businessCategories());
  }

  @override
  Future<Expense> create({
    required String taxpayerId,
    required ExpenseCreate expense,
  }) async {
    creates.add(expense);
    await createGate?.future;
    _call('create', expense.id ?? '-');
    final stored = expenses.where((e) => e.id == expense.id).firstOrNull;
    if (stored != null) return stored;
    final created = Expense(
      id: expense.id ?? nextExpenseId(),
      taxpayerId: taxpayerId,
      regimeCode: expense.regimeCode,
      categoryCode: expense.categoryCode,
      amountCents: expense.amountCents,
      ivaRate: expense.ivaRate,
      ivaCents: expense.ivaCents,
      paymentMethod: expense.paymentMethod,
      paidOn: expense.paidOn,
      hasCfdi: expense.hasCfdi,
      cfdiUuid: expense.cfdiUuid,
      isrWithheldCents: expense.isrWithheldCents ?? 0,
      ivaWithheldCents: expense.ivaWithheldCents ?? 0,
      description: expense.description,
      deductible: expense.deductible ?? true,
      deductibleSource: expense.deductible == null
          ? DeductibleSource.categoryDefault
          : DeductibleSource.user,
      warnings: const [],
      legend: expenseLegend,
      createdAt: DateTime.utc(2026, 10, 13),
      updatedAt: DateTime.utc(2026, 10, 13),
    );
    expenses.add(created);
    return created.copyWith(warnings: _takeWarnings());
  }

  List<Warning> _takeWarnings() {
    final warnings = warnNext;
    warnNext = const [];
    return warnings;
  }

  @override
  Future<Expense> get({
    required String taxpayerId,
    required String expenseId,
  }) async {
    _call('get', expenseId);
    return expenses.firstWhere((e) => e.id == expenseId);
  }

  @override
  Future<Expense> update({
    required String taxpayerId,
    required String expenseId,
    required Map<String, Object?> patch,
  }) async {
    patches.add(patch);
    _call('update', expenseId);
    final index = expenses.indexWhere((e) => e.id == expenseId);
    final merged = {...expenses[index].toJson(), ...patch};
    if (patch.containsKey('deductible')) {
      merged['deductible_source'] = patch['deductible'] == null
          ? DeductibleSource.categoryDefault.json
          : DeductibleSource.user.json;
      merged['deductible'] ??= true;
    }
    final updated = Expense.fromJson(merged);
    expenses[index] = updated;
    return updated.copyWith(warnings: _takeWarnings());
  }

  @override
  Future<void> delete({
    required String taxpayerId,
    required String expenseId,
  }) async {
    _call('delete', expenseId);
    deleted.add(expenseId);
  }

  @override
  Future<Expense> restore({
    required String taxpayerId,
    required String expenseId,
  }) async {
    _call('restore', expenseId);
    final refusal = failRestoreOf[expenseId];
    if (refusal != null) throw refusal;
    deleted.remove(expenseId);
    return expenses.firstWhere((e) => e.id == expenseId);
  }
}
