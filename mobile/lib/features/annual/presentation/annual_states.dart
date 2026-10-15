import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

const String _parametersUnavailable = 'parameters_unavailable';
const String _sectionNotPresent = 'section_not_present';
const String _outOfRange = 'out_of_range';

/// A failed annual estimate or breakdown load: why, in es-MX, and a retry.
class AnnualFailure extends StatelessWidget {
  /// The failure of [error], retried by [onRetry].
  const new({required this.error, required this.onRetry, super.key});

  /// What the load threw.
  final Object error;

  /// Loads again.
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => LoadFailed(
    message: _message(AppLocalizations.of(context)),
    onRetry: onRetry,
  );

  String _message(AppLocalizations l10n) => switch (error) {
    ApiUnreachable() => l10n.estimateUnreachable,
    ApiNotFound() => l10n.estimateNotFound,
    ApiRejected(code: ErrorCode.businessRuleViolation, :final details)
        when details.any((d) => d.code == _parametersUnavailable) =>
      l10n.annualParametersUnavailable,
    ApiRejected(code: ErrorCode.businessRuleViolation, :final details)
        when details.any((d) => d.code == _sectionNotPresent) =>
      l10n.annualSectionNotPresent,
    ApiValidationFailure(:final details)
        when details.any((d) => d.code == _outOfRange) =>
      l10n.annualFutureYear,
    _ => l10n.annualLoadFailed,
  };
}
