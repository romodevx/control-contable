import 'package:control_contable/core/fiscal/regimes_held.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';
import 'package:control_contable/features/expenses/presentation/expense_copy.dart';
import 'package:control_contable/features/expenses/presentation/expense_form_page.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The new-expense page (M4-10): its own page, not a dialog (13 inputs,
/// OQ-E2; `ui-forms-and-modals.md`), for the active taxpayer's regimes in
/// effect today.
class ExpenseFormScreen extends ConsumerWidget {
  /// Creates the page.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final active = ref.watch(activeTaxpayerProvider);
    final taxpayer = active.value;
    if (taxpayer == null) {
      return _Status(
        active.isLoading ? l10n.expensesLoading : l10n.expensesNoTaxpayer,
        busy: active.isLoading,
      );
    }
    final today = businessToday(ref.read(clockProvider)());
    final regimes = regimesHeldIn(taxpayer, RecordMonth.ofDate(today).period);
    if (regimes.isEmpty) return _Status(l10n.homeNoRegime);
    return ExpenseFormPage(
      key: ValueKey(taxpayer.id),
      taxpayerId: taxpayer.id,
      createModel: () => ExpenseFormModel(
        taxpayer: taxpayer,
        regimes: regimes,
        today: today,
        paymentMethod: ref.read(lastPaymentMethodProvider),
        copy: expenseFormCopy(l10n),
        newId: ref.read(clientIdProvider),
        lastSupplierFor: (category) =>
            ref.read(supplierMemoryProvider)[category],
      ),
    );
  }
}

class _Status extends StatelessWidget {
  const new(this.message, {this.busy = false});

  final String message;
  final bool busy;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(AppLocalizations.of(context).expenseFormTitle)),
    body: Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: busy ? LoadStatus(message: message, busy: true) : Text(message),
    ),
  );
}
