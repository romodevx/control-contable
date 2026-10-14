import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/tax_input_def.dart';
import 'package:control_contable/api/generated/models/tax_input_def_kind.dart';
import 'package:control_contable/api/generated/models/tax_input_value.dart';
import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/core/forms/checkbox_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/features/tax_engine/application/tax_input_entries.dart';
import 'package:control_contable/features/tax_engine/domain/tax_input_months.dart';
import 'package:control_contable/features/taxpayers/domain/decimal_rate.dart';

/// What the save sends, and which form field each value came from (by the
/// path the server names it with: `values.<index>`).
typedef TaxInputsUpdatePlan = ({
  List<TaxInputValue> values,
  Map<String, AppFormFieldModel> fieldsByPath,
});

/// The tax-inputs form of one regime and fiscal year: one entry per
/// definition, in the server's order, prefilled with the stored values.
/// Every field is optional (a missing input is never an error; the
/// estimate computes with zero and says so); the client checks formats
/// only and the server is the authority.
class TaxInputsFormModel {
  /// The form over [inputs] as of [current] (`YYYY-MM`). [monthLabel]
  /// names a month's field ("Pago de enero"); [checkRate] is the rate rule.
  new(
    TaxInputs inputs, {
    required String current,
    required String Function(int month) monthLabel,
    required FieldValidator checkRate,
  }) {
    entries = [
      for (final def in inputs.definitions)
        _entryOf(
          def,
          [
            for (final value in inputs.values)
              if (value.key == def.key) value,
          ],
          fiscalYear: inputs.fiscalYear,
          current: current,
          monthLabel: monthLabel,
          checkRate: checkRate,
        ),
    ];
  }

  /// The fields, their submit state and the error summary.
  final AppFormController form = AppFormController(
    marks: RequirementMark.requiredFields,
  );

  /// One per definition, in the server's order.
  late final List<TaxInputEntry> entries;

  TaxInputEntry _entryOf(
    TaxInputDef def,
    List<TaxInputValue> stored, {
    required int fiscalYear,
    required String current,
    required String Function(int month) monthLabel,
    required FieldValidator checkRate,
  }) {
    final first = stored.firstOrNull;
    return switch (def.kind) {
      TaxInputDefKind.money => MoneyEntry(
        def,
        _money(def.label, first?.valueCents),
      ),
      TaxInputDefKind.rate => RateEntry(
        def,
        form.textField(
          label: def.label,
          isRequired: false,
          validator: checkRate,
          canonicalize: (value) => parseDecimalRate(value) ?? value,
          initialText: switch (first?.valueRate) {
            final rate? => parseDecimalRate(rate) ?? rate,
            null => '',
          },
        ),
      ),
      TaxInputDefKind.boolean => FlagEntry(
        def,
        form.add(
          AppCheckboxController(
            label: def.label,
            isRequired: false,
            initialValue: first?.valueBool ?? false,
          ),
        ),
        wasStored: first?.valueBool != null,
      ),
      TaxInputDefKind.moneyMonthly => _monthly(
        def,
        stored,
        fiscalYear: fiscalYear,
        current: current,
        monthLabel: monthLabel,
      ),
      TaxInputDefKind.$unknown => KeptEntry(def, stored),
    };
  }

  MonthlyEntry _monthly(
    TaxInputDef def,
    List<TaxInputValue> stored, {
    required int fiscalYear,
    required String current,
    required String Function(int month) monthLabel,
  }) {
    final byMonth = {
      for (final value in stored) ?value.month: value.valueCents,
    };
    final months = paymentMonths(
      fiscalYear: fiscalYear,
      current: current,
      stored: byMonth.keys.toSet(),
    );
    return MonthlyEntry(def, {
      for (final month in months)
        month: _money(monthLabel(month), byMonth[month]),
    });
  }

  MoneyFieldController _money(String label, int? centavos) => form.moneyField(
    label: label,
    isRequired: false,
    initialText: centavos == null ? '' : formatCentavos(centavos),
  );

  /// The field to open the form at for [key] (the first input the
  /// estimate found missing): a monthly key's first empty month, else its
  /// first month; `null` when no field edits [key].
  AppFormFieldModel? focusTarget(String? key) {
    final entry = entries.where((e) => e.def.key == key).firstOrNull;
    return switch (entry) {
      MoneyEntry(:final field) => field,
      RateEntry(:final field) => field,
      FlagEntry(:final field) => field,
      MonthlyEntry(:final months) =>
        months.values.where((f) => f.value.isEmpty).firstOrNull ??
            months.values.firstOrNull,
      KeptEntry() || null => null,
    };
  }

  /// The index (in `monthGroups`) of the half of the year holding
  /// [field], when it is a month of the monthly grid.
  int? groupOf(AppFormFieldModel? field) {
    for (final entry in entries.whereType<MonthlyEntry>()) {
      final month = entry.months.entries
          .where((e) => identical(e.value, field))
          .firstOrNull
          ?.key;
      if (month == null) continue;
      return monthGroups(entry.months.keys.toList())
          .indexWhere((group) => group.contains(month));
    }
    return null;
  }

  /// The complete set of values to store, in definition order then month;
  /// an empty field stores nothing for its key.
  TaxInputsUpdatePlan toUpdate() {
    final values = <TaxInputValue>[];
    final fieldsByPath = <String, AppFormFieldModel>{};
    void add(TaxInputValue value, [AppFormFieldModel? field]) {
      if (field != null) fieldsByPath['values.${values.length}'] = field;
      values.add(value);
    }

    for (final entry in entries) {
      final key = entry.def.key;
      switch (entry) {
        case MoneyEntry(:final field):
          if (field.centavos case final cents?) {
            add(TaxInputValue(key: key, valueCents: cents), field);
          }
        case RateEntry(:final field):
          if (parseDecimalRate(field.value) case final rate?) {
            add(TaxInputValue(key: key, valueRate: rate), field);
          }
        case FlagEntry(:final field, :final wasStored):
          if (field.checked || wasStored) {
            add(TaxInputValue(key: key, valueBool: field.checked), field);
          }
        case MonthlyEntry(:final months):
          for (final MapEntry(key: month, value: field) in months.entries) {
            if (field.centavos case final cents?) {
              add(
                TaxInputValue(key: key, month: month, valueCents: cents),
                field,
              );
            }
          }
        case KeptEntry(:final stored):
          stored.forEach(add);
      }
    }
    return (values: values, fieldsByPath: fieldsByPath);
  }

  /// Shows the server's verdict ([details], by `values.<index>` of [plan])
  /// on the fields the values came from. Answers the half of the year
  /// holding the first rejected month (to open it) and whether a detail
  /// named no field (for a form-level message).
  ({int? group, bool hasUnmatched}) showVerdict(
    List<api.FieldError> details,
    TaxInputsUpdatePlan plan, {
    required FieldError Function(api.FieldError detail) describe,
  }) {
    final verdict = matchServerFieldErrors(
      details,
      fieldsByPath: plan.fieldsByPath,
      describe: describe,
    );
    form.showServerErrors(verdict.matched);
    return (
      group: groupOf(verdict.matched.keys.firstOrNull),
      hasUnmatched: verdict.unmatched.isNotEmpty,
    );
  }

  /// Releases the fields.
  void dispose() => form.dispose();
}
