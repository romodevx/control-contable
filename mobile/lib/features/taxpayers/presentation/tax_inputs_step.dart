import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/tax_input_def.dart';
import 'package:control_contable/api/generated/models/tax_input_def_kind.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/domain/decimal_rate.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Optional step 5 (`module_tax_engine` on, OQ-3): the per-fiscal-year
/// inputs a regime's estimate uses (coeficiente de utilidad, year-to-date
/// totals), one regime per page. Every field is optional; *Omitir* moves
/// on. Written through M5's tax-inputs endpoint once the taxpayer exists.
class TaxInputsStep extends StatefulWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The onboarding.
  final OnboardingFlow flow;

  @override
  State<TaxInputsStep> createState() => _TaxInputsStepState();
}

typedef _Input = ({TaxInputDef def, AppFieldController field});

class _TaxInputsStepState extends State<TaxInputsStep> {
  // Every field is optional: none carries a mark; the intro says so.
  final AppFormController _form = AppFormController(
    marks: RequirementMark.requiredFields,
  );
  List<_Input>? _inputs;

  // Fixed when the page opens (see RegimeOptionsStep).
  late final Regime _regime = _flow.inputRegimes[_flow.page];

  OnboardingFlow get _flow => widget.flow;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_inputs != null) return;
    final l10n = AppLocalizations.of(context);
    _inputs = [
      for (final def in _regime.requiredInputs)
        if (_fieldFor(def, l10n) case final field?) (def: def, field: field),
    ];
  }

  AppFieldController? _fieldFor(TaxInputDef def, AppLocalizations l10n) =>
      switch (def.kind) {
        TaxInputDefKind.money => _form.add(
          MoneyFieldController(label: def.label, isRequired: false),
        ),
        TaxInputDefKind.rate => _form.textField(
          label: def.label,
          isRequired: false,
          validator: (value) => checkRate(value, l10n),
          canonicalize: (value) => parseDecimalRate(value) ?? value,
        ),
        // Monthly amounts and yes/no inputs belong to the per-regime
        // tax-inputs form (M5-27), not to this optional onboarding step.
        TaxInputDefKind.moneyMonthly ||
        TaxInputDefKind.boolean ||
        TaxInputDefKind.$unknown => null,
      };

  List<TaxInputValue> _values() => [
    for (final (:def, :field) in _inputs!)
      if (field.value.isNotEmpty)
        switch (field) {
          MoneyFieldController(:final centavos) => TaxInputValue(
            key: def.key,
            valueCents: centavos,
          ),
          _ => TaxInputValue(
            key: def.key,
            valueRate: parseDecimalRate(field.value),
          ),
        },
  ];

  @override
  void dispose() {
    _form.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: _flow,
      builder: (context, _) {
        final notice = _flow.notice;
        return AppForm(
          controller: _form,
          onSave: () => _flow.saveTaxInputs(_values()),
          children: [
            StepHeader(
              flow: _flow,
              title: l10n.onboardingInputsTitle,
              intro: l10n.onboardingInputsIntro('${_flow.fiscalYear}'),
            ),
            Text(
              regimeTitle(_regime),
              style: Theme.of(context).textTheme.titleMedium,
            ),
            if (notice != null)
              FormNotice(message: describeNotice(notice, l10n)),
            for (final (:def, :field) in _inputs!)
              AppTextField(
                field: field,
                keyboardType: decimalKeyboard,
                textInputAction: TextInputAction.done,
                autofillHints: const [],
                autocorrect: false,
                prefixText: def.kind == TaxInputDefKind.money ? r'$ ' : null,
              ),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton(
                onPressed: () => _flow.saveTaxInputs(const []),
                child: Text(l10n.onboardingInputsSkip),
              ),
            ),
          ],
        );
      },
    );
  }
}
