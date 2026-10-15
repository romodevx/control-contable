import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Step 2: the display name and the RFC (two fields). The RFC is accepted
/// with spaces, dashes and lower case and rewritten upper-case on blur;
/// its length follows the person type chosen in step 1. Format errors show
/// on blur, empty fields on submit; the server's verdict (sent back here
/// from the last step) shows on its field.
class IdentityStep extends StatefulWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The onboarding.
  final OnboardingFlow flow;

  @override
  State<IdentityStep> createState() => _IdentityStepState();
}

class _IdentityStepState extends State<IdentityStep> {
  final AppFormController _form = AppFormController();
  AppFieldController? _nameField;
  late final AppFieldController _rfc;

  OnboardingFlow get _flow => widget.flow;
  PersonType get _personType => _flow.personType ?? PersonType.fisica;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_nameField != null) return;
    // The labels and validators need the localized copy (not in initState).
    final l10n = AppLocalizations.of(context);
    final name = _nameField = _form.textField(
      label: l10n.onboardingDisplayNameLabel,
      initialText: _flow.displayName,
      validator: (value) => checkDisplayName(value, l10n),
    );
    final rfc = _rfc = _form.textField(
      label: l10n.onboardingRfcLabel,
      initialText: _flow.rfc,
      validator: (value) => checkTypedRfc(value, _personType, l10n),
      canonicalize: normalizeRfc,
    );
    for (final field in [name, rfc]) {
      field.text.addListener(_remember);
    }
    _showServerVerdict(l10n);
  }

  void _remember() => _flow.rememberIdentity(
    displayName: _nameField!.text.text,
    rfc: _rfc.text.text,
  );

  void _showServerVerdict(AppLocalizations l10n) {
    final rejection = _flow.rfcRejection;
    final nameRejection = _flow.displayNameRejection;
    _form.showServerErrors({
      if (rejection != null)
        _rfc: describeRfcRejection(rejection, _personType, l10n),
      if (nameRejection != null)
        _nameField!: describeDisplayNameRejection(nameRejection, l10n),
    });
  }

  Future<void> _save() async =>
      _flow.submitIdentity(displayName: _nameField!.value, rfc: _rfc.value);

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
      primaryLabel: l10n.onboardingNext,
      children: [
        StepHeader(flow: _flow, title: l10n.onboardingIdentityTitle),
        AppTextField(
          field: _nameField!,
          keyboardType: TextInputType.name,
          textInputAction: TextInputAction.next,
          autofillHints: const [AutofillHints.name],
          helperText: l10n.onboardingDisplayNameHelp,
        ),
        AppTextField(
          field: _rfc,
          keyboardType: TextInputType.text,
          textInputAction: TextInputAction.done,
          autofillHints: const [],
          textCapitalization: TextCapitalization.characters,
          autocorrect: false,
          helperText: _personType == PersonType.moral
              ? l10n.onboardingRfcHelpMoral
              : l10n.onboardingRfcHelpFisica,
        ),
      ],
    );
  }
}
