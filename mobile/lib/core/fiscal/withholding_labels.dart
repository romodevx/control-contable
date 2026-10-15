import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/yes_no.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/l10n/app_localizations.dart';

/// What was sold or bought, in plain words (MODULE-withholdings §2.1 Q2,
/// §2.2 E3). There is no "subcontratación": its 6 % was repealed (WH-07).
String operationKindLabel(OperationKind kind, AppLocalizations l10n) =>
    switch (kind) {
      OperationKind.professionalServices =>
        l10n.withholdingOperationProfessional,
      OperationKind.businessServices => l10n.withholdingOperationBusiness,
      OperationKind.rentRealEstate => l10n.withholdingOperationRentRealEstate,
      OperationKind.rentMovable => l10n.withholdingOperationRentMovable,
      OperationKind.saleOfGoods => l10n.withholdingOperationGoods,
      OperationKind.groundFreight => l10n.withholdingOperationFreight,
      OperationKind.commission => l10n.withholdingOperationCommission,
      OperationKind.scrap => l10n.withholdingOperationScrap,
      OperationKind.other ||
      OperationKind.$unknown => l10n.withholdingOperationOther,
    };

/// Every operation, as choice options in the catalog's order.
List<ChoiceOption> operationKindOptions(AppLocalizations l10n) => [
  for (final kind in OperationKind.$valuesDefined)
    (value: kind.json!, label: operationKindLabel(kind, l10n)),
];

/// The options of a yes/no question.
List<ChoiceOption> yesNoOptions(AppLocalizations l10n) => [
  (value: yes, label: l10n.withholdingYes),
  (value: no, label: l10n.withholdingNo),
];

/// What a withholding warning says (§6), in the app's es-MX copy; the
/// deduction note names [remitDueOn]. Any other code reads the server's
/// message, which the contract makes safe to show.
String withholdingWarningText(
  Warning warning,
  AppLocalizations l10n, {
  String? remitDueOn,
}) => switch (warning.code) {
  WarningCode.withholdingCpaUnverified => l10n.withholdingWarningCpaUnverified,
  WarningCode.foreignCounterpartyNotSupported => l10n.withholdingWarningForeign,
  WarningCode.supplierRegimeUnknown =>
    l10n.withholdingWarningSupplierRegimeUnknown,
  WarningCode.deductionRequiresRemittance when remitDueOn != null =>
    l10n.withholdingWarningDeduction(parseBusinessDate(remitDueOn)),
  _ => warning.message,
};
