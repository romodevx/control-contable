import 'dart:async';

import 'package:control_contable/core/theme/app_tokens.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// A row that opens [location]: [title], an optional [subtitle] (a counter,
/// the months concerned) and a chevron. The whole row is the target, at
/// least 44 × 44, read as one button.
class AnnualLinkRow extends StatelessWidget {
  /// The row of [title] opening [location].
  const new({
    required this.title,
    required this.location,
    this.subtitle,
    super.key,
  });

  /// What the row opens.
  final String title;

  /// A detail under the title.
  final String? subtitle;

  /// Where the row goes.
  final String location;

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    final detail = subtitle;
    // Its own semantics node: a lone row would otherwise merge into the
    // card around it, and the whole card would be announced as this button.
    // Its own transparent Material: the cards it sits on paint a surface,
    // which would hide the ink of the nearest Material ancestor.
    return Semantics(
      container: true,
      child: Material(
        type: MaterialType.transparency,
        child: ListTile(
          contentPadding: EdgeInsets.zero,
          title: Text(title, style: text.bodyLarge),
          subtitle: detail == null
              ? null
              : Text(
                  detail,
                  style: text.bodyMedium?.copyWith(
                    color: context.tokens.colors.muted,
                  ),
                ),
          trailing: const Icon(Icons.chevron_right),
          onTap: () => unawaited(context.push<void>(location)),
        ),
      ),
    );
  }
}
