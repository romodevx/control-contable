import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/income_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/features/incomes/domain/income_basis.dart';

/// What the new-income form sends, and where the server's verdicts land.
extension IncomeFormRequest on IncomeFormModel {
  /// The fields a server path names, for `matchServerFieldErrors`.
  Map<String, AppFormFieldModel> get fieldsByPath {
    final accrual = basis == IncomeBasis.accrual;
    return {
      'amount_cents': amount,
      'iva_cents': ivaAmount,
      'iva_rate': ivaRate,
      'counterparty_type': counterparty,
      if (regimes.length > 1) 'regime_code': regime,
      if (activities.length > 1) 'platform_activity': activity,
      'isr_withheld_cents': isrWithheld,
      'iva_withheld_cents': ivaWithheld,
      'collected_on': accrual ? otherDate : bookingDate,
      'invoiced_on': accrual ? bookingDate : otherDate,
      'cfdi_uuid': cfdiUuid,
      'description': description,
    };
  }

  /// A 625 income from several activities needs one chosen: its error,
  /// shown before anything is sent.
  Map<AppFormFieldModel, FieldError> get missingChoices => {
    if (activities.length > 1 && activity.value == null)
      activity: const MissingChoice(),
  };

  /// The request body (call only after the form validated).
  IncomeCreate toCreate() {
    final accrual = basis == IncomeBasis.accrual;
    final other = otherDate.date;
    final otherDay = other == null ? null : businessDate(other);
    final withheld = amounts;
    return IncomeCreate(
      id: sessionId,
      regimeCode: regimeCode,
      counterpartyType: payer,
      amountCents: amount.centavos!,
      ivaRate: IvaRate.fromJson(ivaRate.value!),
      ivaCents: amounts.ivaCents,
      isrWithheldCents: withheld.isrWithheldCents == 0
          ? null
          : withheld.isrWithheldCents,
      ivaWithheldCents: withheld.ivaWithheldCents == 0
          ? null
          : withheld.ivaWithheldCents,
      invoicedOn: accrual ? bookingDay : otherDay,
      collectedOn: accrual ? otherDay : bookingDay,
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
      'counterparty_type': wanted.counterpartyType.json,
      'amount_cents': wanted.amountCents,
      'iva_rate': wanted.ivaRate.json,
      'iva_cents': wanted.ivaCents,
      'isr_withheld_cents': wanted.isrWithheldCents ?? 0,
      'iva_withheld_cents': wanted.ivaWithheldCents ?? 0,
      'invoiced_on': wanted.invoicedOn,
      'collected_on': wanted.collectedOn,
      'platform_activity': wanted.platformActivity?.json,
      'cfdi_uuid': wanted.cfdiUuid,
      'description': wanted.description,
    };
    final current = stored.toJson();
    return {
      for (final MapEntry(:key, :value) in desired.entries)
        if (current[key] != value) key: value,
    };
  }
}
