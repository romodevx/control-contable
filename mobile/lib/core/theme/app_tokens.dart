import 'dart:ui' show lerpDouble;

import 'package:control_contable/core/theme/app_colors.dart';
import 'package:flutter/material.dart';

/// Corner radii of a direction: control, button, card.
@immutable
class AppRadii {
  /// Creates the radii set.
  const new({required this.control, required this.button, required this.card});

  /// Inputs, chips, small controls.
  final double control;

  /// Buttons.
  final double button;

  /// Cards and sheets.
  final double card;

  /// Interpolates every radius.
  AppRadii lerp(AppRadii other, double t) => AppRadii(
    control: lerpDouble(control, other.control, t)!,
    button: lerpDouble(button, other.button, t)!,
    card: lerpDouble(card, other.card, t)!,
  );
}

/// Faces and type scale (logical pixels) of a direction.
@immutable
class AppTypography {
  /// Creates the typography set.
  const new({
    required this.displayFamily,
    required this.bodyFamily,
    required this.hero,
    required this.page,
    required this.title,
    required this.section,
    required this.input,
    required this.body,
    required this.help,
    required this.caption,
  });

  /// Hero amounts, screen and section titles.
  final String displayFamily;

  /// Body, UI and numbers.
  final String bodyFamily;

  /// Hero amount size.
  final double hero;

  /// Page title size.
  final double page;

  /// Card/dialog title size.
  final double title;

  /// Section heading size.
  final double section;

  /// Input and button text size (house rule: ≥ 16).
  final double input;

  /// Body text size.
  final double body;

  /// Help text size.
  final double help;

  /// Caption size.
  final double caption;
}

/// The design-token theme extension. A direction is one light and one dark
/// [AppTokens]; widgets read them with `context.tokens` and never hard-code
/// a color, radius or face.
@immutable
class AppTokens extends ThemeExtension<AppTokens> {
  /// Creates the token set.
  const new({
    required this.colors,
    required this.radii,
    required this.typography,
  });

  /// Semantic colors.
  final AppColors colors;

  /// Corner radii.
  final AppRadii radii;

  /// Faces and type scale.
  final AppTypography typography;

  @override
  AppTokens copyWith({
    AppColors? colors,
    AppRadii? radii,
    AppTypography? typography,
  }) => AppTokens(
    colors: colors ?? this.colors,
    radii: radii ?? this.radii,
    typography: typography ?? this.typography,
  );

  @override
  AppTokens lerp(covariant AppTokens? other, double t) {
    if (other == null) return this;
    return AppTokens(
      colors: colors.lerp(other.colors, t),
      radii: radii.lerp(other.radii, t),
      // Faces cannot interpolate; switch at the midpoint.
      typography: t < 0.5 ? typography : other.typography,
    );
  }
}

/// `context.tokens`: the active direction's tokens for the current brightness.
extension AppTokensContext on BuildContext {
  /// The [AppTokens] of the nearest [Theme]. Always registered by the app
  /// theme, so a missing extension is a programming error.
  AppTokens get tokens => Theme.of(this).extension<AppTokens>()!;
}
