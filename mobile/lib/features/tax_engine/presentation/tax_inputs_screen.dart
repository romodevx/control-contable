import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/tax_engine/application/tax_inputs.dart';
import 'package:control_contable/features/tax_engine/presentation/tax_inputs_form_page.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

const String _regimeNotActive = 'regime_not_active';
const String _parametersUnavailable = 'parameters_unavailable';

/// The tax inputs of one regime and fiscal year (M5-27), the target of the
/// estimate's "Completa tus datos": loads them, then shows the form opened
/// at [focusKey].
class TaxInputsScreen extends ConsumerWidget {
  /// The inputs of [regime] in [fiscalYear] for [taxpayerId].
  const new({
    required this.taxpayerId,
    required this.regime,
    this.fiscalYear,
    this.focusKey,
    super.key,
  });

  /// Whose inputs.
  final String taxpayerId;

  /// The regime code.
  final String regime;

  /// The fiscal year; the current one (America/Mexico_City) when the link
  /// carries none.
  final int? fiscalYear;

  /// The input to open the form at.
  final String? focusKey;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final query = (
      taxpayerId: taxpayerId,
      regime: regime,
      fiscalYear:
          fiscalYear ??
          periodMonth(currentPeriod(ref.read(clockProvider)())).year,
    );
    final inputs = ref.watch(taxInputsProvider(query));
    if (inputs.value case final loaded?) {
      return TaxInputsFormPage(
        key: ValueKey(query),
        query: query,
        inputs: loaded,
        focusKey: focusKey,
      );
    }
    return Scaffold(
      appBar: AppBar(title: Text(l10n.taxInputsTitle)),
      body: ListView(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        children: [
          if (inputs.error case final error?)
            LoadFailed(
              message: _failureMessage(error, l10n),
              onRetry: () => ref.invalidate(taxInputsProvider(query)),
            )
          else
            LoadStatus(message: l10n.taxInputsLoading, busy: true),
        ],
      ),
    );
  }

  String _failureMessage(Object error, AppLocalizations l10n) =>
      switch (error) {
        ApiUnreachable() => l10n.estimateUnreachable,
        ApiNotFound() => l10n.estimateNotFound,
        ApiRejected(code: ErrorCode.businessRuleViolation, :final details)
            when details.any((d) => d.code == _regimeNotActive) =>
          l10n.taxInputsRegimeNotActive,
        ApiRejected(code: ErrorCode.businessRuleViolation, :final details)
            when details.any((d) => d.code == _parametersUnavailable) =>
          l10n.estimateParametersUnavailable,
        _ => l10n.taxInputsLoadFailed,
      };
}
