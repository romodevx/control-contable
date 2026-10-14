/// Which date books an income (MODULE-incomes §2.2): the ISR date, the one
/// the regime's basis requires and the one that decides its list month.
enum IncomeBasis {
  /// PM Título II: the invoice date (`invoiced_on`); IVA still waits for
  /// the collection ("Por cobrar").
  accrual,

  /// Every persona física regime, 626 PM, 603 and 610: the collection date
  /// (`collected_on`).
  cash,
}

/// The accrual regimes of spec §2.2. The server keeps this as catalog data
/// (`regimes.income_basis`) and is the authority (`400 required`); the
/// contract does not expose it, so the form uses this only to choose which
/// date to show first.
const Set<String> _accrualRegimes = {'601', '620', '622', '623', '624'};

/// The basis of [regimeCode].
IncomeBasis incomeBasisOf(String regimeCode) =>
    _accrualRegimes.contains(regimeCode)
    ? IncomeBasis.accrual
    : IncomeBasis.cash;
