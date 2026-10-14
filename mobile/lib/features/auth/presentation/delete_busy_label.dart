import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';

const double _spinnerSize = 20;
const double _spinnerStroke = 2;

/// The busy content of the destructive button: a spinner beside [label],
/// read by screen readers as [announcement] (*Eliminando cuenta…*).
class DeleteBusyLabel extends StatelessWidget {
  /// The busy state of a button reading [label].
  const new({required this.label, required this.announcement, super.key});

  /// The button's own label, kept so its width does not jump.
  final String label;

  /// What screen readers hear while it runs.
  final String announcement;

  @override
  Widget build(BuildContext context) => Semantics(
    label: announcement,
    excludeSemantics: true,
    child: Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox.square(
          dimension: _spinnerSize,
          child: CircularProgressIndicator(
            strokeWidth: _spinnerStroke,
            color: Theme.of(context).colorScheme.onError,
          ),
        ),
        const SizedBox(width: AppSpacing.xs),
        Text(label),
      ],
    ),
  );
}
