import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/forms/form_footer.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/core/widgets/captioned_picker.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/settings_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// `details[].code` when records exist after the chosen month.
const String _recordsAfterEnd = 'records_after_end_period';

/// Ends [regime] at a chosen month, in place (one field: no dialog). The
/// month defaults to the current one and is picked from a list, so it is
/// never out of range. A regime with records asks once more and keeps them
/// (`409 regime_has_records` → `confirm=true`). Allowed while read-only.
class EndRegimePanel extends ConsumerStatefulWidget {
  /// The panel ending [regime] of [taxpayerId]; [onClose] folds it.
  const new({
    required this.taxpayerId,
    required this.regime,
    required this.onClose,
    super.key,
  });

  /// The taxpayer.
  final String taxpayerId;

  /// The regime to end.
  final TaxpayerRegime regime;

  /// Folds the panel (cancel, or after ending).
  final VoidCallback onClose;

  @override
  ConsumerState<EndRegimePanel> createState() => _EndRegimePanelState();
}

class _EndRegimePanelState extends ConsumerState<EndRegimePanel> {
  late final List<String> _choices = endPeriodChoices(
    start: widget.regime.startPeriod,
    current: currentPeriod(ref.read(clockProvider)()),
  );
  late String _endPeriod = _choices.first;
  bool _keepRecords = false;
  bool _busy = false;
  String? _problem;

  Future<void> _end() async {
    if (_busy) return;
    setState(() => _busy = true);
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    final code = widget.regime.regimeCode;
    try {
      await ref
          .read(taxpayerSettingsProvider)
          .run(
            WriteAccess.always,
            (repository) => repository.endRegime(
              taxpayerId: widget.taxpayerId,
              regimeCode: code,
              endPeriod: _endPeriod,
              confirm: _keepRecords,
            ),
          );
      showSettingsToast(messenger, l10n.settingsRegimeEndedDone(code));
      if (mounted) widget.onClose();
    } on ApiFailure catch (failure) {
      if (!mounted) return;
      setState(() {
        _busy = false;
        _keepRecords |=
            failure is ApiRejected &&
            failure.code == ErrorCode.regimeHasRecords;
        _problem = _describe(failure, l10n);
      });
    }
  }

  String _describe(ApiFailure failure, AppLocalizations l10n) =>
      switch (failure) {
        ApiRejected(code: ErrorCode.regimeHasRecords) =>
          l10n.settingsEndRegimeHasRecords,
        ApiRejected(:final details)
            when details.any((d) => d.code == _recordsAfterEnd) =>
          l10n.settingsEndRegimeRecordsAfter,
        _ => settingsFailureMessage(failure, l10n),
      };

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final months = MaterialLocalizations.of(context);
    final label = _keepRecords
        ? l10n.settingsEndRegimeKeep
        : l10n.settingsEndRegimeConfirm;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        CaptionedPicker(
          caption: l10n.settingsEndRegimeLabel,
          picker: DropdownButtonFormField<String>(
            initialValue: _endPeriod,
            isExpanded: true,
            items: [
              for (final period in _choices)
                DropdownMenuItem(
                  value: period,
                  child: Text(months.formatMonthYear(periodMonth(period))),
                ),
            ],
            onChanged: (period) => setState(() => _endPeriod = period!),
          ),
        ),
        if (_problem case final problem?) ...[
          const SizedBox(height: AppSpacing.xs),
          FormNotice(message: problem),
        ],
        const SizedBox(height: AppSpacing.xs),
        Wrap(
          spacing: AppSpacing.xs,
          children: [
            FilledButton(
              onPressed: _end,
              child: _busy
                  ? BusyLabel(label: label, announcement: l10n.formSaving)
                  : Text(label),
            ),
            TextButton(onPressed: widget.onClose, child: Text(l10n.formCancel)),
          ],
        ),
      ],
    );
  }
}
