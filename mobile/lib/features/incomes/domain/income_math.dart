import 'package:control_contable/api/generated/models/decimal_rate.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:flutter/foundation.dart';

/// IVA percent of each rate that charges IVA (LIVA 1 and the border
/// decree); `rate_0` and `exempt` charge none.
const Map<IvaRate, int> _ivaPercent = {IvaRate.rate16: 16, IvaRate.rate8: 8};

const int _percent = 100;

/// Decimal places of a `DecimalRate` (contract pattern).
const int _rateScale = 6;

/// The IVA [amountCents] (the subtotal) carries at [rate], rounded half up
/// to the centavo. The server accepts ±100 centavos around it (spec §2.5);
/// the user may correct it to the CFDI. Integer arithmetic only.
int ivaCentsFor(int amountCents, IvaRate rate) {
  final percent = _ivaPercent[rate];
  if (percent == null) return 0;
  return _roundHalfUp(
    BigInt.from(amountCents) * BigInt.from(percent),
    BigInt.from(_percent),
  );
}

/// [rate] (a `DecimalRate` such as `0.666667`) of [baseCents], rounded
/// half up: the withholding the defaults suggest. Never through a float.
int withheldCents(int baseCents, DecimalRate rate) {
  final [whole, ...rest] = rate.split('.');
  final fraction = (rest.isEmpty ? '' : rest.first).padRight(_rateScale, '0');
  final millionths = BigInt.parse('$whole$fraction');
  return _roundHalfUp(
    BigInt.from(baseCents) * millionths,
    BigInt.from(10).pow(_rateScale),
  );
}

int _roundHalfUp(BigInt numerator, BigInt denominator) =>
    ((numerator * BigInt.two + denominator) ~/ (denominator * BigInt.two))
        .toInt();

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
