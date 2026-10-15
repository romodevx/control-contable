import 'package:control_contable/api/generated/models/income_basis.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';

export 'package:control_contable/api/generated/models/income_basis.dart'
    show IncomeBasis;

/// Which date books an income of the held regime [code] (MODULE-incomes
/// §2.2): [IncomeBasis.accrual] books on the invoice date (`invoiced_on`),
/// [IncomeBasis.cash] on the collection date (`collected_on`).
///
/// The server's catalog decides it (`TaxpayerRegime.income_basis`,
/// ADR-0010); the app keeps no table of regime codes. A regime not in
/// [held], or a value this app does not know, reads as cash: the form then
/// shows the collection date first, and the server still answers
/// `400 required` on `invoiced_on` for an accrual regime.
IncomeBasis incomeBasisOf(String code, Iterable<TaxpayerRegime> held) {
  final served = held.where((r) => r.regimeCode == code).firstOrNull;
  return served?.incomeBasis == IncomeBasis.accrual
      ? IncomeBasis.accrual
      : IncomeBasis.cash;
}
