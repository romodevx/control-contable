import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/presentation/annual_input_list.dart';
import 'package:control_contable/features/annual/presentation/filed_payments_form.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The provisional payments filed for one regime in a year (ANN-21): twelve
/// inline rows (one field each, ui-forms-and-modals.md) prefilled with the
/// app's monthly estimates, to confirm one by one or all at once.
class FiledPaymentsScreen extends ConsumerWidget {
  /// The payments of [regime] for [taxpayerId] in [fiscalYear].
  const new({
    required this.taxpayerId,
    required this.fiscalYear,
    required this.regime,
    super.key,
  });

  /// Whose payments.
  final String taxpayerId;

  /// Which year.
  final int fiscalYear;

  /// Which regime (`612`).
  final String regime;

  FiledPaymentsQuery get _query =>
      (taxpayerId: taxpayerId, fiscalYear: fiscalYear, regime: regime);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.annualInputPayments),
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(kMinInteractiveDimension),
          child: Text(l10n.estimateRegimeCode(regime)),
        ),
      ),
      body: annualInputBody(
        ref.watch(filedPaymentsProvider(_query)),
        loading: l10n.annualPaymentsLoading,
        failed: l10n.annualPaymentsLoadFailed,
        onRetry: () => ref.invalidate(filedPaymentsProvider(_query)),
        // A saved set reloads as a new form; a refused one keeps the typing.
        data: (payments) => FiledPaymentsForm(
          key: ValueKey(payments.updatedAt),
          query: _query,
          payments: payments,
        ),
      ),
    );
  }
}
