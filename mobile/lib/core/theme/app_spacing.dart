/// Spacing and sizes shared by every design direction (DIRECTIONS.md, "What
/// every direction shares"; ~/.claude/rules/ui-forms-and-modals.md).
abstract final class AppSpacing {
  /// 4-point grid.
  static const double xxs = 4;

  /// 8.
  static const double xs = 8;

  /// 12.
  static const double sm = 12;

  /// 16: screen gutter.
  static const double md = 16;

  /// 20: tight gap between sections.
  static const double lg = 20;

  /// 24: gap between sections.
  static const double xl = 24;

  /// 32.
  static const double xxl = 32;

  /// 48.
  static const double xxxl = 48;

  /// Horizontal screen gutter.
  static const double gutter = md;
}

/// Control sizes fixed by the house form rules.
abstract final class AppSizes {
  /// Minimum touch target, both axes (house rule, stricter than WCAG 2.2).
  static const double minTarget = 44;

  /// Input and primary button height.
  static const double control = 52;
}
