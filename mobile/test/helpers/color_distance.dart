import 'dart:math' as math;

import 'package:flutter/painting.dart';

/// Color-distance checks of the dataviz palette validator
/// (`validate_palette.js`), so chart tokens are held to the same floor in
/// CI: Delta E is the Euclidean OKLab distance ×100, and color-vision
/// deficiency is simulated with Machado, Oliveira & Fernandes (2009) at
/// severity 1.0 on linear RGB.

/// Normal-vision floor: below it, full-color readers cannot tell a pair
/// apart (a hard fail, secondary encoding does not excuse it).
const double normalVisionFloor = 15;

/// CVD target: protan and deutan Delta E at or above it need no secondary
/// encoding.
const double cvdTarget = 8;

const _protan = [
  [0.152286, 1.052583, -0.204868],
  [0.114503, 0.786281, 0.099216],
  [-0.003882, -0.048116, 1.051998],
];

const _deutan = [
  [0.367322, 0.860646, -0.227968],
  [0.280085, 0.672501, 0.047413],
  [-0.011820, 0.042940, 0.968881],
];

double _toLinear(double channel) => channel <= 0.04045
    ? channel / 12.92
    : math.pow((channel + 0.055) / 1.055, 2.4).toDouble();

List<double> _linearRgb(Color color) => [
  _toLinear(color.r),
  _toLinear(color.g),
  _toLinear(color.b),
];

List<double> _oklab(List<double> rgb) {
  final [r, g, b] = rgb;
  double cbrt(double v) => v.sign * math.pow(v.abs(), 1 / 3);
  final l = cbrt(0.4122214708 * r + 0.5363588147 * g + 0.0514500430 * b);
  final m = cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b);
  final s = cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b);
  return [
    0.2104542553 * l + 0.7936177850 * m - 0.0040720468 * s,
    1.9779984951 * l - 2.4285922050 * m + 0.4505937099 * s,
    0.0259040371 * l + 0.7827717662 * m - 0.8086757660 * s,
  ];
}

List<double> _simulate(List<List<double>> matrix, List<double> rgb) => [
  for (final row in matrix) row[0] * rgb[0] + row[1] * rgb[1] + row[2] * rgb[2],
];

double _distance(List<double> a, List<double> b) =>
    100 *
    math.sqrt(
      math.pow(a[0] - b[0], 2) +
          math.pow(a[1] - b[1], 2) +
          math.pow(a[2] - b[2], 2),
    );

/// Delta E between [a] and [b] under normal vision.
double normalDeltaE(Color a, Color b) =>
    _distance(_oklab(_linearRgb(a)), _oklab(_linearRgb(b)));

/// The worse of the protan and deutan Delta E between [a] and [b].
double cvdDeltaE(Color a, Color b) {
  final la = _linearRgb(a);
  final lb = _linearRgb(b);
  return [_protan, _deutan]
      .map((m) => _distance(_oklab(_simulate(m, la)), _oklab(_simulate(m, lb))))
      .reduce(math.min);
}
