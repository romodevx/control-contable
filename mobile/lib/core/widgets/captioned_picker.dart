import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';

/// A visible [caption] above a [picker] that screen readers announce once:
/// the picker carries the caption as its label and the caption itself is
/// kept out of the semantics tree (otherwise VoiceOver and TalkBack read
/// "Ejercicio, Ejercicio, 2026").
class CaptionedPicker extends StatelessWidget {
  /// [picker] under [caption], drawn in [style] (the body text when `null`).
  const new({
    required this.caption,
    required this.picker,
    this.style,
    super.key,
  });

  /// The visible caption, also the picker's spoken label.
  final String caption;

  /// The control, usually a `DropdownButtonFormField`.
  final Widget picker;

  /// The caption's text style.
  final TextStyle? style;

  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    mainAxisSize: MainAxisSize.min,
    children: [
      ExcludeSemantics(child: Text(caption, style: style)),
      const SizedBox(height: AppSpacing.xxs),
      Semantics(label: caption, child: picker),
    ],
  );
}
