import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/features/tax_engine/application/prior_payment_suggestion.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Under an empty month: the app's own estimate of that month's payment,
/// labelled as such, and "Usar este monto" to copy it into [field]
/// (MODULE-tax-engine §8, OQ-T1). Only a suggestion: nothing is filled or
/// saved until the user uses it, and it goes away once the month has a
/// value. Nothing shows while it loads, when it fails, or when there is no
/// trustworthy figure.
class PriorPaymentSuggestionLine extends ConsumerWidget {
  /// The suggestion of [query] for [field].
  const new({required this.field, required this.query, super.key});

  /// The month's amount.
  final MoneyFieldController field;

  /// Whose month.
  final PriorPaymentQuery query;

  @override
  Widget build(BuildContext context, WidgetRef ref) => ListenableBuilder(
    listenable: field.text,
    builder: (context, _) {
      if (field.value.isNotEmpty) return const SizedBox.shrink();
      final cents = ref.watch(priorPaymentSuggestionProvider(query)).value;
      if (cents == null) return const SizedBox.shrink();
      final l10n = AppLocalizations.of(context);
      final text = Theme.of(context).textTheme;
      return Padding(
        padding: const EdgeInsets.only(top: AppSpacing.xxs),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.taxInputsSuggestion(formatPesos(cents)),
              style: text.bodyMedium?.copyWith(
                color: context.tokens.colors.muted,
              ),
            ),
            TextButton(
              style: TextButton.styleFrom(
                minimumSize: const Size.square(AppSizes.minTarget),
              ),
              onPressed: () => field.text.text = formatCentavos(cents),
              child: Text(l10n.taxInputsUseSuggestion),
            ),
          ],
        ),
      );
    },
  );
}
