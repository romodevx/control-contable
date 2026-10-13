import 'package:flutter/material.dart';

/// Semantic color tokens of a design direction, one set per brightness.
///
/// Names match docs/design/DIRECTIONS.md so a direction's palette table can be
/// copied in verbatim. Widgets never use raw hex values: they read these.
@immutable
class AppColors {
  /// Creates a full palette; every token is required.
  const new({
    required this.bg,
    required this.surface,
    required this.text,
    required this.muted,
    required this.outline,
    required this.border,
    required this.primary,
    required this.onPrimary,
    required this.heroBg,
    required this.heroText,
    required this.heroMuted,
    required this.iva,
    required this.isr,
    required this.favor,
    required this.legendBg,
    required this.legendText,
    required this.error,
  });

  /// Screen background.
  final Color bg;

  /// Cards, inputs, bottom bar.
  final Color surface;

  /// Body text.
  final Color text;

  /// Secondary text and help.
  final Color muted;

  /// Input and button borders (≥ 3:1 against [surface]).
  final Color outline;

  /// Decorative dividers; never the only boundary of a control.
  final Color border;

  /// Primary button, active navigation, links.
  final Color primary;

  /// Text and icons on [primary].
  final Color onPrimary;

  /// Estimate card background.
  final Color heroBg;

  /// Estimate card text.
  final Color heroText;

  /// Estimate card secondary text.
  final Color heroMuted;

  /// IVA chart series.
  final Color iva;

  /// ISR chart series.
  final Color isr;

  /// "Saldo a favor" amounts (always paired with text, never color alone).
  final Color favor;

  /// Mandatory legend block background.
  final Color legendBg;

  /// Mandatory legend block text.
  final Color legendText;

  /// Errors.
  final Color error;

  /// Returns a copy with the given tokens replaced.
  AppColors copyWith({
    Color? bg,
    Color? surface,
    Color? text,
    Color? muted,
    Color? outline,
    Color? border,
    Color? primary,
    Color? onPrimary,
    Color? heroBg,
    Color? heroText,
    Color? heroMuted,
    Color? iva,
    Color? isr,
    Color? favor,
    Color? legendBg,
    Color? legendText,
    Color? error,
  }) => AppColors(
    bg: bg ?? this.bg,
    surface: surface ?? this.surface,
    text: text ?? this.text,
    muted: muted ?? this.muted,
    outline: outline ?? this.outline,
    border: border ?? this.border,
    primary: primary ?? this.primary,
    onPrimary: onPrimary ?? this.onPrimary,
    heroBg: heroBg ?? this.heroBg,
    heroText: heroText ?? this.heroText,
    heroMuted: heroMuted ?? this.heroMuted,
    iva: iva ?? this.iva,
    isr: isr ?? this.isr,
    favor: favor ?? this.favor,
    legendBg: legendBg ?? this.legendBg,
    legendText: legendText ?? this.legendText,
    error: error ?? this.error,
  );

  /// Linear interpolation between two palettes, token by token.
  AppColors lerp(AppColors other, double t) {
    Color mix(Color a, Color b) => Color.lerp(a, b, t)!;
    return AppColors(
      bg: mix(bg, other.bg),
      surface: mix(surface, other.surface),
      text: mix(text, other.text),
      muted: mix(muted, other.muted),
      outline: mix(outline, other.outline),
      border: mix(border, other.border),
      primary: mix(primary, other.primary),
      onPrimary: mix(onPrimary, other.onPrimary),
      heroBg: mix(heroBg, other.heroBg),
      heroText: mix(heroText, other.heroText),
      heroMuted: mix(heroMuted, other.heroMuted),
      iva: mix(iva, other.iva),
      isr: mix(isr, other.isr),
      favor: mix(favor, other.favor),
      legendBg: mix(legendBg, other.legendBg),
      legendText: mix(legendText, other.legendText),
      error: mix(error, other.error),
    );
  }
}
