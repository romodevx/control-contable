import 'package:intl/intl.dart';

/// Centavos per peso.
const _centavosPerPeso = 100;

/// Integer digits accepted (up to 9,999,999,999,999.99): far above any
/// personal tax amount, and far below where `int` arithmetic overflows.
const _maxIntegerDigits = 13;

/// A peso amount as people type it in es-MX: optional `$`, comma thousands
/// separators in groups of three (or none), a dot and one or two decimals.
final _amount = RegExp(r'^(\d{1,3}(?:,\d{3})+|\d+)?(?:\.(\d{1,2}))?$');

final _currencyAndSpaces = RegExp(r'[\s$]');

final _groupedPesos = NumberFormat('#,##0', 'es_MX');

/// The amount in [text] as integer centavos, or `null` when it is not a
/// non-negative peso amount.
///
/// Accepts `1,234.50`, `1234.5`, `$ 1,234.50`, `.5`. Rejects the European
/// `1.234,50`, misplaced commas (`1,23`) and more than two decimals: a
/// misread amount is worse than asking again.
int? parseCentavos(String text) {
  final match = _amount.firstMatch(text.replaceAll(_currencyAndSpaces, ''));
  if (match == null) return null;
  final pesos = match.group(1)?.replaceAll(',', '') ?? '';
  final decimals = match.group(2) ?? '';
  if (pesos.isEmpty && decimals.isEmpty) return null;
  if (pesos.length > _maxIntegerDigits) return null;
  final centavos = decimals.padRight(2, '0');
  return int.parse(pesos.isEmpty ? '0' : pesos) * _centavosPerPeso +
      int.parse(centavos);
}

/// The minus sign of a negative amount (U+2212, never a hyphen: DESIGN
/// "Numbers").
const String _minus = '−';

/// [centavos] as an es-MX currency amount: `$1,234.50`, `−$2,066.67`.
/// Integer arithmetic only.
String formatPesos(int centavos) {
  final sign = centavos < 0 ? _minus : '';
  return '$sign\$${formatCentavos(centavos.abs())}';
}

/// [centavos] in the es-MX canonical form, without the currency sign:
/// `1,234.50`. Integer arithmetic only: no binary-float rounding.
String formatCentavos(int centavos) {
  final pesos = _groupedPesos.format(centavos ~/ _centavosPerPeso);
  final cents = (centavos % _centavosPerPeso).toString().padLeft(2, '0');
  return '$pesos.$cents';
}
