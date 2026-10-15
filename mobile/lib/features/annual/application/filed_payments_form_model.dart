import 'package:control_contable/api/generated/models/filed_payment_entry.dart';
import 'package:control_contable/api/generated/models/filed_payment_month.dart';
import 'package:control_contable/api/generated/models/filed_payment_source.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';

/// One month of the filed payments form: the server's facts, the month's
/// name and, when the regime was held that month, its amount field.
typedef FiledPaymentRow = ({
  FiledPaymentMonth month,
  String label,
  MoneyFieldController? field,
});

/// The twelve provisional payments of one regime (ANN-21), one inline
/// amount per month prefilled with what the app knows (confirmed, the tax
/// input, or its own estimate). Absent from the PUT = not confirmed.
class FiledPaymentsFormModel {
  /// The form over [payments]; [monthLabel] names each month's field.
  new({
    required FiledPayments payments,
    required String Function(FiledPaymentMonth month) monthLabel,
  }) {
    rows = [
      for (final month in payments.months)
        (
          month: month,
          label: monthLabel(month),
          field: month.inWindow
              ? form.moneyField(
                  label: monthLabel(month),
                  isRequired: false,
                  initialText: _prefill(month),
                )
              : null,
        ),
    ];
  }

  /// The fields and submit state. Every month is optional, so none is
  /// marked "(opcional)" (the fewer kind is marked).
  final AppFormController form = AppFormController(
    marks: RequirementMark.requiredFields,
  );

  /// The twelve months, in order.
  late final List<FiledPaymentRow> rows;

  /// What "Guardar" sends: the months already confirmed and every month
  /// the user typed, each with a value. A cleared month is dropped.
  FiledPaymentsUpdate toUpdate() => _entries(
    (row) => row.month.source == FiledPaymentSource.filed || row.field!.isDirty,
  );

  /// What "Confirmar todos" sends: every month with a value shown.
  FiledPaymentsUpdate confirmAll() => _entries((_) => true);

  FiledPaymentsUpdate _entries(bool Function(FiledPaymentRow row) sends) =>
      FiledPaymentsUpdate(
        months: [
          for (final row in rows)
            if (row.field?.centavos case final cents? when sends(row))
              FiledPaymentEntry(month: row.month.month, amountCents: cents),
        ],
      );

  static String _prefill(FiledPaymentMonth month) =>
      switch (month.effectiveCents) {
        final cents? => formatCentavos(cents),
        null => '',
      };

  /// Releases the fields.
  void dispose() => form.dispose();
}
