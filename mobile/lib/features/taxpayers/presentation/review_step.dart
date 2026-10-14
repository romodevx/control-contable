import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/domain/masked_rfc.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/features/taxpayers/presentation/warning_block.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// After the create, when the server warned (non-blocking): the taxpayer
/// exists. A check-digit mismatch (D-RFC) shows "Revisa que tu RFC esté
/// bien escrito" over the RFC, editable: *Continuar* as typed confirms it;
/// a changed RFC is corrected first (`PATCH`). A regime pair under CPA
/// review shows the server's reason.
class ReviewStep extends StatefulWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The onboarding.
  final OnboardingFlow flow;

  @override
  State<ReviewStep> createState() => _ReviewStepState();
}

class _ReviewStepState extends State<ReviewStep> {
  final AppFormController _form = AppFormController();
  AppFieldController? _rfcField;
  RfcRejection? _shownRejection;

  OnboardingFlow get _flow => widget.flow;
  PersonType get _personType => _flow.personType ?? PersonType.fisica;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_rfcField != null) return;
    final l10n = AppLocalizations.of(context);
    _rfcField = _form.textField(
      label: l10n.onboardingRfcLabel,
      initialText: normalizeRfc(_flow.rfc),
      validator: (value) => checkTypedRfc(value, _personType, l10n),
      canonicalize: normalizeRfc,
    );
    _flow.addListener(_showServerVerdict);
  }

  /// Shows a refused correction on the RFC (once per verdict).
  void _showServerVerdict() {
    final rejection = _flow.rfcRejection;
    if (rejection == _shownRejection) return;
    _shownRejection = rejection;
    if (rejection == null) return;
    final l10n = AppLocalizations.of(context);
    _form.showServerErrors({
      _rfcField!: describeRfcRejection(rejection, _personType, l10n),
    });
  }

  Future<void> _save() =>
      _flow.confirmReview(rfc: _flow.rfcWarning ? _rfcField!.value : _flow.rfc);

  @override
  void dispose() {
    _flow.removeListener(_showServerVerdict);
    _form.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: _flow,
      builder: (context, _) {
        final created = _flow.created;
        final notice = _flow.notice;
        return AppForm(
          controller: _form,
          onSave: _save,
          primaryLabel: l10n.onboardingContinue,
          children: [
            StepHeader(flow: _flow, title: l10n.onboardingReviewTitle),
            if (created != null)
              FormNotice(
                tone: FormNoticeTone.info,
                message: l10n.onboardingReviewCreated(
                  created.displayName,
                  displayRfc(created.rfcMasked),
                ),
              ),
            if (notice != null)
              FormNotice(message: describeNotice(notice, l10n)),
            if (_flow.rfcWarning) ...[
              WarningBlock(
                title: l10n.onboardingRfcCheckWarning,
                body: l10n.onboardingRfcCheckBody,
              ),
              AppTextField(
                field: _rfcField!,
                keyboardType: TextInputType.text,
                textInputAction: TextInputAction.done,
                autofillHints: const [],
                textCapitalization: TextCapitalization.characters,
                autocorrect: false,
              ),
            ],
            for (final reason in _flow.cpaReviewReasons)
              WarningBlock(title: l10n.onboardingCpaWarningTitle, body: reason),
          ],
        );
      },
    );
  }
}
