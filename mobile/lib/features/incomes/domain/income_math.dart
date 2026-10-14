import 'package:flutter/foundation.dart';

/// The money of one income, in centavos: what the form shows live and the
/// list row totals.
@immutable
class IncomeAmounts {
  /// The amounts of one income.
  const new({
    required this.amountCents,
    required this.ivaCents,
    this.isrWithheldCents = 0,
    this.ivaWithheldCents = 0,
  });

  /// Subtotal before IVA.
  final int amountCents;

  /// IVA charged.
  final int ivaCents;

  /// ISR the payer withheld.
  final int isrWithheldCents;

  /// IVA the payer withheld.
  final int ivaWithheldCents;

  /// Subtotal plus IVA (the CFDI total).
  int get totalCents => amountCents + ivaCents;

  /// Everything withheld.
  int get withheldCents => isrWithheldCents + ivaWithheldCents;

  /// What reaches the user: the total minus the withholdings.
  int get netCents => totalCents - withheldCents;
}
