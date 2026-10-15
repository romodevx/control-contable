import 'package:control_contable/api/generated/models/filed_payment_month.dart';
import 'package:control_contable/api/generated/models/filed_payment_source.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/features/annual/application/filed_payments_form_model.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// One month's inline row: its amount field, labelled by the month, with
/// where the prefilled amount comes from under it ("Estimado por la app —
/// confírmalo"). A month outside the regime's window is a plain line.
class FiledPaymentRowView extends StatelessWidget {
  /// The row of [row].
  const new({required this.row, super.key});

  /// The month and its field.
  final FiledPaymentRow row;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final field = row.field;
    if (field == null) {
      return MergeSemantics(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(row.label, style: Theme.of(context).textTheme.bodyLarge),
            Text(
              l10n.annualPaymentsOutOfWindow,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
          ],
        ),
      );
    }
    return AppMoneyField(
      field: field,
      textInputAction: TextInputAction.next,
      helperText: filedPaymentSourceLabel(row.month, l10n),
    );
  }
}

/// Where [month]'s prefilled amount comes from, in words.
String filedPaymentSourceLabel(
  FiledPaymentMonth month,
  AppLocalizations l10n,
) => switch (month.source) {
  FiledPaymentSource.filed => l10n.annualPaymentsSourceFiled,
  FiledPaymentSource.taxInput => l10n.annualPaymentsSourceTaxInput,
  FiledPaymentSource.estimatedByApp when month.appEstimateIncomplete =>
    l10n.annualPaymentsSourceIncomplete,
  FiledPaymentSource.estimatedByApp => l10n.annualPaymentsSourceEstimated,
  FiledPaymentSource.none ||
  FiledPaymentSource.$unknown => l10n.annualPaymentsSourceNone,
};
