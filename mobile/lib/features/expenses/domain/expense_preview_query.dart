import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/supplier_kind.dart';
import 'package:control_contable/api/generated/models/supplier_regime.dart';

/// The facts `previewExpenseWithholdings` derives from (persona moral
/// taxpayers only, MODULE-withholdings §2.2, §3.1). A `null` optional fact
/// is left to the server's default. Records compare by value, so an
/// unchanged form asks nothing again.
typedef ExpensePreviewQuery = ({
  String regime,
  String? on,
  int amountCents,
  IvaRate ivaRate,
  int ivaCents,
  String categoryCode,
  SupplierKind supplierKind,
  SupplierRegime? supplierRegime,
  OperationKind? operationKind,
  bool? residentialRent,
});
