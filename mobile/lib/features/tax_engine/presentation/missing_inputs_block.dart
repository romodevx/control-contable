import 'dart:async';

import 'package:control_contable/api/generated/models/missing_input.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// What an `incomplete` regime is missing ("Falta: …", each input with its
/// months) and the CTA to the tax-inputs form at [location].
class MissingInputsBlock extends StatelessWidget {
  /// The [missing] inputs, completed at [location].
  const new({required this.missing, required this.location, super.key});

  /// The inputs the server reported missing, in its order.
  final List<MissingInput> missing;

  /// The tax-inputs form, opened at the first missing key.
  final String location;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final months = MaterialLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(l10n.estimateMissingTitle, style: text.titleSmall),
        ),
        for (final input in missing)
          Padding(
            padding: const EdgeInsets.only(top: AppSpacing.xxs),
            child: Text(
              [
                input.label,
                if (input.months case final periods? when periods.isNotEmpty)
                  periods
                      .map((p) => months.formatMonthYear(periodMonth(p)))
                      .join(', '),
              ].join(': '),
              style: text.bodyMedium,
            ),
          ),
        const SizedBox(height: AppSpacing.sm),
        OutlinedButton(
          onPressed: () => unawaited(context.push<void>(location)),
          child: Text(l10n.estimateCompleteInputs),
        ),
      ],
    );
  }
}
