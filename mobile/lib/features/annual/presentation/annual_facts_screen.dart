import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/presentation/annual_facts_form.dart';
import 'package:control_contable/features/annual/presentation/annual_input_list.dart';
import 'package:control_contable/features/annual/presentation/regime_tax_inputs_list.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// "Datos del año" (ANN-21): a persona física's per-year facts as a form,
/// and every taxpayer's regime tax inputs (losses, PTU, depreciation)
/// read-only with a link to the form that owns each.
class AnnualFactsScreen extends ConsumerWidget {
  /// The facts of [taxpayerId] in [fiscalYear].
  const new({required this.taxpayerId, required this.fiscalYear, super.key});

  /// Whose facts.
  final String taxpayerId;

  /// Which year.
  final int fiscalYear;

  AnnualQuery get _query => (taxpayerId: taxpayerId, fiscalYear: fiscalYear);

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final active = ref.watch(activeTaxpayerProvider).value;
    final isMoral =
        active?.id == taxpayerId && active?.personType == PersonType.moral;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.annualInputFacts)),
      body: annualInputBody(
        ref.watch(annualFactsProvider(_query)),
        loading: l10n.annualFactsLoading,
        failed: l10n.annualFactsLoadFailed,
        onRetry: () => ref.invalidate(annualFactsProvider(_query)),
        data: (facts) => isMoral
            ? _MoralFacts(query: _query, facts: facts)
            : AnnualFactsForm(query: _query, facts: facts),
      ),
    );
  }
}

/// A persona moral has no per-year facts here (`person_type_not_allowed`):
/// only its regimes' inputs.
class _MoralFacts extends StatelessWidget {
  const new({required this.query, required this.facts});

  final AnnualQuery query;
  final AnnualInputs facts;

  @override
  Widget build(BuildContext context) => AnnualInputList(
    children: [
      Text(
        AppLocalizations.of(context).annualFactsMoralNote,
        style: Theme.of(context).textTheme.bodyLarge,
      ),
      RegimeTaxInputsList(query: query, inputs: facts.regimeTaxInputs),
    ],
  );
}
