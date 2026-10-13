import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/widgets.dart';

/// Room the sticky footer takes (primary row, secondary row, padding). A
/// focused field scrolls at least this far clear of the bottom, so the
/// footer never covers it (WCAG 2.2 · 2.4.11).
const double formFooterClearance = AppSizes.control * 2 + AppSpacing.md * 3;

/// Gives a form's fields its controller and its submit action (Enter from a
/// single-line field, Ctrl/Cmd+Enter from any field). Put in place by
/// `AppForm`.
class AppFormScope extends InheritedWidget {
  /// Scopes [controller] and [submit] over [child].
  const new({
    required this.controller,
    required this.submit,
    required super.child,
    super.key,
  });

  /// The form's fields and submit state.
  final AppFormController controller;

  /// Validates and saves with the primary action.
  final VoidCallback submit;

  /// The nearest scope. Form fields only work inside an `AppForm`.
  static AppFormScope of(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<AppFormScope>();
    assert(scope != null, 'Form fields must be placed inside an AppForm.');
    return scope!;
  }

  @override
  bool updateShouldNotify(AppFormScope oldWidget) =>
      controller != oldWidget.controller || submit != oldWidget.submit;
}
