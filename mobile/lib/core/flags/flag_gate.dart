import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter/widgets.dart';

/// Placeholder: implemented after the failing test (TDD).
class FlagGate extends StatelessWidget {
  const new({required this.module, required this.child, super.key});

  final AppModule module;
  final Widget child;

  @override
  Widget build(BuildContext context) => child;
}
