import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// The screens that complete an annual input.
enum AnnualInput {
  /// Personal deductions (ANN-19).
  personalDeductions,

  /// Salary constancias (ANN-20).
  salaryStatements,

  /// Per-year annual facts (ANN-21).
  annualInputs,

  /// Provisional payments filed (ANN-21).
  filedPayments,
}

/// What [input]'s screen is called.
String annualInputTitle(AppLocalizations l10n, AnnualInput input) =>
    switch (input) {
      AnnualInput.personalDeductions => l10n.annualInputDeductions,
      AnnualInput.salaryStatements => l10n.annualInputSalary,
      AnnualInput.annualInputs => l10n.annualInputFacts,
      AnnualInput.filedPayments => l10n.annualInputPayments,
    };

/// Where an annual input will be captured, until its screen ships
/// (ANN-19..21): its title and a plain line saying it is coming.
class AnnualInputPendingScreen extends StatelessWidget {
  /// The placeholder of [input].
  const new({required this.input, super.key});

  /// The input this location completes.
  final AnnualInput input;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(annualInputTitle(l10n, input))),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: Text(
          l10n.annualInputPending,
          style: Theme.of(context).textTheme.bodyLarge,
        ),
      ),
    );
  }
}
