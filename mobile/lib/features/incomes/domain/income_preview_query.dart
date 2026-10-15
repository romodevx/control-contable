import 'package:control_contable/api/generated/models/counterparty_kind.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';

/// The facts `previewIncomeWithholdings` derives from (MODULE-withholdings
/// §3.1). A `null` optional fact is left to the server's default. Records
/// compare by value, so an unchanged form asks nothing again.
typedef IncomePreviewQuery = ({
  String regime,
  String? on,
  int amountCents,
  IvaRate ivaRate,
  int ivaCents,
  CounterpartyKind counterpartyKind,
  OperationKind? operationKind,
  bool? residentialRent,
  bool? viaPlatform,
  PlatformActivity? platformActivity,
});
