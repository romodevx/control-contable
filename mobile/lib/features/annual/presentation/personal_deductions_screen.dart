import 'dart:async';

import 'package:control_contable/api/generated/models/personal_deduction_list.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/domain/deduction_rules.dart';
import 'package:control_contable/features/annual/presentation/annual_input_list.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/annual/presentation/deduction_category_group.dart';
import 'package:control_contable/features/annual/presentation/deduction_delete.dart';
import 'package:control_contable/features/annual/presentation/deductions_summary.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// The personal deductions of a year (ANN-19): registered vs applied with
/// the caps that bind, one group per category, delete + Undo, and the
/// add button (the form is its own screen).
class PersonalDeductionsScreen extends ConsumerWidget {
  /// The deductions of [taxpayerId] in [fiscalYear].
  const new({required this.taxpayerId, required this.fiscalYear, super.key});

  /// Whose deductions.
  final String taxpayerId;

  /// Which year.
  final int fiscalYear;

  AnnualQuery get _query => (taxpayerId: taxpayerId, fiscalYear: fiscalYear);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final list = ref.watch(personalDeductionsProvider(_query));
    return Scaffold(
      appBar: AppBar(title: Text(l10n.annualInputDeductions)),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _add(context, ref, list.value),
        icon: const Icon(Icons.add),
        label: Text(l10n.annualDeductionsAdd),
      ),
      body: annualInputBody(
        list,
        loading: l10n.annualDeductionsLoading,
        failed: l10n.annualDeductionsLoadFailed,
        onRetry: () => ref.invalidate(personalDeductionsProvider(_query)),
        data: (deductions) => _List(query: _query, deductions: deductions),
      ),
    );
  }

  void _add(BuildContext context, WidgetRef ref, PersonalDeductionList? list) {
    if (blockedAsReadOnly(context, ref)) return;
    if (_isFull(list)) {
      ref
          .read(annualDeductionsUndoProvider)
          .notify(
            ScaffoldMessenger.of(context),
            AppLocalizations.of(context).annualDeductionsFull('$fiscalYear'),
          );
      return;
    }
    unawaited(
      context.push<void>(
        AnnualRoutes.newPersonalDeduction(taxpayerId, fiscalYear),
      ),
    );
  }
}

bool _isFull(PersonalDeductionList? list) =>
    (list?.items.length ?? 0) >= deductionsPerYearLimit;

class _List extends ConsumerWidget {
  const new({required this.query, required this.deductions});

  final AnnualQuery query;
  final PersonalDeductionList deductions;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final year = '${query.fiscalYear}';
    final registered = deductions.totals.fold(
      0,
      (sum, total) => sum + total.registeredCents,
    );
    return AnnualInputList(
      hasFloatingAction: true,
      children: [
        DeductionsSummary(
          key: const ValueKey('deductions-summary'),
          query: query,
          registeredCents: registered,
        ),
        if (_isFull(deductions))
          Text(l10n.annualDeductionsFull(year), style: text.bodyLarge),
        if (deductions.items.isEmpty)
          Text(l10n.annualDeductionsEmpty(year), style: text.bodyLarge),
        for (final total in deductions.totals)
          DeductionCategoryGroup(
            key: ValueKey('deduction-group-${total.category.json}'),
            total: total,
            items: deductionsOf(total.category, deductions.items),
            onDelete: (deduction) =>
                unawaited(deletePersonalDeduction(context, ref, deduction)),
          ),
      ],
    );
  }
}
