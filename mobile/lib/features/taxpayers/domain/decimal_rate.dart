/// A `DecimalRate` typed in es-MX: up to 3 integer digits and 6 decimals,
/// a point or a lone comma as the decimal mark (`0.25`, `.25`, `0,25`).
final RegExp _typedRate = RegExp(r'^(\d{0,3})(?:[.,](\d{0,6}))?$');

/// Decimal places a `DecimalRate` may carry (contract pattern).
const int _scale = 6;

/// Any digit: a rate needs at least one.
final RegExp _digit = RegExp(r'\d');

/// [typed] as a canonical `DecimalRate` string (`0.25`, `7.5`, `1`), or
/// `null` when it is not one. Never parsed through a float: the value
/// stays a string end to end.
String? parseDecimalRate(String typed) {
  final match = _typedRate.firstMatch(typed.trim());
  if (match == null) return null;
  final whole = match.group(1)!;
  final fraction = (match.group(2) ?? '').replaceFirst(RegExp(r'0+$'), '');
  if (!typed.contains(_digit)) return null;
  final integer = whole.replaceFirst(RegExp('^0+'), '');
  final integerPart = integer.isEmpty ? '0' : integer;
  return fraction.isEmpty ? integerPart : '$integerPart.$fraction';
}

/// Compares two canonical rates exactly (negative, zero or positive).
int compareDecimalRates(String a, String b) => _scaled(a).compareTo(_scaled(b));

/// Whether [rate] lies within the inclusive bounds, when given.
bool isRateWithin(String rate, {String? min, String? max}) =>
    (min == null || compareDecimalRates(rate, min) >= 0) &&
    (max == null || compareDecimalRates(rate, max) <= 0);

/// The rate as an integer number of millionths.
BigInt _scaled(String rate) {
  final [whole, ...rest] = rate.split('.');
  final fraction = (rest.isEmpty ? '' : rest.first).padRight(_scale, '0');
  return BigInt.parse('$whole$fraction');
}
