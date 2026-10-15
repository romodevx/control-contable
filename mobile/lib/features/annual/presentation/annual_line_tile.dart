import 'package:control_contable/api/generated/models/annual_breakdown_line.dart';
import 'package:control_contable/api/generated/models/annual_breakdown_line_operation.dart';
import 'package:control_contable/features/tax_engine/presentation/computation_line_tile.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// `AnnualLineSource` values of the contract.
const String _records = 'records';
const String _computed = 'computed';
const String _salary = 'salary_statements';
const String _deductions = 'personal_deductions';
const String _annualInputs = 'annual_inputs';
const String _filed = 'filed_payment';
const String _estimated = 'estimated_by_app';
const String _taxInputPrefix = 'tax_input:';
const String _paramsPrefix = 'params:';

/// One step of an annual breakdown (ANN-18), shown by the monthly
/// breakdown's [ComputationLineTile]: where each figure comes from (a
/// provisional payment filed, taken from the tax inputs or estimated by
/// the app), the regime it belongs to, and every "Tope aplicado" or
/// excluded deduction as a remark of its own.
class AnnualLineTile extends StatelessWidget {
  /// The tile of [line]; [dependsOnMissingInputs] marks its amount.
  const new({
    required this.line,
    required this.dependsOnMissingInputs,
    super.key,
  });

  /// The line.
  final AnnualBreakdownLine line;

  /// Whether the section is incomplete and this figure is not a plain sum
  /// of the records.
  final bool dependsOnMissingInputs;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final regime = line.regimeCode;
    return ComputationLineTile(
      line: ComputationLine(
        concept: line.concept,
        amountCents: line.amountCents,
        role: switch (line.operation) {
          AnnualBreakdownLineOperation.subtract =>
            ComputationLineRole.subtraction,
          AnnualBreakdownLineOperation.result => ComputationLineRole.result,
          AnnualBreakdownLineOperation.roundingAdjustment =>
            ComputationLineRole.rounding,
          AnnualBreakdownLineOperation.info => ComputationLineRole.note,
          _ => ComputationLineRole.step,
        },
        rate: line.rate,
        legalBasis: line.legalBasis,
        sourceLabel: annualSourceLabel(l10n, line.source),
        regimeLabel: regime == null ? null : l10n.estimateRegimeCode(regime),
        needsCpaReview: line.needsCpaReview,
      ),
      dependsOnMissingInputs: dependsOnMissingInputs && line.source != _records,
    );
  }
}

/// Where an annual figure comes from, in words; `null` for a source this
/// app version does not know.
String? annualSourceLabel(AppLocalizations l10n, String source) =>
    switch (source) {
      _records => l10n.breakdownSourceRecords,
      _computed => l10n.breakdownSourceComputed,
      _salary => l10n.annualSourceSalary,
      _deductions => l10n.annualSourceDeductions,
      _annualInputs => l10n.annualSourceFacts,
      _filed => l10n.annualSourceFiled,
      _estimated => l10n.annualSourceEstimated,
      _ when source.startsWith(_taxInputPrefix) => l10n.breakdownSourceInput,
      _ when source.startsWith(_paramsPrefix) => l10n.breakdownSourceParams,
      _ => null,
    };
