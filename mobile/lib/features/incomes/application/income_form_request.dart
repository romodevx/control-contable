import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';

/// What the income form sends, and where the server's verdicts land. No
/// withholding amount is ever sent: the server derives it from the facts
/// (ADR-0008).
extension IncomeFormRequest on IncomeFormModel {
  /// The fields a server path names, for `matchServerFieldErrors`.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    'amount_cents': amount,
    'iva_cents': ivaAmount,
    'iva_rate': ivaRate,
    ...facts.fieldsByPath,
    if (regimes.length > 1) 'regime_code': regime,
    if (activities.length > 1) 'platform_activity': activity,
    'collected_on': collectedOn,
    'invoiced_on': invoicedOn,
    'cfdi_uuid': cfdiUuid,
    'description': description,
  };

  /// A 625 income from several activities needs one chosen: its error,
  /// shown before anything is sent.
  Map<AppFormFieldModel, FieldError> get missingChoices => {
    if (activities.length > 1 && activity.value == null)
      activity: const MissingChoice(),
  };

  /// The request body (call only after the form validated).
  IncomeCreate toCreate() {
    return IncomeCreate(
      id: sessionId,
      regimeCode: regimeCode,
      // Q1 is required: the form validated it.
      counterpartyKind: facts.kind!,
      operationKind: facts.operation,
      residentialRent: facts.residentialRentSent,
      viaPlatform: facts.viaPlatformSent,
      amountCents: amount.centavos!,
      ivaRate: IvaRate.fromJson(ivaRate.value!),
      ivaCents: amounts.ivaCents,
      invoicedOn: _dayOf(invoicedOn.date),
      collectedOn: _dayOf(collectedOn.date),
      platformActivity: platformActivity,
      cfdiUuid: cfdiUuid.value.isEmpty ? null : cfdiUuid.value.toUpperCase(),
      description: description.value.isEmpty ? null : description.value,
    );
  }

  /// The merge-patch of an edit (call only after the form validated): the
  /// `updateIncome` keys whose value differs from [stored], with `null`
  /// where an optional field was cleared. Empty when nothing changed.
  Map<String, Object?> toPatch(Income stored) {
    final wanted = toCreate();
    final desired = <String, Object?>{
      'regime_code': wanted.regimeCode,
      'amount_cents': wanted.amountCents,
      'iva_rate': wanted.ivaRate.json,
      'iva_cents': wanted.ivaCents,
      'invoiced_on': wanted.invoicedOn,
      'collected_on': wanted.collectedOn,
      'platform_activity': wanted.platformActivity?.json,
      'cfdi_uuid': wanted.cfdiUuid,
      'description': wanted.description,
      'counterparty_kind': wanted.counterpartyKind.json,
      'operation_kind': ?wanted.operationKind?.json,
      'residential_rent': ?wanted.residentialRent,
      'via_platform': ?wanted.viaPlatform,
    };
    final current = stored.toJson();
    return {
      for (final MapEntry(:key, :value) in desired.entries)
        if (current[key] != value) key: value,
    };
  }

  String? _dayOf(DateTime? date) => date == null ? null : businessDate(date);
}
