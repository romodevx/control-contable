import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/domain/annual_payment_regimes.dart';
import 'package:control_contable/features/annual/presentation/annual_link_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// "Datos para tu anual" (MODULE-annual §8): the screens that feed the
/// estimate. Personas físicas: personal deductions and salary constancias
/// with how many are registered (left out when the list cannot be read);
/// everyone: the provisional payments filed per regime and the year's
/// facts. A known read-only user is told what still works.
class AnnualInputsBlock extends ConsumerWidget {
  /// The block of [estimate].
  const new({required this.estimate, super.key});

  /// The estimate shown.
  final AnnualEstimate estimate;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final text = Theme.of(context).textTheme;
    final taxpayerId = estimate.taxpayerId;
    final year = estimate.fiscalYear;
    final query = (taxpayerId: taxpayerId, fiscalYear: year);
    final isFisica = estimate.personType == PersonType.fisica;
    String? count(int? registered) =>
        registered == null ? null : l10n.annualCountRegistered(registered);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Semantics(
          header: true,
          child: Text(l10n.annualInputsTitle, style: text.titleMedium),
        ),
        if (ref.watch(readOnlyAccessProvider))
          Padding(
            key: const ValueKey('annual-read-only'),
            padding: const EdgeInsets.only(top: AppSpacing.xs),
            child: Text(l10n.annualReadOnly, style: text.bodyMedium),
          ),
        if (isFisica) ...[
          AnnualLinkRow(
            title: l10n.annualInputDeductions,
            subtitle: count(
              ref.watch(personalDeductionsProvider(query)).value?.items.length,
            ),
            location: AnnualRoutes.personalDeductions(taxpayerId, year),
          ),
          AnnualLinkRow(
            title: l10n.annualInputSalary,
            subtitle: count(
              ref.watch(salaryStatementsProvider(query)).value?.items.length,
            ),
            location: AnnualRoutes.salaryStatements(taxpayerId, year),
          ),
        ],
        for (final regime in provisionalPaymentRegimes(estimate))
          AnnualLinkRow(
            title: l10n.annualInputPayments,
            subtitle: l10n.estimateRegimeCode(regime),
            location: AnnualRoutes.filedPayments(taxpayerId, year, regime),
          ),
        AnnualLinkRow(
          title: l10n.annualInputFacts,
          location: AnnualRoutes.annualInputs(taxpayerId, year),
        ),
      ],
    );
  }
}
