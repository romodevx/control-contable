/// Digits a percentage moves the decimal point by.
const int _percentShift = 2;

/// A no-break space keeps the figure and its `%` on one line.
const String _beforePercent = ' %';

final RegExp _trailingZeros = RegExp(r'0+$');
final RegExp _leadingZeros = RegExp('^0+(?=.)');

/// A `DecimalRate` ([rate], `0.2136`) as an es-MX percentage (`21.36 %`).
/// The decimal point moves on the string itself: the rate never passes
/// through a binary float, so nothing is rounded.
String formatRatePercent(String rate) {
  final [whole, ...rest] = rate.split('.');
  final fraction = (rest.isEmpty ? '' : rest.first).padRight(
    _percentShift,
    '0',
  );
  final integer = '$whole${fraction.substring(0, _percentShift)}'.replaceFirst(
    _leadingZeros,
    '',
  );
  final decimals = fraction
      .substring(_percentShift)
      .replaceFirst(_trailingZeros, '');
  final figure = decimals.isEmpty ? integer : '$integer.$decimals';
  return '$figure$_beforePercent';
}
