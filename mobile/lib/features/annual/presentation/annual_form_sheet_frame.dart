import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:flutter/material.dart';

/// The frame of a form inside a bottom sheet: its [title] as a heading,
/// then [form] (an `AppForm` with `fitContent`), lifted above the keyboard
/// so it never covers the fields. The sheet is as tall as its content.
class AnnualFormSheetFrame extends StatelessWidget {
  /// A sheet titled [title] holding [form].
  const new({required this.title, required this.form, super.key});

  /// The sheet's heading.
  final String title;

  /// The form.
  final Widget form;

  @override
  Widget build(BuildContext context) => Padding(
    padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
    child: Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.gutter,
            AppSpacing.lg,
            AppSpacing.gutter,
            0,
          ),
          child: Semantics(
            header: true,
            child: Text(
              title,
              style: Theme.of(context).textTheme.headlineSmall,
            ),
          ),
        ),
        Flexible(child: form),
      ],
    ),
  );
}
