import 'package:control_contable/core/fiscal/regimes_held.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';
import 'package:control_contable/features/incomes/presentation/income_form_copy.dart';
import 'package:control_contable/features/incomes/presentation/income_form_page.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The new-income page (M3-09): its own page, not a dialog (11 inputs;
/// `ui-forms-and-modals.md`), for the active taxpayer's regimes in effect
/// today.
class IncomeFormScreen extends ConsumerWidget {
  /// Creates the page.
  const new({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final active = ref.watch(activeTaxpayerProvider);
    final taxpayer = active.value;
    if (taxpayer == null) {
      return Scaffold(
        appBar: AppBar(title: Text(l10n.incomeFormTitle)),
        body: Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: active.isLoading
              ? LoadStatus(message: l10n.incomesLoading, busy: true)
              : Text(l10n.incomesNoTaxpayer),
        ),
      );
    }
    final today = businessToday(ref.read(clockProvider)());
    final regimes = regimesHeldIn(taxpayer, RecordMonth.ofDate(today).period);
    if (regimes.isEmpty) {
      return Scaffold(
        appBar: AppBar(title: Text(l10n.incomeFormTitle)),
        body: Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: Text(l10n.homeNoRegime),
        ),
      );
    }
    return IncomeFormPage(
      key: ValueKey(taxpayer.id),
      taxpayerId: taxpayer.id,
      createModel: () => IncomeFormModel(
        regimes: regimes,
        today: today,
        counterparty: ref.read(lastCounterpartyProvider),
        copy: incomeFormCopy(l10n),
        newId: ref.read(clientIdProvider),
      ),
    );
  }
}
