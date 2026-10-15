import 'package:control_contable/api/generated/models/annual_missing_input.dart';
import 'package:control_contable/features/annual/application/annual_input_location.dart';
import 'package:control_contable/features/annual/presentation/annual_link_row.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// What an `incomplete` section is missing ("Falta:"), each input a row
/// that opens the screen completing it, with the months concerned. An
/// input this app cannot route to is still named, without a link.
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
      return Text([input.label, ?months].join(': '), style: text.bodyMedium);
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
