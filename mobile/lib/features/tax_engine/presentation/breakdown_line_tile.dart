import 'package:control_contable/api/generated/models/breakdown_line.dart';
import 'package:control_contable/api/generated/models/breakdown_line_operation.dart';
import 'package:control_contable/features/tax_engine/presentation/computation_line_tile.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// `source` prefixes of the contract (`BreakdownLine.source`).
const String _records = 'records';
const String _computed = 'computed';
const String _taxInputPrefix = 'tax_input:';
const String _paramsPrefix = 'params:';

/// One step of a monthly breakdown (M5-26), shown by [ComputationLineTile]:
/// the rounding adjustment is muted and explained instead of cited; a
/// result is ruled off from the steps above it.
class BreakdownLineTile extends StatelessWidget {
  /// The tile of [line]; [dependsOnMissingInputs] marks its amount.
  const new({
    required this.line,
    required this.dependsOnMissingInputs,
    super.key,
  });

  /// The line.
  final BreakdownLine line;

  /// Whether the regime is incomplete and this figure is not a plain sum
  /// of the records (those do not depend on the missing inputs).
  final bool dependsOnMissingInputs;

  @override
  Widget build(BuildContext context) => ComputationLineTile(
    line: ComputationLine(
      concept: line.concept,
      amountCents: line.amountCents,
      role: switch (line.operation) {
        BreakdownLineOperation.subtract => ComputationLineRole.subtraction,
        BreakdownLineOperation.result => ComputationLineRole.result,
        BreakdownLineOperation.roundingAdjustment =>
          ComputationLineRole.rounding,
        _ => ComputationLineRole.step,
      },
      rate: line.rate,
      legalBasis: line.legalBasis,
      sourceLabel: _sourceLabel(AppLocalizations.of(context)),
      needsCpaReview: line.needsCpaReview,
    ),
    dependsOnMissingInputs: dependsOnMissingInputs,
  );

  String? _sourceLabel(AppLocalizations l10n) {
    final source = line.source;
    if (source == _records) return l10n.breakdownSourceRecords;
    if (source == _computed) return l10n.breakdownSourceComputed;
    if (source.startsWith(_taxInputPrefix)) return l10n.breakdownSourceInput;
    if (source.startsWith(_paramsPrefix)) return l10n.breakdownSourceParams;
    return null;
  }
}

/// Whether [line] sums the records (independent of the tax inputs).
bool isRecordsLine(BreakdownLine line) => line.source == _records;
