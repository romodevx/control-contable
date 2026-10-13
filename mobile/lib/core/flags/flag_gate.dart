import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Renders [child] only while [module] is enabled; otherwise nothing (the
/// child is not in the tree, so it is neither shown nor announced).
///
/// Use it for entry points into a module (menu items, cards, buttons).
class FlagGate extends ConsumerWidget {
  /// Gates [child] behind [module].
  const new({required this.module, required this.child, super.key});

  /// The module whose flag decides.
  final AppModule module;

  /// What to show while the module is enabled.
  final Widget child;

  @override
  Widget build(BuildContext context, WidgetRef ref) =>
      ref.useFlag(module) ? child : const SizedBox.shrink();
}
