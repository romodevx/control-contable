import 'package:control_contable/api/generated/models/regime_option_def.dart';
import 'package:control_contable/api/generated/models/regime_option_def_type.dart';
import 'package:control_contable/core/forms/app_choice_fields.dart';
import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/features/taxpayers/domain/decimal_rate.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// Keyboard of a decimal field (a rate): digits and a decimal mark.
const TextInputType decimalKeyboard = TextInputType.numberWithOptions(
  decimal: true,
);

/// One per-regime option rendered from its `option_defs` entry: the form
/// field holding its value, the widget showing it, and its value for the
/// request (`null` when it sends nothing).
abstract class OptionField {
  /// The field for [def] in [form], starting at [initial] (the remembered
  /// value or the catalog default); `null` for a type this build does not
  /// know (the server fills or refuses it).
  static OptionField? create(
    RegimeOptionDef def,
    Object? initial,
    AppFormController form,
    AppLocalizations l10n,
  ) => switch (def.type) {
    RegimeOptionDefType.boolean => _BooleanOption(def, initial, form),
    RegimeOptionDefType.valueEnum => _EnumOption(def, initial, form),
    RegimeOptionDefType.multiEnum => _MultiEnumOption(def, initial, form),
    RegimeOptionDefType.rate => _RateOption(def, initial, form, l10n),
    RegimeOptionDefType.$unknown => null,
  };

  /// The option's key in `options`.
  String get key;

  /// The form field (for errors and the summary).
  AppFormFieldModel get field;

  /// The value to send, or `null` to send nothing.
  Object? get value;

  /// The widget showing the field.
  Widget build(AppLocalizations l10n);
}

class _BooleanOption implements OptionField {
  new(this.def, Object? initial, AppFormController form)
    : field = form.checkboxField(label: def.label, isRequired: false)
        ..checked = initial == true;

  final RegimeOptionDef def;

  @override
  final AppCheckboxController field;

  @override
  String get key => def.key;

  @override
  Object get value => field.checked;

  @override
  Widget build(AppLocalizations l10n) => ListenableBuilder(
    listenable: field,
    builder: (context, _) => SwitchListTile(
      value: field.checked,
      focusNode: field.focusNode,
      contentPadding: EdgeInsets.zero,
      title: Text(def.label),
      onChanged: (checked) => field.checked = checked,
    ),
  );
}

class _EnumOption implements OptionField {
  new(this.def, Object? initial, AppFormController form)
    : field = form.add(
        ChoiceFieldController(
          label: def.label,
          isRequired: def.requiredValue,
          initialValue: initial is String ? initial : null,
        ),
      );

  final RegimeOptionDef def;

  @override
  final ChoiceFieldController field;

  @override
  String get key => def.key;

  @override
  Object? get value => field.value;

  @override
  Widget build(AppLocalizations l10n) =>
      AppChoiceField(field: field, options: _optionsOf(def, l10n));
}

class _MultiEnumOption implements OptionField {
  new(this.def, Object? initial, AppFormController form)
    : field = form.add(
        MultiChoiceFieldController(
          label: def.label,
          isRequired: def.requiredValue,
          initialValues: {
            if (initial is List<Object?>) ...initial.whereType<String>(),
          },
        ),
      );

  final RegimeOptionDef def;

  @override
  final MultiChoiceFieldController field;

  @override
  String get key => def.key;

  /// The chosen values in catalog order.
  @override
  Object? get value {
    final chosen = [
      for (final allowed in def.allowedValues ?? const <String>[])
        if (field.values.contains(allowed)) allowed,
    ];
    return chosen.isEmpty ? null : chosen;
  }

  @override
  Widget build(AppLocalizations l10n) =>
      AppMultiChoiceField(field: field, options: _optionsOf(def, l10n));
}

class _RateOption implements OptionField {
  new(this.def, Object? initial, AppFormController form, AppLocalizations l10n)
    : field = form.textField(
        label: def.label,
        isRequired: def.requiredValue,
        initialText: initial is String ? initial : '',
        validator: (value) =>
            checkRate(value, l10n, min: def.min, max: def.max),
        canonicalize: (value) => parseDecimalRate(value) ?? value,
      );

  final RegimeOptionDef def;

  @override
  final AppFieldController field;

  @override
  String get key => def.key;

  @override
  Object? get value => parseDecimalRate(field.value);

  @override
  Widget build(AppLocalizations l10n) {
    final (min, max) = (def.min, def.max);
    return AppTextField(
      field: field,
      keyboardType: decimalKeyboard,
      textInputAction: TextInputAction.done,
      autofillHints: const [],
      autocorrect: false,
      helperText: min != null && max != null
          ? l10n.onboardingRateHelp(min, max)
          : null,
    );
  }
}

/// The client check of a typed rate against inclusive [min] and [max].
FieldError? checkRate(
  String value,
  AppLocalizations l10n, {
  String? min,
  String? max,
}) {
  final rate = parseDecimalRate(value);
  if (rate == null) return CustomFieldError(l10n.onboardingRateInvalid);
  if (isRateWithin(rate, min: min, max: max)) return null;
  return CustomFieldError(
    l10n.onboardingRateOutOfRange(min ?? '0', max ?? '999'),
  );
}

List<ChoiceOption> _optionsOf(RegimeOptionDef def, AppLocalizations l10n) => [
  for (final value in def.allowedValues ?? const <String>[])
    (value: value, label: optionValueLabel(def.key, value, l10n)),
];
