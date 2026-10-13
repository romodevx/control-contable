import 'package:control_contable/core/theme/app_colors.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';

const _tabularFigures = [FontFeature.tabularFigures()];
const _outlineWidth = 1.0;
const _focusWidth = 2.0;

/// Builds the Material 3 theme for one [AppTokens] set.
///
/// The only place tokens are mapped onto Material; the tokens are also
/// registered as a [ThemeExtension] for what Material has no slot for (hero,
/// legend, chart series). Sizes live in the [TextTheme], so system text
/// scaling still applies.
ThemeData buildAppTheme(AppTokens tokens, Brightness brightness) {
  final c = tokens.colors;
  final scheme = _colorScheme(c, brightness);
  return ThemeData(
    useMaterial3: true,
    brightness: brightness,
    colorScheme: scheme,
    scaffoldBackgroundColor: c.bg,
    fontFamily: tokens.typography.bodyFamily,
    textTheme: _textTheme(tokens),
    materialTapTargetSize: MaterialTapTargetSize.padded,
    appBarTheme: AppBarTheme(
      backgroundColor: c.bg,
      foregroundColor: c.text,
      elevation: 0,
      scrolledUnderElevation: 0,
    ),
    cardTheme: CardThemeData(
      color: c.surface,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(tokens.radii.card),
        side: BorderSide(color: c.border),
      ),
    ),
    dividerTheme: DividerThemeData(color: c.border, thickness: 1),
    inputDecorationTheme: _inputTheme(tokens),
    filledButtonTheme: FilledButtonThemeData(style: _buttonStyle(tokens)),
    outlinedButtonTheme: OutlinedButtonThemeData(
      style: _buttonStyle(tokens)
          .copyWith(side: WidgetStatePropertyAll(BorderSide(color: c.outline))),
    ),
    textButtonTheme: TextButtonThemeData(
      style: TextButton.styleFrom(
        minimumSize: const Size(AppSizes.minTarget, AppSizes.minTarget),
      ),
    ),
    navigationBarTheme: NavigationBarThemeData(
      backgroundColor: c.surface,
      elevation: 0,
      indicatorColor: Colors.transparent,
    ),
    extensions: [tokens],
  );
}

ColorScheme _colorScheme(AppColors c, Brightness brightness) => ColorScheme(
  brightness: brightness,
  primary: c.primary,
  onPrimary: c.onPrimary,
  secondary: c.isr,
  onSecondary: c.onPrimary,
  error: c.error,
  onError: c.surface,
  surface: c.surface,
  onSurface: c.text,
  onSurfaceVariant: c.muted,
  outline: c.outline,
  outlineVariant: c.border,
);

TextTheme _textTheme(AppTokens tokens) {
  final t = tokens.typography;
  final c = tokens.colors;
  TextStyle display(double size, FontWeight weight) => TextStyle(
    fontFamily: t.displayFamily,
    fontSize: size,
    fontWeight: weight,
    color: c.text,
    fontFeatures: _tabularFigures,
  );
  TextStyle body(double size, FontWeight weight, Color color) => TextStyle(
    fontFamily: t.bodyFamily,
    fontSize: size,
    fontWeight: weight,
    color: color,
    fontFeatures: _tabularFigures,
  );
  return TextTheme(
    displayLarge: display(t.hero, FontWeight.w700),
    headlineMedium: display(t.page, FontWeight.w600),
    titleLarge: display(t.title, FontWeight.w600),
    titleMedium: display(t.section, FontWeight.w600),
    bodyLarge: body(t.input, FontWeight.w400, c.text),
    bodyMedium: body(t.body, FontWeight.w400, c.text),
    bodySmall: body(t.help, FontWeight.w400, c.muted),
    labelLarge: body(t.input, FontWeight.w600, c.text),
    labelSmall: body(t.caption, FontWeight.w500, c.muted),
  );
}

InputDecorationTheme _inputTheme(AppTokens tokens) {
  final c = tokens.colors;
  OutlineInputBorder border(Color color, double width) => OutlineInputBorder(
    borderRadius: BorderRadius.circular(tokens.radii.control),
    borderSide: BorderSide(color: color, width: width),
  );
  return InputDecorationTheme(
    filled: true,
    fillColor: c.surface,
    constraints: const BoxConstraints(minHeight: AppSizes.control),
    contentPadding: const EdgeInsets.symmetric(
      horizontal: AppSpacing.md,
      vertical: AppSpacing.sm,
    ),
    hintStyle: TextStyle(color: c.muted),
    helperStyle: TextStyle(color: c.muted, fontSize: tokens.typography.help),
    errorStyle: TextStyle(color: c.error, fontSize: tokens.typography.help),
    border: border(c.outline, _outlineWidth),
    enabledBorder: border(c.outline, _outlineWidth),
    focusedBorder: border(c.primary, _focusWidth),
    errorBorder: border(c.error, _outlineWidth),
    focusedErrorBorder: border(c.error, _focusWidth),
  );
}

ButtonStyle _buttonStyle(AppTokens tokens) => ButtonStyle(
  minimumSize: const WidgetStatePropertyAll(
    Size(AppSizes.minTarget, AppSizes.control),
  ),
  shape: WidgetStatePropertyAll(
    RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(tokens.radii.button),
    ),
  ),
  textStyle: WidgetStatePropertyAll(
    TextStyle(
      fontFamily: tokens.typography.bodyFamily,
      fontSize: tokens.typography.input,
      fontWeight: FontWeight.w600,
    ),
  ),
);
