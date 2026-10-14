import 'package:control_contable/core/forms/money.dart';
import 'package:intl/intl.dart';

const int _centavosPerPeso = 100;

/// The minus sign (U+2212) of negative amounts (DIRECTIONS.md "Numbers").
const String minusSign = '−';

final NumberFormat _groupedPesos = NumberFormat('#,##0', 'es_MX');

/// [centavos] as es-MX money: `$2,284.29`, `−$924.00`. Integer arithmetic
/// only; the app formats and never rounds (MODULE-tax-engine §2.7).
String formatMoney(int centavos) {
  final sign = centavos < 0 ? minusSign : '';
  return '$sign\$${formatCentavos(centavos.abs())}';
}

/// [centavos] as a screen reader should say them: `2,284 pesos con 29
/// centavos`, `menos 924 pesos`. A bare `$2,284.29` is read inconsistently
/// (the `$` and the decimal point are spelled out by some readers).
String spokenMoney(int centavos) {
  final magnitude = centavos.abs();
  final pesos = magnitude ~/ _centavosPerPeso;
  final cents = magnitude % _centavosPerPeso;
  final pesosWords =
      '${_groupedPesos.format(pesos)} ${pesos == 1 ? 'peso' : 'pesos'}';
  final centsWords = cents == 0
      ? ''
      : ' con $cents ${cents == 1 ? 'centavo' : 'centavos'}';
  final sign = centavos < 0 ? 'menos ' : '';
  return '$sign$pesosWords$centsWords';
}
