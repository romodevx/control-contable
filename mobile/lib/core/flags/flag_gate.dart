import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Renders [child] only while [module] is enabled; otherwise [fallback],
/// nothing by default (the child is not in the tree, so it is neither shown
/// nor announced).
///
/// Use it for entry points into a module (menu items, cards, buttons), and
/// with a [fallback] where a module replaces what was there before (the
/// dashboard replacing the placeholder home).
class FlagGate extends ConsumerWidget {
  /// Gates [child] behind [module].
  const new({
    required this.module,
    required this.child,
    this.fallback = const SizedBox.shrink(),
    super.key,
  });

  /// The module whose flag decides.
  final AppModule module;

  /// What to show while the module is enabled.
  final Widget child;

  /// What to show while the module is disabled (or the flags still load).
  final Widget fallback;

  @override
  Widget build(BuildContext context, WidgetRef ref) =>
      ref.useFlag(module) ? child : fallback;
}
