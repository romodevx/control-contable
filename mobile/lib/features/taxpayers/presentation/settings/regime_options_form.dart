import 'dart:async';

import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/settings_feedback.dart';
import 'package:control_contable/features/taxpayers/presentation/warning_block.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Prefix of an option's field path in a `400` (`options.<key>`).
const String _optionsPath = 'options';

/// The options of [regime] (the same fields as onboarding step 4), from
/// the [current] values. Warns first that a change applies to every period
/// of the regime (OQ-7: options are not versioned). Saving sends them all
/// (`PATCH`, merge semantics) and returns to the settings.
class RegimeOptionsForm extends ConsumerStatefulWidget {
  /// The form for [regime] of [taxpayerId].
  const new({
    required this.taxpayerId,
    required this.regime,
    required this.current,
    super.key,
  });

  /// The taxpayer.
  final String taxpayerId;

  /// The catalog regime (its `option_defs`).
  final Regime regime;

  /// The options the taxpayer holds.
  final Map<String, dynamic> current;

  @override
  ConsumerState<RegimeOptionsForm> createState() => _RegimeOptionsFormState();
}

class _RegimeOptionsFormState extends ConsumerState<RegimeOptionsForm> {
  final AppFormController _form = AppFormController();
  List<OptionField>? _optionFields;
  String? _problem;

  List<OptionField> get _fields => _optionFields!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final l10n = AppLocalizations.of(context);
    _optionFields ??= [
      for (final def in widget.regime.optionDefs)
        ?OptionField.create(def, widget.current[def.key], _form, l10n),
    ];
  }

  Future<void> _save() async {
    final l10n = AppLocalizations.of(context);
    final messenger = ScaffoldMessenger.of(context);
    setState(() => _problem = null);
    try {
      await ref
          .read(taxpayerSettingsProvider)
          .run(
            WriteAccess.needsPlan,
            (repository) => repository.updateRegimeOptions(
              taxpayerId: widget.taxpayerId,
              regimeCode: widget.regime.code,
              options: {
                for (final option in _fields)
                  if (option.value case final Object value) option.key: value,
              },
            ),
          );
      showSettingsToast(messenger, l10n.optionsSaved);
      if (mounted) context.pop();
    } on ApiEntitlementRequired {
      if (mounted) unawaited(showPaywall(context));
    } on ApiValidationFailure catch (failure) {
      _showRejected(failure, l10n);
    } on ApiFailure catch (failure) {
      if (mounted) {
        setState(() => _problem = settingsFailureMessage(failure, l10n));
      }
    }
  }

  void _showRejected(ApiValidationFailure failure, AppLocalizations l10n) {
    final rejected = {
      for (final detail in failure.details)
        if (detail.field.split('.') case [_optionsPath, final key, ...]) key,
    };
    _form.showServerErrors({
      for (final option in _fields)
        if (rejected.contains(option.key))
          option.field: CustomFieldError(l10n.onboardingOptionRejected),
    });
  }

  @override
  void dispose() {
    _form.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AppForm(
      controller: _form,
      onSave: _save,
      onCancel: () => context.pop(),
      children: [
        Semantics(
          header: true,
          child: Text(
            regimeTitle(widget.regime),
            style: Theme.of(context).textTheme.titleLarge,
          ),
        ),
        WarningBlock(
          title: l10n.onboardingOptionsTitle,
          body: l10n.optionsEditWarning,
        ),
        if (_problem case final problem?) FormNotice(message: problem),
        for (final option in _fields) option.build(l10n),
      ],
    );
  }
}
