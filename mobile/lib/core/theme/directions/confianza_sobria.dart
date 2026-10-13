import 'package:control_contable/core/theme/app_colors.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/theme/design_direction.dart';
import 'package:flutter/painting.dart';

// Direction "A · Confianza sobria" (docs/design/DIRECTIONS.md, PR #3).
// Contrast of every documented pair is asserted in theme_tokens_test.dart.

const _radii = AppRadii(control: 6, button: 8, card: 12);

const _typography = AppTypography(
  displayFamily: 'Source Serif 4',
  bodyFamily: 'Public Sans',
  hero: 44,
  page: 28,
  title: 20,
  section: 18,
  input: 17,
  body: 16,
  help: 14,
  caption: 13,
);

const _light = AppColors(
  bg: Color(0xFFF5F7FA),
  surface: Color(0xFFFFFFFF),
  text: Color(0xFF0E1B2C),
  muted: Color(0xFF4A5568),
  outline: Color(0xFF7A8797),
  border: Color(0xFFE1E6ED),
  primary: Color(0xFF1D4E89),
  onPrimary: Color(0xFFFFFFFF),
  heroBg: Color(0xFF1D4E89),
  heroText: Color(0xFFFFFFFF),
  heroMuted: Color(0xFFC7D7EC),
  iva: Color(0xFF1D4E89),
  isr: Color(0xFF2B7470),
  favor: Color(0xFF1C7046),
  legendBg: Color(0xFFFFF4D6),
  legendText: Color(0xFF5C3F00),
  error: Color(0xFFB42318),
);

const _dark = AppColors(
  bg: Color(0xFF0B1420),
  surface: Color(0xFF13202F),
  text: Color(0xFFE8EEF5),
  muted: Color(0xFFA9B6C6),
  outline: Color(0xFF6B7D93),
  border: Color(0xFF24344A),
  primary: Color(0xFF8DB8EC),
  onPrimary: Color(0xFF0B1420),
  heroBg: Color(0xFF16365C),
  heroText: Color(0xFFE8EEF5),
  heroMuted: Color(0xFFB4C6DD),
  iva: Color(0xFF8DB8EC),
  isr: Color(0xFF6FC7C0),
  favor: Color(0xFF6FD49A),
  legendBg: Color(0xFF2B2310),
  legendText: Color(0xFFF2D58A),
  error: Color(0xFFFF8A80),
);

/// "A · Confianza sobria": institutional ink blue, serif amounts, small
/// radii, no shadows.
const confianzaSobria = DesignDirection(
  name: 'A · Confianza sobria',
  light: AppTokens(colors: _light, radii: _radii, typography: _typography),
  dark: AppTokens(colors: _dark, radii: _radii, typography: _typography),
);
