import 'package:control_contable/api/generated/models/breakdown_line.dart';
import 'package:control_contable/api/generated/models/breakdown_line_operation.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/domain/rate_percent.dart';
import 'package:control_contable/features/tax_engine/presentation/citation_disclosure.dart';
import 'package:control_contable/features/tax_engine/presentation/money_row.dart';
import 'package:control_contable/features/tax_engine/presentation/status_chip.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// `source` prefixes of the contract (`BreakdownLine.source`).
const String _records = 'records';
const String _computed = 'computed';
const String _taxInputPrefix = 'tax_input:';
const String _paramsPrefix = 'params:';

const double _cpaIconSize = 18;

/// One step of the computation: the concept and its amount (a subtraction
/// shows and reads as one), the rate of a multiplication, where the figure
/// comes from, the ⚠️ of a CPA-unverified rule, and the legal basis. The
/// rounding adjustment is muted and explained instead of cited; a result
/// is ruled off from the steps above it.
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

  bool get _isRounding =>
      line.operation == BreakdownLineOperation.roundingAdjustment;

  bool get _isResult => line.operation == BreakdownLineOperation.result;

  int get _shownCentavos => line.operation == BreakdownLineOperation.subtract
      ? -line.amountCents
      : line.amountCents;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.tokens.colors;
    final text = Theme.of(context).textTheme;
    final muted = colors.muted;
    final rate = line.rate;
    final source = _sourceLabel(l10n);
    return Semantics(
      container: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: _isResult
              ? Border(top: BorderSide(color: colors.border))
              : null,
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            spacing: AppSpacing.xxs,
            children: [
              MoneyRow(
                label: line.concept,
                centavos: _shownCentavos,
                color: _isRounding ? muted : null,
                labelStyle: _isResult ? text.titleSmall : text.bodyLarge,
                isIncomplete: dependsOnMissingInputs,
              ),
              if (rate != null)
                Text(
                  l10n.breakdownRate(formatRatePercent(rate)),
                  style: text.bodyMedium?.copyWith(color: muted),
                ),
              if (_isRounding)
                Text(
                  l10n.breakdownRoundingHelp,
                  style: text.bodyMedium?.copyWith(color: muted),
                ),
              if (source != null || line.needsCpaReview)
                Wrap(
                  spacing: AppSpacing.xs,
                  runSpacing: AppSpacing.xxs,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    if (source != null) StatusChip(label: source, color: muted),
                    if (line.needsCpaReview) const _CpaReviewMark(),
                  ],
                ),
              if (line.legalBasis.isNotEmpty)
                CitationDisclosure(citations: line.legalBasis),
            ],
          ),
        ),
      ),
    );
  }

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

/// ⚠️ "Sin validar por un contador": the icon never stands alone.
class _CpaReviewMark extends StatelessWidget {
  const new();

  @override
  Widget build(BuildContext context) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Icon(
        Icons.warning_amber_rounded,
        size: _cpaIconSize,
        color: context.tokens.colors.legendText,
      ),
      const SizedBox(width: AppSpacing.xxs),
      Flexible(
        child: Text(
          AppLocalizations.of(context).breakdownCpaLine,
          style: Theme.of(context).textTheme.labelLarge,
        ),
      ),
    ],
  );
}
