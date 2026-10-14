import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/inline_text_editor.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/settings_feedback.dart';
import 'package:control_contable/features/taxpayers/presentation/taxpayer_tile.dart';
import 'package:control_contable/features/taxpayers/presentation/warning_block.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The taxpayer's name and RFC, each edited in place. The RFC is never
/// shown or prefilled in full (S-3): correcting it means typing it again.
/// A check-digit mismatch is saved and asks the user to double-check
/// (D-RFC). Archived taxpayers are read-only here (writes answer `422`).
class DataSection extends ConsumerStatefulWidget {
  /// The data of [taxpayer].
  const new({required this.taxpayer, super.key});

  /// The taxpayer.
  final Taxpayer taxpayer;

  @override
  ConsumerState<DataSection> createState() => _DataSectionState();
}

class _DataSectionState extends ConsumerState<DataSection> {
  bool _rfcWarning = false;

  Taxpayer get _taxpayer => widget.taxpayer;
  bool get _editable => _taxpayer.archivedAt == null;

  /// Whether a write may start; a known read-only user gets the paywall.
  bool _canWrite() {
    if (!ref.read(taxpayerSettingsProvider).knownReadOnly) return true;
    unawaited(showPaywall(context));
    return false;
  }

  Future<FieldError?> _rename(String name) => _write(
    (repository) =>
        repository.rename(taxpayerId: _taxpayer.id, displayName: name),
    (_) => AppLocalizations.of(context).settingsNameSaved,
  );

  Future<FieldError?> _correctRfc(String typed) {
    setState(() => _rfcWarning = false);
    return _write(
      (repository) => repository.correctRfc(
        taxpayerId: _taxpayer.id,
        rfc: normalizeRfc(typed),
      ),
      (updated) {
        final warned = (updated.warnings ?? const []).any(
          (w) => w.code == rfcCheckDigitWarning,
        );
        setState(() => _rfcWarning = warned);
        return warned
            ? null
            : AppLocalizations.of(context).settingsRfcCorrected;
      },
      fieldError: (failure) => switch (rfcRejectionOf(failure)) {
        final rejection? => describeRfcRejection(
          rejection,
          _taxpayer.personType,
          AppLocalizations.of(context),
        ),
        null => null,
      },
    );
  }

  /// Runs [write]; a success shows [confirmation] (if any) and closes the
  /// editor; a failure a field owns stays on the field.
  Future<FieldError?> _write(
    Future<Taxpayer> Function(TaxpayersRepository) write,
    String? Function(Taxpayer) confirmation, {
    FieldError? Function(ApiFailure)? fieldError,
  }) async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    try {
      final updated = await ref
          .read(taxpayerSettingsProvider)
          .run(WriteAccess.needsPlan, write);
      if (!mounted) return null;
      if (confirmation(updated) case final message?) {
        showSettingsToast(messenger, message);
      }
      return null;
    } on ApiEntitlementRequired {
      if (mounted) unawaited(showPaywall(context));
      return null;
    } on ApiFailure catch (failure) {
      return fieldError?.call(failure) ??
          CustomFieldError(settingsFailureMessage(failure, l10n));
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final theme = Theme.of(context);
    final name = Text(
      _taxpayer.displayName,
      style: theme.textTheme.titleMedium,
    );
    final rfc = TaxpayerSummary(taxpayer: _taxpayer);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (!_editable) ...[
          name,
          rfc,
        ] else ...[
          InlineTextEditor(
            label: l10n.onboardingDisplayNameLabel,
            display: name,
            editLabel: l10n.settingsEditName,
            initialText: _taxpayer.displayName,
            validator: (value) => checkDisplayName(value, l10n),
            canOpen: _canWrite,
            onSave: _rename,
            fieldBuilder: (field) => AppTextField(
              field: field,
              keyboardType: TextInputType.name,
              textInputAction: TextInputAction.done,
              autofillHints: const [AutofillHints.name],
            ),
          ),
          InlineTextEditor(
            label: l10n.onboardingRfcLabel,
            display: rfc,
            editLabel: l10n.settingsCorrectRfc,
            validator: (value) =>
                checkTypedRfc(value, _taxpayer.personType, l10n),
            canonicalize: normalizeRfc,
            canOpen: _canWrite,
            onSave: _correctRfc,
            fieldBuilder: (field) => AppTextField(
              field: field,
              keyboardType: TextInputType.text,
              textInputAction: TextInputAction.done,
              autofillHints: const [],
              textCapitalization: TextCapitalization.characters,
              autocorrect: false,
              helperText: l10n.settingsRfcHelp,
            ),
          ),
        ],
        if (_rfcWarning)
          WarningBlock(
            title: l10n.onboardingRfcCheckWarning,
            body: l10n.onboardingRfcCheckBody,
          ),
      ],
    );
  }
}
