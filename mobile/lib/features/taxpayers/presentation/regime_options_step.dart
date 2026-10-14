import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Step 4: the options of one chosen regime per page, rendered from its
/// `option_defs` (boolean → switch, enum → radio group, multi_enum →
/// checkboxes, rate → es-MX decimal field), prefilled with the catalog
/// defaults. The last page creates the taxpayer.
class RegimeOptionsStep extends StatefulWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The onboarding.
  final OnboardingFlow flow;

  @override
  State<RegimeOptionsStep> createState() => _RegimeOptionsStepState();
}

class _RegimeOptionsStepState extends State<RegimeOptionsStep> {
  final AppFormController _form = AppFormController();
  List<OptionField>? _optionFields;
  Map<String, Set<String>>? _shownRejections;

  // Fixed when the page opens: the flow moves on while this page is still
  // animating out, and a late notification must not write another regime.
  late final int _page = _flow.page;
  late final List<Regime> _regimes = _flow.optionRegimes;
  late final Regime _regime = _regimes[_page];

  OnboardingFlow get _flow => widget.flow;
  List<OptionField> get _fields => _optionFields!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_optionFields != null) return;
    final l10n = AppLocalizations.of(context);
    final initial = _flow.optionsFor(_regime);
    _optionFields = [
      for (final def in _regime.optionDefs)
        ?OptionField.create(def, initial[def.key], _form, l10n),
    ];
    _form.addListener(_remember);
    _flow.addListener(_showServerVerdict);
    _showServerVerdict();
  }

  Map<String, Object> _values() => {
    for (final option in _fields)
      if (option.value case final Object value) option.key: value,
  };

  void _remember() => _flow.rememberOptions(_regime.code, _values());

  /// Marks the options the server refused (once per verdict).
  void _showServerVerdict() {
    final rejections = _flow.optionRejections;
    if (identical(rejections, _shownRejections)) return;
    _shownRejections = rejections;
    final rejected = rejections[_regime.code] ?? const {};
    final message = AppLocalizations.of(context).onboardingOptionRejected;
    _form.showServerErrors({
      for (final option in _fields)
        if (rejected.contains(option.key))
          option.field: CustomFieldError(message),
    });
  }

  @override
  void dispose() {
    _flow.removeListener(_showServerVerdict);
    _form.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isLast = _page == _regimes.length - 1;
    return ListenableBuilder(
      listenable: _flow,
      builder: (context, _) {
        final notice = _flow.notice;
        return AppForm(
          controller: _form,
          onSave: () => _flow.submitOptions(_regime.code, _values()),
          primaryLabel: isLast ? l10n.onboardingCreate : l10n.onboardingNext,
          children: [
            StepHeader(
              flow: _flow,
              title: l10n.onboardingOptionsTitle,
              intro: l10n.onboardingOptionsRegime(
                _page + 1,
                _regimes.length,
                _regime.code,
                _regime.name,
              ),
            ),
            if (notice != null)
              FormNotice(message: describeNotice(notice, l10n)),
            for (final option in _fields) option.build(l10n),
          ],
        );
      },
    );
  }
}
