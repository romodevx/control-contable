import 'package:control_contable/api/generated/models/annual_missing_input.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/annual/application/annual_input_location.dart';
import 'package:control_contable/features/annual/presentation/annual_link_row.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// What an `incomplete` section is missing ("Falta:"), each input a row
/// that opens the screen completing it, with the months concerned. An
/// input with no screen in this app version (ANN-19..21 not shipped, or an
/// unknown target) is still named with its months, but is not a link.
class AnnualMissingInputs extends StatelessWidget {
  /// The [missing] inputs of [taxpayerId]'s [fiscalYear].
  const new({
    required this.missing,
    required this.taxpayerId,
    required this.fiscalYear,
    super.key,
  });

  /// The inputs the server reported missing, in its order.
  final List<AnnualMissingInput> missing;

  /// Whose estimate it is.
  final String taxpayerId;

  /// The fiscal year of the estimate.
  final int fiscalYear;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    final months = MaterialLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(
            AppLocalizations.of(context).estimateMissingTitle,
            style: text.titleSmall,
          ),
        ),
        for (final input in missing)
          _row(input, _monthsOf(input, months), text),
      ],
    );
  }

  Widget _row(AnnualMissingInput input, String? months, TextTheme text) {
    final location = annualInputLocation(
      input,
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
    );
    if (location == null) {
      return _NamedInput(title: input.label, months: months);
    }
    return AnnualLinkRow(
      key: ValueKey('missing-${input.key}'),
      title: input.label,
      subtitle: months,
      location: location,
    );
  }

  static String? _monthsOf(
    AnnualMissingInput input,
    MaterialLocalizations months,
  ) => input.months.isEmpty
      ? null
      : input.months
            .map((period) => months.formatMonthYear(periodMonth(period)))
            .join(', ');
}

/// A missing input named without a link: its label and, under it, the
/// months concerned. Read as one sentence, never as a button.
class _NamedInput extends StatelessWidget {
  const new({required this.title, required this.months});

  final String title;
  final String? months;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    final detail = months;
    return MergeSemantics(
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(title, style: text.bodyLarge),
            if (detail != null)
              Text(
                detail,
                style: text.bodyMedium?.copyWith(
                  color: context.tokens.colors.muted,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
