import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/fiscal/regimes_held.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/expenses/presentation/expense_copy.dart';
import 'package:control_contable/features/expenses/presentation/expense_form_page.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which expense an edit loads: an expense id of the active taxpayer.
typedef ExpenseRef = ({String taxpayerId, String expenseId});

/// One stored expense, read when the edit page is opened by its location
/// (no expense handed over by the list).
final FutureProviderFamily<Expense, ExpenseRef> storedExpenseProvider =
    FutureProvider.autoDispose.family<Expense, ExpenseRef>(
      (ref, key) => ref
          .read(expensesRepositoryProvider)
          .get(taxpayerId: key.taxpayerId, expenseId: key.expenseId),
      name: 'storedExpenseProvider',
      retry: (retryCount, error) => null,
    );

/// The edit-expense page (M4-10, US-E2): the same form page as a new
/// expense, prefilled with the stored one. [initial] is the expense the
/// list handed over; without it (a deep link) the expense is loaded first.
class ExpenseEditScreen extends ConsumerWidget {
  /// The page for [expenseId].
  const new({required this.expenseId, this.initial, super.key});

  /// Which expense.
  final String expenseId;

  /// The expense as the list showed it, when opened from the list.
  final Expense? initial;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final taxpayer = ref.watch(activeTaxpayerProvider).value;
    if (taxpayer == null) return _Status(l10n.expensesLoading, busy: true);
    final handed = initial;
    if (handed != null && handed.id == expenseId) {
      return _form(ref, l10n, taxpayer, handed);
    }
    final stored = ref.watch(
      storedExpenseProvider((taxpayerId: taxpayer.id, expenseId: expenseId)),
    );
    return switch (stored) {
      AsyncData(:final value) => _form(ref, l10n, taxpayer, value),
      AsyncError() => _Status(l10n.expensesLoadFailed),
      _ => _Status(l10n.expensesLoading, busy: true),
    };
  }

  Widget _form(
    WidgetRef ref,
    AppLocalizations l10n,
    Taxpayer taxpayer,
    Expense expense,
  ) {
    final regimes = _regimesFor(taxpayer, expense);
    if (regimes.isEmpty) return _Status(l10n.homeNoRegime);
    return ExpenseFormPage(
      key: ValueKey(expense.id),
      taxpayerId: taxpayer.id,
      createModel: () => ExpenseFormModel(
        taxpayer: taxpayer,
        regimes: regimes,
        today: businessToday(ref.read(clockProvider)()),
        paymentMethod: ref.read(lastPaymentMethodProvider),
        copy: expenseFormCopy(l10n),
        newId: ref.read(clientIdProvider),
        editing: expense,
      ),
    );
  }

  /// The regimes held in the expense's month, its own regime always first
  /// (it may have ended since).
  static List<TaxpayerRegime> _regimesFor(Taxpayer taxpayer, Expense expense) {
    final period = RecordMonth.ofDate(parseBusinessDate(expense.paidOn)).period;
    final own = taxpayer.regimes
        .where((r) => r.regimeCode == expense.regimeCode)
        .firstOrNull;
    return [
      ?own,
      for (final regime in regimesHeldIn(taxpayer, period))
        if (regime.regimeCode != expense.regimeCode) regime,
    ];
  }
}

class _Status extends StatelessWidget {
  const new(this.message, {this.busy = false});

  final String message;
  final bool busy;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(AppLocalizations.of(context).expenseEditTitle)),
    body: Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: busy ? LoadStatus(message: message, busy: true) : Text(message),
    ),
  );
}
