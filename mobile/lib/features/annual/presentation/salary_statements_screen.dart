import 'dart:async';

import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/api/generated/models/salary_statement_list.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/presentation/annual_input_card.dart';
import 'package:control_contable/features/annual/presentation/annual_input_list.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/annual/presentation/salary_statement_sheet.dart';
import 'package:control_contable/features/annual/presentation/salary_statement_tile.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The salary constancias of a year (ANN-20): the year totals, one row per
/// employer (tap to edit), and the add button; the form is a sheet.
class SalaryStatementsScreen extends ConsumerWidget {
  /// The constancias of [taxpayerId] in [fiscalYear].
  const new({required this.taxpayerId, required this.fiscalYear, super.key});

  /// Whose constancias.
  final String taxpayerId;

  /// Which year.
  final int fiscalYear;

  AnnualQuery get _query => (taxpayerId: taxpayerId, fiscalYear: fiscalYear);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(l10n.annualInputSalary)),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _open(context, ref),
        icon: const Icon(Icons.add),
        label: Text(l10n.annualSalaryAdd),
      ),
      body: annualInputBody(
        ref.watch(salaryStatementsProvider(_query)),
        loading: l10n.annualSalaryLoading,
        failed: l10n.annualSalaryLoadFailed,
        onRetry: () => ref.invalidate(salaryStatementsProvider(_query)),
        data: (list) => _List(
          list: list,
          onEdit: (statement) => _open(context, ref, editing: statement),
        ),
      ),
    );
  }

  void _open(BuildContext context, WidgetRef ref, {SalaryStatement? editing}) {
    if (editing == null && blockedAsReadOnly(context, ref)) return;
    unawaited(
      showSalaryStatementSheet(
        context,
        taxpayerId: taxpayerId,
        fiscalYear: fiscalYear,
        editing: editing,
      ),
    );
  }
}

class _List extends StatelessWidget {
  const new({required this.list, required this.onEdit});

  final SalaryStatementList list;
  final ValueChanged<SalaryStatement> onEdit;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final totals = list.totals;
    return AnnualInputList(
      hasFloatingAction: true,
      children: [
        if (list.items.isEmpty)
          Text(
            l10n.annualSalaryEmpty('${list.fiscalYear}'),
            style: text.bodyLarge,
          )
        else
          AnnualInputCard(
            key: const ValueKey('salary-totals'),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Semantics(
                  header: true,
                  child: Text(l10n.annualSalaryTotals, style: text.titleSmall),
                ),
                for (final (label, cents) in [
                  (l10n.annualSalaryAccumulable, totals.accumulableCents),
                  (l10n.annualSalaryExempt, totals.exemptCents),
                  (l10n.annualSalaryIsrWithheld, totals.isrWithheldCents),
                ])
                  AnnualAmountRow(label: label, amount: formatPesos(cents)),
              ],
            ),
          ),
        for (final (index, statement) in list.items.indexed)
          SalaryStatementTile(
            key: ValueKey('salary-${statement.id}'),
            statement: statement,
            name:
                statement.employerLabel ?? l10n.annualSalaryUnnamed(index + 1),
            onEdit: () => onEdit(statement),
          ),
      ],
    );
  }
}
