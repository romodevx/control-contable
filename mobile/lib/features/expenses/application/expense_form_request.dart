import 'package:control_contable/api/generated/models/deductible_source.dart';
import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_create.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';

/// The `updateExpense` key of the deductibility override.
const String _deductibleKey = 'deductible';

/// What the expense form sends, and where the server's verdicts land.
extension ExpenseFormRequest on ExpenseFormModel {
  /// The fields a server path names, for `matchServerFieldErrors`.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    'category_code': category,
    'amount_cents': amount,
    'iva_rate': ivaRate,
    'paid_on': paidOn,
    'payment_method': paymentMethod,
    'has_cfdi': hasCfdi,
    if (regimes.length > 1) 'regime_code': regime,
    'iva_cents': ivaAmount,
    'deductible': deductible,
    if (withCfdi) 'cfdi_uuid': cfdiUuid,
    'description': description,
    if (isMoral) ...{
      'isr_withheld_cents': isrWithheld,
      'iva_withheld_cents': ivaWithheld,
    },
  };

  /// The request body (call only after the form validated). The folio
  /// travels only with a CFDI, the withholdings only for a persona moral.
  ExpenseCreate toCreate() {
    final uuid = cfdiUuid.value;
    int? withheld(int? cents) => isMoral && (cents ?? 0) > 0 ? cents : null;
    return ExpenseCreate(
      id: sessionId,
      regimeCode: regimeCode,
      categoryCode: category.value!,
      amountCents: amount.centavos!,
      ivaRate: IvaRate.fromJson(ivaRate.value!),
      ivaCents: ivaCents,
      paymentMethod: method,
      paidOn: businessDate(paidOn.date!),
      hasCfdi: withCfdi,
      cfdiUuid: withCfdi && uuid.isNotEmpty ? uuid.toUpperCase() : null,
      deductible: deductibleOverride,
      isrWithheldCents: withheld(isrWithheld.centavos),
      ivaWithheldCents: withheld(ivaWithheld.centavos),
      description: description.value.isEmpty ? null : description.value,
    );
  }

  /// The merge-patch of an edit (call only after the form validated): the
  /// `updateExpense` keys whose value differs from [stored], with `null`
  /// where an optional field or the override was cleared. Empty when
  /// nothing changed.
  Map<String, Object?> toPatch(Expense stored) {
    final wanted = toCreate();
    final desired = <String, Object?>{
      'regime_code': wanted.regimeCode,
      'category_code': wanted.categoryCode,
      'amount_cents': wanted.amountCents,
      'iva_rate': wanted.ivaRate.json,
      'iva_cents': wanted.ivaCents,
      'payment_method': wanted.paymentMethod.json,
      'paid_on': wanted.paidOn,
      'has_cfdi': wanted.hasCfdi,
      'cfdi_uuid': wanted.cfdiUuid,
      'isr_withheld_cents': wanted.isrWithheldCents ?? 0,
      'iva_withheld_cents': wanted.ivaWithheldCents ?? 0,
      'description': wanted.description,
    };
    final current = stored.toJson();
    final storedOverride = stored.deductibleSource == DeductibleSource.user
        ? stored.deductible
        : null;
    return {
      for (final MapEntry(:key, :value) in desired.entries)
        if (current[key] != value) key: value,
      if (wanted.deductible != storedOverride)
        _deductibleKey: wanted.deductible,
    };
  }
}
