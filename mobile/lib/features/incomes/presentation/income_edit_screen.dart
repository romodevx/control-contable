import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/fiscal/regimes_held.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/incomes/presentation/income_form_copy.dart';
import 'package:control_contable/features/incomes/presentation/income_form_page.dart';
import 'package:control_contable/features/incomes/presentation/income_labels.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which income an edit loads: an income id of the active taxpayer.
typedef IncomeRef = ({String taxpayerId, String incomeId});

/// One stored income, read when the edit page is opened by its location
/// (no income handed over by the list).
final FutureProviderFamily<Income, IncomeRef> storedIncomeProvider =
    FutureProvider.autoDispose.family<Income, IncomeRef>(
      (ref, key) => ref
          .read(incomesRepositoryProvider)
          .get(taxpayerId: key.taxpayerId, incomeId: key.incomeId),
      name: 'storedIncomeProvider',
      retry: (retryCount, error) => null,
    );

/// The edit-income page (M3-09, US-I3): the same form page as a new
/// income, prefilled with the stored one. [initial] is the income the list
/// handed over; without it (a deep link) the income is loaded first
/// (`storedIncomeProvider`).
class IncomeEditScreen extends ConsumerWidget {
  /// The page for [incomeId].
  const new({required this.incomeId, this.initial, super.key});

  /// Which income.
  final String incomeId;

  /// The income as the list showed it, when opened from the list.
  final Income? initial;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final taxpayer = ref.watch(activeTaxpayerProvider).value;
    if (taxpayer == null) return _Status(l10n.incomesLoading, busy: true);
    final handed = initial;
    if (handed != null && handed.id == incomeId) {
      return _form(ref, l10n, taxpayer, handed);
    }
    final stored = ref.watch(
      storedIncomeProvider((taxpayerId: taxpayer.id, incomeId: incomeId)),
    );
    return switch (stored) {
      AsyncData(:final value) => _form(ref, l10n, taxpayer, value),
      AsyncError() => _Status(l10n.incomesLoadFailed),
      _ => _Status(l10n.incomesLoading, busy: true),
    };
  }

  Widget _form(
    WidgetRef ref,
    AppLocalizations l10n,
    Taxpayer taxpayer,
    Income income,
  ) {
    final regimes = _regimesFor(taxpayer, income);
    if (regimes.isEmpty) return _Status(l10n.homeNoRegime);
    return IncomeFormPage(
      key: ValueKey(income.id),
      taxpayerId: taxpayer.id,
      createModel: () => IncomeFormModel(
        regimes: regimes,
        today: businessToday(ref.read(clockProvider)()),
        isMoral: taxpayer.personType == PersonType.moral,
        copy: incomeFormCopy(l10n),
        newId: ref.read(clientIdProvider),
        editing: income,
      ),
    );
  }

  /// The regimes held in the income's month, its own regime always first
  /// (it may have ended since).
  static List<TaxpayerRegime> _regimesFor(Taxpayer taxpayer, Income income) {
    final period = RecordMonth.ofDate(bookingDateOf(income, taxpayer.regimes))
        .period;
    final held = regimesHeldIn(taxpayer, period);
    final own = taxpayer.regimes
        .where((r) => r.regimeCode == income.regimeCode)
        .firstOrNull;
    return [
      ?own,
      for (final regime in held)
        if (regime.regimeCode != income.regimeCode) regime,
    ];
  }
}

class _Status extends StatelessWidget {
  const new(this.message, {this.busy = false});

  final String message;
  final bool busy;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text(AppLocalizations.of(context).incomeEditTitle)),
    body: Padding(
      padding: const EdgeInsets.all(AppSpacing.gutter),
      child: busy ? LoadStatus(message: message, busy: true) : Text(message),
    ),
  );
}
