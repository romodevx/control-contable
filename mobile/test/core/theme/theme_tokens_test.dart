import 'dart:math' as math;

import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:control_contable/core/theme/design_direction.dart';
import 'package:control_contable/core/theme/directions/confianza_sobria.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';

/// WCAG 2.x contrast ratio, the method DIRECTIONS.md uses.
double _contrast(Color a, Color b) {
  final la = a.computeLuminance();
  final lb = b.computeLuminance();
  return (math.max(la, lb) + 0.05) / (math.min(la, lb) + 0.05);
}

const _textMin = 4.5;
const _nonTextMin = 3.0;

void _expectContrast(Map<String, (Color, Color)> pairs, double minimum) {
  for (final MapEntry(key: name, value: (fg, bg)) in pairs.entries) {
    expect(_contrast(fg, bg), greaterThanOrEqualTo(minimum), reason: name);
  }
}

ThemeData _themeOfHome(WidgetTester tester) =>
    Theme.of(tester.element(find.byType(HomeScreen)));

void main() {
  group('direction A · Confianza sobria resolves through the theme', () {
    testWidgets('light', (tester) async {
      await pumpControlContableApp(tester);
      final theme = _themeOfHome(tester);
      final tokens = theme.extension<AppTokens>()!;

      expect(tokens.colors.primary, const Color(0xFF1D4E89));
      expect(tokens.colors.bg, const Color(0xFFF5F7FA));
      expect(theme.brightness, Brightness.light);
      expect(theme.colorScheme.primary, tokens.colors.primary);
      expect(theme.colorScheme.surface, tokens.colors.surface);
      expect(theme.colorScheme.outline, tokens.colors.outline);
      expect(theme.scaffoldBackgroundColor, tokens.colors.bg);
      expect(tokens.typography.displayFamily, 'Source Serif 4');
      expect(tokens.typography.bodyFamily, 'Public Sans');
    });

    testWidgets('dark', (tester) async {
      await pumpControlContableApp(tester, brightness: Brightness.dark);
      final theme = _themeOfHome(tester);
      final tokens = theme.extension<AppTokens>()!;

      expect(tokens.colors.primary, const Color(0xFF8DB8EC));
      expect(tokens.colors.bg, const Color(0xFF0B1420));
      expect(theme.brightness, Brightness.dark);
      expect(theme.colorScheme.primary, tokens.colors.primary);
      expect(theme.colorScheme.onPrimary, tokens.colors.onPrimary);
    });

    testWidgets('widgets read tokens through context.tokens', (tester) async {
      await pumpControlContableApp(tester, brightness: Brightness.dark);
      final context = tester.element(find.byType(HomeScreen));

      expect(context.tokens, same(Theme.of(context).extension<AppTokens>()));
    });
  });

  testWidgets('swapping the direction re-themes without touching widgets', (
    tester,
  ) async {
    final swapped = DesignDirection(
      name: 'test',
      light: confianzaSobria.light.copyWith(
        colors: confianzaSobria.light.colors.copyWith(
          primary: const Color(0xFFA13F1C),
        ),
      ),
      dark: confianzaSobria.dark,
    );

    await pumpControlContableApp(
      tester,
      overrides: [designDirectionProvider.overrideWithValue(swapped)],
    );

    expect(_themeOfHome(tester).colorScheme.primary, const Color(0xFFA13F1C));
  });

  group('every documented contrast pair passes AA', () {
    final modes = {
      'light': confianzaSobria.light.colors,
      'dark': confianzaSobria.dark.colors,
    };
    for (final MapEntry(key: mode, value: c) in modes.entries) {
      test(mode, () {
        final textPairs = {
          'text/bg': (c.text, c.bg),
          'text/surface': (c.text, c.surface),
          'muted/surface': (c.muted, c.surface),
          'onPrimary/primary': (c.onPrimary, c.primary),
          'heroText/heroBg': (c.heroText, c.heroBg),
          'heroMuted/heroBg': (c.heroMuted, c.heroBg),
          'favor/surface': (c.favor, c.surface),
          'legendText/legendBg': (c.legendText, c.legendBg),
          'error/surface': (c.error, c.surface),
        };
        final nonTextPairs = {
          'iva/surface': (c.iva, c.surface),
          'isr/surface': (c.isr, c.surface),
          'outline/surface': (c.outline, c.surface),
        };
        _expectContrast(textPairs, _textMin);
        _expectContrast(nonTextPairs, _nonTextMin);
      });
    }
  });

  test('tokens interpolate for animated theme changes', () {
    final light = confianzaSobria.light;
    final dark = confianzaSobria.dark;

    expect(light.lerp(dark, 0).colors.primary, light.colors.primary);
    expect(light.lerp(dark, 1).colors.primary, dark.colors.primary);
    expect(light.lerp(null, 0.5), same(light));
  });
}
